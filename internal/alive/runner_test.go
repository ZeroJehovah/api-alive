package alive

import (
	"context"
	"strings"
	"testing"
)

type staticProvider struct {
	name    string
	command string
}

func (p staticProvider) Name() string { return p.name }

func (p staticProvider) ShellCommand(string, PromptCase) string { return p.command }

func TestRunnerDryRunReturnsOneResultPerModel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a", "b", "c"}
	cfg.LoopCount = 3
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
		if res.Attempts != 1 {
			t.Fatalf("dry-run attempts = %d, want 1", res.Attempts)
		}
		if res.Model == "" || res.Prompt == "" || res.Expected == "" || res.ShellCommand == "" {
			t.Fatalf("incomplete result: %#v", res)
		}
	}
	if count != len(cfg.Models) {
		t.Fatalf("got %d results, want %d", count, len(cfg.Models))
	}
}

func TestRunnerRetriesUntilExpectedOutputMatches(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.LoopCount = 3
	r := Runner{
		Config: cfg,
		Provider: staticProvider{
			name:    "test",
			command: `n=$(cat tries 2>/dev/null || echo 0); n=$((n+1)); echo "$n" > tries; if [ "$n" -lt 3 ]; then echo NO; else echo OK; fi`,
		},
		Prompts: []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	ch, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res := <-ch
	if !res.Success {
		t.Fatalf("result failed: %#v", res)
	}
	if res.Attempts != 3 {
		t.Fatalf("attempts = %d, want 3", res.Attempts)
	}
}

func TestRunnerReportsConfiguredAttemptsAfterAllFailures(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.LoopCount = 2
	r := Runner{
		Config:   cfg,
		Provider: staticProvider{name: "test", command: "echo NO"},
		Prompts:  []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	ch, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res := <-ch
	if res.Success {
		t.Fatalf("result succeeded unexpectedly: %#v", res)
	}
	if res.Attempts != 2 {
		t.Fatalf("attempts = %d, want 2", res.Attempts)
	}
}

func TestRunnerChecksExpectedOutput(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	r := Runner{
		Config:   cfg,
		Provider: staticProvider{name: "test", command: "echo OK"},
		Prompts:  []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	ch, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res := <-ch
	if !res.Success {
		t.Fatalf("result failed: %#v", res)
	}
}

func TestRunnerFailsWhenExpectedOutputIsMissing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	r := Runner{
		Config:   cfg,
		Provider: staticProvider{name: "test", command: "echo NO"},
		Prompts:  []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	ch, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res := <-ch
	if res.Success {
		t.Fatalf("result succeeded unexpectedly: %#v", res)
	}
	if !strings.Contains(res.Error, "expected output") {
		t.Fatalf("error %q does not explain expected-output mismatch", res.Error)
	}
}
