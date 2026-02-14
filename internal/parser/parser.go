package parser

import (
	"fmt"

	"github.com/donaldgifford/cmakefmt/internal/lexer"
)

// Parser builds an AST from a stream of tokens.
type Parser struct {
	tokens []lexer.Token
	pos    int
}

// New creates a new Parser for the given tokens.
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse processes all tokens and returns the root File node.
func (p *Parser) Parse() (*File, error) {
	elements, err := p.parseElements(false)
	if err != nil {
		return nil, err
	}
	return &File{Elements: elements}, nil
}

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) parseElements(inBlock bool) ([]Node, error) {
	var elements []Node
	blankCount := 0

	for {
		tok := p.peek()

		switch tok.Type {
		case lexer.TokenEOF:
			if blankCount > 0 {
				elements = append(elements, &BlankLine{Count: blankCount, Line: tok.Line})
			}
			return elements, nil

		case lexer.TokenNewline:
			p.advance()
			blankCount++
			if blankCount > 1 {
				// Will emit a BlankLine when we hit the next non-blank token.
			}
			continue

		case lexer.TokenSpace:
			// Consume leading space, might be needed for commands.
			p.advance()
			continue

		case lexer.TokenLineComment, lexer.TokenBracketComment:
			if blankCount > 1 {
				elements = append(elements, &BlankLine{Count: blankCount - 1, Line: tok.Line})
			}
			blankCount = 0
			comment := p.parseComment()
			elements = append(elements, comment)

		case lexer.TokenIdentifier:
			if blankCount > 1 {
				elements = append(elements, &BlankLine{Count: blankCount - 1, Line: tok.Line})
			}
			blankCount = 0

			// Check if this is a block closer or middle — if so, return to parent.
			if inBlock && (lexer.IsBlockCloser(tok.Value) || lexer.IsBlockMiddle(tok.Value)) {
				return elements, nil
			}

			cmd, err := p.parseCommand()
			if err != nil {
				return nil, err
			}

			// If the command opens a block, parse the block body.
			if lexer.IsBlockOpener(cmd.Name) {
				block, err := p.parseBlock(cmd)
				if err != nil {
					return nil, err
				}
				elements = append(elements, block)
			} else {
				elements = append(elements, cmd)
			}

		default:
			// Skip unexpected tokens.
			p.advance()
			blankCount = 0
		}
	}
}

func (p *Parser) parseComment() *Comment {
	tok := p.advance()
	return &Comment{
		Value: tok.Value,
		Line:  tok.Line,
	}
}

func (p *Parser) parseCommand() (*CommandInvocation, error) {
	nameTok := p.advance() // Identifier
	cmd := &CommandInvocation{
		Name:      nameTok.Value,
		NameToken: nameTok,
		Line:      nameTok.Line,
	}

	// Skip whitespace between name and '('.
	p.skipSpaceAndNewlines()

	if p.peek().Type != lexer.TokenLeftParen {
		return nil, fmt.Errorf("expected '(' after command %q at line %d, got %s",
			cmd.Name, nameTok.Line, p.peek().Type)
	}
	p.advance() // consume '('

	args, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	cmd.Arguments = args

	if p.peek().Type != lexer.TokenRightParen {
		return nil, fmt.Errorf("expected ')' to close command %q at line %d, got %s",
			cmd.Name, nameTok.Line, p.peek().Type)
	}
	p.advance() // consume ')'

	// Check for trailing inline comment.
	p.skipSpaces()
	if p.peek().Type == lexer.TokenLineComment {
		cmd.TrailingComment = p.advance().Value
	}

	return cmd, nil
}

func (p *Parser) parseArguments() ([]Argument, error) {
	var args []Argument

	for {
		p.skipSpaceAndNewlines()
		tok := p.peek()

		switch tok.Type {
		case lexer.TokenRightParen, lexer.TokenEOF:
			return args, nil
		case lexer.TokenLeftParen:
			p.advance()
			children, err := p.parseArguments()
			if err != nil {
				return nil, err
			}
			if p.peek().Type == lexer.TokenRightParen {
				p.advance()
			}
			args = append(args, Argument{
				Token:    tok,
				Children: children,
			})
		case lexer.TokenQuotedArg, lexer.TokenUnquotedArg, lexer.TokenBracketArg:
			args = append(args, Argument{Token: p.advance()})
		case lexer.TokenLineComment:
			// Comments inside argument lists are preserved as arguments.
			args = append(args, Argument{Token: p.advance()})
		default:
			p.advance() // skip unexpected
		}
	}
}

func (p *Parser) parseBlock(opener *CommandInvocation) (*BlockNode, error) {
	block := &BlockNode{Opener: opener}

	body, err := p.parseElements(true)
	if err != nil {
		return nil, err
	}
	block.Body = body

	// Handle else/elseif middles.
	for p.peek().Type == lexer.TokenIdentifier && lexer.IsBlockMiddle(p.peek().Value) {
		mid, err := p.parseCommand()
		if err != nil {
			return nil, err
		}
		midBody, err := p.parseElements(true)
		if err != nil {
			return nil, err
		}
		block.Middles = append(block.Middles, BlockMiddle{
			Command: mid,
			Body:    midBody,
		})
	}

	// Parse the closer.
	if p.peek().Type == lexer.TokenIdentifier && lexer.IsBlockCloser(p.peek().Value) {
		closer, err := p.parseCommand()
		if err != nil {
			return nil, err
		}
		block.Closer = closer
	}

	return block, nil
}

func (p *Parser) skipSpaces() {
	for p.peek().Type == lexer.TokenSpace {
		p.advance()
	}
}

func (p *Parser) skipSpaceAndNewlines() {
	for {
		t := p.peek().Type
		if t == lexer.TokenSpace || t == lexer.TokenNewline {
			p.advance()
			continue
		}
		break
	}
}
