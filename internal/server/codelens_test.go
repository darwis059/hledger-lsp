package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
)

func TestCodeLens_BalancedTransaction(t *testing.T) {
	srv := NewServer()
	settings := srv.getSettings()
	settings.Features.CodeLens = true
	srv.setSettings(settings)

	content := `2024-01-15 grocery
    expenses:food  $50
    assets:cash  $-50`

	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.CodeLens(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Contains(t, result[0].Command.Title, "balanced")
}

func TestCodeLens_UnbalancedTransaction(t *testing.T) {
	srv := NewServer()
	settings := srv.getSettings()
	settings.Features.CodeLens = true
	srv.setSettings(settings)

	content := `2024-01-15 grocery
    expenses:food  $50
    assets:cash  $-40`

	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.CodeLens(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Contains(t, result[0].Command.Title, "unbalanced")
}

func TestCodeLens_EmptyDocument(t *testing.T) {
	srv := NewServer()
	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, "")

	params := &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.CodeLens(context.Background(), params)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestCodeLens_FeatureDisabled(t *testing.T) {
	srv := NewServer()
	settings := srv.getSettings()
	settings.Features.CodeLens = false
	srv.setSettings(settings)

	content := `2024-01-15 grocery
    expenses:food  $50
    assets:cash  $-50`

	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.CodeLens(context.Background(), params)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCodeLens_CommodityDirectivePrecision(t *testing.T) {
	srv := NewServer()
	settings := srv.getSettings()
	settings.Features.CodeLens = true
	srv.setSettings(settings)

	// 3 * 33.337 = 100.011; diff = 0.011
	// Without directive: posting precision 0 → tolerance 0.5 → balanced (wrong)
	// With directive precision 2: tolerance 0.005, |0.011| > 0.005 → unbalanced (correct)
	content := `commodity $1,000.00

2024-01-15 buy
    assets:stock  3 AAPL @ $33.337
    assets:cash  -$100`

	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.CodeLens(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Contains(t, result[0].Command.Title, "unbalanced",
		"CodeLens must use commodity directive precision for balance check")
}

func TestCodeLens_OneLensPerTransaction(t *testing.T) {
	srv := NewServer()
	settings := srv.getSettings()
	settings.Features.CodeLens = true
	srv.setSettings(settings)

	content := `2024-01-15 grocery
    expenses:food  $50
    assets:cash  $-50

2024-01-16 store
    expenses:clothing  $30
    assets:cash`

	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.CodeLens(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, result, 2)
}
