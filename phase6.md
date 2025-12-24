# Phase 6: Polish & UX

**Timeline:** Week 6  
**Goal:** Add polish, help system, keyboard shortcuts, loading states, and improved user experience

---

## Overview

This phase focuses on making Flux feel complete and professional. It adds a help system, proper keyboard shortcuts, loading indicators, token counting, session management, and context visibility.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Help overlay | `/help` command and `?` toggle | P0 |
| Keyboard shortcuts | Full shortcut system | P0 |
| Loading spinners | Visual feedback during AI calls | P0 |
| Token counting | Display estimated tokens | P1 |
| Session management | `/clear`, `/new` commands | P0 |
| Context viewer | `/context` to see current context | P1 |
| Error display | Improved error messages | P1 |
| Welcome screen | First-run experience | P2 |

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/ui/components/help.go` | Help overlay component |
| `internal/ui/components/spinner.go` | Loading spinner wrapper |
| `internal/ui/keybindings.go` | Keyboard shortcut definitions |
| `internal/commands/session.go` | Session management commands |
| `internal/tokens/counter.go` | Token estimation |

### Modified Files

| File | Changes |
|------|---------|
| `internal/ui/model.go` | Add help toggle, shortcuts, spinner |
| `internal/ui/styles.go` | Add help overlay styles |
| `internal/ui/components/statusbar.go` | Add token count display |
| `internal/commands/handler.go` | Add help and session commands |

---

## Detailed File Specifications

### 1. `internal/ui/keybindings.go`

```go
package ui

import (
    "github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keyboard shortcuts
type KeyMap struct {
    Quit        key.Binding
    Help        key.Binding
    Send        key.Binding
    Clear       key.Binding
    NewSession  key.Binding
    Cancel      key.Binding
    ScrollUp    key.Binding
    ScrollDown  key.Binding
    PageUp      key.Binding
    PageDown    key.Binding
    Top         key.Binding
    Bottom      key.Binding
    FocusInput  key.Binding
    CycleModel  key.Binding
}

// DefaultKeyMap returns the default keyboard shortcuts
func DefaultKeyMap() KeyMap {
    return KeyMap{
        Quit: key.NewBinding(
            key.WithKeys("ctrl+c", "ctrl+d"),
            key.WithHelp("ctrl+c", "quit"),
        ),
        Help: key.NewBinding(
            key.WithKeys("?", "ctrl+h"),
            key.WithHelp("?", "toggle help"),
        ),
        Send: key.NewBinding(
            key.WithKeys("enter"),
            key.WithHelp("enter", "send message"),
        ),
        Clear: key.NewBinding(
            key.WithKeys("ctrl+l"),
            key.WithHelp("ctrl+l", "clear conversation"),
        ),
        NewSession: key.NewBinding(
            key.WithKeys("ctrl+n"),
            key.WithHelp("ctrl+n", "new session"),
        ),
        Cancel: key.NewBinding(
            key.WithKeys("esc"),
            key.WithHelp("esc", "cancel/close"),
        ),
        ScrollUp: key.NewBinding(
            key.WithKeys("up", "k"),
            key.WithHelp("↑/k", "scroll up"),
        ),
        ScrollDown: key.NewBinding(
            key.WithKeys("down", "j"),
            key.WithHelp("↓/j", "scroll down"),
        ),
        PageUp: key.NewBinding(
            key.WithKeys("pgup", "ctrl+u"),
            key.WithHelp("pgup", "page up"),
        ),
        PageDown: key.NewBinding(
            key.WithKeys("pgdown", "ctrl+d"),
            key.WithHelp("pgdn", "page down"),
        ),
        Top: key.NewBinding(
            key.WithKeys("home", "g"),
            key.WithHelp("home", "go to top"),
        ),
        Bottom: key.NewBinding(
            key.WithKeys("end", "G"),
            key.WithHelp("end", "go to bottom"),
        ),
        FocusInput: key.NewBinding(
            key.WithKeys("i", "tab"),
            key.WithHelp("i/tab", "focus input"),
        ),
        CycleModel: key.NewBinding(
            key.WithKeys("ctrl+m"),
            key.WithHelp("ctrl+m", "cycle model"),
        ),
    }
}

// ShortHelp returns a short help string for the status bar
func (k KeyMap) ShortHelp() []key.Binding {
    return []key.Binding{
        k.Send,
        k.Help,
        k.Quit,
    }
}

// FullHelp returns all keybindings for the help overlay
func (k KeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Send, k.Clear, k.NewSession},
        {k.ScrollUp, k.ScrollDown, k.PageUp, k.PageDown},
        {k.Help, k.Cancel, k.Quit},
    }
}
```

---

### 2. `internal/ui/components/help.go`

```go
package components

