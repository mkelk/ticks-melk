package tick

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Learnings are split in two files. The head, .tick/learnings.md, is read in
// full by every implementer and at every planning pass, so its size is a
// per-agent context tax: it holds hard rules plus an index of the body's
// categories. The body, .tick/learnings-ref.md, holds the Problem/Cause/Rule
// entries and is looked up on demand. Caps are size budgets in bytes, not line
// counts: a line cap is met by merging entries into longer lines, which keeps
// every byte the reader pays for. The engine owns the format; the project owns
// the content.
const (
	// LearningsHeadFile is the always-read head, relative to .tick/.
	LearningsHeadFile = "learnings.md"
	// LearningsBodyFile is the looked-up body, relative to .tick/.
	LearningsBodyFile = "learnings-ref.md"

	// LearningsHeadBudgetBytes is the recommended maximum size of the head,
	// in UTF-8 bytes (~4 KB, ~1k tokens).
	LearningsHeadBudgetBytes = 4096
	// LearningsBodyBudgetBytes is the recommended maximum size of the body,
	// in UTF-8 bytes (~16 KB, ~4k tokens).
	LearningsBodyBudgetBytes = 16384
	// LearningsEntryBytes is the recommended maximum size of one body entry —
	// one Problem/Cause/Rule block — so a single entry cannot take over the
	// file. It applies to the body only.
	LearningsEntryBytes = 800
)

// learningsPreviewRunes is how much of a fat entry's first line a warning shows.
const learningsPreviewRunes = 60

// LearningsEntry is one body entry over LearningsEntryBytes.
type LearningsEntry struct {
	Line    int    // 1-based line number of the entry's first line
	Preview string // the entry's first line, cut to ~60 characters
	Bytes   int    // the entry's size in bytes (its lines joined by "\n")
}

// LearningsReport is the measurement of one learnings file against its budget.
// The zero value (an absent file) has no warnings.
type LearningsReport struct {
	File       string           // display path, e.g. ".tick/learnings.md"
	Bytes      int              // total file size in bytes
	Budget     int              // the file's budget in bytes
	OverBudget bool             // Bytes > Budget
	FatEntries []LearningsEntry // body entries over LearningsEntryBytes, in file order
	hint       string           // what to do when OverBudget
}

// LearningsCheck is the measurement of both learnings files.
type LearningsCheck struct {
	Head LearningsReport
	Body LearningsReport
}

// MeasureLearningsHead measures head content against LearningsHeadBudgetBytes.
// The entry limit does not apply to the head.
func MeasureLearningsHead(content []byte) LearningsReport {
	return LearningsReport{
		File:       ".tick/" + LearningsHeadFile,
		Bytes:      len(content),
		Budget:     LearningsHeadBudgetBytes,
		OverBudget: len(content) > LearningsHeadBudgetBytes,
		hint:       "keep it to hard rules and an index of the body's categories; move entries to the body",
	}
}

// MeasureLearningsBody measures body content against LearningsBodyBudgetBytes
// and each entry against LearningsEntryBytes.
//
// An entry is a run of non-blank lines between blank lines, under a "##"
// category header. Lines starting with "#" (headers) are not entries, and
// nothing before the first "##" header (the title and preamble) is.
func MeasureLearningsBody(content []byte) LearningsReport {
	r := LearningsReport{
		File:       ".tick/" + LearningsBodyFile,
		Bytes:      len(content),
		Budget:     LearningsBodyBudgetBytes,
		OverBudget: len(content) > LearningsBodyBudgetBytes,
		hint:       "compact it at the next retro; merging entries into longer lines is not compaction",
	}

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

// CheckLearnings measures <tickDir>/learnings.md (head) and
// <tickDir>/learnings-ref.md (body) against their budgets.
//
// A missing or unreadable file yields a zero report for that file: this lint
// must never block or fail a command.
func CheckLearnings(tickDir string) LearningsCheck {
	var c LearningsCheck
	if content, err := os.ReadFile(filepath.Join(tickDir, LearningsHeadFile)); err == nil {
		c.Head = MeasureLearningsHead(content)
	}
	if content, err := os.ReadFile(filepath.Join(tickDir, LearningsBodyFile)); err == nil {
		c.Body = MeasureLearningsBody(content)
	}
	return c
}

// Warnings renders both reports as warning lines, head first, each naming its
// own file; none when both are within budget or absent.
func (c LearningsCheck) Warnings() []string {
	return append(c.Head.Warnings(), c.Body.Warnings()...)
}

// Warnings renders the report as warning lines (without trailing newlines);
// none when the file is within budget.
func (r LearningsReport) Warnings() []string {
	var w []string
	if r.OverBudget {
		w = append(w, fmt.Sprintf(
			"warning: %s is %d bytes (budget %d) — %s",
			r.File, r.Bytes, r.Budget, r.hint))
	}
	for _, e := range r.FatEntries {
		w = append(w, fmt.Sprintf(
			"warning: %s:%d entry %q is %d bytes (limit %d) — tighten it at the next retro",
			r.File, e.Line, e.Preview, e.Bytes, LearningsEntryBytes))
	}
	return w
}
