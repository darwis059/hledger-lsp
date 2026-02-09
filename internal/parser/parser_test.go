package parser

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/juev/hledger-lsp/internal/ast"
)

func TestParser_SimpleTransaction(t *testing.T) {
	input := `2024-01-15 grocery store
    expenses:food  $50.00
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, 2024, tx.Date.Year)
	assert.Equal(t, 1, tx.Date.Month)
	assert.Equal(t, 15, tx.Date.Day)
	assert.Equal(t, "grocery store", tx.Description)
	assert.Equal(t, ast.StatusNone, tx.Status)
	require.Len(t, tx.Postings, 2)

	p1 := tx.Postings[0]
	assert.Equal(t, "expenses:food", p1.Account.Name)
	require.NotNil(t, p1.Amount)
	assert.Equal(t, "$", p1.Amount.Commodity.Symbol)
	assert.True(t, p1.Amount.Quantity.Equal(decimal.NewFromFloat(50.00)))

	p2 := tx.Postings[1]
	assert.Equal(t, "assets:cash", p2.Account.Name)
	assert.Nil(t, p2.Amount)
}

func TestParser_TransactionWithStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		status ast.Status
	}{
		{
			name: "cleared",
			input: `2024-01-15 * grocery store
    expenses:food  $50
    assets:cash`,
			status: ast.StatusCleared,
		},
		{
			name: "pending",
			input: `2024-01-15 ! grocery store
    expenses:food  $50
    assets:cash`,
			status: ast.StatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)
			assert.Equal(t, tt.status, journal.Transactions[0].Status)
		})
	}
}

func TestParser_TransactionWithCode(t *testing.T) {
	input := `2024-01-15 * (12345) grocery store
    expenses:food  $50
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)
	assert.Equal(t, "12345", journal.Transactions[0].Code)
}

func TestParser_TransactionWithPayeeAndNote(t *testing.T) {
	input := `2024-01-15 Grocery Store | weekly shopping
    expenses:food  $50
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)
	assert.Equal(t, "Grocery Store", journal.Transactions[0].Payee)
	assert.Equal(t, "weekly shopping", journal.Transactions[0].Note)
}

func TestParser_Issue12_CompleteJournal(t *testing.T) {
	input := `decimal-mark .

2026-01-20 18a Brock Street Cafe
    Expenses:Eating out    £10
    Assets:Checking

2026-01-20 J Random Hacker
    Expenses:Contracting   £50
    Assets:Checking

2026-01-20 Australian friend
    Expenses:Kangaroo food   AU$50
    Assets:Checking

2026-01-20 salary
    Assets:Checking   £2000
    Assets:Pension    500 PensionCredits @@ £500
    Income:Salary

2026-01-20 Steam | Magic: The Gathering Arena
    Expenses:Games   £5
    Assets:Checking

2026-01-20 Art shop | "Mona Lisa" print
    Expenses:Home    £20
    Assets:Checking

2026-01-20 Opening balances
    Assets:Checking          £3026.13
    Equity:Opening balances
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 7)

	// Verify decimal-mark directive
	require.Len(t, journal.Directives, 1)
	dir, ok := journal.Directives[0].(ast.DecimalMarkDirective)
	require.True(t, ok)
	assert.Equal(t, ".", dir.Mark)

	// Verify transaction 1: numbers in payee
	tx1 := journal.Transactions[0]
	assert.Equal(t, "18a Brock Street Cafe", tx1.Description)

	// Verify transaction 3: multi-char currency
	tx3 := journal.Transactions[2]
	require.Len(t, tx3.Postings, 2)
	assert.Equal(t, "AU$", tx3.Postings[0].Amount.Commodity.Symbol)

	// Verify transaction 4: mixed-case commodity
	tx4 := journal.Transactions[3]
	require.Len(t, tx4.Postings, 3)
	assert.Equal(t, "PensionCredits", tx4.Postings[1].Amount.Commodity.Symbol)

	// Verify transaction 5: colon after pipe
	tx5 := journal.Transactions[4]
	assert.Equal(t, "Steam", tx5.Payee)
	assert.Equal(t, "Magic: The Gathering Arena", tx5.Note)

	// Verify transaction 6: quotes in description
	tx6 := journal.Transactions[5]
	assert.Equal(t, "Art shop", tx6.Payee)
	assert.Equal(t, `"Mona Lisa" print`, tx6.Note)

	// Verify transaction 7: year-like amounts
	tx7 := journal.Transactions[6]
	require.Len(t, tx7.Postings, 2)
	assert.Equal(t, "3026.13", tx7.Postings[0].Amount.RawQuantity)
}

func TestParser_MixedCaseCommodity(t *testing.T) {
	input := `2026-01-20 salary
    Assets:Checking   £2000
    Assets:Pension    500 PensionCredits @@ £500
    Income:Salary
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Postings, 3)

	p2 := tx.Postings[1]
	assert.Equal(t, "Assets:Pension", p2.Account.Name)
	require.NotNil(t, p2.Amount)
	assert.Equal(t, "500", p2.Amount.RawQuantity)
	assert.Equal(t, "PensionCredits", p2.Amount.Commodity.Symbol)
	require.NotNil(t, p2.Cost)
	assert.True(t, p2.Cost.IsTotal)
}

func TestParser_MultiCharCurrency(t *testing.T) {
	input := `2026-01-20 Australian friend
    Expenses:Kangaroo food   AU$50
    Assets:Checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Postings, 2)

	p1 := tx.Postings[0]
	assert.Equal(t, "Expenses:Kangaroo food", p1.Account.Name)
	require.NotNil(t, p1.Amount)
	assert.Equal(t, "AU$", p1.Amount.Commodity.Symbol)
	assert.Equal(t, "50", p1.Amount.RawQuantity)
}

func TestParser_YearLikeAmounts(t *testing.T) {
	input := `2026-01-20 Opening balances
    Assets:Checking          £3026.13
    Equity:Opening balances
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Postings, 2)

	p1 := tx.Postings[0]
	assert.Equal(t, "Assets:Checking", p1.Account.Name)
	require.NotNil(t, p1.Amount)
	assert.Equal(t, "3026.13", p1.Amount.RawQuantity)
	assert.Equal(t, "£", p1.Amount.Commodity.Symbol)
}

func TestParser_ColonAfterPipe(t *testing.T) {
	input := `2026-01-20 Steam | Magic: The Gathering Arena
    Expenses:Games   £5
    Assets:Checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, "Steam", tx.Payee)
	assert.Equal(t, "Magic: The Gathering Arena", tx.Note)
}

