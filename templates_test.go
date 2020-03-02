package main

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_formatContainerImage(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "non ecr image",
			input:    "jboss/keycloak:7.0.0",
			expected: "jboss/keycloak:7.0.0",
		},
		{
			desc:     "correct ecr image",
			input:    "00000000000.dkr.ecr.ap-southeast-2.amazonaws.com/foo-bar:0.5.0",
			expected: "foo-bar:0.5.0",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			actual := formatContainerImage(c.input)
			assert.Equal(t, template.HTML(c.expected), actual)
		})
	}
}