import (
    "strings"
    
    "github.com/charmbracelet/lipgloss"
)

// HelpOverlay displays help information
type HelpOverlay struct {
    visible bool
    width   int
    height  int
}

func NewHelpOverlay() HelpOverlay {
    return HelpOverlay{}
}

func (h *HelpOverlay) Toggle() {
    h.visible = !h.visible
}

func (h *HelpOverlay) Show() {
    h.visible = true
}

func (h *HelpOverlay) Hide() {
    h.visible = false
}

func (h HelpOverlay) IsVisible() bool {
    return h.visible
}

func (h *HelpOverlay) SetSize(width, height int) {
    h.width = width
    h.height = height
}

func (h HelpOverlay) View() string {
    if !h.visible {
        return ""
    }
    
    // Styles
    overlayStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#7D56F4")).
        Padding(1, 2).
        Width(60)
    
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#7D56F4")).
        MarginBottom(1)
    
    sectionStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00D4AA")).
        MarginTop(1)
    
    keyStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFB86C"))
    
    descStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA"))
    
    // Build help content
    var builder strings.Builder
    
    builder.WriteString(titleStyle.Render("⚡ Flux Help"))
    builder.WriteString("\n\n")
    
    // Keyboard shortcuts
    builder.WriteString(sectionStyle.Render("Keyboard Shortcuts"))
    builder.WriteString("\n")
    
    shortcuts := []struct{ key, desc string }{
        {"Enter", "Send message"},
        {"Ctrl+C", "Quit Flux"},
        {"Ctrl+L", "Clear conversation"},
        {"Ctrl+N", "New session"},
        {"Esc", "Cancel/close"},
        {"?", "Toggle this help"},
        {"↑/↓", "Scroll messages"},
        {"PgUp/PgDn", "Page scroll"},
    }
    
    for _, s := range shortcuts {
        builder.WriteString(
            keyStyle.Render(padRight(s.key, 12)) +
            descStyle.Render(s.desc) + "\n",
        )
    }
    
    // Commands
    builder.WriteString("\n")
    builder.WriteString(sectionStyle.Render("Slash Commands"))
    builder.WriteString("\n")
    
    commands := []struct{ cmd, desc string }{
        {"/help", "Show this help"},
        {"/clear", "Clear conversation"},
        {"/new", "Start new session"},
        {"/context", "View current context"},
        {"/diff", "Show git diff"},
        {"/staged", "Show staged changes"},
        {"/log [n]", "Show recent commits"},
        {"/blame <file>", "Show git blame"},
        {"/commit", "Generate commit message"},
        {"/read <file>", "Add file to context"},
        {"/symbols <file>", "List symbols in file"},
        {"/explain <file> <sym>", "Explain a symbol"},
        {"/doc <file> <sym>", "Generate docs"},
        {"/project", "Show project structure"},
        {"/deps", "Show dependencies"},
    }
    
    for _, c := range commands {
        builder.WriteString(
            keyStyle.Render(padRight(c.cmd, 22)) +
            descStyle.Render(c.desc) + "\n",
        )
    }
    
    builder.WriteString("\n")
    builder.WriteString(lipgloss.NewStyle().
        Foreground(lipgloss.Color("#626262")).
        Render("Press ? or Esc to close"))
    
    content := overlayStyle.Render(builder.String())
    
    // Center the overlay
    return centerOverlay(content, h.width, h.height)
}

