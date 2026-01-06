package git

import (
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// DiffOptions configures diff generation
type DiffOptions struct {
	Staged  bool   // Show staged changes only
	File    string // Specific file (empty = all)
	Context int    // Lines of context (default 3)
}

// GetDiff returns the diff as a string
func (r *Repo) GetDiff(opts DiffOptions) (string, error) {
	// For now, we'll use a simplified approach since generating full unified diffs
	// manually with go-git is complex. We'll list changed files and their status.
	// In a full implementation, we would iterate through patches.

	status, err := r.worktree.Status()
	if err != nil {
		return "", err
	}

	if status.IsClean() {
		return "No changes detected.", nil
	}

	var builder strings.Builder

	// Build diff output
	for file, fileStatus := range status {
		if opts.File != "" && file != opts.File {
			continue
		}

		// Filter based on staged/unstaged
		// Status codes: ' ' (Unmodified), 'M' (Modified), 'A' (Added), 'D' (Deleted), etc.
		// Staging is the first char, Worktree is the second.

		if opts.Staged {
			if fileStatus.Staging == gogit.Unmodified && fileStatus.Staging != gogit.Untracked {
				continue
			}
		} else {
			if fileStatus.Worktree == gogit.Unmodified {
				continue
			}
		}

		statusChar := getStatusChar(fileStatus, opts.Staged)
		builder.WriteString(fmt.Sprintf("%s %s\n", statusChar, file))
	}

	return builder.String(), nil
}

// GetDiffStats returns summary statistics
func (r *Repo) GetDiffStats(staged bool) (*DiffStats, error) {
	status, err := r.worktree.Status()
	if err != nil {
		return nil, err
	}

	stats := &DiffStats{}

	for _, s := range status {
		if staged {
			switch s.Staging {
			case gogit.Added:
				stats.Added++
			case gogit.Modified:
				stats.Modified++
			case gogit.Deleted:
				stats.Deleted++
			}
		} else {
			switch s.Worktree {
			case gogit.Added, gogit.Untracked:
				stats.Added++
			case gogit.Modified:
				stats.Modified++
			case gogit.Deleted:
				stats.Deleted++
			}
		}
	}

	return stats, nil
}

type DiffStats struct {
	Added    int
	Modified int
	Deleted  int
}

func (d DiffStats) String() string {
	return fmt.Sprintf("+%d ~%d -%d", d.Added, d.Modified, d.Deleted)
}

func getStatusChar(s *gogit.FileStatus, staged bool) string {
	var code gogit.StatusCode
	if staged {
		code = s.Staging
	} else {
		code = s.Worktree
	}

	switch code {
	case gogit.Added:
		return "A"
	case gogit.Modified:
		return "M"
	case gogit.Deleted:
		return "D"
	case gogit.Renamed:
		return "R"
	case gogit.Copied:
		return "C"
	case gogit.Untracked:
		return "?"
	default:
		return " "
	}
}
