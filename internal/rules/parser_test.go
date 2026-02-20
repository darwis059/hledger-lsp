package rules

import (
	"testing"
)

func TestParse_EmptyInput(t *testing.T) {
	rf, _ := Parse("")
	if rf == nil {
		t.Fatal("expected non-nil RulesFile")
	}
}

func TestParse_Comments(t *testing.T) {
	input := "# top comment\n; another comment\n* star comment"
	rf, _ := Parse(input)
	if len(rf.Comments) != 3 {
		t.Errorf("comments: got %d, want 3", len(rf.Comments))
	}
}

func TestParse_SkipDirective(t *testing.T) {
	input := "skip 1"
	rf, _ := Parse(input)
	if len(rf.Directives) != 1 {
		t.Fatalf("directives: got %d, want 1", len(rf.Directives))
	}
	d := rf.Directives[0]
	if d.Name != "skip" {
		t.Errorf("name: got %q, want skip", d.Name)
	}
	if d.Value != "1" {
		t.Errorf("value: got %q, want 1", d.Value)
	}
}

func TestParse_FlagDirective(t *testing.T) {
	input := "newest-first"
	rf, _ := Parse(input)
	if len(rf.Directives) != 1 {
		t.Fatalf("directives: got %d, want 1", len(rf.Directives))
	}
	d := rf.Directives[0]
	if d.Name != "newest-first" {
		t.Errorf("name: got %q, want newest-first", d.Name)
	}
	if d.Value != "" {
		t.Errorf("value: got %q, want empty", d.Value)
	}
}

func TestParse_FieldsDirective(t *testing.T) {
	input := "fields date, description, amount"
	rf, _ := Parse(input)
	if len(rf.FieldsDefs) != 1 {
		t.Fatalf("fieldsDefs: got %d, want 1", len(rf.FieldsDefs))
	}
	fd := rf.FieldsDefs[0]
	if len(fd.Names) != 3 {
		t.Errorf("field names count: got %d, want 3", len(fd.Names))
	}
	if fd.Names[0] != "date" {
		t.Errorf("name[0]: got %q, want date", fd.Names[0])
	}
	if fd.Names[2] != "amount" {
		t.Errorf("name[2]: got %q, want amount", fd.Names[2])
	}
}

func TestParse_IncludeDirective(t *testing.T) {
	input := "include common.rules"
	rf, _ := Parse(input)
	if len(rf.Includes) != 1 {
		t.Fatalf("includes: got %d, want 1", len(rf.Includes))
	}
	inc := rf.Includes[0]
	if inc.Path != "common.rules" {
		t.Errorf("path: got %q, want common.rules", inc.Path)
	}
}

func TestParse_SourceDirective(t *testing.T) {
	input := "source data.csv"
	rf, _ := Parse(input)
	if len(rf.Sources) != 1 {
		t.Fatalf("sources: got %d, want 1", len(rf.Sources))
	}
	src := rf.Sources[0]
	if src.Path != "data.csv" {
		t.Errorf("path: got %q, want data.csv", src.Path)
	}
}

func TestParse_TopLevelFieldAssignment(t *testing.T) {
	input := "account1 expenses:food"
	rf, _ := Parse(input)
	if len(rf.Assignments) != 1 {
		t.Fatalf("assignments: got %d, want 1", len(rf.Assignments))
	}
	a := rf.Assignments[0]
	if a.Field != "account1" {
		t.Errorf("field: got %q, want account1", a.Field)
	}
	if a.Value != "expenses:food" {
		t.Errorf("value: got %q, want expenses:food", a.Value)
	}
}

func TestParse_IfBlock(t *testing.T) {
	input := "if PATTERN\n  account1 expenses:food\n  description groceries"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 1 || block.Patterns[0] != "PATTERN" {
		t.Errorf("patterns: got %v, want [PATTERN]", block.Patterns)
	}
	if len(block.Assignments) != 2 {
		t.Errorf("assignments: got %d, want 2", len(block.Assignments))
	}
	if block.Assignments[0].Field != "account1" {
		t.Errorf("field[0]: got %q, want account1", block.Assignments[0].Field)
	}
}

func TestParse_IfBlockMultiPattern(t *testing.T) {
	input := "if PATTERN1\n& PATTERN2\n  account1 expenses:food"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 2 {
		t.Errorf("patterns: got %v, want 2 patterns", block.Patterns)
	}
}

func TestParse_MultipleDirectives(t *testing.T) {
	input := "skip 1\nseparator ,\ndate-format %Y-%m-%d"
	rf, _ := Parse(input)
	if len(rf.Directives) != 3 {
		t.Errorf("directives: got %d, want 3", len(rf.Directives))
	}
}

func TestParse_Range(t *testing.T) {
	input := "skip 1"
	rf, _ := Parse(input)
	d := rf.Directives[0]
	if d.Range.Start.Line != 1 {
		t.Errorf("start line: got %d, want 1", d.Range.Start.Line)
	}
}

func TestParse_Errors_UnknownDirective(t *testing.T) {
	input := "bogus-directive value"
	rf, _ := Parse(input)
	if len(rf.Directives) != 0 {
		t.Errorf("expected no directives for unknown text, got %d", len(rf.Directives))
	}
	if len(rf.Assignments) != 0 {
		t.Errorf("expected no assignments, got %d", len(rf.Assignments))
	}
	if len(rf.IfBlocks) != 0 {
		t.Errorf("expected no if blocks, got %d", len(rf.IfBlocks))
	}
}

