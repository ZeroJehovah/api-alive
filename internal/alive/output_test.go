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
		Attempts:   3,
		DurationMS: 1234,
		Error:      "abcdefghijklmnopqrstuvwxyz1234567890",
		Output:     "Reconnecting...\nNO",
		Prompt:     "Say OK.",
		Expected:   "OK",
	}, len("longer-model"))

	got := buf.String()
	if !strings.HasSuffix(strings.TrimSpace(got), "failed") {
		t.Fatalf("status is not last in %q", got)
	}
	for _, expected := range []string{"gpt-5", "1234ms", "attempts=3", `error="abcdefghijklmnopqrstuvwxyz1..."`} {
		if !strings.Contains(got, expected) {
			t.Fatalf("output missing %q: %q", expected, got)
		}
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("got %q, want exactly one line", got)
	}
	if idx := strings.Index(got, "1234ms"); idx != len("longer-model")+2+4 {
		t.Fatalf("duration starts at index %d in %q", idx, got)
	}
	for _, unexpected := range []string{"Reconnecting", "prompt=", "expected=", "output:"} {
		if strings.Contains(got, unexpected) {
			t.Fatalf("output contains %q: %q", unexpected, got)
		}
	}
}

func TestPrintHumanAlignedPrintsSuccessStatusLastWithoutError(t *testing.T) {
	var buf bytes.Buffer
	PrintHumanAligned(&buf, Result{
		Model:      "gpt-5",
		Success:    true,
		Attempts:   1,
		DurationMS: 99,
	}, len("gpt-5"))

	got := buf.String()
	if !strings.HasSuffix(strings.TrimSpace(got), "success") {
		t.Fatalf("status is not last in %q", got)
	}
	if strings.Contains(got, "error=") {
		t.Fatalf("success output contains error field: %q", got)
	}
}
