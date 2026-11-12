package tick

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTickFile writes content to <dir>/<name>, creating dir as needed.
func writeTickFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

const learningsPreamble = "# Learnings\n\nRepo-specific gotchas. Size budget 16 KB.\n\n"

const (
	headPath = ".tick/" + LearningsHeadFile
	bodyPath = ".tick/" + LearningsBodyFile
)

func TestLearningsFileNames(t *testing.T) {
	if LearningsHeadFile != "learnings.md" {
		t.Errorf("LearningsHeadFile = %q, want learnings.md", LearningsHeadFile)
	}
	if LearningsBodyFile != "learnings-ref.md" {
		t.Errorf("LearningsBodyFile = %q, want learnings-ref.md", LearningsBodyFile)
	}
}

func TestLearningsLimitConstants(t *testing.T) {
	if LearningsHeadBudgetBytes != 4096 {
		t.Errorf("LearningsHeadBudgetBytes = %d, want 4096", LearningsHeadBudgetBytes)
	}
	if LearningsBodyBudgetBytes != 16384 {
		t.Errorf("LearningsBodyBudgetBytes = %d, want 16384", LearningsBodyBudgetBytes)
	}
	if LearningsEntryBytes != 800 {
		t.Errorf("LearningsEntryBytes = %d, want 800", LearningsEntryBytes)
	}
}

// Neither file present: nothing to say.
func TestCheckLearnings_BothAbsentSilent(t *testing.T) {
	c := CheckLearnings(t.TempDir())
	if w := c.Warnings(); len(w) != 0 {
		t.Errorf("no files: warnings %q, want none", w)
	}
}

// An absent body gives no warning, even next to an over-budget head (which
// warns only about itself).
func TestCheckLearnings_AbsentBodyNoWarning(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsHeadFile, strings.Repeat("h", LearningsHeadBudgetBytes+1))
	w := CheckLearnings(dir).Warnings()
	if len(w) != 1 {
		t.Fatalf("warnings = %q, want exactly the head warning", w)
	}
	if strings.Contains(w[0], LearningsBodyFile) {
		t.Errorf("absent body produced a body warning: %q", w[0])
	}
}

// An absent head stays silent; the body is still checked.
func TestCheckLearnings_AbsentHeadSilent(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsBodyFile, strings.Repeat("b", LearningsBodyBudgetBytes+1))
	w := CheckLearnings(dir).Warnings()
	if len(w) != 1 || !strings.HasPrefix(w[0], "warning: "+bodyPath+" ") {
		t.Errorf("warnings = %q, want exactly the body warning", w)
	}
}

func TestCheckLearnings_EmptyFilesSilent(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsHeadFile, "")
	writeTickFile(t, dir, LearningsBodyFile, "")
	if w := CheckLearnings(dir).Warnings(); len(w) != 0 {
		t.Errorf("empty files: warnings %q, want none", w)
	}
}

// Head boundary: 4096 bytes is within budget, 4097 warns naming the head file.
func TestCheckLearnings_HeadBoundary(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsHeadFile, strings.Repeat("h", LearningsHeadBudgetBytes))
	if w := CheckLearnings(dir).Warnings(); len(w) != 0 {
		t.Errorf("head of %d bytes: warnings %q, want none", LearningsHeadBudgetBytes, w)
	}

	writeTickFile(t, dir, LearningsHeadFile, strings.Repeat("h", LearningsHeadBudgetBytes+1))
	c := CheckLearnings(dir)
	if !c.Head.OverBudget || c.Head.Bytes != LearningsHeadBudgetBytes+1 {
		t.Errorf("head report = %+v, want over budget at %d bytes", c.Head, LearningsHeadBudgetBytes+1)
	}
	w := c.Warnings()
	want := fmt.Sprintf("warning: %s is %d bytes (budget %d)", headPath, LearningsHeadBudgetBytes+1, LearningsHeadBudgetBytes)
	if len(w) != 1 || !strings.HasPrefix(w[0], want) {
		t.Errorf("warnings = %q, want one starting %q", w, want)
	}
	if len(w) == 1 && !strings.Contains(w[0], "move entries to the body") {
		t.Errorf("head warning %q lacks the head hint", w[0])
	}
}

