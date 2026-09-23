package tick

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// .tick/learnings.md is read in full by every implementer and at every
// planning pass, so its size is a per-agent context tax. The cap is a size
// budget in bytes, not a line count: a line cap is met by merging entries into
// longer lines, which keeps every byte the reader pays for.
const (
	// LearningsBudgetBytes is the recommended maximum size of
	// .tick/learnings.md, in UTF-8 bytes (~16 KB, ~4k tokens).
	LearningsBudgetBytes = 16384
	// LearningsEntryBytes is the recommended maximum size of one entry — one
	// Problem/Cause/Rule block — so a single entry cannot take over the file.
	LearningsEntryBytes = 800
)

// learningsPreviewRunes is how much of a fat entry's first line a warning shows.
const learningsPreviewRunes = 60

// LearningsEntry is one entry over LearningsEntryBytes.
type LearningsEntry struct {
	Line    int    // 1-based line number of the entry's first line
	Preview string // the entry's first line, cut to ~60 characters
	Bytes   int    // the entry's size in bytes (its lines joined by "\n")
}

// LearningsReport is the measurement of a learnings file against the budget.
type LearningsReport struct {
	Bytes      int              // total file size in bytes
	OverBudget bool             // Bytes > LearningsBudgetBytes
	FatEntries []LearningsEntry // entries over LearningsEntryBytes, in file order
}

// MeasureLearnings measures learnings content against the size budget.
//
// An entry is a run of non-blank lines between blank lines, under a "##"
// category header. Lines starting with "#" (headers) are not entries, and
// nothing before the first "##" header (the title and preamble) is.
func MeasureLearnings(content []byte) LearningsReport {
	r := LearningsReport{Bytes: len(content)}
	r.OverBudget = r.Bytes > LearningsBudgetBytes

	lines := strings.Split(string(content), "\n")
	inBody := false
	var block []string
	start := 0
	flush := func() {
		if len(block) == 0 {
			return
		}
		size := len(strings.Join(block, "\n"))
		if size > LearningsEntryBytes {
			r.FatEntries = append(r.FatEntries, LearningsEntry{
				Line:    start,
				Preview: learningsPreview(block[0]),
				Bytes:   size,
			})
		}
		block = nil
	}
	for i, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#"):
			flush()
			if strings.HasPrefix(trimmed, "##") {
				inBody = true
			}
		case trimmed == "":
			flush()
		case inBody:
			if len(block) == 0 {
				start = i + 1
			}
			block = append(block, line)
		}
	}
	flush()
	return r
}

func learningsPreview(line string) string {
	s := strings.TrimSpace(line)
	if utf8.RuneCountInString(s) <= learningsPreviewRunes {
		return s
	}
	return string([]rune(s)[:learningsPreviewRunes]) + "…"
}

// CheckLearnings measures <tickDir>/learnings.md against the size budget.
//
// A missing or unreadable file yields a zero report: this lint must never
// block or fail a command.
func CheckLearnings(tickDir string) LearningsReport {
	content, err := os.ReadFile(filepath.Join(tickDir, "learnings.md"))
	if err != nil {
		return LearningsReport{}
	}
	return MeasureLearnings(content)
}

// Warnings renders the report as warning lines (without trailing newlines);
// none when the file is within budget.
func (r LearningsReport) Warnings() []string {
	var w []string
	if r.OverBudget {
		w = append(w, fmt.Sprintf(
			"warning: .tick/learnings.md is %d bytes (budget %d) — compact it at the next retro; merging entries into longer lines is not compaction",
			r.Bytes, LearningsBudgetBytes))
	}
	for _, e := range r.FatEntries {
		w = append(w, fmt.Sprintf(
			"warning: .tick/learnings.md:%d entry %q is %d bytes (limit %d) — tighten it at the next retro",
			e.Line, e.Preview, e.Bytes, LearningsEntryBytes))
	}
	return w
}
