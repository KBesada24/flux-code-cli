# Phase 4: Git Integration

**Timeline:** Week 4  
**Goal:** Integrate go-git for repository-aware context, implement git commands for developer workflows

---

## Overview

This phase adds deep git integration using go-git. Developers can use commands like `/diff`, `/staged`, `/blame` to provide git context to the AI. The AI becomes aware of what changes the developer is working on.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Repository detection | Detect if in a git repo | P0 |
| `/diff` command | Show unstaged changes | P0 |
| `/staged` command | Show staged changes | P0 |
| `/log` command | Show recent commits | P0 |
| `/blame` command | Show file blame info | P1 |
| `/commit` command | Generate commit message | P1 |
| `/branch` command | Show branch info | P2 |
| Git status in status bar | Show branch + dirty state | P1 |

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/git/repo.go` | Repository operations wrapper |
| `internal/git/diff.go` | Diff generation and parsing |
| `internal/git/blame.go` | Blame information extraction |
| `internal/git/status.go` | Status and branch info |
| `internal/commands/handler.go` | Slash command parser |
| `internal/commands/git.go` | Git command implementations |

### Modified Files

| File | Changes |
|------|---------|
| `internal/ui/model.go` | Add command handling, git status |
| `internal/ui/components/statusbar.go` | Show git branch |

---

## Detailed File Specifications

### 1. `internal/git/repo.go`

```go
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
```

---

### 2. `internal/git/diff.go`

```go
package git