func padRight(s string, length int) string {
    if len(s) >= length {
        return s
    }
    return s + strings.Repeat(" ", length-len(s))
}

func centerOverlay(content string, width, height int) string {
    lines := strings.Split(content, "\n")
    contentHeight := len(lines)
    contentWidth := lipgloss.Width(content)
    
    // Calculate padding
    topPadding := (height - contentHeight) / 2
    leftPadding := (width - contentWidth) / 2
    
    if topPadding < 0 {
        topPadding = 0
    }
    if leftPadding < 0 {
        leftPadding = 0
    }
    
    // Build centered content
    var result strings.Builder
    
    // Top padding
    for i := 0; i < topPadding; i++ {
        result.WriteString("\n")
    }
    
    // Content with left padding
    for _, line := range lines {
        result.WriteString(strings.Repeat(" ", leftPadding))
        result.WriteString(line)
        result.WriteString("\n")
    }
    
    return result.String()
}
```

---

### 3. `internal/ui/components/spinner.go`

```go
package components

import (
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Spinner wraps the bubbles spinner with custom styling
type Spinner struct {
    spinner spinner.Model
    active  bool
    message string
}

func NewSpinner() Spinner {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA"))
    
    return Spinner{
        spinner: s,
        message: "Thinking...",
    }
}

func (s Spinner) Init() tea.Cmd {
    return s.spinner.Tick
}

func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
    if !s.active {
        return s, nil
    }
    
    var cmd tea.Cmd
    s.spinner, cmd = s.spinner.Update(msg)
    return s, cmd
}

func (s Spinner) View() string {
    if !s.active {
        return ""
    }
    
    return s.spinner.View() + " " + s.message
}

func (s *Spinner) Start(message string) tea.Cmd {
    s.active = true
    s.message = message
    return s.spinner.Tick
}

func (s *Spinner) Stop() {
    s.active = false
}

func (s Spinner) IsActive() bool {
    return s.active
}

func (s *Spinner) SetMessage(message string) {
    s.message = message
}

// SpinnerStyles returns different spinner styles
func SpinnerStyles() map[string]spinner.Spinner {
    return map[string]spinner.Spinner{
        "dot":     spinner.Dot,
        "line":    spinner.Line,
        "mini":    spinner.MiniDot,
        "jump":    spinner.Jump,
        "pulse":   spinner.Pulse,
        "points":  spinner.Points,
        "globe":   spinner.Globe,
        "moon":    spinner.Moon,
        "monkey":  spinner.Monkey,
    }
}
```

---

### 4. `internal/tokens/counter.go`

```go
package tokens

import (
    "strings"
    "unicode"
)

// EstimateTokens provides a rough token estimate for text
// Uses a simple heuristic: ~4 characters per token for code
func EstimateTokens(text string) int {
    if text == "" {
        return 0
    }
    
    // More accurate estimation based on content type
    // Code tends to have more tokens per character due to punctuation
    
    charCount := len(text)
    wordCount := len(strings.Fields(text))
    
    // Estimate based on mix of characters and words
    // Words are roughly 1.3 tokens on average
    // Characters are roughly 4 per token
    
    wordTokens := float64(wordCount) * 1.3
    charTokens := float64(charCount) / 4.0
    
    // Use weighted average
    estimate := int((wordTokens + charTokens) / 2)
    
    if estimate < 1 && charCount > 0 {
        return 1
    }
    
    return estimate
}

