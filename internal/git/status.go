package git

import (
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Status represents the repository status
type Status struct {
	Branch    string
	Dirty     bool
	Staged    []string
	Modified  []string
	Untracked []string
}

// GetStatus returns the current repository status
func (r *Repo) GetStatus() (*Status, error) {
	branch, err := r.CurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	status, err := r.worktree.Status()
	if err != nil {
		return nil, err
	}

	result := &Status{
		Branch: branch,
		Dirty:  !status.IsClean(),
	}

	for file, s := range status {
		if s.Staging != gogit.Unmodified && s.Staging != gogit.Untracked {
			result.Staged = append(result.Staged, file)
		}
		if s.Worktree == gogit.Modified {
			result.Modified = append(result.Modified, file)
		}
		if s.Worktree == gogit.Untracked {
			result.Untracked = append(result.Untracked, file)
		}
	}

	return result, nil
}

// FormatForStatusBar returns a short status for the status bar
func (s *Status) FormatForStatusBar() string {
	var parts []string

	parts = append(parts, s.Branch)

	if s.Dirty {
		parts = append(parts, "*")
	}

	if len(s.Staged) > 0 {
		parts = append(parts, fmt.Sprintf("+%d", len(s.Staged)))
	}

	return strings.Join(parts, "")
}

// GetLog returns recent commits
func (r *Repo) GetLog(n int) ([]CommitInfo, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	iter, err := r.repo.Log(&gogit.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	count := 0

	err = iter.ForEach(func(c *object.Commit) error {
		if count >= n {
			return fmt.Errorf("done") // Stop iteration
		}

		commits = append(commits, CommitInfo{
			Hash:    c.Hash.String()[:7],
			Author:  c.Author.Name,
			Date:    c.Author.When.Format("2006-01-02 15:04"),
			Message: strings.Split(c.Message, "\n")[0],
		})
		count++
		return nil
	})

	if err != nil && err.Error() != "done" {
		return nil, err
	}

	return commits, nil
}

type CommitInfo struct {
	Hash    string
	Author  string
	Date    string
	Message string
}

func (c CommitInfo) Format() string {
	return fmt.Sprintf("%s %s (%s) %s", c.Hash, c.Message, c.Author, c.Date)
}
