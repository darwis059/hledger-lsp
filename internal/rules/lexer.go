package rules

import "strings"

// DirectiveInfo holds the name and human-readable description of a top-level directive.
type DirectiveInfo struct {
	Name   string
	Detail string
}

// KnownDirectives is the single authoritative list of recognised top-level directives.
// Both the lexer and completion provider derive their data from this slice.
var KnownDirectives = []DirectiveInfo{
	{"skip", "skip N lines"},
	{"fields", "define CSV field names"},
	{"separator", "field separator character"},
	{"source", "source CSV file"},
	{"date-format", "date parsing format"},
	{"decimal-mark", "decimal mark character"},
	{"timezone", "timezone for dates"},
	{"encoding", "file encoding"},
	{"balance-type", "balance assertion type"},
	{"include", "include another rules file"},
	{"newest-first", "process rows newest-first"},
	{"intra-day-reversed", "intra-day rows are reversed"},
	{"archive", "archive processed files"},
}

// knownDirectives is a fast-lookup set derived from KnownDirectives.
var knownDirectives = func() map[string]bool {
	m := make(map[string]bool, len(KnownDirectives))
	for _, d := range KnownDirectives {
		m[d.Name] = true
	}
	return m
}()

// BuiltinFieldNames is the authoritative list of fixed field names offered in completions.
// Dynamic patterns (accountN, amountN-in, etc.) are validated by isBuiltinField but not listed here.
var BuiltinFieldNames = []string{
	"date", "date2", "status", "code", "description", "payee", "note",
	"comment",
	"account1", "account2", "account3",
	"amount", "amount1", "amount2", "amount-in", "amount-out",
	"currency", "currency1", "currency2",
	"balance", "balance1", "balance2",
}

var builtinFieldPrefixes = []string{"account", "amount", "currency", "balance", "comment"}

func isBuiltinField(name string) bool {
	switch name {
	case "date", "date2", "status", "code", "description", "payee", "note",
		"comment", "amount", "amount-in", "amount-out", "currency", "balance":
		return true
	}
	for _, p := range builtinFieldPrefixes {
		if strings.HasPrefix(name, p) {
			rest := name[len(p):]
			if rest == "" {
				return true
			}
			digits := 0
			for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
				digits++
			}
			suffix := rest[digits:]
			if suffix == "" || (digits > 0 && (suffix == "-in" || suffix == "-out")) {
				return true
			}
		}
	}
	return false
}

// Lexer tokenises a rules file line-by-line.
type Lexer struct {
	lines   []string
	lineIdx int // 0-based current line index
	pending []Token
	offset  int // byte offset of start of current line
}

func NewLexer(input string) *Lexer {
	return &Lexer{lines: splitLines(input)}
}

func splitLines(input string) []string {
	return strings.Split(input, "\n")
}

func (l *Lexer) Next() Token {
	for len(l.pending) == 0 {
		if l.lineIdx >= len(l.lines) {
			return Token{Type: TokenEOF, Pos: Position{Line: l.lineIdx + 1, Column: 1, Offset: l.offset}}
		}

		rawLine := l.lines[l.lineIdx]
		line := strings.TrimRight(rawLine, "\r")
		lineNum := l.lineIdx + 1
		lineOffset := l.offset

		lineTokens := scanLine(line, lineNum, lineOffset)
		l.lineIdx++
		l.offset += len(rawLine) + 1 // +1 for '\n'; rawLine includes \r for CRLF files

		// Append newline separator between lines (not after last line)
		if l.lineIdx < len(l.lines) {
			lineTokens = append(lineTokens, Token{
				Type:  TokenNewline,
				Value: "\n",
				Pos:   Position{Line: lineNum, Column: len(line) + 1, Offset: lineOffset + len(rawLine)},
				End:   Position{Line: lineNum, Column: len(line) + 2, Offset: lineOffset + len(rawLine) + 1},
			})
		}

		l.pending = lineTokens
	}

	tok := l.pending[0]
	l.pending = l.pending[1:]
	return tok
}

// scanLine returns the tokens for a single line (without EOF).
func scanLine(line string, lineNum int, baseOffset int) []Token {
	if line == "" {
		return nil
	}

	isIndented := line[0] == ' ' || line[0] == '\t'
	if isIndented {
		return scanIndentedLine(line, lineNum, baseOffset)
	}
	return scanTopLevelLine(line, lineNum, baseOffset)
}

