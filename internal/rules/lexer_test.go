package rules

import (
	"testing"
)

func collectTokens(l *Lexer) []Token {
	var tokens []Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func tokenTypes(tokens []Token) []TokenType {
	types := make([]TokenType, len(tokens))
	for i, t := range tokens {
		types[i] = t.Type
	}
	return types
}

func TestLexer_Comment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "hash comment",
			input: "# this is a comment",
			want:  []TokenType{TokenComment, TokenEOF},
		},
		{
			name:  "semicolon comment",
			input: "; this is a comment",
			want:  []TokenType{TokenComment, TokenEOF},
		},
		{
			name:  "star comment",
			input: "* this is a comment",
			want:  []TokenType{TokenComment, TokenEOF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			got := tokenTypes(collectTokens(l))
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("token[%d]: got %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLexer_Directive(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTypes []TokenType
		wantVals  []string
	}{
		{
			name:      "skip with number",
			input:     "skip 1",
			wantTypes: []TokenType{TokenDirective, TokenText, TokenEOF},
			wantVals:  []string{"skip", "1", ""},
		},
		{
			name:      "skip without number",
			input:     "skip",
			wantTypes: []TokenType{TokenDirective, TokenEOF},
			wantVals:  []string{"skip", ""},
		},
		{
			name:      "fields directive",
			input:     "fields date, description, amount",
			wantTypes: []TokenType{TokenDirective, TokenText, TokenEOF},
			wantVals:  []string{"fields", "date, description, amount", ""},
		},
		{
			name:      "separator directive",
			input:     "separator ,",
			wantTypes: []TokenType{TokenDirective, TokenText, TokenEOF},
			wantVals:  []string{"separator", ",", ""},
		},
		{
			name:      "date-format directive",
			input:     "date-format %Y-%m-%d",
			wantTypes: []TokenType{TokenDirective, TokenText, TokenEOF},
			wantVals:  []string{"date-format", "%Y-%m-%d", ""},
		},
		{
			name:      "newest-first flag directive",
			input:     "newest-first",
			wantTypes: []TokenType{TokenDirective, TokenEOF},
			wantVals:  []string{"newest-first", ""},
		},
		{
			name:      "include directive",
			input:     "include common.rules",
			wantTypes: []TokenType{TokenDirective, TokenText, TokenEOF},
			wantVals:  []string{"include", "common.rules", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tokens := collectTokens(l)
			types := tokenTypes(tokens)
			if len(types) != len(tt.wantTypes) {
				t.Fatalf("got types %v, want %v", types, tt.wantTypes)
			}
			for i := range types {
				if types[i] != tt.wantTypes[i] {
					t.Errorf("type[%d]: got %v, want %v", i, types[i], tt.wantTypes[i])
				}
				if tt.wantVals[i] != "" && tokens[i].Value != tt.wantVals[i] {
					t.Errorf("value[%d]: got %q, want %q", i, tokens[i].Value, tt.wantVals[i])
				}
			}
		})
	}
}

func TestLexer_IfBlock(t *testing.T) {
	input := "if PATTERN\n  account1 expenses:food"
	l := NewLexer(input)
	tokens := collectTokens(l)
	types := tokenTypes(tokens)

	// Expect: IfKeyword, Regex/Text, Newline, Indent, FieldName, Text, EOF
	wantFirst := TokenIfKeyword
	if types[0] != wantFirst {
		t.Errorf("first token: got %v, want %v", types[0], wantFirst)
	}
	if tokens[1].Value != "PATTERN" {
		t.Errorf("regex value: got %q, want %q", tokens[1].Value, "PATTERN")
	}
}

func TestLexer_FieldAssignment(t *testing.T) {
	input := "  account1 expenses:food"
	l := NewLexer(input)
	tokens := collectTokens(l)
	types := tokenTypes(tokens)

	// Indented line: Indent, FieldName, Text, EOF
	if types[0] != TokenIndent {
		t.Errorf("first: got %v, want Indent", types[0])
	}
	if types[1] != TokenFieldName {
		t.Errorf("second: got %v, want FieldName", types[1])
	}
	if tokens[2].Value != "expenses:food" {
		t.Errorf("value: got %q, want expenses:food", tokens[2].Value)
	}
}

func TestLexer_MultipleLines(t *testing.T) {
	input := "# comment\nskip 1\nfields date, amount"
	l := NewLexer(input)
	tokens := collectTokens(l)
	types := tokenTypes(tokens)

	// Comment, Newline, Directive(skip), Text(1), Newline, Directive(fields), Text, EOF
	if types[0] != TokenComment {
		t.Errorf("first: got %v, want Comment", types[0])
	}
	if types[1] != TokenNewline {
		t.Errorf("second: got %v, want Newline", types[1])
	}
	if types[2] != TokenDirective {
		t.Errorf("third: got %v, want Directive", types[2])
	}
}

func TestLexer_Position(t *testing.T) {
	input := "skip 1"
	l := NewLexer(input)
	tok := l.Next()
	if tok.Pos.Line != 1 {
		t.Errorf("line: got %d, want 1", tok.Pos.Line)
	}
	if tok.Pos.Column != 1 {
		t.Errorf("column: got %d, want 1", tok.Pos.Column)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "unix newlines",
			input: "a\nb\nc",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "crlf newlines",
			input: "a\r\nb\r\nc",
			want:  []string{"a\r", "b\r", "c"},
		},
		{
			name:  "crlf trailing newline",
			input: "a\r\nb\r\n",
			want:  []string{"a\r", "b\r", ""},
		},
		{
			name:  "single line with crlf",
			input: "skip 1\r\n",
			want:  []string{"skip 1\r", ""},
		},
		{
			name:  "mixed lf and crlf",
			input: "a\r\nb\nc",
			want:  []string{"a\r", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitLines(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitLines(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLexer_CRLF_Offsets(t *testing.T) {
	// "skip 1\r\nfields date" — 7 bytes for "skip 1\r", then \n, then "fields date"
	// fields token must start at byte offset 8, not 7
	input := "skip 1\r\nfields date"
	l := NewLexer(input)
	tokens := collectTokens(l)

	var fieldsToken Token
	for _, tok := range tokens {
		if tok.Value == "fields" {
			fieldsToken = tok
			break
		}
	}
	if fieldsToken.Value != "fields" {
		t.Fatal("fields token not found")
	}
	wantOffset := 8 // len("skip 1\r") + len("\n") = 7 + 1
	if fieldsToken.Pos.Offset != wantOffset {
		t.Errorf("fields token offset = %d, want %d", fieldsToken.Pos.Offset, wantOffset)
	}

	// Also verify newline token byte offsets
	var newlineTok Token
	for _, tok := range tokens {
		if tok.Type == TokenNewline {
			newlineTok = tok
			break
		}
	}
	if newlineTok.Type != TokenNewline {
		t.Fatal("newline token not found")
	}
	// \n is at byte 7 (after "skip 1\r"), ends at byte 8
	if newlineTok.Pos.Offset != 7 {
		t.Errorf("newline Pos.Offset = %d, want 7", newlineTok.Pos.Offset)
	}
	if newlineTok.End.Offset != 8 {
		t.Errorf("newline End.Offset = %d, want 8", newlineTok.End.Offset)
	}
}

func TestLexer_CRLF_TokenValues(t *testing.T) {
	input := "skip 1\r\nfields date, amount"
	l := NewLexer(input)
	tokens := collectTokens(l)
	types := tokenTypes(tokens)

	wantTypes := []TokenType{TokenDirective, TokenText, TokenNewline, TokenDirective, TokenText, TokenEOF}
	if len(types) != len(wantTypes) {
		t.Fatalf("got types %v, want %v", types, wantTypes)
	}
	if tokens[0].Value != "skip" {
		t.Errorf("token[0].Value = %q, want %q", tokens[0].Value, "skip")
	}
	if tokens[1].Value != "1" {
		t.Errorf("token[1].Value = %q, want %q", tokens[1].Value, "1")
	}
	if tokens[3].Value != "fields" {
		t.Errorf("token[3].Value = %q, want %q", tokens[3].Value, "fields")
	}
}

func TestIsBuiltinField(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// exact matches
		{"date", true},
		{"date2", true},
		{"amount", true},
		{"amount-in", true},
		{"amount-out", true},
		{"balance", true},
		{"currency", true},
		{"comment", true},
		// numbered variants
		{"account1", true},
		{"account12", true},
		{"amount1", true},
		{"currency2", true},
		{"balance3", true},
		// digit+suffix variants
		{"amount1-in", true},
		{"amount2-out", true},
		{"account12-in", true},
		// invalid
		{"foo", false},
		{"account-in", false},
		{"amount-xyz", false},
		{"amount1-inx", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinField(tt.name)
			if got != tt.want {
				t.Errorf("isBuiltinField(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
