# Phase 7: Advanced Features

**Timeline:** Future (Post-MVP)  
**Goal:** Add advanced developer workflow features, code review mode, refactoring, test generation, and GitHub integration

---

## Overview

This phase adds advanced features that make Flux a complete AI coding assistant. These are post-MVP features that enhance the developer experience significantly.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| `/review` command | Code review mode for changes | P1 |
| `/refactor` command | Suggest refactoring | P1 |
| `/test` command | Generate test suggestions | P1 |
| GitHub integration | PR and issue context | P2 |
| Conversation persistence | Save/load sessions | P2 |
| Plugin system | Extend with custom commands | P3 |
| MCP integration | Model Context Protocol support | P2 |

---

## Files to Create

### New Files

| File | Purpose |
|------|---------|
| `internal/commands/review.go` | Code review implementation |
| `internal/commands/refactor.go` | Refactoring suggestions |
| `internal/commands/test.go` | Test generation |
| `internal/github/client.go` | GitHub API wrapper |
| `internal/github/pr.go` | Pull request operations |
| `internal/github/issues.go` | Issue operations |
| `internal/session/persistence.go` | Session save/load |
| `internal/plugins/loader.go` | Plugin system |
| `internal/mcp/client.go` | MCP client |

---

## Feature Specifications

### 1. Code Review Mode (`/review`)

Provides structured code review for git changes.

```go
// internal/commands/review.go

package commands

import (
    "fmt"
    "strings"
    
    "github.com/yourusername/flux/internal/git"
)

type ReviewMode string

const (
    ReviewStaged   ReviewMode = "staged"
    ReviewUnstaged ReviewMode = "unstaged"
    ReviewCommit   ReviewMode = "commit"
    ReviewPR       ReviewMode = "pr"
)

type ReviewOptions struct {
    Mode       ReviewMode
    Focus      []string // Areas to focus on: security, performance, style, bugs
    CommitHash string   // For commit review
    PRNumber   int      // For PR review
}

func ExecuteReview(args []string) CommandResult {
    opts := parseReviewArgs(args)
    
    repo, err := git.Open("")
    if err != nil {
        return CommandResult{Error: err}
    }
    
    var diff string
    switch opts.Mode {
    case ReviewStaged:
        diff, err = repo.GetDiff(git.DiffOptions{Staged: true})
    case ReviewUnstaged:
        diff, err = repo.GetDiff(git.DiffOptions{Staged: false})
    case ReviewCommit:
        diff, err = repo.GetCommitDiff(opts.CommitHash)
    }
    
    if err != nil {
        return CommandResult{Error: err}
    }
    
    prompt := buildReviewPrompt(diff, opts.Focus)
    
    return CommandResult{
        Output:    prompt,
        AddToChat: true,
    }
}

func buildReviewPrompt(diff string, focus []string) string {
    var builder strings.Builder
    
    builder.WriteString("Please review the following code changes:\n\n")
    builder.WriteString("```diff\n")
    builder.WriteString(diff)
    builder.WriteString("\n```\n\n")
    
    builder.WriteString("Provide a thorough code review covering:\n\n")
    
    if len(focus) == 0 {
        focus = []string{"bugs", "security", "performance", "style", "maintainability"}
    }
    
    for _, area := range focus {
        switch area {
        case "bugs":
            builder.WriteString("1. **Potential Bugs**: Logic errors, edge cases, null/nil handling\n")
        case "security":
            builder.WriteString("2. **Security Issues**: Injection vulnerabilities, auth issues, data exposure\n")
        case "performance":
            builder.WriteString("3. **Performance**: Inefficient algorithms, unnecessary allocations, N+1 queries\n")
        case "style":
            builder.WriteString("4. **Code Style**: Naming, formatting, idiomatic patterns\n")
        case "maintainability":
            builder.WriteString("5. **Maintainability**: Complexity, documentation, testability\n")
        }
    }
    
    builder.WriteString("\nFor each issue found, provide:\n")
    builder.WriteString("- The specific line/section\n")
    builder.WriteString("- Why it's a problem\n")
    builder.WriteString("- A suggested fix\n")
    
    return builder.String()
}

