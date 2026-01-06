package git

import (
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// BlameResult contains blame information for a file
type BlameResult struct {
	Lines []BlameLine
}

// BlameLine represents a single line's blame info
type BlameLine struct {
	LineNumber int
	Hash       string
	Author     string
	Date       string
	Content    string
}

// Blame returns blame information for a file
func (r *Repo) Blame(file string) (*BlameResult, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	blame, err := gogit.Blame(commit, file)
	if err != nil {
		return nil, fmt.Errorf("failed to blame %s: %w", file, err)
	}

	result := &BlameResult{
		Lines: make([]BlameLine, len(blame.Lines)),
	}

	for i, line := range blame.Lines {
		result.Lines[i] = BlameLine{
			LineNumber: i + 1,
			Hash:       line.Hash.String()[:7],
			Author:     line.Author,
			Date:       line.Date.Format("2006-01-02"),
			Content:    line.Text,
		}
	}

	return result, nil
}

// BlameRange returns blame for specific line range
func (r *Repo) BlameRange(file string, startLine, endLine int) (*BlameResult, error) {
	full, err := r.Blame(file)
	if err != nil {
		return nil, err
	}

	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(full.Lines) {
		endLine = len(full.Lines)
	}

	return &BlameResult{
		Lines: full.Lines[startLine-1 : endLine],
	}, nil
}

// FormatBlame formats blame output for display
func (b *BlameResult) Format() string {
	var builder strings.Builder

	for _, line := range b.Lines {
		builder.WriteString(fmt.Sprintf(
			"%4d │ %s │ %-12s │ %s │ %s\n",
			line.LineNumber,
			line.Hash,
			truncate(line.Author, 12),
			line.Date,
			line.Content,
		))
	}

	return builder.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s + strings.Repeat(" ", max-len(s))
	}
	return s[:max-1] + "…"
}