func TestParse_IfBlockNoAssignments_RangeEndNonZero(t *testing.T) {
	input := "if PATTERN"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("expected 1 if block, got %d", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if block.Range.End.Column == 0 {
		t.Error("expected non-zero Column in IfBlock.Range.End when no assignments")
	}
	if block.Range.End.Offset == 0 {
		t.Error("expected non-zero Offset in IfBlock.Range.End when no assignments")
	}
}

func TestParse_IfBlockRangeEndComplete(t *testing.T) {
	input := "if PATTERN\n  account1 expenses:food"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("expected 1 if block, got %d", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if block.Range.End.Column == 0 {
		t.Error("expected non-zero Column in IfBlock.Range.End")
	}
	if block.Range.End.Offset == 0 {
		t.Error("expected non-zero Offset in IfBlock.Range.End")
	}
	if block.Range.End.Line != 2 {
		t.Errorf("expected End.Line=2, got %d", block.Range.End.Line)
	}
}

func TestParse_IfBlockContinuationOnlyRangeEnd(t *testing.T) {
	input := "if PAT\n& PAT2"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("expected 1 if block, got %d", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if block.Range.End.Line != 2 {
		t.Errorf("expected End.Line=2 for continuation-only if block, got %d", block.Range.End.Line)
	}
}

func TestParse_IfBlockORPatterns(t *testing.T) {
	input := "if\nPATTERN1\nPATTERN2\n  account1 expenses:food"
	rf, diags := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 2 {
		t.Errorf("patterns: got %v, want 2", block.Patterns)
	}
	if len(block.Assignments) != 1 {
		t.Errorf("assignments: got %d, want 1", len(block.Assignments))
	}
	for _, d := range diags {
		if d.Code == "UNKNOWN_DIRECTIVE" {
			t.Errorf("unexpected UNKNOWN_DIRECTIVE diagnostic: %s", d.Message)
		}
	}
}

func TestParse_IfBlockFieldMatchers(t *testing.T) {
	input := "if\n%description GOOGLE PLAY\n%comment Google Play\n  account2 Expenses:Services"
	rf, diags := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 2 {
		t.Errorf("patterns: got %v, want 2", block.Patterns)
	}
	if len(block.Patterns) >= 1 && block.Patterns[0] != "%description GOOGLE PLAY" {
		t.Errorf("patterns[0]: got %q, want %%description GOOGLE PLAY", block.Patterns[0])
	}
	if len(block.Patterns) >= 2 && block.Patterns[1] != "%comment Google Play" {
		t.Errorf("patterns[1]: got %q, want %%comment Google Play", block.Patterns[1])
	}
	if len(block.Assignments) != 1 {
		t.Errorf("assignments: got %d, want 1", len(block.Assignments))
	}
	for _, d := range diags {
		if d.Code == "UNKNOWN_DIRECTIVE" {
			t.Errorf("unexpected UNKNOWN_DIRECTIVE diagnostic: %s", d.Message)
		}
	}
}

func TestParse_IfBlockMixedORAndAND(t *testing.T) {
	input := "if\n%description PAT1\n& %comment PAT2\n  account1 expenses"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 2 {
		t.Errorf("patterns: got %v, want 2", block.Patterns)
	}
	if len(block.Assignments) != 1 {
		t.Errorf("assignments: got %d, want 1", len(block.Assignments))
	}
}

func TestParse_IfBlockORPatternsOnly(t *testing.T) {
	input := "if\nPATTERN1\nPATTERN2"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 2 {
		t.Errorf("patterns: got %v, want 2", block.Patterns)
	}
	if len(block.Assignments) != 0 {
		t.Errorf("assignments: got %d, want 0", len(block.Assignments))
	}
	if block.Range.End.Line != 3 {
		t.Errorf("Range.End.Line: got %d, want 3", block.Range.End.Line)
	}
}

func TestParse_IfBlockInlineAndORPatterns(t *testing.T) {
	input := "if INLINE\nOR_PAT\n  account1 expenses"
	rf, _ := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 2 {
		t.Errorf("patterns: got %v, want 2", block.Patterns)
	}
	if len(block.Patterns) >= 1 && block.Patterns[0] != "INLINE" {
		t.Errorf("patterns[0]: got %q, want INLINE", block.Patterns[0])
	}
	if len(block.Patterns) >= 2 && block.Patterns[1] != "OR_PAT" {
		t.Errorf("patterns[1]: got %q, want OR_PAT", block.Patterns[1])
	}
	if len(block.Assignments) != 1 {
		t.Errorf("assignments: got %d, want 1", len(block.Assignments))
	}
}

func TestParse_IfBlockORPatternStopsAfterAssignment(t *testing.T) {
	input := "if\n  account1 expenses\nbare-text"
	rf, diags := Parse(input)
	if len(rf.IfBlocks) != 1 {
		t.Fatalf("ifBlocks: got %d, want 1", len(rf.IfBlocks))
	}
	block := rf.IfBlocks[0]
	if len(block.Patterns) != 0 {
		t.Errorf("patterns: got %v, want 0", block.Patterns)
	}
	if len(block.Assignments) != 1 {
		t.Errorf("assignments: got %d, want 1", len(block.Assignments))
	}
	foundUnknown := false
	for _, d := range diags {
		if d.Code == "UNKNOWN_DIRECTIVE" {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Error("expected UNKNOWN_DIRECTIVE diagnostic for bare-text")
	}
}
