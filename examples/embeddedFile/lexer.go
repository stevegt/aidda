package embedded

import (
	"strings"
)

// Lexer struct will tokenize the input string.
type Lexer struct {
	input string // The input string to tokenize.
	pos   int    // Current position within the input.
}

// Token represents a lexical unit in the input stream.
type Token struct {
	Type    string // Type of the token, e.g., Text, FileStart, Newline.
	Payload string // The payload of the token, holding relevant data.
	Src     string // The source text represented by the token.
}

// NewLexer initializes a new Lexer with the provided input.
func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

// Checkpoint returns the current position in the input.
func (l *Lexer) Checkpoint() int {
	return l.pos
}

// Rollback moves the current position back to the specified checkpoint.
func (l *Lexer) Rollback(cp int) {
	l.pos = cp
}

// try attempts to produce a token using the provided function.
// If unsuccessful, it rolls back to the previous checkpoint.
func (l *Lexer) try(fn func() *Token) *Token {
	cp := l.Checkpoint()
	token := fn()
	if token == nil {
		l.Rollback(cp)
	}
	return token
}

// Methods for generating specific tokens (fileStart, fileEnd, tripleBacktick, newline, text).

func (l *Lexer) fileStart() *Token {
	if strings.HasPrefix(l.input[l.pos:], "File: ") {
		return l.emitWithSkip("FileStart", "File: ")
	}
	return nil
}

func (l *Lexer) fileEnd() *Token {
	if strings.HasPrefix(l.input[l.pos:], "EOF_") {
		return l.emitWithSkip("FileEnd", "EOF_")
	}
	return nil
}

func (l *Lexer) tripleBacktick() *Token {
	if strings.HasPrefix(l.input[l.pos:], "```") {
		return l.emitWithSkip("TripleBacktick", "```")
	}
	return nil
}

func (l *Lexer) newline() *Token {
	if l.pos < len(l.input) && l.input[l.pos] == '\n' {
		l.pos++
		return &Token{Type: "Newline", Src: "\n"}
	}
	return nil
}

func (l *Lexer) role() *Token {
	if strings.HasPrefix(l.input[l.pos:], "USER: ") {
		return l.emitWithSkip("Role", "USER: ")
	} else if strings.HasPrefix(l.input[l.pos:], "AI: ") {
		return l.emitWithSkip("Role", "AI: ")
	}
	return nil
}

func (l *Lexer) text() *Token {
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
	return &Token{Type: "Text", Src: l.input[start:l.pos]}
}

// Next produces the next token from the input.
func (l *Lexer) Next() Token {
	if l.pos >= len(l.input) {
		return Token{Type: "EOF"}
	}

	// Attempt to match token types in order of specificity.
	token := l.try(l.fileStart) 
	if token == nil { 
		token = l.try(l.fileEnd) 
	}
	if token == nil { 
		token = l.try(l.tripleBacktick) 
	}
	if token == nil { 
		token = l.try(l.newline) 
	}
	if token == nil { 
		token = l.try(l.text) 
	}
	if token == nil {
		// If no token match is found, return EOF to avoid an infinite loop.
		return Token{Type: "EOF"}
	}
	return *token
}

// emitWithSkip generates a token by skipping a given substring in the input.
func (l *Lexer) emitWithSkip(tokenType, skip string) *Token {
	end := l.pos + len(skip)
	return &Token{Type: tokenType, Payload: l.input[l.pos:end], Src: l.input[l.pos:end]}
}
