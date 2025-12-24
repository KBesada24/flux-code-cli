# Phase 5: Code Intelligence

**Timeline:** Week 5  
**Goal:** Integrate tree-sitter for intelligent code parsing, symbol extraction, and smart context building

---

## Overview

This phase adds code intelligence using tree-sitter. Instead of dumping entire files, Flux can extract specific functions, classes, and symbols. This allows for more precise context and better AI responses.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Tree-sitter integration | Multi-language code parsing | P0 |
| Symbol extraction | Extract functions, classes, types | P0 |
| `/symbols` command | List symbols in a file | P0 |
| `/read` enhancement | Smart extraction vs full file | P0 |
| `/explain` command | Explain specific symbol | P1 |
| `/doc` command | Generate docs for symbol | P1 |
| Context priority system | Smart token budget management | P1 |
| Language detection | Auto-detect file language | P2 |

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/parser/treesitter.go` | Tree-sitter wrapper |
| `internal/parser/symbols.go` | Symbol extraction logic |
| `internal/parser/languages.go` | Language grammar loaders |
| `internal/parser/queries.go` | Tree-sitter query definitions |
| `internal/context/builder.go` | Smart context builder |
| `internal/context/priority.go` | Context priority system |
| `internal/commands/code.go` | Code-related commands |

### Modified Files

| File | Changes |
|------|---------|
| `internal/commands/handler.go` | Add code command routing |
| `internal/ui/model.go` | Integrate context builder |

---

## Detailed File Specifications

### 1. `internal/parser/languages.go`

```go
package parser

import (
    "path/filepath"
    "strings"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
    "github.com/smacker/go-tree-sitter/javascript"
    "github.com/smacker/go-tree-sitter/typescript/typescript"
    "github.com/smacker/go-tree-sitter/python"
    "github.com/smacker/go-tree-sitter/rust"
    "github.com/smacker/go-tree-sitter/java"
    "github.com/smacker/go-tree-sitter/c"
    "github.com/smacker/go-tree-sitter/cpp"
    "github.com/smacker/go-tree-sitter/ruby"
    "github.com/smacker/go-tree-sitter/bash"
)

// Language represents a supported programming language
type Language struct {
    Name      string
    TSLang    *sitter.Language
    Extensions []string
    Queries   LanguageQueries
}

// LanguageQueries contains tree-sitter queries for a language
type LanguageQueries struct {
    Functions   string
    Classes     string
    Types       string
    Imports     string
    Comments    string
}

