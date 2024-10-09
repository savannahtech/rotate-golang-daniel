package commandexecutor

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/danielboakye/filechangestracker/internal/config"
)

//go:generate mockgen -destination=../../mocks/commandexecutor/mock_commandexecutor.go -package=commandexecutormock -source=commandexecutor.go
type CommandExecutor interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	IsWorkerThreadAlive() bool
	AddCommands(commands []string) error
}

type commandExecutor struct {
	commandQueue        chan string
	appLogger           *slog.Logger
	config              *config.Config
	mu                  sync.Mutex
	workerLastHeartbeat time.Time
}

var commandWhitelist = []string{
	"touch",
	"mkdir",
}

func New(appLogger *slog.Logger, cfg *config.Config) CommandExecutor {
	return &commandExecutor{
		commandQueue: make(chan string, 100),
		appLogger:    appLogger,
		config:       cfg,
	}
}

func (f *commandExecutor) drainCommandQueue() {
	for newCmd := range f.commandQueue {
		err := f.executeCommand(newCmd)
		if err != nil {
			f.appLogger.Error("error-executing-command", slog.String("error", err.Error()))
		}
	}
}

func (f *commandExecutor) workerThread(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Heartbeat every 10 seconds
	defer func() {
		ticker.Stop()
		if f.commandQueue != nil {
			f.drainCommandQueue() // process all queued commands before shutdown
			close(f.commandQueue)
			f.commandQueue = nil
		}
	}()

	for {
		select {
		case <-ctx.Done():
			f.appLogger.Info("command-executor-shutdown")
			return
		case <-ticker.C:
			f.mu.Lock()
			f.workerLastHeartbeat = time.Now()
			f.mu.Unlock()
		case newCmd := <-f.commandQueue:
			err := f.executeCommand(newCmd)
			if err != nil {
				f.appLogger.Error("error-executing-command", slog.String("error", err.Error()))
			}
		}
	}
}

func isCommandWhitelisted(command string) bool {
	for _, allowedCmd := range commandWhitelist {
		if command == allowedCmd {
			return true
		}
	}
	return false
}

func parseCommand(input string) (command string, args []string, err error) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return command, args, fmt.Errorf("no command provided")
	}

	// Check for 'sudo' and strip it if present
	if strings.ToLower(tokens[0]) == "sudo" {
		tokens = tokens[1:]
		if len(tokens) == 0 {
			return command, args, fmt.Errorf("no command provided after sudo")
		}
	}

	command = tokens[0]
	args = tokens[1:]
	return command, args, nil
}

func (f *commandExecutor) executeCommand(input string) error {
	command, args, err := parseCommand(input)
	if err != nil {
		return err
	}

	if !isCommandWhitelisted(command) {
		return fmt.Errorf("execution blocked: command: %s is not whitelisted", command)
	}

	cmd := exec.Command(command, args...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	return nil
}

func (f *commandExecutor) Start(ctx context.Context) error {
	go f.workerThread(ctx)

	return nil
}

func (f *commandExecutor) Stop(ctx context.Context) error {
	return nil
}

func (f *commandExecutor) IsWorkerThreadAlive() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return time.Since(f.workerLastHeartbeat) < 2*time.Minute
}

func (f *commandExecutor) AddCommands(commands []string) error {
	for _, cmd := range commands {
		f.commandQueue <- cmd
	}

	return nil
}
