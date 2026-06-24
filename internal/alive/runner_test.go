package alive

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunnerRetriesUntilExpectedOutputMatches(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.ModelLoopCounts = map[string]int{"a": 3}
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
	cfg.ModelLoopCounts = map[string]int{"a": 2}
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

func TestRunnerEmitsAttemptEventsBeforeFinalResult(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.ModelLoopCounts = map[string]int{"a": 2}
	r := Runner{
		Config:         cfg,
		commandBuilder: staticCommand(`n=$(cat tries 2>/dev/null || echo 0); n=$((n+1)); echo "$n" > tries; if [ "$n" -lt 2 ]; then echo NO; else echo OK; fi`),
		Prompts:        []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	events, err := r.RunEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var got []Event
	for event := range events {
		got = append(got, event)
	}
	if len(got) != 3 {
		t.Fatalf("event count = %d, want 3: %#v", len(got), got)
	}
	if got[0].Type != EventAttempt || got[0].Result.Attempts != 1 || got[0].Result.Success {
		t.Fatalf("first event = %#v, want failed attempt 1", got[0])
	}
	if got[1].Type != EventAttempt || got[1].Result.Attempts != 2 || !got[1].Result.Success {
		t.Fatalf("second event = %#v, want successful attempt 2", got[1])
	}
	if got[2].Type != EventResult || got[2].Result.Attempts != 2 || len(got[2].Result.AttemptResults) != 2 {
		t.Fatalf("final event = %#v, want aggregate result with two attempts", got[2])
	}
}

func TestRunnerAttemptDurationsArePerAttempt(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.ModelLoopCounts = map[string]int{"a": 2}
	r := Runner{
		Config: cfg,
		commandBuilder: staticCommand(`n=$(cat tries 2>/dev/null || echo 0); n=$((n+1)); echo "$n" > tries; ` +
			`if [ "$n" -eq 1 ]; then sleep 1; echo NO; else echo OK; fi`),
		Prompts: []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	events, err := r.RunEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var attempts []Result
	for event := range events {
		if event.Type == EventAttempt {
			attempts = append(attempts, event.Result)
		}
	}
	if len(attempts) != 2 {
		t.Fatalf("attempt count = %d, want 2: %#v", len(attempts), attempts)
	}
	if attempts[0].DurationMS < 800 {
		t.Fatalf("first attempt duration = %dms, want at least 800ms", attempts[0].DurationMS)
	}
	if attempts[1].DurationMS >= attempts[0].DurationMS {
		t.Fatalf("second attempt duration = %dms, want per-attempt duration shorter than first attempt %dms", attempts[1].DurationMS, attempts[0].DurationMS)
	}
}

func TestRunnerCancelStopsAfterCurrentAttempt(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a"}
	cfg.ModelLoopCounts = map[string]int{"a": 3}
	cfg.TimeoutSeconds = 10
	r := Runner{
		Config:         cfg,
		commandBuilder: staticCommand("sleep 5"),
		Prompts:        []PromptCase{{Input: "Say OK.", Expected: "OK"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	timer := time.AfterFunc(50*time.Millisecond, cancel)
	defer timer.Stop()

	events, err := r.RunEvents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var got []Event
	for event := range events {
		got = append(got, event)
	}
	if len(got) != 2 {
		t.Fatalf("event count = %d, want 2: %#v", len(got), got)
	}
	if got[0].Type != EventAttempt || got[0].Result.Attempts != 1 || got[0].Result.Error != context.Canceled.Error() {
		t.Fatalf("attempt event = %#v, want canceled attempt 1", got[0])
	}
	if got[1].Type != EventResult || got[1].Result.Attempts != 1 || len(got[1].Result.AttemptResults) != 1 {
		t.Fatalf("final event = %#v, want one canceled attempt", got[1])
	}
}

func staticCommand(command string) func(string, PromptCase) string {
	return func(string, PromptCase) string { return command }
}
