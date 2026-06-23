package alive

import (
	"context"
	"strings"
	"testing"
)

func TestRunnerRetriesUntilExpectedOutputMatches(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.LoopCount = 3
	r := Runner{
		Config:         cfg,
		commandBuilder: staticCommand(`n=$(cat tries 2>/dev/null || echo 0); n=$((n+1)); echo "$n" > tries; if [ "$n" -lt 3 ]; then echo NO; else echo OK; fi`),
		Prompts:        []PromptCase{{Input: "Say OK.", Expected: "OK"}},
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
		Config:         cfg,
		commandBuilder: staticCommand("echo NO"),
		Prompts:        []PromptCase{{Input: "Say OK.", Expected: "OK"}},
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

func TestRunnerUsesLastErrorLineOnCommandFailure(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	r := Runner{
		Config: cfg,
		commandBuilder: staticCommand(strings.Join([]string{
			`printf '%s\n' 'Reading additional input from stdin...'`,
			`printf '%s\n' 'ERROR: first retry failed'`,
			`printf '%s\n' 'ERROR: exceeded retry limit, last status: 429 Too Many Requests'`,
			`exit 1`,
		}, "; ")),
		Prompts: []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	ch, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res := <-ch
	if res.Success {
		t.Fatalf("result succeeded unexpectedly: %#v", res)
	}
	if res.Error != "ERROR: exceeded retry limit, last status: 429 Too Many Requests" {
		t.Fatalf("error = %q", res.Error)
	}
	if res.ExitCode == nil || *res.ExitCode != 1 {
		t.Fatalf("exit code = %v, want 1", res.ExitCode)
	}
}

func TestRunnerChecksExpectedOutput(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	r := Runner{
		Config:         cfg,
		commandBuilder: staticCommand("echo OK"),
		Prompts:        []PromptCase{{Input: "Say OK.", Expected: "OK"}},
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
		Config:         cfg,
		commandBuilder: staticCommand("echo NO"),
		Prompts:        []PromptCase{{Input: "Say OK.", Expected: "OK"}},
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

func staticCommand(command string) func(string, PromptCase) string {
	return func(string, PromptCase) string { return command }
}