import (
    "fmt"
    "strings"
    
    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

// DiffOptions configures diff generation
type DiffOptions struct {
    Staged   bool   // Show staged changes only
    File     string // Specific file (empty = all)
    Context  int    // Lines of context (default 3)
}

// GetDiff returns the diff as a string
func (r *Repo) GetDiff(opts DiffOptions) (string, error) {
    status, err := r.worktree.Status()
    if err != nil {
        return "", err
    }
    
    if status.IsClean() {
        return "No changes detected.", nil
    }
    
    var builder strings.Builder
    
    // Get HEAD commit for comparison
    head, err := r.repo.Head()
    if err != nil {
        return "", err
    }
    
    headCommit, err := r.repo.CommitObject(head.Hash())
    if err != nil {
        return "", err
    }
    
    headTree, err := headCommit.Tree()
    if err != nil {
        return "", err
    }
    
    // Build diff output
    for file, fileStatus := range status {
        if opts.File != "" && file != opts.File {
            continue
        }
        
        // Filter based on staged/unstaged
        if opts.Staged && fileStatus.Staging == gogit.Unmodified {
            continue
        }
        if !opts.Staged && fileStatus.Worktree == gogit.Unmodified {
            continue
        }
        
        statusChar := getStatusChar(fileStatus, opts.Staged)
        builder.WriteString(fmt.Sprintf("%s %s\n", statusChar, file))
    }
    
    // For actual diff content, we need to compare trees
    if opts.Staged {
        // Compare index to HEAD
        builder.WriteString("\n--- Staged Changes ---\n")
        // Generate patch...
    } else {
        // Compare worktree to index
        builder.WriteString("\n--- Unstaged Changes ---\n")
        // Generate patch...
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

func getStatusChar(s gogit.FileStatus, staged bool) string {
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

// GetUnifiedDiff generates a unified diff for a specific file
func (r *Repo) GetUnifiedDiff(file string, staged bool) (string, error) {
    // Get file contents from HEAD
    head, err := r.repo.Head()
    if err != nil {
        return "", err
    }
    
    commit, err := r.repo.CommitObject(head.Hash())
    if err != nil {
        return "", err
    }
    
    tree, err := commit.Tree()
    if err != nil {
        return "", err
    }
    
    // Get the file from HEAD
    headFile, err := tree.File(file)
    var headContent string
    if err == nil {
        headContent, _ = headFile.Contents()
    }
    
    // Get current file content
    currentContent, err := r.readFile(file)
    if err != nil && headContent == "" {
        return "", fmt.Errorf("file not found: %s", file)
    }
    
    // Generate unified diff
    return generateUnifiedDiff(file, headContent, currentContent), nil
}

func (r *Repo) readFile(file string) (string, error) {
    path := filepath.Join(r.path, file)
    content, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return string(content), nil
}

func generateUnifiedDiff(filename, old, new string) string {
    // Simple diff implementation
    // For production, use a proper diff library
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("--- a/%s\n", filename))
    builder.WriteString(fmt.Sprintf("+++ b/%s\n", filename))
    
    oldLines := strings.Split(old, "\n")
    newLines := strings.Split(new, "\n")
    
    // Basic line-by-line comparison
    // In production, use Myers diff algorithm
    builder.WriteString("@@ -1 +1 @@\n")
    
    for _, line := range oldLines {
        if !contains(newLines, line) {
            builder.WriteString("-" + line + "\n")
        }
    }
    for _, line := range newLines {
        if !contains(oldLines, line) {
            builder.WriteString("+" + line + "\n")
        }
    }
    
    return builder.String()
}
```

---

### 3. `internal/git/blame.go`

```go
package git

import (
    "fmt"
    "strings"
    
    "github.com/go-git/go-git/v5/plumbing/object"
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
```

---

### 4. `internal/git/status.go`

```go
package git

import (
    "fmt"
    "strings"
    
    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

// Status represents the repository status
type Status struct {
    Branch   string
    Dirty    bool
    Ahead    int
    Behind   int
    Staged   []string
    Modified []string
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
```

---

### 5. `internal/commands/handler.go`

```go
package commands

import (
    "strings"
)

// Command represents a parsed slash command
type Command struct {
    Name string
    Args []string
    Raw  string
}

// Parse parses a slash command from input
func Parse(input string) *Command {
    input = strings.TrimSpace(input)
    
    if !strings.HasPrefix(input, "/") {
        return nil
    }
    
    parts := strings.Fields(input)
    if len(parts) == 0 {
        return nil
    }
    
    name := strings.TrimPrefix(parts[0], "/")
    args := parts[1:]
    
    return &Command{
        Name: strings.ToLower(name),
        Args: args,
        Raw:  input,
    }
}

// IsCommand returns true if the input starts with /
func IsCommand(input string) bool {
    return strings.HasPrefix(strings.TrimSpace(input), "/")
}

// CommandResult represents the result of a command execution
type CommandResult struct {
    Output    string
    AddToChat bool   // If true, add to chat as context
    Error     error
}
```

---

### 6. `internal/commands/git.go`

```go
package commands

import (
    "fmt"
    "strconv"
    "strings"
    
    "github.com/yourusername/flux/internal/git"
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
```

---

### 7. Updated Status Bar with Git Info

```go
// internal/ui/components/statusbar.go

package components

import (
    "fmt"
    
    "github.com/charmbracelet/lipgloss"
    "github.com/yourusername/flux/internal/git"
)

type StatusBar struct {
    width     int
    gitStatus string
    model     string
    provider  string
}

func NewStatusBar() StatusBar {
    return StatusBar{}
}

func (s *StatusBar) Update() {
    // Update git status
    if repo, err := git.Open(""); err == nil {
        if status, err := repo.GetStatus(); err == nil {
            s.gitStatus = status.FormatForStatusBar()
        }
    }
}

func (s StatusBar) View() string {
    leftStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#626262"))
    
    gitStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00D4AA")).
        Bold(true)
    
    modelStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#7D56F4"))
    
    left := leftStyle.Render("Ctrl+C quit • Enter send • /help commands")
    
    var right string
    if s.gitStatus != "" {
        right = gitStyle.Render(" "+s.gitStatus) + " │ "
    }
    right += modelStyle.Render(s.model)
    
    // Calculate padding
    padding := s.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
    if padding < 0 {
        padding = 0
    }
    
    return left + strings.Repeat(" ", padding) + right
}

func (s *StatusBar) SetWidth(w int) {
    s.width = w
}

func (s *StatusBar) SetModel(provider, model string) {
    s.provider = provider
    s.model = fmt.Sprintf("%s/%s", provider, model)
}
```

---

## Testing

### Unit Tests

| Test File | Tests |
|-----------|-------|
| `internal/git/repo_test.go` | Open, IsRepo, CurrentBranch |
| `internal/git/diff_test.go` | GetDiff, GetDiffStats |
| `internal/git/blame_test.go` | Blame, BlameRange |
| `internal/git/status_test.go` | GetStatus, GetLog |
| `internal/commands/handler_test.go` | Parse, IsCommand |
| `internal/commands/git_test.go` | All git command handlers |

### Test Repository Setup

```bash
# Create test repo for unit tests
func setupTestRepo(t *testing.T) string {
    dir := t.TempDir()
    
    repo, _ := gogit.PlainInit(dir, false)
    w, _ := repo.Worktree()
    
    // Create initial file
    ioutil.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)
    w.Add("test.txt")
    w.Commit("Initial commit", &gogit.CommitOptions{
        Author: &object.Signature{Name: "Test", Email: "test@test.com"},
    })
    
    return dir
}
```

### Manual Testing Checklist

- [ ] `/diff` shows unstaged changes
- [ ] `/staged` shows staged changes
- [ ] `/log` shows recent commits
- [ ] `/log 5` shows exactly 5 commits
- [ ] `/blame <file>` shows blame info
- [ ] `/branch` shows current branch and status
- [ ] `/status` shows full git status
- [ ] `/commit` generates AI commit message from staged changes
- [ ] Status bar shows branch name
- [ ] Status bar shows dirty indicator (*)
- [ ] Commands work in subdirectories of repo
- [ ] Error message when not in git repo

---

## Dependencies

```bash
go get github.com/go-git/go-git/v5
```

---

## Acceptance Criteria

1. **Detection:** Correctly detects git repositories
2. **Diff:** Shows accurate diff output
3. **Staged:** Distinguishes staged from unstaged
4. **Log:** Shows commit history correctly
5. **Blame:** Accurate line-by-line blame
6. **Status bar:** Shows branch + dirty state
7. **Commands:** All slash commands work

---

## Definition of Done

- [ ] All git wrapper functions implemented
- [ ] All slash commands working
- [ ] Status bar shows git info
- [ ] Unit tests for git operations
- [ ] Integration tests with test repo
- [ ] Manual testing completed
- [ ] Error handling for non-git directories
- [ ] Committed to version control

---

## Notes

- go-git is pure Go, no external git binary needed
- Blame can be slow on large files - consider caching
- Status bar updates on each render - may need debouncing
- Consider lazy loading git status for performance
