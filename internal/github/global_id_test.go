package github

import "testing"

func TestNormalizeID(t *testing.T) {
	project := "petere/chefswiz"

	id, err := NormalizeID(project, "a1b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "a1b" {
		t.Fatalf("expected a1b, got %s", id)
	}

	id, err = NormalizeID(project, "petere/chefswiz:a1b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "a1b" {
		t.Fatalf("expected a1b, got %s", id)
	}

	_, err = NormalizeID(project, "someoneelse/repo:a1b")
	if err == nil {
		t.Fatalf("expected mismatch error")
	}
}
