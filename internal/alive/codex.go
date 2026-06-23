package alive

import (
	"context"
	"os/exec"
	"strings"
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

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shellForContext(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-lc", command)
}
