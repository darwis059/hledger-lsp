package rules

import (
	"strings"

	"github.com/juev/hledger-lsp/internal/lsputil"
)

// CompletionItem is a simple completion suggestion for rules files.
type CompletionItem struct {
	Label  string
	Detail string
	Kind   CompletionItemKind
}

// CompletionItemKind mirrors LSP completion item kinds relevant to rules.
type CompletionItemKind int

const (
	KindKeyword  CompletionItemKind = 14
	KindField    CompletionItemKind = 5
	KindValue    CompletionItemKind = 12
	KindVariable CompletionItemKind = 6
)

// Complete returns completion items for the given cursor position in a rules file line.
//
// Parameters:
//   - line: the current line text
//   - col: 0-based cursor column in UTF-16 code units (as used by LSP)
//   - workspaceAccounts: account names from the journal workspace (may be nil)
func Complete(line string, col int, workspaceAccounts []string) []CompletionItem {
	// No completions in comments
	if len(line) > 0 && (line[0] == '#' || line[0] == ';' || line[0] == '*') {
		return nil
	}

	byteCol := lsputil.UTF16OffsetToByteOffset(line, col)
	prefix := line[:byteCol]

	// Start of line (empty or no significant prefix) -> directives
	if strings.TrimSpace(prefix) == "" {
		return directiveCompletions()
	}

	// Indented line: field name value
	if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
		return completeIndentedLine(prefix, workspaceAccounts)
	}

	// Top-level directive value completions
	return completeDirectiveValue(prefix)
}

func completeDirectiveValue(prefix string) []CompletionItem {
	switch {
	case strings.HasPrefix(prefix, "fields "):
		return fieldNameCompletions()

	case strings.HasPrefix(prefix, "separator "):
		return separatorCompletions()

	case strings.HasPrefix(prefix, "date-format "):
		return dateFormatCompletions()

	case strings.HasPrefix(prefix, "decimal-mark "):
		return decimalMarkCompletions()

	case strings.HasPrefix(prefix, "balance-type "):
		return balanceTypeCompletions()

	case strings.HasPrefix(prefix, "include "):
		return nil

	case strings.HasPrefix(prefix, "source "):
		return nil
	}

	return directiveCompletions()
}

func completeIndentedLine(prefix string, workspaceAccounts []string) []CompletionItem {
	trimmed := strings.TrimLeft(prefix, " \t")
	if trimmed == "" {
		return fieldNameCompletions()
	}

	word, rest := splitWord(trimmed)
	if isBuiltinField(word) && rest != "" {
		// After a field name -> value completions based on field type
		if strings.HasPrefix(word, "account") {
			return accountCompletions(workspaceAccounts)
		}
		return nil
	}

	// Partial field name
	return fieldNameCompletions()
}

func directiveCompletions() []CompletionItem {
	// KnownDirectives covers all top-level directives; "if" and "end" are added as keywords.
	items := make([]CompletionItem, 0, len(KnownDirectives)+2)
	for _, d := range KnownDirectives {
		items = append(items, CompletionItem{Label: d.Name, Detail: d.Detail, Kind: KindKeyword})
	}
	items = append(items, CompletionItem{Label: "if", Detail: "conditional block", Kind: KindKeyword})
	items = append(items, CompletionItem{Label: "end", Detail: "stop processing", Kind: KindKeyword})
	return items
}

func fieldNameCompletions() []CompletionItem {
	items := make([]CompletionItem, len(BuiltinFieldNames))
	for i, f := range BuiltinFieldNames {
		items[i] = CompletionItem{Label: f, Detail: "field name", Kind: KindField}
	}
	return items
}

func separatorCompletions() []CompletionItem {
	seps := []struct{ label, detail string }{
		{",", "comma"},
		{";", "semicolon"},
		{"|", "pipe"},
		{"TAB", "tab character"},
		{"SPACE", "space character"},
	}
	items := make([]CompletionItem, len(seps))
	for i, s := range seps {
		items[i] = CompletionItem{Label: s.label, Detail: s.detail, Kind: KindValue}
	}
	return items
}

func dateFormatCompletions() []CompletionItem {
	formats := []string{
		"%Y-%m-%d",
		"%m/%d/%Y",
		"%d/%m/%Y",
		"%Y/%m/%d",
		"%d-%m-%Y",
		"%m-%d-%Y",
		"%d.%m.%Y",
		"%Y.%m.%d",
	}
	items := make([]CompletionItem, len(formats))
	for i, f := range formats {
		items[i] = CompletionItem{Label: f, Detail: "date format", Kind: KindValue}
	}
	return items
}

func decimalMarkCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: ".", Detail: "period (US/UK)", Kind: KindValue},
		{Label: ",", Detail: "comma (EU)", Kind: KindValue},
	}
}

func balanceTypeCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "=", Detail: "partial balance assertion", Kind: KindValue},
		{Label: "==", Detail: "total balance assertion", Kind: KindValue},
		{Label: "=*", Detail: "partial balance assignment", Kind: KindValue},
		{Label: "==*", Detail: "total balance assignment", Kind: KindValue},
	}
}

func accountCompletions(workspaceAccounts []string) []CompletionItem {
	if len(workspaceAccounts) == 0 {
		return nil
	}
	items := make([]CompletionItem, len(workspaceAccounts))
	for i, acc := range workspaceAccounts {
		items[i] = CompletionItem{Label: acc, Detail: "account", Kind: KindVariable}
	}
	return items
}