var languages = map[string]*Language{
    "go": {
        Name:       "Go",
        TSLang:     golang.GetLanguage(),
        Extensions: []string{".go"},
        Queries: LanguageQueries{
            Functions: `(function_declaration name: (identifier) @name) @func`,
            Types:     `(type_declaration (type_spec name: (type_identifier) @name)) @type`,
            Imports:   `(import_declaration) @import`,
        },
    },
    "javascript": {
        Name:       "JavaScript",
        TSLang:     javascript.GetLanguage(),
        Extensions: []string{".js", ".jsx", ".mjs"},
        Queries: LanguageQueries{
            Functions: `[
                (function_declaration name: (identifier) @name) @func
                (arrow_function) @func
                (method_definition name: (property_identifier) @name) @func
            ]`,
            Classes: `(class_declaration name: (identifier) @name) @class`,
            Imports: `(import_statement) @import`,
        },
    },
    "typescript": {
        Name:       "TypeScript",
        TSLang:     typescript.GetLanguage(),
        Extensions: []string{".ts", ".tsx"},
        Queries: LanguageQueries{
            Functions: `[
                (function_declaration name: (identifier) @name) @func
                (arrow_function) @func
                (method_definition name: (property_identifier) @name) @func
            ]`,
            Classes:   `(class_declaration name: (type_identifier) @name) @class`,
            Types:     `(interface_declaration name: (type_identifier) @name) @type`,
            Imports:   `(import_statement) @import`,
        },
    },
    "python": {
        Name:       "Python",
        TSLang:     python.GetLanguage(),
        Extensions: []string{".py"},
        Queries: LanguageQueries{
            Functions: `(function_definition name: (identifier) @name) @func`,
            Classes:   `(class_definition name: (identifier) @name) @class`,
            Imports:   `[(import_statement) (import_from_statement)] @import`,
        },
    },
    "rust": {
        Name:       "Rust",
        TSLang:     rust.GetLanguage(),
        Extensions: []string{".rs"},
        Queries: LanguageQueries{
            Functions: `(function_item name: (identifier) @name) @func`,
            Types:     `[(struct_item name: (type_identifier) @name) (enum_item name: (type_identifier) @name)] @type`,
            Imports:   `(use_declaration) @import`,
        },
    },
    "java": {
        Name:       "Java",
        TSLang:     java.GetLanguage(),
        Extensions: []string{".java"},
        Queries: LanguageQueries{
            Functions: `(method_declaration name: (identifier) @name) @func`,
            Classes:   `(class_declaration name: (identifier) @name) @class`,
            Imports:   `(import_declaration) @import`,
        },
    },
    "c": {
        Name:       "C",
        TSLang:     c.GetLanguage(),
        Extensions: []string{".c", ".h"},
        Queries: LanguageQueries{
            Functions: `(function_definition declarator: (function_declarator declarator: (identifier) @name)) @func`,
            Types:     `(struct_specifier name: (type_identifier) @name) @type`,
        },
    },
    "cpp": {
        Name:       "C++",
        TSLang:     cpp.GetLanguage(),
        Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx"},
        Queries: LanguageQueries{
            Functions: `(function_definition declarator: (function_declarator declarator: (identifier) @name)) @func`,
            Classes:   `(class_specifier name: (type_identifier) @name) @class`,
        },
    },
    "ruby": {
        Name:       "Ruby",
        TSLang:     ruby.GetLanguage(),
        Extensions: []string{".rb"},
        Queries: LanguageQueries{
            Functions: `(method name: (identifier) @name) @func`,
            Classes:   `(class name: (constant) @name) @class`,
        },
    },
    "bash": {
        Name:       "Bash",
        TSLang:     bash.GetLanguage(),
        Extensions: []string{".sh", ".bash"},
        Queries: LanguageQueries{
            Functions: `(function_definition name: (word) @name) @func`,
        },
    },
}

// GetLanguage returns the language for a file extension
func GetLanguage(filename string) *Language {
    ext := strings.ToLower(filepath.Ext(filename))
    
    for _, lang := range languages {
        for _, langExt := range lang.Extensions {
            if ext == langExt {
                return lang
            }
        }
    }
    
    return nil
}

// GetLanguageByName returns a language by its name
func GetLanguageByName(name string) *Language {
    return languages[strings.ToLower(name)]
}

// SupportedLanguages returns a list of supported language names
func SupportedLanguages() []string {
    var names []string
    for name := range languages {
        names = append(names, name)
    }
    return names
}
```

---

### 2. `internal/parser/treesitter.go`

```go
package parser

import (
    "context"
    "fmt"
    "os"
    
    sitter "github.com/smacker/go-tree-sitter"
)

// Parser wraps tree-sitter parsing functionality
type Parser struct {
    parser *sitter.Parser
}

// NewParser creates a new parser instance
func NewParser() *Parser {
    return &Parser{
        parser: sitter.NewParser(),
    }
}

// ParseFile parses a file and returns the syntax tree
func (p *Parser) ParseFile(filename string) (*ParsedFile, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    return p.ParseContent(filename, content)
}

// ParseContent parses content with a given filename for language detection
func (p *Parser) ParseContent(filename string, content []byte) (*ParsedFile, error) {
    lang := GetLanguage(filename)
    if lang == nil {
        return nil, fmt.Errorf("unsupported language for file: %s", filename)
    }
    
    p.parser.SetLanguage(lang.TSLang)
    
    tree, err := p.parser.ParseCtx(context.Background(), nil, content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse: %w", err)
    }
    
    return &ParsedFile{
        Filename: filename,
        Content:  content,
        Tree:     tree,
        Language: lang,
    }, nil
}

