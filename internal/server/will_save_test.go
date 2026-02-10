package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
)

func TestWillSaveWaitUntil_ReturnsEdits(t *testing.T) {
	srv := NewServer()
	content := "2024-01-15 grocery\n    expenses:food  $50\n    assets:cash   $-50"
	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.WillSaveTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		Reason:       protocol.TextDocumentSaveReasonManual,
	}

	edits, err := srv.WillSaveWaitUntil(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, edits)
}

func TestWillSaveWaitUntil_FormattingDisabled(t *testing.T) {
	srv := NewServer()
	settings := srv.getSettings()
	settings.Features.Formatting = false
	srv.setSettings(settings)

	content := "2024-01-15 grocery\n    expenses:food  $50\n    assets:cash"
	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.WillSaveTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		Reason:       protocol.TextDocumentSaveReasonManual,
	}

	edits, err := srv.WillSaveWaitUntil(context.Background(), params)
	require.NoError(t, err)
	assert.Nil(t, edits)
}

func TestWillSaveWaitUntil_UnknownDocument(t *testing.T) {
	srv := NewServer()

	params := &protocol.WillSaveTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///nonexistent.journal"},
		Reason:       protocol.TextDocumentSaveReasonManual,
	}

	edits, err := srv.WillSaveWaitUntil(context.Background(), params)
	require.NoError(t, err)
	assert.Nil(t, edits)
}
