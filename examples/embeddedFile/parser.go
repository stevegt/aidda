package embedded

import (
	"encoding/json"
	"strings"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer  *Lexer  // Lexer instance
	tokens []Token // Token buffer
	pos    int     // Current token position
}

// NewParser initializes a new parser from a lexer.
func NewParser(lexer *Lexer) *Parser {
	return &Parser{lexer: lexer, tokens: []Token{}}
}

// ASTNode represents a node in the abstract syntax tree.
type ASTNode struct {
	Type     string     `json:"Type"`
	Content  string     `json:"Content"`
	Name     string     `json:"Name,omitempty"`
	Language string     `json:"Language,omitempty"`
	Children []*ASTNode `json:"Children,omitempty"`
}

// NewASTNode creates a new AST node given its type and content.
func NewASTNode(nodeType, content string) *ASTNode {
	return &ASTNode{Type: nodeType, Content: content}
}

func (p *Parser) next() Token {
	if p.pos >= len(p.tokens) {
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	return p.tokens[p.pos]
}

func (p *Parser) parseTokenAsText() *ASTNode {
	var content strings.Builder
	for {
		token := p.peek()
		switch token.Type {
		case "EOF", "FileStart":
			break
		default:
			content.WriteString(token.Data + "\n")
			p.next()
			continue
		}
		break
	}
	// Remove the trailing newline from content before returning the node.
	return NewASTNode("Text", strings.TrimSuffix(content.String(), "\n"))
}

func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "")
	for {
		token := p.peek()
		if token.Type == "EOF" {
			break
		}
		textNode := p.parseTokenAsText()
		root.Children = append(root.Children, textNode)
	}
	return root
}

// Implement parseFile according to your parsing requirements for file blocks.
// Placeholder implementation provided for simplicity.
func (p *Parser) parseFile() *ASTNode {
	return NewASTNode("File", "Placeholder implementation. Complete as required.")
}

// Parse runs the parser on the lexer's output and generates an AST.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	root := parser.parseRoot()
	return root, nil
}

// AsJSON returns the AST as a JSON string.
func (n *ASTNode) AsJSON() string {
	buf, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}
