# Flux CLI - Project Plan

A terminal-based AI coding assistant CLI tool written in Go, similar to Claude Code, OpenCode, and Gemini CLI. Users can type `flux` in the terminal to open an interactive TUI for AI-powered code assistance using open-source AI models with their own API keys.

---

## 1. Project Overview

### Goals
- Create a TUI-based CLI tool invoked by typing `flux` in the terminal
- Support BYOK (Bring Your Own Key) for open-source AI models (Ollama, OpenRouter, Groq, Together AI, etc.)
- Provide an interactive chat interface similar to Claude Code/Gemini CLI
- Enable code context awareness (read files, understand project structure)
- Render AI responses with proper markdown formatting and syntax highlighting

### Target Features (MVP)
- Interactive prompt input with multi-line support
- Streaming AI responses with markdown rendering
- Conversation history within session
- File reading and context injection
- Configuration file for API keys and model selection
- Keyboard shortcuts (Ctrl+C to quit, etc.)

---

## 2. Technology Stack

### Core Framework Libraries

| Library | Purpose | Documentation Reference |
|---------|---------|------------------------|
| **Bubble Tea** (`github.com/charmbracelet/bubbletea`) | TUI framework with Elm architecture (Model-View-Update pattern) | High-quality Go TUI framework |
| **Bubbles** (`github.com/charmbracelet/bubbles`) | Pre-built TUI components (textarea, viewport, spinner, etc.) | Input fields, scrollable areas, loading indicators |
| **Lip Gloss** (`github.com/charmbracelet/lipgloss`) | Terminal styling (colors, borders, padding) | CSS-like styling for terminal |
| **Glamour** (`github.com/charmbracelet/glamour`) | Markdown rendering for terminal | Render AI responses with syntax highlighting |
| **Cobra** (`github.com/spf13/cobra`) | CLI command structure and flags | Command-line argument parsing |
| **Viper** (`github.com/spf13/viper`) | Configuration management | API keys, model settings, preferences |

### AI/HTTP Libraries

| Library | Purpose |
|---------|---------|
| `net/http` (stdlib) | HTTP client for API calls |
| `bufio` (stdlib) | Streaming response handling (SSE) |
| `encoding/json` (stdlib) | JSON parsing for API requests/responses |

### Developer-Focused Libraries

| Library | Purpose | Replaces |
|---------|---------|----------|
| **go-git** (`github.com/go-git/go-git/v5`) | Pure Go git implementation - access diffs, staged files, commit history, blame without shelling out | Basic `os.ReadFile` for context |
| **go-tree-sitter** (`github.com/smacker/go-tree-sitter`) | Incremental code parsing - extract functions, classes, symbols for intelligent context | Simple regex-based project detection |
| **go-github** (`github.com/google/go-github/v57`) | GitHub API client - PR context, issues, code search | N/A (new capability) |

---

## 3. Project Structure

```
flux/
├── cmd/
│   └── root.go              # Cobra root command (entry point)
├── internal/
│   ├── app/
│   │   └── app.go           # Main Bubble Tea application
│   ├── ui/
│   │   ├── model.go         # Main TUI model (state)
│   │   ├── update.go        # Message handling (Update function)
│   │   ├── view.go          # UI rendering (View function)
│   │   ├── styles.go        # Lip Gloss style definitions
│   │   └── components/
│   │       ├── input.go     # Text input component
│   │       ├── messages.go  # Chat message display
│   │       ├── statusbar.go # Status bar component
│   │       └── help.go      # Help overlay
│   ├── ai/
│   │   ├── client.go        # AI API client interface
│   │   ├── openai.go        # OpenAI-compatible API (works with most providers)
│   │   ├── ollama.go        # Ollama local API
│   │   └── streaming.go     # SSE streaming handler
│   ├── config/
│   │   ├── config.go        # Viper configuration loading
│   │   └── models.go        # Configuration structs
│   ├── git/
│   │   ├── repo.go          # go-git repository operations
│   │   ├── diff.go          # Git diff parsing and context
│   │   └── blame.go         # Git blame for code attribution
│   ├── parser/
│   │   ├── treesitter.go    # Tree-sitter code parsing
│   │   ├── symbols.go       # Extract functions, classes, types
│   │   └── languages.go     # Language grammar loaders
│   └── context/
│       ├── builder.go       # Smart context builder (combines git + parser)
│       └── project.go       # Project structure detection
├── main.go                  # Application entry point
├── go.mod
├── go.sum
├── config.example.yaml      # Example configuration file
└── README.md
```