func TestParser_QuotesInDescription(t *testing.T) {
	input := `2026-01-20 Art shop | "Mona Lisa" print
    Expenses:Home    £20
    Assets:Checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, "Art shop", tx.Payee)
	assert.Equal(t, `"Mona Lisa" print`, tx.Note)
}

func TestParser_PayeeWithNumbers(t *testing.T) {
	input := `2026-01-20 18a Brock Street Cafe
    Expenses:Eating out    £10
    Assets:Checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, "18a Brock Street Cafe", tx.Description)
	assert.Equal(t, 2026, tx.Date.Year)
	assert.Equal(t, 1, tx.Date.Month)
	assert.Equal(t, 20, tx.Date.Day)
}

func TestParser_PayeeDirective(t *testing.T) {
	input := "payee Whole Foods\n"
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	directive, ok := journal.Directives[0].(ast.PayeeDirective)
	require.True(t, ok, "expected PayeeDirective, got %T", journal.Directives[0])
	assert.Equal(t, "Whole Foods", directive.Name)
}

func TestParser_TagDirective(t *testing.T) {
	input := "tag project\n"
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	directive, ok := journal.Directives[0].(ast.TagDirective)
	require.True(t, ok, "expected TagDirective, got %T", journal.Directives[0])
	assert.Equal(t, "project", directive.Name)
}

func TestParser_AliasDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		original string
		alias    string
		isRegex  bool
	}{
		{
			name:     "simple alias",
			input:    "alias checking = assets:bank:checking\n",
			original: "checking",
			alias:    "assets:bank:checking",
			isRegex:  false,
		},
		// TODO: Regex aliases with slashes need special handling in lexer
		// {
		// 	name:     "regex alias",
		// 	input:    "alias /foo/ = bar\n",
		// 	original: "foo",
		// 	alias:    "bar",
		// 	isRegex:  true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Directives, 1)

			directive, ok := journal.Directives[0].(ast.AliasDirective)
			require.True(t, ok, "expected AliasDirective, got %T", journal.Directives[0])
			assert.Equal(t, tt.original, directive.Original)
			assert.Equal(t, tt.alias, directive.Alias)
			assert.Equal(t, tt.isRegex, directive.IsRegex)
		})
	}
}

func TestParser_AutoPostingRule(t *testing.T) {
	input := `= expenses:food
    (budget:food)    $-1
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.AutoPostingRules, 1)

	rule := journal.AutoPostingRules[0]
	assert.Equal(t, "expenses:food", rule.Query)
	require.Len(t, rule.Postings, 1)
	assert.Equal(t, "budget:food", rule.Postings[0].Account.Name)
}

func TestParser_PeriodicTransaction(t *testing.T) {
	input := `~ monthly
    expenses:rent    $2000
    assets:checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.PeriodicTransactions, 1)

	ptx := journal.PeriodicTransactions[0]
	assert.Equal(t, "monthly", ptx.Period)
	require.Len(t, ptx.Postings, 2)
	assert.Equal(t, "expenses:rent", ptx.Postings[0].Account.Name)
}

func TestParser_CommentBlock(t *testing.T) {
	input := `account assets:checking

comment
This is a multi-line comment
that should be ignored
2024-01-01 Invalid transaction
    expenses:food  $50
end comment

2024-01-15 Real transaction
    expenses:food  $30
    assets:checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)

	// Should have 1 account directive and 1 transaction (comment block ignored)
	require.Len(t, journal.Directives, 1)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, 2024, tx.Date.Year)
	assert.Equal(t, 1, tx.Date.Month)
	assert.Equal(t, 15, tx.Date.Day)
	assert.Equal(t, "Real transaction", tx.Description)
}

func TestParser_DecimalMarkDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "period decimal mark",
			input:    "decimal-mark .\n",
			expected: ".",
			wantErr:  false,
		},
		{
			name:     "comma decimal mark",
			input:    "decimal-mark ,\n",
			expected: ",",
			wantErr:  false,
		},
		{
			name:     "invalid mark",
			input:    "decimal-mark :\n",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			if tt.wantErr {
				require.NotEmpty(t, errs)
				return
			}
			require.Empty(t, errs)
			require.Len(t, journal.Directives, 1, "expected 1 directive, got %d", len(journal.Directives))

			directive, ok := journal.Directives[0].(ast.DecimalMarkDirective)
			require.True(t, ok, "expected DecimalMarkDirective, got %T", journal.Directives[0])
			assert.Equal(t, tt.expected, directive.Mark)
		})
	}
}

func TestParser_PostingWithCost(t *testing.T) {
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL @ $150
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	assert.Equal(t, "assets:stocks", p.Account.Name)
	require.NotNil(t, p.Amount)
	assert.Equal(t, "AAPL", p.Amount.Commodity.Symbol)
	assert.True(t, p.Amount.Quantity.Equal(decimal.NewFromInt(10)))

	require.NotNil(t, p.Cost)
	assert.False(t, p.Cost.IsTotal)
	assert.Equal(t, "$", p.Cost.Amount.Commodity.Symbol)
	assert.True(t, p.Cost.Amount.Quantity.Equal(decimal.NewFromInt(150)))
}

func TestParser_PostingWithTotalCost(t *testing.T) {
	input := `2024-01-15 buy stocks
    assets:stocks  10 AAPL @@ $1500
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.Cost)
	assert.True(t, p.Cost.IsTotal)
	assert.True(t, p.Cost.Amount.Quantity.Equal(decimal.NewFromInt(1500)))
}

func TestParser_BalanceAssertion(t *testing.T) {
	input := `2024-01-15 check balance
    assets:checking  $100 = $1000
    income:salary`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.BalanceAssertion)
	assert.False(t, p.BalanceAssertion.IsStrict)
	assert.True(t, p.BalanceAssertion.Amount.Quantity.Equal(decimal.NewFromInt(1000)))
}

func TestParser_StrictBalanceAssertion(t *testing.T) {
	input := `2024-01-15 check balance
    assets:checking  $100 == $1000
    income:salary`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.BalanceAssertion)
	assert.True(t, p.BalanceAssertion.IsStrict)
}

func TestParser_AccountDirective(t *testing.T) {
	input := `account expenses:food`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.AccountDirective)
	require.True(t, ok)
	assert.Equal(t, "expenses:food", dir.Account.Name)
}

