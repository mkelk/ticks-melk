package swarm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Runner executes an epic using Claude Code's native Task tool for parallel subagent orchestration.
// Instead of the ralph iteration loop, it spawns a single claude process that orchestrates
// subagents to work on tasks in parallel waves.
type Runner struct {
	// Command is the path to the claude binary. Defaults to "claude".
	Command string

	// MaxAgents is the maximum number of parallel subagents per wave.
	MaxAgents int

	// Callbacks for status updates
	OnOutput func(chunk string)
	OnStart  func(epicID string)
	OnEnd    func(epicID string, result *Result)
}

// Result contains the outcome of a swarm run.
type Result struct {
	EpicID   string
	Duration time.Duration
	Success  bool
	Error    error
	Output   string
}

// NewRunner creates a new swarm runner with default settings.
func NewRunner(maxAgents int) *Runner {
	return &Runner{
		Command:   "claude",
		MaxAgents: maxAgents,
	}
}

// Available checks if the claude CLI is installed and accessible.
func (r *Runner) Available() bool {
	_, err := exec.LookPath(r.command())
	return err == nil
}

// Run executes the epic using swarm orchestration.
// workDir is the directory to run in (worktree path or empty for cwd).
func (r *Runner) Run(ctx context.Context, epicID string, workDir string) (*Result, error) {
	start := time.Now()

	if r.OnStart != nil {
		r.OnStart(epicID)
	}

	result := &Result{
		EpicID: epicID,
	}

	// Build the orchestration prompt
	prompt := r.buildPrompt(epicID)

	// Build command arguments
	args := []string{
		"--dangerously-skip-permissions",
		"--print",
		"--verbose",
		prompt,
	}

	cmd := exec.CommandContext(ctx, r.command(), args...)

	// Set working directory if specified
	if workDir != "" {
		cmd.Dir = workDir
	}

	// Set environment
	cmd.Env = append(os.Environ(), "TICK_OWNER=swarm")

	var stdout, stderr bytes.Buffer

	// Stream output if callback is set
	if r.OnOutput != nil {
		cmd.Stdout = &streamWriter{
			buf:      &stdout,
			callback: r.OnOutput,
		}
	} else {
		cmd.Stdout = &stdout
	}
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	result.Duration = time.Since(start)
	result.Output = stdout.String()

	if err != nil {
		if ctx.Err() == context.Canceled {
			result.Error = fmt.Errorf("swarm cancelled")
			result.Success = false
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("swarm timed out")
			result.Success = false
		} else {
			result.Error = fmt.Errorf("swarm failed: %w\nstderr: %s", err, stderr.String())
			result.Success = false
		}
	} else {
		result.Success = true
	}

	if r.OnEnd != nil {
		r.OnEnd(epicID, result)
	}

	return result, nil
}

// buildPrompt creates the orchestration prompt for the swarm agent.
func (r *Runner) buildPrompt(epicID string) string {
	return fmt.Sprintf(`You are orchestrating an epic using Claude Code's Task tool for parallel execution.

## Epic: %s

## Instructions

Follow these steps exactly:

### Step 1: Get the dependency graph
Run this command and parse the JSON output:
`+"```"+`bash
tk graph %s --json
`+"```"+`

The output contains:
- "waves": Groups of tasks that can run in parallel
- "max_parallel": Maximum concurrent tasks possible
- Each task has: id, title, description, acceptance, blocked_by

### Step 2: Execute waves sequentially

For each wave in the graph:

1. **Launch subagents** - For each task in the wave (up to %d parallel):
   - Use the Task tool with these parameters:
     - subagent_type: "general-purpose"
     - name: "%s-w<wave_number>-<tick_id>"
     - run_in_background: true
     - mode: "bypassPermissions"
     - prompt: Include the full task details (see template below)

2. **Poll for completion** - Use TaskOutput with block=false to check each agent:
   - Poll every 10 seconds
   - Continue until all agents in the wave complete

3. **Sync results** - For each completed task:
   `+"```"+`bash
   tk close <tick-id> --reason "Completed via swarm"
   `+"```"+`

4. **Proceed to next wave** - Only after all tasks in current wave are done

### Step 3: Finalize

After all waves complete:

1. **Commit any uncommitted work**:
   `+"```"+`bash
   git status
   # If there are changes:
   git add -A
   git commit -m "feat(%s): implement epic tasks via swarm"
   `+"```"+`

2. **Close the epic**:
   `+"```"+`bash
   tk close %s --reason "All tasks completed via swarm"
   `+"```"+`

## Task Agent Prompt Template

For each task, use this prompt structure:

`+"```"+`
You are implementing a task from the Ticks issue tracker.

## Task
Title: <tick-title>
ID: <tick-id>
Epic: %s

## Description
<tick-description>

## Acceptance Criteria
<tick-acceptance>

## Instructions
1. Read and understand the existing codebase relevant to this task
2. Implement the feature/fix as described
3. Write tests that verify the acceptance criteria
4. Ensure all tests pass before completing

## Completion
When done, provide a summary of:
- Files changed
- Tests added/modified
- Any notes for the next task

Do NOT run 'tk close' - the orchestrator will handle tick state.
`+"```"+`

## Error Handling

- If an agent fails, log the error and continue with remaining agents
- Add a note to failed tasks: `+"`tk note <id> \"Agent failed: <reason>\"`"+`
- Report all failures at the end

## Important Notes

- Tasks within a wave are independent and can run in parallel
- Tasks in later waves may depend on earlier waves completing
- Always wait for a wave to fully complete before starting the next
- The Task tool will return an agent ID - use this with TaskOutput to poll

Begin by running 'tk graph %s --json' to get the dependency structure.
`, epicID, epicID, r.MaxAgents, epicID, epicID, epicID, epicID, epicID)
}

// command returns the claude binary path.
func (r *Runner) command() string {
	if r.Command != "" {
		return r.Command
	}
	return "claude"
}

// streamWriter wraps a buffer and calls a callback for each write.
type streamWriter struct {
	buf      *bytes.Buffer
	callback func(string)
}

func (w *streamWriter) Write(p []byte) (n int, err error) {
	n, err = w.buf.Write(p)
	if w.callback != nil && n > 0 {
		w.callback(string(p[:n]))
	}
	return
}