func parseReviewArgs(args []string) ReviewOptions {
    opts := ReviewOptions{
        Mode: ReviewStaged,
    }
    
    for i, arg := range args {
        switch arg {
        case "--staged", "-s":
            opts.Mode = ReviewStaged
        case "--unstaged", "-u":
            opts.Mode = ReviewUnstaged
        case "--commit", "-c":
            if i+1 < len(args) {
                opts.Mode = ReviewCommit
                opts.CommitHash = args[i+1]
            }
        case "--focus", "-f":
            if i+1 < len(args) {
                opts.Focus = strings.Split(args[i+1], ",")
            }
        }
    }
    
    return opts
}
```

**Usage:**
```
/review                      # Review staged changes
/review --unstaged           # Review unstaged changes
/review --commit abc123      # Review specific commit
/review --focus security,bugs # Focus on specific areas
```

---

### 2. Refactoring Suggestions (`/refactor`)

Analyzes code and suggests improvements.

```go
// internal/commands/refactor.go

package commands

import (
    "fmt"
    
    "github.com/yourusername/flux/internal/parser"
)

type RefactorType string

const (
    RefactorGeneral   RefactorType = "general"
    RefactorExtract   RefactorType = "extract"
    RefactorRename    RefactorType = "rename"
    RefactorSimplify  RefactorType = "simplify"
    RefactorModernize RefactorType = "modernize"
)

func ExecuteRefactor(args []string) CommandResult {
    if len(args) < 2 {
        return CommandResult{
            Error: fmt.Errorf("usage: /refactor <file> <symbol> [--type <type>]"),
        }
    }
    
    filename := args[0]
    symbolName := args[1]
    refactorType := parseRefactorType(args[2:])
    
    p := parser.NewParser()
    pf, err := p.ParseFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    symbol, err := parser.FindSymbol(pf, symbolName)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    prompt := buildRefactorPrompt(symbol, pf.Language.Name, refactorType)
    
    return CommandResult{
        Output:    prompt,
        AddToChat: true,
    }
}

func buildRefactorPrompt(symbol *parser.Symbol, lang string, refactorType RefactorType) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("Please suggest refactoring for this %s %s:\n\n",
        lang, symbol.Type))
    
    builder.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", lang, symbol.Content))
    
    switch refactorType {
    case RefactorExtract:
        builder.WriteString("Focus on:\n")
        builder.WriteString("- Extracting reusable functions/methods\n")
        builder.WriteString("- Breaking down complex logic\n")
        builder.WriteString("- Reducing function length\n")
        
    case RefactorSimplify:
        builder.WriteString("Focus on:\n")
        builder.WriteString("- Simplifying complex conditionals\n")
        builder.WriteString("- Removing redundant code\n")
        builder.WriteString("- Using simpler constructs\n")
        
    case RefactorModernize:
        builder.WriteString("Focus on:\n")
        builder.WriteString("- Using modern language features\n")
        builder.WriteString("- Updating deprecated patterns\n")
        builder.WriteString("- Following current best practices\n")
        
    default:
        builder.WriteString("Analyze and suggest improvements for:\n")
        builder.WriteString("- Code clarity and readability\n")
        builder.WriteString("- Reducing complexity\n")
        builder.WriteString("- Improving maintainability\n")
        builder.WriteString("- Performance optimizations\n")
        builder.WriteString("- Error handling\n")
    }
    
    builder.WriteString("\nProvide:\n")
    builder.WriteString("1. Specific issues with the current code\n")
    builder.WriteString("2. Refactored version with explanations\n")
    builder.WriteString("3. Benefits of the changes\n")
    
    return builder.String()
}

