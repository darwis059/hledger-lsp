package include

import "github.com/juev/hledger-lsp/internal/ast"

type ErrorKind int

const (
	ErrorFileNotFound ErrorKind = iota
	ErrorCycleDetected
	ErrorParseError
	ErrorReadError
	ErrorFileTooLarge
	ErrorPathTraversal
	ErrorNotJournal
)

type LoadError struct {
	Kind    ErrorKind
	Path    string
	Message string
	Range   ast.Range
}

func (e LoadError) Error() string {
	return e.Message
}

type FileSource struct {
	Path    string
	Content string
}

type ResolvedJournal struct {
	Primary   *ast.Journal
	Files     map[string]*ast.Journal
	FileOrder []string
	Errors    []LoadError
}

func NewResolvedJournal(primary *ast.Journal) *ResolvedJournal {
	return &ResolvedJournal{
		Primary: primary,
		Files:   make(map[string]*ast.Journal),
	}
}

func (r *ResolvedJournal) AllTransactions() []ast.Transaction {
	var result []ast.Transaction
	if r.Primary != nil {
		result = append(result, r.Primary.Transactions...)
	}
	for _, path := range r.FileOrder {
		if j, ok := r.Files[path]; ok {
			result = append(result, j.Transactions...)
		}
	}
	return result
}

func (r *ResolvedJournal) AllDirectives() []ast.Directive {
	var result []ast.Directive
	if r.Primary != nil {
		result = append(result, r.Primary.Directives...)
	}
	for _, path := range r.FileOrder {
		if j, ok := r.Files[path]; ok {
			result = append(result, j.Directives...)
		}
	}
	return result
}

// FormatDirectives returns directives suitable for formatter commodity format extraction.
// It excludes DecimalMarkDirective from included files because decimal-mark is file-scoped
// per hledger semantics, while primary file directives are passed through unfiltered.
func (r *ResolvedJournal) FormatDirectives() []ast.Directive {
	var result []ast.Directive
	if r.Primary != nil {
		result = append(result, r.Primary.Directives...)
	}
	for _, path := range r.FileOrder {
		if j, ok := r.Files[path]; ok {
			for _, d := range j.Directives {
				if _, ok := d.(ast.DecimalMarkDirective); ok {
					continue
				}
				result = append(result, d)
			}
		}
	}
	return result
}

func (r *ResolvedJournal) AllIncludes() []ast.Include {
	var result []ast.Include
	if r.Primary != nil {
		result = append(result, r.Primary.Includes...)
	}
	for _, path := range r.FileOrder {
		if j, ok := r.Files[path]; ok {
			result = append(result, j.Includes...)
		}
	}
	return result
}
