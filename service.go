package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/pborman/uuid"
)

const (
	cacheBucket  = "org"
	uppAuthority = "http://api.ft.com/system/FT-UPP"
	tmeAuthority = "http://api.ft.com/system/FT-TME"
)

type orgsService interface {
	getOrgs() ([]orgLink, bool)
	getOrgByUUID(uuid string) (org, bool, error)
	isInitialised() bool
	shutdown() error
	orgCount() int
	orgIds() ([]orgUUID, error)
	orgReload() error
}

type orgServiceImpl struct {
	repository    tmereader.Repository
	baseURL       string
	orgUUIDs      []string
	taxonomyName  string
	maxTmeRecords int
	initialised   bool
	cacheFileName string
	db            *bolt.DB
}

func newOrgService(repo tmereader.Repository, baseURL string, taxonomyName string, maxTmeRecords int, cacheFileName string) orgsService {
	s := &orgServiceImpl{repository: repo, baseURL: baseURL, taxonomyName: taxonomyName, maxTmeRecords: maxTmeRecords, initialised: false, cacheFileName: cacheFileName}
	go func(service *orgServiceImpl) {
		err := service.init()
		if err != nil {
			log.Errorf("Error while creating OrgService: [%v]", err.Error())
		}
		service.initialised = true
	}(s)
	return s
}

func (s *orgServiceImpl) isInitialised() bool {
	return s.initialised
}

func (s *orgServiceImpl) shutdown() error {
	if s.db == nil {
		return errors.New("DB not open")
	}
	return s.db.Close()
}

func (s *orgServiceImpl) init() error {
	var err error
	s.db, err = bolt.Open(s.cacheFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf("ERROR opening cache file for init: %v", err.Error())
		return err
	}
	if err = createCacheBucket(s.db); err != nil {
		return err
	}
	var wg sync.WaitGroup
	responseCount := 0
	log.Printf("Fetching organisations from TME\n")
	for {
		terms, err := s.repository.GetTmeTermsFromIndex(responseCount)
		if err != nil {
			return err
		}
		if len(terms) < 1 {
			log.Printf("Finished fetching organisations from TME. Waiting subroutines to terminate\n")
			break
		}
		wg.Add(1)
		go s.initOrgsMap(terms, s.db, &wg)
		responseCount += s.maxTmeRecords
	}
	wg.Wait()
	log.Printf("Added %d orgs UUIDs\n", len(s.orgUUIDs))
	return nil
}

func createCacheBucket(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(cacheBucket))
		if err != nil {
			log.Warnf("Cache bucket [%v] could not be deleted\n", cacheBucket)
		}
		_, err = tx.CreateBucket([]byte(cacheBucket))
		return err
	})

}

func (s *orgServiceImpl) getOrgs() ([]orgLink, bool) {
	links := make([]orgLink, len(s.orgUUIDs))
	if len(s.orgUUIDs) > 0 {
		for _, uuid := range s.orgUUIDs {
			links = append(links, orgLink{APIURL: uuid})
		}
		return links, true
	}
	return links, false
}

func (s *orgServiceImpl) getOrgByUUID(uuid string) (org, bool, error) {
	var cachedValue []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found!", cacheBucket)
		}
		cachedValue = bucket.Get([]byte(uuid))
		return nil
	})

	if err != nil {
		log.Errorf("ERROR reading from cache file for [%v]: %v", uuid, err.Error())
		return org{}, false, err
	}
	if len(cachedValue) == 0 {
		log.Infof("INFO No cached value for [%v]", uuid)
		return org{}, false, nil
	}
	var cachedOrg org
	err = json.Unmarshal(cachedValue, &cachedOrg)
	if err != nil {
		log.Errorf("ERROR unmarshalling cached value for [%v]: %v", uuid, err.Error())
		return org{}, true, err
	}
	return cachedOrg, true, nil

}

func (s *orgServiceImpl) initOrgsMap(terms []interface{}, db *bolt.DB, wg *sync.WaitGroup) {
	var cacheToBeWritten []org
	for _, iTerm := range terms {
		t := iTerm.(term)
		tmeIdentifier := buildTmeIdentifier(t.RawID, s.taxonomyName)
		uuid := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
		s.orgUUIDs = append(s.orgUUIDs, uuid)
		cacheToBeWritten = append(cacheToBeWritten, transformOrg(t, s.taxonomyName))
	}

	go storeOrgToCache(db, cacheToBeWritten, wg)
}

func storeOrgToCache(db *bolt.DB, cacheToBeWritten []org, wg *sync.WaitGroup) {
	defer wg.Done()
	err := db.Batch(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Cache bucket [%v] not found!", cacheBucket)
		}
		for _, anOrg := range cacheToBeWritten {
			marshalledOrg, err := json.Marshal(anOrg)
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(anOrg.UUID), marshalledOrg)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Errorf("ERROR storing to cache: %+v", err)
	}

}

// ADMIN METHODS

func (s *orgServiceImpl) orgCount() int {
	return len(s.orgUUIDs)
}

func (s *orgServiceImpl) orgIds() ([]orgUUID, error) {
	var uuidList []orgUUID
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found!", cacheBucket)
		}

		log.Printf("List created, size = %d", bucket.Stats().KeyN)

		bucket.ForEach(func(k, v []byte) error {
			uuidList = append(uuidList, orgUUID{UUID: string(k)})
			return nil
		})
		return nil
	})

	return uuidList, err
}

func (s *orgServiceImpl) orgReload() error {
	return nil
}
