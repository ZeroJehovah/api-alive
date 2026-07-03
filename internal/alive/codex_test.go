package alive

import (
	"strings"
	"testing"
)

func TestCodexShellCommand(t *testing.T) {
	cmd := CodexShellCommand("codex", "gpt-test", PromptCase{Input: "Say OK."})
	for _, want := range []string{"'codex' exec", "--model 'gpt-test'", "--skip-git-repo-check", "--ephemeral", "'Say OK.'"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command %q does not contain %q", cmd, want)
		}
	}
}

func TestClaudeShellCommand(t *testing.T) {
	cmd := ClaudeShellCommand("claude", "sonnet", PromptCase{Input: "Say OK."})
	for _, want := range []string{"'claude' --model 'sonnet'", "--print", "--no-session-persistence", "'Say OK.'"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command %q does not contain %q", cmd, want)
		}
	}
}