func TestParser_CommodityDirective(t *testing.T) {
	input := `commodity $1000.00`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.CommodityDirective)
	require.True(t, ok)
	assert.Equal(t, "$", dir.Commodity.Symbol)
}

func TestParser_IncludeDirective(t *testing.T) {
	input := `include accounts.journal`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Includes, 1)

	inc := journal.Includes[0]
	assert.Equal(t, "accounts.journal", inc.Path)
}

func TestParser_Comment(t *testing.T) {
	input := `; This is a comment
2024-01-15 test
    expenses:misc  $10
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Comments, 1)
	assert.Equal(t, " This is a comment", journal.Comments[0].Text)
	require.Len(t, journal.Transactions, 1)
}

func TestParser_NegativeAmount(t *testing.T) {
	input := `2024-01-15 withdrawal
    assets:cash  $-50
    assets:bank`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	assert.True(t, p.Amount.Quantity.Equal(decimal.NewFromInt(-50)))
}

func TestParser_MultipleTransactions(t *testing.T) {
	input := `2024-01-15 first
    expenses:food  $50
    assets:cash

2024-01-16 second
    expenses:transport  $20
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 2)

	assert.Equal(t, "first", journal.Transactions[0].Description)
	assert.Equal(t, "second", journal.Transactions[1].Description)
}

func TestParser_CommodityRight(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  50 EUR
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	assert.Equal(t, "EUR", p.Amount.Commodity.Symbol)
	assert.Equal(t, ast.CommodityRight, p.Amount.Commodity.Position)
	assert.True(t, p.Amount.Quantity.Equal(decimal.NewFromInt(50)))
}

func TestParser_DateFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		year  int
		month int
		day   int
	}{
		{
			name: "dashes",
			input: `2024-01-15 test
    e:f  $1
    a:c`,
			year: 2024, month: 1, day: 15,
		},
		{
			name: "slashes",
			input: `2024/01/15 test
    e:f  $1
    a:c`,
			year: 2024, month: 1, day: 15,
		},
		{
			name: "dots",
			input: `2024.01.15 test
    e:f  $1
    a:c`,
			year: 2024, month: 1, day: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)
			assert.Equal(t, tt.year, journal.Transactions[0].Date.Year)
			assert.Equal(t, tt.month, journal.Transactions[0].Date.Month)
			assert.Equal(t, tt.day, journal.Transactions[0].Date.Day)
		})
	}
}

func TestParser_ErrorRecovery(t *testing.T) {
	input := `2024-01-15 valid transaction
    expenses:food  $50
    assets:cash

invalid line without date

2024-01-16 another valid
    expenses:misc  $10
    assets:cash`

	journal, errs := Parse(input)
	assert.NotEmpty(t, errs)
	assert.Len(t, journal.Transactions, 2)
}

func TestParser_Date2(t *testing.T) {
	input := `2024-01-15=2024-01-20 transaction with date2
    expenses:food  $50
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, 2024, tx.Date.Year)
	assert.Equal(t, 1, tx.Date.Month)
	assert.Equal(t, 15, tx.Date.Day)

	require.NotNil(t, tx.Date2)
	assert.Equal(t, 2024, tx.Date2.Year)
	assert.Equal(t, 1, tx.Date2.Month)
	assert.Equal(t, 20, tx.Date2.Day)

	assert.Equal(t, "transaction with date2", tx.Description)
}

func TestParser_Date2Formats(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		year2  int
		month2 int
		day2   int
	}{
		{
			name: "dashes",
			input: `2024-01-15=2024-01-20 test
    e:f  $1
    a:c`,
			year2: 2024, month2: 1, day2: 20,
		},
		{
			name: "slashes",
			input: `2024/01/15=2024/01/20 test
    e:f  $1
    a:c`,
			year2: 2024, month2: 1, day2: 20,
		},
		{
			name: "mixed separators",
			input: `2024-01-15=2024/01/20 test
    e:f  $1
    a:c`,
			year2: 2024, month2: 1, day2: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)
			require.NotNil(t, journal.Transactions[0].Date2)
			assert.Equal(t, tt.year2, journal.Transactions[0].Date2.Year)
			assert.Equal(t, tt.month2, journal.Transactions[0].Date2.Month)
			assert.Equal(t, tt.day2, journal.Transactions[0].Date2.Day)
		})
	}
}

func TestParser_PriceDirective(t *testing.T) {
	input := `P 2024-01-15 EUR $1.08`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.PriceDirective)
	require.True(t, ok)
	assert.Equal(t, 2024, dir.Date.Year)
	assert.Equal(t, 1, dir.Date.Month)
	assert.Equal(t, 15, dir.Date.Day)
	assert.Equal(t, "EUR", dir.Commodity.Symbol)
	assert.Equal(t, "$", dir.Price.Commodity.Symbol)
	assert.True(t, dir.Price.Quantity.Equal(decimal.NewFromFloat(1.08)))
}

func TestParser_PriceDirectiveVariants(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		commodity string
		priceSym  string
		priceQty  float64
	}{
		{
			name:      "stock price",
			input:     `P 2024-01-15 AAPL $185.50`,
			commodity: "AAPL",
			priceSym:  "$",
			priceQty:  185.50,
		},
		{
			name:      "crypto price",
			input:     `P 2024-01-15 BTC $42000.00`,
			commodity: "BTC",
			priceSym:  "$",
			priceQty:  42000.00,
		},
		{
			name:      "currency with right commodity",
			input:     `P 2024-01-15 USD 0.92 EUR`,
			commodity: "USD",
			priceSym:  "EUR",
			priceQty:  0.92,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Directives, 1)

			dir, ok := journal.Directives[0].(ast.PriceDirective)
			require.True(t, ok)
			assert.Equal(t, tt.commodity, dir.Commodity.Symbol)
			assert.Equal(t, tt.priceSym, dir.Price.Commodity.Symbol)
			assert.True(t, dir.Price.Quantity.Equal(decimal.NewFromFloat(tt.priceQty)))
		})
	}
}

func TestParser_VirtualPostings(t *testing.T) {
	input := `2024-01-15 transaction with virtual postings
    expenses:food           $50
    assets:cash            $-50
    [budget:food]          $-50
    [budget:available]      $50
    (tracking:note)`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Postings, 5)

	assert.Equal(t, ast.VirtualNone, tx.Postings[0].Virtual)
	assert.Equal(t, "expenses:food", tx.Postings[0].Account.Name)

	assert.Equal(t, ast.VirtualNone, tx.Postings[1].Virtual)
	assert.Equal(t, "assets:cash", tx.Postings[1].Account.Name)

	assert.Equal(t, ast.VirtualBalanced, tx.Postings[2].Virtual)
	assert.Equal(t, "budget:food", tx.Postings[2].Account.Name)

	assert.Equal(t, ast.VirtualBalanced, tx.Postings[3].Virtual)
	assert.Equal(t, "budget:available", tx.Postings[3].Account.Name)

	assert.Equal(t, ast.VirtualUnbalanced, tx.Postings[4].Virtual)
	assert.Equal(t, "tracking:note", tx.Postings[4].Account.Name)
}

func TestParser_VirtualPostingWithAmount(t *testing.T) {
	input := `2024-01-15 test
    (opening:balance)  $1000
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	assert.Equal(t, ast.VirtualUnbalanced, p.Virtual)
	assert.Equal(t, "opening:balance", p.Account.Name)
	require.NotNil(t, p.Amount)
	assert.True(t, p.Amount.Quantity.Equal(decimal.NewFromInt(1000)))
}

