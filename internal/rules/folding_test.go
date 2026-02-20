package rules

import (
	"testing"
)

func TestFolding_Empty(t *testing.T) {
	ranges := FoldingRanges(nil)
	if ranges == nil {
		t.Fatal("expected non-nil ranges")
	}
}

func TestFolding_IfBlock(t *testing.T) {
	input := "if PATTERN\n  account1 expenses:food\n  description groceries"
	ranges := FoldingRanges(parseRF(input))
	found := false
	for _, r := range ranges {
		if r.StartLine == 0 && r.EndLine >= 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected folding range for if block")
	}
}

func TestFolding_CommentBlock(t *testing.T) {
	input := "# comment 1\n# comment 2\n# comment 3"
	ranges := FoldingRanges(parseRF(input))
	found := false
	for _, r := range ranges {
		if r.StartLine == 0 && r.EndLine == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected folding range for comment block")
	}
}

func TestFolding_SingleLineIfBlock(t *testing.T) {
	// If block with only header, no body -> no fold
	input := "if PATTERN"
	ranges := FoldingRanges(parseRF(input))
	if len(ranges) != 0 {
		t.Errorf("expected no folding ranges for single-line if block, got %d", len(ranges))
	}
}

func TestFolding_ContinuationOnlyIfBlock(t *testing.T) {
	input := "if PAT\n& PAT2"
	ranges := FoldingRanges(parseRF(input))
	found := false
	for _, r := range ranges {
		if r.StartLine == 0 && r.EndLine == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected folding range StartLine=0 EndLine=1 for continuation-only if block, got %v", ranges)
	}
}

func TestFolding_NoFoldableContent_ReturnsEmptySlice(t *testing.T) {
	// Non-empty input with no foldable regions should return empty slice, not nil
	input := "skip 1\nfields date, amount"
	ranges := FoldingRanges(parseRF(input))
	if ranges == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(ranges) != 0 {
		t.Errorf("expected 0 ranges, got %d", len(ranges))
	}
}
