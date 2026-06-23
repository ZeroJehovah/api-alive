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
