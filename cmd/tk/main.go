package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pengelbrecht/ticks/internal/config"
	"github.com/pengelbrecht/ticks/internal/github"
	"github.com/pengelbrecht/ticks/internal/tick"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}

	switch args[1] {
	case "init":
		return runInit()
	case "whoami":
		return runWhoami(args[2:])
	case "show":
		return runShow(args[2:])
	case "create":
		return runCreate(args[2:])
	case "block":
		return runBlock(args[2:])
	case "unblock":
		return runUnblock(args[2:])
	case "update":
		return runUpdate(args[2:])
	case "close":
		return runClose(args[2:])
	case "reopen":
		return runReopen(args[2:])
	case "--help", "-h":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[1])
		printUsage()
		return 2
	}
}

func runInit() int {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}

	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	owner, err := github.DetectOwner(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect owner: %v\n", err)
		return 5
	}

	tickDir := filepath.Join(root, ".tick")
	if err := os.MkdirAll(filepath.Join(tickDir, "issues"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create .tick directory: %v\n", err)
		return 6
	}

	if err := config.Save(filepath.Join(tickDir, "config.json"), config.Default()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write config: %v\n", err)
		return 6
	}

	if err := os.WriteFile(filepath.Join(tickDir, ".gitignore"), []byte(".index.json\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .gitignore: %v\n", err)
		return 6
	}

	if err := github.EnsureGitAttributes(root); err != nil {
		fmt.Fprintf(os.Stderr, "failed to update .gitattributes: %v\n", err)
		return 6
	}
	if err := github.ConfigureMergeDriver(root); err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure merge driver: %v\n", err)
		return 6
	}

	fmt.Printf("Detected GitHub repo: %s\n", project)
	fmt.Printf("Detected user: %s\n\n", owner)
	fmt.Println("Initialized .tick/")

	return 0
}

func runWhoami(args []string) int {
	fs := flag.NewFlagSet("whoami", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "output as json")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	owner, err := github.DetectOwner(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect owner: %v\n", err)
		return 5
	}

	if *jsonOutput {
		payload := map[string]string{"owner": owner, "project": project}
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(payload); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode json: %v\n", err)
			return 6
		}
		return 0
	}

	fmt.Printf("Owner: %s\n", owner)
	fmt.Printf("Project: %s\n", project)
	return 0
}

