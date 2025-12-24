# Phase 2: Core Functionality

**Timeline:** Week 2  
**Goal:** Implement interactive text input, scrollable message viewport, markdown rendering, and configuration system

---

## Overview

This phase adds the core interactive elements: users can type messages, view them in a scrollable area with markdown rendering, and configure the app via YAML files.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Multi-line text input | Textarea component for user prompts | P0 |
| Scrollable viewport | View message history with scrolling | P0 |
| Markdown rendering | Render AI responses with Glamour | P0 |
| Configuration loading | Viper-based config from YAML | P0 |
| Message display | Show user/assistant messages with styling | P0 |
| Keyboard navigation | Scroll with arrow keys, page up/down | P1 |

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/ui/components/input.go` | Textarea wrapper component |
| `internal/ui/components/messages.go` | Message list rendering |
| `internal/ui/components/viewport.go` | Scrollable viewport wrapper |
| `internal/config/config.go` | Viper configuration loading |
| `internal/config/types.go` | Configuration struct definitions |
| `config.example.yaml` | Example configuration file |

### Modified Files

| File | Changes |
|------|---------|
| `internal/ui/model.go` | Add textarea, viewport, messages state |
| `internal/ui/styles.go` | Add message styles (user/assistant) |
| `internal/app/app.go` | Load config before starting TUI |

---

## Detailed File Specifications

### 1. `internal/config/types.go`

```go
package config

type Config struct {
    Provider  string              `mapstructure:"provider"`
    Providers map[string]Provider `mapstructure:"providers"`
    UI        UIConfig            `mapstructure:"ui"`
    System    SystemConfig        `mapstructure:"system"`
}

type Provider struct {
    APIKey  string `mapstructure:"api_key"`
    BaseURL string `mapstructure:"base_url"`
    Model   string `mapstructure:"model"`
}

type UIConfig struct {
    Theme              string `mapstructure:"theme"`
    WordWrap           int    `mapstructure:"word_wrap"`
    ShowTokens         bool   `mapstructure:"show_tokens"`
    SyntaxHighlighting bool   `mapstructure:"syntax_highlighting"`
}

type SystemConfig struct {
    Prompt string `mapstructure:"system_prompt"`
}
```

**Requirements:**
- Support multiple providers
- Environment variable expansion for API keys
- Sensible defaults for all fields

---

### 2. `internal/config/config.go`

```go
package config

import (
    "os"
    "github.com/spf13/viper"
)

var cfg *Config

func Load() (*Config, error) {
    v := viper.New()
    
    // Defaults
    v.SetDefault("provider", "ollama")
    v.SetDefault("ui.theme", "dark")
    v.SetDefault("ui.word_wrap", 80)
    v.SetDefault("ui.show_tokens", true)
    v.SetDefault("ui.syntax_highlighting", true)
    
    // Config paths
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
    }
    
    // Unmarshal
    cfg = &Config{}
    if err := v.Unmarshal(cfg); err != nil {
        return nil, err
    }
    
    // Expand environment variables in API keys
    for name, provider := range cfg.Providers {
        provider.APIKey = os.ExpandEnv(provider.APIKey)
        cfg.Providers[name] = provider
    }
    
    return cfg, nil
}

func Get() *Config {
    return cfg
}
```

**Requirements:**
- Search multiple config paths
- Fall back to defaults if no config file
- Expand `${ENV_VAR}` in API key fields
- Thread-safe global access via `Get()`

---

### 3. `internal/ui/components/input.go`

```go
package components

import (
    "github.com/charmbracelet/bubbles/textarea"
    tea "github.com/charmbracelet/bubbletea"
)

type Input struct {
    textarea textarea.Model
    focused  bool
}

func NewInput() Input {
    ta := textarea.New()
    ta.Placeholder = "Type your message... (Enter to send)"
    ta.Focus()
    ta.CharLimit = 4000
    ta.SetWidth(80)
    ta.SetHeight(3)
    ta.ShowLineNumbers = false
    
    return Input{
        textarea: ta,
        focused:  true,
    }
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
    var cmd tea.Cmd
    i.textarea, cmd = i.textarea.Update(msg)
    return i, cmd
}

