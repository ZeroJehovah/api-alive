package alive

import (
	"strings"
	"testing"
)

func TestDefaultPromptsCountAndShape(t *testing.T) {
	if got := len(DefaultPrompts); got != 100 {
		t.Fatalf("len(DefaultPrompts) = %d, want 100", got)
	}
	for i, p := range DefaultPrompts {
		if p.Input == "" || p.Expected == "" {
			t.Fatalf("prompt %d has empty input or expected: %#v", i, p)
		}
		if len(p.Input) > 80 || len(p.Expected) > 40 {
			t.Fatalf("prompt %d is not short enough: %#v", i, p)
		}
	}
}

func TestDefaultPromptsAvoidProbeLikePhrasing(t *testing.T) {
	for i, p := range DefaultPrompts {
		input := strings.ToLower(p.Input)
		for _, phrase := range []string{
			"say ok",
			"say hello",
			"say hi",
			"reply ping",
			"reply pong",
			"return ok",
			"return ready",
			"return alive",
			"echo ",
			"type x",
			"type y",
			"type z",
		} {
			if strings.Contains(input, phrase) {
				t.Fatalf("prompt %d uses probe-like phrase %q: %#v", i, phrase, p)
			}
		}
	}
}
