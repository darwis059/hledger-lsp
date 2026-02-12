package server

import (
	"maps"

	"go.lsp.dev/protocol"

	"github.com/juev/hledger-lsp/internal/ast"
	"github.com/juev/hledger-lsp/internal/include"
)

func findCurrentTransactionIndex(transactions []ast.Transaction, lspLine int) int {
	astLine := lspLine + 1
	for i, tx := range transactions {
		if tx.Range.Start.Line <= astLine && astLine <= tx.Range.End.Line {
			return i
		}
	}
	return -1
}

func journalWithoutTransaction(journal *ast.Journal, txIndex int) *ast.Journal {
	if txIndex < 0 {
		return journal
	}

	filtered := make([]ast.Transaction, 0, len(journal.Transactions)-1)
	filtered = append(filtered, journal.Transactions[:txIndex]...)
	filtered = append(filtered, journal.Transactions[txIndex+1:]...)

	return &ast.Journal{
		Transactions:         filtered,
		PeriodicTransactions: journal.PeriodicTransactions,
		AutoPostingRules:     journal.AutoPostingRules,
		Directives:           journal.Directives,
		Comments:             journal.Comments,
		Includes:             journal.Includes,
	}
}

func resolvedWithoutTransaction(resolved *include.ResolvedJournal, cursorLine int, docURI protocol.DocumentURI) *include.ResolvedJournal {
	docPath := uriToPath(docURI)

	for path, journal := range resolved.Files {
		if path == docPath {
			txIdx := findCurrentTransactionIndex(journal.Transactions, cursorLine)
			if txIdx < 0 {
				return resolved
			}
			filtered := journalWithoutTransaction(journal, txIdx)
			newFiles := make(map[string]*ast.Journal, len(resolved.Files))
			maps.Copy(newFiles, resolved.Files)
			newFiles[path] = filtered
			return &include.ResolvedJournal{
				Primary:   resolved.Primary,
				Files:     newFiles,
				FileOrder: resolved.FileOrder,
				Errors:    resolved.Errors,
			}
		}
	}

	txIdx := findCurrentTransactionIndex(resolved.Primary.Transactions, cursorLine)
	if txIdx < 0 {
		return resolved
	}

	filteredPrimary := journalWithoutTransaction(resolved.Primary, txIdx)

	return &include.ResolvedJournal{
		Primary:   filteredPrimary,
		Files:     resolved.Files,
		FileOrder: resolved.FileOrder,
		Errors:    resolved.Errors,
	}
}
