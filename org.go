package main

//model aligned with v2-org-transformer
type org struct {
	UUID                   string                 `json:"uuid"`
	ProperName             string                 `json:"properName"`
	PrefLabel              string                 `json:"prefLabel"`
	Type                   string                 `json:"type"`
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers,omitempty"`
}

type alternativeIdentifiers struct {
	TME   []string `json:"TME,omitempty"`
	Uuids []string `json:"uuids,omitempty"`
}

type orgLink struct {
	APIURL string `json:"apiUrl"`
}

type orgUUID struct {
	UUID string `json:"uuid"`
}
