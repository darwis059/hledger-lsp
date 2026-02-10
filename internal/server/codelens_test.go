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

func TestCodeLens_OneLensPerTransaction(t *testing.T) {
	srv := NewServer()
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
