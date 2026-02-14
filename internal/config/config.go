package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// DefaultConfigFile is the default config file name.
	DefaultConfigFile = ".cmakefmt"
)

// Config holds all formatting configuration.
type Config struct {
	// LineLength is the maximum line length before wrapping.
	LineLength int `json:"line_length"`
	// Indent is the number of spaces per indentation level.
	Indent int `json:"indent"`
	// TabSize is the tab width (used if UseTabs is true).
	TabSize int `json:"tab_size"`
	// UseTabs uses tabs instead of spaces for indentation.
	UseTabs bool `json:"use_tabs"`
	// CommandCase controls the casing of command names.
	// Values: "lower", "upper", "unchanged".
	CommandCase string `json:"command_case"`
	// KeywordCase controls the casing of keyword arguments.
	// Values: "upper", "lower", "unchanged".
	KeywordCase string `json:"keyword_case"`
	// DanglingParenthesis puts the closing ')' on its own line
	// when arguments span multiple lines.
	DanglingParenthesis bool `json:"dangling_parenthesis"`
	// TrailingNewline ensures the file ends with a newline.
	TrailingNewline bool `json:"trailing_newline"`
	// TrimTrailingWhitespace removes trailing whitespace from lines.
	TrimTrailingWhitespace bool `json:"trim_trailing_whitespace"`
	// MaxBlankLines is the maximum number of consecutive blank lines to keep.
	MaxBlankLines int `json:"max_blank_lines"`
	// SpaceBeforeParen inserts a space before '(' in command invocations.
	SpaceBeforeParen bool `json:"space_before_paren"`
	// LineEnding controls line ending style: "lf", "crlf", or "auto".
	LineEnding string `json:"line_ending"`

	// Rules enables or disables specific formatting rules.
	Rules map[string]RuleConfig `json:"rules"`
}

// RuleConfig holds per-rule configuration options.
type RuleConfig struct {
	Enabled *bool                  `json:"enabled"`
	Options map[string]interface{} `json:"options"`
}

// IsEnabled returns whether a rule is enabled (default: true).
func (rc RuleConfig) IsEnabled() bool {
	if rc.Enabled == nil {
		return true
	}
	return *rc.Enabled
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		LineLength:             80,
		Indent:                 2,
		TabSize:                4,
		UseTabs:                false,
		CommandCase:            "lower",
		KeywordCase:            "upper",
		DanglingParenthesis:    true,
		TrailingNewline:        true,
		TrimTrailingWhitespace: true,
		MaxBlankLines:          2,
		SpaceBeforeParen:       false,
		LineEnding:             "lf",
		Rules:                  make(map[string]RuleConfig),
	}
}

// Load reads configuration from a file. Supports both JSON (.json) and
// a simple YAML-like format (.cmakefmt). If the file doesn't exist, returns defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	content := strings.TrimSpace(string(data))

	// Try JSON first.
	if strings.HasPrefix(content, "{") {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config as JSON: %w", err)
		}
		return cfg, nil
	}

	// Otherwise, parse as simple YAML-like key: value format.
	if err := parseSimpleConfig(content, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// parseSimpleConfig parses a flat YAML-like config file.
// Supports: key: value lines, comments (#), and blank lines.
// This avoids requiring an external YAML dependency.
func parseSimpleConfig(content string, cfg *Config) error {
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip nested sections (rules: etc.) for now.
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Skip if value is empty (section header like "rules:").
		if value == "" {
			continue
		}

		switch key {
		case "line_length":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.LineLength = v
			}
		case "indent":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.Indent = v
			}
		case "tab_size":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.TabSize = v
			}
		case "use_tabs":
			cfg.UseTabs = parseBool(value)
		case "command_case":
			cfg.CommandCase = value
		case "keyword_case":
			cfg.KeywordCase = value
		case "dangling_parenthesis":
			cfg.DanglingParenthesis = parseBool(value)
		case "trailing_newline":
			cfg.TrailingNewline = parseBool(value)
		case "trim_trailing_whitespace":
			cfg.TrimTrailingWhitespace = parseBool(value)
		case "max_blank_lines":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.MaxBlankLines = v
			}
		case "space_before_paren":
			cfg.SpaceBeforeParen = parseBool(value)
		case "line_ending":
			cfg.LineEnding = value
		}
	}

	return scanner.Err()
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes" || s == "1"
}

// FindConfigFile searches for a config file by walking up from the given
// directory. Returns the path to the config file, or empty string if not found.
func FindConfigFile(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, DefaultConfigFile)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Merge applies non-zero values from the override config onto the base config.
func Merge(base, override *Config) *Config {
	result := *base
	if override.LineLength != 0 {
		result.LineLength = override.LineLength
	}
	if override.Indent != 0 {
		result.Indent = override.Indent
	}
	if override.CommandCase != "" {
		result.CommandCase = override.CommandCase
	}
	if override.KeywordCase != "" {
		result.KeywordCase = override.KeywordCase
	}
	return &result
}
