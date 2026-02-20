package rules

import (
	"strings"

	"github.com/juev/hledger-lsp/internal/ast"
)

// Parse parses a rules file and returns its AST and any parse-time diagnostics.
func Parse(input string) (*RulesFile, []Diagnostic) {
	p := &parser{lexer: NewLexer(input)}
	p.advance()
	return p.parseFile()
}

type parser struct {
	lexer   *Lexer
	current Token
	rf      *RulesFile
	diags   []Diagnostic
}

func (p *parser) advance() {
	p.current = p.lexer.Next()
}

func (p *parser) parseFile() (*RulesFile, []Diagnostic) {
	rf := &RulesFile{}
	p.rf = rf

	for p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenNewline:
			p.advance()

		case TokenComment:
			rf.Comments = append(rf.Comments, p.parseComment())

		case TokenIfKeyword:
			rf.IfBlocks = append(rf.IfBlocks, p.parseIfBlock())

		case TokenEndKeyword:
			p.advance()

		case TokenDirective:
			p.parseDirective(rf)

		case TokenIndent:
			// Top-level indented line — field assignment outside if block
			p.parseIndentedAssignment(rf)

		case TokenFieldName:
			// Unindented field assignment
			a := p.parseFieldAssignmentFromCurrent()
			rf.Assignments = append(rf.Assignments, a)

		default:
			if p.current.Type == TokenText && p.current.Pos.Column == 1 {
				tok := p.current
				p.diags = append(p.diags, Diagnostic{
					Range:    tokenRange(tok),
					Message:  "unknown directive: " + firstWord(tok.Value),
					Code:     "UNKNOWN_DIRECTIVE",
					Severity: SeverityWarning,
				})
			}
			p.advance()
		}
	}

	return rf, p.diags
}

func (p *parser) parseComment() RulesComment {
	tok := p.current
	p.advance()
	return RulesComment{
		Text:  tok.Value,
		Range: tokenRange(tok),
	}
}

func (p *parser) parseDirective(rf *RulesFile) {
	dirTok := p.current
	p.advance()

	var valueTok *Token
	if p.current.Type == TokenText {
		t := p.current
		valueTok = &t
		p.advance()
	}

	name := dirTok.Value
	value := ""
	if valueTok != nil {
		value = valueTok.Value
	}

	rng := tokenRange(dirTok)
	if valueTok != nil {
		rng.End = ast.Position{Line: valueTok.End.Line, Column: valueTok.End.Column, Offset: valueTok.End.Offset}
	}

	switch name {
	case "fields":
		fd := FieldsDirective{Range: rng}
		for f := range strings.SplitSeq(value, ",") {
			fn := strings.TrimSpace(f)
			if fn != "" {
				fd.Names = append(fd.Names, fn)
			}
		}
		rf.FieldsDefs = append(rf.FieldsDefs, fd)

	case "include":
		rf.Includes = append(rf.Includes, IncludeDirective{
			Path:  value,
			Range: rng,
		})

	case "source":
		rf.Sources = append(rf.Sources, SourceDirective{
			Path:  value,
			Range: rng,
		})

	default:
		rf.Directives = append(rf.Directives, SimpleDirective{
			Name:  name,
			Value: value,
			Range: rng,
		})
	}
}

