# Phase 1: Foundation

**Timeline:** Week 1  
**Goal:** Set up project structure, basic CLI entry point, and TUI skeleton

---

## Overview

This phase establishes the foundational architecture for Flux CLI. By the end, running `flux` in the terminal will launch a basic TUI with a header, empty message area, input field, and status bar.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Go module initialization | Set up `go.mod` with all dependencies | P0 |
| Cobra root command | `flux` command entry point | P0 |
| Bubble Tea app skeleton | Basic Model-View-Update structure | P0 |
| TUI layout | Header, viewport, input, status bar | P0 |
| Lip Gloss styling | Color palette and component styles | P1 |
| Graceful exit | Ctrl+C / Esc to quit | P0 |

---

## Files to Create

### Project Root

```
flux/
├── main.go
├── go.mod
├── go.sum
└── README.md
```

### Source Files

| File | Purpose | Key Functions/Types |
|------|---------|---------------------|
| `main.go` | Entry point, calls `cmd.Execute()` | `main()` |
| `cmd/root.go` | Cobra root command definition | `rootCmd`, `Execute()`, `init()` |
| `internal/app/app.go` | Bubble Tea program initialization | `Run()`, `NewProgram()` |
| `internal/ui/model.go` | Main TUI state model | `Model` struct, `Init()`, `Update()`, `View()` |
| `internal/ui/styles.go` | Lip Gloss style definitions | Color constants, style variables |

---

## Detailed File Specifications

### 1. `main.go`

```go
// Entry point - minimal, just calls cmd.Execute()
package main

import "github.com/yourusername/flux/cmd"

func main() {
    cmd.Execute()
}
```

**Requirements:**
- Import cmd package
- Call Execute() and handle errors

---

### 2. `cmd/root.go`

```go
package cmd

// Root command that launches the TUI
var rootCmd = &cobra.Command{
    Use:   "flux",
    Short: "AI-powered coding assistant",
    Long:  `Flux is a terminal-based AI coding assistant...`,
    Run: func(cmd *cobra.Command, args []string) {
        // Launch the TUI
        app.Run()
    },
}
```

**Requirements:**
- Define `flux` as the root command
- No subcommands in Phase 1
- Add version flag (`--version`, `-v`)
- Call `app.Run()` to start TUI

**Flags to implement:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--version` | bool | false | Print version and exit |

---

### 3. `internal/app/app.go`

```go
package app

// Run initializes and starts the Bubble Tea program
func Run() error {
    model := ui.NewModel()
    p := tea.NewProgram(model, tea.WithAltScreen())
    _, err := p.Run()
    return err
}
```

**Requirements:**
- Create new Model instance
- Use alternate screen mode (`tea.WithAltScreen()`)
- Handle program errors gracefully
- Return error to caller for logging

---

### 4. `internal/ui/model.go`

```go
package ui

type Model struct {
    width    int
    height   int
    ready    bool
    quitting bool
}

// Init returns initial command (none for Phase 1)
func (m Model) Init() tea.Cmd {
    return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc", "q":
            m.quitting = true
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.ready = true
    }
    return m, nil
}

// View renders the UI
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }
    if m.quitting {
        return "Goodbye!\n"
    }
    
    return lipgloss.JoinVertical(
        lipgloss.Left,
        m.renderHeader(),
        m.renderMessages(),
        m.renderInput(),
        m.renderStatusBar(),
    )
}
```

**Requirements:**
- Track window dimensions
- Handle resize events
- Implement quit functionality
- Render 4-part layout (header, messages, input, status)

**Helper methods to implement:**
| Method | Returns | Description |
|--------|---------|-------------|
| `renderHeader()` | string | App title + model indicator |
| `renderMessages()` | string | Empty placeholder (Phase 1) |
| `renderInput()` | string | Input prompt placeholder |
| `renderStatusBar()` | string | Keybinding hints |

---

### 5. `internal/ui/styles.go`

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
    PrimaryColor   = lipgloss.Color("#7D56F4")  // Purple
    SecondaryColor = lipgloss.Color("#00D4AA")  // Teal
    ErrorColor     = lipgloss.Color("#FF6B6B")  // Red
    WarningColor   = lipgloss.Color("#FFB86C")  // Orange
    MutedColor     = lipgloss.Color("#626262")  // Gray
    TextColor      = lipgloss.Color("#FAFAFA")  // White
    BgColor        = lipgloss.Color("#1E1E1E")  // Dark
)

// Component styles
var (
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(PrimaryColor).
        Padding(0, 1)

    MessageAreaStyle = lipgloss.NewStyle().
        Padding(1, 2)

    InputStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(SecondaryColor).
        Padding(0, 1)

    StatusBarStyle = lipgloss.NewStyle().
        Foreground(MutedColor).
        Padding(0, 1)

    // Logo/brand style
    LogoStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(PrimaryColor)
)
```

**Requirements:**
- Define consistent color palette
- Create reusable component styles
- Support terminal color profiles (ANSI, 256, TrueColor)

---

## Testing

### Unit Tests

| Test File | Tests | Description |
|-----------|-------|-------------|
| `internal/ui/model_test.go` | Model initialization | Verify NewModel returns valid state |
| | Quit handling | Verify Ctrl+C, Esc, q trigger quit |
| | Window resize | Verify dimensions update correctly |
| `internal/ui/styles_test.go` | Style rendering | Verify styles apply without panic |

### Manual Testing Checklist

- [ ] `go build -o flux .` compiles without errors
- [ ] `./flux` launches TUI in alternate screen
- [ ] Window resize updates layout correctly
- [ ] `Ctrl+C` exits cleanly
- [ ] `Esc` exits cleanly
- [ ] `q` exits cleanly
- [ ] Exit message "Goodbye!" displays briefly
- [ ] No panic on very small terminal sizes (< 20 cols)

### Test Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/ui/...

# Build and run
go build -o flux . && ./flux
```

---

## Dependencies

```bash
# Initialize module
go mod init github.com/yourusername/flux

# Install Phase 1 dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/spf13/cobra
```

---

## Acceptance Criteria

1. **Build:** `go build` succeeds with no errors
2. **Launch:** Running `./flux` opens a full-screen TUI
3. **Layout:** UI shows header, message area, input area, status bar
4. **Styling:** Colors and borders render correctly
5. **Exit:** All quit keys (Ctrl+C, Esc, q) work
6. **Resize:** Terminal resize updates layout without crash
7. **Tests:** All unit tests pass

---

## Definition of Done

- [ ] All files created and compiling
- [ ] Unit tests written and passing
- [ ] Manual testing checklist completed
- [ ] Code reviewed (if applicable)
- [ ] README.md updated with build instructions
- [ ] Committed to version control

---

## Notes

- Keep the TUI minimal - no actual functionality yet
- Focus on clean architecture that scales
- Placeholder text is fine for message area and input
- Status bar should show: `Ctrl+C quit • q quit`