func runCreate(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	description := fs.String("description", "", "detailed description")
	fs.StringVar(description, "d", "", "detailed description")
	priority := fs.Int("priority", 2, "priority 0-4")
	fs.IntVar(priority, "p", 2, "priority 0-4")
	typeFlag := fs.String("type", tick.TypeTask, "type")
	fs.StringVar(typeFlag, "t", tick.TypeTask, "type")
	ownerFlag := fs.String("owner", "", "owner")
	fs.StringVar(ownerFlag, "o", "", "owner")
	labelsFlag := fs.String("labels", "", "comma-separated labels")
	fs.StringVar(labelsFlag, "l", "", "comma-separated labels")
	blockedFlag := fs.String("blocked-by", "", "comma-separated blocker ids")
	fs.StringVar(blockedFlag, "b", "", "comma-separated blocker ids")
	parentFlag := fs.String("parent", "", "parent epic id")
	discoveredFlag := fs.String("discovered-from", "", "source tick id")
	jsonOutput := fs.Bool("json", false, "output as json")
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	title := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if title == "" {
		fmt.Fprintln(os.Stderr, "title is required")
		return 2
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	cfg, err := config.Load(filepath.Join(root, ".tick", "config.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 6
	}

	creator, err := github.DetectOwner(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect owner: %v\n", err)
		return 5
	}

	owner := creator
	if strings.TrimSpace(*ownerFlag) != "" {
		owner = strings.TrimSpace(*ownerFlag)
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	gen := tick.NewIDGenerator(nil)
	id, newLen, err := gen.Generate(func(candidate string) bool {
		_, err := os.Stat(filepath.Join(root, ".tick", "issues", candidate+".json"))
		return err == nil
	}, cfg.IDLength)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate id: %v\n", err)
		return 6
	}

	now := time.Now().UTC()
	t := tick.Tick{
		ID:             id,
		Title:          title,
		Description:    strings.TrimSpace(*description),
		Status:         tick.StatusOpen,
		Priority:       *priority,
		Type:           strings.TrimSpace(*typeFlag),
		Owner:          owner,
		Labels:         splitCSV(*labelsFlag),
		BlockedBy:      splitCSV(*blockedFlag),
		Parent:         strings.TrimSpace(*parentFlag),
		DiscoveredFrom: strings.TrimSpace(*discoveredFlag),
		CreatedBy:      creator,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := store.Write(t); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write tick: %v\n", err)
		return 6
	}

	if newLen != cfg.IDLength {
		cfg.IDLength = newLen
		if err := config.Save(filepath.Join(root, ".tick", "config.json"), cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to update config: %v\n", err)
			return 6
		}
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(t); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode json: %v\n", err)
			return 6
		}
		return 0
	}

	fmt.Printf("%s\n", t.ID)
	return 0
}

func runShow(args []string) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "output as json")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "id is required")
		return 2
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	id, err := github.NormalizeID(project, fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %v\n", err)
		return 4
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	t, err := store.Read(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read tick: %v\n", err)
		return 4
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(t); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode json: %v\n", err)
			return 6
		}
		return 0
	}

	fmt.Printf("%s  P%d %s  %s  @%s\n\n", t.ID, t.Priority, t.Type, t.Status, t.Owner)
	fmt.Printf("%s\n\n", t.Title)

	if strings.TrimSpace(t.Description) != "" {
		fmt.Println("Description:")
		fmt.Printf("  %s\n\n", t.Description)
	}

	if strings.TrimSpace(t.Notes) != "" {
		fmt.Println("Notes:")
		for _, line := range strings.Split(t.Notes, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	if len(t.Labels) > 0 {
		fmt.Printf("Labels: %s\n", strings.Join(t.Labels, ", "))
	}
	if len(t.BlockedBy) > 0 {
		var blocked []string
		for _, blocker := range t.BlockedBy {
			blk, err := store.Read(blocker)
			if err != nil {
				blocked = append(blocked, fmt.Sprintf("%s (unknown)", blocker))
				continue
			}
			blocked = append(blocked, fmt.Sprintf("%s (%s)", blocker, blk.Status))
		}
		fmt.Printf("Blocked by: %s\n", strings.Join(blocked, ", "))
	}

	fmt.Printf("Created: %s by %s\n", formatTime(t.CreatedAt), t.CreatedBy)
	fmt.Printf("Updated: %s\n\n", formatTime(t.UpdatedAt))

	fmt.Printf("Global: %s:%s\n", project, t.ID)

	return 0
}

func runBlock(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tk block <id> <blocker-id>")
		return 2
	}
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	id, err := github.NormalizeID(project, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %v\n", err)
		return 4
	}
	blockerID, err := github.NormalizeID(project, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid blocker id: %v\n", err)
		return 4
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	t, err := store.Read(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read tick: %v\n", err)
		return 4
	}
	if _, err := store.Read(blockerID); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read blocker tick: %v\n", err)
		return 4
	}

	t.BlockedBy = appendUnique(t.BlockedBy, blockerID)
	t.UpdatedAt = time.Now().UTC()

	if err := store.Write(t); err != nil {
		fmt.Fprintf(os.Stderr, "failed to update tick: %v\n", err)
		return 6
	}

	return 0
}

func runUnblock(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tk unblock <id> <blocker-id>")
		return 2
	}
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	id, err := github.NormalizeID(project, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %v\n", err)
		return 4
	}
	blockerID, err := github.NormalizeID(project, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid blocker id: %v\n", err)
		return 4
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	t, err := store.Read(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read tick: %v\n", err)
		return 4
	}

	t.BlockedBy = removeString(t.BlockedBy, blockerID)
	t.UpdatedAt = time.Now().UTC()

	if err := store.Write(t); err != nil {
		fmt.Fprintf(os.Stderr, "failed to update tick: %v\n", err)
		return 6
	}

	return 0
}

