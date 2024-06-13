package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanUserQuery(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "With comments",
			input:  "# This is a comment\nquery\n# Another comment",
			expect: "query",
		},
		{
			name:   "Without comments",
			input:  "query\nquery2",
			expect: "query\nquery2",
		},
		{
			name:   "Mixed spaces and tabs in comments",
			input:  "#This is a comment\n\t#Indented comment\nquery",
			expect: "query",
		},
		{
			name:   "Only comments",
			input:  "# Just a comment\n#Another one",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanUserQuery(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestFormatValidActions(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]string
		expect string
	}{
		{
			name: "Multiple actions",
			input: map[string]string{
				"queryGopls":         "queryGopls args...",
				"fetchLinesFromFile": "fetchLinesFromFile path startLine endLine",
			},
			expect: "Valid actions:\nqueryGopls: queryGopls args...\nfetchLinesFromFile: fetchLinesFromFile path startLine endLine\n",
		},
		{
			name: "Single action",
			input: map[string]string{
				"runTests": "runTests packagePath",
			},
			expect: "Valid actions:\nrunTests: runTests packagePath\n",
		},
		{
			name:   "No actions",
			input:  map[string]string{},
			expect: "Valid actions:\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValidActions(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestParseActions(t *testing.T) {
	tests := []struct {
		name     string
		response string
		valid    map[string]string
		expect   []Action
	}{
		{
			name: "Valid single action",
			response: "runTests ./...\n",
			valid: map[string]string{
				"runTests": "Run go tests on package",
			},
			expect: []Action{
				{Name: "runTests", Args: []string{"./..."}},
			},
		},
		{
			name: "Multiple actions",
			response: "runTests ./pkg1\nqueryGopls symbol definition\n",
			valid: map[string]string{
				"runTests":  "Run go tests on package",
				"queryGopls": "Query gopls server",
			},
			expect: []Action{
				{Name: "runTests", Args: []string{"./pkg1"}},
				{Name: "queryGopls", Args: []string{"symbol", "definition"}},
			},
		},
		{
			name:     "No valid actions",
			response: "nonexistentAction param1\n",
			valid: map[string]string{
				"runTests":  "Run go tests on package",
			},
			expect: []Action{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseActions(tt.response, tt.valid)
			assert.Equal(t, tt.expect, result)
		})
	}
}
