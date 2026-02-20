package rules

import (
	"github.com/juev/hledger-lsp/internal/ast"
)

// DiagnosticSeverity mirrors LSP diagnostic severity levels.
type DiagnosticSeverity int

const (
	SeverityError   DiagnosticSeverity = 1
	SeverityWarning DiagnosticSeverity = 2
)

// Diagnostic represents a single rules-file validation issue.
type Diagnostic struct {
	Range    ast.Range
	Message  string
	Code     string
	Severity DiagnosticSeverity
}

var validBalanceTypes = map[string]bool{
	"=": true, "==": true, "=*": true, "==*": true,
}

// Diagnostics returns validation diagnostics for a parsed rules file.
// parseDiags must be the diagnostics returned by Parse.
func Diagnostics(rf *RulesFile, parseDiags []Diagnostic) []Diagnostic {
	if rf == nil {
		return []Diagnostic{}
	}
	var diags []Diagnostic

	if len(rf.FieldsDefs) > 1 {
		for _, fd := range rf.FieldsDefs[1:] {
			diags = append(diags, Diagnostic{
				Range:    fd.Range,
				Message:  "duplicate 'fields' directive",
				Code:     "DUPLICATE_FIELDS",
				Severity: SeverityWarning,
			})
		}
	}

	for _, d := range rf.Directives {
		switch d.Name {
		case "decimal-mark":
			if d.Value != "." && d.Value != "," {
				diags = append(diags, Diagnostic{
					Range:    d.Range,
					Message:  "invalid decimal-mark: must be '.' or ','",
					Code:     "INVALID_DECIMAL_MARK",
					Severity: SeverityError,
				})
			}
		case "balance-type":
			if d.Value != "" && !validBalanceTypes[d.Value] {
				diags = append(diags, Diagnostic{
					Range:    d.Range,
					Message:  "invalid balance-type: must be one of =, ==, =*, ==*",
					Code:     "INVALID_BALANCE_TYPE",
					Severity: SeverityError,
				})
			}
		}
	}

	diags = append(diags, parseDiags...)
	return diags
}