// ParsedFile represents a parsed source file
type ParsedFile struct {
    Filename string
    Content  []byte
    Tree     *sitter.Tree
    Language *Language
}

// RootNode returns the root node of the syntax tree
func (pf *ParsedFile) RootNode() *sitter.Node {
    return pf.Tree.RootNode()
}

// Query executes a tree-sitter query on the parsed file
func (pf *ParsedFile) Query(queryStr string) ([]*QueryMatch, error) {
    query, err := sitter.NewQuery([]byte(queryStr), pf.Language.TSLang)
    if err != nil {
        return nil, fmt.Errorf("invalid query: %w", err)
    }
    
    cursor := sitter.NewQueryCursor()
    cursor.Exec(query, pf.RootNode())
    
    var matches []*QueryMatch
    
    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }
        
        qm := &QueryMatch{
            Captures: make(map[string]*sitter.Node),
        }
        
        for _, capture := range match.Captures {
            name := query.CaptureNameForId(capture.Index)
            qm.Captures[name] = capture.Node
        }
        
        matches = append(matches, qm)
    }
    
    return matches, nil
}

// QueryMatch represents a single query match
type QueryMatch struct {
    Captures map[string]*sitter.Node
}

// GetContent returns the content of a node
func (pf *ParsedFile) GetContent(node *sitter.Node) string {
    return node.Content(pf.Content)
}

// GetNodeRange returns start and end positions
func (pf *ParsedFile) GetNodeRange(node *sitter.Node) (startLine, endLine int) {
    return int(node.StartPoint().Row) + 1, int(node.EndPoint().Row) + 1
}
```

---

### 3. `internal/parser/symbols.go`

```go
package parser

import (
    "fmt"
    "strings"
)

// SymbolType represents the type of code symbol
type SymbolType string

const (
    SymbolFunction SymbolType = "function"
    SymbolClass    SymbolType = "class"
    SymbolType     SymbolType = "type"
    SymbolImport   SymbolType = "import"
)

// Symbol represents a code symbol (function, class, etc.)
type Symbol struct {
    Name      string
    Type      SymbolType
    StartLine int
    EndLine   int
    Content   string
    Signature string // For functions: the signature line
}

// ExtractSymbols extracts all symbols from a parsed file
func ExtractSymbols(pf *ParsedFile) ([]Symbol, error) {
    var symbols []Symbol
    
    // Extract functions
    if pf.Language.Queries.Functions != "" {
        funcs, err := extractSymbolsWithQuery(pf, pf.Language.Queries.Functions, SymbolFunction)
        if err == nil {
            symbols = append(symbols, funcs...)
        }
    }
    
    // Extract classes
    if pf.Language.Queries.Classes != "" {
        classes, err := extractSymbolsWithQuery(pf, pf.Language.Queries.Classes, SymbolClass)
        if err == nil {
            symbols = append(symbols, classes...)
        }
    }
    
    // Extract types
    if pf.Language.Queries.Types != "" {
        types, err := extractSymbolsWithQuery(pf, pf.Language.Queries.Types, SymbolType)
        if err == nil {
            symbols = append(symbols, types...)
        }
    }
    
    return symbols, nil
}

func extractSymbolsWithQuery(pf *ParsedFile, queryStr string, symType SymbolType) ([]Symbol, error) {
    matches, err := pf.Query(queryStr)
    if err != nil {
        return nil, err
    }
    
    var symbols []Symbol
    
    for _, match := range matches {
        var name string
        var node *sitter.Node
        
        // Get the name capture
        if nameNode, ok := match.Captures["name"]; ok {
            name = pf.GetContent(nameNode)
        }
        
        // Get the main node (func, class, type)
        for key, n := range match.Captures {
            if key != "name" {
                node = n
                break
            }
        }
        
        if node == nil {
            continue
        }
        
        startLine, endLine := pf.GetNodeRange(node)
        content := pf.GetContent(node)
        
        // Get signature (first line for functions)
        signature := content
        if idx := strings.Index(content, "\n"); idx != -1 {
            signature = content[:idx]
        }
        
        symbols = append(symbols, Symbol{
            Name:      name,
            Type:      symType,
            StartLine: startLine,
            EndLine:   endLine,
            Content:   content,
            Signature: strings.TrimSpace(signature),
        })
    }
    
    return symbols, nil
}

