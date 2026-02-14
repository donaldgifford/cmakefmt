// Package main is the entry point for the cmakefmt
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/donaldgifford/cmakefmt/internal/config"
	"github.com/donaldgifford/cmakefmt/internal/formatter"
	"github.com/donaldgifford/cmakefmt/internal/formatter/rules"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	// Flags.
	checkFlag := flag.Bool("check", false, "Check if files are formatted (exit 1 if not)")
	diffFlag := flag.Bool("diff", false, "Show diff of formatting changes")
	inPlaceFlag := flag.Bool("i", false, "Format files in-place")
	configPath := flag.String("config", "", "Path to config file (default: search for .cmakefmt)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	listRulesFlag := flag.Bool("list-rules", false, "List all available formatting rules")

	// Config overrides.
	lineLength := flag.Int("line-length", 0, "Override max line length")
	indent := flag.Int("indent", 0, "Override indentation width")
	commandCase := flag.String("command-case", "", "Override command casing (lower, upper, unchanged)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cmakefmt [flags] [files...]\n\n")
		fmt.Fprintf(os.Stderr, "A formatter for CMake files.\n\n")
		fmt.Fprintf(os.Stderr, "With no file arguments, reads from stdin and writes to stdout.\n")
		fmt.Fprintf(os.Stderr, "This mode is designed for use with editor integrations like conform.nvim.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  cmakefmt CMakeLists.txt              # Print formatted output to stdout\n")
		fmt.Fprintf(os.Stderr, "  cmakefmt -i CMakeLists.txt           # Format in-place\n")
		fmt.Fprintf(os.Stderr, "  cmakefmt -check .                    # Check all CMake files (for CI)\n")
		fmt.Fprintf(os.Stderr, "  cmakefmt < CMakeLists.txt            # Read from stdin (for editors)\n")
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("cmakefmt %s (%s)\n", version, commit)
		os.Exit(0)
	}

	if *listRulesFlag {
		listRules()
		os.Exit(0)
	}

	// Load config.
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fatal("Error loading config: %v", err)
	}

	// Apply CLI overrides.
	if *lineLength > 0 {
		cfg.LineLength = *lineLength
	}
	if *indent > 0 {
		cfg.Indent = *indent
	}
	if *commandCase != "" {
		cfg.CommandCase = *commandCase
	}

	fmtr := formatter.New(cfg)

	args := flag.Args()

	// No file arguments: stdin/stdout mode (for conform.nvim and editors).
	if len(args) == 0 {
		if err := formatStdin(fmtr); err != nil {
			fatal("Error: %v", err)
		}
		return
	}

	// Collect files to format.
	files, err := collectFiles(args)
	if err != nil {
		fatal("Error collecting files: %v", err)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No CMake files found")
		os.Exit(0)
	}

	// Process files.
	hasChanges := false
	hasErrors := false

	for _, file := range files {
		input, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			hasErrors = true
			continue
		}

		output, err := fmtr.Format(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", file, err)
			hasErrors = true
			continue
		}

		if string(input) == string(output) {
			continue
		}

		hasChanges = true

		if *checkFlag {
			fmt.Fprintf(os.Stderr, "%s: would be reformatted\n", file)
			if *diffFlag {
				printDiff(file, string(input), string(output))
			}
			continue
		}

		if *diffFlag {
			printDiff(file, string(input), string(output))
			continue
		}

		if *inPlaceFlag {
			if err := os.WriteFile(file, output, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", file, err)
				hasErrors = true
				continue
			}
			fmt.Fprintf(os.Stderr, "Formatted %s\n", file)
		} else {
			os.Stdout.Write(output)
		}
	}

	if hasErrors {
		os.Exit(2)
	}
	if *checkFlag && hasChanges {
		os.Exit(1)
	}
}

func formatStdin(fmtr *formatter.Formatter) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	output, err := fmtr.Format(input)
	if err != nil {
		return fmt.Errorf("formatting: %w", err)
	}

	_, err = os.Stdout.Write(output)
	return err
}

func loadConfig(path string) (*config.Config, error) {
	if path != "" {
		return config.Load(path)
	}

	// Search upward from cwd.
	cwd, err := os.Getwd()
	if err != nil {
		return config.DefaultConfig(), nil
	}

	found := config.FindConfigFile(cwd)
	if found != "" {
		return config.Load(found)
	}

	return config.DefaultConfig(), nil
}

func collectFiles(args []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			files = append(files, arg)
			continue
		}

		// Walk directory for CMake files.
		err = filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip hidden dirs and build dirs.
				base := filepath.Base(path)
				if strings.HasPrefix(base, ".") || base == "build" || base == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if isCMakeFile(path) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func isCMakeFile(path string) bool {
	base := filepath.Base(path)
	if base == "CMakeLists.txt" {
		return true
	}
	ext := filepath.Ext(path)
	return ext == ".cmake"
}

func printDiff(filename, original, formatted string) {
	origLines := strings.Split(original, "\n")
	fmtLines := strings.Split(formatted, "\n")

	fmt.Fprintf(os.Stderr, "--- %s (original)\n", filename)
	fmt.Fprintf(os.Stderr, "+++ %s (formatted)\n", filename)

	// Simple unified-ish diff output.
	maxLines := len(origLines)
	if len(fmtLines) > maxLines {
		maxLines = len(fmtLines)
	}

	for i := range maxLines {
		var origLine, fmtLine string
		if i < len(origLines) {
			origLine = origLines[i]
		}
		if i < len(fmtLines) {
			fmtLine = fmtLines[i]
		}

		if origLine != fmtLine {
			if i < len(origLines) {
				fmt.Fprintf(os.Stderr, "-%s\n", origLine)
			}
			if i < len(fmtLines) {
				fmt.Fprintf(os.Stderr, "+%s\n", fmtLine)
			}
		}
	}
	fmt.Fprintln(os.Stderr)
}

func listRules() {
	// Import to ensure init() runs.
	allRules := getAllRules()
	fmt.Println("Available formatting rules:")
	fmt.Println()
	for _, r := range allRules {
		fmt.Printf("  %-25s %s\n", r.name, r.description)
	}
}

type ruleInfo struct {
	name        string
	description string
}

func getAllRules() []ruleInfo {
	// We reference the rules package directly.
	var infos []ruleInfo
	for _, r := range rules.All() {
		infos = append(infos, ruleInfo{
			name:        r.Name(),
			description: r.Description(),
		})
	}
	return infos
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