// Body boundary: 16384 bytes is within budget, 16385 warns naming the body file.
func TestCheckLearnings_BodyBoundary(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsBodyFile, strings.Repeat("b", LearningsBodyBudgetBytes))
	if w := CheckLearnings(dir).Warnings(); len(w) != 0 {
		t.Errorf("body of %d bytes: warnings %q, want none", LearningsBodyBudgetBytes, w)
	}

	writeTickFile(t, dir, LearningsBodyFile, strings.Repeat("b", LearningsBodyBudgetBytes+1))
	w := CheckLearnings(dir).Warnings()
	want := fmt.Sprintf("warning: %s is %d bytes (budget %d)", bodyPath, LearningsBodyBudgetBytes+1, LearningsBodyBudgetBytes)
	if len(w) != 1 || !strings.HasPrefix(w[0], want) {
		t.Errorf("warnings = %q, want one starting %q", w, want)
	}
	if len(w) == 1 && !strings.Contains(w[0], "not compaction") {
		t.Errorf("body warning %q lacks the body hint", w[0])
	}
}

// An unreadable file (here a directory in its place) yields a zero report for
// that file, while the other file is still checked — both ways round.
func TestCheckLearnings_UnreadableFileZeroReport(t *testing.T) {
	cases := []struct {
		unreadable, other, content, wantPrefix string
	}{
		{LearningsHeadFile, LearningsBodyFile, strings.Repeat("b", LearningsBodyBudgetBytes+1), "warning: " + bodyPath + " "},
		{LearningsBodyFile, LearningsHeadFile, strings.Repeat("h", LearningsHeadBudgetBytes+1), "warning: " + headPath + " "},
	}
	for _, tc := range cases {
		t.Run(tc.unreadable, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, tc.unreadable), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			writeTickFile(t, dir, tc.other, tc.content)
			w := CheckLearnings(dir).Warnings()
			if len(w) != 1 || !strings.HasPrefix(w[0], tc.wantPrefix) {
				t.Errorf("warnings = %q, want exactly one starting %q", w, tc.wantPrefix)
			}
		})
	}
}

// Body entry boundary: an 800-byte entry is fine, an 801-byte one warns with
// its line number, naming the body file.
func TestCheckLearnings_BodyEntryBoundary(t *testing.T) {
	dir := t.TempDir()
	at := strings.Repeat("a", LearningsEntryBytes)
	writeTickFile(t, dir, LearningsBodyFile, "# Learnings ref\n\n## H\n\n"+at+"\n")
	if w := CheckLearnings(dir).Warnings(); len(w) != 0 {
		t.Errorf("entry of %d bytes: warnings %q, want none", LearningsEntryBytes, w)
	}

	writeTickFile(t, dir, LearningsBodyFile, "# Learnings ref\n\n## H\n\n"+at+"a\n")
	w := CheckLearnings(dir).Warnings()
	want := fmt.Sprintf("warning: %s:5 entry ", bodyPath)
	if len(w) != 1 || !strings.HasPrefix(w[0], want) ||
		!strings.Contains(w[0], fmt.Sprintf("%d bytes (limit %d)", LearningsEntryBytes+1, LearningsEntryBytes)) {
		t.Errorf("warnings = %q, want one starting %q naming size and limit", w, want)
	}
}

// The entry limit applies to the body only: a fat entry in the head does not
// warn while the head is within its own budget.
func TestCheckLearnings_EntryLimitNotAppliedToHead(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsHeadFile, "# Learnings\n\n## Rules\n\n"+strings.Repeat("r", LearningsEntryBytes+100)+"\n")
	if w := CheckLearnings(dir).Warnings(); len(w) != 0 {
		t.Errorf("fat head entry: warnings %q, want none", w)
	}
}

