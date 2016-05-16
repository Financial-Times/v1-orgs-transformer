package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransform(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		term term
		org  org
	}{
		{"Trasform term to location", term{CanonicalName: "European Union", RawID: "Nstein_GL_US_NY_Municipality_942968"}, org{UUID: "6a7edb42-c27a-3186-a0b9-7e3cdc91e16b", ProperName: "European Union", Identifiers: []identifier{
			identifier{Authority: "http://api.ft.com/system/FT-TME", IdentifierValue: "TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-T04="}}, Type: "Organisation"}},
	}

	for _, test := range tests {
		expectedLocation := transformOrg(test.term, "ON")

		assert.Equal(test.org, expectedLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
	}

}
