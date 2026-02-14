package formatter

import (
	"fmt"
	"strings"

	"github.com/donaldgifford/cmakefmt/internal/config"
	"github.com/donaldgifford/cmakefmt/internal/formatter/rules"
	"github.com/donaldgifford/cmakefmt/internal/lexer"
	"github.com/donaldgifford/cmakefmt/internal/parser"
)

// Formatter formats CMake source code.
type Formatter struct {
	Config *config.Config
}

// New creates a new Formatter with the given configuration.
func New(cfg *config.Config) *Formatter {
	return &Formatter{Config: cfg}
}

// Format takes CMake source and returns formatted output.
func (f *Formatter) Format(input []byte) ([]byte, error) {
	// Tokenize.
	lex := lexer.New(string(input))
	tokens, err := lex.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	// Parse.
	p := parser.New(tokens)
	file, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}

	// Apply rules.
	enabledRules := rules.EnabledRules(f.Config)
	for _, rule := range enabledRules {
		if err := rule.Apply(file, f.Config); err != nil {
			return nil, fmt.Errorf("rule %q error: %w", rule.Name(), err)
		}
	}

	// Emit formatted output.
	var buf strings.Builder
	f.emitFile(&buf, file, 0)

	result := buf.String()

	// Apply trailing newline.
	if f.Config.TrailingNewline && !strings.HasSuffix(result, "\n") {
		result += f.newline()
	}

	// Apply line ending normalization.
	result = f.normalizeLineEndings(result)

	// Apply trailing whitespace removal.
	if f.Config.TrimTrailingWhitespace {
		result = trimTrailingWhitespace(result)
	}

	return []byte(result), nil
}

func (f *Formatter) emitFile(buf *strings.Builder, file *parser.File, depth int) {
	prevWasBlank := false
	for _, el := range file.Elements {
		switch n := el.(type) {
		case *parser.CommandInvocation:
			prevWasBlank = false
			f.emitCommand(buf, n, depth)
			buf.WriteString(f.newline())
		case *parser.Comment:
			prevWasBlank = false
			f.emitIndent(buf, depth)
			buf.WriteString(n.Value)
			buf.WriteString(f.newline())
		case *parser.BlankLine:
			count := n.Count
			if count > f.Config.MaxBlankLines {
				count = f.Config.MaxBlankLines
			}
			if !prevWasBlank {
				for range count {
					buf.WriteString(f.newline())
				}
				prevWasBlank = true
			}
		case *parser.BlockNode:
			prevWasBlank = false
			f.emitBlock(buf, n, depth)
		}
	}
}

func (f *Formatter) emitBlock(buf *strings.Builder, block *parser.BlockNode, depth int) {
	// Emit opener.
	if block.Opener != nil {
		f.emitCommand(buf, block.Opener, depth)
		buf.WriteString(f.newline())
	}

	// Emit body (indented).
	f.emitElements(buf, block.Body, depth+1)

	// Emit middles (else/elseif).
	for _, mid := range block.Middles {
		if mid.Command != nil {
			f.emitCommand(buf, mid.Command, depth)
			buf.WriteString(f.newline())
		}
		f.emitElements(buf, mid.Body, depth+1)
	}

	// Emit closer.
	if block.Closer != nil {
		f.emitCommand(buf, block.Closer, depth)
		buf.WriteString(f.newline())
	}
}

func (f *Formatter) emitElements(buf *strings.Builder, elements []parser.Node, depth int) {
	prevWasBlank := false
	for _, el := range elements {
		switch n := el.(type) {
		case *parser.CommandInvocation:
			prevWasBlank = false
			f.emitCommand(buf, n, depth)
			buf.WriteString(f.newline())
		case *parser.Comment:
			prevWasBlank = false
			f.emitIndent(buf, depth)
			buf.WriteString(n.Value)
			buf.WriteString(f.newline())
		case *parser.BlankLine:
			count := n.Count
			if count > f.Config.MaxBlankLines {
				count = f.Config.MaxBlankLines
			}
			if !prevWasBlank {
				for range count {
					buf.WriteString(f.newline())
				}
				prevWasBlank = true
			}
		case *parser.BlockNode:
			prevWasBlank = false
			f.emitBlock(buf, n, depth)
		}
	}
}

