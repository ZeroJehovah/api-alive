package alive

import (
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
