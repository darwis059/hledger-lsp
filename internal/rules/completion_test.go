package rules

import (
	"testing"
)

func TestComplete_LineStart(t *testing.T) {
	// At start of empty line -> should offer all known directives
	items := Complete("", 0, nil)
	labels := itemLabels(items)
	if !contains(labels, "skip") {
		t.Error("expected 'skip' directive completion")
	}
	if !contains(labels, "fields") {
		t.Error("expected 'fields' directive completion")
	}
	if !contains(labels, "if") {
		t.Error("expected 'if' keyword completion")
	}
}

func TestComplete_FieldsValue(t *testing.T) {
	// After "fields " -> built-in field names
	items := Complete("fields ", 7, nil)
	labels := itemLabels(items)
	if !contains(labels, "date") {
		t.Error("expected 'date' field completion")
	}
	if !contains(labels, "amount") {
		t.Error("expected 'amount' field completion")
	}
}

func TestComplete_SeparatorValue(t *testing.T) {
	items := Complete("separator ", 10, nil)
	labels := itemLabels(items)
	if !contains(labels, ",") {
		t.Error("expected ',' separator completion")
	}
	if !contains(labels, "TAB") {
		t.Error("expected 'TAB' separator completion")
	}
}

func TestComplete_DateFormatValue(t *testing.T) {
	items := Complete("date-format ", 12, nil)
	labels := itemLabels(items)
	if !contains(labels, "%Y-%m-%d") {
		t.Error("expected date format completion to include year-month-day format")
	}
}

func TestComplete_DecimalMarkValue(t *testing.T) {
	items := Complete("decimal-mark ", 13, nil)
	labels := itemLabels(items)
	if !contains(labels, ".") {
		t.Error("expected '.' decimal mark completion")
	}
	if !contains(labels, ",") {
		t.Error("expected ',' decimal mark completion")
	}
}

func TestComplete_BalanceTypeValue(t *testing.T) {
	items := Complete("balance-type ", 13, nil)
	labels := itemLabels(items)
	if !contains(labels, "=") {
		t.Error("expected '=' balance type completion")
	}
	if !contains(labels, "==") {
		t.Error("expected '==' balance type completion")
	}
}

func TestComplete_AccountValue(t *testing.T) {
	// In indented line after account1 -> workspace accounts
	accounts := []string{"expenses:food", "assets:bank"}
	items := Complete("  account1 ", 11, accounts)
	labels := itemLabels(items)
	if !contains(labels, "expenses:food") {
		t.Error("expected 'expenses:food' account completion")
	}
}

func TestComplete_NoCompletionInComment(t *testing.T) {
	// In comment line -> no completions
	items := Complete("# this is a comment", 5, nil)
	if len(items) != 0 {
		t.Errorf("expected no completions in comment, got: %d", len(items))
	}
}

func TestComplete_ColBeyondLineLength(t *testing.T) {
	// col > len(line) falls back to prefix="" -> directive completions
	items := Complete("fields", 100, nil)
	labels := itemLabels(items)
	if !contains(labels, "skip") {
		t.Error("expected directive completions when col > len(line), missing 'skip'")
	}
	if !contains(labels, "if") {
		t.Error("expected directive completions when col > len(line), missing 'if'")
	}
}

func itemLabels(items []CompletionItem) []string {
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label
	}
	return labels
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
