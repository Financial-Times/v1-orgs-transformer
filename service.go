package main

import (
	"net/http"
	"log"
	"github.com/pborman/uuid"
)

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type orgsService interface {
	getOrgs() ([]orgLink, bool)
	getOrgByUUID(uuid string) (org, bool)
}

type orgServiceImpl struct {
	repository repository
	baseURL    string
	IdMap      map[string]string
	orgLinks   []orgLink
}

func newOrgService(repo repository, baseURL string) (orgsService, error) {

	s := &orgServiceImpl{repository: repo, baseURL: baseURL}
	err := s.init()
	if err != nil {
		return &orgServiceImpl{}, err
	}
	return s, nil
}

func (s *orgServiceImpl) init() error {
	s.IdMap = make(map[string]string)
	responseCount := 0
	log.Printf("Fetching organisations from TME\n")
	for {
		tax, err := s.repository.getOrgsTaxonomy(responseCount)
		if err != nil {
			return err
		}
		if (len(tax.Terms) < 1) {
			break
		}
		s.initOrgsMap(tax.Terms)
		responseCount += MaxRecords
	}
	log.Printf("Added %d orgs links\n", len(s.orgLinks))
	return nil
}

func (s *orgServiceImpl) getOrgs() ([]orgLink, bool) {
	if len(s.orgLinks) > 0 {
		return s.orgLinks, true
	}
	return s.orgLinks, false
}

func (s *orgServiceImpl) getOrgByUUID(uuid string) (org, bool) {
	rawId, found := s.IdMap[uuid]
	if !found {
		return org{}, false
	}
	term, err := s.repository.getSingleOrgTaxonomy(rawId)
	if err != nil {
		return org{}, false
	}
	return transformOrg(term), true
}

func (s *orgServiceImpl) initOrgsMap(terms []term) {
	for _, t := range terms {
		tmeIdentifier := buildTmeIdentifier(t.RawID)
		uuid := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
		s.IdMap[uuid] = t.RawID
		s.orgLinks = append(s.orgLinks, orgLink{APIURL: s.baseURL + uuid})
	}
}
