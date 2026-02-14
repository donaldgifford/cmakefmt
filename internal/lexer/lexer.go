package lexer

import (
	"strings"
	"unicode"
)

// Lexer tokenizes CMake source code.
type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	tokens []Token
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{
		input: input,
		line:  1,
		col:   1,
	}
}

// Tokenize processes the entire input and returns all tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		if err := l.next(); err != nil {
			return nil, err
		}
	}
	l.emit(TokenEOF, "")
	return l.tokens, nil
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekAt(offset int) byte {
	idx := l.pos + offset
	if idx >= len(l.input) {
		return 0
	}
	return l.input[idx]
}

func (l *Lexer) advance() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) emit(typ TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:  typ,
		Value: value,
		Line:  l.line,
		Col:   l.col,
	})
}

func (l *Lexer) emitAt(typ TokenType, value string, line, col int) {
	l.tokens = append(l.tokens, Token{
		Type:  typ,
		Value: value,
		Line:  line,
		Col:   col,
	})
}

func (l *Lexer) next() error {
	ch := l.peek()

	switch {
	case ch == '\n':
		line, col := l.line, l.col
		l.advance()
		l.emitAt(TokenNewline, "\n", line, col)
	case ch == '\r' && l.peekAt(1) == '\n':
		line, col := l.line, l.col
		l.advance()
		l.advance()
		l.emitAt(TokenNewline, "\r\n", line, col)
	case ch == '\r':
		line, col := l.line, l.col
		l.advance()
		l.emitAt(TokenNewline, "\r", line, col)
	case ch == ' ' || ch == '\t':
		l.lexSpace()
	case ch == '#':
		return l.lexComment()
	case ch == '(':
		line, col := l.line, l.col
		l.advance()
		l.emitAt(TokenLeftParen, "(", line, col)
	case ch == ')':
		line, col := l.line, l.col
		l.advance()
		l.emitAt(TokenRightParen, ")", line, col)
	case ch == '"':
		return l.lexQuotedArg()
	case ch == '[' && l.isBracketStart():
		return l.lexBracketArg()
	case isIdentStart(ch):
		l.lexIdentifierOrUnquoted()
	default:
		l.lexUnquotedArg()
	}
	return nil
}

func (l *Lexer) lexSpace() {
	line, col := l.line, l.col
	start := l.pos
	for l.pos < len(l.input) && (l.peek() == ' ' || l.peek() == '\t') {
		l.advance()
	}
	l.emitAt(TokenSpace, l.input[start:l.pos], line, col)
}

func (l *Lexer) lexComment() error {
	line, col := l.line, l.col
	// Check for bracket comment: #[=*[
	if l.peekAt(1) == '[' {
		eqCount := 0
		i := 2
		for l.peekAt(i) == '=' {
			eqCount++
			i++
		}
		if l.peekAt(i) == '[' {
			return l.lexBracketComment(eqCount, line, col)
		}
	}

	// Line comment
	start := l.pos
	for l.pos < len(l.input) && l.peek() != '\n' && l.peek() != '\r' {
		l.advance()
	}
	l.emitAt(TokenLineComment, l.input[start:l.pos], line, col)
	return nil
}

func (l *Lexer) lexBracketComment(eqCount int, line, col int) error {
	start := l.pos
	// Skip #[=*[
	l.advance() // #
	l.advance() // [
	for range eqCount {
		l.advance() // =
	}
	l.advance() // [

	closer := "]" + strings.Repeat("=", eqCount) + "]"
	for l.pos < len(l.input) {
		idx := strings.Index(l.input[l.pos:], closer)
		if idx < 0 {
			// Consume rest of input
			for l.pos < len(l.input) {
				l.advance()
			}
			l.emitAt(TokenBracketComment, l.input[start:l.pos], line, col)
			return nil
		}
		// Advance to end of closer
		for range idx + len(closer) {
			l.advance()
		}
		l.emitAt(TokenBracketComment, l.input[start:l.pos], line, col)
		return nil
	}
	l.emitAt(TokenBracketComment, l.input[start:l.pos], line, col)
	return nil
}

func (l *Lexer) lexQuotedArg() error {
	line, col := l.line, l.col
	start := l.pos
	l.advance() // opening "

	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == '\\' {
			l.advance() // backslash
			l.advance() // escaped char
			continue
		}
		if ch == '"' {
			l.advance() // closing "
			l.emitAt(TokenQuotedArg, l.input[start:l.pos], line, col)
			return nil
		}
		l.advance()
	}
	// Unterminated quoted arg — emit what we have
	l.emitAt(TokenQuotedArg, l.input[start:l.pos], line, col)
	return nil
}

func (l *Lexer) isBracketStart() bool {
	i := 1
	for l.peekAt(i) == '=' {
		i++
	}
	return l.peekAt(i) == '['
}

