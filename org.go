package main

//model aligned with v2-org-transformer
type org struct {
	UUID                   string                 `json:"uuid"`
	ProperName             string                 `json:"properName"`
	PrefLabel              string                 `json:"prefLabel"`
	PrimaryType            string                 `json:"type"`
	TypeHierarchy          []string               `json:"types"`
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers,omitempty"`
	Aliases                []string               `json:"aliases,omitempty"`
}

type alternativeIdentifiers struct {
	TME   []string `json:"TME,omitempty"`
	Uuids []string `json:"uuids,omitempty"`
}

type orgLink struct {
	APIURL string `json:"apiUrl"`
}

type orgUUID struct {
	UUID string `json:"ID"`
}

var primaryType = "Organisation"
var orgTypes = []string{"Thing", "Concept", "Organisation"}
