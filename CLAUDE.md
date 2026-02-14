# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

cmakefmt is a fast, opinionated CMake formatter written in Go. Single binary, zero external dependencies. Designed for editor integration (stdin/stdout mode) and CI checks (--check mode with exit codes).

## Build & Development Commands

```bash
make build          # Build binary to bin/cmakefmt (injects version/commit via ldflags)
make test           # Run all tests with race detector
make lint           # Run golangci-lint (Uber Go Style config in .golangci.yml)
make clean          # Remove bin/ directory
make install        # Install binary to $GOPATH/bin
make fmt            # Format testdata/ with cmakefmt (dogfooding)
make check          # Check testdata/ formatting (dry run)
```

Run a single test:
```bash
go test -race -v ./internal/lexer/ -run TestTokenizeSimpleCommand
```

## Architecture

The formatter uses a **Lexer → Parser → Rules → Formatter** pipeline:

1. **Lexer** (`internal/lexer/`) — Tokenizes CMake source into a typed token stream. Handles all CMake constructs: quoted/unquoted/bracket arguments, line/bracket comments, and identifies commands by lookahead for `(`. Contains the keyword list (PUBLIC, PRIVATE, etc.) and block opener/closer detection (if/endif, foreach/endforeach, etc.).

2. **Parser** (`internal/parser/`) — Recursive descent parser building an AST from the token stream. AST nodes: `File`, `CommandInvocation`, `Argument`, `Comment`, `BlankLine`, `BlockNode` (nested if/foreach/while/function/macro blocks with middles for else/elseif).

3. **Rules** (`internal/formatter/rules/`) — Extensible rule system using a registry pattern (rules self-register via `init()`). Each rule implements `Name()`, `Description()`, and `Apply(node, config)` to transform the AST in place. Current rules: `command_case`, `keyword_case`.

4. **Formatter** (`internal/formatter/`) — Emits formatted source from the AST. Handles indentation, single-line vs multi-line decisions based on line length, dangling parentheses, trailing whitespace removal, blank line limiting, and line ending normalization.

5. **CLI** (`cmd/cmakefmt/main.go`) — Entry point handling three modes: stdin/stdout (no args, for editors), file arguments, and directory recursion. Supports `--check`, `--diff`, `-i` (in-place), `--config`, and rule override flags.

## Configuration

Formatter config is defined in `internal/config/config.go`. Config files (`.cmakefmt`) are discovered by walking up from the working directory. Supports JSON and simple YAML-like key-value format. Key defaults: 80 char line length, 2-space indent, lowercase commands, uppercase keywords, dangling parentheses enabled.

## Code Style

- Follows **Uber Go Style Guide** — enforced by golangci-lint with strict settings (gocyclo max 15, funlen max 100 lines, gofumpt formatting)
- Zero external Go dependencies (stdlib only)
- Go 1.25.7
