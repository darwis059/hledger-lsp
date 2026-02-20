// Package filetype detects hledger file types from URI extensions.
package filetype

import "strings"

type FileType int

const (
	Unknown FileType = iota
	Journal
	Rules
)

func (f FileType) String() string {
	switch f {
	case Journal:
		return "journal"
	case Rules:
		return "rules"
	default:
		return "unknown"
	}
}

// Detect returns the FileType for the given URI based on its file extension.
// The uri parameter is a file path or LSP file URI (file://...); query parameters are not present in LSP URIs.
func Detect(uri string) FileType {
	if uri == "" {
		return Unknown
	}
	lower := strings.ToLower(uri)
	switch {
	case strings.HasSuffix(lower, ".journal"),
		strings.HasSuffix(lower, ".hledger"),
		strings.HasSuffix(lower, ".j"),
		strings.HasSuffix(lower, ".ledger"):
		return Journal
	case strings.HasSuffix(lower, ".rules"):
		return Rules
	default:
		return Unknown
	}
}

// IsJournal reports whether the URI refers to an hledger journal file.
func IsJournal(uri string) bool { return Detect(uri) == Journal }

// IsRules reports whether the URI refers to an hledger rules file.
func IsRules(uri string) bool { return Detect(uri) == Rules }
