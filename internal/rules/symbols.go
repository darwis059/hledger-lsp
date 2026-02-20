package rules

import "github.com/juev/hledger-lsp/internal/ast"

// Symbol represents a document symbol in a rules file.
type Symbol struct {
	Name  string
	Kind  SymbolKind
	Range ast.Range
}

// SymbolKind mirrors the LSP symbol kinds.
type SymbolKind int

const (
	SymbolKindModule   SymbolKind = 2
	SymbolKindClass    SymbolKind = 5
	SymbolKindField    SymbolKind = 8
	SymbolKindVariable SymbolKind = 13
)

// Symbols returns document symbols for a parsed rules file.
func Symbols(rf *RulesFile) []Symbol {
	if rf == nil {
		return []Symbol{}
	}
	syms := []Symbol{}

	for _, fd := range rf.FieldsDefs {
		syms = append(syms, Symbol{
			Name:  "fields",
			Kind:  SymbolKindField,
			Range: fd.Range,
		})
	}

	for _, inc := range rf.Includes {
		syms = append(syms, Symbol{
			Name:  "include " + inc.Path,
			Kind:  SymbolKindModule,
			Range: inc.Range,
		})
	}

	for _, src := range rf.Sources {
		syms = append(syms, Symbol{
			Name:  "source " + src.Path,
			Kind:  SymbolKindModule,
			Range: src.Range,
		})
	}

	for _, block := range rf.IfBlocks {
		name := "if"
		if len(block.Patterns) > 0 {
			name = "if " + block.Patterns[0]
		}
		syms = append(syms, Symbol{
			Name:  name,
			Kind:  SymbolKindClass,
			Range: block.Range,
		})
	}

	for _, d := range rf.Directives {
		name := d.Name
		if d.Value != "" {
			name = d.Name + " " + d.Value
		}
		syms = append(syms, Symbol{
			Name:  name,
			Kind:  SymbolKindVariable,
			Range: d.Range,
		})
	}

	for _, a := range rf.Assignments {
		syms = append(syms, Symbol{
			Name:  a.Field + " " + a.Value,
			Kind:  SymbolKindField,
			Range: a.Range,
		})
	}

	return syms
}
