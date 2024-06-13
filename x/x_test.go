package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for cleanUserQuery
func TestCleanUserQuery_QueriesWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "Query with special characters",
			input:  "SELECT * FROM users WHERE name='John #Doe'; #comment",
			expect: "SELECT * FROM users WHERE name='John #Doe';",
		},
		{
			name:   "Complex query spread across multiple lines",
			input:  "SELECT *\n# This is a comment\nFROM users;\n# Another comment\nWHERE id > 10;",
			expect: "SELECT *\nFROM users;\nWHERE id > 10;",
		},
		{
			name:   "Query ending with comment",
			input:  "SELECT * FROM users; #final comment",
			expect: "SELECT * FROM users;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanUserQuery(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// Tests for formatValidActions
func TestFormatValidActions_WithSortedOutput(t *testing.T) {
	actions := map[string]string{
		"zAction": "Last action",
		"aAction": "First action",
		"mAction": "Middle action",
	}
	expected := "Valid actions:\naAction: First action\nmAction: Middle action\nzAction: Last 	action\n"
	result := formatValidActions(actions)
	assert.Equal(t, expected, result)
}

// Tests for parseActions
func TestParseActions_WithInvalidActions(t *testing.T) {
	response := "invalidAction param1\nvalidAction param2 param3\n"
	valid := map[string]string{
		"validAction": "A valid action",
	}
	expected := []Action{
		{Name: "validAction", Args: []string{"param2", "param3"}},
	}
	result := parseActions(response, valid)
	assert.Equal(t, expected, result)
}

func TestParseActions_WithSpecialCharacters(t *testing.T) {
	response := "specialAction 'param with spaces' \"another param\"\n"
	valid := map[string]string{
		"specialAction": "Handles special character cases",
	}
	expected := []Action{
		{Name: "specialAction", Args: []string{"'param with spaces'", "\"another param\""}},
	}
	result := parseActions(response, valid)
	assert.Equal(t, expected, result)
}
