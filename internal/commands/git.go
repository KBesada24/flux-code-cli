package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kbesada/flux-code-cli/internal/git"
)

// ExecuteGitCommand handles git-related slash commands
func ExecuteGitCommand(cmd *Command) CommandResult {
	repo, err := git.Open("")
	if err != nil {
		return CommandResult{
			Error: fmt.Errorf("not in a git repository: %w", err),
		}
	}

	switch cmd.Name {
	case "diff":
		return executeDiff(repo, cmd.Args)
	case "staged":
		return executeStaged(repo)
	case "log":
		return executeLog(repo, cmd.Args)
	case "blame":
		return executeBlame(repo, cmd.Args)
	case "branch":
		return executeBranch(repo)
	case "status":
		return executeStatus(repo)
	case "commit":
		return executeCommitMsg(repo)
	default:
		return CommandResult{
			Error: fmt.Errorf("unknown command: /%s", cmd.Name),
		}
	}
}

func executeDiff(repo *git.Repo, args []string) CommandResult {
	opts := git.DiffOptions{Staged: false}

	if len(args) > 0 {
		opts.File = args[0]
	}

	diff, err := repo.GetDiff(opts)
	if err != nil {
		return CommandResult{Error: err}
	}

	return CommandResult{
		Output:    formatDiffForContext(diff),
		AddToChat: true,
	}
}

func executeStaged(repo *git.Repo) CommandResult {
	diff, err := repo.GetDiff(git.DiffOptions{Staged: true})
	if err != nil {
		return CommandResult{Error: err}
	}

	return CommandResult{
		Output:    formatDiffForContext(diff),
		AddToChat: true,
	}
}

func executeLog(repo *git.Repo, args []string) CommandResult {
	n := 10 // default
	if len(args) > 0 {
		if parsed, err := strconv.Atoi(args[0]); err == nil {
			n = parsed
		}
	}

	commits, err := repo.GetLog(n)
	if err != nil {
		return CommandResult{Error: err}
	}

	var builder strings.Builder
	builder.WriteString("## Recent Commits\n\n")
	for _, c := range commits {
		builder.WriteString(fmt.Sprintf("- `%s` %s (%s)\n", c.Hash, c.Message, c.Author))
	}

	return CommandResult{
		Output:    builder.String(),
		AddToChat: true,
	}
}

func executeBlame(repo *git.Repo, args []string) CommandResult {
	if len(args) == 0 {
		return CommandResult{
			Error: fmt.Errorf("usage: /blame <file> [start-line] [end-line]"),
		}
	}

	file := args[0]

	var result *git.BlameResult
	var err error

	if len(args) >= 3 {
		start, _ := strconv.Atoi(args[1])
		end, _ := strconv.Atoi(args[2])
		result, err = repo.BlameRange(file, start, end)
	} else {
		result, err = repo.Blame(file)
	}

	if err != nil {
		return CommandResult{Error: err}
	}

	return CommandResult{
		Output:    result.Format(),
		AddToChat: true,
	}
}

func executeBranch(repo *git.Repo) CommandResult {
	branch, err := repo.CurrentBranch()
	if err != nil {
		return CommandResult{Error: err}
	}

	status, err := repo.GetStatus()
	if err != nil {
		return CommandResult{Error: err}
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("## Current Branch: %s\n\n", branch))

	if status.Dirty {
		builder.WriteString("Status: **dirty** (uncommitted changes)\n\n")
		if len(status.Staged) > 0 {
			builder.WriteString(fmt.Sprintf("- %d staged files\n", len(status.Staged)))
		}
		if len(status.Modified) > 0 {
			builder.WriteString(fmt.Sprintf("- %d modified files\n", len(status.Modified)))
		}
		if len(status.Untracked) > 0 {
			builder.WriteString(fmt.Sprintf("- %d untracked files\n", len(status.Untracked)))
		}
	} else {
		builder.WriteString("Status: **clean**\n")
	}

	return CommandResult{
		Output:    builder.String(),
		AddToChat: true,
	}
}

func executeStatus(repo *git.Repo) CommandResult {
	status, err := repo.GetStatus()
	if err != nil {
		return CommandResult{Error: err}
	}

	var builder strings.Builder
	builder.WriteString("## Git Status\n\n")
	builder.WriteString(fmt.Sprintf("Branch: %s\n\n", status.Branch))

	if len(status.Staged) > 0 {
		builder.WriteString("### Staged\n")
		for _, f := range status.Staged {
			builder.WriteString(fmt.Sprintf("- %s\n", f))
		}
		builder.WriteString("\n")
	}

	if len(status.Modified) > 0 {
		builder.WriteString("### Modified\n")
		for _, f := range status.Modified {
			builder.WriteString(fmt.Sprintf("- %s\n", f))
		}
		builder.WriteString("\n")
	}

	if len(status.Untracked) > 0 {
		builder.WriteString("### Untracked\n")
		for _, f := range status.Untracked {
			builder.WriteString(fmt.Sprintf("- %s\n", f))
		}
	}

	return CommandResult{
		Output:    builder.String(),
		AddToChat: true,
	}
}

func executeCommitMsg(repo *git.Repo) CommandResult {
	diff, err := repo.GetDiff(git.DiffOptions{Staged: true})
	if err != nil {
		return CommandResult{Error: err}
	}

	if diff == "No changes detected." {
		return CommandResult{
			Output: "No staged changes. Stage changes with `git add` first.",
		}
	}

	// Return diff with instruction for AI to generate commit message
	prompt := fmt.Sprintf(`Based on the following staged changes, generate a concise and descriptive commit message following conventional commits format (e.g., feat:, fix:, docs:, refactor:).

%s

Generate only the commit message, nothing else.`, diff)

	return CommandResult{
		Output:    prompt,
		AddToChat: true,
	}
}

func formatDiffForContext(diff string) string {
	return fmt.Sprintf("## Git Diff\n\n```diff\n%s\n```", diff)
}
