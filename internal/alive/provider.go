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

type DirectCommandProvider interface {
	CommandContext(ctx context.Context, model string, prompt PromptCase) *exec.Cmd
}

type CodexProvider struct{ Command string }

type ClaudeProvider struct{ Command string }

type CodexWSLProvider struct {
	WSLCommand   string
	Distribution string
	CodexCommand string
}

func NewProvider(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "codex":
		return CodexProvider{Command: cfg.CodexCommand}, nil
	case "codex-wsl":
		return CodexWSLProvider{
			WSLCommand:   cfg.WSLCommand,
			Distribution: strings.TrimSpace(cfg.WSLDistro),
			CodexCommand: cfg.CodexCommand,
		}, nil
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

func (p CodexWSLProvider) Name() string { return "codex-wsl" }

func (p CodexWSLProvider) ShellCommand(model string, prompt PromptCase) string {
	parts := []string{shellQuote(p.wslCommand())}
	for _, arg := range p.wslArgs(model, prompt) {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

func (p CodexWSLProvider) CommandContext(ctx context.Context, model string, prompt PromptCase) *exec.Cmd {
	return exec.CommandContext(ctx, p.wslCommand(), p.wslArgs(model, prompt)...)
}

func (p CodexWSLProvider) wslCommand() string {
	command := strings.TrimSpace(p.WSLCommand)
	if command == "" {
		return "wsl.exe"
	}
	return command
}

func (p CodexWSLProvider) codexCommand() string {
	command := strings.TrimSpace(p.CodexCommand)
	if command == "" {
		return "codex"
	}
	return command
}

func (p CodexWSLProvider) wslArgs(model string, prompt PromptCase) []string {
	args := make([]string, 0, 7)
	if distro := strings.TrimSpace(p.Distribution); distro != "" {
		args = append(args, "-d", distro)
	}
	return append(args, "--", "sh", "-lc", p.wslScript(model, prompt))
}

func (p CodexWSLProvider) wslScript(model string, prompt PromptCase) string {
	codexCommand := p.codexCommand()
	codexArgs := strings.Join([]string{
		posixShellQuote(codexCommand),
		"exec",
		"--model", posixShellQuote(model),
		"--skip-git-repo-check",
		"--ephemeral",
		posixShellQuote(prompt.Input),
	}, " ")
	return strings.Join([]string{
		"tmp=$(mktemp -d)",
		`trap 'rm -rf "$tmp"' EXIT`,
		`cd "$tmp"`,
		"exec " + codexArgs,
	}, "; ")
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
	return posixShellQuote(s)
}

func posixShellQuote(s string) string {
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
