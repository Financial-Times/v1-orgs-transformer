package main

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/pborman/uuid"
	"net/http"
	"time"
)

const cacheBucket = "org"
const cacheFileName = "cache.db"

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type orgsService interface {
	getOrgs() ([]orgLink, bool)
	getOrgByUUID(uuid string) (org, bool)
}

type orgServiceImpl struct {
	repository    tmereader.Repository
	baseURL       string
	orgLinks      []orgLink
	taxonomyName  string
	maxTmeRecords int
}

func newOrgService(repo tmereader.Repository, baseURL string, taxonomyName string, maxTmeRecords int) (orgsService, error) {

	s := &orgServiceImpl{repository: repo, baseURL: baseURL, taxonomyName: taxonomyName, maxTmeRecords: maxTmeRecords}
	err := s.init()
	if err != nil {
		return &orgServiceImpl{}, err
	}
	return s, nil
}

func (s *orgServiceImpl) init() error {
	db, err := bolt.Open(cacheFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()
	if err = createCacheBucket(db); err != nil {
		return err
	}

	responseCount := 0
	log.Printf("Fetching organisations from TME\n")
	for {
		terms, err := s.repository.GetTmeTermsFromIndex(responseCount)
		if err != nil {
			return err
		}
		if len(terms) < 1 {
			log.Printf("Finished fetching organisations from TME\n")
			break
		}
		s.initOrgsMap(terms, db)
		responseCount += s.maxTmeRecords
	}
	log.Printf("Added %d orgs links\n", len(s.orgLinks))
	return nil
}

func createCacheBucket(db *bolt.DB) (error) {
	return db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(cacheBucket))

		_, err := tx.CreateBucket([]byte(cacheBucket))
		if err != nil {
			return err
		}
		return nil
	})

}

func (s *orgServiceImpl) getOrgs() ([]orgLink, bool) {
	//TODO implement 503 response when init is still ongoing
	if len(s.orgLinks) > 0 {
		return s.orgLinks, true
	}
	return s.orgLinks, false
}

func (s *orgServiceImpl) getOrgByUUID(uuid string) (org, bool) {
	//TODO implement 503 response when init is still ongoing
	db, err := bolt.Open(cacheFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf(err.Error())
		return org{}, false
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
		log.Errorf(err.Error())
		return org{}, false
	}
	if cachedValue == nil || len(cachedValue) == 0 {
		log.Errorf(err.Error())
		return org{}, false
	}
	var cachedOrg org
	err = json.Unmarshal(cachedValue, &cachedOrg)
	if err != nil {
		log.Errorf(err.Error())
		return org{}, false
	}
	return cachedOrg, true

}

func (s *orgServiceImpl) initOrgsMap(terms []interface{}, db *bolt.DB) {
	var cacheToBeWritten []org
	for _, iTerm := range terms {
		t := iTerm.(term)
		tmeIdentifier := buildTmeIdentifier(t.RawID, s.taxonomyName)
		uuid := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
		s.orgLinks = append(s.orgLinks, orgLink{APIURL: s.baseURL + uuid})
		cacheToBeWritten = append(cacheToBeWritten, transformOrg(t, s.taxonomyName))
	}
	storeOrgToCache(db, cacheToBeWritten)
}

func storeOrgToCache(db *bolt.DB, cacheToBeWritten []org) {
	start := time.Now()

	defer func(startTime time.Time) {
		log.Printf("Done, elapsed time: %+v, size: %v\n", time.Since(startTime), len(cacheToBeWritten))
	}(start)

	err := db.Batch(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(cacheBucket))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found!", cacheBucket)
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
		log.Errorf("ERROR store: %+v", err)
	}

}
