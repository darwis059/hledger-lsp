package rules

import (
	"testing"
)

func TestDiagnostics_NilRulesFile(t *testing.T) {
	diags := Diagnostics(nil, nil)
	if diags == nil {
		t.Error("expected non-nil slice from Diagnostics(nil)")
	}
	if len(diags) != 0 {
		t.Errorf("expected empty diagnostics for nil rules file, got %d", len(diags))
	}
}

func TestSymbols_NilRulesFile(t *testing.T) {
	syms := Symbols(nil)
	if syms == nil {
		t.Error("expected non-nil slice from Symbols(nil)")
	}
	if len(syms) != 0 {
		t.Errorf("expected empty symbols for nil rules file, got %d", len(syms))
	}
}

func TestLinks_NilRulesFile(t *testing.T) {
	links := Links(nil, "/dir/")
	if links == nil {
		t.Error("expected non-nil slice from Links(nil, ...)")
	}
	if len(links) != 0 {
		t.Errorf("expected empty links for nil rules file, got %d", len(links))
	}
}
