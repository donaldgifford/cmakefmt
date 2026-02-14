package lexer

import (
	"testing"
)

func TestTokenizeSimpleCommand(t *testing.T) {
	input := `set(FOO "bar")`
	lex := New(input)
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []TokenType{
		TokenIdentifier,  // set
		TokenLeftParen,    // (
		TokenUnquotedArg,  // FOO
		TokenSpace,        // " "
		TokenQuotedArg,    // "bar"
		TokenRightParen,   // )
		TokenEOF,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token %d: expected %s, got %s (%q)", i, expected[i], tok.Type, tok.Value)
		}
	}
}

func TestTokenizeComment(t *testing.T) {
	input := "# this is a comment\nset(X Y)\n"
	lex := New(input)
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TokenLineComment {
		t.Errorf("expected LineComment, got %s", tokens[0].Type)
	}
	if tokens[0].Value != "# this is a comment" {
		t.Errorf("expected comment value, got %q", tokens[0].Value)
	}
}

func TestTokenizeBracketArgument(t *testing.T) {
	input := `set(X [=[hello world]=])`
	lex := New(input)
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the bracket arg token.
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenBracketArg {
			found = true
			if tok.Value != "[=[hello world]=]" {
				t.Errorf("expected bracket arg value, got %q", tok.Value)
			}
		}
	}
	if !found {
		t.Error("no bracket argument token found")
	}
}

func TestTokenizeMultiLineCommand(t *testing.T) {
	input := "target_link_libraries(\n  mylib\n  PUBLIC\n  fmt::fmt\n)\n"
	lex := New(input)
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TokenIdentifier {
		t.Errorf("expected Identifier, got %s", tokens[0].Type)
	}
	if tokens[0].Value != "target_link_libraries" {
		t.Errorf("expected 'target_link_libraries', got %q", tokens[0].Value)
	}
}

func TestIsBlockOpener(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"if", true},
		{"IF", true},
		{"foreach", true},
		{"while", true},
		{"function", true},
		{"macro", true},
		{"set", false},
		{"endif", false},
	}

	for _, tt := range tests {
		got := IsBlockOpener(tt.input)
		if got != tt.want {
			t.Errorf("IsBlockOpener(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsBlockCloser(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"endif", true},
		{"ENDIF", true},
		{"endforeach", true},
		{"endwhile", true},
		{"endfunction", true},
		{"endmacro", true},
		{"if", false},
		{"set", false},
	}

	for _, tt := range tests {
		got := IsBlockCloser(tt.input)
		if got != tt.want {
			t.Errorf("IsBlockCloser(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
