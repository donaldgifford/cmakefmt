package parser

import "github.com/donaldgifford/cmakefmt/internal/lexer"

// Node is the interface all AST nodes implement.
type Node interface {
	nodeType() string
}

// File is the root node of a CMake file.
type File struct {
	Elements []Node
}

func (f *File) nodeType() string { return "File" }

// CommandInvocation represents a command call like `set(FOO "bar")`.
type CommandInvocation struct {
	Name      string
	NameToken lexer.Token
	Arguments []Argument
	// LeadingSpace is whitespace before the command name on the same line.
	LeadingSpace string
	// TrailingComment is an inline comment after the closing paren.
	TrailingComment string
	Line            int
}

func (c *CommandInvocation) nodeType() string { return "Command" }

// Argument represents a single argument to a command.
type Argument struct {
	Token lexer.Token
	// Nested arguments (for parenthesized groups).
	Children []Argument
}

// Comment represents a standalone comment (not attached to a command).
type Comment struct {
	Value string
	Line  int
	// LeadingSpace is the whitespace before the comment.
	LeadingSpace string
}

func (c *Comment) nodeType() string { return "Comment" }

// BlankLine represents one or more consecutive blank lines.
type BlankLine struct {
	Count int
	Line  int
}

func (b *BlankLine) nodeType() string { return "BlankLine" }

// BlockNode represents an indented block (if/endif, foreach/endforeach, etc.).
type BlockNode struct {
	Opener *CommandInvocation
	Body   []Node
	Closer *CommandInvocation
	// Middles holds elseif/else commands and their body sections.
	Middles []BlockMiddle
}

func (b *BlockNode) nodeType() string { return "Block" }

// BlockMiddle represents an else/elseif and the body following it.
type BlockMiddle struct {
	Command *CommandInvocation
	Body    []Node
}
