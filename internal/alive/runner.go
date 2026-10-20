package alive

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Model          string        `json:"model"`
	Success        bool          `json:"success"`
	Attempts       int           `json:"attempts"`
	ExitCode       *int          `json:"exit_code,omitempty"`
	Error          string        `json:"error,omitempty"`
	Duration       time.Duration `json:"-"`
	DurationMS     int64         `json:"duration_ms"`
	AttemptStarted time.Time     `json:"-"`
	UpdatedAt      string        `json:"updated_at,omitempty"`
	Prompt         string        `json:"prompt"`
	Expected       string        `json:"expected"`
	Output         string        `json:"output,omitempty"`
	AttemptResults []Result      `json:"attempt_results,omitempty"`
}

type Runner struct {
	Config         Config
	Prompts        []PromptCase
	commandBuilder func(model string, prompt PromptCase) string
	attemptSpacing time.Duration
}

type Event struct {
	Type   string `json:"type"`
	Result Result `json:"result"`
}

const (
	EventAttemptProgress = "attempt_progress"
	EventAttempt         = "attempt"
	EventResult          = "result"
)

func (r Runner) Run(ctx context.Context) (<-chan Result, error) {
	events, err := r.RunEvents(ctx)
	if err != nil {
		return nil, err
	}
	results := make(chan Result)
	go func() {
		defer close(results)
		for event := range events {
			if event.Type == EventResult {
				results <- event.Result
			}
		}
	}()
	return results, nil
}

func (r Runner) RunEvents(ctx context.Context) (<-chan Event, error) {
	r.Config.ApplyDefaults()
	if err := r.Config.Validate(); err != nil {
		return nil, err
	}
	if len(r.Prompts) == 0 {
		return nil, errors.New("prompts are empty")
	}

	events := make(chan Event)
	var wg sync.WaitGroup
	seed := time.Now().UnixNano()

	for i, model := range r.Config.Models {
		model := model
		rng := rand.New(rand.NewSource(seed + int64(i)))
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := r.runModelEvents(ctx, model, rng, events)
			events <- Event{Type: EventResult, Result: result}
		}()
	}

	go func() {
		wg.Wait()
		close(events)
	}()
	return events, nil
}

func (r Runner) runModel(parent context.Context, model string, rng *rand.Rand) Result {
	events := make(chan Event)
	go func() {
		for range events {
		}
	}()
	result := r.runModelEvents(parent, model, rng, events)
	close(events)
	return result
}

func (r Runner) runModelEvents(parent context.Context, model string, rng *rand.Rand, events chan<- Event) Result {
	started := time.Now()
	res := Result{
		Model: model,
	}

	tmp, err := os.MkdirTemp("", "api-alive-*")
	if err != nil {
		res.Error = err.Error()
		res.Duration = time.Since(started)
		res.DurationMS = res.Duration.Milliseconds()
		return res
	}
	defer os.RemoveAll(tmp)

	loopCount := r.Config.LoopCountForModel(model)
	attemptResults := make([]Result, 0)
	if loopCount > 0 {
		attemptResults = make([]Result, 0, loopCount)
	}
	for attempt := 1; loopCount == 0 || attempt <= loopCount; attempt++ {
		if attempt > 1 {
			if err := waitForNextAttempt(parent, res.AttemptStarted, r.modelAttemptSpacing()); err != nil {
				res.Error = err.Error()
				res.AttemptResults = attemptResults
				return res
			}
		}
		prompt := r.Prompts[rng.Intn(len(r.Prompts))]
		res = r.runAttempt(parent, model, prompt, tmp, attempt, func(progress Result) {
			events <- Event{Type: EventAttemptProgress, Result: progress}
		})
		attemptResults = append(attemptResults, res)
		events <- Event{Type: EventAttempt, Result: res}
		if res.Success {
			res.AttemptResults = attemptResults
			return res
		}
		if parent.Err() != nil {
			res.Error = parent.Err().Error()
			res.AttemptResults = attemptResults
			return res
		}
	}
	res.AttemptResults = attemptResults
	return res
}

