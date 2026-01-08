package github

import (
	"fmt"
	"os"
	"strings"
)

// DetectOwner resolves owner via TICK_OWNER or gh.
func DetectOwner(run CommandRunner) (string, error) {
	if owner := strings.TrimSpace(os.Getenv("TICK_OWNER")); owner != "" {
		return owner, nil
	}

	if run == nil {
		run = defaultRunner
	}

	out, err := run("gh", "api", "user", "--jq", ".login")
	if err != nil {
		return "", fmt.Errorf("failed to resolve owner via gh: %w", err)
	}

	owner := strings.TrimSpace(string(out))
	if owner == "" {
		return "", fmt.Errorf("gh returned empty owner")
	}

	return owner, nil
}