// EstimateCodeTokens is optimized for code
func EstimateCodeTokens(code string) int {
    if code == "" {
        return 0
    }
    
    // Code has more special characters, so slightly different ratio
    // Roughly 3.5 characters per token for code
    
    return len(code) / 3
}

// FormatTokenCount returns a human-readable token count
func FormatTokenCount(tokens int) string {
    if tokens < 1000 {
        return fmt.Sprintf("%d", tokens)
    }
    
    if tokens < 1000000 {
        return fmt.Sprintf("%.1fK", float64(tokens)/1000)
    }
    
    return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
}

// TokenBudget represents available token budget
type TokenBudget struct {
    Max       int
    Used      int
    Reserved  int // For system prompt, etc.
}

func NewTokenBudget(maxTokens int) *TokenBudget {
    return &TokenBudget{
        Max:      maxTokens,
        Reserved: 500, // Default reservation for system prompt
    }
}

func (tb *TokenBudget) Available() int {
    return tb.Max - tb.Used - tb.Reserved
}

func (tb *TokenBudget) Add(tokens int) bool {
    if tb.Used + tokens > tb.Max - tb.Reserved {
        return false
    }
    tb.Used += tokens
    return true
}

func (tb *TokenBudget) Reset() {
    tb.Used = 0
}

func (tb *TokenBudget) Percentage() float64 {
    return float64(tb.Used) / float64(tb.Max-tb.Reserved) * 100
}
```

---

### 5. `internal/commands/session.go`

```go
package commands

import (
    "fmt"
    "strings"
    
    "github.com/yourusername/flux/internal/context"
)

// Session holds the current session state
type Session struct {
    Messages      []Message
    ContextBuilder *context.ContextBuilder
    TokensUsed    int
}

// SessionManager manages chat sessions
type SessionManager struct {
    current  *Session
    history  []*Session // Previous sessions
    maxHistory int
}

func NewSessionManager() *SessionManager {
    return &SessionManager{
        current:    newSession(),
        history:    []*Session{},
        maxHistory: 10,
    }
}

func newSession() *Session {
    return &Session{
        Messages:       []Message{},
        ContextBuilder: context.NewContextBuilder(8000), // ~8K tokens for context
    }
}

func (sm *SessionManager) Current() *Session {
    return sm.current
}

func (sm *SessionManager) NewSession() {
    // Save current to history
    if len(sm.current.Messages) > 0 {
        sm.history = append(sm.history, sm.current)
        
        // Trim history if needed
        if len(sm.history) > sm.maxHistory {
            sm.history = sm.history[1:]
        }
    }
    
    sm.current = newSession()
}

func (sm *SessionManager) Clear() {
    sm.current.Messages = []Message{}
    sm.current.ContextBuilder.Clear()
    sm.current.TokensUsed = 0
}

// ExecuteSessionCommand handles session-related commands
func ExecuteSessionCommand(cmd *Command, sm *SessionManager) CommandResult {
    switch cmd.Name {
    case "clear":
        sm.Clear()
        return CommandResult{
            Output: "Conversation cleared.",
        }
        
    case "new":
        sm.NewSession()
        return CommandResult{
            Output: "Started new session.",
        }
        
    case "context":
        return executeShowContext(sm.Current())
        
    case "help":
        return executeHelp()
        
    default:
        return CommandResult{
            Error: fmt.Errorf("unknown session command: /%s", cmd.Name),
        }
    }
}

func executeShowContext(session *Session) CommandResult {
    items := session.ContextBuilder.GetItems()
    
    if len(items) == 0 {
        return CommandResult{
            Output: "No context items. Use `/read`, `/diff`, or `/symbols` to add context.",
        }
    }
    
    var builder strings.Builder
    builder.WriteString("## Current Context\n\n")
    
    totalTokens := 0
    for _, item := range items {
        builder.WriteString(fmt.Sprintf("- **%s** (%s) - ~%d tokens\n",
            item.Source, item.Type, item.Tokens))
        totalTokens += item.Tokens
    }
    
    builder.WriteString(fmt.Sprintf("\n**Total:** ~%d tokens", totalTokens))
    
    return CommandResult{
        Output: builder.String(),
    }
}

