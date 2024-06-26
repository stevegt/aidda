package parser

import (
	"encoding/json"
	// . "github.com/stevegt/goadapt"
	"embedded/lexer"

	. "github.com/stevegt/goadapt"
)

// Parser prepares tokens from a lexer into an Abstract Syntax Tree.
type Parser struct {
	lexer *lexer.Lexer // Lexer instance
	pos   int          // Current token position
}

// NewParser initializes a new parser from a lexer.
func NewParser(lexer *lexer.Lexer) *Parser {
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

// nextNode returns the next node from the lexer.
func (p *Parser) nextNode() *ASTNode {
	var node *ASTNode
	for {
		node = p.Try(p.parseRole)
		if node != nil {
			break
		}
		node = p.Try(p.parseEOF)
		if node != nil {
			break
		}
		node = p.Try(p.parseFile)
		if node != nil {
			break
		}
		node = p.Try(p.parseCodeBlock, "")
		if node != nil {
			break
		}
		node = p.parseAnyAsText()
		break
	}
	return node
}

func (p *Parser) parse() *ASTNode {
	root := NewASTNode("Root", "")
	for {
		node := p.nextNode()
		root.Children = append(root.Children, node)
		if node.Type == "EOF" {
			break
		}
	}
	root.ConcatenateTextNodes()
	return root
}

func (p *Parser) parseRole(args ...any) *ASTNode {
	lex := p.lexer
	token := lex.Next()
	if token.Type != "Role" {
		return nil
	}
	roleNode := &ASTNode{Type: "Role", Name: token.Payload, Content: token.Src}
	// there might be a newline after the role name -- if so, skip it
	p.Try(p.parseNewline)
	// parse the role's children -- they might be anything other
	// than a role
	for {
		cp := lex.Checkpoint()
		node := p.nextNode()
		if node.Type == "EOF" {
			// we're at the end of the input
			lex.Rollback(cp)
			break
		}
		if node.Type == "Role" {
			// we've hit another role, so we're done with this one
			lex.Rollback(cp)
			break
		}
		roleNode.Children = append(roleNode.Children, node)
	}
	return roleNode
}

func (p *Parser) parseAnyAsText() *ASTNode {
	lex := p.lexer
	token := lex.Next()
	textNode := &ASTNode{Type: "Text", Content: token.Src}
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
	fileNode.Name = fileStartToken.Payload
	if p.Try(p.parseNewline) == nil {
		return nil
	}
	codeNode := p.parseCodeBlock(fileNode.Name)
	if codeNode == nil {
		return nil
	}
	fileNode.Language = codeNode.Language
	fileNode.Children = append(fileNode.Children, codeNode.Children...)
	return fileNode
}

func (p *Parser) parseCodeBlock(args ...any) *ASTNode {
	fileName := args[0].(string)
	lex := p.lexer
	codeNode := NewASTNode("CodeBlock", "")
	// cpFirstBacktick := lex.Checkpoint()
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
		if fileName == "" {
			// no file name was given, so we're just looking for the end of the code block
			backtickNode := p.Try(p.parseTripleBacktick)
			if backtickNode != nil {
				// end of code block
				break
			}
		} else {
			// we're looking for code block end followed by a file end
			cpBacktick := lex.Checkpoint()
			backtickNode := p.Try(p.parseTripleBacktick)
			fileEndNode := p.Try(p.parseFileEnd, fileName)
			if backtickNode != nil && fileEndNode != nil {
				// properly-formed end of file block
				break
			}
			lex.Rollback(cpBacktick)
		}
		// anything else is just text
		textNode := p.parseAnyAsText()
		codeNode.Children = append(codeNode.Children, textNode)
	}
	return codeNode
}

func (p *Parser) parseNewline(args ...any) *ASTNode {
	token := p.lexer.Next()
	if token.Type == "Newline" {
		return NewASTNode("Newline", token.Src)
	}
	return nil
}

func (p *Parser) parseNewlineOrEOF(args ...any) *ASTNode {
	node := p.Try(p.parseNewline)
	if node == nil {
		node = p.Try(p.parseEOF)
	}
	return node
}

func (p *Parser) parseFileEnd(args ...any) *ASTNode {
	fileName := args[0].(string)
	token := p.lexer.Next()
	if token.Type == "FileEnd" && token.Payload == fileName {
		if p.Try(p.parseNewlineOrEOF) == nil {
			return nil
		}
		return NewASTNode("FileEnd", "")
	}
	return nil
}

func (p *Parser) parseTripleBacktick(args ...any) *ASTNode {
	token := p.lexer.Next()
	if token.Type == "TripleBacktick" {
		node := NewASTNode("TripleBacktick", "")
		node.Language = token.Payload
		if p.Try(p.parseNewlineOrEOF) == nil {
			return nil
		}
		return node
	}
	return nil
}

// Parse create and runs a parser on the lexer's output and generates an AST.
func Parse(lexer *lexer.Lexer) (*ASTNode, error) {
	parser := NewParser(lexer)
	root := parser.parse()
	return root, nil
}

// AsJSON returns the AST as a JSON string.
func (n *ASTNode) AsJSON(pretty bool) string {
	var buf []byte
	var err error
	if pretty {
		buf, err = json.MarshalIndent(n, "", "  ")
	} else {
		buf, err = json.Marshal(n)
	}
	Ck(err)
	return string(buf)
}
