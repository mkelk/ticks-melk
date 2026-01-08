package github

import (
	"fmt"
	"strings"
)

// NormalizeID accepts short or global IDs and returns the short ID.
func NormalizeID(project, input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("id is required")
	}

	parts := strings.SplitN(input, ":", 2)
	if len(parts) == 1 {
		return input, nil
	}

	if project == "" {
		return "", fmt.Errorf("project is required for global ids")
	}
	if parts[0] != project {
		return "", fmt.Errorf("global id project mismatch: %s", parts[0])
	}
	if parts[1] == "" {
		return "", fmt.Errorf("global id missing tick id")
	}
	return parts[1], nil
}