func (i Input) View() string {
    return i.textarea.View()
}

func (i Input) Value() string {
    return i.textarea.Value()
}

func (i *Input) Reset() {
    i.textarea.Reset()
}

func (i *Input) SetWidth(w int) {
    i.textarea.SetWidth(w)
}

func (i *Input) Focus() tea.Cmd {
    return i.textarea.Focus()
}

func (i *Input) Blur() {
    i.textarea.Blur()
}
```

**Requirements:**
- Wrap bubbles/textarea
- Support multi-line input
- Configurable width based on terminal size
- Clear/reset functionality

---

### 4. `internal/ui/components/messages.go`

```go
package components

import (
    "fmt"
    "time"
    
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/lipgloss"
)

type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
)

type Message struct {
    Role      Role
    Content   string
    Timestamp time.Time
}

type Messages struct {
    items    []Message
    renderer *glamour.TermRenderer
    width    int
}

func NewMessages(width int) Messages {
    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(width),
    )
    
    return Messages{
        items:    []Message{},
        renderer: r,
        width:    width,
    }
}

func (m *Messages) Add(role Role, content string) {
    m.items = append(m.items, Message{
        Role:      role,
        Content:   content,
        Timestamp: time.Now(),
    })
}

func (m *Messages) Clear() {
    m.items = []Message{}
}

func (m Messages) Render() string {
    var output string
    
    for _, msg := range m.items {
        switch msg.Role {
        case RoleUser:
            output += m.renderUserMessage(msg)
        case RoleAssistant:
            output += m.renderAssistantMessage(msg)
        }
        output += "\n"
    }
    
    return output
}

func (m Messages) renderUserMessage(msg Message) string {
    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#3C3C3C")).
        Padding(0, 1).
        MarginBottom(1)
    
    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#7D56F4")).
        Render("You")
    
    return header + "\n" + style.Render(msg.Content)
}

func (m Messages) renderAssistantMessage(msg Message) string {
    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00D4AA")).
        Render("Assistant")
    
    // Render markdown
    rendered, err := m.renderer.Render(msg.Content)
    if err != nil {
        rendered = msg.Content
    }
    
    return header + "\n" + rendered
}

func (m *Messages) SetWidth(w int) {
    m.width = w
    m.renderer, _ = glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(w),
    )
}
```

**Requirements:**
- Store message history
- Render user messages with simple styling
- Render assistant messages with Glamour (markdown)
- Support dynamic width for terminal resize

---

### 5. `internal/ui/components/viewport.go`

```go
package components

import (
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
)

type Viewport struct {
    viewport viewport.Model
    ready    bool
}

func NewViewport(width, height int) Viewport {
    vp := viewport.New(width, height)
    vp.YPosition = 0
    
    return Viewport{
        viewport: vp,
        ready:    true,
    }
}

func (v Viewport) Update(msg tea.Msg) (Viewport, tea.Cmd) {
    var cmd tea.Cmd
    v.viewport, cmd = v.viewport.Update(msg)
    return v, cmd
}

func (v Viewport) View() string {
    return v.viewport.View()
}

func (v *Viewport) SetContent(content string) {
    v.viewport.SetContent(content)
}

func (v *Viewport) SetSize(width, height int) {
    v.viewport.Width = width
    v.viewport.Height = height
}

func (v *Viewport) GotoBottom() {
    v.viewport.GotoBottom()
}

func (v Viewport) ScrollPercent() float64 {
    return v.viewport.ScrollPercent()
}
```

**Requirements:**
- Wrap bubbles/viewport
- Auto-scroll to bottom on new messages
- Support manual scrolling
- Track scroll position for status bar

---

### 6. Updated `internal/ui/model.go`

```go
package ui

