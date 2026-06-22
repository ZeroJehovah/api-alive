package alive

import (
	"context"
	"testing"
)

func TestRunnerDryRunReturnsOneResultPerModel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a", "b", "c"}
	r := Runner{Config: cfg, Provider: CodexProvider{Command: "codex"}, Prompts: DefaultPrompts, DryRun: true}
	ch, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for res := range ch {
		count++
		if !res.Success {
			t.Fatalf("dry-run result failed: %#v", res)
		}
		if res.Model == "" || res.Prompt == "" || res.Expected == "" || res.ShellCommand == "" {
			t.Fatalf("incomplete result: %#v", res)
		}
	}
	if count != len(cfg.Models) {
		t.Fatalf("got %d results, want %d", count, len(cfg.Models))
	}
}