func (p *parser) parseIfBlock() IfBlock {
	startTok := p.current
	block := IfBlock{}

	// Collect inline pattern from TokenRegex on the "if" line
	p.advance() // consume "if"
	if p.current.Type == TokenRegex || p.current.Type == TokenText {
		block.Patterns = append(block.Patterns, p.current.Value)
		p.advance()
	}

	// Skip newline after "if PATTERN"
	if p.current.Type == TokenNewline {
		p.advance()
	}

	// Collect continuation patterns (&) and indented assignments
	var lastPatternEnd ast.Position
loop:
	for p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenIndent:
			a := p.parseIndentedAssignmentDirect()
			if a.Field != "" {
				block.Assignments = append(block.Assignments, a)
			}
		case TokenText:
			if strings.HasPrefix(p.current.Value, "& ") || p.current.Value == "&" {
				tok := p.current
				pat := strings.TrimPrefix(strings.TrimPrefix(p.current.Value, "& "), "&")
				pat = strings.TrimSpace(pat)
				if pat != "" {
					block.Patterns = append(block.Patterns, pat)
				} else {
					p.diags = append(p.diags, Diagnostic{
						Range:    tokenRange(tok),
						Message:  "empty continuation pattern",
						Code:     "EMPTY_PATTERN",
						Severity: SeverityWarning,
					})
				}
				lastPatternEnd = ast.Position{Line: tok.End.Line, Column: tok.End.Column, Offset: tok.End.Offset}
				p.advance()
				if p.current.Type == TokenNewline {
					p.advance()
				}
			} else if len(block.Assignments) == 0 && p.current.Pos.Column == 1 {
				tok := p.current
				block.Patterns = append(block.Patterns, tok.Value)
				lastPatternEnd = ast.Position{
					Line: tok.End.Line, Column: tok.End.Column, Offset: tok.End.Offset,
				}
				p.advance()
				if p.current.Type == TokenNewline {
					p.advance()
				}
			} else {
				break loop
			}
		case TokenNewline:
			p.advance()
		case TokenEndKeyword, TokenDirective, TokenIfKeyword, TokenComment:
			break loop
		default:
			break loop
		}
	}

	blockEnd := ast.Position{Line: startTok.Pos.Line, Column: startTok.End.Column, Offset: startTok.End.Offset}
	if len(block.Assignments) > 0 {
		blockEnd = block.Assignments[len(block.Assignments)-1].Range.End
	} else if lastPatternEnd.Line > 0 {
		blockEnd = lastPatternEnd
	}
	block.Range = ast.Range{
		Start: posFromToken(startTok),
		End:   blockEnd,
	}
	return block
}

// parseIndentedAssignment handles a top-level indented line as a field assignment
// outside of any if block (adds to rf.Assignments).
func (p *parser) parseIndentedAssignment(rf *RulesFile) {
	a := p.parseIndentedAssignmentDirect()
	if a.Field != "" {
		rf.Assignments = append(rf.Assignments, a)
	}
}

// parseIndentedAssignmentDirect parses an indented field assignment and returns it.
func (p *parser) parseIndentedAssignmentDirect() FieldAssignment {
	p.advance() // consume TokenIndent

	if p.current.Type == TokenFieldName {
		return p.parseFieldAssignmentFromCurrent()
	}

	// skip non-field indented line; multi-word text signals an unknown field name
	if p.current.Type == TokenText {
		tok := p.current
		if firstWord(tok.Value) != tok.Value {
			p.diags = append(p.diags, Diagnostic{
				Range:    tokenRange(tok),
				Message:  "unknown field name: " + firstWord(tok.Value),
				Code:     "UNKNOWN_FIELD",
				Severity: SeverityWarning,
			})
		}
		p.advance()
	}
	if p.current.Type == TokenNewline {
		p.advance()
	}
	return FieldAssignment{}
}

func (p *parser) parseFieldAssignmentFromCurrent() FieldAssignment {
	fieldTok := p.current
	p.advance()

	a := FieldAssignment{
		Field: fieldTok.Value,
		Range: tokenRange(fieldTok),
	}

	if p.current.Type == TokenText {
		a.Value = p.current.Value
		a.Range.End = ast.Position{Line: p.current.End.Line, Column: p.current.End.Column, Offset: p.current.End.Offset}
		p.advance()
	}

	if p.current.Type == TokenNewline {
		p.advance()
	}

	return a
}

func tokenRange(tok Token) ast.Range {
	return ast.Range{
		Start: posFromToken(tok),
		End:   ast.Position{Line: tok.End.Line, Column: tok.End.Column, Offset: tok.End.Offset},
	}
}

func posFromToken(tok Token) ast.Position {
	return ast.Position{Line: tok.Pos.Line, Column: tok.Pos.Column, Offset: tok.Pos.Offset}
}
