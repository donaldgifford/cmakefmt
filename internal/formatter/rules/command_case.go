package rules

import (
	"strings"

	"github.com/donaldgifford/cmakefmt/internal/config"
	"github.com/donaldgifford/cmakefmt/internal/parser"
)

func init() {
	Register(&CommandCaseRule{})
}

// CommandCaseRule normalizes command name casing (e.g., SET -> set).
type CommandCaseRule struct{}

func (r *CommandCaseRule) Name() string        { return "command_case" }
func (r *CommandCaseRule) Description() string  { return "Normalize command name casing" }

func (r *CommandCaseRule) Apply(node parser.Node, cfg *config.Config) error {
	switch n := node.(type) {
	case *parser.File:
		for _, el := range n.Elements {
			if err := r.Apply(el, cfg); err != nil {
				return err
			}
		}
	case *parser.CommandInvocation:
		n.Name = applyCase(n.Name, cfg.CommandCase)
	case *parser.BlockNode:
		if n.Opener != nil {
			n.Opener.Name = applyCase(n.Opener.Name, cfg.CommandCase)
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
				mid.Command.Name = applyCase(mid.Command.Name, cfg.CommandCase)
			}
			for _, el := range mid.Body {
				if err := r.Apply(el, cfg); err != nil {
					return err
				}
			}
		}
		if n.Closer != nil {
			n.Closer.Name = applyCase(n.Closer.Name, cfg.CommandCase)
		}
	}
	return nil
}

func applyCase(s, caseStyle string) string {
	switch caseStyle {
	case "lower":
		return strings.ToLower(s)
	case "upper":
		return strings.ToUpper(s)
	default:
		return s
	}
}
