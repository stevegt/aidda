package embedded

import (
	"regexp"
)

// Node represents a generic node in the tree.
type Node interface {
	Type() string
	Content() string
}

// Text represents a text node.
type Text struct {
	Data string
}

func (t *Text) Type() string {
	return "Text"
}

func (t *Text) Content() string {
	return t.Data
}

// File represents a file node.
type File struct {
	Name string
	Data string
}

func (f *File) Type() string {
	return "File"
}

func (f *File) Content() string {
	return f.Data
}

// Root represents the root node of the tree.
type Root struct {
	Data     string
	children []Node
}

func (r *Root) Type() string {
	return "Root"
}

func (r *Root) Content() string {
	return r.Data
}

func (r *Root) Children() []Node {
	return r.children
}

// Parse function to parse the input byte slice into a tree structure.
func Parse(input []byte) (*Root, error) {
	if len(input) == 0 {
		// Allow empty input to return a new Root with no children.
		return &Root{Data: "root"}, nil
	}

	var children []Node

	backticks := "```"
	fileBlockRegexString := `(?ms)^File: ([^\n]+)\n` + backticks + `(.*?)` + backticks + `\nEOF_([^\n]+)\n?`
	fileBlockRegex := regexp.MustCompile(fileBlockRegexString)

	matches := fileBlockRegex.FindAllSubmatchIndex(input, -1)
	lastIndex := 0
	for _, match := range matches {
		preMatchText := string(input[lastIndex:match[0]])
		if len(preMatchText) > 0 {
			children = append(children, &Text{Data: preMatchText})
		}

		fileName := string(input[match[2]:match[3]])
		fileContent := string(input[match[4]:match[5]])
		eofName := string(input[match[6]:match[7]])

		if fileName != eofName {
			continue // EOF marker does not match file name; skip this block.
		}

		children = append(children, &File{Name: fileName, Data: fileContent})

		// Update lastIndex to just after the EOF marker.
		lastIndex = match[7] + len("\n")
	}

	// Append any trailing text after the last file block as a Text node.
	if lastIndex <= len(input) {
		trailingText := string(input[lastIndex:])
		if len(trailingText) > 0 {
			children = append(children, &Text{Data: trailingText})
		}
	}

	root := &Root{
		Data:     "root",
		children: children,
	}

	return root, nil
}
