package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/juev/hledger-lsp/internal/ast"
	"github.com/juev/hledger-lsp/internal/include"
)

func TestFindCurrentTransactionIndex_CursorInside(t *testing.T) {
	// AST uses 1-based lines, LSP cursor uses 0-based
	transactions := []ast.Transaction{
		{Range: ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}}},
		{Range: ast.Range{Start: ast.Position{Line: 5}, End: ast.Position{Line: 7}}},
		{Range: ast.Range{Start: ast.Position{Line: 9}, End: ast.Position{Line: 11}}},
	}

	assert.Equal(t, 0, findCurrentTransactionIndex(transactions, 0))
	assert.Equal(t, 0, findCurrentTransactionIndex(transactions, 1))
	assert.Equal(t, 0, findCurrentTransactionIndex(transactions, 2))
	assert.Equal(t, 1, findCurrentTransactionIndex(transactions, 4))
	assert.Equal(t, 1, findCurrentTransactionIndex(transactions, 5))
	assert.Equal(t, 2, findCurrentTransactionIndex(transactions, 9))
}

func TestFindCurrentTransactionIndex_CursorBetween(t *testing.T) {
	transactions := []ast.Transaction{
		{Range: ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}}},
		{Range: ast.Range{Start: ast.Position{Line: 5}, End: ast.Position{Line: 7}}},
	}

	assert.Equal(t, -1, findCurrentTransactionIndex(transactions, 3))
}

func TestFindCurrentTransactionIndex_CursorAfterAll(t *testing.T) {
	transactions := []ast.Transaction{
		{Range: ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}}},
	}

	assert.Equal(t, -1, findCurrentTransactionIndex(transactions, 5))
}

func TestFindCurrentTransactionIndex_Empty(t *testing.T) {
	assert.Equal(t, -1, findCurrentTransactionIndex(nil, 0))
	assert.Equal(t, -1, findCurrentTransactionIndex([]ast.Transaction{}, 0))
}

func TestJournalWithoutTransaction_RemovesCorrectIndex(t *testing.T) {
	journal := &ast.Journal{
		Transactions: []ast.Transaction{
			{Description: "first"},
			{Description: "second"},
			{Description: "third"},
		},
		Directives: []ast.Directive{
			ast.AccountDirective{Account: ast.Account{Name: "assets:cash"}},
		},
	}

	result := journalWithoutTransaction(journal, 1)

	assert.Len(t, result.Transactions, 2)
	assert.Equal(t, "first", result.Transactions[0].Description)
	assert.Equal(t, "third", result.Transactions[1].Description)
	assert.Len(t, result.Directives, 1)
}

func TestJournalWithoutTransaction_NegativeIndex(t *testing.T) {
	journal := &ast.Journal{
		Transactions: []ast.Transaction{
			{Description: "first"},
		},
	}

	result := journalWithoutTransaction(journal, -1)
	assert.Same(t, journal, result)
}

func TestJournalWithoutTransaction_DoesNotMutateOriginal(t *testing.T) {
	journal := &ast.Journal{
		Transactions: []ast.Transaction{
			{Description: "first"},
			{Description: "second"},
		},
	}

	result := journalWithoutTransaction(journal, 0)

	assert.Len(t, journal.Transactions, 2, "original unchanged")
	assert.Len(t, result.Transactions, 1)
	assert.Equal(t, "second", result.Transactions[0].Description)
}

func TestJournalWithoutTransaction_SingleTransaction(t *testing.T) {
	journal := &ast.Journal{
		Transactions: []ast.Transaction{
			{Description: "only"},
		},
	}

	result := journalWithoutTransaction(journal, 0)
	assert.Empty(t, result.Transactions)
}

func TestResolvedWithoutTransaction_FiltersPrimary(t *testing.T) {
	primary := &ast.Journal{
		Transactions: []ast.Transaction{
			{
				Description: "first",
				Range:       ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}},
			},
			{
				Description: "second",
				Range:       ast.Range{Start: ast.Position{Line: 5}, End: ast.Position{Line: 7}},
			},
		},
	}
	resolved := &include.ResolvedJournal{
		Primary:   primary,
		Files:     map[string]*ast.Journal{"included.journal": {Transactions: []ast.Transaction{{Description: "inc"}}}},
		FileOrder: []string{"included.journal"},
	}

	result := resolvedWithoutTransaction(resolved, 5, "file:///main.journal")

	assert.Len(t, result.Primary.Transactions, 1)
	assert.Equal(t, "first", result.Primary.Transactions[0].Description)
	assert.Equal(t, resolved.Files, result.Files)
	assert.Equal(t, resolved.FileOrder, result.FileOrder)
}

func TestResolvedWithoutTransaction_CursorNotInTransaction(t *testing.T) {
	primary := &ast.Journal{
		Transactions: []ast.Transaction{
			{Range: ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}}},
		},
	}
	resolved := &include.ResolvedJournal{Primary: primary}

	result := resolvedWithoutTransaction(resolved, 10, "file:///main.journal")

	assert.Same(t, resolved, result)
}

func TestResolvedWithoutTransaction_FiltersIncludedFile(t *testing.T) {
	primary := &ast.Journal{
		Transactions: []ast.Transaction{
			{Description: "main-tx", Range: ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}}},
		},
	}
	includedJournal := &ast.Journal{
		Transactions: []ast.Transaction{
			{Description: "inc-first", Range: ast.Range{Start: ast.Position{Line: 1}, End: ast.Position{Line: 3}}},
			{Description: "inc-second", Range: ast.Range{Start: ast.Position{Line: 5}, End: ast.Position{Line: 5}}},
		},
	}
	resolved := &include.ResolvedJournal{
		Primary:   primary,
		Files:     map[string]*ast.Journal{"/tmp/included.journal": includedJournal},
		FileOrder: []string{"/tmp/included.journal"},
	}

	result := resolvedWithoutTransaction(resolved, 0, "file:///tmp/included.journal")

	assert.Same(t, resolved.Primary, result.Primary)
	assert.Len(t, result.Files["/tmp/included.journal"].Transactions, 1)
	assert.Equal(t, "inc-second", result.Files["/tmp/included.journal"].Transactions[0].Description)
}