func (l *Lexer) lexBracketArg() error {
	line, col := l.line, l.col
	start := l.pos
	l.advance() // [
	eqCount := 0
	for l.peek() == '=' {
		eqCount++
		l.advance()
	}
	l.advance() // [

	closer := "]" + strings.Repeat("=", eqCount) + "]"
	for l.pos < len(l.input) {
		idx := strings.Index(l.input[l.pos:], closer)
		if idx < 0 {
			for l.pos < len(l.input) {
				l.advance()
			}
			l.emitAt(TokenBracketArg, l.input[start:l.pos], line, col)
			return nil
		}
		for range idx + len(closer) {
			l.advance()
		}
		l.emitAt(TokenBracketArg, l.input[start:l.pos], line, col)
		return nil
	}
	l.emitAt(TokenBracketArg, l.input[start:l.pos], line, col)
	return nil
}

func (l *Lexer) lexIdentifierOrUnquoted() {
	line, col := l.line, l.col
	start := l.pos

	// First consume identifier-like chars.
	for l.pos < len(l.input) && isIdentChar(l.peek()) {
		l.advance()
	}

	// Check if what follows (skipping space) is '(' — if so, this is a command name.
	saved := l.pos
	for saved < len(l.input) && (l.input[saved] == ' ' || l.input[saved] == '\t') {
		saved++
	}
	if saved < len(l.input) && l.input[saved] == '(' {
		l.emitAt(TokenIdentifier, l.input[start:l.pos], line, col)
		return
	}

	// Not a command — continue consuming as unquoted argument.
	// Unquoted args can contain /, ., -, +, $, {, }, <, >, @, etc.
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' ||
			ch == '(' || ch == ')' || ch == '#' || ch == '"' {
			break
		}
		l.advance()
	}

	l.emitAt(TokenUnquotedArg, l.input[start:l.pos], line, col)
}

func (l *Lexer) lexUnquotedArg() {
	line, col := l.line, l.col
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' ||
			ch == '(' || ch == ')' || ch == '#' || ch == '"' {
			break
		}
		l.advance()
	}
	if l.pos > start {
		l.emitAt(TokenUnquotedArg, l.input[start:l.pos], line, col)
	}
}

func isIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// IsKeyword returns true if the given string is a CMake keyword used in
// command arguments (e.g., PUBLIC, PRIVATE, INTERFACE, REQUIRED, etc.).
func IsKeyword(s string) bool {
	upper := strings.ToUpper(s)
	_, ok := cmakeKeywords[upper]
	return ok
}

// IsBlockOpener returns true if the command name opens a block
// (e.g., if, foreach, while, function, macro, block).
func IsBlockOpener(cmd string) bool {
	lower := strings.ToLower(cmd)
	_, ok := blockOpeners[lower]
	return ok
}

// IsBlockCloser returns true if the command name closes a block
// (e.g., endif, endforeach, endwhile, endfunction, endmacro, endblock).
func IsBlockCloser(cmd string) bool {
	lower := strings.ToLower(cmd)
	_, ok := blockClosers[lower]
	return ok
}

// IsBlockMiddle returns true if the command is a mid-block keyword
// (e.g., else, elseif).
func IsBlockMiddle(cmd string) bool {
	lower := strings.ToLower(cmd)
	return lower == "else" || lower == "elseif"
}

// IsCMakeIdentifier returns true if the rune is valid in a CMake identifier.
func IsCMakeIdentifier(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

var blockOpeners = map[string]bool{
	"if": true, "foreach": true, "while": true,
	"function": true, "macro": true, "block": true,
}

var blockClosers = map[string]bool{
	"endif": true, "endforeach": true, "endwhile": true,
	"endfunction": true, "endmacro": true, "endblock": true,
}

var cmakeKeywords = map[string]bool{
	"PUBLIC": true, "PRIVATE": true, "INTERFACE": true,
	"REQUIRED": true, "COMPONENTS": true, "OPTIONAL_COMPONENTS": true,
	"CONFIG": true, "MODULE": true, "NO_MODULE": true,
	"QUIET": true, "EXACT": true,
	"DESTINATION": true, "RENAME": true, "PERMISSIONS": true,
	"TARGETS": true, "FILES": true, "PROGRAMS": true, "DIRECTORY": true,
	"EXPORT": true, "NAMESPACE": true, "FILE": true,
	"PROPERTIES": true, "PROPERTY": true,
	"IMPORTED": true, "GLOBAL": true, "ALIAS": true,
	"STATIC": true, "SHARED": true, "MODULE_LIBRARY": true, "OBJECT": true,
	"COMMAND": true, "DEPENDS": true, "WORKING_DIRECTORY": true,
	"COMMENT": true, "VERBATIM": true,
	"APPEND": true, "PARENT_SCOPE": true, "CACHE": true, "FORCE": true,
	"BOOL": true, "STRING": true, "FILEPATH": true, "PATH": true, "INTERNAL": true,
}