func TestParser_TagsInTransactionComment(t *testing.T) {
	input := `2024-01-15 Business dinner  ; client:acme, project:alpha
    expenses:meals  $50
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Comments, 1)
	require.Len(t, tx.Comments[0].Tags, 2)

	assert.Equal(t, "client", tx.Comments[0].Tags[0].Name)
	assert.Equal(t, "acme", tx.Comments[0].Tags[0].Value)

	assert.Equal(t, "project", tx.Comments[0].Tags[1].Name)
	assert.Equal(t, "alpha", tx.Comments[0].Tags[1].Value)
}

func TestParser_TagWithoutValue(t *testing.T) {
	input := `2024-01-15 test  ; billable:
    expenses:meals  $50
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Comments, 1)
	require.Len(t, tx.Comments[0].Tags, 1)

	assert.Equal(t, "billable", tx.Comments[0].Tags[0].Name)
	assert.Equal(t, "", tx.Comments[0].Tags[0].Value)
}

func TestParser_TagsInPostingComment(t *testing.T) {
	input := `2024-01-15 test
    expenses:meals  $50  ; date:2024-01-16, receipt:123
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.Len(t, p.Tags, 2)

	assert.Equal(t, "date", p.Tags[0].Name)
	assert.Equal(t, "2024-01-16", p.Tags[0].Value)

	assert.Equal(t, "receipt", p.Tags[1].Name)
	assert.Equal(t, "123", p.Tags[1].Value)
}

func TestParser_YearDirective(t *testing.T) {
	tests := []struct {
		name  string
		input string
		year  int
	}{
		{
			name:  "Y directive",
			input: "Y2026",
			year:  2026,
		},
		{
			name:  "Y with space",
			input: "Y 2026",
			year:  2026,
		},
		{
			name:  "year directive",
			input: "year 2025",
			year:  2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Directives, 1)

			dir, ok := journal.Directives[0].(ast.YearDirective)
			require.True(t, ok)
			assert.Equal(t, tt.year, dir.Year)
		})
	}
}

func TestParser_PartialDate(t *testing.T) {
	input := `Y2026
01-02 Магазин
    Расходы:Продукты  100 RUB
    Активы:Банк`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, 2026, tx.Date.Year)
	assert.Equal(t, 1, tx.Date.Month)
	assert.Equal(t, 2, tx.Date.Day)
	assert.Equal(t, "Магазин", tx.Description)
}

func TestParser_PartialDateWithoutYear(t *testing.T) {
	input := `01-02 test
    e:f  $1
    a:c`

	_, errs := Parse(input)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Message, "partial date requires Y directive")
}

func TestParser_UnicodeAccountDirective(t *testing.T) {
	input := `account Активы:Банк`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.AccountDirective)
	require.True(t, ok)
	assert.Equal(t, "Активы:Банк", dir.Account.Name)
}

func TestParser_UnicodeTransaction(t *testing.T) {
	input := `2024-01-15 Покупка продуктов
    Расходы:Продукты  100 RUB
    Активы:Наличные`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	assert.Equal(t, "Покупка продуктов", tx.Description)
	assert.Equal(t, "Расходы:Продукты", tx.Postings[0].Account.Name)
	assert.Equal(t, "Активы:Наличные", tx.Postings[1].Account.Name)
}

func TestParser_CommodityDirectiveWithFormat(t *testing.T) {
	input := `commodity RUB
  format 1.000,00 RUB`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.CommodityDirective)
	require.True(t, ok)
	assert.Equal(t, "RUB", dir.Commodity.Symbol)
	assert.Equal(t, "1.000,00 RUB", dir.Format)
}

func TestParser_CommodityDirectiveMultipleSubdirs(t *testing.T) {
	input := `commodity EUR
  format 1.000,00 EUR
  note European currency`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.CommodityDirective)
	require.True(t, ok)
	assert.Equal(t, "EUR", dir.Commodity.Symbol)
	assert.Equal(t, "1.000,00 EUR", dir.Format)
	assert.Equal(t, "European currency", dir.Note)
}

func TestParser_CommodityDirectiveInlineFormat(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedSymbol string
		expectedFormat string
	}{
		{"symbol right USD", "commodity 1.000,00 USD", "USD", "1.000,00 USD"},
		{"symbol right EUR", "commodity 1.000,00 EUR", "EUR", "1.000,00 EUR"},
		{"symbol right BTC", "commodity 1.00000000 BTC", "BTC", "1.00000000 BTC"},
		{"symbol left dollar", "commodity $1000.00", "$", "$1000.00"},
		{"symbol left euro", "commodity €1.000,00", "€", "€1.000,00"},
		{"symbol left Turkish lira", "commodity ₺1.000,00", "₺", "₺1.000,00"},
		{"symbol right Turkish lira", "commodity 1.000,00 ₺", "₺", "1.000,00 ₺"},
		{"symbol left Indian rupee", "commodity ₹1,00,000.00", "₹", "₹1,00,000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Directives, 1)

			dir, ok := journal.Directives[0].(ast.CommodityDirective)
			require.True(t, ok)
			assert.Equal(t, tt.expectedSymbol, dir.Commodity.Symbol)
			assert.Equal(t, tt.expectedFormat, dir.Format, "inline format should be saved")
		})
	}
}

func TestParser_DefaultCommodityDirective(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedSymbol string
		expectedFormat string
	}{
		{"USD format with comma", "D $1,000.00", "$", "$1,000.00"},
		{"USD format no comma", "D $1000.00", "$", "$1000.00"},
		{"EUR format", "D 1.000,00 EUR", "EUR", "1.000,00 EUR"},
		{"RUB format", "D 1 000,00 RUB", "RUB", "1 000,00 RUB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Directives, 1)

			dir, ok := journal.Directives[0].(ast.DefaultCommodityDirective)
			require.True(t, ok, "expected DefaultCommodityDirective, got %T", journal.Directives[0])
			assert.Equal(t, tt.expectedSymbol, dir.Symbol)
			assert.Equal(t, tt.expectedFormat, dir.Format)
		})
	}
}

func TestParser_DefaultCommodityWithTransaction(t *testing.T) {
	input := `D $1,000.00

2024-01-15 test
    expenses:food  $50.00
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs, "parse errors: %v", errs)
	require.Len(t, journal.Directives, 1, "expected 1 directive")
	require.Len(t, journal.Transactions, 1, "expected 1 transaction")

	dir, ok := journal.Directives[0].(ast.DefaultCommodityDirective)
	require.True(t, ok, "expected DefaultCommodityDirective, got %T", journal.Directives[0])
	assert.Equal(t, "$", dir.Symbol)
	assert.Equal(t, "$1,000.00", dir.Format)

	// Verify transaction parsed correctly
	tx := journal.Transactions[0]
	assert.Equal(t, "test", tx.Description)
	require.Len(t, tx.Postings, 2)
	assert.Equal(t, "expenses:food", tx.Postings[0].Account.Name)
}

