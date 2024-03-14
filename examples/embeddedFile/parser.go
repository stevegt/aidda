package embedded

import (
	"encoding/json"
	"strings"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer  *Lexer  // Lexer instance
	tokens []Token // Token buffer
	pos    int     // Current position in token buffer
}

// NewParser initializes a new parser from a lexer.
func NewParser(lexer *Lexer) *Parser {
	return &Parser{
		lexer:  lexer,
		tokens: make([]Token, 0),
	}
}

// ASTNode represents a node in the abstract syntax tree.
type ASTNode struct {
	Type     string     `json:"Type"`
	Content  string     `json:"Content,omitempty"`
	Name     string     `json:"Name,omitempty"`
	Language string     `json:"Language,omitempty"`
	Children []*ASTNode `json:"Children,omitempty"`
}

// next retrieves the next token from the lexer or the buffer.
func (p *Parser) next() Token {
	if p.pos >= len(p.tokens) {
		p.tokens = append(p.tokens, p.lexer.Next())
	}

	token := p.tokens[p.pos]
	p.pos++
	return token
}

// peek looks at the next token without consuming it.
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		p.tokens = append(p.tokens, p.lexer.Next())
	}
	return p.tokens[p.pos]
}

// parseText gathers consecutive text tokens until another token type is encountered.
func (p *Parser) parseText() *ASTNode {
	text := strings.Builder{}

	for {
		token := p.peek()
		if token.Type != "Text" && token.Type != "TripleBacktick" {
			break
		}

		p.next() // Consume the token

		if token.Type == "TripleBacktick" {
			text.WriteString("```\n") // Retain backticks and newline for accurate representation
		} else {
			text.WriteString(token.Data)
			if p.peek().Type == "Text" {
				text.WriteString("\n") // Separate text tokens with newlines except before a non-Text token
			}
		}
	}

	if text.Len() == 0 {
		return nil // No text collected
	}

	return &ASTNode{
		Type:    "Text",
		Content: text.String(),
	}
}

// parseFile handles file blocks, including capturing its name, optional language, and content.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next()

	fileNode := &ASTNode{
		Type: "File",
		Name: startToken.Data,
	}

	for {
		token := p.next()

		switch token.Type {
		case "EOF", "FileEnd":
			return fileNode
		case "TripleBacktick":
			// Potential language line immediately after opening backticks
			if langToken := p.peek(); langToken.Type == "Text" {
				fileNode.Language = langToken.Data
				p.next() // Consume language
				p.next() // Assume and consume the closing TripleBacktick
				continue
			}
		default:
			fileNode.Content += token.Data + "\n"
		}
	}
}

// Parse runs the parser on the lexer's output to generate an AST.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	root := &ASTNode{Type: "Root"}

	for token := parser.peek(); token.Type != "EOF"; token = parser.peek() {
		if token.Type == "FileStart" {
			root.Children = append(root.Children, parser.parseFile())
		} else {
			textNode := parser.parseText()
			if textNode != nil { // Only add non-empty text nodes
				root.Children = append(root.Children, textNode)
			}
		}
	}

	return root, nil
}

// AsJSON serializes the ASTNode to a JSON string.
func (n *ASTNode) AsJSON() string {
	buf, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}