func executeHelp() CommandResult {
    help := `## Flux Commands

### Session
- \`/help\` - Show this help
- \`/clear\` - Clear conversation  
- \`/new\` - Start new session
- \`/context\` - View current context

### Git
- \`/diff\` - Show unstaged changes
- \`/staged\` - Show staged changes
- \`/log [n]\` - Show recent commits
- \`/blame <file>\` - Show git blame
- \`/commit\` - Generate commit message
- \`/branch\` - Show branch info
- \`/status\` - Show git status

### Code
- \`/read <file> [symbol]\` - Add file/symbol to context
- \`/symbols <file>\` - List symbols in file
- \`/explain <file> <symbol>\` - Explain code
- \`/doc <file> <symbol>\` - Generate documentation
- \`/project\` - Show project structure
- \`/deps\` - Show dependencies

### Keyboard Shortcuts
- \`Enter\` - Send message
- \`Ctrl+C\` - Quit
- \`Ctrl+L\` - Clear conversation
- \`Ctrl+N\` - New session
- \`?\` - Toggle help
- \`↑/↓\` - Scroll messages`

    return CommandResult{
        Output: help,
    }
}
```

---

### 6. Updated `internal/ui/model.go`

```go
package ui

import (
    "github.com/yourusername/flux/internal/ai"
    "github.com/yourusername/flux/internal/commands"
    "github.com/yourusername/flux/internal/tokens"
    "github.com/yourusername/flux/internal/ui/components"
    
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type Model struct {
    // Components
    input     components.Input
    viewport  components.Viewport
    messages  components.Messages
    statusBar components.StatusBar
    help      components.HelpOverlay
    spinner   components.Spinner
    
    // Keybindings
    keys KeyMap
    
    // AI
    aiClient       ai.Client
    streaming      bool
    streamBuf      string
    cancelFn       context.CancelFunc
    
    // Session
    sessionManager *commands.SessionManager
    
    // Tokens
    tokenBudget *tokens.TokenBudget
    
    // State
    width    int
    height   int
    ready    bool
    quitting bool
    err      error
    
    // Focus
    inputFocused bool
}

func NewModel(client ai.Client) Model {
    return Model{
        input:          components.NewInput(),
        messages:       components.NewMessages(80),
        statusBar:      components.NewStatusBar(),
        help:           components.NewHelpOverlay(),
        spinner:        components.NewSpinner(),
        keys:           DefaultKeyMap(),
        aiClient:       client,
        sessionManager: commands.NewSessionManager(),
        tokenBudget:    tokens.NewTokenBudget(128000), // 128K context
        inputFocused:   true,
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        components.TextareaBlink,
        m.spinner.Init(),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Handle help overlay first
    if m.help.IsVisible() {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Cancel) {
                m.help.Hide()
                return m, nil
            }
        }
        return m, nil
    }
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Global shortcuts
        switch {
        case key.Matches(msg, m.keys.Quit):
            if m.streaming {
                if m.cancelFn != nil {
                    m.cancelFn()
                }
                m.streaming = false
                m.spinner.Stop()
                return m, nil
            }
            m.quitting = true
            return m, tea.Quit
            
        case key.Matches(msg, m.keys.Help):
            m.help.Toggle()
            return m, nil
            
        case key.Matches(msg, m.keys.Clear):
            m.sessionManager.Clear()
            m.messages.Clear()
            m.viewport.SetContent("")
            return m, nil
            
        case key.Matches(msg, m.keys.NewSession):
            m.sessionManager.NewSession()
            m.messages.Clear()
            m.viewport.SetContent("")
            return m, nil
            
        case key.Matches(msg, m.keys.Cancel):
            if m.streaming {
                if m.cancelFn != nil {
                    m.cancelFn()
                }
                m.streaming = false
                m.spinner.Stop()
            }
            return m, nil
            
        case key.Matches(msg, m.keys.Send):
            if !m.streaming && m.input.Value() != "" {
                return m, m.handleInput()
            }
        }
        
    case streamChunkMsg:
        m.streamBuf += string(msg)
        m.updateStreamingMessage()
        return m, nil
        
    case streamDoneMsg:
        m.finalizeStream()
        m.spinner.Stop()
        return m, nil
        
    case streamErrMsg:
        m.err = msg.err
        m.streaming = false
        m.spinner.Stop()
        m.messages.Add(components.RoleAssistant, "Error: "+msg.err.Error())
        m.viewport.SetContent(m.messages.Render())
        return m, nil
        
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.handleResize()
        m.ready = true
    }
    
    // Update spinner
    if m.spinner.IsActive() {
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        cmds = append(cmds, cmd)
    }
    
    // Update input (only if not streaming)
    if !m.streaming && m.inputFocused {
        var cmd tea.Cmd
        m.input, cmd = m.input.Update(msg)
        cmds = append(cmds, cmd)
    }
    
    // Update viewport
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m *Model) handleInput() tea.Cmd {
    input := m.input.Value()
    m.input.Reset()
    
    // Check if it's a command
    if commands.IsCommand(input) {
        cmd := commands.Parse(input)
        result := m.executeCommand(cmd)
        
        if result.Error != nil {
            m.messages.Add(components.RoleSystem, "Error: "+result.Error.Error())
        } else if result.Output != "" {
            if result.AddToChat {
                // Add as context for AI
                m.messages.Add(components.RoleUser, input)
                m.messages.Add(components.RoleSystem, result.Output)
            } else {
                // Just display
                m.messages.Add(components.RoleSystem, result.Output)
            }
        }
        
        m.viewport.SetContent(m.messages.Render())
        m.viewport.GotoBottom()
        return nil
    }
    
    // Regular message - send to AI
    m.messages.Add(components.RoleUser, input)
    m.viewport.SetContent(m.messages.Render())
    m.viewport.GotoBottom()
    
    m.streaming = true
    m.streamBuf = ""
    
    return tea.Batch(
        m.spinner.Start("Thinking..."),
        m.streamCommand(),
    )
}

