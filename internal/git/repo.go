package git

import (
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repo wraps go-git repository operations
type Repo struct {
	repo     *gogit.Repository
	worktree *gogit.Worktree
	path     string
}

// Open opens the git repository at the given path or current directory
func Open(path string) (*Repo, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Find repository root
	root, err := findRepoRoot(path)
	if err != nil {
		return nil, err
	}

	repo, err := gogit.PlainOpen(root)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	return &Repo{
		repo:     repo,
		worktree: worktree,
		path:     root,
	}, nil
}

// findRepoRoot walks up the directory tree to find .git
func findRepoRoot(path string) (string, error) {
	for {
		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return path, nil
		}

		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("not a git repository")
		}
		path = parent
	}
}

// IsRepo returns true if we're in a git repository
func IsRepo() bool {
	_, err := Open("")
	return err == nil
}

// Path returns the repository root path
func (r *Repo) Path() string {
	return r.path
}

// Head returns the current HEAD reference
func (r *Repo) Head() (*plumbing.Reference, error) {
	return r.repo.Head()
}

// CurrentBranch returns the current branch name
func (r *Repo) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", err
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	// Detached HEAD - return short hash
	return head.Hash().String()[:7], nil
}

// IsDirty returns true if there are uncommitted changes
func (r *Repo) IsDirty() (bool, error) {
	status, err := r.worktree.Status()
	if err != nil {
		return false, err
	}
	return !status.IsClean(), nil
}
