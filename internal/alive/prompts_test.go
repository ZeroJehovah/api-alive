package alive

import "testing"

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
