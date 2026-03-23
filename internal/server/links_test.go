package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestDocumentLink_IncludeDirective(t *testing.T) {
	srv := NewServer()
	content := `include accounts.journal
include data/transactions.journal

2024-01-15 test
    expenses:food  $50
    assets:cash`

	docURI := protocol.DocumentURI("file:///home/user/main.journal")
	srv.documents.Store(docURI, content)

	params := &protocol.DocumentLinkParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	}

	result, err := srv.DocumentLink(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, uint32(0), result[0].Range.Start.Line)
	assert.Equal(t, string(uri.File("/home/user/accounts.journal")), string(result[0].Target))

	assert.Equal(t, uint32(1), result[1].Range.Start.Line)
	assert.Equal(t, string(uri.File("/home/user/data/transactions.journal")), string(result[1].Target))
}

func TestDocumentLink_NoIncludes(t *testing.T) {
	srv := NewServer()
	content := `2024-01-15 test
    expenses:food  $50
    assets:cash`

	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, content)

	params := &protocol.DocumentLinkParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.DocumentLink(context.Background(), params)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDocumentLink_EmptyDocument(t *testing.T) {
	srv := NewServer()
	uri := protocol.DocumentURI("file:///test.journal")
	srv.documents.Store(uri, "")

	params := &protocol.DocumentLinkParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	}

	result, err := srv.DocumentLink(context.Background(), params)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDocumentLink_DocumentNotFound(t *testing.T) {
	srv := NewServer()

	params := &protocol.DocumentLinkParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///nonexistent.journal"},
	}

	result, err := srv.DocumentLink(context.Background(), params)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestDocumentLink_RulesIncludeWithSpaces(t *testing.T) {
	srv := NewServer()
	content := "include /path/with spaces/other.rules\n"

	rulesURI := protocol.DocumentURI("file:///home/user/test.rules")
	srv.documents.Store(rulesURI, content)

	params := &protocol.DocumentLinkParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: rulesURI},
	}

	result, err := srv.DocumentLink(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, result, 1)

	expected := string(uri.File("/path/with spaces/other.rules"))
	assert.Equal(t, expected, string(result[0].Target))
}
