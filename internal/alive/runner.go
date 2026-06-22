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
		prompt := r.Prompts[rand.New(rand.NewSource(seed+int64(i))).Intn(len(r.Prompts))]
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- r.runOne(ctx, model, prompt)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	return results, nil
}

func (r Runner) runOne(parent context.Context, model string, prompt PromptCase) Result {
	started := time.Now()
	res := Result{
		Provider: r.Provider.Name(),
		Model:    model,
		Prompt:   prompt.Input,
		Expected: prompt.Expected,
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
	res.Output = trimOutput(string(outputBytes), r.Config.MaxOutputChars)
	if err != nil {
		res.Success = false
		res.Error = err.Error()
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
