package server

import (
	"context"
	"strings"

	"go.lsp.dev/protocol"

	"github.com/juev/hledger-lsp/internal/formatter"
	"github.com/juev/hledger-lsp/internal/lsputil"
	"github.com/juev/hledger-lsp/internal/parser"
)

type lineKind int

const (
	lineEmpty lineKind = iota
	lineWhitespaceOnly
	lineTransactionHeader
	linePosting
	lineDirective
	lineComment
	lineOther
)

var directiveKeywords = map[string]struct{}{
	"account": {}, "alias": {}, "apply": {}, "assert": {}, "bucket": {}, "capture": {},
	"check": {}, "comment": {}, "commodity": {}, "D": {}, "decimal-mark": {}, "def": {},
	"define": {}, "end": {}, "eval": {}, "expr": {}, "include": {}, "payee": {}, "P": {},
	"tag": {}, "test": {}, "Y": {}, "year": {},
}

func classifyLine(line string) lineKind {
	if len(line) == 0 {
		return lineEmpty
	}

	if strings.TrimSpace(line) == "" {
		return lineWhitespaceOnly
	}

	first := line[0]

	if first == ' ' || first == '\t' {
		return linePosting
	}

	if first == ';' || first == '#' || first == '*' {
		return lineComment
	}

	if first >= '0' && first <= '9' || first == '~' || first == '=' {
		return lineTransactionHeader
	}

	word := line
	if idx := strings.IndexAny(line, " \t"); idx != -1 {
		word = line[:idx]
	}
	if _, ok := directiveKeywords[word]; ok {
		return lineDirective
	}

	return lineOther
}

func (s *Server) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	doc, ok := s.GetDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	switch params.Ch {
	case "\n":
		return s.onTypeNewline(doc, params)
	case "\t":
		return s.onTypeTab(doc, params)
	default:
		return nil, nil
	}
}

func (s *Server) onTypeNewline(doc string, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	line := int(params.Position.Line)
	if line == 0 {
		return nil, nil
	}

	lines := splitLines(doc)
	if line-1 >= len(lines) {
		return nil, nil
	}
	prevLine := lines[line-1]

	settings := s.getSettings()
	indent := strings.Repeat(" ", settings.Formatting.IndentSize)

	kind := classifyLine(prevLine)
	var newIndent string
	switch kind {
	case lineTransactionHeader, linePosting:
		newIndent = indent
	default:
		newIndent = ""
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

func (s *Server) onTypeTab(doc string, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	line := int(params.Position.Line)
	lines := splitLines(doc)

	if line >= len(lines) {
		return nil, nil
	}

	if classifyLine(lines[line]) != linePosting {
		return nil, nil
	}

	alignCol := s.getAlignmentColumn(doc, params.TextDocument.URI)
	if alignCol <= 0 {
		return nil, nil
	}

	cursorChar := int(params.Position.Character)
	tabPos := cursorChar - 1
	if tabPos < 0 {
		return nil, nil
	}

	if tabPos >= alignCol {
		return nil, nil
	}

	spacesNeeded := alignCol - tabPos

	return []protocol.TextEdit{{
		Range: protocol.Range{
			Start: protocol.Position{Line: uint32(line), Character: uint32(tabPos)},
			End:   protocol.Position{Line: uint32(line), Character: uint32(cursorChar)},
		},
		NewText: strings.Repeat(" ", spacesNeeded),
	}}, nil
}

func (s *Server) getAlignmentColumn(doc string, uri protocol.DocumentURI) int {
	if cached, ok := s.alignmentCache.Load(uri); ok {
		return cached.(int)
	}

	journal, _ := parser.Parse(doc)
	if len(journal.Transactions) == 0 {
		return 0
	}

	settings := s.getSettings()
	alignCol := formatter.CalculateGlobalAlignmentColumnWithIndent(journal.Transactions, settings.Formatting.IndentSize)
	if settings.Formatting.MinAlignmentColumn > 0 && alignCol < settings.Formatting.MinAlignmentColumn-1 {
		alignCol = settings.Formatting.MinAlignmentColumn - 1
	}

	s.alignmentCache.Store(uri, alignCol)
	return alignCol
}
