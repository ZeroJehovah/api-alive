package alive

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"syscall"
)

func CodexShellCommand(command, model string, prompt PromptCase) string {
	return strings.Join([]string{
		shellQuote(command),
		"exec",
		"--model", shellQuote(model),
		"--skip-git-repo-check",
		"--ephemeral",
		shellQuote(prompt.Input),
	}, " ")
}

func ClaudeShellCommand(command, model string, prompt PromptCase) string {
	return strings.Join([]string{
		shellQuote(command),
		"--model", shellQuote(model),
		"--print",
		"--no-session-persistence",
		shellQuote(prompt.Input),
	}, " ")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shellForContext(ctx context.Context, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			return err
		}
		return nil
	}
	return cmd
}
