package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"test-api-alive/internal/alive"
)

func TestRunListPrintsConfiguredModels(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := alive.DefaultConfig()
	cfg.Models = []string{"gpt-5", "gpt-5-mini"}
	if err := alive.SaveConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runWithArgs([]string{"list", "--config", configPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if got, want := stdout.String(), "gpt-5\ngpt-5-mini\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunAddAndRemoveUpdateConfigModels(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")

	var stdout, stderr bytes.Buffer
	code := runWithArgs([]string{"add", "--config", configPath, "gpt-5", "gpt-5-mini", "gpt-5"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("add exit code = %d, stderr = %q", code, stderr.String())
	}
	cfg, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"gpt-5", "gpt-5-mini"}; !reflect.DeepEqual(cfg.Models, want) {
		t.Fatalf("models after add = %#v, want %#v", cfg.Models, want)
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithArgs([]string{"remove", "--config", configPath, "gpt-5"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("remove exit code = %d, stderr = %q", code, stderr.String())
	}
	cfg, err = alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"gpt-5-mini"}; !reflect.DeepEqual(cfg.Models, want) {
		t.Fatalf("models after remove = %#v, want %#v", cfg.Models, want)
	}
}

func TestRunExcludeFiltersModelsForCurrentProbe(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := alive.DefaultConfig()
	cfg.Models = []string{"aaa/gpt-5.5", "bbb/gpt-5", "aaa-mini"}
	if err := alive.SaveConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runWithArgs([]string{"exclude", "--config", configPath, "--dry-run", "aaa"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if out := stdout.String(); strings.Contains(out, "aaa/gpt-5.5") || strings.Contains(out, "aaa-mini") {
		t.Fatalf("stdout contains excluded model: %q", out)
	}
	if out := stdout.String(); !strings.Contains(out, "bbb/gpt-5") {
		t.Fatalf("stdout does not contain kept model: %q", out)
	}

	loaded, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Models, cfg.Models) {
		t.Fatalf("exclude changed config models = %#v, want %#v", loaded.Models, cfg.Models)
	}
}

func TestRunExcludeFailsWhenAllModelsAreFiltered(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runWithArgs([]string{"exclude", "--models", "aaa/gpt-5.5,aaa-mini", "--dry-run", "aaa"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "at least one model is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunExcludeRequiresPrefix(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runWithArgs([]string{"exclude"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "at least one model prefix is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