// FindSymbol finds a specific symbol by name
func FindSymbol(pf *ParsedFile, name string) (*Symbol, error) {
    symbols, err := ExtractSymbols(pf)
    if err != nil {
        return nil, err
    }
    
    for _, sym := range symbols {
        if sym.Name == name {
            return &sym, nil
        }
    }
    
    return nil, fmt.Errorf("symbol not found: %s", name)
}

// FindSymbolsOfType finds all symbols of a specific type
func FindSymbolsOfType(pf *ParsedFile, symType SymbolType) ([]Symbol, error) {
    symbols, err := ExtractSymbols(pf)
    if err != nil {
        return nil, err
    }
    
    var filtered []Symbol
    for _, sym := range symbols {
        if sym.Type == symType {
            filtered = append(filtered, sym)
        }
    }
    
    return filtered, nil
}

// FormatSymbolList formats symbols as a string list
func FormatSymbolList(symbols []Symbol) string {
    var builder strings.Builder
    
    // Group by type
    funcs := filterByType(symbols, SymbolFunction)
    classes := filterByType(symbols, SymbolClass)
    types := filterByType(symbols, SymbolType)
    
    if len(funcs) > 0 {
        builder.WriteString("### Functions\n")
        for _, s := range funcs {
            builder.WriteString(fmt.Sprintf("- `%s` (lines %d-%d)\n", s.Name, s.StartLine, s.EndLine))
        }
        builder.WriteString("\n")
    }
    
    if len(classes) > 0 {
        builder.WriteString("### Classes\n")
        for _, s := range classes {
            builder.WriteString(fmt.Sprintf("- `%s` (lines %d-%d)\n", s.Name, s.StartLine, s.EndLine))
        }
        builder.WriteString("\n")
    }
    
    if len(types) > 0 {
        builder.WriteString("### Types\n")
        for _, s := range types {
            builder.WriteString(fmt.Sprintf("- `%s` (lines %d-%d)\n", s.Name, s.StartLine, s.EndLine))
        }
    }
    
    return builder.String()
}

func filterByType(symbols []Symbol, t SymbolType) []Symbol {
    var result []Symbol
    for _, s := range symbols {
        if s.Type == t {
            result = append(result, s)
        }
    }
    return result
}
```

---

### 4. `internal/context/builder.go`

```go
package context

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/yourusername/flux/internal/parser"
)

// ContextBuilder builds smart context for AI prompts
type ContextBuilder struct {
    items       []ContextItem
    maxTokens   int
    currentTokens int
    parser      *parser.Parser
}

// ContextItem represents a piece of context
type ContextItem struct {
    Type     ItemType
    Priority int
    Source   string // file path or description
    Content  string
    Tokens   int
}

type ItemType string

const (
    ItemGitDiff    ItemType = "git_diff"
    ItemGitStaged  ItemType = "git_staged"
    ItemFile       ItemType = "file"
    ItemSymbol     ItemType = "symbol"
    ItemError      ItemType = "error"
    ItemProject    ItemType = "project"
)

// Priority levels (higher = more important)
const (
    PriorityError   = 100
    PriorityGitDiff = 90
    PrioritySymbol  = 80
    PriorityFile    = 70
    PriorityProject = 50
)

// NewContextBuilder creates a new context builder
func NewContextBuilder(maxTokens int) *ContextBuilder {
    return &ContextBuilder{
        items:     []ContextItem{},
        maxTokens: maxTokens,
        parser:    parser.NewParser(),
    }
}

