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

// Parse initializes the parsing process and generates an Abstract Syntax Tree from lexer tokens.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := &Parser{lexer: lexer}
	return parser.parseRoot(), nil
}

// parseRoot creates the root of the AST and parses sections of the input.
func (p *Parser) parseRoot() *ASTNode {
	root := &ASTNode{Type: "Root"} // Creating root AST node without content, name or language.
	for {
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

// parseText collects consecutive Text and "TripleBacktick" tokens into a single Text node, correcting newlines.
func (p *Parser) parseText() *ASTNode {
	var contentBuilder strings.Builder
	for {
		token := p.peek()
		if token.Type != "Text" && token.Type != "TripleBacktick" {
			break
		}
		if contentBuilder.Len() > 0 {
			contentBuilder.WriteString("\n")
		}
		contentBuilder.WriteString(token.Data)
		p.next()
	}
	return &ASTNode{Type: "Text", Content: contentBuilder.String()}
}

// parseFile processes FileStart and FileEnd tokens into a File AST node, handling content and language.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next() // Consume FileStart token.
	fileNode := &ASTNode{Type: "File", Name: startToken.Data}
	var contentBuilder strings.Builder

	// Looking for file language after "TripleBacktick"
	if p.peek().Type == "TripleBacktick" {
		p.next()
		if p.peek().Type == "Text" {
			fileNode.Language = p.next().Data // Setting language of the file.
			p.next()                           // Assuming next TripleBacktick closes code block.
		}
	}

	// Collect content until FileEnd token is found matching the file name.
	for p.peek().Type != "EOF" && !(p.peek().Type == "FileEnd" && p.peek().Data == startToken.Data) {
		token := p.next()
		if token.Type == "Text" {
			if contentBuilder.Len() > 0 {
				contentBuilder.WriteString("\n")
			}
			contentBuilder.WriteString(token.Data)
		}
	}

	if p.peek().Type == "FileEnd" {
		p.next() // Consume the FileEnd token.
	}

	fileNode.Content = contentBuilder.String()

	return fileNode
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
