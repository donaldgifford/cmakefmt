package formatter

import (
	"strings"
	"testing"

	"github.com/donaldgifford/cmakefmt/internal/config"

	// Import rules to trigger init() registration.
	_ "github.com/donaldgifford/cmakefmt/internal/formatter/rules"
)

func TestFormatSimpleCommand(t *testing.T) {
	input := `SET(FOO "bar")`
	cfg := config.DefaultConfig()
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "set(") {
		t.Errorf("expected lowercase command, got:\n%s", result)
	}
}

func TestFormatPreservesComments(t *testing.T) {
	input := "# This is a comment\nset(FOO bar)\n"
	cfg := config.DefaultConfig()
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "# This is a comment") {
		t.Errorf("expected comment to be preserved, got:\n%s", result)
	}
}

func TestFormatBlockIndentation(t *testing.T) {
	input := `if(FOO)
set(BAR baz)
endif()
`
	cfg := config.DefaultConfig()
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := string(output)
	lines := strings.Split(result, "\n")

	// The set() command should be indented.
	found := false
	for _, line := range lines {
		if strings.Contains(line, "set(") {
			found = true
			if !strings.HasPrefix(line, "  ") {
				t.Errorf("expected indented set(), got: %q", line)
			}
		}
	}
	if !found {
		t.Errorf("set() command not found in output:\n%s", result)
	}
}

func TestFormatUpperCommandCase(t *testing.T) {
	input := `set(FOO "bar")`
	cfg := config.DefaultConfig()
	cfg.CommandCase = "upper"
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "SET(") {
		t.Errorf("expected uppercase command, got:\n%s", result)
	}
}

func TestFormatTrailingNewline(t *testing.T) {
	input := `set(FOO bar)`
	cfg := config.DefaultConfig()
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := string(output)
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestFormatKeywordCase(t *testing.T) {
	input := `target_link_libraries(mylib public fmt::fmt)`
	cfg := config.DefaultConfig()
	cfg.KeywordCase = "upper"
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "PUBLIC") {
		t.Errorf("expected uppercase PUBLIC keyword, got:\n%s", result)
	}
}

func TestCheckModeDetectsChanges(t *testing.T) {
	input := `SET(FOO "bar")`
	cfg := config.DefaultConfig()
	fmtr := New(cfg)

	output, err := fmtr.Format([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) == input {
		t.Error("expected formatting changes, got identical output")
	}
}
