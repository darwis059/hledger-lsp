package rules

import (
	"testing"
)

func parseRF(input string) *RulesFile {
	rf, _ := Parse(input)
	return rf
}

func TestSymbols_Empty(t *testing.T) {
	syms := Symbols(parseRF(""))
	if syms == nil {
		t.Fatal("expected non-nil symbols slice")
	}
}

func TestSymbols_FieldsDirective(t *testing.T) {
	input := "fields date, description, amount"
	syms := Symbols(parseRF(input))
	found := false
	for _, s := range syms {
		if s.Name == "fields" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected symbol for 'fields' directive")
	}
}

func TestSymbols_IncludeDirective(t *testing.T) {
	input := "include common.rules"
	syms := Symbols(parseRF(input))
	found := false
	for _, s := range syms {
		if s.Name == "include common.rules" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected symbol for 'include common.rules'")
	}
}

func TestSymbols_IfBlock(t *testing.T) {
	input := "if PATTERN\n  account1 expenses:food"
	syms := Symbols(parseRF(input))
	found := false
	for _, s := range syms {
		if s.Name == "if PATTERN" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected symbol for 'if PATTERN'")
	}
}

func TestSymbols_SourceDirective(t *testing.T) {
	input := "source data.csv"
	syms := Symbols(parseRF(input))
	found := false
	for _, s := range syms {
		if s.Name == "source data.csv" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected symbol for 'source data.csv'")
	}
}

func TestSymbols_SimpleDirective(t *testing.T) {
	input := "skip 1"
	syms := Symbols(parseRF(input))
	found := false
	for _, s := range syms {
		if s.Name == "skip 1" && s.Kind == SymbolKindVariable {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected symbol for 'skip 1' directive with SymbolKindVariable")
	}
}

func TestSymbols_Assignment(t *testing.T) {
	input := "  account1 expenses:food"
	syms := Symbols(parseRF(input))
	found := false
	for _, s := range syms {
		if s.Name == "account1 expenses:food" && s.Kind == SymbolKindField {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected symbol for 'account1 expenses:food' assignment with SymbolKindField")
	}
}

func TestSymbols_Range(t *testing.T) {
	input := "fields date, amount"
	syms := Symbols(parseRF(input))
	if len(syms) == 0 {
		t.Fatal("expected symbols")
	}
	sym := syms[0]
	if sym.Range.Start.Line != 1 {
		t.Errorf("start line: got %d, want 1", sym.Range.Start.Line)
	}
}
