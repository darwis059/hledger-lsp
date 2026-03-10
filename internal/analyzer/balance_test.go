package analyzer

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/juev/hledger-lsp/internal/parser"
)

func TestCheckBalance_SimpleBalanced(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $50
    assets:cash  $-50`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
	assert.Empty(t, result.Differences)
}

func TestCheckBalance_InferredAmount(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $50
    assets:cash`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
	assert.Equal(t, 1, result.InferredIdx)
}

func TestCheckBalance_Unbalanced(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $50
    assets:cash  $-40`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced)
	assert.Equal(t, decimal.NewFromInt(10), result.Differences["$"])
}

func TestCheckBalance_MultiCommodity(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $50
    expenses:rent  EUR 100
    assets:cash  $-50
    assets:bank  EUR -100`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
}

func TestCheckBalance_MultiCommodity_Unbalanced(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $50
    expenses:rent  EUR 100
    assets:cash  $-50
    assets:bank  EUR -90`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced)
	assert.Equal(t, decimal.NewFromInt(10), result.Differences["EUR"])
}

func TestCheckBalance_MultipleInferred_Error(t *testing.T) {
	input := `2024-01-15 test
    expenses:food
    assets:cash`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced)
}

func TestCheckBalance_WithCost_UnitPrice(t *testing.T) {
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL @ $150
    assets:cash  $-1500`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
}

func TestCheckBalance_WithCost_TotalPrice(t *testing.T) {
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL @@ $1500
    assets:cash  $-1500`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
}

func TestCheckBalance_VirtualUnbalanced_Exempt(t *testing.T) {
	t.Skip("Parser does not yet support virtual postings (task 3.4)")
}

func TestCheckBalance_ZeroAmount(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $0
    assets:cash  $0`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
}

func TestCheckBalance_NegativeAmounts(t *testing.T) {
	input := `2024-01-15 refund
    assets:cash  $100
    expenses:food  $-100`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced)
}

func TestCheckBalance_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		balanced bool
	}{
		{
			name: "simple balanced",
			input: `2024-01-15 test
    expenses:food  $50
    assets:cash  $-50`,
			balanced: true,
		},
		{
			name: "inferred single posting",
			input: `2024-01-15 test
    expenses:food  $50
    assets:cash`,
			balanced: true,
		},
		{
			name: "unbalanced by $10",
			input: `2024-01-15 test
    expenses:food  $50
    assets:cash  $-40`,
			balanced: false,
		},
		{
			name: "three postings balanced",
			input: `2024-01-15 test
    expenses:food  $30
    expenses:drinks  $20
    assets:cash  $-50`,
			balanced: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := parser.Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)

			result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

			assert.Equal(t, tt.balanced, result.Balanced)
		})
	}
}

func TestCheckBalance_MultiCurrencyInferred(t *testing.T) {
	input := `2024-01-01 opening balances
    assets:bank  1000 RUB
    assets:cash  100 USD
    equity:opening`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "multi-currency transaction with single inferred posting should be balanced")
	assert.Equal(t, 2, result.InferredIdx)
}

func TestCheckBalance_MultiCurrencyWithBalanceAssertion(t *testing.T) {
	input := `2024-01-01 opening balances
    assets:bank  1000 RUB = 1000 RUB
    assets:cash  100 USD = 100 USD
    equity:opening`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "multi-currency with balance assertions should be balanced")
}

func TestCheckBalance_MultiCurrencyExplicitlyBalanced(t *testing.T) {
	input := `2024-01-01 test
    assets:bank  1000 RUB
    assets:cash  100 USD
    equity:rub  -1000 RUB
    equity:usd  -100 USD`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "explicitly balanced multi-currency should be balanced")
}

func TestCheckBalance_BalanceAssertionOnly_NotCountedAsInferred(t *testing.T) {
	input := `2024-01-15 opening balances
    assets:bank  1000 CNY
    assets:cash  = 500 CNY
    assets:wallet  = 200 CNY
    equity:opening`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "balance-assertion-only postings should not count as inferred")
}

func TestCheckBalance_AllBalanceAssertionOnly_Balanced(t *testing.T) {
	input := `2024-01-15 check balances
    assets:bank  1000 CNY
    assets:cash  = 500 CNY
    assets:wallet  = 200 CNY
    assets:savings  -1000 CNY`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "all balance-assertion-only postings contribute zero, explicit amounts should balance")
}

