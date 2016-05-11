package main

//TODO model aligned with v2-org-transformer
type org struct {
	UUID          string       `json:"uuid"`
	ProperName    string       `json:"properName"`
	Type          string       `json:"type"`
	Identifiers   []identifier `json:"identifiers,omitempty"`
}

type identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}

type orgLink struct {
	APIURL string `json:"apiUrl"`
}