// AddFile adds a file to the context
func (cb *ContextBuilder) AddFile(path string) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    tokens := estimateTokens(string(content))
    
    cb.items = append(cb.items, ContextItem{
        Type:     ItemFile,
        Priority: PriorityFile,
        Source:   path,
        Content:  string(content),
        Tokens:   tokens,
    })
    
    return nil
}

// AddSymbol adds a specific symbol from a file
func (cb *ContextBuilder) AddSymbol(path, symbolName string) error {
    pf, err := cb.parser.ParseFile(path)
    if err != nil {
        return err
    }
    
    symbol, err := parser.FindSymbol(pf, symbolName)
    if err != nil {
        return err
    }
    
    content := fmt.Sprintf("// From %s (lines %d-%d)\n%s", 
        path, symbol.StartLine, symbol.EndLine, symbol.Content)
    
    tokens := estimateTokens(content)
    
    cb.items = append(cb.items, ContextItem{
        Type:     ItemSymbol,
        Priority: PrioritySymbol,
        Source:   fmt.Sprintf("%s:%s", path, symbolName),
        Content:  content,
        Tokens:   tokens,
    })
    
    return nil
}

// AddGitDiff adds git diff to context
func (cb *ContextBuilder) AddGitDiff(diff string) {
    tokens := estimateTokens(diff)
    
    cb.items = append(cb.items, ContextItem{
        Type:     ItemGitDiff,
        Priority: PriorityGitDiff,
        Source:   "git diff",
        Content:  diff,
        Tokens:   tokens,
    })
}

// AddError adds an error/stack trace to context
func (cb *ContextBuilder) AddError(errorText string) {
    tokens := estimateTokens(errorText)
    
    cb.items = append(cb.items, ContextItem{
        Type:     ItemError,
        Priority: PriorityError,
        Source:   "error",
        Content:  errorText,
        Tokens:   tokens,
    })
}

// Build returns the final context string, respecting token limits
func (cb *ContextBuilder) Build() string {
    // Sort by priority (highest first)
    sortByPriority(cb.items)
    
    var selected []ContextItem
    totalTokens := 0
    
    for _, item := range cb.items {
        if totalTokens + item.Tokens <= cb.maxTokens {
            selected = append(selected, item)
            totalTokens += item.Tokens
        }
    }
    
    // Build output
    var builder strings.Builder
    
    for _, item := range selected {
        builder.WriteString(formatContextItem(item))
        builder.WriteString("\n\n---\n\n")
    }
    
    return builder.String()
}

// GetItems returns all context items
func (cb *ContextBuilder) GetItems() []ContextItem {
    return cb.items
}

// Clear removes all context
func (cb *ContextBuilder) Clear() {
    cb.items = []ContextItem{}
    cb.currentTokens = 0
}

// TotalTokens returns the estimated total token count
func (cb *ContextBuilder) TotalTokens() int {
    total := 0
    for _, item := range cb.items {
        total += item.Tokens
    }
    return total
}

func formatContextItem(item ContextItem) string {
    var header string
    
    switch item.Type {
    case ItemGitDiff:
        header = "## Git Changes"
    case ItemGitStaged:
        header = "## Staged Changes"
    case ItemFile:
        header = fmt.Sprintf("## File: %s", item.Source)
    case ItemSymbol:
        header = fmt.Sprintf("## Symbol: %s", item.Source)
    case ItemError:
        header = "## Error"
    case ItemProject:
        header = "## Project Structure"
    default:
        header = "## Context"
    }
    
    return fmt.Sprintf("%s\n\n%s", header, item.Content)
}

func sortByPriority(items []ContextItem) {
    // Simple bubble sort (items list is small)
    for i := 0; i < len(items)-1; i++ {
        for j := 0; j < len(items)-i-1; j++ {
            if items[j].Priority < items[j+1].Priority {
                items[j], items[j+1] = items[j+1], items[j]
            }
        }
    }
}

// estimateTokens provides a rough token estimate
// Rule of thumb: ~4 chars per token for code
func estimateTokens(s string) int {
    return len(s) / 4
}
```

---

### 5. `internal/commands/code.go`

```go
package commands

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/yourusername/flux/internal/parser"
)

