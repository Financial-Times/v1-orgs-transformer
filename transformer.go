package main

import (
	"encoding/base64"
	"encoding/xml"

	"github.com/pborman/uuid"
)

func transformOrg(tmeTerm term, taxonomyName string) org {
	tmeIdentifier := buildTmeIdentifier(tmeTerm.RawID, taxonomyName)
	orgUUID := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
	orgAliasList := buildAliasList(tmeTerm.Aliases)
	return org{
		UUID:       orgUUID,
		ProperName: tmeTerm.CanonicalName,
		PrefLabel:  tmeTerm.CanonicalName,
		AlternativeIdentifiers: alternativeIdentifiers{
			TME:   []string{tmeIdentifier},
			Uuids: []string{orgUUID},
		},
		Type:    "Organisation",
		Aliases: orgAliasList,
	}
}

func buildTmeIdentifier(rawID string, tmeTermTaxonomyName string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawID))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(tmeTermTaxonomyName))
	return id + "-" + taxonomyName
}

func buildAliasList(aList aliases) []string {
	aliasList := make([]string, len(aList.Alias))
	for k, v := range aList.Alias {
		aliasList[k] = v.Name
	}
	return aliasList
}

type orgTransformer struct {
}

func (*orgTransformer) UnMarshallTaxonomy(contents []byte) ([]interface{}, error) {
	t := taxonomy{}
	err := xml.Unmarshal(contents, &t)
	if err != nil {
		return nil, err
	}
	interfaces := make([]interface{}, len(t.Terms))
	for i, d := range t.Terms {
		interfaces[i] = d
	}
	return interfaces, nil
}

func (*orgTransformer) UnMarshallTerm(content []byte) (interface{}, error) {
	dummyTerm := term{}
	err := xml.Unmarshal(content, &dummyTerm)
	if err != nil {
		return term{}, err
	}
	return dummyTerm, nil
}
