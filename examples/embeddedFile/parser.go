package embedded

// Parser parses tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer  *Lexer       // Lexer instance to draw tokens from
	tokens []Token      // Buffer of tokens read from the lexer
	pos    int          // Current position in the tokens buffer
}

// NewParser creates a new Parser instance given a lexer.
func NewParser(lexer *Lexer) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

// next returns the next token from the buffer, loading new tokens from the lexer as needed.
func (p *Parser) next() Token {
	if p.pos >= len(p.tokens) { // Need more tokens
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}

// peek returns the next token without advancing the parser's position.
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) { // Need more tokens
		token := p.lexer.Next()
		p.tokens = append(p.tokens, token)
	}
	return p.tokens[p.pos]
}

// parseRoot parses the input into a root AST node.
func (p *Parser) parseRoot() *ASTNode {
	root := NewASTNode("Root", "")

	var lastText *ASTNode
	for {
		token := p.peek()
		switch token.Type {
		case "EOF":
			return root
		case "Text":
			if lastText != nil {
				// Append new text content to the last Text node
				lastText.Content += "\n" + token.Data
				p.next() // consume the Text token
			} else {
				lastText = p.parseText()
				root.addChild(lastText)
			}
		case "FileStart":
			child := p.parseFile()
			root.addChild(child)
			lastText = nil // reset lastText pointer
		default:
			p.next() // Skip unexpected token
			lastText = nil // reset lastText pointer
		}
	}
}

// parseText consumes and returns a text node.
func (p *Parser) parseText() *ASTNode {
	token := p.next() // consume the text token
	if token.Type == "Text" {
		return NewASTNode("Text", token.Data)
	}
	return NewASTNode("Error", "expected a text token")
}

// parseFile parses a file block, including its content, and returns a file node.
func (p *Parser) parseFile() *ASTNode {
	p.next() // consume the FileStart token, it's not directly used

	fileNode := NewASTNode("File", "")

	// Collect content until FileEnd or EOF. Include the newline as per test requirement.
	content := ""
	language := ""
	inCodeBlock := false
	for {
		token := p.peek()
		if token.Type == "FileEnd" {
			fileNode.Name = token.Data // Set file node name after consuming FileEnd token
			p.next() // consume the FileEnd token
			break
		} else if token.Type == "EOF" {
			break // stop at EOF even if matching FileEnd not found
		} else if token.Type == "TripleBacktick" {
			inCodeBlock = !inCodeBlock
			if inCodeBlock { // Capture language if starting a code block
				language = token.Data
				p.next() // consume TripleBacktick token to fetch language
				continue
			}
		}

		// Consume and collect text or non-code block token data
		if token.Type == "Text" || (!inCodeBlock && token.Type != "TripleBacktick") {
			content += token.Data + "\n" // Include the newline character
		}
		p.next() // consume token
	}
	fileNode.Content = content
	if language != "" {
		fileNode.Language = language // Set language property to the AST node
	}
	return fileNode
}

// Parse runs the parser on the lexer's output and returns the root AST node.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	root := parser.parseRoot()
	return root, nil // Placeholder for now. Add error handling as required.
}

// ASTNode represents a node in the abstract syntax tree.
type ASTNode struct {
	Type      string
	Content   string
	Name      string    // To hold file name if this node represents a file
	Language  string    // To hold language info if this node represents a block of code
	children  []*ASTNode // Children nodes
}

// NewASTNode creates a new ASTNode given a type and content.
func NewASTNode(nodeType, content string) *ASTNode {
	return &ASTNode{
		Type:    nodeType,
		Content: content,
	}
}

// addChild adds a child node to this AST node.
func (n *ASTNode) addChild(child *ASTNode) {
	n.children = append(n.children, child)
}

// Children returns the child nodes of this AST node.
func (n *ASTNode) Children() []*ASTNode {
	return n.children
}
