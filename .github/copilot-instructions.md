# Flux Code CLI - AI Agent Instructions

Flux Code is a terminal-native AI coding assistant built in Go. It uses the Bubble Tea framework for its TUI and is designed to integrate with local (Ollama) and cloud AI models.

## 🏗 Architecture & Core Components

### TUI Architecture (Bubble Tea)
The application follows The Elm Architecture (Model-View-Update) via [Bubble Tea](https://github.com/charmbracelet/bubbletea).

- **Entry Point**: `cmd/root.go` initializes the CLI via Cobra; `internal/app/app.go` starts the Bubble Tea program.
- **State Management**: `internal/ui/model.go` holds the global state (`Model` struct).
- **Components**: UI elements are modularized in `internal/ui/components/` (e.g., `input`, `messages`, `viewport`).
    - **Input**: Handles user text entry.
    - **Viewport**: Displays the chat history.
    - **Messages**: Manages the list of chat messages.

### Configuration
- **Viper**: Used for configuration management in `internal/config/config.go`.
- **Defaults**: Defaults are set in code; overrides come from `$HOME/.config/flux/config.yaml` or local `config.yaml`.

### Missing/Planned Components (Reference `PLAN.md`)
- **AI Integration**: `internal/ai` (Client, Ollama, OpenAI) is currently **missing**.
- **Git Integration**: `internal/git` (Repo, Diff) is **missing**.
- **Context Parsing**: `internal/parser` (Tree-sitter) is **missing**.

## 🧩 Patterns & Conventions

- **Bubble Tea**:
    - `Init()`: Initial command (e.g., cursor blink).
    - `Update(msg)`: Handles events (keys, window resize) and returns `(Model, Cmd)`.
    - `View()`: Renders the UI as a string.
    - **Commands**: Use `tea.Cmd` for side effects (I/O, timers). Return `nil` if no side effect.
- **Styling**: Use [Lip Gloss](https://github.com/charmbracelet/lipgloss) for terminal styling. Define styles in `styles.go`.
- **Error Handling**:
    - `app.Run()` returns errors to `cmd/root.go` for printing.
    - TUI errors should be displayed in the UI or logged, avoiding panic.

## 🛠 Development Workflow

- **Run**: `go run main.go`
- **Build**: `go build -o flux main.go`
- **Test**: `go test ./...`
- **Dependencies**: Managed via `go.mod`. Run `go mod tidy` after adding imports.

## 📂 Key Files

- [main.go](main.go): Application entry point.
- [cmd/root.go](cmd/root.go): Cobra command definition.
- [internal/app/app.go](internal/app/app.go): Application lifecycle and TUI startup.
- [internal/ui/model.go](internal/ui/model.go): Main TUI model and update loop.
- [internal/config/config.go](internal/config/config.go): Configuration loading logic.
- [PLAN.md](PLAN.md): Roadmap for missing features.

## 🚀 Implementation Priorities

When implementing new features, refer to `PLAN.md`. The immediate next steps likely involve:
1.  Scaffolding `internal/ai` to connect to Ollama/OpenAI.
2.  Implementing the `internal/git` package for context awareness.
3.  Connecting the UI input to the AI client.
