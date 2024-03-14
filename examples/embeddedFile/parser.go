package embedded

import (
	"encoding/json"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer  *Lexer  // Lexer instance
	tokens []Token // Token buffer
	pos    int     // Current position in token buffer
}

// NewParser initializes a new parser from a lexer.
func NewParser(lexer *Lexer) *Parser {
	return &Parser{lexer: lexer}
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

// next retrieves the next token from the lexer or the buffer.
func (p *Parser) next() Token {
	if p.pos >= len(p.tokens) {
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}

// peek looks at the next token without consuming it.
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	return p.tokens[p.pos]
}

// parseRoot initializes and fills the root node of the AST.
func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "")
	for {
		token := p.peek()
		switch token.Type {
		case "EOF":
			return root
		case "Text":
			textNode := p.parseText()
			root.Children = append(root.Children, textNode)
		case "FileStart":
			fileNode := p.parseFile()
			root.Children = append(root.Children, fileNode)
		default:
			p.next() // Consume unknown tokens to avoid infinite loop
		}
	}
}

// parseText collects consecutive Text tokens into a single Text node.
func (p *Parser) parseText() *ASTNode {
	content := ""
	for {
		token := p.peek()
		if token.Type != "Text" {
			break
		}
		content += token.Data + "\n" // Append the text data
		p.next()                      // Consume the token
	}
	return NewASTNode("Text", content)
}

// parseFile extracts file content from the input, delimited by FileStart and EOF markers.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next() // consume FileStart
	fileNode := NewASTNode("File", "")
	fileNode.Name = startToken.Data
	content := ""
	for {
		token := p.next()
		if token.Type == "FileEnd" && token.Data == startToken.Data {
			break // Found matching EOF marker, end file block
		} else if token.Type == "EOF" {
			// If EOF token is encountered before closing FileEnd, treat the rest as Text
			fileNode = NewASTNode("Text", "File: "+startToken.Data+"\n"+content)
			break
		}
		content += token.Data + "\n"
	}
	fileNode.Content = content
	return fileNode
}

// Parse runs the parser on the lexer's output to generate an AST.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	root := parser.parseRoot()
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

