package embedded

import (
	"errors"
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
		return nil, errors.New("input cannot be empty")
	}

	var children []Node

	backticks := "```"
	fileBlockRegexString := `(?m)^File: ([^\n]+)\n` + backticks + `([\s\S]*?)` + backticks + `\nEOF_([^\n]+)`
	fileBlockRegex := regexp.MustCompile(fileBlockRegexString)

	matches := fileBlockRegex.FindAllSubmatchIndex(input, -1)

	lastIndex := 0
	for _, match := range matches {
		fileName := string(input[match[2]:match[3]])
		fileContent := string(input[match[4]:match[5]])
		eofName := string(input[match[6]:match[7]])

		// Validate that the EOF name matches the file name before adding the node.
		if fileName != eofName {
			continue // Skip this match as the EOF marker does not match the file name.
		}

		// Include text before the first file block as a Text node.
		if lastIndex < match[0] {
			textData := string(input[lastIndex:match[0]])
			trimmedTextData := textData[:len(textData)-1] // Trim the trailing newline before the file block starts.
			children = append(children, &Text{Data: trimmedTextData})
		}

		children = append(children, &File{Name: fileName, Data: fileContent})

		lastIndex = match[1] + 1 // Increment lastIndex to move past the newline following the EOF marker.
	}

	// Append any trailing text after the last match as a Text node.
	if lastIndex < len(input) {
		children = append(children, &Text{Data: string(input[lastIndex:])})
	}

	root := &Root{
		Data:     "root",
		children: children,
	}

	return root, nil
}