func parseRefactorType(args []string) RefactorType {
    for i, arg := range args {
        if (arg == "--type" || arg == "-t") && i+1 < len(args) {
            switch args[i+1] {
            case "extract":
                return RefactorExtract
            case "simplify":
                return RefactorSimplify
            case "modernize":
                return RefactorModernize
            }
        }
    }
    return RefactorGeneral
}
```

**Usage:**
```
/refactor main.go ProcessOrder              # General refactoring
/refactor main.go ProcessOrder --type extract  # Focus on extraction
/refactor main.go ProcessOrder --type simplify # Focus on simplification
```

---

### 3. Test Generation (`/test`)

Generates test suggestions for code.

```go
// internal/commands/test.go

package commands

import (
    "fmt"
    "path/filepath"
    "strings"
    
    "github.com/yourusername/flux/internal/parser"
)

func ExecuteTest(args []string) CommandResult {
    if len(args) < 1 {
        return CommandResult{
            Error: fmt.Errorf("usage: /test <file> [symbol]"),
        }
    }
    
    filename := args[0]
    
    p := parser.NewParser()
    pf, err := p.ParseFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    var symbols []parser.Symbol
    
    if len(args) > 1 {
        // Specific symbol
        symbol, err := parser.FindSymbol(pf, args[1])
        if err != nil {
            return CommandResult{Error: err}
        }
        symbols = []parser.Symbol{*symbol}
    } else {
        // All functions
        symbols, err = parser.FindSymbolsOfType(pf, parser.SymbolFunction)
        if err != nil {
            return CommandResult{Error: err}
        }
    }
    
    prompt := buildTestPrompt(filename, symbols, pf.Language.Name)
    
    return CommandResult{
        Output:    prompt,
        AddToChat: true,
    }
}

func buildTestPrompt(filename string, symbols []parser.Symbol, lang string) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("Generate comprehensive tests for the following %s code:\n\n", lang))
    
    for _, sym := range symbols {
        builder.WriteString(fmt.Sprintf("### %s\n\n```%s\n%s\n```\n\n",
            sym.Name, strings.ToLower(lang), sym.Content))
    }
    
    testFramework := getTestFramework(lang)
    
    builder.WriteString(fmt.Sprintf("Use the %s testing framework.\n\n", testFramework))
    
    builder.WriteString("Generate tests that cover:\n")
    builder.WriteString("1. **Happy path**: Normal expected inputs\n")
    builder.WriteString("2. **Edge cases**: Empty values, boundaries, limits\n")
    builder.WriteString("3. **Error cases**: Invalid inputs, error conditions\n")
    builder.WriteString("4. **Null/nil handling**: Missing or undefined values\n\n")
    
    builder.WriteString("For each test:\n")
    builder.WriteString("- Use descriptive test names\n")
    builder.WriteString("- Include setup and assertions\n")
    builder.WriteString("- Add comments explaining what's being tested\n\n")
    
    testFilename := generateTestFilename(filename, lang)
    builder.WriteString(fmt.Sprintf("The tests should go in `%s`.\n", testFilename))
    
    return builder.String()
}

func getTestFramework(lang string) string {
    frameworks := map[string]string{
        "Go":         "testing package with testify",
        "JavaScript": "Jest",
        "TypeScript": "Jest with ts-jest",
        "Python":     "pytest",
        "Rust":       "#[cfg(test)] module",
        "Java":       "JUnit 5",
    }
    
    if f, ok := frameworks[lang]; ok {
        return f
    }
    return "standard testing framework"
}

func generateTestFilename(filename, lang string) string {
    ext := filepath.Ext(filename)
    base := strings.TrimSuffix(filename, ext)
    
    switch lang {
    case "Go":
        return base + "_test.go"
    case "JavaScript", "TypeScript":
        return base + ".test" + ext
    case "Python":
        return "test_" + filepath.Base(filename)
    case "Rust":
        return filename // Tests in same file
    case "Java":
        return base + "Test.java"
    default:
        return base + "_test" + ext
    }
}
```

**Usage:**
```
/test handlers.go                    # Generate tests for all functions
/test handlers.go CreateUser         # Generate tests for specific function
```

---

### 4. GitHub Integration

Connect to GitHub API for PR and issue context.

```go
// internal/github/client.go

