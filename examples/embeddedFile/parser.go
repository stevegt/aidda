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
// This function is essential for creating new instances of ASTNode.
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
		case "FileStart":
			fileNode := p.parseFile()
			root.Children = append(root.Children, fileNode)
		default:
			textNode := p.collectText() // Collecting text outside of File blocks in a separate method
			root.Children = append(root.Children, textNode)
		}
	}
}

// collectText gathers consecutive non-file tokens.
func (p *Parser) collectText() *ASTNode {
	content := ""
	for {
		token := p.peek()
		if token.Type == "EOF" || token.Type == "FileStart" {
			break
		}
		content += token.Data
		if token.Type == "TripleBacktick" {
			content += "\n" // Add newline for triple backticks
		}
		p.next()
	}
	// Removing the trailing newlines only for content that ends with triple backticks
	content = strings.TrimRight(content, "\n")
	return NewASTNode("Text", content)
}

// parseFile extracts file content, correctly handling the EOF and language.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next() // consume FileStart
	fileNode := NewASTNode("File", "")
	fileNode.Name = startToken.Data
	content := ""
	var languageLine bool // flag to identify language specification line
	for {
		token := p.next()
		if token.Type == "FileEnd" && token.Data == startToken.Data {
			break // Found matching EOF marker, end file block
		} else if token.Type == "EOF" {
			break // Handle case without EOF marker
		} else if token.Type == "TripleBacktick" {
			if !languageLine { // Skip language line if already captured
				fileNode.Language = token.Data
				languageLine = true
				continue
			}
		}
		// Accumulate file content
		content += token.Data + "\n"
	}
	fileNode.Content = strings.TrimSuffix(content, "\n") // Remove trailing newline
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