---

## 4. Core Components Architecture

### 4.1 Bubble Tea Model (State)

```go
type Model struct {
    // UI Components
    textarea     textarea.Model    // User input field
    viewport     viewport.Model    // Scrollable message history
    spinner      spinner.Model     // Loading indicator
    help         help.Model        // Help display

    // Application State
    messages     []Message         // Conversation history
    loading      bool              // AI response in progress
    streaming    string            // Current streaming response
    err          error             // Error state

    // Configuration
    config       *Config           // User configuration
    aiClient     ai.Client         // AI provider client

    // Window dimensions
    width        int
    height       int
}
```

### 4.2 Message Types

```go
type Message struct {
    Role      string    // "user" | "assistant" | "system"
    Content   string    // Message content
    Timestamp time.Time
}

// Bubble Tea Messages
type streamChunkMsg string      // Streaming response chunk
type streamDoneMsg struct{}     // Stream completed
type streamErrMsg error         // Stream error
type windowSizeMsg tea.WindowSizeMsg
```

### 4.3 Update Loop (Message Handling)

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc":
            return m, tea.Quit
        case "enter":
            if !m.loading && m.textarea.Value() != "" {
                return m.sendMessage()
            }
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Resize components
    case streamChunkMsg:
        m.streaming += string(msg)
        m.viewport.SetContent(m.renderMessages())
    case streamDoneMsg:
        m.finalizeMessage()
    case streamErrMsg:
        m.err = msg
    }
    // Update sub-components
    return m, nil
}
```

### 4.4 View Rendering

```go
func (m Model) View() string {
    // Build layout:
    // ┌─────────────────────────────────────┐
    // │ Flux - AI Coding Assistant          │ <- Header
    // ├─────────────────────────────────────┤
    // │                                     │
    // │ [Conversation History - Viewport]   │ <- Scrollable
    // │                                     │
    // ├─────────────────────────────────────┤
    // │ > User input textarea               │ <- Input
    // ├─────────────────────────────────────┤
    // │ Ctrl+C: Quit | Enter: Send          │ <- Status bar
    // └─────────────────────────────────────┘

    header := m.renderHeader()
    messages := m.viewport.View()
    input := m.textarea.View()
    status := m.renderStatusBar()

    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        messages,
        input,
        status,
    )
}
```

---

## 5. AI Provider Integration

### 5.1 Provider Interface

```go
type Client interface {
    // Stream sends a prompt and streams the response
    Stream(ctx context.Context, messages []Message) (<-chan string, <-chan error)

    // Complete sends a prompt and returns the full response
    Complete(ctx context.Context, messages []Message) (string, error)

    // SetModel changes the active model
    SetModel(model string)
}
```

### 5.2 Supported Providers (OpenAI-Compatible)

Most providers support the OpenAI API format:

| Provider | Base URL | Notes |
|----------|----------|-------|
| **OpenRouter** | `https://openrouter.ai/api/v1` | Multiple models, pay-per-token |
| **Groq** | `https://api.groq.com/openai/v1` | Fast inference, free tier |
| **Together AI** | `https://api.together.xyz/v1` | Open-source models |
| **Ollama** | `http://localhost:11434/v1` | Local, no API key needed |
| **LM Studio** | `http://localhost:1234/v1` | Local, no API key needed |
| **Mistral** | `https://api.mistral.ai/v1` | Mistral models |

### 5.3 Streaming Implementation (SSE)

```go
func (c *OpenAIClient) Stream(ctx context.Context, messages []Message) (<-chan string, <-chan error) {
    chunks := make(chan string)
    errs := make(chan error, 1)

    go func() {
        defer close(chunks)
        defer close(errs)

        req := c.buildRequest(messages)
        req.Stream = true

        resp, err := c.httpClient.Do(req)
        if err != nil {
            errs <- err
            return
        }
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            if strings.HasPrefix(line, "data: ") {
                data := strings.TrimPrefix(line, "data: ")
                if data == "[DONE]" {
                    return
                }
                // Parse SSE chunk, extract delta content
                var chunk StreamChunk
                json.Unmarshal([]byte(data), &chunk)
                if content := chunk.Choices[0].Delta.Content; content != "" {
                    chunks <- content
                }
            }
        }
    }()

    return chunks, errs
}
```