package github

import (
    "context"
    "os"
    
    "github.com/google/go-github/v57/github"
    "golang.org/x/oauth2"
)

type Client struct {
    client *github.Client
    owner  string
    repo   string
}

func NewClient() (*Client, error) {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        return nil, fmt.Errorf("GITHUB_TOKEN not set")
    }
    
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(context.Background(), ts)
    
    client := github.NewClient(tc)
    
    // Detect owner/repo from git remote
    owner, repo, err := detectRepo()
    if err != nil {
        return nil, err
    }
    
    return &Client{
        client: client,
        owner:  owner,
        repo:   repo,
    }, nil
}

func detectRepo() (string, string, error) {
    // Parse from git remote URL
    // git remote get-url origin -> github.com/owner/repo
    // Implementation details...
    return "", "", nil
}

// GetPR fetches pull request details
func (c *Client) GetPR(number int) (*PRInfo, error) {
    pr, _, err := c.client.PullRequests.Get(
        context.Background(),
        c.owner,
        c.repo,
        number,
    )
    if err != nil {
        return nil, err
    }
    
    // Get diff
    diff, _, err := c.client.PullRequests.GetRaw(
        context.Background(),
        c.owner,
        c.repo,
        number,
        github.RawOptions{Type: github.Diff},
    )
    if err != nil {
        return nil, err
    }
    
    return &PRInfo{
        Number:      number,
        Title:       pr.GetTitle(),
        Body:        pr.GetBody(),
        Author:      pr.GetUser().GetLogin(),
        State:       pr.GetState(),
        BaseBranch:  pr.GetBase().GetRef(),
        HeadBranch:  pr.GetHead().GetRef(),
        Diff:        diff,
        ChangedFiles: pr.GetChangedFiles(),
        Additions:   pr.GetAdditions(),
        Deletions:   pr.GetDeletions(),
    }, nil
}

type PRInfo struct {
    Number       int
    Title        string
    Body         string
    Author       string
    State        string
    BaseBranch   string
    HeadBranch   string
    Diff         string
    ChangedFiles int
    Additions    int
    Deletions    int
}

// GetIssue fetches issue details
func (c *Client) GetIssue(number int) (*IssueInfo, error) {
    issue, _, err := c.client.Issues.Get(
        context.Background(),
        c.owner,
        c.repo,
        number,
    )
    if err != nil {
        return nil, err
    }
    
    // Get comments
    comments, _, err := c.client.Issues.ListComments(
        context.Background(),
        c.owner,
        c.repo,
        number,
        nil,
    )
    if err != nil {
        return nil, err
    }
    
    issueComments := make([]Comment, len(comments))
    for i, c := range comments {
        issueComments[i] = Comment{
            Author: c.GetUser().GetLogin(),
            Body:   c.GetBody(),
        }
    }
    
    return &IssueInfo{
        Number:   number,
        Title:    issue.GetTitle(),
        Body:     issue.GetBody(),
        Author:   issue.GetUser().GetLogin(),
        State:    issue.GetState(),
        Labels:   extractLabels(issue.Labels),
        Comments: issueComments,
    }, nil
}

type IssueInfo struct {
    Number   int
    Title    string
    Body     string
    Author   string
    State    string
    Labels   []string
    Comments []Comment
}

type Comment struct {
    Author string
    Body   string
}
```

**Commands:**
```
/pr 123              # Get PR context
/pr 123 --review     # Review PR
/issue 456           # Get issue context
/issue 456 --fix     # Suggest fix for issue
```

---

### 5. Session Persistence

Save and load conversation sessions.

```go
// internal/session/persistence.go

package session

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type PersistedSession struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Title     string    `json:"title"`
    Messages  []Message `json:"messages"`
    Provider  string    `json:"provider"`
    Model     string    `json:"model"`
}

type Message struct {
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}