func (r Runner) runAttempt(parent context.Context, model string, prompt PromptCase, tmp string, attempt int, emitProgress func(Result)) Result {
	started := time.Now()
	res := Result{
		Model:          model,
		Attempts:       attempt,
		AttemptStarted: started,
		Prompt:         prompt.Input,
		Expected:       prompt.Expected,
	}

	ctx, cancel := context.WithTimeout(parent, r.Config.Timeout())
	defer cancel()

	reportFirstError := func(message string) {
		if emitProgress == nil {
			return
		}
		progress := res
		progress.Success = false
		progress.Error = message
		progress.Duration = time.Since(started)
		progress.DurationMS = progress.Duration.Milliseconds()
		emitProgress(progress)
	}
	var executed attemptExecution
	if r.commandBuilder != nil {
		executed = runShellAttempt(ctx, r.shellCommand(model, prompt), tmp)
	} else if r.Config.Provider == ProviderClaude {
		executed = runClaudeHTTPAttempt(ctx, r.Config.BaseURL, model, prompt, reportFirstError)
	} else {
		executed = runCodexHTTPAttempt(ctx, r.Config.BaseURL, model, prompt, reportFirstError)
	}
	res.Output = trimOutput(executed.Output, r.Config.MaxOutputChars)
	res.ExitCode = executed.ExitCode
	if executed.Err != nil {
		res.Success = false
		res.Error = executed.LastError
		if res.Error == "" {
			res.Error = commandFailureMessage(executed.Err, executed.Output)
		}
		if ctx.Err() == context.DeadlineExceeded {
			res.Error = timeoutAttemptMessage(r.Config.Timeout(), executed.FirstError, executed.LastError)
		}
		if ctx.Err() == context.Canceled {
			res.Error = context.Canceled.Error()
		}
	} else {
		if outputContainsExpected(res.Output, prompt.Expected) {
			res.Success = true
		} else {
			res.Success = false
			res.Error = fmt.Sprintf("expected output %q not found", prompt.Expected)
		}
	}
	res.Duration = time.Since(started)
	res.DurationMS = res.Duration.Milliseconds()
	return res
}

func runShellAttempt(ctx context.Context, command, dir string) attemptExecution {
	cmd := shellForContext(ctx, command)
	cmd.Dir = dir
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	result := attemptExecution{Output: output, Err: err}
	if err != nil {
		result.LastError = commandFailureMessage(err, output)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code := exitErr.ExitCode()
			result.ExitCode = &code
		}
	}
	return result
}

func timeoutAttemptMessage(timeout time.Duration, firstError, lastError string) string {
	message := fmt.Sprintf("timeout after %s", timeout)
	if firstError != "" {
		message += "; first error: " + firstError
	}
	if lastError != "" && lastError != firstError {
		message += "; last error: " + lastError
	}
	return message
}

func (r Runner) modelAttemptSpacing() time.Duration {
	if r.attemptSpacing > 0 {
		return r.attemptSpacing
	}
	return r.Config.BatchMinInterval()
}

func waitForNextAttempt(ctx context.Context, previousStarted time.Time, spacing time.Duration) error {
	if spacing <= 0 || previousStarted.IsZero() {
		return nil
	}
	wait := time.Until(previousStarted.Add(spacing))
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r Runner) shellCommand(model string, prompt PromptCase) string {
	return r.commandBuilder(model, prompt)
}

func commandFailureMessage(err error, output string) string {
	if msg := lastErrorLine(output); msg != "" {
		return msg
	}
	if msg := lastOutputLine(output); msg != "" {
		return msg
	}
	return err.Error()
}

func lastErrorLine(output string) string {
	var last string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "ERROR:") {
			last = line
		}
	}
	return last
}

func lastOutputLine(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}

func outputContainsExpected(output, expected string) bool {
	output = strings.TrimSpace(output)
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}
	if output == expected {
		return true
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == expected {
			return true
		}
	}
	return strings.Contains(output, expected)
}

func trimOutput(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 20 {
		return s[:max]
	}
	return s[:max-20] + "...<output truncated>"
}
