package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func setupTestRepo(t *testing.T) string {
	dir := t.TempDir()

	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create initial file
	filePath := filepath.Join(dir, "test.txt")
	err = os.WriteFile(filePath, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = w.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = w.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	return dir
}

func TestRepo_CurrentBranch(t *testing.T) {
	dir := setupTestRepo(t)

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get branch: %v", err)
	}

	if branch != "master" && branch != "main" {
		t.Errorf("expected branch master or main, got %s", branch)
	}
}

func TestRepo_IsDirty(t *testing.T) {
	dir := setupTestRepo(t)

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	// Should be clean initially
	dirty, err := repo.IsDirty()
	if err != nil {
		t.Fatalf("failed to check dirty: %v", err)
	}
	if dirty {
		t.Error("expected clean repo")
	}

	// Modify file
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("modified"), 0644)

	// Should be dirty now
	dirty, err = repo.IsDirty()
	if err != nil {
		t.Fatalf("failed to check dirty: %v", err)
	}
	if !dirty {
		t.Error("expected dirty repo")
	}
}
