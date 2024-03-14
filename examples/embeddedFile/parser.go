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

// ASTNode structure
type ASTNode struct {
	Type     string     `json:"Type"`
	Content  string     `json:"Content"`
	Name     string     `json:"Name,omitempty"`
	Language string     `json:"Language,omitempty"`
	Children []*ASTNode `json:"Children,omitempty"`
}

// NewParser initializes a new parser from the provided lexer.
func NewParser(lexer *Lexer) *Parser {
	return &Parser{lexer: lexer}
}

// NewASTNode creates and returns a new AST node.
func NewASTNode(nodeType, content string) *ASTNode {
	return &ASTNode{Type: nodeType, Content: content}
}

// next retrieves and consumes the next token.
func (p *Parser) next() Token {
	if p.pos >= len(p.tokens) {
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}

// peek returns the next token without consuming it.
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	return p.tokens[p.pos]
}

// parseRoot parses the token stream starting from the root.
func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "")
	for {
		token := p.peek()
		switch token.Type {
		case "EOF":
			return root
		case "FileStart":
			root.Children = append(root.Children, p.parseFile())
		default:
			root.Children = append(root.Children, p.parseText())
		}
	}
}

// parseText handles aggregation of text, including triple backticks as text.
func (p *Parser) parseText() *ASTNode {
	var content strings.Builder
	for {
		token := p.peek()
		if token.Type != "Text" && token.Type != "TripleBacktick" && token.Type != "FileStart" || token.Type == "EOF" {
			break
		}
		if token.Type == "TripleBacktick" {
			content.WriteString("```\n")
		} else {
			content.WriteString(token.Data)
			content.WriteRune('\n') // Newline as separator for actual text lines
		}
		p.next()
	}
	return NewASTNode("Text", content.String())
}

// parseFile extracts and constructs a file node from the token stream.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next()
	fileNode := NewASTNode("File", "")
	fileNode.Name = startToken.Data

	var content strings.Builder
	for {
		token := p.next()
		if token.Type == "FileEnd" && token.Data == fileNode.Name || token.Type == "EOF" {
			break
		}
		content.WriteString(token.Data + "\n")
	}
	fileNode.Content = strings.TrimRight(content.String(), "\n") // Trim the last newline added above
	return fileNode
}

// Parse initiates parsing and returns the AST root node.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	return parser.parseRoot(), nil
}

// AsJSON serializes the ASTNode to a JSON string.
func (n *ASTNode) AsJSON() string {
	buf, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return ""
	}
	return string(buf)
}
