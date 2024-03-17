package parser

import (
	"embedded/lexer"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"testing"

	. "github.com/stevegt/goadapt"
)

// pt (parser test) asserts that the given node's type, content, and
// children count are equal to the given values.  If all assertions
// pass, pt returns the child nodes of the given node.
func pt(t *testing.T, node *ASTNode, typ, content string, childrenCount int) (children []*ASTNode) {
	// get caller line number
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", file, line)
	pass := true
	msg := ""
	if node == nil {
		msg += "expected non-nil node\n"
		pass = false
	} else {
		if node.Type != typ {
			msg += fmt.Sprintf("expected node type %q, got %q\n", typ, node.Type)
			pass = false
		}
		if node.Content != content {
			msg += fmt.Sprintf("expected node content %q, got %q\n", content, node.Content)
			pass = false
		}
		if len(node.Children) != childrenCount {
			msg += fmt.Sprintf("expected %d children, got %d\n", childrenCount, len(node.Children))
			pass = false
		}
	}
	if !pass {
		if node != nil {
			msg += node.AsJSON(true)
		}
		// print message and fail test
		t.Logf("%s: %s", caller, msg)
		t.FailNow()
	}
	return node.Children
}

// The parser uses the backtracking lexer to process input as it
// encounters different tokens.
func TestParseEmptyInput(t *testing.T) {
	lex := lexer.NewLexer("")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal("should not have error on empty input")
	}
	rootChildren := pt(t, ast, "Root", "", 1)
	pt(t, rootChildren[0], "EOF", "", 0)
}