func TestCheckBalance_BalanceAssertionPlusTwoInferred_MultipleInferred(t *testing.T) {
	input := `2024-01-15 test
    assets:bank  = 500 CNY
    expenses:food
    assets:cash`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced, "two truly inferred postings should still be MULTIPLE_INFERRED even with balance assertion posting")
}

func TestCheckBalance_ExplicitAmountPlusBalanceAssertionPlusOneInferred(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  100 CNY
    assets:cash  = 500 CNY
    equity:opening`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "explicit amount + balance-assertion-only + 1 inferred should be balanced")
}

func TestCheckBalance_QuotedCommodityWithTotalCostAndBalanceAssertion(t *testing.T) {
	input := `2024-06-01 sell stock
    assets:broker  "STOCK" - 100 @@ 5000 CNY = 0 "STOCK"
    assets:cash  5000 CNY`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "stock sale with total cost and balance assertion should be balanced")
}

func TestDecimalPrecision(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		precision int32
	}{
		{"integer", "100", 0},
		{"one decimal", "1.5", 1},
		{"two decimals", "1.00", 2},
		{"three decimals", "1.234", 3},
		{"four decimals", "6.8237", 4},
		{"zero", "0", 0},
		{"negative", "-1.50", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decimal.NewFromString(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.precision, decimalPrecision(d))
		})
	}
}

func TestToleranceForPrecision(t *testing.T) {
	tests := []struct {
		name      string
		precision int32
		expected  string
	}{
		{"precision 0", 0, "0.5"},
		{"precision 1", 1, "0.05"},
		{"precision 2", 2, "0.005"},
		{"precision 3", 3, "0.0005"},
		{"precision 4", 4, "0.00005"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, err := decimal.NewFromString(tt.expected)
			require.NoError(t, err)
			assert.True(t, toleranceForPrecision(tt.precision).Equal(expected),
				"toleranceForPrecision(%d) = %s, want %s", tt.precision, toleranceForPrecision(tt.precision), expected)
		})
	}
}

func TestCheckBalance_CostRounding_WithinTolerance(t *testing.T) {
	// 3.00 * 0.333 = 0.999 EUR; balance = 0.999 - 1.00 = -0.001
	// Posting amounts: 3.00 (prec 2) mapped to EUR, 1.00 (prec 2) → max 2
	// Tolerance = 0.005; |0.001| < 0.005 → balanced
	input := `2024-01-15 exchange
    assets:foreign  3.00 USD @ 0.333 EUR
    assets:eur  -1.00 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "cost rounding 0.001 within tolerance 0.005")
}

func TestCheckBalance_CostRounding_ExceedsTolerance(t *testing.T) {
	// 3.00 * 0.337 = 1.011 EUR; balance = 1.011 - 1.00 = 0.011
	// Precision 2, tolerance 0.005; 0.011 > 0.005 → unbalanced
	input := `2024-01-15 exchange
    assets:foreign  3.00 USD @ 0.337 EUR
    assets:eur  -1.00 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced, "cost rounding 0.011 exceeds tolerance 0.005")
}

func TestCheckBalance_CostPrecisionExcluded(t *testing.T) {
	// 5 * 0.2006 = 1.003 EUR; balance = 1.003 - 1 = 0.003
	// Posting amounts: 5 (prec 0) mapped to EUR, 1 (prec 0) → max 0
	// Tolerance = 0.5; |0.003| < 0.5 → balanced
	// If cost precision (4 from 0.2006) were included: max = 4, tolerance = 0.00005
	// 0.003 > 0.00005 → would be unbalanced
	input := `2024-01-15 exchange
    assets:foreign  5 USD @ 0.2006 EUR
    assets:eur  -1 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "cost amount precision (4) must NOT tighten tolerance; "+
		"posting precision 0 → tolerance 0.5, imbalance 0.003 within tolerance")
}

