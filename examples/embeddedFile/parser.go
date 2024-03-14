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
			textNode := p.parseText() // Collecting text outside of File blocks in a separate method
			if textNode != nil {
				root.Children = append(root.Children, textNode)
			}
		}
	}
}

// parseText gathers consecutive text tokens until another token type is encountered.
func (p *Parser) parseText() *ASTNode {
	content := strings.Builder{}
	for {
		token := p.peek()
		if token.Type != "Text" && token.Type != "TripleBacktick" {
			break
		}
		if token.Type == "EOF" { // Assuring End of File handling
			break
		}
		p.next() // Consume the token.
		if token.Type == "TripleBacktick" {
			content.WriteString("```\n")
		} else {
			content.WriteString(token.Data + "\n")
		}
	}
	if content.Len() > 0 {
		// See if there is a need to remove a trailing newline.
		finalContent := strings.TrimSuffix(content.String(), "\n")
		return NewASTNode("Text", finalContent)
	}
	return nil
}

// parseFile extracts file content, correctly handling the EOF and language.
func (p *Parser) parseFile() *ASTNode {
	startToken := p.next() // consume FileStart
	fileNode := NewASTNode("File", "")
	fileNode.Name = startToken.Data
	content := strings.Builder{}
	languageParsed := false
	for {
		token := p.next()
		if token.Type == "FileEnd" && token.Data == startToken.Data {
			break // Found matching EOF marker, end file block
		} else if token.Type == "EOF" {
			break // Handle case without EOF marker
		} else if token.Type == "TripleBacktick" {
			if !languageParsed { // Capturing the language identifier
				languageParsed = true
				fileNode.Language = p.next().Data // Consuming and setting language
				p.next() // Skip closing triple backtick
				continue
			}
		} else {
			content.WriteString(token.Data + "\n")
		}
	}
	fileNode.Content = strings.TrimSuffix(content.String(), "\n") // Remove trailing newline
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