func (m *Model) executeCommand(cmd *commands.Command) commands.CommandResult {
    // Route to appropriate handler
    switch cmd.Name {
    case "help", "clear", "new", "context":
        return commands.ExecuteSessionCommand(cmd, m.sessionManager)
    case "diff", "staged", "log", "blame", "branch", "status", "commit":
        return commands.ExecuteGitCommand(cmd)
    case "read", "symbols", "explain", "doc", "project", "deps":
        return commands.ExecuteCodeCommand(cmd)
    default:
        return commands.CommandResult{
            Error: fmt.Errorf("unknown command: /%s. Type /help for available commands.", cmd.Name),
        }
    }
}

func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }
    if m.quitting {
        return "Goodbye!\n"
    }
    
    // Main layout
    main := lipgloss.JoinVertical(
        lipgloss.Left,
        m.renderHeader(),
        m.renderMainArea(),
        m.renderInputArea(),
        m.statusBar.View(),
    )
    
    // Overlay help if visible
    if m.help.IsVisible() {
        return m.help.View()
    }
    
    return main
}

func (m Model) renderHeader() string {
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#7D56F4"))
    
    modelStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#626262"))
    
    title := titleStyle.Render("⚡ Flux")
    model := modelStyle.Render(" • " + m.aiClient.GetModel())
    
    return title + model + "\n"
}

func (m Model) renderMainArea() string {
    return m.viewport.View()
}

