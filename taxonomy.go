package main

type taxonomy struct {
	Terms []term `xml:"term"`
}

//TODO revise fields
type term struct {
	CanonicalName string `xml:"name"`
	RawID         string `xml:"id"`
}
