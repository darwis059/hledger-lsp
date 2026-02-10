package server

import (
	"context"
	"strings"

	"go.lsp.dev/protocol"

	"github.com/juev/hledger-lsp/internal/lsputil"
)

func (s *Server) OnTypeFormatting(_ context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	if params.Ch != "\n" {
		return nil, nil
	}

	doc, ok := s.GetDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	line := int(params.Position.Line)
	if line == 0 {
		return nil, nil
	}

	lines := splitLines(doc)
	prevLine := lines[line-1]
	trimmed := strings.TrimSpace(prevLine)

	settings := s.getSettings()
	indent := strings.Repeat(" ", settings.Formatting.IndentSize)

	var newIndent string
	if trimmed == "" {
		newIndent = ""
	} else {
		newIndent = indent
	}

	var currentLineLen uint32
	if line < len(lines) {
		currentLineContent := lines[line]
		if strings.TrimSpace(currentLineContent) == "" {
			currentLineLen = uint32(lsputil.UTF16Len(currentLineContent))
		}
	}

	return []protocol.TextEdit{{
		Range: protocol.Range{
			Start: protocol.Position{Line: uint32(line), Character: 0},
			End:   protocol.Position{Line: uint32(line), Character: currentLineLen},
		},
		NewText: newIndent,
	}}, nil
}
