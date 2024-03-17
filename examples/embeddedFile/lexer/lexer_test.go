package lexer

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/stevegt/goadapt"
)

// lt (lexer test) asserts that the next token's type and data are
// equal to the given values.
func lt(t *testing.T, lexer *Lexer, typ, src, payload string) {
	token := lexer.Next()
	pass := true
	msg := ""
	if token.Type != typ {
		msg += fmt.Sprintf("expected token type %q, got %q\n", typ, token.Type)
		pass = false
	}
	if token.Src != src {
		msg += fmt.Sprintf("expected token src %q, got %q\n", src, token.Src)
		pass = false
	}
	if token.Payload != payload {
		msg += fmt.Sprintf("expected token payload %q, got %q\n", payload, token.Payload)
		pass = false
	}
	if !pass {
		msg += Spf("token: %#v", token)
		t.Fatal(msg)
	}
}

func TestLexerEmptyInput(t *testing.T) {
	// The lexer should return an EOF token when the input is empty.
	lexer := NewLexer("")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerNewlines(t *testing.T) {
	// The lexer should return a newline token for each empty line in the input.
	lexer := NewLexer("\n\n\n")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerWhitespace(t *testing.T) {
	// The lexer should return a two tokens for each non-empty line in
	// the input, including lines with only whitespace.  The first
	// token may be a Text token with the line's content if any, and the
	// second token should be a Newline token.
	lexer := NewLexer("  \n \n\n")
	lt(t, lexer, "Text", "  ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", " ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
}

func TestLexerText(t *testing.T) {
	// The lexer should return a text token (if not empty) and a
	// newline token for for each line in the input.
	lexer := NewLexer("foo\nbar\n\nbaz\n")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerTextWithWhitespace(t *testing.T) {
	// The lexer should return a text and newline token for each line in the
	// input, including leading/trailing whitespace.
	lexer := NewLexer("  foo\n  bar \n  baz  \n")
	lt(t, lexer, "Text", "  foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "  bar ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "  baz  ", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerTripleBacktick(t *testing.T) {
	// The lexer should return TripleBacktick and Newline tokens for each
	// line in the input that starts with three backticks, and
	// an EOF token at the end of the input.
	lexer := NewLexer("```\n```\n```\n")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerNotTripleBacktick(t *testing.T) {
	// The lexer should not return a TripleBacktick token for a line
	// that does not start with three backticks.
	lexer := NewLexer("```\n ```\n````\n")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", " ```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "````", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerFileStart(t *testing.T) {
	// The lexer should return a FileStart token for each File block start marker
	lexer := NewLexer("File: foo\nFile: bar\n")
	lt(t, lexer, "FileStart", "File: foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerNotFileStart(t *testing.T) {
	// The lexer should not return a FileStart token for a line that does not start with "File: ".
	lexer := NewLexer(" File: foo\nFile: bar\nNotFile: baz\n")
	lt(t, lexer, "Text", " File: foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "NotFile: baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerFileEnd(t *testing.T) {
	// The lexer should return a FileEnd token for each File block end marker
	lexer := NewLexer("EOF_foo\nEOF_bar\n")
	lt(t, lexer, "FileEnd", "EOF_foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileEnd", "EOF_bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

// TestLexerBacktracking tests the lexer's Checkpoint and Rollback methods to ensure they work as expected.
func TestLexerBacktracking(t *testing.T) {
	// The lexer should be able to backtrack and reprocess input from a certain point.
	// Each line in the input file should be one token
	lexer := NewLexer("foo\nbar\nbaz\nbing\nbong\n")
	cp := lexer.Checkpoint()
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lexer.Rollback(cp)
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	cp = lexer.Checkpoint()
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lexer.Rollback(cp)
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bing", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bong", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerMissingNewline(t *testing.T) {
	// The lexer should handle input without a trailing newline.
	lexer := NewLexer("foo\nbar")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerFunctional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md"
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}

	lexer := NewLexer(string(buf))
	lt(t, lexer, "Text", "test line before file", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileEnd", "EOF_foo", "foo")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "test line after eof", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerMixedContent(t *testing.T) {
	// The lexer should return a mix of Text, FileStart, FileEnd, and TripleBacktick tokens.
	lexer := NewLexer("foo\nFile: bar\n```\n\nbaz\n```\nEOF_bar\n")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileStart", "File: bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "FileEnd", "EOF_bar", "bar")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

func TestLexerBacktickLanguage(t *testing.T) {
	// The lexer should handle input with a language identifier after the opening backticks.
	lexer := NewLexer("```go\npackage main\n```\n")
	lt(t, lexer, "TripleBacktick", "```go", "go")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "package main", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "TripleBacktick", "```", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "EOF", "", "")
}

// Role tokens signify the start of a USER: or AI: section in an LLM
// chat log.  The lexer should return a Role token for each line in
// the input that starts with "USER: " or "AI: ", with the role name
// as the token's payload. Any text on the same line after the "USER: "
// or "AI: " should be returned as a Text token.
func TestLexerUserAIStart(t *testing.T) {
	lexer := NewLexer("USER: foo\nbaz\nAI: bar\nUSER:\nbing\n USER: baz")
	lt(t, lexer, "Role", "USER: ", "USER")
	lt(t, lexer, "Text", "foo", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "baz", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Role", "AI: ", "AI")
	lt(t, lexer, "Text", "bar", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Role", "USER:", "USER")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", "bing", "")
	lt(t, lexer, "Newline", "\n", "")
	lt(t, lexer, "Text", " USER: baz", "")
	lt(t, lexer, "EOF", "", "")
}
