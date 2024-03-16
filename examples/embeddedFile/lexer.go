package embedded

import (
	"strings"
)

// Lexer is responsible for tokenizing an input string based on specific
// delimiters and markers, facilitating the parsing of embedded content.
type Lexer struct {
	input string // The input string to tokenize
	pos   int    // Current position within the input
}

// Token represents a lexical token with an associated type and data.
type Token struct {
	Type    string // The type of the token (e.g., "Text", "FileStart", "FileEnd", "TripleBacktick")
	Payload string // The data contained within the token
	Src     string // The original source text for the token
}

// NewLexer creates a new Lexer for tokenizing the given input string.
func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

// Checkpoint returns the current position in the input for potential backtracking.
func (l *Lexer) Checkpoint() int {
	return l.pos
}

// Rollback reverts the current position to the given position.
func (l *Lexer) Rollback(cp int) {
	l.pos = cp
}

// try is a helper function that attempts to generate a token given a
// matching function. If the matching function returns a token, it is
// returned; otherwise, the lexer's position is rolled back and nil is
// returned.
func (l *Lexer) try(fn func() *Token) *Token {
	cp := l.Checkpoint()
	token := fn()
	if token != nil {
		return token
	}
	l.Rollback(cp)
	return nil
}

// Next retrieves the next token, advancing the lexer's position.
func (l *Lexer) Next() Token {
	if l.pos >= len(l.input) {
		return Token{Type: "EOF"}
	}

	// tripleBacktick := "```"
	/*
		backtick := "`"
		tbpat := `^` + tripleBacktick + `([^` + backtick + `]*|$)`
		tbre := regexp.MustCompile(tbpat)
	*/

	// recursive descent lexer
	var token *Token
	for {
		token = l.try(l.fileStart)
		if token != nil {
			break
		}
		token = l.try(l.fileEnd)
		if token != nil {
			break
		}
		token = l.try(l.tripleBacktick)
		if token != nil {
			break
		}
		token = l.try(l.newline)
		if token != nil {
			break
		}
		token = l.text()
		break
	}
	return *token
}

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
	if strings.HasPrefix(l.input[l.pos:], "```") && (l.pos+3 == len(l.input) || l.input[l.pos+3] != '`') {
		return l.emitWithSkip("TripleBacktick", "```")
	}
	return nil
}

func (l *Lexer) newline() *Token {
	if strings.HasPrefix(l.input[l.pos:], "\r\n") {
		l.pos += 2
		return &Token{Type: "Newline", Src: "\r\n"}
	}
	if strings.HasPrefix(l.input[l.pos:], "\n") {
		l.pos++
		return &Token{Type: "Newline", Src: "\n"}
	}
	return nil
}

func (l *Lexer) text() *Token {
	start := l.pos
	end := strings.IndexAny(l.input[start:], "\r\n")
	if end == -1 {
		end = len(l.input)
	} else {
		end += start
	}
	src := l.input[start:end]
	l.pos = end
	return &Token{Type: "Text", Src: src}
}

// emitWithSkip generates a token with a specified type by skipping over
// a predefined marker and collecting the subsequent text.
func (l *Lexer) emitWithSkip(tokenType, startMarker string) *Token {
	payloadStart := l.pos + len(startMarker)
	endPos := strings.IndexAny(l.input[payloadStart:], "\r\n")
	if endPos == -1 {
		endPos = len(l.input)
	} else {
		endPos += payloadStart
	}
	src := l.input[l.pos:endPos]
	payload := l.input[payloadStart:endPos]
	l.pos = endPos
	return &Token{Type: tokenType, Payload: payload, Src: src}
}
