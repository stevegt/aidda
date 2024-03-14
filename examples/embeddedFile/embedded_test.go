package embedded

import (
	"io/ioutil"
	"regexp"
	"testing"

	. "github.com/stevegt/goadapt"
)

func TestLexerEmptyInput(t *testing.T) {
	// The lexer should not return any tokens when the input is empty.
	lexer := NewLexer("")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 1, "expected 1 token on empty input, got %#v", tokens)
	Tassert(t, tokens[0].Type == "EOF", "expected EOF token, got %#v", tokens[0])
}

func TestLexerNewlinesOnly(t *testing.T) {
	// The lexer should return a single token for each line in the
	// input, including empty lines.
	lexer := NewLexer("\n\n\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	for i, token := range tokens {
		if i == 3 {
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		} else {
			Tassert(t, token.Type == "Text" && token.Data == "", "expected empty Text token, got %#v", token)
		}
	}
}

func TestLexerWhitespaceOnly(t *testing.T) {
	// The lexer should return a single token for each line in the
	// input, including lines with only whitespace.
	lexer := NewLexer("  \n \n\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "Text" && token.Data == "  ", "expected '  ', got %#v", token)
		case 1:
			Tassert(t, token.Type == "Text" && token.Data == " ", "expected ' ', got %#v", token)
		case 2:
			Tassert(t, token.Type == "Text" && token.Data == "", "expected empty Text token, got %#v", token)
		case 3:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

func TestLexerTextOnly(t *testing.T) {
	// The lexer should return a single token for each line in the
	// input, including empty lines.
	lexer := NewLexer("foo\nbar\n\nbaz\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 5, "expected 5 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "Text" && token.Data == "foo", "expected 'foo', got %#v", token)
		case 1:
			Tassert(t, token.Type == "Text" && token.Data == "bar", "expected 'bar', got %#v", token)
		case 2:
			Tassert(t, token.Type == "Text" && token.Data == "", "expected empty Text token, got %#v", token)
		case 3:
			Tassert(t, token.Type == "Text" && token.Data == "baz", "expected 'baz', got %#v", token)
		case 4:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

func TestLexerTextWithWhitespace(t *testing.T) {
	// The lexer should return a single token for each line in the
	// input, including empty lines and leading/trailing whitespace.
	lexer := NewLexer("  foo\n  bar \n  baz  \n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "Text" && token.Data == "  foo", "expected '  foo', got %#v", token)
		case 1:
			Tassert(t, token.Type == "Text" && token.Data == "  bar ", "expected '  bar ', got %#v", token)
		case 2:
			Tassert(t, token.Type == "Text" && token.Data == "  baz  ", "expected '  baz  ', got %#v", token)
		case 3:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

func TestLexerTripleBacktick(t *testing.T) {
	// The lexer should return a TripleBacktick token for each set of three backticks.
	lexer := NewLexer("```\n```\n```\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	for i, token := range tokens {
		if i == 3 {
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		} else {
			Tassert(t, token.Type == "TripleBacktick", "expected TripleBacktick token, got %#v", token)
			Tassert(t, token.Data == "", "expected empty TripleBacktick token, got %#v", token)
		}
	}
}

func TestLexerNotTripleBacktick(t *testing.T) {
	// The lexer should not return a TripleBacktick token for a line
	// that does not start with three backticks.
	lexer := NewLexer("```\n ```\n````\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "TripleBacktick", "expected TripleBacktick token, got %#v", token)
			Tassert(t, token.Data == "", "expected empty TripleBacktick token, got %#v", token)
		case 1:
			Tassert(t, token.Type == "Text", "expected Text token, got %#v", token)
			Tassert(t, token.Data == " ```", "expected ' ```', got %#v", token)
		case 2:
			Tassert(t, token.Type == "Text", "expected Text token, got %#v", token)
			Tassert(t, token.Data == "````", "expected '````', got %#v", token)
		case 3:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

func TestLexerFileStart(t *testing.T) {
	// The lexer should return a FileStart token for each File block start marker
	lexer := NewLexer("File: foo\nFile: bar\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 3, "expected 3 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "FileStart", "expected FileStart token, got %#v", token)
			Tassert(t, token.Data == "foo", "expected 'foo', got %#v", token)
		case 1:
			Tassert(t, token.Type == "FileStart", "expected FileStart token, got %#v", token)
			Tassert(t, token.Data == "bar", "expected 'bar', got %#v", token)
		case 2:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

func TestLexerNotFileStart(t *testing.T) {
	// The lexer should not return a FileStart token for a line that does not start with "File: ".
	lexer := NewLexer(" File: foo\nFile: bar\nNotFile: baz\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "Text", "expected Text token, got %#v", token)
			Tassert(t, token.Data == " File: foo", "expected ' File: foo', got %#v", token)
		case 1:
			Tassert(t, token.Type == "FileStart", "expected FileStart token, got %#v", token)
			Tassert(t, token.Data == "bar", "expected 'bar', got %#v", token)
		case 2:
			Tassert(t, token.Type == "Text", "expected Text token, got %#v", token)
			Tassert(t, token.Data == "NotFile: baz", "expected 'NotFile: baz', got %#v", token)
		case 3:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

func TestLexerFileEnd(t *testing.T) {
	// The lexer should return a FileEnd token for each File block end marker
	lexer := NewLexer("EOF_foo\nEOF_bar\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 3, "expected 3 tokens, got %#v", tokens)
	for i, token := range tokens {
		switch i {
		case 0:
			Tassert(t, token.Type == "FileEnd", "expected FileEnd token, got %#v", token)
			Tassert(t, token.Data == "foo", "expected 'foo', got %#v", token)
		case 1:
			Tassert(t, token.Type == "FileEnd", "expected FileEnd token, got %#v", token)
			Tassert(t, token.Data == "bar", "expected 'bar', got %#v", token)
		case 2:
			Tassert(t, token.Type == "EOF", "expected EOF token, got %#v", token)
		}
	}
}

// TestLexerBacktracking tests the lexer's Checkpoint and Rollback methods to ensure they work as expected.
func TestLexerBacktracking(t *testing.T) {
	// The lexer should be able to backtrack and reprocess input from a certain point.
	// Each line in the input file should be one token
	lexer := NewLexer("foo\nbar\nbaz\nbing\nbong\n")
	lexer.Checkpoint()
	token := lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "foo", "expected 'foo', got %#v", token)
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "bar", "expected 'bar', got %#v", token)
	lexer.Rollback()
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "foo", "expected 'foo', got %#v", token)
	lexer.Checkpoint()
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "bar", "expected 'bar', got %#v", token)
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "baz", "expected 'baz', got %#v", token)
	lexer.Rollback()
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "bar", "expected 'bar', got %#v", token)
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "baz", "expected 'baz', got %#v", token)
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "bing", "expected 'bing', got %#v", token)
	token = lexer.Next()
	Tassert(t, token.Type == "Text" && token.Data == "bong", "expected 'bong', got %#v", token)
	token = lexer.Next()
	Tassert(t, token.Type == "EOF", "expected EOF, got %#v", token)
}

func TestLexerFunctional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md"
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}

	lexer := NewLexer(string(buf))
	// Run runs the lexer all the way through without backtracking and
	// returns a slice of tokens.
	tokens := lexer.Run()
	Tassert(t, len(tokens) > 0, "expected tokens, got %#v", tokens)

	// each line in the input file should be one token
	expectedTokens := []Token{
		{Type: "Text", Data: "test line before file"},
		{Type: "Text", Data: ""},
		{Type: "FileStart", Data: "foo"},
		{Type: "TripleBacktick", Data: ""},
		{Type: "Text", Data: "bar"},
		{Type: "TripleBacktick", Data: ""},
		{Type: "FileEnd", Data: "foo"},
		{Type: "Text", Data: "test line after eof"},
		{Type: "EOF", Data: ""},
	}

	// Ensure the lexer returns the expected tokens.
	for i, token := range tokens {
		if token.Type != expectedTokens[i].Type || token.Data != expectedTokens[i].Data {
			t.Fatalf("unexpected token %d: expected %v, got %v", i, expectedTokens[i], token)
		}
	}
}

func TestLexerMixedContent(t *testing.T) {
	// The lexer should return a mix of Text, FileStart, FileEnd, and TripleBacktick tokens.
	lexer := NewLexer("foo\nFile: bar\n```\n\nbaz\n```\nEOF_bar\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 8, "expected 8 tokens, got %#v", tokens)
	expectedTokens := []Token{
		{Type: "Text", Data: "foo"},
		{Type: "FileStart", Data: "bar"},
		{Type: "TripleBacktick", Data: ""},
		{Type: "Text", Data: ""},
		{Type: "Text", Data: "baz"},
		{Type: "TripleBacktick", Data: ""},
		{Type: "FileEnd", Data: "bar"},
		{Type: "EOF", Data: ""},
	}
	for i, token := range tokens {
		if token.Type != expectedTokens[i].Type || token.Data != expectedTokens[i].Data {
			t.Fatalf("unexpected token %d: expected %v, got %v", i, expectedTokens[i], token)
		}
	}
}

func TestLexerMissingNewline(t *testing.T) {
	// The lexer should handle input without a trailing newline.
	lexer := NewLexer("foo\nbar")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 3, "expected 3 tokens, got %#v", tokens)
	expectedTokens := []Token{
		{Type: "Text", Data: "foo"},
		{Type: "Text", Data: "bar"},
		{Type: "EOF", Data: ""},
	}
	for i, token := range tokens {
		if token.Type != expectedTokens[i].Type || token.Data != expectedTokens[i].Data {
			t.Fatalf("unexpected token %d: expected %v, got %v", i, expectedTokens[i], token)
		}
	}
}

func TestLexerBacktickLanguage(t *testing.T) {
	// The lexer should handle input with a language identifier after the opening backticks.
	lexer := NewLexer("```go\npackage main\n```\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 4, "expected 4 tokens, got %#v", tokens)
	expectedTokens := []Token{
		{Type: "TripleBacktick", Data: "go"},
		{Type: "Text", Data: "package main"},
		{Type: "TripleBacktick", Data: ""},
		{Type: "EOF", Data: ""},
	}
	for i, token := range tokens {
		if token.Type != expectedTokens[i].Type || token.Data != expectedTokens[i].Data {
			t.Fatalf("unexpected token %d: expected %v, got %v", i, expectedTokens[i], token)
		}
	}
}

func XXXTestLexerTextWithLeadingWhitespace(t *testing.T) {
	// The lexer should not remove whitespace from text lines.
	lexer := NewLexer("  foo\n  bar \n File: baz\n ```\n  \n EOF_baz\n")
	tokens := lexer.Run()
	Tassert(t, len(tokens) == 7, "expected 7 tokens, got %#v", tokens)
	expectedTokens := []Token{
		{Type: "Text", Data: "  foo"},
		{Type: "Text", Data: "  bar "},
		{Type: "Text", Data: " File: baz"},
		{Type: "Text", Data: " ```"},
		{Type: "Text", Data: "  "},
		{Type: "Text", Data: " EOF_baz"},
		{Type: "EOF", Data: ""},
	}
	for i, token := range tokens {
		if token.Type != expectedTokens[i].Type || token.Data != expectedTokens[i].Data {
			t.Fatalf("unexpected token %d: expected %v, got %v", i, expectedTokens[i], token)
		}
	}
}

// Helper function to split a byte slice into lines.  Each line
// includes a newline if the original line had one.
func bytesToLines(buf []byte) []string {
	txt := string(buf)
	// use regexp to split on \n or \r\n
	re := regexp.MustCompile(`(?ms)(^.*?(\r\n|\n))`)
	lines := re.FindAllString(txt, -1)
	return lines
}

// The parser uses the backtracking lexer and a state machine to
// process input as it encounters different tokens.
func TestParseEmptyInput(t *testing.T) {
	lex := NewLexer("")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal("should not have error on empty input")
	}
	if ast == nil || len(ast.Children) != 0 {
		t.Fatal("ast should be a non-nil root node with no children on empty input")
	}
}

// TestParseShowJSON tests the parser's ability to generate a JSON representation of the AST.
func TestParseASTString(t *testing.T) {
	lex := NewLexer("foo\nFile: bar\n```\nbaz\n```\nEOF_bar\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	j := ast.AsJSON()
	Tassert(t, j != "", "expected non-empty JSON string, got %q", j)
}

// TestParseTextOnly tests the parser's behavior when the input contains only text.
func TestParseTextOnly(t *testing.T) {
	lex := NewLexer("test line 1\ntest line 2\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a single Text child.
	children := ast.Children
	if len(children) != 1 {
		Pl(ast.AsJSON())
		t.Fatalf("expected 1 child node, got %d", len(children))
	}
	Tassert(t, children[0].Type == "Text", "expected child to be of type %q, got %q", "Text", children[0].Type)
	Tassert(t, children[0].Content == "test line 1\ntest line 2\n", "expected child content to be %q, got %q", "test line 1\ntest line 2\n", children[0].Content)
}

// TestParseFileBlock tests the parser's behavior when the input contains a single file block.
func TestParseFileBlock(t *testing.T) {
	lex := NewLexer("File: foo\n```\nbar\n```\nEOF_foo\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a single File
	// child.  The File child should have the name "foo" and the content
	// "bar\n".
	children := ast.Children
	if len(children) != 1 {
		Pl(ast.AsJSON())
		t.Fatalf("expected 1 child node, got %d", len(children))
	}
	Tassert(t, children[0].Type == "File", "expected child to be of type %q, got %q", "File", children[0].Type)
	Tassert(t, children[0].Content == "bar\n", "expected child content to be %q, got %q", "bar\n", children[0].Content)
}

// TestParseFileBlockWithLanguage tests the parser's behavior when the input contains a file block with a language identifier.
func TestParseFileBlockWithLanguage(t *testing.T) {
	lex := NewLexer("File: foo\n```go\npackage main\n```\nEOF_foo\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a single File child.
	children := ast.Children
	if len(children) != 1 {
		Pl(ast.AsJSON())
		t.Fatalf("expected 1 child node, got %d", len(children))
	}
	Tassert(t, children[0].Type == "File", "expected child to be of type %q, got %q", "File", children[0].Type)
	Tassert(t, children[0].Content == "package main\n", "expected child content to be %q, got %q", "package main\n", children[0].Content)
	Tassert(t, children[0].Language == "go", "expected child language to be %q, got %q", "go", children[0].Language)
}

// TestParseTripleBacktickOnly tests the parser's behavior when the input contains only triple backticks.
func TestParseTripleBacktickOnly(t *testing.T) {
	lex := NewLexer("```\n```\n```\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a single Text child.
	children := ast.Children
	if len(children) != 1 {
		Pl(ast.AsJSON())
		t.Fatalf("expected 1 child node, got %d", len(children))
	}
	Tassert(t, children[0].Type == "Text", "expected child to be of type %q, got %q", "Text", children[0].Type)
	Tassert(t, children[0].Content == "```\n```\n```\n", "expected child content to be %q, got %q", "```\n```\n```\n", children[0].Content)
}

/*
func TestParseIncorrectEOFMarker(t *testing.T) {
	// Expected behavior is to ignore the incorrect EOF marker,
	// including it as if it were part of the file content.

	buf, err := ioutil.ReadFile("input_incorrect_eof.md")
	Ck(err)
	Pf("buf: %q\n", buf)

	lex := NewLexer(string(buf))
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	children := ast.Children
	Tassert(t, len(children) == 2, "expected %d children nodes, got %d", 2, len(children))
	Tassert(t, children[0].Type() == "Text", "expected first child to be of type %q, got %q", "Text", children[0].Type())
	Tassert(t, children[1].Type() == "Text", "expected second child to be of type %q, got %q", "Text", children[1].Type())

	// get the input file lines
	lines := bytesToLines(buf)

	// the first child should contain the first line of the input file
	Tassert(t, children[0].Content() == lines[0], "expected first child content to be %q, got %q", lines[0], children[0].Content())

	// the second child should contain the rest of the input file
	restOfInput := strings.Join(lines[1:], "")
	Tassert(t, children[1].Content() == restOfInput, "expected second child content to be %q, got %q", restOfInput, children[1].Content())

}
*/

/*
func XXXTestParseFunctional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md" // Assuming the input file is in the test directory with the name 'input.md'.
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	Pf("buf: %q\n", buf)

	lex := NewLexer(string(buf))
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// split buf into lines
	lines := bytesToLines(buf)

	// ensure the root node is a Root type
	Tassert(t, ast.Type() == "Root", "ast type expected %q, got %q", "Root", ast.Type())

	// ensure the root node has 3 children
	children := ast.Children
	Tassert(t, len(children) == 3, "root node expected %d children, got %d", 3, len(children))

	// ensure the first child is a Text type
	Tassert(t, children[0].Type() == "Text", "first child expected %q, got %q", "Text", children[0].Type())

	// ensure the second child is a File type
	Tassert(t, children[1].Type() == "File", "second child expected %q, got %q", "File", children[1].Type())

	// ensure the third child is a Text type
	Tassert(t, children[2].Type() == "Text", "third child expected %q, got %q", "Text", children[2].Type())

	// ensure the first child content matches the first 2 lines of the input file
	firstTwoLines := lines[0] + lines[1]
	firstContent := children[0].Content()
	Tassert(t, firstContent == firstTwoLines, "first child content: expected %q, got %q", firstTwoLines, firstContent)

	// ensure the second child content matches lines 4-6 of the input file
	secondContent := children[1].Content()
	expectedContent := strings.Join(lines[3:6], "")
	Pf("expectedContent: %q\n", expectedContent)
	Tassert(t, secondContent == expectedContent, "second child content: expected %q, got %q", expectedContent, secondContent)

	// ensure the third child content matches the last line of the input file
	lastLine := lines[len(lines)-1]
	thirdContent := children[2].Content()
	Tassert(t, thirdContent == lastLine, "third child content: expected %q, got %q", lastLine, thirdContent)
}

// TestParseEmbeddedFileBlocks tests the parser's ability to handle file blocks embedded within other file blocks.
func XXXTestParseEmbeddedFileBlocks(t *testing.T) {
	buf, err := ioutil.ReadFile("input_embedded_files.md")
	Ck(err)
	Pf("buf: %q\n", buf)

	lex := NewLexer(string(buf))
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to parse the input correctly, resulting in
	// a root node with 1 File child.
	children := ast.Children
	Tassert(t, len(children) == 1, "expected %d children nodes, got %d", 1, len(children))

	// the child should be of type File
	Tassert(t, children[0].Type() == "File", "expected child to be of type %q, got %q", "File", children[0].Type())

	// the child should contain the entire input file without the
	// first and last lines, which are the File and EOF markers for the
	// outer file block
	lines := bytesToLines(buf)
	expectedContent := strings.Join(lines[1:len(lines)-1], "")
	Tassert(t, children[0].Content() == expectedContent, "expected child content to be %q, got %q", expectedContent, children[0].Content())
}
*/

/*
1. **TestParseMultipleFiles**: Verify the parser correctly handles input containing multiple file blocks.
2. **TestParseNoEOF**: Test parsing input containing a file block without an EOF marker.
4. **TestParseSpecialCharactersInContent**: Check how the parser deals with special characters or escape sequences within the text or file content.
5. **TestParseWhitespaceHandling**: Verify the parser's behavior with unusual whitespace patterns, such as leading/trailing whitespaces in file names, file content, or around EOF markers.
6. **TestParseInvalidUTF8**: Determine how the parser reacts to invalid UTF-8 sequences within the input.
7. **TestParseLongFileContent**: Test the parser's ability to handle very long file contents to verify if there are any issues with buffer sizes or memory management.
8. **TestParseSingleLineFileBlock**: Ensure that the parser correctly handles a file block defined in a single line.
9. **TestParseFileNameCollisions**: Test how the parser behaves when two file blocks have the same name but different EOF markers or content.
10. **TestParseEmptyFileContent**: Verify the behavior when a file block has no content between the start and EOF markers.
11. **TestParseCommentLines**: Include tests for parsing input with lines that should be ignored, such as comments or annotations within the text.
12. **TestParseUnexpectedEOFLocation**: Test cases where the EOF marker appears in unexpected locations, such as before the file content or at the very beginning/end of the input.
13. **TestParseRobustnessAgainstMalformedInput**: Check the parser's robustness against various forms of malformed input, including incomplete file blocks, missing file names, and abrupt endings.
14. **TestParseConcurrency**: If applicable, test the parser's behavior and correctness under concurrent execution to ensure thread safety if the Parse function is anticipated to be called from multiple goroutines.
15. **TestParseErrorHandling**: Include tests that verify the parser returns meaningful error messages or codes for various error conditions, ensuring that clients can respond appropriately to different failure modes.
*/