func TestCheckBalance_HigherPostingPrecision_TighterTolerance(t *testing.T) {
	// 3.000 * 0.3333 = 0.9999 EUR; balance = 0.9999 - 1.000 = -0.0001
	// Posting amounts: 3.000 (prec 3) mapped to EUR, 1.000 (prec 3) → max 3
	// Tolerance = 0.0005; |0.0001| < 0.0005 → balanced
	input := `2024-01-15 exchange
    assets:foreign  3.000 USD @ 0.3333 EUR
    assets:eur  -1.000 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "precision 3 tolerance 0.0005; |0.0001| within tolerance")
}

func TestCheckBalance_HigherPostingPrecision_ExceedsTighterTolerance(t *testing.T) {
	// 3.000 * 0.333 = 0.999 EUR; balance = 0.999 - 1.000 = -0.001
	// Posting amounts: 3.000 (prec 3) mapped to EUR, 1.000 (prec 3) → max 3
	// Tolerance = 0.0005; |0.001| > 0.0005 → unbalanced
	input := `2024-01-15 exchange
    assets:foreign  3.000 USD @ 0.333 EUR
    assets:eur  -1.000 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced, "precision 3 tolerance 0.0005; |0.001| exceeds tolerance")
}

func TestCheckBalance_MultiCommodity_DifferentTolerances(t *testing.T) {
	// EUR: 3.00 * 0.333 = 0.999; balance = -0.001; prec 2, tol 0.005 → OK
	// CHF: 3.000 * 0.3333 = 0.9999; balance = -0.0001; prec 3, tol 0.0005 → OK
	input := `2024-01-15 exchange
    assets:usd1  3.00 USD @ 0.333 EUR
    assets:eur  -1.00 EUR
    assets:usd2  3.000 GBP @ 0.3333 CHF
    assets:chf  -1.000 CHF`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.True(t, result.Balanced, "each commodity uses its own precision for tolerance")
}

func TestCheckBalance_UserToleranceOverridesPrecision(t *testing.T) {
	// 3.00 * 0.335 = 1.005 EUR; balance = 1.005 - 1.00 = 0.00543 (simulated)
	// Precision 2 → precisionTolerance = 0.005; 0.00543 > 0.005 → unbalanced with default
	// userTolerance = 0.01; 0.00543 < 0.01 → balanced with user tolerance
	input := `2024-01-15 exchange
    assets:foreign  3.00 USD @ 0.33510 EUR
    assets:eur  -1.00 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	resultDefault := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.False(t, resultDefault.Balanced, "should be unbalanced with default tolerance")

	userTol, _ := decimal.NewFromString("0.01")
	resultUser := CheckBalance(&journal.Transactions[0], userTol, nil)
	assert.True(t, resultUser.Balanced, "should be balanced with user tolerance 0.01")
}

func TestCheckBalance_PrecisionToleranceWinsWhenHigher(t *testing.T) {
	// Precision 0 → precisionTolerance = 0.5
	// userTolerance = 0.001 → max(0.5, 0.001) = 0.5
	// Imbalance 0.003 < 0.5 → balanced (precision tolerance wins)
	input := `2024-01-15 exchange
    assets:foreign  5 USD @ 0.2006 EUR
    assets:eur  -1 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	userTol, _ := decimal.NewFromString("0.001")
	result := CheckBalance(&journal.Transactions[0], userTol, nil)
	assert.True(t, result.Balanced, "precision tolerance 0.5 should win over user 0.001")
}

func TestCheckBalance_DirectivePrecisionAsFloor(t *testing.T) {
	input := `commodity $1,000.00

2024-01-15 buy
    assets:stock  3 AAPL @ $33.333
    assets:cash  -$100`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	dp := ExtractDirectivePrecisions(journal.Directives)
	result := CheckBalance(&journal.Transactions[0], decimal.Zero, dp)
	assert.True(t, result.Balanced,
		"with commodity directive precision 2 → tolerance 0.005, diff 0.001 should balance")
}

func TestCheckBalance_NoDirective_IntegerPrecision(t *testing.T) {
	input := `2024-01-15 buy
    assets:stock  3 AAPL @ $33.333
    assets:cash  -$100`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.True(t, result.Balanced,
		"without directive, precision 0 → tolerance 0.5, diff 0.001 should balance")
}

func TestCheckBalance_DirectivePrecisionLowerThanTransaction(t *testing.T) {
	input := `commodity $1,000.00

2024-01-15 exchange
    assets:foreign  3.000 USD @ 0.33510 EUR
    assets:eur  -1.000 EUR`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	dp := ExtractDirectivePrecisions(journal.Directives)
	result := CheckBalance(&journal.Transactions[0], decimal.Zero, dp)
	assert.False(t, result.Balanced,
		"transaction precision 3 > directive precision 2 → use 3, tolerance 0.0005, diff 0.0053 should NOT balance")
}

