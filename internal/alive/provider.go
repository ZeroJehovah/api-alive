package alive

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Provider interface {
	Name() string
	ShellCommand(model string, prompt PromptCase) string
}

type CodexProvider struct{ Command string }

type ClaudeProvider struct{ Command string }

func NewProvider(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "codex":
		return CodexProvider{Command: cfg.CodexCommand}, nil
	case "claude":
		return ClaudeProvider{Command: cfg.ClaudeCommand}, nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", cfg.Provider)
	}
}

func (p CodexProvider) Name() string { return "codex" }

func (p CodexProvider) ShellCommand(model string, prompt PromptCase) string {
	// Codex CLI non-interactive mode. Each run is executed inside its own temp dir by the runner.
	return strings.Join([]string{
		shellQuote(p.Command),
		"exec",
		"--model", shellQuote(model),
		"--skip-git-repo-check",
		"--ephemeral",
		shellQuote(prompt.Input),
	}, " ")
}

func (p ClaudeProvider) Name() string { return "claude" }

func (p ClaudeProvider) ShellCommand(model string, prompt PromptCase) string {
	// Claude Code-compatible placeholder adapter; kept separate so flags can evolve without touching the runner.
	return strings.Join([]string{
		shellQuote(p.Command),
		"--model", shellQuote(model),
		"--print",
		shellQuote(prompt.Input),
	}, " ")
}

func shellQuote(s string) string {
	if runtime.GOOS == "windows" {
		return strconv.Quote(s)
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shellFor(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-lc", command)
}

func shellForContext(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-lc", command)
}