// TestParseShowJSON tests the parser's ability to generate a JSON representation of the AST.
func TestParseASTString(t *testing.T) {
	lex := lexer.NewLexer("foo\nFile: bar\n```\nbaz\n```\nEOF_bar\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	j := ast.AsJSON(true)
	Tassert(t, j != "", "expected non-empty JSON string, got %q", j)
}

// TestParseTextOnly tests the parser's behavior when the input contains only text.
func TestParseTextOnly(t *testing.T) {
	lex := lexer.NewLexer("test line 1\ntest line 2\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with Text and EOF
	// children.
	rootChildren := pt(t, ast, "Root", "", 2)
	pt(t, rootChildren[0], "Text", "test line 1\ntest line 2\n", 0)
}

// TestParseCodeBlockOnly tests the parser's behavior when the input contains only a code block.
func TestParseCodeBlockOnly(t *testing.T) {
	lex := lexer.NewLexer("```\nfoo\nbar\n```\n")
	parser := NewParser(lex)
	node := parser.parseCodeBlock("")
	// Expected behavior is to return a single CodeBlock node with
	// four Text children: "foo", "\n", "bar", and "\n".
	cbChildren := pt(t, node, "CodeBlock", "", 4)
	pt(t, cbChildren[0], "Text", "foo", 0)
	pt(t, cbChildren[1], "Text", "\n", 0)
	pt(t, cbChildren[2], "Text", "bar", 0)
	pt(t, cbChildren[3], "Text", "\n", 0)
}

// TestParseCodeBlockWithLanguage tests the parser's behavior when the input contains a code block with a language identifier.
func TestParseCodeBlockWithLanguage(t *testing.T) {
	lex := lexer.NewLexer("```go\npackage main\n```\n")
	parser := NewParser(lex)
	node := parser.parseCodeBlock("")
	// Expected behavior is to return a single CodeBlock node with the
	// Text children "package main" and "\n".
	cbChildren := pt(t, node, "CodeBlock", "", 2)
	pt(t, cbChildren[0], "Text", "package main", 0)
	pt(t, cbChildren[1], "Text", "\n", 0)
}

// TestParseFileBlock tests the parser's behavior when the input contains a single file block.
func TestParseFileBlock(t *testing.T) {
	lex := lexer.NewLexer("File: foo\n```\nbar\n```\nEOF_foo\n")
	ast, err := Parse(lex)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to return a root node with a File and EOF
	// children.  The File child should have the name "foo" and a
	// single Text child with the content "bar\n".
	rootChildren := pt(t, ast, "Root", "", 2)
	fileChildren := pt(t, rootChildren[0], "File", "", 1)
	pt(t, fileChildren[0], "Text", "bar\n", 0)

	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseFileBlockWithLanguage tests the parser's behavior when the input contains a file block with a language identifier.
func TestParseFileBlockWithLanguage(t *testing.T) {
	lex := lexer.NewLexer("File: foo\n```go\npackage main\n```\nEOF_foo\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with File and EOF
	// children.  The File child should have language "go" and a single
	// Text child with content "package main\n".
	rootChildren := pt(t, ast, "Root", "", 2)
	fileChildren := pt(t, rootChildren[0], "File", "", 1)
	pt(t, fileChildren[0], "Text", "package main\n", 0)

	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseTripleBacktickOnly tests the parser's behavior when the input contains only triple backticks.
func TestParseTripleBacktickOnly(t *testing.T) {
	lex := lexer.NewLexer("```\n```\n```\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with a single
	// empty CodeBlock child followed by a Text child with the content
	// "```\n" and an EOF.
	rootChildren := pt(t, ast, "Root", "", 3)
	pt(t, rootChildren[0], "CodeBlock", "", 0)
	pt(t, rootChildren[1], "Text", "```\n", 0)
	pt(t, rootChildren[2], "EOF", "", 0)
}

// TestParseBacktracking tests the parser's ability to backtrack and reprocess input from a certain point.
func TestParseBacktracking(t *testing.T) {
	input := "File: bar\n```\nbaz\ntrailing text\n"
	lex := lexer.NewLexer(input)
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with 2 children: a
	// Text node and an EOF node.  The Text node should have the
	// content "File: bar\n```\nbaz\ntrailing text\n". The parser will
	// checkpoint the lexer, then start a File node when it sees the
	// File: line, then it will hit end of input without finding an
	// EOF_ token, at which point it will rollback the lexer and
	// re-parse the File: line and everything after it as a Text node.
	// In order for this to work, the parser must checkpoint the lexer
	// before it starts parsing any new node, and it must rollback the
	// lexer if it encounters an error while parsing a node.  The
	// parser must have some sense of node priority, trying more
	// complex nodes like File before simpler nodes like Text.
	rootChildren := pt(t, ast, "Root", "", 2)
	pt(t, rootChildren[0], "Text", input, 0)
	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseNoEOF tests the parser's behavior when the input contains a file block without an EOF marker.
func TestParseNoEOF(t *testing.T) {
	lex := lexer.NewLexer("File: foo\n```\nbar\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with a single Text
	// child and an EOF child.  The Text child should have the content
	// "File: foo\n```\nbar\n".
	rootChildren := pt(t, ast, "Root", "", 2)
	pt(t, rootChildren[0], "Text", "File: foo\n```\nbar\n", 0)
	pt(t, rootChildren[1], "EOF", "", 0)
}

func TestParseIncorrectEOFMarker(t *testing.T) {
	lex := lexer.NewLexer("File: foo\n```\nbar\n```\nEOF_baz\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to ignore the incorrect EOF marker,
	// treat it and the file start marker as text, with a code block
	// in between.
	rootChildren := pt(t, ast, "Root", "", 4)
	pt(t, rootChildren[0], "Text", "File: foo\n", 0)
	cbChildren := pt(t, rootChildren[1], "CodeBlock", "", 1)
	pt(t, cbChildren[0], "Text", "bar\n", 0)
	pt(t, rootChildren[2], "Text", "EOF_baz\n", 0)
	pt(t, rootChildren[3], "EOF", "", 0)
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

func TestParseFunctional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md" // Assuming the input file is in the test directory with the name 'input.md'.
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("buf: %q\n", buf)
	lex := lexer.NewLexer(string(buf))
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	t.Log(ast.AsJSON(true))

	// split buf into lines
	lines := bytesToLines(buf)

	// test line before file
	//
	// File: foo
	// ```
	// bar
	// ```
	// EOF_foo
	// test line after eof

	rootChildren := pt(t, ast, "Root", "", 4)
	pt(t, rootChildren[0], "Text", lines[0]+lines[1], 0)
	fileChildren := pt(t, rootChildren[1], "File", "", 1)
	pt(t, fileChildren[0], "Text", lines[4], 0)
	pt(t, rootChildren[2], "Text", lines[7], 0)
	pt(t, rootChildren[3], "EOF", "", 0)

}

// TestParseEmbeddedFileBlocks tests the parser's ability to handle file blocks embedded within other file blocks.
func TestParseEmbeddedFileBlocks(t *testing.T) {
	buf, err := ioutil.ReadFile("input_embedded_files.md")
	Ck(err)
	t.Logf("buf: %q\n", buf)
	lex := lexer.NewLexer(string(buf))
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)

	// File: outer_file.md
	// ```
	// Some content
	// File: inner_file.md
	// ```
	// inner_file.md content
	// ```
	// EOF_inner_file.md
	// ```
	// EOF_outer_file.md

	// Expected behavior is to parse the input correctly, resulting in
	// a root node with 1 File child and an EOF child.  The File child
	// should have the name "foo" and a single Text child with the
	// content "Some content\nFile: inner_file.md\n```\ninner_file.md content\n```\nEOF_inner_file.md\n".
	rootChildren := pt(t, ast, "Root", "", 2)
	fileChildren := pt(t, rootChildren[0], "File", "", 1)
	pt(t, fileChildren[0], "Text", "Some content\nFile: inner_file.md\n```\ninner_file.md content\n```\nEOF_inner_file.md\n", 0)
	pt(t, rootChildren[1], "EOF", "", 0)
}

// TestParseRole tests the parser's ability to handle USER: and AI:
// role sections in an LLM chat log. Nested inside the role sections
// are other nodes like File, CodeBlock, or Text.
func TestParseRole(t *testing.T) {
	lex := lexer.NewLexer("USER: bong\nbaz\nAI: bar\nUSER:\nbing\nFile: foo\n```\nboom\n```\nEOF_foo\nAI:\nok\n")
	ast, err := Parse(lex)
	Tassert(t, err == nil, "expected no error, got %v", err)
	// Expected behavior is to return a root node with 5 children:
	// Four alternating USER and AI Role nodes, and an EOF node.  The
	// USER and AI Role nodes should each have Text and/or File
	// children.
	rootChildren := pt(t, ast, "Root", "", 5)
	user1Children := pt(t, rootChildren[0], "Role", "USER: ", 1)
	pt(t, user1Children[0], "Text", "bong\nbaz\n", 0)
	ai1Children := pt(t, rootChildren[1], "Role", "AI: ", 1)
	pt(t, ai1Children[0], "Text", "bar\n", 0)
	user2Children := pt(t, rootChildren[2], "Role", "USER:", 2)
	pt(t, user2Children[0], "Text", "bing\n", 0)
	fileChildren := pt(t, user2Children[1], "File", "", 1)
	pt(t, fileChildren[0], "Text", "boom\n", 0)
	ai2Children := pt(t, rootChildren[3], "Role", "AI:", 1)
	pt(t, ai2Children[0], "Text", "ok\n", 0)
	pt(t, rootChildren[4], "EOF", "", 0)
}
