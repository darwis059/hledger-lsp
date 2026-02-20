package rules

import (
	"testing"
)

func TestLinks_Empty(t *testing.T) {
	links := Links(parseRF(""), "")
	if links == nil {
		t.Fatal("expected non-nil links slice")
	}
}

func TestLinks_IncludeDirective(t *testing.T) {
	input := "include common.rules"
	links := Links(parseRF(input), "/home/user/import/")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Path != "/home/user/import/common.rules" {
		t.Errorf("path: got %q, want /home/user/import/common.rules", links[0].Path)
	}
}

func TestLinks_SourceDirective(t *testing.T) {
	input := "source data.csv"
	links := Links(parseRF(input), "/home/user/import/")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Path != "/home/user/import/data.csv" {
		t.Errorf("path: got %q, want /home/user/import/data.csv", links[0].Path)
	}
}

func TestLinks_AbsolutePath(t *testing.T) {
	input := "include /absolute/path.rules"
	links := Links(parseRF(input), "/any/dir/")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Path != "/absolute/path.rules" {
		t.Errorf("path: got %q, want /absolute/path.rules", links[0].Path)
	}
}

func TestLinks_Range(t *testing.T) {
	input := "include common.rules"
	links := Links(parseRF(input), "/dir/")
	if len(links) == 0 {
		t.Fatal("expected links")
	}
	if links[0].Range.Start.Line != 1 {
		t.Errorf("start line: got %d, want 1", links[0].Range.Start.Line)
	}
}

func TestLinks_IncludeRangeCoverPathOnly(t *testing.T) {
	// "include common.rules" — range should start at "common.rules", not "include"
	input := "include common.rules"
	links := Links(parseRF(input), "/dir/")
	if len(links) == 0 {
		t.Fatal("expected links")
	}
	lnk := links[0]
	// "include " is 8 chars (col 1-8), so path starts at col 9
	if lnk.Range.Start.Column != 9 {
		t.Errorf("Start.Column: got %d, want 9", lnk.Range.Start.Column)
	}
	// "common.rules" is 12 chars, ends at col 21
	if lnk.Range.End.Column != 21 {
		t.Errorf("End.Column: got %d, want 21", lnk.Range.End.Column)
	}
}

func TestLinks_SourceRangeCoverPathOnly(t *testing.T) {
	// "source data.csv" — range should start at "data.csv", not "source"
	input := "source data.csv"
	links := Links(parseRF(input), "/dir/")
	if len(links) == 0 {
		t.Fatal("expected links")
	}
	lnk := links[0]
	// "source " is 7 chars (col 1-7), so path starts at col 8
	if lnk.Range.Start.Column != 8 {
		t.Errorf("Start.Column: got %d, want 8", lnk.Range.Start.Column)
	}
	// "data.csv" is 8 chars, ends at col 16
	if lnk.Range.End.Column != 16 {
		t.Errorf("End.Column: got %d, want 16", lnk.Range.End.Column)
	}
}

func TestLinks_SourceWithPipe(t *testing.T) {
	input := "source data.csv | sort"
	links := Links(parseRF(input), "/base")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Path != "/base/data.csv" {
		t.Errorf("path: got %q, want /base/data.csv", links[0].Path)
	}
}