type SessionStore struct {
    baseDir string
}

func NewSessionStore() (*SessionStore, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }
    
    baseDir := filepath.Join(homeDir, ".config", "flux", "sessions")
    
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, err
    }
    
    return &SessionStore{baseDir: baseDir}, nil
}

func (s *SessionStore) Save(session *PersistedSession) error {
    if session.ID == "" {
        session.ID = generateID()
    }
    session.UpdatedAt = time.Now()
    
    filename := filepath.Join(s.baseDir, session.ID+".json")
    
    data, err := json.MarshalIndent(session, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}

func (s *SessionStore) Load(id string) (*PersistedSession, error) {
    filename := filepath.Join(s.baseDir, id+".json")
    
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var session PersistedSession
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, err
    }
    
    return &session, nil
}

func (s *SessionStore) List() ([]SessionSummary, error) {
    entries, err := os.ReadDir(s.baseDir)
    if err != nil {
        return nil, err
    }
    
    var summaries []SessionSummary
    
    for _, entry := range entries {
        if filepath.Ext(entry.Name()) != ".json" {
            continue
        }
        
        id := strings.TrimSuffix(entry.Name(), ".json")
        session, err := s.Load(id)
        if err != nil {
            continue
        }
        
        summaries = append(summaries, SessionSummary{
            ID:        session.ID,
            Title:     session.Title,
            UpdatedAt: session.UpdatedAt,
            Messages:  len(session.Messages),
        })
    }
    
    // Sort by updated time
    sort.Slice(summaries, func(i, j int) bool {
        return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
    })
    
    return summaries, nil
}

func (s *SessionStore) Delete(id string) error {
    filename := filepath.Join(s.baseDir, id+".json")
    return os.Remove(filename)
}

type SessionSummary struct {
    ID        string
    Title     string
    UpdatedAt time.Time
    Messages  int
}

func generateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

**Commands:**
```
/save [name]         # Save current session
/load                # List and load sessions
/load <id>           # Load specific session
/sessions            # List all sessions
/delete <id>         # Delete a session
```

---

### 6. Plugin System

Allow custom commands via plugins.

```go
// internal/plugins/loader.go

package plugins

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

type Plugin struct {
    Name        string
    Description string
    Command     string
    Path        string
}

type PluginManager struct {
    plugins map[string]*Plugin
    baseDir string
}

func NewPluginManager() (*PluginManager, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }
    
    baseDir := filepath.Join(homeDir, ".config", "flux", "plugins")
    
    pm := &PluginManager{
        plugins: make(map[string]*Plugin),
        baseDir: baseDir,
    }
    
    if err := pm.loadPlugins(); err != nil {
        return nil, err
    }
    
    return pm, nil
}

func (pm *PluginManager) loadPlugins() error {
    if _, err := os.Stat(pm.baseDir); os.IsNotExist(err) {
        return nil // No plugins directory
    }
    
    entries, err := os.ReadDir(pm.baseDir)
    if err != nil {
        return err
    }
    
    for _, entry := range entries {
        if !entry.IsDir() && isExecutable(entry) {
            name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
            pm.plugins[name] = &Plugin{
                Name:    name,
                Command: "/" + name,
                Path:    filepath.Join(pm.baseDir, entry.Name()),
            }
        }
    }
    
    return nil
}

func (pm *PluginManager) Execute(name string, args []string, input string) (string, error) {
    plugin, ok := pm.plugins[name]
    if !ok {
        return "", fmt.Errorf("unknown plugin: %s", name)
    }
    
    cmd := exec.Command(plugin.Path, args...)
    cmd.Stdin = strings.NewReader(input)
    
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("plugin error: %w", err)
    }
    
    return string(output), nil
}

func (pm *PluginManager) List() []*Plugin {
    var plugins []*Plugin
    for _, p := range pm.plugins {
        plugins = append(plugins, p)
    }
    return plugins
}

func (pm *PluginManager) IsPlugin(name string) bool {
    _, ok := pm.plugins[name]
    return ok
}

func isExecutable(entry os.DirEntry) bool {
    info, err := entry.Info()
    if err != nil {
        return false
    }
    return info.Mode()&0111 != 0
}
```