// Both files over: warnings come per file, head first, each naming its own file.
func TestCheckLearnings_WarningsPerFile(t *testing.T) {
	dir := t.TempDir()
	writeTickFile(t, dir, LearningsHeadFile, strings.Repeat("h", LearningsHeadBudgetBytes+1))
	writeTickFile(t, dir, LearningsBodyFile, "## H\n\n"+strings.Repeat("e", LearningsEntryBytes+1)+"\n\n"+
		strings.Repeat("small\n\n", LearningsBodyBudgetBytes/7))
	w := CheckLearnings(dir).Warnings()
	if len(w) != 3 {
		t.Fatalf("warnings = %q, want head, body and entry", w)
	}
	if !strings.HasPrefix(w[0], "warning: "+headPath+" ") {
		t.Errorf("w[0] = %q, want the head warning", w[0])
	}
	if !strings.HasPrefix(w[1], "warning: "+bodyPath+" ") {
		t.Errorf("w[1] = %q, want the body budget warning", w[1])
	}
	if !strings.HasPrefix(w[2], "warning: "+bodyPath+":3 ") {
		t.Errorf("w[2] = %q, want the body entry warning at line 3", w[2])
	}
}

// A 150-line body whose lines are long — entries merged into long lines to
// meet the old line cap — is over the byte budget and warns.
func TestMeasureLearningsBody_150LongLinesWarns(t *testing.T) {
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

	r := MeasureLearningsBody([]byte(content))
	if r.Bytes != len(content) {
		t.Errorf("Bytes = %d, want %d", r.Bytes, len(content))
	}
	if !r.OverBudget {
		t.Fatalf("OverBudget = false for %d bytes in 150 lines, want true", r.Bytes)
	}
	w := r.Warnings()
	if len(w) == 0 || !strings.Contains(w[0], fmt.Sprintf("%d bytes (budget %d)", r.Bytes, LearningsBodyBudgetBytes)) {
		t.Errorf("warnings = %q, want a total-size warning naming bytes and budget", w)
	}
}

// A 400-line body under the byte budget, made of small entries, does not
// warn: line count does not matter.
func TestMeasureLearningsBody_400ShortLinesUnderBudgetDoesNotWarn(t *testing.T) {
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
	if len(content) > LearningsBodyBudgetBytes {
		t.Fatalf("fixture is %d bytes, want <= %d", len(content), LearningsBodyBudgetBytes)
	}

	r := MeasureLearningsBody([]byte(content))
	if r.OverBudget || len(r.FatEntries) != 0 || len(r.Warnings()) != 0 {
		t.Errorf("got %+v, warnings %q; want no warning", r, r.Warnings())
	}
}

// One body entry over the per-entry limit is named (first ~60 chars and its
// size), even when the file as a whole is within budget.
func TestMeasureLearningsBody_FatEntryNamed(t *testing.T) {
	fat := "**Problem:** the fat one " + strings.Repeat("y", 700) + "\n" +
		"**Cause:** packed\n" +
		"**Rule:** " + strings.Repeat("z", 100)
	content := learningsPreamble +
		"## Build\n\n" +
		"**Problem:** small\n**Cause:** c\n**Rule:** r\n\n" +
		fat + "\n\n" +
		"## Tests\n\n" +
		"**Problem:** also small\n**Rule:** r\n"

	r := MeasureLearningsBody([]byte(content))
	if r.OverBudget {
		t.Errorf("OverBudget = true for %d bytes, want false", r.Bytes)
	}
	if len(r.FatEntries) != 1 {
		t.Fatalf("FatEntries = %+v, want exactly one", r.FatEntries)
	}
	e := r.FatEntries[0]
	if e.Line != 11 {
		t.Errorf("entry Line = %d, want 11", e.Line)
	}
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

// The title and preamble before the first ## header, and headers themselves,
// are not entries — however long.
func TestMeasureLearningsBody_PreambleAndHeadersAreNotEntries(t *testing.T) {
	long := strings.Repeat("p", 900)
	content := "# Learnings\n\n" + long + "\n\n## " + long + "\n\n**Rule:** r\n"
	r := MeasureLearningsBody([]byte(content))
	if len(r.FatEntries) != 0 {
		t.Errorf("preamble or header counted as an entry: %+v", r.FatEntries)
	}
}