---

## 6. Configuration System

### 6.1 Configuration File (`~/.config/flux/config.yaml`)

```yaml
# Flux CLI Configuration

# Default AI provider
provider: openrouter

# Provider configurations
providers:
  openrouter:
    api_key: ${OPENROUTER_API_KEY}  # Environment variable reference
    base_url: https://openrouter.ai/api/v1
    model: anthropic/claude-3-haiku

  groq:
    api_key: ${GROQ_API_KEY}
    base_url: https://api.groq.com/openai/v1
    model: llama-3.1-70b-versatile

  ollama:
    base_url: http://localhost:11434/v1
    model: codellama:13b

  together:
    api_key: ${TOGETHER_API_KEY}
    base_url: https://api.together.xyz/v1
    model: meta-llama/Llama-3-70b-chat-hf

# UI preferences
ui:
  theme: dark                    # dark | light | auto
  word_wrap: 80                  # Line wrap width
  show_tokens: true              # Display token count
  syntax_highlighting: true      # Enable code highlighting

# System prompt (injected into every conversation)
system_prompt: |
  You are a helpful AI coding assistant. You help users with programming tasks,
  code reviews, debugging, and explaining code concepts. Be concise and practical.
```

### 6.2 Viper Configuration Loading

```go
func LoadConfig() (*Config, error) {
    v := viper.New()

    // Set defaults
    v.SetDefault("provider", "ollama")
    v.SetDefault("ui.theme", "dark")
    v.SetDefault("ui.word_wrap", 80)

    // Config file paths
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("$HOME/.config/flux")
    v.AddConfigPath(".")

    // Environment variables
    v.SetEnvPrefix("FLUX")
    v.AutomaticEnv()

    // Read config
    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
        // Config file not found, use defaults
    }

    var config Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

---

## 7. UI/UX Design

### 7.1 Main Interface Layout

```
╭─────────────────────────────────────────────────────────────╮
│  flux  AI Coding Assistant         Model: llama-3.1-70b    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  You:                                                       │
│  How do I implement a binary search in Go?                  │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  Assistant:                                                 │
│  Here's a binary search implementation in Go:               │
│                                                             │
│  ```go                                                      │
│  func binarySearch(arr []int, target int) int {            │
│      left, right := 0, len(arr)-1                          │
│      for left <= right {                                   │
│          mid := left + (right-left)/2                      │
│          if arr[mid] == target {                           │
│              return mid                                    │
│          } else if arr[mid] < target {                     │
│              left = mid + 1                                │
│          } else {                                          │
│              right = mid - 1                               │
│          }                                                 │
│      }                                                     │
│      return -1                                             │
│  }                                                         │
│  ```                                                       │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  > Type your message...                                     │
├─────────────────────────────────────────────────────────────┤
│  Ctrl+C quit • Enter send • Ctrl+L clear • ? help          │
╰─────────────────────────────────────────────────────────────╯
```

### 7.2 Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+C` / `Esc` | Quit application |
| `Ctrl+L` | Clear conversation |
| `Ctrl+N` | New conversation |
| `↑` / `↓` | Scroll history |
| `?` | Toggle help |
| `Tab` | Cycle through providers |

### 7.3 Lip Gloss Styling

```go
var (
    // Color palette
    primaryColor   = lipgloss.Color("#7D56F4")
    secondaryColor = lipgloss.Color("#00D4AA")
    errorColor     = lipgloss.Color("#FF6B6B")
    mutedColor     = lipgloss.Color("#626262")

    // Styles
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(primaryColor).
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(primaryColor).
        Padding(0, 1)

    userMessageStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#3C3C3C")).
        Padding(1, 2).
        MarginBottom(1)

    assistantMessageStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Padding(1, 2).
        MarginBottom(1)

    inputStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(secondaryColor).
        Padding(0, 1)

    statusBarStyle = lipgloss.NewStyle().
        Foreground(mutedColor).
        Padding(0, 1)
)
```

---

## 8. Developer-Focused Features

### 8.1 Git-Aware Context (go-git)

Unlike basic file reading, Flux uses **go-git** to provide intelligent git context:

| Command | Description | Example Use Case |
|---------|-------------|------------------|
| `/diff` | Show unstaged changes with AI analysis | "What bugs might this diff introduce?" |
| `/staged` | Show staged changes for review | "Review my staged changes before commit" |
| `/blame <file>` | Show git blame with context | "Who wrote this function and why?" |
| `/log [n]` | Recent commit history | "Summarize the last 5 commits" |
| `/branch` | Current branch info + recent commits | "What's the state of this branch?" |
| `/stash` | List stashed changes | "What did I stash last week?" |

```go
// Example: Get diff context for AI
func (g *GitContext) GetDiffContext() (string, error) {
    repo, _ := git.PlainOpen(".")
    worktree, _ := repo.Worktree()
    status, _ := worktree.Status()

    var context strings.Builder
    context.WriteString("## Git Status\n")
    for file, s := range status {
        context.WriteString(fmt.Sprintf("- %s: %c%c\n", file, s.Staging, s.Worktree))
    }

    // Get actual diff content
    head, _ := repo.Head()
    commit, _ := repo.CommitObject(head.Hash())
    tree, _ := commit.Tree()

    // Compare working tree to HEAD
    changes, _ := worktree.Diff(&git.DiffOptions{})
    context.WriteString("\n## Changes\n```diff\n")
    context.WriteString(changes.String())
    context.WriteString("\n```")

    return context.String(), nil
}
```

### 8.2 Intelligent Code Parsing (tree-sitter)

**Replaces** simple regex-based detection with **tree-sitter** for:

| Feature | Description | Benefit |
|---------|-------------|---------|
| **Symbol extraction** | Extract functions, classes, types, interfaces | "Explain the `ProcessOrder` function" - AI gets just that function |
| **Scope-aware context** | Understand code structure | Include only relevant imports/dependencies |
| **Multi-language support** | Go, Python, JS/TS, Rust, Java, C/C++ | Works across polyglot codebases |
| **Incremental parsing** | Re-parse only changed portions | Fast context updates on file changes |

```go
// Example: Extract function by name using tree-sitter
func ExtractFunction(code []byte, funcName string, lang *sitter.Language) (string, error) {
    parser := sitter.NewParser()
    parser.SetLanguage(lang)

    tree, _ := parser.ParseCtx(context.Background(), nil, code)
    root := tree.RootNode()

    // Query for function definitions
    query := `(function_declaration name: (identifier) @name) @func`
    q, _ := sitter.NewQuery([]byte(query), lang)
    qc := sitter.NewQueryCursor()
    qc.Exec(q, root)

    for {
        match, ok := qc.NextMatch()
        if !ok {
            break
        }
        for _, capture := range match.Captures {
            if capture.Node.Content(code) == funcName {
                // Return the full function node
                funcNode := capture.Node.Parent()
                return funcNode.Content(code), nil
            }
        }
    }
    return "", fmt.Errorf("function %s not found", funcName)
}
```

**Supported Languages (via go-tree-sitter):**
- Go, Python, JavaScript, TypeScript, Rust
- Java, C, C++, C#, Ruby, PHP
- HTML, CSS, JSON, YAML, Markdown
- Bash, SQL, and 50+ more

### 8.3 Smart Context Commands

| Command | Old Behavior | New Behavior (Dev-Focused) |
|---------|--------------|----------------------------|
| `/read <file>` | Dump entire file | Parse with tree-sitter, extract relevant symbols only |
| `/project` | List files | Analyze structure + extract public APIs/exports |
| `/context` | N/A | Show what's currently in AI context (files, symbols, git state) |
| `/symbols <file>` | N/A | List all functions/classes/types in file |
| `/deps` | N/A | Show dependencies (go.mod, package.json, etc.) |
| `/test` | N/A | Find and show related test files |
| `/impl <interface>` | N/A | Find implementations of an interface |

### 8.4 Context Priority System

Smart context management to fit within token limits:

```go
type ContextItem struct {
    Type     ContextType  // Git, File, Symbol, Error
    Priority int          // Higher = more important
    Content  string
    Tokens   int          // Estimated token count
}

