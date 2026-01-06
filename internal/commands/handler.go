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
	AddToChat bool // If true, add to chat as context
	Error     error
}
