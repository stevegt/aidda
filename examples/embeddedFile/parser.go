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
		content += token.Data + "\n" // Ensure each text block ends with a newline
	}
	return NewASTNode("Text", content)
}

func (p *Parser) parseFile() *ASTNode {
	fileStartToken := p.next() // consume FileStart
	fileNode := NewASTNode("File", "")
	fileNode.Name = fileStartToken.Data
	// Reset language for each file parse attempt to ensure it isn't incorrectly carried over
	fileNode.Language = ""
	openingCodeBlock := false // Track the state of code block opening/closing
	
	for {
		token := p.peek()
		if token.Type == "FileEnd" && token.Data == fileNode.Name {
			p.next() // consume FileEnd
			return fileNode // Finished file block
		} else if token.Type == "TripleBacktick" {
			if !openingCodeBlock {
				openingCodeBlock = true
				languageToken := p.next() // Capture the possible language specification
				if languageToken.Data != "" {
					fileNode.Language = languageToken.Data
				}
			} else {
				openingCodeBlock = false
				p.next() // Closing backtick, simply consume it
			}
		} else if token.Type == "Text" && openingCodeBlock {
			// Append text within triple backticks to File content
			fileNode.Content += token.Data + "\n"
			p.next()
		} else if token.Type == "EOF" {
			// Incomplete file block encountered, treat remaining content as Text
			break
		} else {
			p.next() // Consume and ignore unexpected tokens
		}
	}
	return NewASTNode("Text", "Incorrect File Block")
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
