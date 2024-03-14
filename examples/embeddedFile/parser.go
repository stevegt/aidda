package embedded

import (
    "encoding/json"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
    lexer  *Lexer  // The lexer instance
    tokens []Token // Buffered tokens for lookahead
    pos    int     // Current position in the buffered tokens
}

// ASTNode represents a node in the abstract syntax tree.
type ASTNode struct {
    Type     string     `json:"Type"`
    Content  string     `json:"Content,omitempty"`
    Name     string     `json:"Name,omitempty"`
    Language string     `json:"Language,omitempty"`
    Children []*ASTNode `json:"Children,omitempty"`
}

// NewParser creates a new instance of Parser.
func NewParser(lexer *Lexer) *Parser {
    return &Parser{
        lexer: lexer,
    }
}

// Parse generates the AST from lexer tokens.
func Parse(lexer *Lexer) (*ASTNode, error) {
    parser := NewParser(lexer)
    
    root := &ASTNode{Type: "Root"}
    for {
        token := parser.next()
        if token.Type == "EOF" {
            break
        }

        switch token.Type {
        case "Text", "TripleBacktick":
            parser.handleText(root, token)
        case "FileStart":
            parser.handleFileStart(root)
        }
    }

    return root, nil
}

// handleText processes text tokens, appending them directly to the root or the last file node as appropriate.
func (p *Parser) handleText(root *ASTNode, token Token) {
    // This method needs to correctly aggregate consecutive Text tokens, considering TripleBacktick as text as well.

    // Example implementation detail.
}

// handleFileStart initiates processing of a file block, extracting name, optional language, and content.
func (p *Parser) handleFileStart(root *ASTNode) {
    // This method needs to manage the file block creation, including correctly handling a language identifier if present.

    // Example implementation detail.
}

// next retrieves the next token, either from the buffered tokens or directly from the lexer.
func (p *Parser) next() Token {
    if p.pos < len(p.tokens) {
        token := p.tokens[p.pos]
        p.pos++
        return token
    }
    return p.lexer.Next()
}

// AsJSON serializes the ASTNode to JSON string.
func (n *ASTNode) AsJSON() string {
    bytes, err := json.MarshalIndent(n, "", "  ")
    if err != nil {
        return ""
    }
    return string(bytes)
}
