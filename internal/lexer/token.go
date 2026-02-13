package lexer

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// TokenEOF indicates end of input.
	TokenEOF TokenType = iota
	// TokenNewline is a newline character (or \r\n).
	TokenNewline
	// TokenSpace is horizontal whitespace (spaces/tabs, not newlines).
	TokenSpace
	// TokenIdentifier is a command name (e.g., "set", "if", "target_link_libraries").
	TokenIdentifier
	// TokenLeftParen is '('.
	TokenLeftParen
	// TokenRightParen is ')'.
	TokenRightParen
	// TokenQuotedArg is a quoted argument including quotes (e.g., `"hello world"`).
	TokenQuotedArg
	// TokenUnquotedArg is an unquoted argument (e.g., `foo`, `${VAR}`, `$<GENEX>`).
	TokenUnquotedArg
	// TokenBracketArg is a bracket argument (e.g., `[=[content]=]`).
	TokenBracketArg
	// TokenLineComment is a line comment (e.g., `# this is a comment`).
	TokenLineComment
	// TokenBracketComment is a bracket comment (e.g., `#[=[comment]=]`).
	TokenBracketComment
)

var tokenNames = map[TokenType]string{
	TokenEOF:            "EOF",
	TokenNewline:        "Newline",
	TokenSpace:          "Space",
	TokenIdentifier:     "Identifier",
	TokenLeftParen:      "LeftParen",
	TokenRightParen:     "RightParen",
	TokenQuotedArg:      "QuotedArg",
	TokenUnquotedArg:    "UnquotedArg",
	TokenBracketArg:     "BracketArg",
	TokenLineComment:    "LineComment",
	TokenBracketComment: "BracketComment",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", int(t))
}

// Token represents a lexical token with its position in the source.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Value, t.Line, t.Col)
}