// ExecuteCodeCommand handles code-related slash commands
func ExecuteCodeCommand(cmd *Command) CommandResult {
    switch cmd.Name {
    case "symbols":
        return executeSymbols(cmd.Args)
    case "read":
        return executeRead(cmd.Args)
    case "explain":
        return executeExplain(cmd.Args)
    case "doc":
        return executeDoc(cmd.Args)
    case "deps":
        return executeDeps()
    case "project":
        return executeProject()
    default:
        return CommandResult{
            Error: fmt.Errorf("unknown code command: /%s", cmd.Name),
        }
    }
}

func executeSymbols(args []string) CommandResult {
    if len(args) == 0 {
        return CommandResult{
            Error: fmt.Errorf("usage: /symbols <file>"),
        }
    }
    
    filename := args[0]
    
    p := parser.NewParser()
    pf, err := p.ParseFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    symbols, err := parser.ExtractSymbols(pf)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    output := fmt.Sprintf("## Symbols in %s\n\n%s", filename, parser.FormatSymbolList(symbols))
    
    return CommandResult{
        Output:    output,
        AddToChat: false, // Just display, don't add to context
    }
}

func executeRead(args []string) CommandResult {
    if len(args) == 0 {
        return CommandResult{
            Error: fmt.Errorf("usage: /read <file> [symbol]"),
        }
    }
    
    filename := args[0]
    
    // Check if file exists
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return CommandResult{
            Error: fmt.Errorf("file not found: %s", filename),
        }
    }
    
    // If symbol specified, extract just that symbol
    if len(args) > 1 {
        symbolName := args[1]
        
        p := parser.NewParser()
        pf, err := p.ParseFile(filename)
        if err != nil {
            // Fall back to full file if parsing fails
            return readFullFile(filename)
        }
        
        symbol, err := parser.FindSymbol(pf, symbolName)
        if err != nil {
            return CommandResult{Error: err}
        }
        
        output := fmt.Sprintf("## %s from %s (lines %d-%d)\n\n```%s\n%s\n```",
            symbol.Name, filename, symbol.StartLine, symbol.EndLine,
            getLanguageForMarkdown(filename), symbol.Content)
        
        return CommandResult{
            Output:    output,
            AddToChat: true,
        }
    }
    
    // No symbol specified - intelligent truncation
    return readFileIntelligent(filename)
}

func readFullFile(filename string) CommandResult {
    content, err := os.ReadFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    output := fmt.Sprintf("## File: %s\n\n```%s\n%s\n```",
        filename, getLanguageForMarkdown(filename), string(content))
    
    return CommandResult{
        Output:    output,
        AddToChat: true,
    }
}

func readFileIntelligent(filename string) CommandResult {
    content, err := os.ReadFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    // If file is small enough, return full content
    if len(content) < 8000 { // ~2000 tokens
        return readFullFile(filename)
    }
    
    // Try to parse and extract public symbols only
    p := parser.NewParser()
    pf, err := p.ParseFile(filename)
    if err != nil {
        // Can't parse - truncate with warning
        truncated := string(content[:8000])
        output := fmt.Sprintf("## File: %s (truncated)\n\n```%s\n%s\n...\n```\n\n*File truncated. Use `/read %s <symbol>` for specific functions.*",
            filename, getLanguageForMarkdown(filename), truncated, filename)
        
        return CommandResult{
            Output:    output,
            AddToChat: true,
        }
    }
    
    // Extract symbols and show summary + key functions
    symbols, _ := parser.ExtractSymbols(pf)
    
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("## File: %s (summary)\n\n", filename))
    builder.WriteString(parser.FormatSymbolList(symbols))
    builder.WriteString("\n\n*Use `/read " + filename + " <symbol>` to see specific implementations.*")
    
    return CommandResult{
        Output:    builder.String(),
        AddToChat: true,
    }
}

