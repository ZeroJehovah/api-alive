package alive

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintHumanAlignedPrintsOneSummaryLine(t *testing.T) {
	var buf bytes.Buffer
	PrintHumanAligned(&buf, Result{
		Model:      "gpt-5",
		Success:    false,
		DurationMS: 1234,
		Error:      "Reconnecting...",
		Output:     "Reconnecting...\nNO",
		Prompt:     "Say OK.",
		Expected:   "OK",
	}, len("longer-model"))

	got := buf.String()
	fields := strings.Fields(got)
	if len(fields) != 3 || fields[0] != "gpt-5" || fields[1] != "1234ms" || fields[2] != "failed" {
		t.Fatalf("got fields %#v from %q", fields, got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("got %q, want exactly one line", got)
	}
	if idx := strings.Index(got, "1234ms"); idx != len("longer-model")+2+4 {
		t.Fatalf("duration starts at index %d in %q", idx, got)
	}
	for _, unexpected := range []string{"Reconnecting", "prompt=", "expected=", "output:", "error="} {
		if strings.Contains(got, unexpected) {
			t.Fatalf("output contains %q: %q", unexpected, got)
		}
	}
}