// Priority ranking (highest first):
// 1. Current error/stack trace (if debugging)
// 2. Git diff (what user is working on)
// 3. Explicitly referenced files/symbols
// 4. Related test files
// 5. Import/dependency context
// 6. Project structure overview
```

### 8.5 Developer Workflow Commands

| Command | Description |
|---------|-------------|
| `/fix` | Auto-detect errors in git diff, suggest fixes |
| `/review` | Code review mode for staged changes |
| `/commit` | Generate commit message from staged changes |
| `/test` | Suggest tests for current changes |
| `/refactor <symbol>` | Suggest refactoring for a function/class |
| `/explain <symbol>` | Deep explanation of a code symbol |
| `/doc <symbol>` | Generate documentation for a symbol |

---

## 9. Implementation Phases

> **Detailed phase documentation:** Each phase has a dedicated file with complete specifications, file lists, code examples, and testing requirements.

### Phase 1: Foundation (Week 1) → [phase1.md](./phase1.md)
- [x] Project setup with Go modules
- [ ] Basic Cobra CLI structure (`flux` command)
- [ ] Bubble Tea application skeleton
- [ ] Basic TUI layout (header, messages, input, status)
- [ ] Lip Gloss styling

### Phase 2: Core Functionality (Week 2) → [phase2.md](./phase2.md)
- [ ] Text input component (multi-line textarea)
- [ ] Viewport for scrollable messages
- [ ] Message rendering with Glamour (markdown)
- [ ] Viper configuration loading
- [ ] Basic AI client interface

### Phase 3: AI Integration (Week 3) → [phase3.md](./phase3.md)
- [ ] OpenAI-compatible API client
- [ ] SSE streaming support
- [ ] Streaming response display
- [ ] Error handling and retry logic
- [ ] Multiple provider support

### Phase 4: Git Integration (Week 4) → [phase4.md](./phase4.md)
- [ ] go-git repository detection and initialization
- [ ] `/diff`, `/staged`, `/log` commands
- [ ] `/blame` with context
- [ ] `/commit` message generation
- [ ] Git status in status bar

### Phase 5: Code Intelligence (Week 5) → [phase5.md](./phase5.md)
- [ ] tree-sitter integration with language grammars
- [ ] Symbol extraction (functions, classes, types)
- [ ] `/symbols`, `/explain`, `/doc` commands
- [ ] Smart context builder with priority system
- [ ] `/read` with intelligent truncation

### Phase 6: Polish & UX (Week 6) → [phase6.md](./phase6.md)
- [ ] Help overlay
- [ ] Keyboard shortcuts
- [ ] Loading spinners
- [ ] Token counting display
- [ ] Session management (clear, new)
- [ ] Context viewer (`/context`)

### Phase 7: Advanced Features (Future) → [phase7.md](./phase7.md)
- [ ] `/review` code review mode
- [ ] `/refactor` suggestions
- [ ] `/test` generation
- [ ] GitHub integration (PRs, issues)
- [ ] Conversation history persistence
- [ ] Plugin system

---

## 10. Dependencies (go.mod)

```go
module github.com/yourusername/flux

go 1.21

require (
    // TUI Framework
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.17.1
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/glamour v0.6.0

    // CLI & Config
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0

    // Developer Tools
    github.com/go-git/go-git/v5 v5.11.0    // Git operations
    github.com/smacker/go-tree-sitter v0.0.0-20240514083259-c5d1f3f5f99e  // Code parsing
    github.com/google/go-github/v57 v57.0.0  // GitHub API (optional)
)
```

---

## 11. Getting Started Commands

```bash
# Initialize the project
mkdir flux && cd flux
go mod init github.com/yourusername/flux

# Install dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/glamour
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/go-git/go-git/v5
go get github.com/smacker/go-tree-sitter

# Initialize Cobra
go install github.com/spf13/cobra-cli@latest
cobra-cli init

# Build and run
go build -o flux .
./flux
```

---

## 12. References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lip Gloss Styling](https://github.com/charmbracelet/lipgloss)
- [Glamour Markdown](https://github.com/charmbracelet/glamour)
- [Cobra CLI](https://github.com/spf13/cobra)
- [Viper Config](https://github.com/spf13/viper)
- [go-git](https://github.com/go-git/go-git) - Pure Go git implementation
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter) - Go bindings for tree-sitter
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)

---

## 13. Similar Projects for Reference

- [Claude Code](https://github.com/anthropics/anthropic-quickstarts) - Anthropic's CLI
- [Aider](https://github.com/paul-gauthier/aider) - AI pair programming in terminal
- [OpenCode](https://github.com/opencode-ai/opencode) - Terminal AI assistant
- [Gemini CLI](https://github.com/google-gemini/gemini-cli) - Google's Gemini CLI
