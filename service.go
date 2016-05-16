package main

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/pborman/uuid"
	"sync"
	"time"
)

const cacheBucket = "org"
const cacheFileName = "cache.db"

type orgsService interface {
	getOrgs() ([]orgLink, bool)
	getOrgByUUID(uuid string) (org, bool, error)
	isInitialised() bool
}

type orgServiceImpl struct {
	repository    tmereader.Repository
	baseURL       string
	orgLinks      []orgLink
	taxonomyName  string
	maxTmeRecords int
	initialised   bool
}

func newOrgService(repo tmereader.Repository, baseURL string, taxonomyName string, maxTmeRecords int) orgsService {
	s := &orgServiceImpl{repository: repo, baseURL: baseURL, taxonomyName: taxonomyName, maxTmeRecords: maxTmeRecords, initialised: false}
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

func (s *orgServiceImpl) init() error {
	db, err := bolt.Open(cacheFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf("ERROR opening cache file for init: %v", err.Error())
		return err
	}
	defer db.Close()
	if err = createCacheBucket(db); err != nil {
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
		go s.initOrgsMap(terms, db, &wg)
		responseCount += s.maxTmeRecords
	}
	wg.Wait()
	log.Printf("Added %d orgs links\n", len(s.orgLinks))
	return nil
}

func createCacheBucket(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(cacheBucket))
		if err != nil {
			log.Warnf("Cache bucket [%v] could not be deleted\n", cacheBucket)
		}
		_, err = tx.CreateBucket([]byte(cacheBucket))
		if err != nil {
			return err
		}
		return nil
	})

}

func (s *orgServiceImpl) getOrgs() ([]orgLink, bool) {
	if len(s.orgLinks) > 0 {
		return s.orgLinks, true
	}
	return s.orgLinks, false
}

func (s *orgServiceImpl) getOrgByUUID(uuid string) (org, bool, error) {
	db, err := bolt.Open(cacheFileName, 0600, &bolt.Options{ReadOnly: true, Timeout: 10 * time.Second})
	if err != nil {
		log.Errorf("ERROR opening cache file for [%v]: %v", uuid, err.Error())
		return org{}, false, err
	}
	defer db.Close()
	var cachedValue []byte
	err = db.View(func(tx *bolt.Tx) error {
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
	if cachedValue == nil || len(cachedValue) == 0 {
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
		s.orgLinks = append(s.orgLinks, orgLink{APIURL: s.baseURL + uuid})
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