func runUpdate(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	var title, description, notes, status, typ, owner optionalString
	var priority optionalInt
	var addLabels, removeLabels optionalString
	jsonOutput := fs.Bool("json", false, "output as json")

	fs.Var(&title, "title", "new title")
	fs.Var(&description, "description", "new description")
	fs.Var(&notes, "notes", "replace notes")
	fs.Var(&status, "status", "new status")
	fs.Var(&priority, "priority", "new priority")
	fs.Var(&typ, "type", "new type")
	fs.Var(&owner, "owner", "new owner")
	fs.Var(&addLabels, "add-labels", "labels to add")
	fs.Var(&removeLabels, "remove-labels", "labels to remove")

	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "id is required")
		return 2
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	id, err := github.NormalizeID(project, fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %v\n", err)
		return 4
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	t, err := store.Read(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read tick: %v\n", err)
		return 4
	}

	if title.set {
		t.Title = title.value
	}
	if description.set {
		t.Description = description.value
	}
	if notes.set {
		t.Notes = notes.value
	}
	if status.set {
		t.Status = status.value
		if status.value == tick.StatusClosed {
			now := time.Now().UTC()
			t.ClosedAt = &now
		} else {
			t.ClosedAt = nil
			t.ClosedReason = ""
		}
	}
	if priority.set {
		t.Priority = priority.value
	}
	if typ.set {
		t.Type = typ.value
	}
	if owner.set {
		t.Owner = owner.value
	}
	if addLabels.set {
		for _, label := range splitCSV(addLabels.value) {
			t.Labels = appendUnique(t.Labels, label)
		}
	}
	if removeLabels.set {
		for _, label := range splitCSV(removeLabels.value) {
			t.Labels = removeString(t.Labels, label)
		}
	}

	t.UpdatedAt = time.Now().UTC()
	if err := store.Write(t); err != nil {
		fmt.Fprintf(os.Stderr, "failed to update tick: %v\n", err)
		return 6
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(t); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode json: %v\n", err)
			return 6
		}
	}
	return 0
}

func runClose(args []string) int {
	fs := flag.NewFlagSet("close", flag.ContinueOnError)
	reason := fs.String("reason", "", "close reason")
	jsonOutput := fs.Bool("json", false, "output as json")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "id is required")
		return 2
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	id, err := github.NormalizeID(project, fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %v\n", err)
		return 4
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	t, err := store.Read(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read tick: %v\n", err)
		return 4
	}

	now := time.Now().UTC()
	t.Status = tick.StatusClosed
	t.ClosedAt = &now
	t.ClosedReason = strings.TrimSpace(*reason)
	t.UpdatedAt = now

	if err := store.Write(t); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close tick: %v\n", err)
		return 6
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(t); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode json: %v\n", err)
			return 6
		}
	}
	return 0
}

func runReopen(args []string) int {
	fs := flag.NewFlagSet("reopen", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "output as json")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "id is required")
		return 2
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repo root: %v\n", err)
		return 3
	}
	project, err := github.DetectProject(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect project: %v\n", err)
		return 5
	}
	id, err := github.NormalizeID(project, fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %v\n", err)
		return 4
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	t, err := store.Read(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read tick: %v\n", err)
		return 4
	}

	t.Status = tick.StatusOpen
	t.ClosedAt = nil
	t.ClosedReason = ""
	t.UpdatedAt = time.Now().UTC()

	if err := store.Write(t); err != nil {
		fmt.Fprintf(os.Stderr, "failed to reopen tick: %v\n", err)
		return 6
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(t); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode json: %v\n", err)
			return 6
		}
	}
	return 0
}

func splitCSV(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func appendUnique(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append(values, value)
}

func removeString(values []string, value string) []string {
	out := values[:0]
	for _, item := range values {
		if item == value {
			continue
		}
		out = append(out, item)
	}
	return out
}

type optionalString struct {
	value string
	set   bool
}

func (o *optionalString) String() string {
	return o.value
}

func (o *optionalString) Set(value string) error {
	o.value = value
	o.set = true
	return nil
}

type optionalInt struct {
	value int
	set   bool
}

func (o *optionalInt) String() string {
	return fmt.Sprintf("%d", o.value)
}

func (o *optionalInt) Set(value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	o.value = parsed
	o.set = true
	return nil
}

func formatTime(value time.Time) string {
	return value.Local().Format("2006-01-02 15:04")
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytesTrimSpace(out)), nil
}

func bytesTrimSpace(in []byte) []byte {
	start := 0
	for start < len(in) && (in[start] == ' ' || in[start] == '\n' || in[start] == '\t' || in[start] == '\r') {
		start++
	}
	end := len(in)
	for end > start && (in[end-1] == ' ' || in[end-1] == '\n' || in[end-1] == '\t' || in[end-1] == '\r') {
		end--
	}
	return in[start:end]
}

func printUsage() {
	fmt.Println("Usage: tk <command> [--help]")
	fmt.Println("Commands: init, whoami, show, create, block, unblock, update, close, reopen")
}