func executeExplain(args []string) CommandResult {
    if len(args) < 2 {
        return CommandResult{
            Error: fmt.Errorf("usage: /explain <file> <symbol>"),
        }
    }
    
    filename := args[0]
    symbolName := args[1]
    
    p := parser.NewParser()
    pf, err := p.ParseFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    symbol, err := parser.FindSymbol(pf, symbolName)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    prompt := fmt.Sprintf(`Please explain the following %s in detail:

\`\`\`%s
%s
\`\`\`

Explain:
1. What it does
2. How it works
3. Any important patterns or techniques used
4. Potential edge cases or gotchas`,
        symbol.Type, getLanguageForMarkdown(filename), symbol.Content)
    
    return CommandResult{
        Output:    prompt,
        AddToChat: true,
    }
}

func executeDoc(args []string) CommandResult {
    if len(args) < 2 {
        return CommandResult{
            Error: fmt.Errorf("usage: /doc <file> <symbol>"),
        }
    }
    
    filename := args[0]
    symbolName := args[1]
    
    p := parser.NewParser()
    pf, err := p.ParseFile(filename)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    symbol, err := parser.FindSymbol(pf, symbolName)
    if err != nil {
        return CommandResult{Error: err}
    }
    
    prompt := fmt.Sprintf(`Generate documentation for the following %s:

\`\`\`%s
%s
\`\`\`

Generate documentation in the appropriate format for %s (e.g., JSDoc for JavaScript, GoDoc for Go, docstrings for Python).
Include:
1. Description
2. Parameters (if applicable)
3. Return value (if applicable)
4. Example usage`,
        symbol.Type, getLanguageForMarkdown(filename), symbol.Content, pf.Language.Name)
    
    return CommandResult{
        Output:    prompt,
        AddToChat: true,
    }
}

func executeDeps() CommandResult {
    var builder strings.Builder
    builder.WriteString("## Project Dependencies\n\n")
    
    // Check for common dependency files
    depFiles := map[string]string{
        "go.mod":         "Go",
        "package.json":   "Node.js",
        "Cargo.toml":     "Rust",
        "pyproject.toml": "Python",
        "requirements.txt": "Python",
        "pom.xml":        "Java (Maven)",
        "build.gradle":   "Java (Gradle)",
    }
    
    found := false
    for file, lang := range depFiles {
        if content, err := os.ReadFile(file); err == nil {
            found = true
            builder.WriteString(fmt.Sprintf("### %s (%s)\n\n", lang, file))
            builder.WriteString("```\n")
            
            // Truncate if too long
            if len(content) > 2000 {
                builder.Write(content[:2000])
                builder.WriteString("\n... (truncated)")
            } else {
                builder.Write(content)
            }
            builder.WriteString("\n```\n\n")
        }
    }
    
    if !found {
        builder.WriteString("No dependency files found.\n")
    }
    
    return CommandResult{
        Output:    builder.String(),
        AddToChat: true,
    }
}

func executeProject() CommandResult {
    var builder strings.Builder
    builder.WriteString("## Project Structure\n\n")
    
    // Walk directory tree (limited depth)
    err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        
        // Skip hidden directories and common non-source dirs
        if info.IsDir() {
            name := info.Name()
            if strings.HasPrefix(name, ".") || 
               name == "node_modules" || 
               name == "vendor" ||
               name == "__pycache__" ||
               name == "target" ||
               name == "dist" ||
               name == "build" {
                return filepath.SkipDir
            }
        }
        
        // Calculate depth
        depth := strings.Count(path, string(os.PathSeparator))
        if depth > 3 {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }
        
        // Format output
        indent := strings.Repeat("  ", depth)
        if info.IsDir() {
            builder.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, info.Name()))
        } else {
            builder.WriteString(fmt.Sprintf("%s📄 %s\n", indent, info.Name()))
        }
        
        return nil
    })
    
    if err != nil {
        return CommandResult{Error: err}
    }
    
    return CommandResult{
        Output:    builder.String(),
        AddToChat: true,
    }
}

