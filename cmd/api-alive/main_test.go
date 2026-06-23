package main

import (
	"bytes"
	"path/filepath"
	"reflect"
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
