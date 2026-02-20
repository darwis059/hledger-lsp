package rules

import (
	"testing"
)

func TestDiagnostics_BareAmpersand(t *testing.T) {
	// A bare "&" continuation line with no pattern should produce an EMPTY_PATTERN warning.
	input := "if PAT\n&\n  account1 expenses:food"
	diags := diagsFor(input)
	found := false
	for _, d := range diags {
		if d.Code == "EMPTY_PATTERN" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected EMPTY_PATTERN diagnostic for bare '&' continuation line")
	}
}

func TestDiagnostics_AmpersandWithPattern_NoDiagnostic(t *testing.T) {
	// "& PATTERN" should NOT produce an EMPTY_PATTERN diagnostic.
	input := "if PAT\n& OTHERPATTERN\n  account1 expenses:food"
	diags := diagsFor(input)
	for _, d := range diags {
		if d.Code == "EMPTY_PATTERN" {
			t.Errorf("unexpected EMPTY_PATTERN for '& PATTERN': %v", d)
		}
	}
}
