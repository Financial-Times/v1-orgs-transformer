package main

import (
	"github.com/pborman/uuid"
	"encoding/base64"
)

func transformOrg(t term) org {
	tmeIdentifier := buildTmeIdentifier(t.RawID)

	return org{
		UUID:          uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String(),
		CanonicalName: t.CanonicalName,
		TmeIdentifier: tmeIdentifier,
		Type:          "Organization",
	}
}

func buildTmeIdentifier(rawId string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawId))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(TaxonomyName))
	return id + "-" + taxonomyName
}
