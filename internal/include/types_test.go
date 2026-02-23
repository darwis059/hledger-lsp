package include

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/juev/hledger-lsp/internal/ast"
)

func TestFormatDirectives_FiltersDecimalMarkFromIncludes(t *testing.T) {
	primary := &ast.Journal{
		Directives: []ast.Directive{
			ast.CommodityDirective{Commodity: ast.Commodity{Symbol: "RUB"}, Format: "1.000,00 RUB"},
			ast.DecimalMarkDirective{Mark: ","},
		},
	}

	included := &ast.Journal{
		Directives: []ast.Directive{
			ast.DecimalMarkDirective{Mark: "."},
			ast.DefaultCommodityDirective{Symbol: "$", Format: "$1,000.00"},
		},
	}

	resolved := &ResolvedJournal{
		Primary:   primary,
		Files:     map[string]*ast.Journal{"included.journal": included},
		FileOrder: []string{"included.journal"},
	}

	result := resolved.FormatDirectives()

	require.Len(t, result, 3, "primary commodity + primary decimal-mark + included D directive")

	var hasDecimalMarkComma, hasDecimalMarkDot bool
	var hasCommodityRUB, hasDefaultD bool
	for _, d := range result {
		switch dd := d.(type) {
		case ast.DecimalMarkDirective:
			if dd.Mark == "," {
				hasDecimalMarkComma = true
			}
			if dd.Mark == "." {
				hasDecimalMarkDot = true
			}
		case ast.CommodityDirective:
			hasCommodityRUB = true
		case ast.DefaultCommodityDirective:
			hasDefaultD = true
		}
	}

	assert.True(t, hasDecimalMarkComma, "decimal-mark from primary should be preserved")
	assert.False(t, hasDecimalMarkDot, "decimal-mark from included file should be filtered")
	assert.True(t, hasCommodityRUB, "commodity directive from primary should be preserved")
	assert.True(t, hasDefaultD, "D directive from included file should be preserved")
}

func TestFormatDirectives_PreservesAllDirectivesFromIncludes(t *testing.T) {
	primary := &ast.Journal{}

	included := &ast.Journal{
		Directives: []ast.Directive{
			ast.CommodityDirective{Commodity: ast.Commodity{Symbol: "EUR"}, Format: "1 000,00 EUR"},
			ast.DefaultCommodityDirective{Symbol: "$", Format: "$1,000.00"},
			ast.AccountDirective{Account: ast.Account{Name: "expenses:food"}},
		},
	}

	resolved := &ResolvedJournal{
		Primary:   primary,
		Files:     map[string]*ast.Journal{"included.journal": included},
		FileOrder: []string{"included.journal"},
	}

	result := resolved.FormatDirectives()
	require.Len(t, result, 3, "all non-decimal-mark directives from includes should be preserved")
}

func TestFormatDirectives_NilPrimary(t *testing.T) {
	resolved := &ResolvedJournal{
		Primary:   nil,
		Files:     map[string]*ast.Journal{},
		FileOrder: nil,
	}

	result := resolved.FormatDirectives()
	assert.Empty(t, result)
}
