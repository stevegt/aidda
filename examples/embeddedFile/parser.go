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

func (p *Parser) rollBackToLastCheckpoint() {
	if p.pos > 0 {
		// Move one position back
		p.pos--
	}
	if len(p.tokens) > p.pos {
		p.tokens = p.tokens[:p.pos] // Truncate tokens at the current position
	}
	p.lexer.Rollback()
}

func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "")
	for {
	p.lexer.Checkpoint()
	token := p.peek()
	switch token.Type {
		case "EOF":
			return root
		case "Text", "TripleBacktick":
			textNode := p.collateTextAndTripleBackticks()
			root.Children = append(root.Children, textNode)
		case "FileStart":
			fileNode := p.attemptParseFile()
			if fileNode != nil {
				root.Children = append(root.Children, fileNode)
			} else {
				// Parsing as file failed, tokenizer has been rolled back, parse as text
				textNode := p.collateTextAndTripleBackticks()
				root.Children = append(root.Children, textNode)
			}
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

func (p *Parser) attemptParseFile() *ASTNode {
	lastCheckpointRollbackPos := len(p.tokens)
	fileStartToken := p.next() // consume FileStart
	fileNode := NewASTNode("File", "")
	fileNode.Name = fileStartToken.Data
	codeBlockFound := false
	
	for {
		token := p.peek()
		if token.Type == "FileEnd" && token.Data == fileNode.Name {
			p.next() // consume FileEnd
			return fileNode // Successfully finished file block
		} else if token.Type == "EOF" {
			// Encountered EOF without closing FileEnd, rollback to last checkpoint
			p.pos = lastCheckpointRollbackPos // Reset parser's position to before attempting to parse file
			p.rollBackToLastCheckpoint() // Rollback lexer to last checkpoint before "FileStart"
			return nil // Signal failure to parse as file
		} else if token.Type == "TripleBacktick" {
			codeBlockFound = !codeBlockFound
			p.next() // Consume triple backtick
		} else if token.Type == "Text" && codeBlockFound {
			// Append text within triple backticks to File content
			fileNode.Content += token.Data + "\n"
			p.next()
		} else {
			p.next() // Skip unexpected tokens
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

