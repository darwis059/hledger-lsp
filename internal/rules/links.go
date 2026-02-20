package rules

import (
	"path/filepath"
	"strings"

	"github.com/juev/hledger-lsp/internal/ast"
)

// DocumentLink represents a navigable link in a rules file.
type DocumentLink struct {
	Path  string
	Range ast.Range
}

// Links returns document links (include and source paths) in a parsed rules file.
//
// baseDir is the directory of the current file, used to resolve relative paths.
func Links(rf *RulesFile, baseDir string) []DocumentLink {
	if rf == nil {
		return []DocumentLink{}
	}
	links := []DocumentLink{}

	for _, inc := range rf.Includes {
		if inc.Path == "" {
			continue
		}
		pathRange := pathOnlyRange(inc.Range, len(inc.Path), len(inc.Path))
		links = append(links, DocumentLink{
			Path:  resolvePath(inc.Path, baseDir),
			Range: pathRange,
		})
	}

	for _, src := range rf.Sources {
		if src.Path == "" {
			continue
		}
		// Strip trailing "| command" for source paths
		path := strings.SplitN(src.Path, "|", 2)[0]
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		pathRange := pathOnlyRange(src.Range, len(src.Path), len(path))
		links = append(links, DocumentLink{
			Path:  resolvePath(path, baseDir),
			Range: pathRange,
		})
	}

	return links
}

// pathOnlyRange computes the range covering just the path portion of a directive.
// fullValueLen is the byte length of the stored value; pathLen is the byte length of the path portion.
func pathOnlyRange(directiveRange ast.Range, fullValueLen, pathLen int) ast.Range {
	end := directiveRange.End
	startCol := end.Column - fullValueLen
	startOffset := end.Offset - fullValueLen
	return ast.Range{
		Start: ast.Position{Line: end.Line, Column: startCol, Offset: startOffset},
		End:   ast.Position{Line: end.Line, Column: startCol + pathLen, Offset: startOffset + pathLen},
	}
}

func resolvePath(p, baseDir string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(baseDir, p))
}
