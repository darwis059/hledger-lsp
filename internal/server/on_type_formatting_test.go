package server

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
)

func (ts *testServer) onTypeFormatting(uri protocol.DocumentURI, line uint32, ch string) ([]protocol.TextEdit, error) {
	params := &protocol.DocumentOnTypeFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		Position:     protocol.Position{Line: line, Character: 0},
		Ch:           ch,
	}
	return ts.OnTypeFormatting(context.Background(), params)
}

func TestOnTypeFormatting_AfterTransactionHeader(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "2024-01-15 grocery store\n"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 1, "\n")
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "    ", edits[0].NewText)
	assert.Equal(t, uint32(1), edits[0].Range.Start.Line)
	assert.Equal(t, uint32(0), edits[0].Range.Start.Character)
}

func TestOnTypeFormatting_AfterPosting(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "2024-01-15 grocery store\n    expenses:food  $50.00\n"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 2, "\n")
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "    ", edits[0].NewText)
}

func TestOnTypeFormatting_AfterEmptyLine(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "2024-01-15 grocery store\n    expenses:food  $50.00\n    assets:cash\n\n"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 4, "\n")
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "", edits[0].NewText)
}

func TestOnTypeFormatting_AfterWhitespaceOnlyLine(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "2024-01-15 grocery store\n    expenses:food  $50.00\n    assets:cash\n    \n"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 4, "\n")
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "", edits[0].NewText)
}

func TestOnTypeFormatting_FirstLine(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "\n2024-01-15 test"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 0, "\n")
	require.NoError(t, err)
	assert.Nil(t, edits)
}

func TestOnTypeFormatting_AfterDirective(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "account expenses:food\n"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 1, "\n")
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "    ", edits[0].NewText)
}

func TestOnTypeFormatting_CustomIndentSize(t *testing.T) {
	tests := []struct {
		name       string
		indentSize int
		expected   string
	}{
		{"indent 2", 2, "  "},
		{"indent 8", 8, strings.Repeat(" ", 8)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestServer()
			settings := ts.getSettings()
			settings.Formatting.IndentSize = tt.indentSize
			ts.setSettings(settings)

			uri := protocol.DocumentURI("file:///test.journal")
			content := "2024-01-15 grocery store\n"

			ts.StoreDocument(uri, content)

			edits, err := ts.onTypeFormatting(uri, 1, "\n")
			require.NoError(t, err)
			require.Len(t, edits, 1)
			assert.Equal(t, tt.expected, edits[0].NewText)
		})
	}
}

func TestOnTypeFormatting_NonNewlineTrigger(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "2024-01-15 grocery store\n"

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 1, "a")
	require.NoError(t, err)
	assert.Nil(t, edits)
}

func TestOnTypeFormatting_DocumentNotFound(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///nonexistent.journal")

	edits, err := ts.onTypeFormatting(uri, 1, "\n")
	require.NoError(t, err)
	assert.Nil(t, edits)
}

func TestOnTypeFormatting_ReplacesEditorAutoIndent(t *testing.T) {
	ts := newTestServer()
	uri := protocol.DocumentURI("file:///test.journal")
	content := "2024-01-15 grocery store\n        "

	ts.StoreDocument(uri, content)

	edits, err := ts.onTypeFormatting(uri, 1, "\n")
	require.NoError(t, err)
	require.Len(t, edits, 1)
	assert.Equal(t, "    ", edits[0].NewText)
	assert.Equal(t, uint32(0), edits[0].Range.Start.Character)
	assert.Equal(t, uint32(8), edits[0].Range.End.Character)
}