import (
    "github.com/yourusername/flux/internal/ui/components"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type Model struct {
    // Components
    input    components.Input
    viewport components.Viewport
    messages components.Messages
    
    // State
    width    int
    height   int
    ready    bool
    quitting bool
}

func NewModel() Model {
    return Model{
        input:    components.NewInput(),
        messages: components.NewMessages(80),
    }
}

func (m Model) Init() tea.Cmd {
    return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc":
            m.quitting = true
            return m, tea.Quit
        case "enter":
            if value := m.input.Value(); value != "" {
                m.messages.Add(components.RoleUser, value)
                m.input.Reset()
                m.viewport.SetContent(m.messages.Render())
                m.viewport.GotoBottom()
            }
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.handleResize()
        m.ready = true
    }
    
    // Update components
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    cmds = append(cmds, cmd)
    
    m.viewport, cmd = m.viewport.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m *Model) handleResize() {
    headerHeight := 3
    statusHeight := 1
    inputHeight := 5
    
    viewportHeight := m.height - headerHeight - statusHeight - inputHeight
    if viewportHeight < 1 {
        viewportHeight = 1
    }
    
    m.viewport.SetSize(m.width-4, viewportHeight)
    m.input.SetWidth(m.width - 4)
    m.messages.SetWidth(m.width - 6)
}

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
        m.viewport.View(),
        m.input.View(),
        m.renderStatusBar(),
    )
}
```

**Requirements:**
- Integrate all components
- Handle Enter key to send messages
- Update viewport content when messages change
- Properly size components on resize

---

### 7. `config.example.yaml`

```yaml
# Flux CLI Configuration
# Copy to ~/.config/flux/config.yaml

# Default AI provider
provider: ollama

# Provider configurations
providers:
  ollama:
    base_url: http://localhost:11434/v1
    model: codellama:13b

  openrouter:
    api_key: ${OPENROUTER_API_KEY}
    base_url: https://openrouter.ai/api/v1
    model: anthropic/claude-3-haiku

  groq:
    api_key: ${GROQ_API_KEY}
    base_url: https://api.groq.com/openai/v1
    model: llama-3.1-70b-versatile

# UI preferences
ui:
  theme: dark
  word_wrap: 80
  show_tokens: true
  syntax_highlighting: true

# System prompt
system_prompt: |
  You are a helpful AI coding assistant. You help users with programming tasks,
  code reviews, debugging, and explaining code concepts. Be concise and practical.
```

---

## Testing

### Unit Tests

| Test File | Tests |
|-----------|-------|
| `internal/config/config_test.go` | Load from file, defaults, env expansion |
| `internal/ui/components/input_test.go` | Value, Reset, SetWidth |
| `internal/ui/components/messages_test.go` | Add, Clear, Render |
| `internal/ui/components/viewport_test.go` | SetContent, scrolling |

### Integration Tests

| Test | Description |
|------|-------------|
| Config loading | Create temp config, verify loading |
| Message flow | Add user message, verify in viewport |

### Manual Testing Checklist

- [ ] Type multi-line message in textarea
- [ ] Press Enter to send message
- [ ] User message appears in viewport with styling
- [ ] Viewport scrolls to show new messages
- [ ] Arrow keys scroll viewport when focused
- [ ] Page Up/Down work for large message lists
- [ ] Config loads from `~/.config/flux/config.yaml`
- [ ] Missing config uses defaults without error
- [ ] Environment variables expand in API keys

### Test Commands

```bash
# Run all tests
go test ./...

# Test config package specifically
go test -v ./internal/config/...

# Test with race detector
go test -race ./...
```

---

## Dependencies

```bash
# New dependencies for Phase 2
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/glamour
go get github.com/spf13/viper
```

---

## Acceptance Criteria

1. **Input:** Multi-line textarea works correctly
2. **Messages:** User messages display with proper styling
3. **Markdown:** Test markdown renders (headers, code blocks, lists)
4. **Scrolling:** Viewport scrolls with keyboard
5. **Config:** Loads from file and falls back to defaults
6. **Env vars:** API keys expand from environment
7. **Resize:** All components resize properly

---

## Definition of Done

- [ ] All new files created
- [ ] Existing files updated
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] Manual testing completed
- [ ] config.example.yaml documented
- [ ] Committed to version control

---

## Notes

- AI responses are still mocked (hardcoded) - real AI comes in Phase 3
- Focus on getting the message flow working smoothly
- Glamour auto-detects light/dark terminal theme
- Textarea should auto-focus on launch
