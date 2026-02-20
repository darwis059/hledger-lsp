package rules

import (
	"testing"
)

func diagsFor(input string) []Diagnostic {
	rf, pd := Parse(input)
	return Diagnostics(rf, pd)
}

func TestDiagnostics_ValidFile(t *testing.T) {
	input := "skip 1\nfields date, description, amount\nseparator ,"
	diags := diagsFor(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for valid file, got: %v", diags)
	}
}

func TestDiagnostics_UnknownDirective(t *testing.T) {
	input := "bogus-directive value"
	diags := diagsFor(input)
	// Unknown directives at top level produce a diagnostic
	if len(diags) == 0 {
		t.Error("expected diagnostic for unknown directive")
	}
}

func TestDiagnostics_DuplicateFields(t *testing.T) {
	input := "fields date, amount\nfields date, amount"
	diags := diagsFor(input)
	found := false
	for _, d := range diags {
		if d.Code == "DUPLICATE_FIELDS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected DUPLICATE_FIELDS diagnostic")
	}
}

func TestDiagnostics_InvalidDecimalMark(t *testing.T) {
	input := "decimal-mark x"
	diags := diagsFor(input)
	found := false
	for _, d := range diags {
		if d.Code == "INVALID_DECIMAL_MARK" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected INVALID_DECIMAL_MARK diagnostic")
	}
}

func TestDiagnostics_InvalidBalanceType(t *testing.T) {
	input := "balance-type bad"
	diags := diagsFor(input)
	found := false
	for _, d := range diags {
		if d.Code == "INVALID_BALANCE_TYPE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected INVALID_BALANCE_TYPE diagnostic")
	}
}

func TestDiagnostics_ValidDecimalMark(t *testing.T) {
	for _, mark := range []string{".", ","} {
		input := "decimal-mark " + mark
		diags := diagsFor(input)
		if len(diags) != 0 {
			t.Errorf("expected no diagnostics for decimal-mark %q, got: %v", mark, diags)
		}
	}
}

func TestDiagnostics_ValidBalanceType(t *testing.T) {
	for _, bt := range []string{"=", "==", "=*", "==*"} {
		input := "balance-type " + bt
		diags := diagsFor(input)
		if len(diags) != 0 {
			t.Errorf("expected no diagnostics for balance-type %q, got: %v", bt, diags)
		}
	}
}

func TestDiagnostics_UnknownIndentedField(t *testing.T) {
	input := "if\n  acount2 expenses:food"
	diags := diagsFor(input)
	found := false
	for _, d := range diags {
		if d.Code == "UNKNOWN_FIELD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected UNKNOWN_FIELD diagnostic for typo in indented field name")
	}
}

func TestDiagnostics_UnknownIndentedField_ValidFieldNotFlagged(t *testing.T) {
	input := "if\n  account1 expenses:food"
	diags := diagsFor(input)
	for _, d := range diags {
		if d.Code == "UNKNOWN_FIELD" {
			t.Errorf("unexpected UNKNOWN_FIELD for valid field name: %v", d)
		}
	}
}

func TestDiagnostics_UnknownIndentedField_BareWordNotFlagged(t *testing.T) {
	input := "if\n  grocery"
	diags := diagsFor(input)
	for _, d := range diags {
		if d.Code == "UNKNOWN_FIELD" {
			t.Errorf("unexpected UNKNOWN_FIELD for bare word (regex pattern): %v", d)
		}
	}
}

func TestDiagnostics_UnknownIndentedField_TopLevel(t *testing.T) {
	input := "  bogusfield somevalue"
	diags := diagsFor(input)
	found := false
	for _, d := range diags {
		if d.Code == "UNKNOWN_FIELD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected UNKNOWN_FIELD diagnostic for indented unknown field outside if-block")
	}
}

func TestDiagnostics_DiagnosticHasRange(t *testing.T) {
	// "bogus-directive" is 15 chars on line 1 starting at column 1
	input := "bogus-directive"
	diags := diagsFor(input)
	if len(diags) == 0 {
		t.Fatal("expected at least one diagnostic")
	}
	d := diags[0]
	if d.Range.Start.Line != 1 {
		t.Errorf("Start.Line: got %d, want 1", d.Range.Start.Line)
	}
	if d.Range.Start.Column != 1 {
		t.Errorf("Start.Column: got %d, want 1", d.Range.Start.Column)
	}
	if d.Range.End.Line != 1 {
		t.Errorf("End.Line: got %d, want 1", d.Range.End.Line)
	}
	if d.Range.End.Column != 16 {
		t.Errorf("End.Column: got %d, want 16", d.Range.End.Column)
	}
}
