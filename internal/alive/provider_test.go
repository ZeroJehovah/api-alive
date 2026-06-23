package alive

import (
	"context"
	"strings"
	"testing"
)

func TestCodexProviderCommand(t *testing.T) {
	cmd := CodexProvider{Command: "codex"}.ShellCommand("gpt-test", PromptCase{Input: "Say OK."})
	for _, want := range []string{"'codex' exec", "--model 'gpt-test'", "--skip-git-repo-check", "--ephemeral", "'Say OK.'"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command %q does not contain %q", cmd, want)
		}
	}
}

func TestClaudeProviderCommand(t *testing.T) {
	cmd := ClaudeProvider{Command: "claude"}.ShellCommand("sonnet", PromptCase{Input: "Say OK."})
	for _, want := range []string{"'claude'", "--model 'sonnet'", "--print", "'Say OK.'"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command %q does not contain %q", cmd, want)
		}
	}
}

func TestCodexWSLProviderCommand(t *testing.T) {
	provider := CodexWSLProvider{
		WSLCommand:   "wsl.exe",
		Distribution: "Ubuntu-22.04",
		CodexCommand: "codex",
	}
	prompt := PromptCase{Input: "Say OK.", Expected: "OK"}

	display := provider.ShellCommand("gpt-test", prompt)
	for _, want := range []string{"'wsl.exe'", "'-d'", "'Ubuntu-22.04'", "'sh'", "'-lc'", "codex", "--model", "gpt-test"} {
		if !strings.Contains(display, want) {
			t.Fatalf("display command %q does not contain %q", display, want)
		}
	}

	cmd := provider.CommandContext(context.Background(), "gpt-test", prompt)
	got := strings.Join(cmd.Args, "\x00")
	for _, want := range []string{"wsl.exe", "-d", "Ubuntu-22.04", "--", "sh", "-lc", "codex", "--model 'gpt-test'"} {
		if !strings.Contains(got, want) {
			t.Fatalf("direct args %q do not contain %q", got, want)
		}
	}
}