func (f *Formatter) emitCommand(buf *strings.Builder, cmd *parser.CommandInvocation, depth int) {
	f.emitIndent(buf, depth)
	buf.WriteString(cmd.Name)

	if f.Config.SpaceBeforeParen {
		buf.WriteString(" ")
	}
	buf.WriteString("(")

	args := filterNonCommentArgs(cmd.Arguments)
	commentArgs := filterCommentArgs(cmd.Arguments)

	if len(args) == 0 {
		buf.WriteString(")")
		f.emitTrailingComment(buf, cmd)
		return
	}

	// Determine if we need to wrap.
	singleLine := f.formatArgsSingleLine(args)
	indentStr := f.indentString(depth)
	commandWidth := len(indentStr) + len(cmd.Name) + 1 + len(singleLine) + 1 // +1 for ( and )

	if commandWidth <= f.Config.LineLength && len(commentArgs) == 0 {
		// Single line.
		buf.WriteString(singleLine)
		buf.WriteString(")")
	} else {
		// Multi-line.
		f.emitArgsMultiLine(buf, cmd.Arguments, depth)
		if f.Config.DanglingParenthesis {
			buf.WriteString(f.newline())
			f.emitIndent(buf, depth)
			buf.WriteString(")")
		} else {
			buf.WriteString(")")
		}
	}

	f.emitTrailingComment(buf, cmd)
}

func (f *Formatter) emitArgsMultiLine(buf *strings.Builder, args []parser.Argument, depth int) {
	for _, arg := range args {
		buf.WriteString(f.newline())
		f.emitIndent(buf, depth+1)
		if arg.Token.Type == lexer.TokenLineComment {
			buf.WriteString(arg.Token.Value)
		} else if len(arg.Children) > 0 {
			// Parenthesized group — flatten for now.
			buf.WriteString("(")
			for i, child := range arg.Children {
				if i > 0 {
					buf.WriteString(" ")
				}
				buf.WriteString(child.Token.Value)
			}
			buf.WriteString(")")
		} else {
			buf.WriteString(arg.Token.Value)
		}
	}
}

func (f *Formatter) formatArgsSingleLine(args []parser.Argument) string {
	var parts []string
	for _, arg := range args {
		if len(arg.Children) > 0 {
			var inner []string
			for _, child := range arg.Children {
				inner = append(inner, child.Token.Value)
			}
			parts = append(parts, "("+strings.Join(inner, " ")+")")
		} else {
			parts = append(parts, arg.Token.Value)
		}
	}
	return strings.Join(parts, " ")
}

func (f *Formatter) emitTrailingComment(buf *strings.Builder, cmd *parser.CommandInvocation) {
	if cmd.TrailingComment != "" {
		buf.WriteString(" ")
		buf.WriteString(cmd.TrailingComment)
	}
}

func (f *Formatter) emitIndent(buf *strings.Builder, depth int) {
	if depth <= 0 {
		return
	}
	if f.Config.UseTabs {
		for range depth {
			buf.WriteByte('\t')
		}
	} else {
		for range depth * f.Config.Indent {
			buf.WriteByte(' ')
		}
	}
}

func (f *Formatter) indentString(depth int) string {
	if depth <= 0 {
		return ""
	}
	if f.Config.UseTabs {
		return strings.Repeat("\t", depth)
	}
	return strings.Repeat(" ", depth*f.Config.Indent)
}

func (f *Formatter) newline() string {
	switch f.Config.LineEnding {
	case "crlf":
		return "\r\n"
	default:
		return "\n"
	}
}

func (f *Formatter) normalizeLineEndings(s string) string {
	switch f.Config.LineEnding {
	case "crlf":
		// First normalize to LF, then convert to CRLF.
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		s = strings.ReplaceAll(s, "\n", "\r\n")
	case "lf":
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
	}
	return s
}

func trimTrailingWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

func filterNonCommentArgs(args []parser.Argument) []parser.Argument {
	var result []parser.Argument
	for _, a := range args {
		if a.Token.Type != lexer.TokenLineComment {
			result = append(result, a)
		}
	}
	return result
}

func filterCommentArgs(args []parser.Argument) []parser.Argument {
	var result []parser.Argument
	for _, a := range args {
		if a.Token.Type == lexer.TokenLineComment {
			result = append(result, a)
		}
	}
	return result
}
