package commandexecutor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v -cover ./internal/commandexecutor/...

// go test -v -cover -run TestIsCommandWhitelisted ./internal/commandexecutor
func TestIsCommandWhitelisted(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"Command is whitelisted - touch", "touch", true},
		{"Command is whitelisted - mkdir", "mkdir", true},
		{"Command is not whitelisted - rm", "rm", false},
		{"Command is not whitelisted - sudo rm", "sudo rm", false},
		{"Empty command", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommandWhitelisted(tt.command)
			assert.Equal(t, result, tt.expected)
		})
	}
}

// go test -v -cover -run TestParseCommand ./internal/commandexecutor
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCmd  string
		expectedArgs []string
	}{
		{"Valid command without args", "ls", "ls", []string{}},
		{"Valid command with args", "ls -la /tmp", "ls", []string{"-la", "/tmp"}},
		{"Command with sudo and args", "sudo ls -la /tmp", "ls", []string{"-la", "/tmp"}},
		{"Command with leading/trailing spaces", "  echo hello world  ", "echo", []string{"hello", "world"}},
		{"Command with sudo in mixed case", "Sudo echo test", "echo", []string{"test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args, err := parseCommand(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCmd, command)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

// go test -v -cover -run TestParseCommand_Errors ./internal/commandexecutor
func TestParseCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{"Command with only sudo", "sudo", "no command provided after sudo"},
		{"Empty input", "", "no command provided"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseCommand(tt.input)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.expectedError)
		})
	}
}

// go test -v -cover -run TestAddCommands ./internal/commandexecutor
func TestAddCommands(t *testing.T) {
	executor := &commandExecutor{
		commandQueue: make(chan string, 3),
	}

	tests := []struct {
		name     string
		commands []string
	}{
		{"Single command", []string{"ls"}},
		{"Multiple commands", []string{"ls", "echo", "cat /etc/passwd"}},
		{"No commands", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.AddCommands(tt.commands)
			require.NoError(t, err)

			for _, expectedCmd := range tt.commands {
				select {
				case cmd := <-executor.commandQueue:
					assert.Equal(t, expectedCmd, cmd)
				case <-time.After(1 * time.Second):
					t.Errorf("expected command %q was not added to the queue in time", expectedCmd)
				}
			}

			assert.Len(t, executor.commandQueue, 0)
		})
	}
}