func TestParser_AccountDirectiveWithComment(t *testing.T) {
	input := `account Активы  ; type:Asset`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.AccountDirective)
	require.True(t, ok)
	assert.Equal(t, "Активы", dir.Account.Name)
	assert.Contains(t, dir.Comment, "type:Asset")
	require.Len(t, dir.Tags, 1)
	assert.Equal(t, "type", dir.Tags[0].Name)
	assert.Equal(t, "Asset", dir.Tags[0].Value)
}

func TestParser_AccountDirectiveWithSubdirs(t *testing.T) {
	input := `account expenses:food
  alias food
  note Food and groceries`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Directives, 1)

	dir, ok := journal.Directives[0].(ast.AccountDirective)
	require.True(t, ok)
	assert.Equal(t, "expenses:food", dir.Account.Name)
	assert.Equal(t, "food", dir.Subdirs["alias"])
	assert.Equal(t, "Food and groceries", dir.Subdirs["note"])
}

func TestParser_SignBeforeCommodity(t *testing.T) {
	input := `2024-01-15 test
    assets:cash  -$100
    expenses:food`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.Amount)
	assert.True(t, p.Amount.Quantity.Equal(decimal.NewFromInt(-100)))
	assert.Equal(t, "$", p.Amount.Commodity.Symbol)
}

func TestParser_SpaceGroupedNumber(t *testing.T) {
	input := `2024-01-15 test
    assets:cash  3 037 850,96 RUB
    expenses:food`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.Amount)
	expected, _ := decimal.NewFromString("3037850.96")
	assert.True(t, p.Amount.Quantity.Equal(expected), "got %s", p.Amount.Quantity.String())
	assert.Equal(t, "RUB", p.Amount.Commodity.Symbol)
}

func TestParser_ScientificNotation(t *testing.T) {
	input := `2024-01-15 test
    assets:cash  1E3 USD
    expenses:food`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.Amount)
	expected := decimal.NewFromInt(1000)
	assert.True(t, p.Amount.Quantity.Equal(expected), "got %s", p.Amount.Quantity.String())
}

func TestParser_PositiveSignBeforeCommodity(t *testing.T) {
	input := `2024-01-15 test
    assets:cash  +$100
    expenses:food`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]
	require.NotNil(t, p.Amount)
	assert.True(t, p.Amount.Quantity.Equal(decimal.NewFromInt(100)), "got %s", p.Amount.Quantity.String())
	assert.Equal(t, "$", p.Amount.Commodity.Symbol)
}

func TestParser_EuropeanNumberFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "european with dot grouping",
			input: `2024-01-15 test
    assets:cash  1.234.567,89 EUR
    expenses:food`,
			expected: "1234567.89",
		},
		{
			name: "us with comma grouping",
			input: `2024-01-15 test
    assets:cash  1,234,567.89 USD
    expenses:food`,
			expected: "1234567.89",
		},
		{
			name: "multiple dots as grouping",
			input: `2024-01-15 test
    assets:cash  1.234.567 EUR
    expenses:food`,
			expected: "1234567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)

			p := journal.Transactions[0].Postings[0]
			require.NotNil(t, p.Amount)
			expected, _ := decimal.NewFromString(tt.expected)
			assert.True(t, p.Amount.Quantity.Equal(expected), "got %s, want %s", p.Amount.Quantity.String(), tt.expected)
		})
	}
}

func TestParser_HledgerNumberFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "dots as grouping 1.2.3 equals 123",
			input: `2024-01-15 test
    assets:cash  1.2.3 EUR
    expenses:food`,
			expected: "123",
		},
		{
			name: "mixed format 1.2,3 equals 12.3",
			input: `2024-01-15 test
    assets:cash  1.2,3 EUR
    expenses:food`,
			expected: "12.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)

			p := journal.Transactions[0].Postings[0]
			require.NotNil(t, p.Amount)
			expected, _ := decimal.NewFromString(tt.expected)
			assert.True(t, p.Amount.Quantity.Equal(expected), "got %s, want %s", p.Amount.Quantity.String(), tt.expected)
		})
	}
}

func TestParseTransactionWithTrailingWhitespace(t *testing.T) {
	input := "2024-01-15 test\n    account:a  100\n    account:b\n    \n"

	journal, errs := Parse(input)

	require.Empty(t, errs, "trailing whitespace should not cause errors")
	require.Len(t, journal.Transactions, 1)
	require.Len(t, journal.Transactions[0].Postings, 2)
}

func TestParseTransactionWithEmptyIndentedLines(t *testing.T) {
	input := "2024-01-15 test\n    account:a  100\n    \n    account:b\n"

	journal, errs := Parse(input)

	require.Empty(t, errs, "empty indented line between postings should not cause errors")
	require.Len(t, journal.Transactions, 1)
	require.Len(t, journal.Transactions[0].Postings, 2)
}

