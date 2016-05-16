package main

import (
	"encoding/base64"
	"encoding/xml"
	"github.com/pborman/uuid"
)

func transformOrg(tmeTerm term, taxonomyName string) org {
	tmeIdentifier := buildTmeIdentifier(tmeTerm.RawID, taxonomyName)

	return org{
		UUID:       uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String(),
		ProperName: tmeTerm.CanonicalName,
		Identifiers: []identifier{
			identifier{Authority: "http://api.ft.com/system/FT-TME", IdentifierValue: tmeIdentifier},
		},
		Type: "Organisation",
	}
}

func buildTmeIdentifier(rawID string, tmeTermTaxonomyName string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawID))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(tmeTermTaxonomyName))
	return id + "-" + taxonomyName
}

type orgTransformer struct {
}

func (*orgTransformer) UnMarshallTaxonomy(contents []byte) ([]interface{}, error) {
	taxonomy := taxonomy{}
	err := xml.Unmarshal(contents, &taxonomy)
	if err != nil {
		return nil, err
	}
	interfaces := make([]interface{}, len(taxonomy.Terms))
	for i, d := range taxonomy.Terms {
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