**Plugin Interface:**
Plugins are executables in `~/.config/flux/plugins/` that:
- Receive arguments via command line
- Receive context via stdin
- Output result to stdout

**Example Plugin (bash):**
```bash
#!/bin/bash
# ~/.config/flux/plugins/weather
# Usage: /weather <city>

city="${1:-London}"
curl -s "wttr.in/${city}?format=3"
```

---

### 7. MCP Integration

Support for Model Context Protocol.

```go
// internal/mcp/client.go

package mcp

import (
    "context"
    "encoding/json"
    "net/rpc"
)

// MCPClient implements the Model Context Protocol client
type MCPClient struct {
    client *rpc.Client
    tools  map[string]Tool
}

type Tool struct {
    Name        string
    Description string
    InputSchema json.RawMessage
}

type ToolCall struct {
    Name      string
    Arguments map[string]interface{}
}

type ToolResult struct {
    Content string
    Error   string
}

func NewMCPClient(address string) (*MCPClient, error) {
    client, err := rpc.Dial("tcp", address)
    if err != nil {
        return nil, err
    }
    
    return &MCPClient{
        client: client,
        tools:  make(map[string]Tool),
    }, nil
}

func (c *MCPClient) Initialize() error {
    var tools []Tool
    err := c.client.Call("MCP.ListTools", struct{}{}, &tools)
    if err != nil {
        return err
    }
    
    for _, tool := range tools {
        c.tools[tool.Name] = tool
    }
    
    return nil
}

func (c *MCPClient) Call(ctx context.Context, call ToolCall) (*ToolResult, error) {
    var result ToolResult
    err := c.client.Call("MCP.CallTool", call, &result)
    if err != nil {
        return nil, err
    }
    
    return &result, nil
}

func (c *MCPClient) GetTools() []Tool {
    var tools []Tool
    for _, t := range c.tools {
        tools = append(tools, t)
    }
    return tools
}

func (c *MCPClient) Close() error {
    return c.client.Close()
}
```

---

## Testing

### Unit Tests

| Test File | Tests |
|-----------|-------|
| `internal/commands/review_test.go` | Review prompt generation |
| `internal/commands/refactor_test.go` | Refactor prompt generation |
| `internal/commands/test_test.go` | Test generation prompts |
| `internal/github/client_test.go` | GitHub API mocking |
| `internal/session/persistence_test.go` | Save/load sessions |
| `internal/plugins/loader_test.go` | Plugin loading/execution |

### Manual Testing

- [ ] `/review` generates useful review prompts
- [ ] `/refactor` extracts and analyzes symbols correctly
- [ ] `/test` generates appropriate test structures
- [ ] GitHub integration fetches PR/issue data
- [ ] Sessions persist and reload correctly
- [ ] Plugins load and execute

---

## Dependencies

```bash
go get github.com/google/go-github/v57
go get golang.org/x/oauth2
```

---

## Acceptance Criteria

1. **Review:** `/review` provides structured code review prompts
2. **Refactor:** `/refactor` analyzes code and suggests improvements
3. **Test:** `/test` generates comprehensive test suggestions
4. **GitHub:** Can fetch PR and issue context when configured
5. **Sessions:** Can save and load conversation history
6. **Plugins:** Custom plugins work correctly

---

## Definition of Done

- [ ] All advanced commands implemented
- [ ] GitHub integration working (optional feature)
- [ ] Session persistence working
- [ ] Plugin system functional
- [ ] Unit tests for all features
- [ ] Documentation for each feature
- [ ] Committed to version control

---

## Notes

- GitHub integration requires `GITHUB_TOKEN` environment variable
- Plugins must be executable files
- Session files are JSON for easy debugging/editing
- MCP integration is optional and experimental
- Consider rate limiting for GitHub API calls
