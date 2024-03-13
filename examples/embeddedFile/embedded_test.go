package embedded

import (
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	. "github.com/stevegt/goadapt"
)

// Helper function to split a byte slice into lines.  Each line
// includes a newline if the original line had one.
func bytesToLines(buf []byte) []string {
	txt := string(buf)
	// use regexp to split on \n or \r\n
	re := regexp.MustCompile(`(?ms)(^.*?(\r\n|\n))`)
	lines := re.FindAllString(txt, -1)
	return lines
}

func TestParse_EmptyInput(t *testing.T) {
	ast, err := Parse([]byte(""))
	if err != nil {
		t.Fatal("should not have error on empty input")
	}
	if ast == nil || len(ast.Children()) != 0 {
		t.Fatal("ast should be a non-nil root node with no children on empty input")
	}

}

func TestParse_Functional(t *testing.T) {
	// Functional test reading input from file
	fn := "input.md" // Assuming the input file is in the test directory with the name 'input.md'.
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	ast, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// split buf into lines
	lines := bytesToLines(buf)

	// ensure the root node is a Root type
	Tassert(t, ast.Type() == "Root", "ast type expected %q, got %q", "Root", ast.Type())

	// ensure the root node has 3 children
	children := ast.Children()
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

	// ensure the third child content matches the last line of the input file
	lastLine := lines[len(lines)-1]
	thirdContent := children[2].Content()
	Tassert(t, thirdContent == lastLine, "third child content: expected %q, got %q", lastLine, thirdContent)
}

func TestParse_IncorrectEOFMarker(t *testing.T) {
	buf, err := ioutil.ReadFile("input_incorrect_eof.md")
	Ck(err)
	ast, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}
	Tassert(t, ast != nil)

	// Expected behavior is to ignore the block with incorrect EOF marker,
	// resulting in two children: one Text node before the incorrect file block and
	// one Text that includes the incorrect EOF marker and trailing text.

	children := ast.Children()
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
