package github

import (
	"errors"
	"os"
	"testing"
)

func TestDetectOwnerEnv(t *testing.T) {
	if err := os.Setenv("TICK_OWNER", "alice"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("TICK_OWNER")

	owner, err := DetectOwner(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "alice" {
		t.Fatalf("expected owner alice, got %s", owner)
	}
}

func TestDetectOwnerGh(t *testing.T) {
	os.Unsetenv("TICK_OWNER")

	owner, err := DetectOwner(func(string, ...string) ([]byte, error) {
		return []byte("bob\n"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "bob" {
		t.Fatalf("expected owner bob, got %s", owner)
	}
}

func TestDetectOwnerGhError(t *testing.T) {
	os.Unsetenv("TICK_OWNER")

	_, err := DetectOwner(func(string, ...string) ([]byte, error) {
		return nil, errors.New("missing gh")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
