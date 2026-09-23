package tick

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLearnings writes content to <dir>/learnings.md, creating dir as needed.
func writeLearnings(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "learnings.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

const learningsPreamble = "# Learnings\n\nRepo-specific gotchas. Size budget 16 KB.\n\n"

func TestCheckLearnings_MissingFile(t *testing.T) {
	r := CheckLearnings(t.TempDir())
	if r.Bytes != 0 || r.OverBudget || len(r.FatEntries) != 0 || len(r.Warnings()) != 0 {
		t.Errorf("missing file: got %+v, want zero report", r)
	}
}

func TestCheckLearnings_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeLearnings(t, dir, "")
	r := CheckLearnings(dir)
	if r.Bytes != 0 || r.OverBudget || len(r.Warnings()) != 0 {
		t.Errorf("empty file: got %+v, want zero report", r)
	}
}

// A 150-line file whose lines are long — entries merged into long lines to meet
// the old line cap — is over the byte budget and warns.
func TestCheckLearnings_150LongLinesWarns(t *testing.T) {
	var b strings.Builder
	b.WriteString(learningsPreamble)
	b.WriteString("## Build\n\n")
	// 144 more lines of ~150 bytes, one line per entry.
	for i := 0; i < 144; i++ {
		fmt.Fprintf(&b, "**Rule %03d:** %s\n", i, strings.Repeat("x", 136))
	}
	content := b.String()
	if n := strings.Count(content, "\n"); n != 150 {
		t.Fatalf("fixture has %d lines, want 150", n)
	}
	dir := t.TempDir()
	writeLearnings(t, dir, content)

	r := CheckLearnings(dir)
	if r.Bytes != len(content) {
		t.Errorf("Bytes = %d, want %d", r.Bytes, len(content))
	}
	if !r.OverBudget {
		t.Fatalf("OverBudget = false for %d bytes in 150 lines, want true", r.Bytes)
	}
	w := r.Warnings()
	if len(w) == 0 || !strings.Contains(w[0], fmt.Sprintf("%d bytes (budget %d)", r.Bytes, LearningsBudgetBytes)) {
		t.Errorf("warnings = %q, want a total-size warning naming bytes and budget", w)
	}
}

// A 400-line file under the byte budget, made of small entries, does not warn:
// line count no longer matters.
func TestCheckLearnings_400ShortLinesUnderBudgetDoesNotWarn(t *testing.T) {
	var b strings.Builder
	b.WriteString(learningsPreamble) // 4 lines
	b.WriteString("## Build\n\n")    // 2 lines
	// 98 entries of 4 lines each (3 content + 1 blank) = 392 lines; total 398.
	for i := 0; i < 98; i++ {
		fmt.Fprintf(&b, "**Problem:** p%02d\n**Cause:** c\n**Rule:** r\n\n", i)
	}
	b.WriteString("end\nend\n")
	content := b.String()
	if n := strings.Count(content, "\n"); n != 400 {
		t.Fatalf("fixture has %d lines, want 400", n)
	}
	if len(content) > LearningsBudgetBytes {
		t.Fatalf("fixture is %d bytes, want <= %d", len(content), LearningsBudgetBytes)
	}
	dir := t.TempDir()
	writeLearnings(t, dir, content)

	r := CheckLearnings(dir)
	if r.OverBudget || len(r.FatEntries) != 0 || len(r.Warnings()) != 0 {
		t.Errorf("got %+v, warnings %q; want no warning", r, r.Warnings())
	}
}

// One entry over the per-entry limit is named (first ~60 chars and its size),
// even when the file as a whole is within budget.
func TestCheckLearnings_FatEntryNamed(t *testing.T) {
	fat := "**Problem:** the fat one " + strings.Repeat("y", 700) + "\n" +
		"**Cause:** packed\n" +
		"**Rule:** " + strings.Repeat("z", 100)
	content := learningsPreamble +
		"## Build\n\n" +
		"**Problem:** small\n**Cause:** c\n**Rule:** r\n\n" +
		fat + "\n\n" +
		"## Tests\n\n" +
		"**Problem:** also small\n**Rule:** r\n"
	dir := t.TempDir()
	writeLearnings(t, dir, content)

	r := CheckLearnings(dir)
	if r.OverBudget {
		t.Errorf("OverBudget = true for %d bytes, want false", r.Bytes)
	}
	if len(r.FatEntries) != 1 {
		t.Fatalf("FatEntries = %+v, want exactly one", r.FatEntries)
	}
	e := r.FatEntries[0]
	if e.Bytes != len(fat) {
		t.Errorf("entry Bytes = %d, want %d", e.Bytes, len(fat))
	}
	if !strings.HasPrefix(e.Preview, "**Problem:** the fat one") || len([]rune(e.Preview)) > 61 {
		t.Errorf("entry Preview = %q, want the entry's first ~60 chars", e.Preview)
	}
	w := r.Warnings()
	if len(w) != 1 || !strings.Contains(w[0], e.Preview) ||
		!strings.Contains(w[0], fmt.Sprintf("%d bytes (limit %d)", e.Bytes, LearningsEntryBytes)) {
		t.Errorf("warnings = %q, want one naming the fat entry and its size", w)
	}
}

// An entry of exactly the limit is not fat; one byte more is.
func TestMeasureLearnings_EntryLimitBoundary(t *testing.T) {
	at := strings.Repeat("a", LearningsEntryBytes)
	if r := MeasureLearnings([]byte("## H\n\n" + at + "\n")); len(r.FatEntries) != 0 {
		t.Errorf("entry of exactly %d bytes flagged: %+v", LearningsEntryBytes, r.FatEntries)
	}
	if r := MeasureLearnings([]byte("## H\n\n" + at + "a\n")); len(r.FatEntries) != 1 {
		t.Errorf("entry of %d bytes not flagged", LearningsEntryBytes+1)
	}
}

// The title and preamble before the first ## header, and headers themselves,
// are not entries — however long.
func TestMeasureLearnings_PreambleAndHeadersAreNotEntries(t *testing.T) {
	long := strings.Repeat("p", 900)
	content := "# Learnings\n\n" + long + "\n\n## " + long + "\n\n**Rule:** r\n"
	r := MeasureLearnings([]byte(content))
	if len(r.FatEntries) != 0 {
		t.Errorf("preamble or header counted as an entry: %+v", r.FatEntries)
	}
}

// Budget boundary: exactly LearningsBudgetBytes is within budget.
func TestMeasureLearnings_BudgetBoundary(t *testing.T) {
	if r := MeasureLearnings(make([]byte, LearningsBudgetBytes)); r.OverBudget {
		t.Errorf("exactly %d bytes reported over budget", LearningsBudgetBytes)
	}
	if r := MeasureLearnings(make([]byte, LearningsBudgetBytes+1)); !r.OverBudget {
		t.Errorf("%d bytes not reported over budget", LearningsBudgetBytes+1)
	}
}

func TestLearningsLimitConstants(t *testing.T) {
	if LearningsBudgetBytes != 16384 {
		t.Errorf("LearningsBudgetBytes = %d, want 16384", LearningsBudgetBytes)
	}
	if LearningsEntryBytes != 800 {
		t.Errorf("LearningsEntryBytes = %d, want 800", LearningsEntryBytes)
	}
}
