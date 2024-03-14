// Package embedded provides a simple Lexer for tokenizing embedded content
// according to predefined rules.
package embedded

import (
	"strings"
)

// Lexer is responsible for tokenizing an input string based on specific
// delimiters and markers, facilitating the parsing of embedded content.
type Lexer struct {
	input       string // The input string to tokenize
	pos         int    // Current position within the input
	checkpoints []int  // Stack for managing backtracking points
}

// Token represents a lexical token with an associated type and data.
type Token struct {
	Type string // The type of the token (e.g., "Text", "FileStart", "FileEnd", "TripleBacktick")
	Data string // The data contained within the token
}

// NewLexer creates a new Lexer for tokenizing the given input string.
func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

// Checkpoint marks the current position in the input for potential backtracking.
func (l *Lexer) Checkpoint() {
	l.checkpoints = append(l.checkpoints, l.pos)
}

// Rollback reverts the current position to the last checkpoint.
func (l *Lexer) Rollback() {
	if len(l.checkpoints) > 0 {
		l.pos = l.checkpoints[len(l.checkpoints)-1]
		l.checkpoints = l.checkpoints[:len(l.checkpoints)-1]
	}
}

// Next retrieves the next token, advancing the lexer's position.
func (l *Lexer) Next() Token {
	if l.pos >= len(l.input) {
		return Token{Type: "EOF"}
	}

	// Enhanced check for newline to emitText method
	// This ensures newline handling is considered in different contexts, including empty lines.
	if l.input[l.pos] == '\n' { // Newline identified; emit as a separate text token
		l.pos++ // Move past the newline character
		return Token{Type: "Text", Data: ""}
	}

	if strings.HasPrefix(l.input[l.pos:], "File: ") {
		return l.emitWithSkip("FileStart", "File: ")
	} else if strings.HasPrefix(l.input[l.pos:], "EOF_") {
		return l.emitWithSkip("FileEnd", "EOF_")
	} else if strings.HasPrefix(l.input[l.pos:], "```") {
		return l.emitWithSkip("TripleBacktick", "```")
	}

	return l.emitText()
}

// emitWithSkip generates a token with a specified type by skipping over
// a predefined marker and collecting the subsequent text.
func (l *Lexer) emitWithSkip(tokenType, startMarker string) Token {
	startPos := l.pos + len(startMarker)
	endPos := strings.Index(l.input[startPos:], "\n")
	if endPos == -1 {
		endPos = len(l.input)
	} else {
		endPos += startPos
	}
	data := strings.TrimSpace(l.input[startPos:endPos])
	l.pos = endPos + 1 // Move past the newline
	return Token{Type: tokenType, Data: data}
}

// emitText gathers and returns a text token ending at a newline or marker.
func (l *Lexer) emitText() Token {
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '\n' && !strings.HasPrefix(l.input[l.pos:], "File: ") && !strings.HasPrefix(l.input[l.pos:], "EOF_") && !strings.HasPrefix(l.input[l.pos:], "```") {
		l.pos++
	}
	data := strings.TrimSpace(l.input[start:l.pos])
	if l.pos < len(l.input) && l.input[l.pos] == '\n' {
		l.pos++ // Move past the newline, if present
	}
	if data == "" { // Avoid emitting empty text tokens except for explicit newlines
		return l.Next() // Recur until a non-empty or meaningful token is found
	}
	return Token{Type: "Text", Data: data}
}

// Run processes the entire input and returns all identified tokens.
func (l *Lexer) Run() []Token {
	var tokens []Token
	for {
		token := l.Next()
		if token.Type == "EOF" {
			tokens = append(tokens, token) // Include EOF in the result
			break
		}
		tokens = append(tokens, token)
	}
	return tokens
}