func TestExtractDirectivePrecisions(t *testing.T) {
	input := `commodity $1,000.00
commodity 1.00000000 BTC

2024-01-15 test
    expenses:food  $50
    assets:cash  $-50`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)

	precisions := ExtractDirectivePrecisions(journal.Directives)
	assert.Equal(t, int32(2), precisions["$"])
	assert.Equal(t, int32(8), precisions["BTC"])
}

func TestExtractDirectivePrecisions_DefaultCommodity(t *testing.T) {
	input := `D $1,000.00

2024-01-15 test
    expenses:food  $50
    assets:cash  $-50`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)

	precisions := ExtractDirectivePrecisions(journal.Directives)
	assert.Equal(t, int32(2), precisions["$"])
}

func TestExtractDirectivePrecisions_Empty(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  $50
    assets:cash  $-50`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)

	precisions := ExtractDirectivePrecisions(journal.Directives)
	assert.Empty(t, precisions)
}

func TestCheckBalance_MultiCommodity_OneExceedsTolerance(t *testing.T) {
	// EUR: 3.00 * 0.333 = 0.999; balance = -0.001; prec 2, tol 0.005 → OK
	// CHF: 3.00 * 0.337 = 1.011; balance = 0.011; prec 2, tol 0.005 → EXCEEDS
	input := `2024-01-15 exchange
    assets:usd1  3.00 USD @ 0.333 EUR
    assets:eur  -1.00 EUR
    assets:usd2  3.00 GBP @ 0.337 CHF
    assets:chf  -1.00 CHF`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)

	assert.False(t, result.Balanced, "CHF exceeds tolerance even though EUR is within")
	assert.Contains(t, result.Differences, "CHF")
}

func TestExtractDirectivePrecisions_ConflictingCommodityAndD(t *testing.T) {
	// When both commodity and D define precision for the same symbol,
	// the last directive in parse order wins (last-write-wins).
	input := `commodity $1,000.00
D $1,000.000

2024-01-15 test
    expenses:food  $50
    assets:cash  $-50`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)

	precisions := ExtractDirectivePrecisions(journal.Directives)
	assert.Equal(t, int32(3), precisions["$"],
		"D directive (precision 3) appears after commodity directive (precision 2), last-write-wins → 3")
}

func TestCheckBalance_LotCost_UnitPrice(t *testing.T) {
	// 10 AAPL {$150} = 10 * $150 = $1500; cash -$1500 → balanced
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL {$150}
    assets:cash  $-1500`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.True(t, result.Balanced, "lot unit cost {$150} should work like @ $150 for balance")
}

func TestCheckBalance_LotCost_TotalPrice(t *testing.T) {
	// 10 AAPL {{$1500}} = total $1500; cash -$1500 → balanced
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL {{$1500}}
    assets:cash  $-1500`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.True(t, result.Balanced, "lot total cost {{$1500}} should work like @@ $1500 for balance")
}

func TestCheckBalance_LotCost_WithCost_CostWins(t *testing.T) {
	// 10 AAPL {$150} @ $180 → Cost ($180) used for balance, not LotPrice ($150)
	// Balance: 10 * $180 = $1800; cash -$1800 → balanced
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL {$150} @ $180
    assets:cash  $-1800`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.True(t, result.Balanced, "when both Cost and LotPrice exist, Cost @ should be used for balance")
}

func TestCheckBalance_LotCost_Unbalanced(t *testing.T) {
	// 10 AAPL {$150} = $1500; cash -$1400 → off by $100
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL {$150}
    assets:cash  $-1400`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.False(t, result.Balanced, "lot cost with wrong amount should be unbalanced")
	assert.Equal(t, decimal.NewFromInt(100), result.Differences["$"])
}

func TestCheckBalance_LotCost_PrecisionMapping(t *testing.T) {
	// 3.000 AAPL {$33.337} = 3.000 * 33.337 = 100.011; diff = 0.011
	// Without fix: prec 3 mapped to AAPL (wrong); $ gets prec 0 → tolerance 0.5 → balanced
	// With fix: prec 3 mapped to $ (correct); $ gets max(3,0) = 3 → tolerance 0.0005 → unbalanced
	input := `2024-01-15 buy
    assets:stocks  3.000 AAPL {$33.337}
    assets:cash  -$100`

	journal, errs := parser.Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	result := CheckBalance(&journal.Transactions[0], decimal.Zero, nil)
	assert.False(t, result.Balanced,
		"lot cost precision mapping: posting prec 3 mapped to $ → tolerance 0.0005, diff 0.011 exceeds tolerance")
}
