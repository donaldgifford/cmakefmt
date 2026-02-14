# cmakefmt

A fast, opinionated CMake formatter written in Go. Single binary, zero
dependencies.

Designed to work as a format-on-save tool with editors (via `conform.nvim`,
etc.) and as a CI check.

## Installation

```bash
go install github.com/yourusername/cmakefmt/cmd/cmakefmt@latest
```

Or download a binary from
[Releases](https://github.com/yourusername/cmakefmt/releases).

## Usage

```bash
# Format a file (prints to stdout)
cmakefmt CMakeLists.txt

# Format in-place
cmakefmt -i CMakeLists.txt

# Format all CMake files in a directory
cmakefmt -i .

# Read from stdin, write to stdout (for editors)
cmakefmt < CMakeLists.txt

# Check formatting (exits 1 if changes needed — for CI)
cmakefmt -check .

# Check with diff output
cmakefmt -check -diff .

# List available rules
cmakefmt -list-rules
```

### CI Usage

```yaml
# GitHub Actions
- name: Check CMake formatting
  run: cmakefmt -check .
```

## Configuration

Create a `.cmakefmt` file in your project root:

```yaml
# Maximum line length before wrapping arguments.
line_length: 80

# Number of spaces per indentation level.
indent: 2

# Command name casing: "lower", "upper", or "unchanged".
command_case: lower

# Keyword argument casing: "upper", "lower", or "unchanged".
keyword_case: upper

# Put closing ')' on its own line when arguments span multiple lines.
dangling_parenthesis: true

# Ensure the file ends with a newline.
trailing_newline: true

# Remove trailing whitespace from lines.
trim_trailing_whitespace: true

# Maximum consecutive blank lines to preserve.
max_blank_lines: 2

# Line ending style: "lf", "crlf", or "auto".
line_ending: lf

# Per-rule configuration.
# rules:
#   command_case:
#     enabled: false
```

The config file is discovered by walking up from the working directory, similar
to `.editorconfig` or `.yamlfmt`.

## Editor Integration

### Neovim (conform.nvim + LazyVim)

```lua
-- In your LazyVim config (e.g., lua/plugins/formatting.lua)
return {
  {
    "stevearc/conform.nvim",
    opts = {
      formatters_by_ft = {
        cmake = { "cmakefmt" },
      },
      formatters = {
        cmakefmt = {
          command = "cmakefmt",
          stdin = true,
        },
      },
    },
  },
}
```

This works because `cmakefmt` reads from stdin and writes to stdout when invoked
with no file arguments — exactly the contract `conform.nvim` expects.

### VS Code

Use the "Run on Save" extension with:

```json
{
  "emeraldwalk.runonsave": {
    "commands": [
      {
        "match": "CMakeLists\\.txt$|\\.cmake$",
        "cmd": "cmakefmt -i ${file}"
      }
    ]
  }
}
```

## Adding New Rules

Rules follow a simple interface pattern inspired by
[checkmake](https://github.com/checkmake/checkmake):

```go
// In internal/formatter/rules/my_rule.go
package rules

func init() {
    Register(&MyRule{})
}

type MyRule struct{}

func (r *MyRule) Name() string        { return "my_rule" }
func (r *MyRule) Description() string { return "Does something useful" }
func (r *MyRule) Apply(node parser.Node, cfg *config.Config) error {
    // Transform the AST node.
    return nil
}
```

The rule is automatically registered via `init()` and can be enabled/disabled
through the config file.

## Architecture

```
cmakefmt/
├── cmd/cmakefmt/          # CLI entrypoint
├── internal/
│   ├── config/            # Config file loading (.cmakefmt)
│   ├── lexer/             # CMake tokenizer
│   ├── parser/            # CMake AST builder
│   └── formatter/         # Formatting engine
│       └── rules/         # Individual formatting rules (add new rules here)
├── testdata/              # Test fixtures
├── .cmakefmt              # Example config
├── .goreleaser.yaml       # Release automation
└── Makefile
```

The pipeline is: **Source → Lexer → Tokens → Parser → AST → Rules → Formatter →
Output**

## Acknowledgments

- [yamlfmt](https://github.com/google/yamlfmt) — Architecture inspiration
  (config, extensible formatters, CI mode)
- [checkmake](https://github.com/checkmake/checkmake) — Rule registration
  pattern
- [gersemi](https://github.com/BlankSpruce/gersemi) — CMake formatting
  heuristics reference
- [cmake-format](https://github.com/cheshirekow/cmake_format) — Original CMake
  formatter

## License

MIT
