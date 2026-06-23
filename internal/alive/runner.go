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
	Provider     string        `json:"provider"`
	Model        string        `json:"model"`
	Success      bool          `json:"success"`
	Attempts     int           `json:"attempts"`
	ExitCode     *int          `json:"exit_code,omitempty"`
	Error        string        `json:"error,omitempty"`
	Duration     time.Duration `json:"-"`
	DurationMS   int64         `json:"duration_ms"`
	Prompt       string        `json:"prompt"`
	Expected     string        `json:"expected"`
	Output       string        `json:"output,omitempty"`
	TempDir      string        `json:"temp_dir,omitempty"`
	ShellCommand string        `json:"shell_command,omitempty"`
}

type Runner struct {
	Config   Config
	Provider Provider
	Prompts  []PromptCase
	DryRun   bool
}

func (r Runner) Run(ctx context.Context) (<-chan Result, error) {
	r.Config.ApplyDefaults()
	if err := r.Config.Validate(); err != nil {
		return nil, err
	}
	if r.Provider == nil {
		return nil, errors.New("provider is nil")
	}
	if len(r.Prompts) == 0 {
		return nil, errors.New("prompts are empty")
	}

	results := make(chan Result)
	var wg sync.WaitGroup
	seed := time.Now().UnixNano()

	for i, model := range r.Config.Models {
		model := model
		rng := rand.New(rand.NewSource(seed + int64(i)))
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- r.runModel(ctx, model, rng)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	return results, nil
}

func (r Runner) runModel(parent context.Context, model string, rng *rand.Rand) Result {
	started := time.Now()
	res := Result{
		Provider: r.Provider.Name(),
		Model:    model,
	}

	tmp, err := os.MkdirTemp("", "api-alive-*")
	if err != nil {
		res.Error = err.Error()
		res.Duration = time.Since(started)
		res.DurationMS = res.Duration.Milliseconds()
		return res
	}
	defer os.RemoveAll(tmp)
	res.TempDir = tmp

	for attempt := 1; attempt <= r.Config.LoopCount; attempt++ {
		prompt := r.Prompts[rng.Intn(len(r.Prompts))]
		res = r.runAttempt(parent, model, prompt, tmp, started, attempt)
		if res.Success {
			return res
		}
	}
	return res
}

func (r Runner) runAttempt(parent context.Context, model string, prompt PromptCase, tmp string, started time.Time, attempt int) Result {
	res := Result{
		Provider: r.Provider.Name(),
		Model:    model,
		Attempts: attempt,
		Prompt:   prompt.Input,
		Expected: prompt.Expected,
		TempDir:  tmp,
	}

	command := r.Provider.ShellCommand(model, prompt)
	res.ShellCommand = command
	if r.DryRun {
		res.Success = true
		res.Output = "dry-run: " + command
		res.Duration = time.Since(started)
		res.DurationMS = res.Duration.Milliseconds()
		return res
	}

	ctx, cancel := context.WithTimeout(parent, r.Config.Timeout())
	defer cancel()

	cmd := shellForContext(ctx, command)
	cmd.Dir = tmp
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	res.Output = trimOutput(output, r.Config.MaxOutputChars)
	if err != nil {
		res.Success = false
		res.Error = commandFailureMessage(err, output)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code := exitErr.ExitCode()
			res.ExitCode = &code
		}
		if ctx.Err() == context.DeadlineExceeded {
			res.Error = fmt.Sprintf("timeout after %s", r.Config.Timeout())
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
