package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pengelbrecht/ticks/internal/config"
	"github.com/pengelbrecht/ticks/internal/github"
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
	fmt.Println("Commands: init, whoami")
}
