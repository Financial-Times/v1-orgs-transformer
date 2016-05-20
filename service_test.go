package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetOrganisations(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name    string
		baseURL string
		terms   []term
		orgs    []orgLink
		found   bool
		err     error
	}{
		{"Success", "localhost:8080/transformers/organsiations/",
			[]term{term{CanonicalName: "European Union", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			[]orgLink{orgLink{APIURL: "localhost:8080/transformers/organsiations/6a7edb42-c27a-3186-a0b9-7e3cdc91e16b"}}, true, nil},
		{"Error on init", "localhost:8080/transformers/organsiations/", []term{}, []orgLink(nil), false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service := newOrgService(&repo, test.baseURL, "ON", 10000, "cache.db")
		time.Sleep(3 * time.Second) //waiting initialization to be finished
		actualOrgansiations, found := service.getOrgs()
		assert.Equal(test.orgs, actualOrgansiations, fmt.Sprintf("%s: Expected organsiations link incorrect", test.name))
		assert.Equal(test.found, found)
	}
}

func TestGetOrganisationByUuid(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name  string
		terms []term
		uuid  string
		org   org
		found bool
		err   error
	}{
		{"Success", []term{term{CanonicalName: "European Union", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			"6a7edb42-c27a-3186-a0b9-7e3cdc91e16b", org{UUID: "6a7edb42-c27a-3186-a0b9-7e3cdc91e16b", ProperName: "European Union", Identifiers: []identifier{
				identifier{Authority: "http://api.ft.com/system/FT-TME", IdentifierValue: "TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-T04="}}, Type: "Organisation"}, true, nil},
		{"Not found", []term{term{CanonicalName: "European Union", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			"some uuid", org{}, false, nil},
		{"Error on init", []term{}, "some uuid", org{}, false, nil},
	}
	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service := newOrgService(&repo, "", "ON", 10000, "cache.db")
		time.Sleep(3 * time.Second) //waiting initialization to be finished
		actualOrganisation, found, err := service.getOrgByUUID(test.uuid)
		assert.Equal(test.org, actualOrganisation, fmt.Sprintf("%s: Expected organsiation incorrect", test.name))
		assert.Equal(test.found, found)
		assert.Equal(test.err, err)
	}
}

type dummyRepo struct {
	terms []term
	err   error
}

func (d *dummyRepo) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
	if startRecord > 0 {
		return nil, d.err
	}
	var interfaces []interface{} = make([]interface{}, len(d.terms))
	for i, data := range d.terms {
		interfaces[i] = data
	}
	return interfaces, d.err
}
func (d *dummyRepo) GetTmeTermById(uuid string) (interface{}, error) {
	return d.terms[0], d.err
}
