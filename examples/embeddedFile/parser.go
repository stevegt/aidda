package embedded

import (
	"encoding/json"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer *Lexer // Lexer instance
	pos   int    // Current token position
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

// ConcatenateTextNodes concatenates adjacent text nodes in the AST.
func (n *ASTNode) ConcatenateTextNodes() {
	for i := 0; i < len(n.Children); i++ {
		child := n.Children[i]
		if child.Type == "Text" {
			for j := i + 1; j < len(n.Children); j++ {
				next := n.Children[j]
				if next.Type == "Text" {
					child.Content += next.Content
					n.Children = append(n.Children[:j], n.Children[j+1:]...)
					j--
				} else {
					break
				}
			}
		}
		child.ConcatenateTextNodes()
	}
}

func (p *Parser) Try(fn func(...any) *ASTNode, args ...any) *ASTNode {
	lex := p.lexer
	cp := lex.Checkpoint()
	node := fn(args...)
	if node == nil {
		lex.Rollback(cp)
		return nil
	}
	return node
}

func (p *Parser) parse() *ASTNode {
	root := NewASTNode("Root", "")
	for {
		var node *ASTNode
		for {
			node = p.Try(p.parseEOF)
			if node != nil {
				break
			}
			node = p.Try(p.parseFile)
			if node != nil {
				break
			}
			node = p.parseText()
			break
		}
		root.Children = append(root.Children, node)
		if node.Type == "EOF" {
			break
		}
	}
	root.ConcatenateTextNodes()
	return root
}

func (p *Parser) parseText() *ASTNode {
	lex := p.lexer
	token := lex.Next()
	textNode := &ASTNode{Type: "Text", Content: token.Data}
	return textNode
}

func (p *Parser) parseEOF(args ...any) *ASTNode {
	token := p.lexer.Next()
	if token.Type == "EOF" {
		return NewASTNode("EOF", "")
	}
	return nil
}

func (p *Parser) parseFile(args ...any) *ASTNode {
	fileStartToken := p.lexer.Next()
	if fileStartToken.Type != "FileStart" {
		return nil
	}
	fileNode := NewASTNode("File", "")
	fileNode.Name = fileStartToken.Data

	codeNode := p.parseCodeBlock(fileNode.Name)
	if codeNode == nil {
		return nil
	}
	fileNode.Children = append(fileNode.Children, codeNode)
	return fileNode
}

func (p *Parser) parseCodeBlock(fileName string) *ASTNode {
	lex := p.lexer
	codeNode := NewASTNode("CodeBlock", "")
	openTickNode := p.parseTripleBacktick()
	if openTickNode == nil {
		return nil
	}
	codeNode.Language = openTickNode.Language
	// collect content until we hit either another triple backtick, a
	// FileEnd token with the same file name, or EOF
	for {
		eofNode := p.Try(p.parseEOF)
		if eofNode != nil {
			// end of input -- malformed code block
			return nil
		}
		cpBacktick := lex.Checkpoint()
		backtickNode := p.Try(p.parseTripleBacktick)
		fileEndNode := p.Try(p.parseFileEnd, fileName)
		if backtickNode != nil {
			// Triple backtick found -- end of code block
			if fileEndNode != nil {
				// properly-formed end of file block -- discard the FileEnd token
				// and close the code block
				break
			}
			// backtick with no following file end
			if fileName == "" {
				// no file name was given, so we're really just looking for the end of the code block
				// and we've found it
				break
			}
			// we're looking for a file end, but we found backticks
			// instead -- rollback and treat the backticks as text
			lex.Rollback(cpBacktick)
		}
		textNode := p.parseText()
		codeNode.Children = append(codeNode.Children, textNode)
	}
	return codeNode
}

func (p *Parser) parseFileEnd(args ...any) *ASTNode {
	fileName := args[0].(string)
	token := p.lexer.Next()
	if token.Type == "FileEnd" && token.Data == fileName {
		return NewASTNode("FileEnd", "")
	}
	return nil
}

func (p *Parser) parseTripleBacktick(args ...any) *ASTNode {
	token := p.lexer.Next()
	if token.Type == "TripleBacktick" {
		node := NewASTNode("TripleBacktick", "")
		node.Language = token.Data
		return node
	}
	return nil
}

// Parse create and runs a parser on the lexer's output and generates an AST.
func Parse(lexer *Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	root := parser.parse()
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