func (m Model) renderInputArea() string {
    var content string
    
    if m.streaming {
        content = m.spinner.View()
    } else {
        content = m.input.View()
    }
    
    style := lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#00D4AA")).
        Padding(0, 1)
    
    return style.Render(content)
}

func (m *Model) handleResize() {
    headerHeight := 2
    statusHeight := 1
    inputHeight := 5
    padding := 2
    
    viewportHeight := m.height - headerHeight - statusHeight - inputHeight - padding
    if viewportHeight < 1 {
        viewportHeight = 1
    }
    
    m.viewport.SetSize(m.width-4, viewportHeight)
    m.input.SetWidth(m.width - 6)
    m.messages.SetWidth(m.width - 8)
    m.statusBar.SetWidth(m.width)
    m.help.SetSize(m.width, m.height)
}
```

---

### 7. Updated Status Bar with Tokens

```go
// internal/ui/components/statusbar.go - additions

func (s StatusBar) View() string {
    leftStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#626262"))
    
    gitStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00D4AA"))
    
    tokenStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFB86C"))
    
    // Left side: shortcuts
    left := leftStyle.Render("Enter send • ? help • Ctrl+C quit")
    
    // Right side: git + tokens + model
    var rightParts []string
    
    if s.gitStatus != "" {
        rightParts = append(rightParts, gitStyle.Render(""+s.gitStatus))
    }
    
    if s.tokenCount > 0 {
        rightParts = append(rightParts, tokenStyle.Render(
            fmt.Sprintf("~%s tokens", tokens.FormatTokenCount(s.tokenCount)),
        ))
    }
    
    rightParts = append(rightParts, s.model)
    
    right := strings.Join(rightParts, " │ ")
    
    // Calculate padding
    padding := s.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
    if padding < 0 {
        padding = 0
    }
    
    return left + strings.Repeat(" ", padding) + right
}

func (s *StatusBar) SetTokenCount(count int) {
    s.tokenCount = count
}
```

---

## Testing

### Unit Tests

| Test File | Tests |
|-----------|-------|
| `internal/ui/keybindings_test.go` | Key matching, help text |
| `internal/ui/components/help_test.go` | Toggle, visibility |
| `internal/ui/components/spinner_test.go` | Start, stop, messages |
| `internal/tokens/counter_test.go` | Token estimation |
| `internal/commands/session_test.go` | Clear, new, context |

### Manual Testing Checklist

- [ ] Press `?` to toggle help overlay
- [ ] Help overlay shows all commands
- [ ] Press `Esc` to close help
- [ ] Spinner appears during AI response
- [ ] Spinner can be cancelled with Ctrl+C
- [ ] `/help` shows help in chat
- [ ] `/clear` clears conversation
- [ ] `/new` starts fresh session
- [ ] `/context` shows current context items
- [ ] Status bar shows git branch
- [ ] Status bar shows token estimate
- [ ] Status bar shows current model
- [ ] Ctrl+L clears conversation
- [ ] Ctrl+N starts new session
- [ ] All keyboard shortcuts work

---

## Acceptance Criteria

1. **Help:** Help overlay displays and toggles correctly
2. **Keyboard:** All shortcuts work as documented
3. **Spinner:** Visual feedback during AI calls
4. **Tokens:** Token count displayed in status bar
5. **Session:** Clear and new session work
6. **Context:** Can view current context items
7. **Polish:** UI feels responsive and complete

---

## Definition of Done

- [ ] Help overlay implemented
- [ ] All keyboard shortcuts working
- [ ] Spinner during AI calls
- [ ] Token counting displayed
- [ ] Session management commands
- [ ] Context viewer working
- [ ] Unit tests passing
- [ ] Manual testing completed
- [ ] UI polish applied
- [ ] Committed to version control

---

## Notes

- Help overlay should be centered and not too large
- Spinner should not block key input (for cancel)
- Token estimation is approximate - consider tiktoken for accuracy
- Status bar should gracefully truncate on narrow terminals