func scanTopLevelLine(line string, lineNum int, baseOffset int) []Token {
	// Comments
	if line[0] == '#' || line[0] == ';' || line[0] == '*' {
		return []Token{{
			Type:  TokenComment,
			Value: line,
			Pos:   Position{Line: lineNum, Column: 1, Offset: baseOffset},
			End:   Position{Line: lineNum, Column: 1 + len(line), Offset: baseOffset + len(line)},
		}}
	}

	// "if" keyword
	if line == "if" || strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "if\t") {
		return scanIfLine(line, lineNum, baseOffset)
	}

	// "end" keyword (standalone)
	if line == "end" {
		return []Token{{
			Type:  TokenEndKeyword,
			Value: "end",
			Pos:   Position{Line: lineNum, Column: 1, Offset: baseOffset},
			End:   Position{Line: lineNum, Column: 4, Offset: baseOffset + 3},
		}}
	}

	word, afterWord := splitWord(line)

	// Known directive
	if knownDirectives[word] {
		tokens := []Token{{
			Type:  TokenDirective,
			Value: word,
			Pos:   Position{Line: lineNum, Column: 1, Offset: baseOffset},
			End:   Position{Line: lineNum, Column: 1 + len(word), Offset: baseOffset + len(word)},
		}}
		if afterWord != "" {
			val := strings.TrimLeft(afterWord, " \t")
			if val != "" {
				valStart := len(line) - len(val)
				tokens = append(tokens, Token{
					Type:  TokenText,
					Value: val,
					Pos:   Position{Line: lineNum, Column: 1 + valStart, Offset: baseOffset + valStart},
					End:   Position{Line: lineNum, Column: 1 + len(line), Offset: baseOffset + len(line)},
				})
			}
		}
		return tokens
	}

	// Unindented field assignment
	if isBuiltinField(word) && afterWord != "" {
		return scanFieldAssignment(line, lineNum, baseOffset, 0)
	}

	// Fallback: plain text
	return []Token{{
		Type:  TokenText,
		Value: line,
		Pos:   Position{Line: lineNum, Column: 1, Offset: baseOffset},
		End:   Position{Line: lineNum, Column: 1 + len(line), Offset: baseOffset + len(line)},
	}}
}

func scanIndentedLine(line string, lineNum int, baseOffset int) []Token {
	indent := 0
	for indent < len(line) && (line[indent] == ' ' || line[indent] == '\t') {
		indent++
	}
	rest := line[indent:]
	if rest == "" {
		return nil
	}

	tokens := []Token{{
		Type:  TokenIndent,
		Value: line[:indent],
		Pos:   Position{Line: lineNum, Column: 1, Offset: baseOffset},
		End:   Position{Line: lineNum, Column: 1 + indent, Offset: baseOffset + indent},
	}}

	fieldTokens := scanFieldAssignment(rest, lineNum, baseOffset+indent, indent)
	return append(tokens, fieldTokens...)
}

func scanFieldAssignment(text string, lineNum int, baseOffset int, indentLen int) []Token {
	word, afterWord := splitWord(text)
	col := indentLen + 1

	if isBuiltinField(word) {
		tokens := []Token{{
			Type:  TokenFieldName,
			Value: word,
			Pos:   Position{Line: lineNum, Column: col, Offset: baseOffset},
			End:   Position{Line: lineNum, Column: col + len(word), Offset: baseOffset + len(word)},
		}}
		if afterWord != "" {
			val := strings.TrimLeft(afterWord, " \t")
			if val != "" {
				valStart := len(text) - len(val)
				tokens = append(tokens, Token{
					Type:  TokenText,
					Value: val,
					Pos:   Position{Line: lineNum, Column: col + valStart, Offset: baseOffset + valStart},
					End:   Position{Line: lineNum, Column: col + len(text), Offset: baseOffset + len(text)},
				})
			}
		}
		return tokens
	}

	return []Token{{
		Type:  TokenText,
		Value: text,
		Pos:   Position{Line: lineNum, Column: col, Offset: baseOffset},
		End:   Position{Line: lineNum, Column: col + len(text), Offset: baseOffset + len(text)},
	}}
}

func scanIfLine(line string, lineNum int, baseOffset int) []Token {
	tokens := []Token{{
		Type:  TokenIfKeyword,
		Value: "if",
		Pos:   Position{Line: lineNum, Column: 1, Offset: baseOffset},
		End:   Position{Line: lineNum, Column: 3, Offset: baseOffset + 2},
	}}

	if len(line) > 2 {
		val := strings.TrimLeft(line[2:], " \t")
		if val != "" {
			valStart := len(line) - len(val)
			tokens = append(tokens, Token{
				Type:  TokenRegex,
				Value: val,
				Pos:   Position{Line: lineNum, Column: 1 + valStart, Offset: baseOffset + valStart},
				End:   Position{Line: lineNum, Column: 1 + len(line), Offset: baseOffset + len(line)},
			})
		}
	}
	return tokens
}

func splitWord(s string) (word, rest string) {
	idx := strings.IndexAny(s, " \t")
	if idx == -1 {
		return s, ""
	}
	return s[:idx], s[idx:]
}

func firstWord(s string) string {
	if idx := strings.IndexAny(s, " \t"); idx != -1 {
		return s[:idx]
	}
	return s
}
