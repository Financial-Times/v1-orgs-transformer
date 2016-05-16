package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testUUID = "bba39990-c78d-3629-ae83-808c333c6dbc"
const getOrganisationsResponse = "[{\"apiUrl\":\"http://localhost:8080/transformers/organisations/bba39990-c78d-3629-ae83-808c333c6dbc\"}]\n"
const getOrganisationByUUIDResponse = "{\"uuid\":\"bba39990-c78d-3629-ae83-808c333c6dbc\",\"properName\":\"European Union\",\"type\":\"Organisation\",\"identifiers\":[{" +
	"\"authority\":\"http://api.ft.com/system/FT-TME\"," +
	"\"identifierValue\":\"MTE3-U3ViamVjdHM=\"" +
	"}]}\n"

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
		{"Success - get organisation by uuid", newRequest("GET", fmt.Sprintf("/transformers/organisations/%s", testUUID)), &dummyService{found: true, initialised: true, orgs: []org{org{UUID: testUUID, ProperName: "European Union", Identifiers: []identifier{identifier{Authority: "http://api.ft.com/system/FT-TME", IdentifierValue: "MTE3-U3ViamVjdHM="}}, Type: "Organisation"}}}, http.StatusOK, "application/json", getOrganisationByUUIDResponse},
		{"Not found - get organisation by uuid", newRequest("GET", fmt.Sprintf("/transformers/organisations/%s", testUUID)), &dummyService{found: false, initialised: true, orgs: []org{org{}}}, http.StatusNotFound, "application/json", ""},
		{"Service unavailable - get organisation by uuid", newRequest("GET", fmt.Sprintf("/transformers/organisations/%s", testUUID)), &dummyService{found: false, initialised: false, orgs: []org{}}, http.StatusServiceUnavailable, "application/json", ""},
		{"Success - get organisations", newRequest("GET", "/transformers/organisations"), &dummyService{found: true, initialised: true, orgs: []org{org{UUID: testUUID}}}, http.StatusOK, "application/json", getOrganisationsResponse},
		{"Not found - get organisations", newRequest("GET", "/transformers/organisations"), &dummyService{found: false, initialised: true, orgs: []org{}}, http.StatusNotFound, "application/json", ""},
		{"Service unavailable - get organisations", newRequest("GET", "/transformers/organisations"), &dummyService{found: false, initialised: false, orgs: []org{}}, http.StatusServiceUnavailable, "application/json", ""},
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
	m.HandleFunc("/transformers/organisations", h.getOrgs).Methods("GET")
	m.HandleFunc("/transformers/organisations/{uuid}", h.getOrgByUUID).Methods("GET")
	return m
}

type dummyService struct {
	found       bool
	orgs        []org
	initialised bool
}

func (s *dummyService) getOrgs() ([]orgLink, bool) {
	var orgLinks []orgLink
	for _, sub := range s.orgs {
		orgLinks = append(orgLinks, orgLink{APIURL: "http://localhost:8080/transformers/organisations/" + sub.UUID})
	}
	return orgLinks, s.found
}

func (s *dummyService) getOrgByUUID(uuid string) (org, bool, error) {
	return s.orgs[0], s.found, nil
}

func (s *dummyService) isInitialised() bool {
	return s.initialised
}
