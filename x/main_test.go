package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanUserQuery(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     string
	}{
		{
			name: "Remove comments",
			input: "SELECT * FROM users # This is a comment\n" +
				"INSERT INTO 'test' VALUES ('1', '2') # Another comment",
			expected: "SELECT * FROM users\n" +
				"INSERT INTO 'test' VALUES ('1', '2')",
		},
		{
			name: "Handle inline comments",
			input: "SELECT * FROM users --ostensible 'comment' in string\n" +
				"INSERT INTO users VALUES ('some#text', 'another#text') -- Comment",
			expected: "SELECT * FROM users --ostensible 'comment' in string\n" +
				"INSERT INTO users VALUES ('some#text', 'another#text')",
		},
		{
			name: "Preserve new lines",
			input: "# Full line comment\n\n" +
				"SELECT 'Something' as `value` -- inline comment\n\n" +
				"# Another comment",
			expected: "\nSELECT 'Something' as `value`\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanUserQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValidActions(t *testing.T) {
    actions := []Action{
        {Name: "push", Args: []string{"origin", "main"}},
        {Name: "fetch", Args: []string{"origin"}},
    }
    expected := "Valid actions:\npush: origin main\nfetch: origin"
    result := formatValidActions(actions)
    assert.Equal(t, expected, result)
}
