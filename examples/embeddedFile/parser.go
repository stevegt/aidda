package embedded

import (
	"encoding/json"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer  *Lexer  // Lexer instance
	tokens []Token // Token buffer
	pos    int     // Current token position
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

func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "")
	for {
		token := p.peek()
		switch token.Type {
		case "EOF":
			return root
		case "Text", "TripleBacktick":
			textNode := p.collateTextAndTripleBackticks()
			root.Children = append(root.Children, textNode)
		case "FileStart":
			fileNode := p.parseFile()
			root.Children = append(root.Children, fileNode)
		default:
			p.next()
		}
	}
}

func (p *Parser) collateTextAndTripleBackticks() *ASTNode {
	content := ""
	for {
		token := p.peek()
		if token.Type != "Text" && token.Type != "TripleBacktick" {
			break
		}
		p.next() // consume token
		if token.Type == "TripleBacktick" {
			content += "```\n"
		} else {
			content += token.Data + "\n" // Ensure each text block ends with a newline
		}
	}
	return NewASTNode("Text", content)
}

func (p *Parser) parseFile() *ASTNode {
	fileStartToken := p.next() // consume FileStart
	fileNode := NewASTNode("File", "")
	fileNode.Name = fileStartToken.Data
	codeBlockFound := false
	
	for {
		token := p.peek()
		if token.Type == "TripleBacktick" {
			codeBlockFound = !codeBlockFound
			if codeBlockFound {
				// Consume the triple backtick and potentially a language tag
				tripleBacktickToken := p.next()
				// Check for language identifier only on the opening triple backtick
				if fileNode.Language == "" && tripleBacktickToken.Data != "" {
					fileNode.Language = tripleBacktickToken.Data
				}
			} else {
				// Consume closing triple backtick
				p.next()
			}
		} else if token.Type == "FileEnd" || token.Type == "EOF" {
			p.next() // consume FileEnd or EOF
			return fileNode // Finished file block
		} else if token.Type == "Text" && codeBlockFound {
			// Append text within triple backticks to File content
			fileNode.Content += token.Data + "\n"
			p.next()
		} else {
			p.next() // Skip unexpected tokens (outside code blocks)
		}
	}
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
