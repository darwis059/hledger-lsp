package rules

import (
	"testing"

	"github.com/juev/hledger-lsp/internal/lsputil"
)

func TestSemanticTokens_Comment(t *testing.T) {
	input := "# top comment"
	tokens := SemanticTokens(input)
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1", len(tokens))
	}
	if tokens[0].TokenType != SemTokenComment {
		t.Errorf("type: got %v, want Comment", tokens[0].TokenType)
	}
	if tokens[0].Line != 0 {
		t.Errorf("line: got %d, want 0", tokens[0].Line)
	}
}

func TestSemanticTokens_Directive(t *testing.T) {
	input := "skip 1"
	tokens := SemanticTokens(input)
	found := false
	for _, tok := range tokens {
		if tok.TokenType == SemTokenDirective {
			found = true
			if tok.Line != 0 || tok.Col != 0 {
				t.Errorf("directive position: got (%d,%d), want (0,0)", tok.Line, tok.Col)
			}
		}
	}
	if !found {
		t.Error("no directive token found for directive")
	}
}

func TestSemanticTokens_IfKeyword(t *testing.T) {
	input := "if PATTERN"
	tokens := SemanticTokens(input)
	found := false
	for _, tok := range tokens {
		if tok.TokenType == SemTokenKeyword && tok.Col == 0 {
			found = true
		}
	}
	if !found {
		t.Error("no keyword token for 'if'")
	}
}

func TestSemanticTokens_FieldName(t *testing.T) {
	input := "  account1 expenses:food"
	tokens := SemanticTokens(input)
	found := false
	for _, tok := range tokens {
		if tok.TokenType == SemTokenParameter {
			found = true
			break
		}
	}
	if !found {
		t.Error("no parameter token for field name")
	}
}

func TestSemanticTokens_FieldValue(t *testing.T) {
	input := "  account1 expenses:food"
	tokens := SemanticTokens(input)
	found := false
	for _, tok := range tokens {
		if tok.TokenType == SemTokenString {
			found = true
			break
		}
	}
	if !found {
		t.Error("no string token for field value")
	}
}

func TestSemanticTokens_Regex(t *testing.T) {
	input := "if PATTERN"
	tokens := SemanticTokens(input)
	found := false
	for _, tok := range tokens {
		if tok.TokenType == SemTokenRegexp {
			found = true
			break
		}
	}
	if !found {
		t.Error("no regexp token for if pattern")
	}
}

func TestSemanticTokens_MultiLine(t *testing.T) {
	input := "# comment\nskip 1\nif PAT\n  account1 expenses:food"
	tokens := SemanticTokens(input)
	if len(tokens) == 0 {
		t.Error("expected tokens for multiline input")
	}
	// Verify line numbers are correct (0-indexed)
	for _, tok := range tokens {
		if tok.Line > 3 {
			t.Errorf("invalid line %d", tok.Line)
		}
	}
}

func TestSemanticTokens_UTF16Column(t *testing.T) {
	// Col must be a UTF-16 offset, not a raw byte offset.
	// "# Привет мир": col=0 (byte 0 == UTF-16 0); Length must be 12 UTF-16 units, not 21 bytes.
	input := "# Привет мир"
	tokens := SemanticTokens(input)
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1", len(tokens))
	}
	tok := tokens[0]
	if tok.Col != 0 {
		t.Errorf("col: got %d, want 0", tok.Col)
	}
	wantLen := uint32(lsputil.UTF16Len(input)) // 12 UTF-16 units, not 21 bytes
	if tok.Length != wantLen {
		t.Errorf("length: got %d, want %d (UTF-16 units)", tok.Length, wantLen)
	}
}
