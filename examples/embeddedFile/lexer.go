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

	tripleBacktick := "```"
	/*
		backtick := "`"
		tbpat := `^` + tripleBacktick + `([^` + backtick + `]*|$)`
		tbre := regexp.MustCompile(tbpat)
	*/

	if strings.HasPrefix(l.input[l.pos:], "File: ") {
		return l.emitWithSkip("FileStart", "File: ")
	} else if strings.HasPrefix(l.input[l.pos:], "EOF_") {
		return l.emitWithSkip("FileEnd", "EOF_")
		// } else if tbre.MatchString(l.input[l.pos:]) {
	} else if strings.HasPrefix(l.input[l.pos:], "```") && (l.pos+3 == len(l.input) || l.input[l.pos+3] != '`') {
		// Enhanced handling to differentiate between opening and closing backticks
		return l.emitWithSkip("TripleBacktick", tripleBacktick)
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
	data := l.input[startPos:endPos]
	l.pos = endPos + 1 // Move past the newline
	return Token{Type: tokenType, Data: data}
}

// emitText gathers and returns a text token ending at a newline.
func (l *Lexer) emitText() Token {
	start := l.pos
	// Move to the end of the line or the end of the input
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
	data := l.input[start:l.pos]
	if l.pos < len(l.input) && l.input[l.pos] == '\n' {
		l.pos++ // Move past the newline, if present
	}
	return Token{Type: "Text", Data: data}
}

// emitCodeBlock differentiates between opening TripleBackticks possibly including a language identifier
// and simple TripleBackticks that close a code block.
func (l *Lexer) emitCodeBlock() Token {
	startPos := l.pos + 3 // Skip the backticks themselves
	l.pos = startPos

	// Find the end of the line to check for a language identifier
	endPos := strings.Index(l.input[startPos:], "\n")
	if endPos == -1 {
		l.pos = len(l.input) // Move to the end if there's no newline
		return Token{Type: "TripleBacktick", Data: ""}
	}

	data := strings.TrimSpace(l.input[startPos : startPos+endPos])
	l.pos = startPos + endPos + 1 // Move past the newline
	return Token{Type: "TripleBacktick", Data: data}
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
