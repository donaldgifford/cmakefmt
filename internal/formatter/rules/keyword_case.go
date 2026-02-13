package rules

import (
	"strings"

	"github.com/donaldgifford/cmakefmt/internal/config"
	"github.com/donaldgifford/cmakefmt/internal/lexer"
	"github.com/donaldgifford/cmakefmt/internal/parser"
)

func init() {
	Register(&KeywordCaseRule{})
}

// KeywordCaseRule normalizes keyword argument casing (e.g., public -> PUBLIC).
type KeywordCaseRule struct{}

func (r *KeywordCaseRule) Name() string        { return "keyword_case" }
func (r *KeywordCaseRule) Description() string  { return "Normalize keyword argument casing" }

func (r *KeywordCaseRule) Apply(node parser.Node, cfg *config.Config) error {
	switch n := node.(type) {
	case *parser.File:
		for _, el := range n.Elements {
			if err := r.Apply(el, cfg); err != nil {
				return err
			}
		}
	case *parser.CommandInvocation:
		for i, arg := range n.Arguments {
			if arg.Token.Type == lexer.TokenUnquotedArg && lexer.IsKeyword(arg.Token.Value) {
				n.Arguments[i].Token.Value = applyCase(arg.Token.Value, cfg.KeywordCase)
			}
		}
	case *parser.BlockNode:
		if n.Opener != nil {
			if err := r.Apply(n.Opener, cfg); err != nil {
				return err
			}
		}
		for _, el := range n.Body {
			if err := r.Apply(el, cfg); err != nil {
				return err
			}
		}
		for _, mid := range n.Middles {
			if mid.Command != nil {
				if err := r.Apply(mid.Command, cfg); err != nil {
					return err
				}
			}
			for _, el := range mid.Body {
				if err := r.Apply(el, cfg); err != nil {
					return err
				}
			}
		}
		if n.Closer != nil {
			if err := r.Apply(n.Closer, cfg); err != nil {
				return err
			}
		}
	}
	return nil
}

// isKeywordLike checks if a string looks like a CMake keyword
// (all uppercase letters and underscores, or a known keyword).
func isKeywordLike(s string) bool {
	if lexer.IsKeyword(s) {
		return true
	}
	// Heuristic: all uppercase letters/digits/underscores, at least 2 chars.
	if len(s) < 2 {
		return false
	}
	for _, c := range s {
		if c != '_' && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	// Exclude things that look like variable references.
	return !strings.HasPrefix(s, "${") && !strings.HasPrefix(s, "$<")
}
