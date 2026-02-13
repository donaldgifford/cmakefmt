package rules

import (
	"github.com/donaldgifford/cmakefmt/internal/config"
	"github.com/donaldgifford/cmakefmt/internal/parser"
)

// Rule is the interface that all formatting rules implement.
// Adding a new rule is as simple as:
//  1. Create a new file in this package.
//  2. Implement the Rule interface.
//  3. Register it in the registry via init().
type Rule interface {
	// Name returns the unique identifier for this rule (e.g., "indent", "command_case").
	Name() string
	// Description returns a human-readable description of what this rule does.
	Description() string
	// Apply transforms the AST node in place according to this rule.
	Apply(node parser.Node, cfg *config.Config) error
}

// registry holds all registered rules.
var registry []Rule

// Register adds a rule to the global registry.
// Call this from init() in each rule file.
func Register(r Rule) {
	registry = append(registry, r)
}

// All returns all registered rules.
func All() []Rule {
	return registry
}

// ByName returns a rule by its name, or nil if not found.
func ByName(name string) Rule {
	for _, r := range registry {
		if r.Name() == name {
			return r
		}
	}
	return nil
}

// EnabledRules returns only the rules that are enabled in the given config.
func EnabledRules(cfg *config.Config) []Rule {
	var enabled []Rule
	for _, r := range registry {
		rc, exists := cfg.Rules[r.Name()]
		if exists && !rc.IsEnabled() {
			continue
		}
		enabled = append(enabled, r)
	}
	return enabled
}
