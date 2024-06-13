package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseActions(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		valid       []Action
		expectError bool
		expected    []Action
	}{
		{
			name:     "Valid single action",
			response: "push origin main",
			valid: []Action{
				{Name: "push", Args: []string{"origin", "main"}},
			},
			expectError: false,
			expected: []Action{
				{Name: "push", Args: []string{"origin", "main"}},
			},
		},
		{
			name:     "Invalid action",
			response: "update origin main",
			valid: []Action{
				{Name: "push", Args: []string{"origin", "main"}},
				{Name: "fetch", Args: []string{"origin"}},
			},
			expectError: true,
		},
		{
			name:     "Multiple actions",
			response: "push origin main\nfetch origin",
			valid: []Action{
				{Name: "push", Args: []string{"origin", "main"}},
				{Name: "fetch", Args: []string{"origin"}},
			},
			expectError: false,
			expected: []Action{
				{Name: "push", Args: []string{"origin", "main"}},
				{Name: "fetch", Args: []string{"origin"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // Corrected function declaration
			actions, err := parseActions(tt.response, tt.valid)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actions)
			}
		})
	}
}

func TestParseArguments(t *testing.T) {
	tests := []struct {
		name     string
		fields   []string
		expected []string
	}{
		{
			name:     "Simple arguments",
			fields:   []string{"arg1", "arg2"},
			expected: []string{"arg1", "arg2"},
		},
		{
			name:     "Arguments with spaces",
			fields:   []string{"'arg 1'", "\"arg 2\""},
			expected: []string{"arg 1", "arg 2"},
		},
		{
			name:     "Mixed arguments",
			fields:   []string{"arg1", "'arg 2'", "arg3"},
			expected: []string{"arg1", "arg 2", "arg3"},
		},
		{
			name:     "Quoted argument with embedded quote",
			fields:   []string{"\"arg'2\"", "'arg\"3'"},
			expected: []string{"arg'2", "arg\"3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArguments(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}
