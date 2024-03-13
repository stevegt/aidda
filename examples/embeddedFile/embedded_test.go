package embedded

import (
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/stevegt/goadapt"
)

func TestParse(t *testing.T) {
	fn := "input.md"
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
	Tassert(t, ast.Type() == "Root", "ast is not a Root type")

	// ensure the root node has 3 children
	children := ast.Children()
	Tassert(t, len(children) == 3)

	// ensure the first child is a Text type
	Tassert(t, children[0].Type() == "Text", "first child is not a Text type")

	// ensure the second child is a File type
	Tassert(t, children[1].Type() == "File", "second child is not a File type")

	// ensure the third child is a Text type
	Tassert(t, children[2].Type() == "Text", "third child is not a Text type")

	// ensure the first child content matches the first 2 lines of the input file
	firstTwoLines := lines[0] + "\n" + lines[1]
	firstContent := children[0].Content()
	Tassert(t, firstContent == firstTwoLines, "first child content does not match first 2 lines of input file")

	// ensure the third child content matches the last two lines of the input file
	lastLines := lines[len(lines)-2] + "\n" + lines[len(lines)-1]
	thirdContent := children[2].Content()
	Tassert(t, thirdContent == lastLines, Spf("third child content: expected %s, got %s", lastLines, thirdContent))
}

// bytesToLines splits a byte slice into lines
func bytesToLines(buf []byte) []string {
	txt := string(buf)
	lines := strings.Split(txt, "\n")
	return lines
}