func TestParseTransactionWithOnlySpacesLine(t *testing.T) {
	input := "2024-01-15 test\n    account:a  100\n    account:b\n        \n"

	journal, errs := Parse(input)

	require.Empty(t, errs, "line with only spaces should not cause errors")
	require.Len(t, journal.Transactions, 1)
	require.Len(t, journal.Transactions[0].Postings, 2)
}

func TestParser_CommodityRange(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSymbol   string
		wantPosition ast.CommodityPosition
	}{
		{
			name: "commodity left (currency symbol)",
			input: `2024-01-15 test
    expenses:food  $50
    assets:cash`,
			wantSymbol:   "$",
			wantPosition: ast.CommodityLeft,
		},
		{
			name: "commodity right",
			input: `2024-01-15 test
    expenses:food  50 EUR
    assets:cash`,
			wantSymbol:   "EUR",
			wantPosition: ast.CommodityRight,
		},
		{
			name: "commodity right multi-char",
			input: `2024-01-15 test
    expenses:food  3.000 FF
    assets:cash`,
			wantSymbol:   "FF",
			wantPosition: ast.CommodityRight,
		},
		{
			name: "commodity right mixed case FFf",
			input: `2024-01-15 test
    expenses:food  3.000 FFf
    assets:cash`,
			wantSymbol:   "FFf",
			wantPosition: ast.CommodityRight,
		},
		{
			name: "commodity right lowercase Rub",
			input: `2024-01-15 test
    expenses:food  100 Rub
    assets:cash`,
			wantSymbol:   "Rub",
			wantPosition: ast.CommodityRight,
		},
		{
			name: "commodity right all lowercase hours",
			input: `2024-01-15 test
    work:project  8 hours
    income:salary`,
			wantSymbol:   "hours",
			wantPosition: ast.CommodityRight,
		},
		{
			name: "commodity right cyrillic Руб",
			input: `2024-01-15 test
    expenses:food  100 Руб
    assets:cash`,
			wantSymbol:   "Руб",
			wantPosition: ast.CommodityRight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)

			p := journal.Transactions[0].Postings[0]
			require.NotNil(t, p.Amount)

			commodity := p.Amount.Commodity
			assert.Equal(t, tt.wantSymbol, commodity.Symbol)
			assert.Equal(t, tt.wantPosition, commodity.Position)

			assert.NotZero(t, commodity.Range.Start.Line, "Range.Start.Line should not be zero")
			assert.NotZero(t, commodity.Range.Start.Column, "Range.Start.Column should not be zero")
			assert.NotZero(t, commodity.Range.End.Line, "Range.End.Line should not be zero")
			assert.NotZero(t, commodity.Range.End.Column, "Range.End.Column should not be zero")

			assert.True(t, commodity.Range.End.Column > commodity.Range.Start.Column ||
				commodity.Range.End.Line > commodity.Range.Start.Line,
				"Range.End should be after Range.Start")
		})
	}
}

func TestIsValidCommodityText(t *testing.T) {
	tests := []struct {
		input string
		want  bool
		desc  string
	}{
		{"USD", true, "uppercase letters"},
		{"usd", true, "lowercase letters"},
		{"Rub", true, "mixed case"},
		{"hours", true, "all lowercase"},
		{"USD2024", true, "letters + digits"},
		{"Руб", true, "cyrillic"},
		{"123", false, "digit-only should be rejected"},
		{"", false, "empty string"},
		{"$", false, "special character"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := isValidCommodityText(tt.input)
			assert.Equal(t, tt.want, got, "isValidCommodityText(%q)", tt.input)
		})
	}
}

func TestParser_CommodityRange_InCostAndAssertion(t *testing.T) {
	input := `2024-01-15 test
    expenses:food  50 EUR @ $1.10 = 100 USD
    assets:cash`

	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	p := journal.Transactions[0].Postings[0]

	require.NotNil(t, p.Amount)
	assert.Equal(t, "EUR", p.Amount.Commodity.Symbol)
	assert.NotZero(t, p.Amount.Commodity.Range.End.Line, "Amount commodity Range.End.Line should not be zero")
	assert.NotZero(t, p.Amount.Commodity.Range.End.Column, "Amount commodity Range.End.Column should not be zero")

	require.NotNil(t, p.Cost)
	assert.Equal(t, "$", p.Cost.Amount.Commodity.Symbol)
	assert.NotZero(t, p.Cost.Amount.Commodity.Range.End.Line, "Cost commodity Range.End.Line should not be zero")
	assert.NotZero(t, p.Cost.Amount.Commodity.Range.End.Column, "Cost commodity Range.End.Column should not be zero")

	require.NotNil(t, p.BalanceAssertion)
	assert.Equal(t, "USD", p.BalanceAssertion.Amount.Commodity.Symbol)
	assert.NotZero(t, p.BalanceAssertion.Amount.Commodity.Range.End.Line, "BalanceAssertion commodity Range.End.Line should not be zero")
	assert.NotZero(t, p.BalanceAssertion.Amount.Commodity.Range.End.Column, "BalanceAssertion commodity Range.End.Column should not be zero")
}

func TestParser_ThousandSeparatorSingleDot(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "single dot thousand separator 3.000",
			input: `2024-01-15 test
    expenses:food  3.000 EUR
    assets:cash`,
			expected: "3000",
		},
		{
			name: "single dot decimal 3.00",
			input: `2024-01-15 test
    expenses:food  3.00 EUR
    assets:cash`,
			expected: "3",
		},
		{
			name: "single dot decimal 3.5",
			input: `2024-01-15 test
    expenses:food  3.5 EUR
    assets:cash`,
			expected: "3.5",
		},
		{
			name: "single comma thousand separator 3,000",
			input: `2024-01-15 test
    expenses:food  3,000 EUR
    assets:cash`,
			expected: "3000",
		},
		{
			name: "larger thousand separator 123.456",
			input: `2024-01-15 test
    expenses:food  123.456 EUR
    assets:cash`,
			expected: "123456",
		},
		{
			name: "hundred with decimal 100.50",
			input: `2024-01-15 test
    expenses:food  100.50 EUR
    assets:cash`,
			expected: "100.5",
		},
		{
			name: "small decimal 0.123",
			input: `2024-01-15 test
    expenses:food  0.123 EUR
    assets:cash`,
			expected: "0.123",
		},
		{
			name: "small decimal 0.999",
			input: `2024-01-15 test
    expenses:food  0.999 EUR
    assets:cash`,
			expected: "0.999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)

			p := journal.Transactions[0].Postings[0]
			require.NotNil(t, p.Amount)
			expected, _ := decimal.NewFromString(tt.expected)
			assert.True(t, p.Amount.Quantity.Equal(expected), "got %s, want %s", p.Amount.Quantity.String(), tt.expected)
		})
	}
}