func getLanguageForMarkdown(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    
    langMap := map[string]string{
        ".go":   "go",
        ".js":   "javascript",
        ".jsx":  "jsx",
        ".ts":   "typescript",
        ".tsx":  "tsx",
        ".py":   "python",
        ".rs":   "rust",
        ".java": "java",
        ".c":    "c",
        ".cpp":  "cpp",
        ".h":    "c",
        ".rb":   "ruby",
        ".sh":   "bash",
        ".json": "json",
        ".yaml": "yaml",
        ".yml":  "yaml",
        ".md":   "markdown",
    }
    
    if lang, ok := langMap[ext]; ok {
        return lang
    }
    return ""
}
```

---

## Testing

### Unit Tests

| Test File | Tests |
|-----------|-------|
| `internal/parser/treesitter_test.go` | ParseFile, ParseContent, Query |
| `internal/parser/symbols_test.go` | ExtractSymbols, FindSymbol |
| `internal/parser/languages_test.go` | GetLanguage, language queries |
| `internal/context/builder_test.go` | AddFile, AddSymbol, Build, priority sorting |
| `internal/commands/code_test.go` | All code commands |

### Test Files

Create test fixtures in `testdata/`:

```
testdata/
├── sample.go
├── sample.js
├── sample.py
├── sample.rs
└── sample.ts
```

### Sample Test

```go
func TestExtractGoSymbols(t *testing.T) {
    content := []byte(`
package main

func Hello(name string) string {
    return "Hello, " + name
}

type User struct {
    Name string
    Age  int
}

func (u *User) Greet() string {
    return Hello(u.Name)
}
`)
    
    p := parser.NewParser()
    pf, err := p.ParseContent("test.go", content)
    require.NoError(t, err)
    
    symbols, err := parser.ExtractSymbols(pf)
    require.NoError(t, err)
    
    // Should find 2 functions and 1 type
    assert.Len(t, symbols, 3)
    
    // Check function names
    names := make(map[string]bool)
    for _, s := range symbols {
        names[s.Name] = true
    }
    
    assert.True(t, names["Hello"])
    assert.True(t, names["User"])
    assert.True(t, names["Greet"])
}
```

### Manual Testing Checklist

- [ ] `/symbols main.go` lists all functions/types
- [ ] `/read main.go` shows intelligent summary for large files
- [ ] `/read main.go FunctionName` extracts specific function
- [ ] `/explain main.go FunctionName` formats explain prompt
- [ ] `/doc main.go FunctionName` formats doc prompt
- [ ] `/deps` shows dependency files
- [ ] `/project` shows directory structure
- [ ] Works with Go, JavaScript, Python, TypeScript
- [ ] Graceful fallback for unsupported languages
- [ ] Large files are truncated appropriately

---

## Dependencies

```bash
go get github.com/smacker/go-tree-sitter
go get github.com/smacker/go-tree-sitter/golang
go get github.com/smacker/go-tree-sitter/javascript
go get github.com/smacker/go-tree-sitter/typescript/typescript
go get github.com/smacker/go-tree-sitter/python
go get github.com/smacker/go-tree-sitter/rust
go get github.com/smacker/go-tree-sitter/java
go get github.com/smacker/go-tree-sitter/c
go get github.com/smacker/go-tree-sitter/cpp
```

---

## Acceptance Criteria

1. **Parsing:** Successfully parse files in all supported languages
2. **Symbols:** Accurate symbol extraction (functions, classes, types)
3. **Smart read:** Large files summarized, small files shown in full
4. **Symbol extraction:** `/read file symbol` works correctly
5. **Context priority:** Higher priority items selected first
6. **Token estimation:** Reasonable token estimates
7. **Language detection:** Correct language from file extension

---

## Definition of Done

- [ ] Tree-sitter integration working
- [ ] All supported languages tested
- [ ] Symbol extraction accurate
- [ ] Context builder with priority system
- [ ] All code commands implemented
- [ ] Unit tests passing
- [ ] Manual testing completed
- [ ] Committed to version control

---

## Notes

- Tree-sitter grammars are compiled into the binary - increases size
- Consider lazy-loading language grammars
- Query strings are language-specific - maintain carefully
- Token estimation is rough - could use tiktoken for accuracy
