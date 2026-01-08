package github

import "os/exec"

// CommandRunner executes a command and returns stdout.
type CommandRunner func(name string, args ...string) ([]byte, error)

func defaultRunner(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}