func TestParser_SubdirectivesEdgeCases(t *testing.T) {
	t.Run("subdirective without value", func(t *testing.T) {
		input := `account expenses:food
  hidden`

		journal, errs := Parse(input)
		require.Empty(t, errs)
		require.Len(t, journal.Directives, 1)

		dir, ok := journal.Directives[0].(ast.AccountDirective)
		require.True(t, ok)
		assert.Equal(t, "", dir.Subdirs["hidden"])
	})

	t.Run("comment between subdirectives", func(t *testing.T) {
		input := `account expenses:food
  alias food
  ; this is a comment
  note Food expenses`

		journal, errs := Parse(input)
		require.Empty(t, errs)
		require.Len(t, journal.Directives, 1)

		dir, ok := journal.Directives[0].(ast.AccountDirective)
		require.True(t, ok)
		assert.Equal(t, "food", dir.Subdirs["alias"])
		assert.Equal(t, "Food expenses", dir.Subdirs["note"])
	})

	t.Run("subdirective at EOF without newline", func(t *testing.T) {
		input := "account expenses:food\n  alias food"

		journal, errs := Parse(input)
		require.Empty(t, errs)
		require.Len(t, journal.Directives, 1)

		dir, ok := journal.Directives[0].(ast.AccountDirective)
		require.True(t, ok)
		assert.Equal(t, "food", dir.Subdirs["alias"])
	})

	t.Run("empty line between subdirectives ends parsing", func(t *testing.T) {
		input := `account expenses:food
  alias food

  note Should not be parsed`

		journal, errs := Parse(input)
		// Parser produces error for orphan indent after blank line
		require.NotEmpty(t, errs)
		require.Len(t, journal.Directives, 1)

		dir, ok := journal.Directives[0].(ast.AccountDirective)
		require.True(t, ok)
		assert.Equal(t, "food", dir.Subdirs["alias"])
		_, hasNote := dir.Subdirs["note"]
		assert.False(t, hasNote)
	})

	t.Run("commodity with multiple subdirectives", func(t *testing.T) {
		input := `commodity EUR
  format 1.000,00 EUR
  alias €
  note European currency
  nomarket`

		journal, errs := Parse(input)
		require.Empty(t, errs)
		require.Len(t, journal.Directives, 1)

		dir, ok := journal.Directives[0].(ast.CommodityDirective)
		require.True(t, ok)
		assert.Equal(t, "EUR", dir.Commodity.Symbol)
		assert.Equal(t, "1.000,00 EUR", dir.Format)
	})

	t.Run("subdirective with special characters in value", func(t *testing.T) {
		input := `account assets:bank
  note Account @ Bank & Trust (savings)`

		journal, errs := Parse(input)
		require.Empty(t, errs)
		require.Len(t, journal.Directives, 1)

		dir, ok := journal.Directives[0].(ast.AccountDirective)
		require.True(t, ok)
		assert.Equal(t, "Account @ Bank & Trust (savings)", dir.Subdirs["note"])
	})
}

func TestParser_DateEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		year      int
		month     int
		day       int
		expectErr bool
	}{
		{
			name: "month 13 parsed without validation",
			input: `2024-13-01 test
    e:f  $1
    a:c`,
			year: 2024, month: 13, day: 1,
			expectErr: false,
		},
		{
			name: "day 32 parsed without validation",
			input: `2024-01-32 test
    e:f  $1
    a:c`,
			year: 2024, month: 1, day: 32,
			expectErr: false,
		},
		{
			name: "february 30 parsed without validation",
			input: `2024-02-30 test
    e:f  $1
    a:c`,
			year: 2024, month: 2, day: 30,
			expectErr: false,
		},
		{
			name: "month 0 parsed without validation",
			input: `2024-00-15 test
    e:f  $1
    a:c`,
			year: 2024, month: 0, day: 15,
			expectErr: false,
		},
		{
			name: "day 0 parsed without validation",
			input: `2024-01-00 test
    e:f  $1
    a:c`,
			year: 2024, month: 1, day: 0,
			expectErr: false,
		},
		{
			name: "leap year feb 29 valid",
			input: `2024-02-29 test
    e:f  $1
    a:c`,
			year: 2024, month: 2, day: 29,
			expectErr: false,
		},
		{
			name: "large year",
			input: `99999-01-15 test
    e:f  $1
    a:c`,
			year: 99999, month: 1, day: 15,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal, errs := Parse(tt.input)
			if tt.expectErr {
				require.NotEmpty(t, errs)
				return
			}
			require.Empty(t, errs)
			require.Len(t, journal.Transactions, 1)
			assert.Equal(t, tt.year, journal.Transactions[0].Date.Year)
			assert.Equal(t, tt.month, journal.Transactions[0].Date.Month)
			assert.Equal(t, tt.day, journal.Transactions[0].Date.Day)
		})
	}
}

func Test_normalizeNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// No separators
		{name: "plain integer", input: "1234", expected: "1234"},
		{name: "plain decimal", input: "12.34", expected: "12.34"},

		// Single comma - decimal separator (European)
		{name: "comma as decimal separator", input: "1234,56", expected: "1234.56"},
		{name: "comma with 2 decimals", input: "12,34", expected: "12.34"},

		// Single comma - thousands separator (when followed by exactly 3 digits)
		{name: "comma as thousands with 4 digits before", input: "1,234", expected: "1234"},
		{name: "comma as thousands with more digits", input: "12,345", expected: "12345"},

		// Leading zeros edge case
		{name: "leading zeros comma decimal", input: "000,50", expected: "000.50"},
		{name: "zero comma three digits", input: "0,123", expected: "0.123"},

		// Dot and comma together - European format (1.234,56)
		{name: "european format", input: "1.234,56", expected: "1234.56"},
		{name: "european with multiple dots", input: "1.234.567,89", expected: "1234567.89"},

		// Dot and comma together - US format (1,234.56)
		{name: "us format", input: "1,234.56", expected: "1234.56"},
		{name: "us with multiple commas", input: "1,234,567.89", expected: "1234567.89"},

		// Multiple commas only (thousands separators)
		{name: "multiple commas no dot", input: "1,234,567", expected: "1234567"},

		// Multiple dots only (thousands separators, European)
		{name: "multiple dots no comma", input: "1.234.567", expected: "1234567"},

		// Dot as thousands separator (when followed by exactly 3 digits)
		{name: "dot as thousands", input: "1.234", expected: "1234"},

		// Edge cases - trailing separators
		{name: "trailing comma", input: "123,", expected: "123."},
		{name: "trailing dot", input: "123.", expected: "123."},

		// Edge cases - leading decimal
		{name: "leading dot", input: ".50", expected: ".50"},
		{name: "leading comma", input: ",50", expected: ".50"},

		// Scientific notation should pass through
		{name: "scientific notation", input: "1E+10", expected: "1E+10"},
		{name: "scientific lowercase", input: "1e-5", expected: "1e-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_PartialDateWithNoCommodityAmount(t *testing.T) {
	input := `Y2019

12/31 * Apple
    Расходы:Развлечения:Музыка       169
    Активы:Тинькофф:Текущий`

	journal, errs := Parse(input)
	require.Empty(t, errs, "parsing should succeed")
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	posting := tx.Postings[0]
	require.NotNil(t, posting.Amount, "amount should not be nil")
	assert.Equal(t, "169", posting.Amount.Quantity.String())
}

