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
)

const (
	cacheBucket = "org"
)

type orgsService interface {
	getOrgs() ([]orgLink, error)
	getOrgByUUID(uuid string) (org, bool, error)
	isInitialised() bool
	isDataLoaded() bool
	shutdown() error
	orgCount() (int, error)
	orgIds() ([]orgUUID, error)
	orgReload() error
}

type orgServiceImpl struct {
	sync.RWMutex
	repository    tmereader.Repository
	baseURL       string
	taxonomyName  string
	maxTmeRecords int
	initialised   bool
	dataLoaded    bool
	cacheFileName string
	db            *bolt.DB
}

func newOrgService(repo tmereader.Repository, baseURL string, taxonomyName string, maxTmeRecords int, cacheFileName string) orgsService {
	s := &orgServiceImpl{repository: repo, baseURL: baseURL, taxonomyName: taxonomyName, maxTmeRecords: maxTmeRecords, initialised: true, dataLoaded: false, cacheFileName: cacheFileName}
	go func(service *orgServiceImpl) {
		err := service.init()
		if err != nil {
			log.Errorf("Error while creating OrgService: [%v]", err.Error())
		}
		s.setDataLoaded(true)
	}(s)
	return s
}

func (s *orgServiceImpl) isInitialised() bool {
	s.RLock()
	defer s.RUnlock()
	return s.initialised
}

func (s *orgServiceImpl) setInitialised(val bool) {
	s.Lock()
	defer s.Unlock()
	s.initialised = val
}

func (s *orgServiceImpl) isDataLoaded() bool {
	s.RLock()
	defer s.RUnlock()
	return s.dataLoaded
}

func (s *orgServiceImpl) setDataLoaded(val bool) {
	s.Lock()
	defer s.Unlock()
	s.dataLoaded = val
}

func (s *orgServiceImpl) shutdown() error {
	if s.db == nil {
		return errors.New("DB not open")
	}
	return s.db.Close()
}

func (s *orgServiceImpl) openDB() error {
	s.Lock()
	defer s.Unlock()
	var err error
	s.db, err = bolt.Open(s.cacheFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf("ERROR opening cache file for init: %v", err.Error())
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(cacheBucket))
		if err != nil {
			log.Warnf("Cache bucket [%v] could not be deleted\n", cacheBucket)
		}
		_, err = tx.CreateBucket([]byte(cacheBucket))
		return err
	})
}

func (s *orgServiceImpl) init() error {
	var wg sync.WaitGroup
	responseCount := 0
	s.dataLoaded = false
	log.Printf("Fetching organisations from TME\n")

	err := s.openDB()
	if err != nil {
		return err
	}

	for {
		log.Printf("Getting terms for responseCount %d", responseCount)
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

	count, _ := s.orgCount()
	log.Printf("Added %d orgs UUIDs\n", count)
	return nil
}

func (s *orgServiceImpl) getOrgs() ([]orgLink, error) {
	s.RLock()
	defer s.RUnlock()
	var linkList []orgLink
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found!", cacheBucket)
		}

		bucket.ForEach(func(k, v []byte) error {
			linkList = append(linkList, orgLink{APIURL: s.baseURL + string(k)})
			return nil
		})
		return nil
	})

	return linkList, err
}

func (s *orgServiceImpl) getOrgByUUID(uuid string) (org, bool, error) {
	s.RLock()
	defer s.RUnlock()
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
		cacheToBeWritten = append(cacheToBeWritten, transformOrg(iTerm.(term), s.taxonomyName))
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

// HELPER METHODS

func (s *orgServiceImpl) orgCount() (int, error) {
	var count int
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found!", cacheBucket)
		}

		count = bucket.Stats().KeyN
		return nil
	})

	return count, err
}

func (s *orgServiceImpl) orgIds() ([]orgUUID, error) {
	s.RLock()
	defer s.RUnlock()
	var uuidList []orgUUID
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found!", cacheBucket)
		}

		bucket.ForEach(func(k, v []byte) error {
			uuidList = append(uuidList, orgUUID{UUID: string(k)})
			return nil
		})
		return nil
	})

	return uuidList, err
}

func (s *orgServiceImpl) orgReload() error {
	err := s.shutdown()
	if err != nil {
		return err
	}

	return s.init()
}
