package alive

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintHumanAlignedPrintsOneSummaryLine(t *testing.T) {
	var buf bytes.Buffer
	longError := strings.Repeat("x", maxHumanErrorChars+10)
	PrintHumanAligned(&buf, Result{
		Model:      "gpt-5",
		Success:    false,
		Attempts:   3,
		DurationMS: 1234,
		Error:      longError,
		Output:     "Reconnecting...\nNO",
		Prompt:     "Say OK.",
		Expected:   "OK",
	}, len("longer-model"))

	got := buf.String()
	if !strings.HasSuffix(strings.TrimSpace(got), "failed") {
		t.Fatalf("status is not last in %q", got)
	}
	for _, expected := range []string{"❌", "gpt-5", "1.234s", "attempts=3", `error="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx..."`} {
		if !strings.Contains(got, expected) {
			t.Fatalf("output missing %q: %q", expected, got)
		}
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("got %q, want exactly one line", got)
	}
	if !strings.HasPrefix(got, "❌ gpt-5") {
		t.Fatalf("output does not start with failure emoji and model: %q", got)
	}
	if strings.Contains(got, "ms") {
		t.Fatalf("output still contains millisecond unit: %q", got)
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
	if !strings.Contains(got, "✅") {
		t.Fatalf("success output missing emoji: %q", got)
	}
	if !strings.Contains(got, "0.099s") {
		t.Fatalf("success output missing second duration: %q", got)
	}
}

func TestPrintHumanAlignedKeepsProviderErrorDetail(t *testing.T) {
	var buf bytes.Buffer
	PrintHumanAligned(&buf, Result{
		Model:      "hotaruapi/gpt-5.5",
		Success:    false,
		Attempts:   1,
		DurationMS: 21735,
		Error:      "ERROR: exceeded retry limit, last status: 429 Too Many Requests",
	}, len("hotaruapi/gpt-5.5"))

	got := buf.String()
	if !strings.Contains(got, `error="ERROR: exceeded retry limit, last status: 429 Too Many Requests"`) {
		t.Fatalf("output missing provider error detail: %q", got)
	}
	if !strings.HasSuffix(strings.TrimSpace(got), "failed") {
		t.Fatalf("status is not last in %q", got)
	}
}
