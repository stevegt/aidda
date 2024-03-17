package embedded

import (
	"strings"
)

// Lexer splits input text into tokens.
type Lexer struct {
	input string
	pos   int
}

// Token represents a lexical unit.
type Token struct {
	Type    string
	Payload string
	Src     string
}

// NewLexer constructs a new Lexer.
func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

// Checkpoint marks the current position.
func (l *Lexer) Checkpoint() int {
	return l.pos
}

// Rollback moves to a saved position.
func (l *Lexer) Rollback(cp int) {
	l.pos = cp
}

// Next returns the next token from the input.
func (l *Lexer) Next() Token {
	if l.pos >= len(l.input) { 
		return Token{Type: "EOF"} 
	}

	var token *Token
	methods := []func() *Token{l.role, l.fileStart, l.fileEnd, l.tripleBacktick, l.newline, l.text}
	for _, method := range methods {
		if token = l.try(method); token != nil {
			break
		}
	}
	return *token
}

func (l *Lexer) try(fn func() *Token) *Token {
	cp := l.Checkpoint()
	token := fn()
	if token == nil {
		l.Rollback(cp)
	}
	return token
}

func (l *Lexer) fileStart() *Token {
	if strings.HasPrefix(l.input[l.pos:], "File: ") {
		return l.emitWithSkip("FileStart", 6) // Skip "File: " (6 characters)
	}
	return nil
}

func (l *Lexer) fileEnd() *Token {
	if strings.HasPrefix(l.input[l.pos:], "EOF_") {
		return l.emitWithSkip("FileEnd", 4) // Skip "EOF_" (4 characters)
	}
	return nil
}

func (l *Lexer) tripleBacktick() *Token {
	if strings.HasPrefix(l.input[l.pos:], "```") {
		end := strings.Index(l.input[l.pos+3:], "```")
		if end == -1 {
			return &Token{Type: "TripleBacktick", Payload: "", Src: l.input[l.pos : l.pos+3]}
		} else {
			language := l.input[l.pos+3 : l.pos+3+end]
			l.pos += 3 // Skip opening backticks, leaving position at start of content or language tag
			return &Token{Type: "TripleBacktick", Payload: language, Src: "```" + language}
		}
	}
	return nil
}

func (l *Lexer) role() *Token {
	if strings.HasPrefix(l.input[l.pos:], "USER: ") || strings.HasPrefix(l.input[l.pos:], "AI: ") {
		prefix := l.input[l.pos : l.pos+5]
		l.pos += 5 // Skip prefix
		start := l.pos
		for l.pos < len(l.input) && l.input[l.pos] != '\n' {
			l.pos++
		}
		return &Token{Type: "Role", Payload: prefix[:len(prefix)-2], Src: prefix + l.input[start:l.pos]}
	}
	return nil
}

func (l *Lexer) newline() *Token {
	if l.input[l.pos] == '\n' {
		l.pos++
		return &Token{Type: "Newline", Src: "\n"}
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

// emitWithSkip handles the common pattern of emitting tokens that start with a certain prefix and end at the next newline.
func (l *Lexer) emitWithSkip(tokenType string, skipLen int) *Token {
	start := l.pos + skipLen
	end := start
	for end < len(l.input) && l.input[end] != '\n' {
		end++
	}
	token := &Token{
		Type:    tokenType,
		Payload: strings.TrimSpace(l.input[start:end]),
		Src:     l.input[l.pos:end],
	}
	l.pos = end
	return token
}
