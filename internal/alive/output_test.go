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
	for _, expected := range []string{"❌", "gpt-5", "1.234s", "attempts=3", `error=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx...`} {
		if !strings.Contains(got, expected) {
			t.Fatalf("output missing %q: %q", expected, got)
		}
	}
	if !strings.Contains(got, `attempts=3  failed   error=`) {
		t.Fatalf("failure status is not before error detail: %q", got)
	}
	if strings.Contains(got, `error="`) {
		t.Fatalf("failure detail is quoted: %q", got)
	}
	if strings.Index(got, "failed") > strings.Index(got, "error=") {
		t.Fatalf("failure detail appears before status: %q", got)
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

func TestPrintHumanAlignedAlignsStatusesBeforeFailureDetail(t *testing.T) {
	var success, failure bytes.Buffer
	width := len("longer-model")
	PrintHumanAligned(&success, Result{
		Model:      "short",
		Success:    true,
		Attempts:   1,
		DurationMS: 1000,
	}, width)
	PrintHumanAligned(&failure, Result{
		Model:      "longer-model",
		Success:    false,
		Attempts:   1,
		DurationMS: 1000,
		Error:      "ERROR: no quota",
	}, width)

	successLine := success.String()
	failureLine := failure.String()
	successIdx := strings.Index(successLine, "success")
	failedIdx := strings.Index(failureLine, "failed")
	if successIdx < 0 || failedIdx < 0 {
		t.Fatalf("missing status in success=%q failure=%q", successLine, failureLine)
	}
	if successIdx != failedIdx {
		t.Fatalf("status columns differ: success=%d failure=%d\nsuccess=%qfailure=%q", successIdx, failedIdx, successLine, failureLine)
	}
	if strings.Index(failureLine, "failed") > strings.Index(failureLine, "error=") {
		t.Fatalf("failure detail appears before status: %q", failureLine)
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
	if !strings.Contains(got, `error=ERROR: exceeded retry limit, last status: 429 Too Many Requests`) {
		t.Fatalf("output missing provider error detail: %q", got)
	}
	if !strings.Contains(got, `failed   error=ERROR: exceeded retry limit, last status: 429 Too Many Requests`) {
		t.Fatalf("provider error detail is not after failed status: %q", got)
	}
	if strings.Contains(got, `error="`) {
		t.Fatalf("provider error detail is quoted: %q", got)
	}
}