func TestParser_TabBetweenAccountAndAmount(t *testing.T) {
	input := "2024-01-15 test\n    expenses:food\t169\n    assets:cash"

	journal, errs := Parse(input)
	require.Empty(t, errs, "tab between account and amount should be valid")
	require.Len(t, journal.Transactions, 1)

	posting := journal.Transactions[0].Postings[0]
	require.NotNil(t, posting.Amount, "amount should be parsed")
	assert.Equal(t, "169", posting.Amount.Quantity.String())
}

func TestParser_MixedWhitespaceBetweenAccountAndAmount(t *testing.T) {
	input := "2024-01-15 test\n    expenses:food  \t  169\n    assets:cash"

	journal, errs := Parse(input)
	require.Empty(t, errs)

	posting := journal.Transactions[0].Postings[0]
	require.NotNil(t, posting.Amount)
	assert.Equal(t, "169", posting.Amount.Quantity.String())
}

func TestParser_ApplyAccount(t *testing.T) {
	input := `apply account business

2024-01-15 Sale
    revenue    $100
    checking

end apply account
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]

	// Original names should be preserved
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)

	// Resolved names should have the prefix
	assert.Equal(t, "business:revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "business:checking", tx.Postings[1].Account.ResolvedName)
}

func TestParser_NestedApplyAccount(t *testing.T) {
	input := `apply account business
apply account europe

2024-01-15 Sale
    revenue    $100
    checking

end apply account
end apply account
`
	journal, errs := Parse(input)
	require.Empty(t, errs)

	tx := journal.Transactions[0]

	// Original names preserved
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)

	// Resolved names with nested prefixes
	assert.Equal(t, "business:europe:revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "business:europe:checking", tx.Postings[1].Account.ResolvedName)
}

func TestParser_ApplyAccountNoEnd(t *testing.T) {
	input := `apply account personal

2024-01-15 Groceries
    expenses:food    $50
    checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs) // NOT an error!

	tx := journal.Transactions[0]

	// Original names preserved
	assert.Equal(t, "expenses:food", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)

	// Resolved names with prefix
	assert.Equal(t, "personal:expenses:food", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "personal:checking", tx.Postings[1].Account.ResolvedName)
}

func TestParser_ApplyAccountComplex(t *testing.T) {
	input := `; Transaction without apply account
2024-01-10 No prefix
    revenue    $50
    checking

apply account business

2024-01-15 With business prefix
    revenue    $100
    checking

apply account europe

2024-01-20 With business:europe prefix
    revenue    $200
    checking

end apply account

2024-01-25 Back to business prefix
    revenue    $150
    checking

end apply account

2024-01-30 No prefix again
    revenue    $75
    checking
`
	journal, errs := Parse(input)
	require.Empty(t, errs)
	require.Len(t, journal.Transactions, 5)

	// Transaction 1: No prefix (original = resolved)
	tx := journal.Transactions[0]
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)
	assert.Equal(t, "revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "checking", tx.Postings[1].Account.ResolvedName)

	// Transaction 2: business prefix
	tx = journal.Transactions[1]
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)
	assert.Equal(t, "business:revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "business:checking", tx.Postings[1].Account.ResolvedName)

	// Transaction 3: business:europe prefix
	tx = journal.Transactions[2]
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)
	assert.Equal(t, "business:europe:revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "business:europe:checking", tx.Postings[1].Account.ResolvedName)

	// Transaction 4: back to business prefix
	tx = journal.Transactions[3]
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)
	assert.Equal(t, "business:revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "business:checking", tx.Postings[1].Account.ResolvedName)

	// Transaction 5: no prefix again
	tx = journal.Transactions[4]
	assert.Equal(t, "revenue", tx.Postings[0].Account.Name)
	assert.Equal(t, "checking", tx.Postings[1].Account.Name)
	assert.Equal(t, "revenue", tx.Postings[0].Account.ResolvedName)
	assert.Equal(t, "checking", tx.Postings[1].Account.ResolvedName)
}

func TestParser_PrefixCommodityAfterBareNumber(t *testing.T) {
	input := "2024-01-15 test\n    Расходы:Продукты  698,43\n    Активы:Альфа  RUB100,00\n    Активы:Бета  RUB11,00"

	journal, errs := Parse(input)
	require.Empty(t, errs, "expected no parse errors, got: %v", errs)
	require.Len(t, journal.Transactions, 1)

	tx := journal.Transactions[0]
	require.Len(t, tx.Postings, 3)

	p1 := tx.Postings[0]
	require.NotNil(t, p1.Amount, "first posting amount should not be nil")
	assert.Equal(t, "698,43", p1.Amount.RawQuantity)

	p2 := tx.Postings[1]
	require.NotNil(t, p2.Amount, "second posting amount should not be nil")
	assert.Equal(t, "RUB", p2.Amount.Commodity.Symbol)
	assert.Equal(t, "100,00", p2.Amount.RawQuantity)

	p3 := tx.Postings[2]
	require.NotNil(t, p3.Amount, "third posting amount should not be nil")
	assert.Equal(t, "RUB", p3.Amount.Commodity.Symbol)
	assert.Equal(t, "11,00", p3.Amount.RawQuantity)
}
