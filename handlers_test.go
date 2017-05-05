package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const testUUID = "bba39990-c78d-3629-ae83-808c333c6dbc"
const getOrganisationsResponse = "[{\"apiUrl\":\"http://localhost:8080/transformers/organisations/bba39990-c78d-3629-ae83-808c333c6dbc\"}]\n"
const getOrganisationByUUIDResponse = "{\"uuid\":\"bba39990-c78d-3629-ae83-808c333c6dbc\",\"properName\":\"European Union\",\"prefLabel\":\"European Union\",\"type\":\"Organisation\",\"types\":[\"Thing\",\"Concept\",\"Organisation\"],\"alternativeIdentifiers\":{" +
	"\"TME\":[\"MTE3-U3ViamVjdHM=\"]," +
	"\"uuids\":[\"bba39990-c78d-3629-ae83-808c333c6dbc\"]" +
	"}}\n"
const testIDs = "{\"ID\":\"bba39990-c78d-3629-ae83-808c333c6dbc\"}\n"

func TestHandlers(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name         string
		req          *http.Request
		dummyService orgsService
		statusCode   int
		contentType  string // Contents of the Content-Type header
		body         string
	}{
		{"Success - get organisation by uuid", newRequest("GET", fmt.Sprintf("/transformers/organisations/%s", testUUID)), &dummyService{found: true, initialised: true, orgs: []org{org{UUID: testUUID, ProperName: "European Union", PrefLabel: "European Union", AlternativeIdentifiers: alternativeIdentifiers{Uuids: []string{testUUID}, TME: []string{"MTE3-U3ViamVjdHM="}}, PrimaryType: primaryType, TypeHierarchy: orgTypes}}}, http.StatusOK, "application/json", getOrganisationByUUIDResponse},
		{"Not found - get organisation by uuid", newRequest("GET", fmt.Sprintf("/transformers/organisations/%s", testUUID)), &dummyService{found: false, initialised: true, orgs: []org{org{}}}, http.StatusNotFound, "application/json", ""},
		{"Service unavailable - get organisation by uuid", newRequest("GET", fmt.Sprintf("/transformers/organisations/%s", testUUID)), &dummyService{found: false, initialised: false, orgs: []org{}}, http.StatusServiceUnavailable, "application/json", ""},
		{"Success - get organisations", newRequest("GET", "/transformers/organisations"), &dummyService{found: true, initialised: true, orgs: []org{org{UUID: testUUID}}}, http.StatusOK, "application/json", getOrganisationsResponse},
		{"Service unavailable - get organisations", newRequest("GET", "/transformers/organisations"), &dummyService{found: false, initialised: false, orgs: []org{}}, http.StatusServiceUnavailable, "application/json", ""},
		{"Success - get count", newRequest("GET", "/transformers/organisations/__count"), &dummyService{found: true, initialised: true, orgs: []org{org{UUID: testUUID}}}, http.StatusOK, "application/json", "1"},
		{"Success - get IDs", newRequest("GET", "/transformers/organisations/__ids"), &dummyService{found: true, initialised: true, orgs: []org{org{UUID: testUUID}}}, http.StatusOK, "application/json", testIDs},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(test.dummyService).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.Equal(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func newRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func router(s orgsService) *mux.Router {
	m := mux.NewRouter()
	h := newOrgsHandler(s)
	m.HandleFunc("/transformers/organisations/__count", h.getOrgCount).Methods("GET")
	m.HandleFunc("/transformers/organisations/__ids", h.getOrgIds).Methods("GET")
	m.HandleFunc("/transformers/organisations/__reload", h.reloadOrgs).Methods("POST")
	m.HandleFunc("/transformers/organisations", h.getOrgs).Methods("GET")
	m.HandleFunc("/transformers/organisations/{uuid}", h.getOrgByUUID).Methods("GET")
	return m
}

type dummyService struct {
	found       bool
	orgs        []org
	initialised bool
}

func (s *dummyService) getOrgs() ([]orgLink, error) {
	var orgLinks []orgLink
	for _, sub := range s.orgs {
		orgLinks = append(orgLinks, orgLink{APIURL: "http://localhost:8080/transformers/organisations/" + sub.UUID})
	}
	return orgLinks, nil
}

func (s *dummyService) getOrgByUUID(uuid string) (org, bool, error) {
	return s.orgs[0], s.found, nil
}

func (s *dummyService) isInitialised() bool {
	return s.initialised
}

func (s *dummyService) shutdown() error {
	return nil
}

func (s *dummyService) orgCount() (int, error) {
	return len(s.orgs), nil
}

func (s *dummyService) orgIds() ([]orgUUID, error) {
	var orgUUIDs []orgUUID
	for _, sub := range s.orgs {
		orgUUIDs = append(orgUUIDs, orgUUID{UUID: sub.UUID})
	}
	return orgUUIDs, nil
}

func (s *dummyService) orgReload() error {
	return nil
}

func (s *dummyService) isDataLoaded() bool {
	return true
}
