package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransform(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		term term
		org  org
	}{
		{"Transform term to org",
			term{CanonicalName: "European Union", RawID: "Nstein_GL_US_NY_Municipality_942968", Aliases: aliases{Alias: []alias{}}},
			org{UUID: "6a7edb42-c27a-3186-a0b9-7e3cdc91e16b", ProperName: "European Union", PrefLabel: "European Union", AlternativeIdentifiers: alternativeIdentifiers{
				TME: []string{"TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-T04="}, Uuids: []string{"6a7edb42-c27a-3186-a0b9-7e3cdc91e16b"}}, PrimaryType: primaryType, TypeHierarchy: orgTypes, Aliases: []string{"European Union"}}},
		{"Transform with aliases",
			term{CanonicalName: "European Union", RawID: "Nstein_GL_US_NY_Municipality_942968", Aliases: aliases{Alias: []alias{alias{Name: "EU"}, alias{Name: "EEC"}}}},
			org{UUID: "6a7edb42-c27a-3186-a0b9-7e3cdc91e16b", ProperName: "European Union", PrefLabel: "European Union", AlternativeIdentifiers: alternativeIdentifiers{
				TME: []string{"TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-T04="}, Uuids: []string{"6a7edb42-c27a-3186-a0b9-7e3cdc91e16b"}}, PrimaryType: primaryType, TypeHierarchy: orgTypes, Aliases: []string{"EU", "EEC", "European Union"}}},
	}

	for _, test := range tests {
		expectedOrg := transformOrg(test.term, "ON")

		assert.Equal(test.org, expectedOrg, fmt.Sprintf("%s: Expected org incorrect", test.name))
	}

}
