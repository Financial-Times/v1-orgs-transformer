package main

type org struct {
	UUID          string `json:"uuid"`
	CanonicalName string `json:"canonicalName"`
	TmeIdentifier string `json:"tmeIdentifier"`
	Type          string `json:"type"`
}

type orgLink struct {
	APIURL string `json:"apiUrl"`
}
