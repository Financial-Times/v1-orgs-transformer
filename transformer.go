package main

import (
	"github.com/pborman/uuid"
	"encoding/base64"
	"encoding/xml"
)

func transformOrg(tmeTerm term, taxonomyName string) org {
	tmeIdentifier := buildTmeIdentifier(tmeTerm.RawID, taxonomyName)

	return org{
		UUID:          uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String(),
		ProperName: tmeTerm.CanonicalName,
		Identifiers: []identifier {
			identifier{Authority:"http://api.ft.com/system/FT-TME", IdentifierValue:tmeIdentifier},
		},
		Type:          "Organization",
	}
}

func buildTmeIdentifier(rawId string, tmeTermTaxonomyName string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawId))
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
	var interfaces []interface{} = make([]interface{}, len(taxonomy.Terms))
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
