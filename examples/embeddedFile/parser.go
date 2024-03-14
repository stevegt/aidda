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
	Content  string     `json:"Content,omitempty"`
	Name     string     `json:"Name,omitempty"`
	Language string     `json:"Language,omitempty"`
	Children []*ASTNode `json:"Children,omitempty"`
}

// NewASTNode creates a new AST node given its type and optional information.
func NewASTNode(nodeType, content, name, language string) *ASTNode {
	return &ASTNode{Type: nodeType, Content: content, Name: name, Language: language}
}

// Parse initializes the parsing process and generates an Abstract Syntax Tree from lexer tokens.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := &Parser{lexer: lexer}
	return parser.parseRoot(), nil
}

// parseRoot creates the root of the AST and parses sections of the input.
func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "", "", "") // Creating root AST node without content, name or language.
	for {
		// Peek at the next token to decide what to do without consuming it.
		token := p.peek()
		switch token.Type {
		case "EOF": // End of tokens.
			return root
		case "FileStart": // Start of a new file content.
			root.Children = append(root.Children, p.parseFile())
		case "Text", "TripleBacktick":
			root.Children = append(root.Children, p.parseText())
		default:
			p.next() // Skip unknown token types.
		}
	}
}

// parseText collects consecutive Text tokens and returns a single Text node.
func (p *Parser) parseText() *ASTNode {
	var contentBuilder strings.Builder
	for p.peek().Type == "Text" || p.peek().Type == "TripleBacktick" {
		token := p.next()
		contentBuilder.WriteString(token.Data)
		if token.Type == "Text" {
			contentBuilder.WriteString("\n") // Add newline for text content, separated by Text tokens.
		}
	}
	content := strings.TrimSuffix(contentBuilder.String(), "\n") // Remove the last newline added.
	return NewASTNode("Text", content, "", "")
}

// parseFile processes FileStart, FileEnd, and internal content into a File AST node.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next() // Consume FileStart token.
	var contentBuilder strings.Builder
	var language string

	// Collect content until FileEnd token is found that matches the file name.
	for !(p.peek().Type == "FileEnd" && p.peek().Data == startToken.Data) && p.peek().Type != "EOF" {
		token := p.next()
		if token.Type == "TripleBacktick" {
			// Next token should specify the language or be the end of the code block.
			languageToken := p.next()
			if languageToken.Type == "Text" { // Language specified.
				language = languageToken.Data
				p.next() // Assume the next TripleBacktick token is the end of the code block, so consume it.
			}
		} else {
			contentBuilder.WriteString(token.Data)
			contentBuilder.WriteString("\n")
		}
	}

	if p.peek().Type == "FileEnd" {
		p.next() // Consume the FileEnd token.
	}

	content := strings.TrimSuffix(contentBuilder.String(), "\n") // Remove the last newline added.

	return NewASTNode("File", content, startToken.Data, language)
}

// peek returns the next token without consuming it, loading more tokens if necessary.
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) { // If we've reached the end of buffered tokens, load more from the lexer.
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	return p.tokens[p.pos]
}

// next consumes and returns the next token.
func (p *Parser) next() Token {
	token := p.peek()
	p.pos++ // Advance to the next token.
	return token
}

// AsJSON converts an ASTNode to its JSON representation.
func (n *ASTNode) AsJSON() string {
	buf, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return ""
	}
	return string(buf)
}

