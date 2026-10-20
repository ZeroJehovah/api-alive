package alive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadConfigMigratesMissingToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"models":["model-a"]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Token) < 32 {
		t.Fatalf("generated token length = %d", len(cfg.Token))
	}
	reloaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Token != cfg.Token {
		t.Fatal("migrated token changed on reload")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"token"`) {
		t.Fatal("migrated config does not persist token")
	}
}

func TestLoadConfigMigratesLegacyBaseURLAndBatchInterval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := `{"models":["model-a"],"token":"test-token","codex_base_url":"https://cpa.example/v1","claude_base_url":"https://cpa.example"}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://cpa.example" || cfg.BatchMinIntervalSeconds != defaultBatchMinIntervalSeconds {
		t.Fatalf("migrated config = base %q interval %d", cfg.BaseURL, cfg.BatchMinIntervalSeconds)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{`"base_url": "https://cpa.example"`, `"batch_min_interval_seconds": 5`} {
		if !strings.Contains(text, want) {
			t.Fatalf("saved config missing %q: %s", want, text)
		}
	}
	for _, legacyKey := range []string{"codex_base_url", "claude_base_url"} {
		if strings.Contains(text, legacyKey) {
			t.Fatalf("saved config still contains %s: %s", legacyKey, text)
		}
	}
}

func TestLoadConfigPreservesExplicitZeroBatchInterval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	input := `{"models":["model-a"],"token":"test-token","base_url":"https://cpa.example","batch_min_interval_seconds":0}`
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BatchMinIntervalSeconds != 0 {
		t.Fatalf("batch interval = %d, want 0", cfg.BatchMinIntervalSeconds)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var persisted map[string]any
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted["batch_min_interval_seconds"] != float64(0) {
		t.Fatalf("explicit zero interval was not preserved on disk: %s", data)
	}
}

func TestAddModelsAppendsUniqueTrimmedModels(t *testing.T) {
	got := AddModels([]string{" gpt-5 ", "gpt-5"}, []string{"gpt-5-mini", "", " gpt-5 "})
	want := []string{"gpt-5", "gpt-5-mini"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AddModels() = %#v, want %#v", got, want)
	}
}

func TestRemoveModelsRemovesConfiguredModels(t *testing.T) {
	got := RemoveModels([]string{"gpt-5", "gpt-5-mini", "gpt-5"}, []string{" gpt-5 "})
	want := []string{"gpt-5-mini"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RemoveModels() = %#v, want %#v", got, want)
	}
}

func TestConfigMigratesLegacyLoopCountToPerModelCounts(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a", "b"}
	cfg.LoopCount = 3
	cfg.ApplyDefaults()
	if cfg.LoopCount != 0 {
		t.Fatalf("legacy loop count = %d, want 0", cfg.LoopCount)
	}
	if cfg.ModelLoopCounts["a"] != 3 || cfg.ModelLoopCounts["b"] != 3 {
		t.Fatalf("model loop counts = %#v, want both 3", cfg.ModelLoopCounts)
	}
}

func TestConfigDefaultsNewModelLoopCountToOne(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a", "b"}
	cfg.ModelLoopCounts = map[string]int{"a": 5}
	cfg.ApplyDefaults()
	if cfg.ModelLoopCounts["a"] != 5 || cfg.ModelLoopCounts["b"] != 1 {
		t.Fatalf("model loop counts = %#v", cfg.ModelLoopCounts)
	}
}

func TestConfigDefaultsCommandsAndRejectsUnknownInternalProvider(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Provider = ""
	cfg.BaseURL = ""
	cfg.Models = []string{"a"}
	cfg.ApplyDefaults()
	if cfg.Provider != ProviderCodex || cfg.BaseURL != defaultBaseURL {
		t.Fatalf("defaults = provider %q base URL %q", cfg.Provider, cfg.BaseURL)
	}

	cfg.Provider = "bad"
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() accepted unsupported provider")
	}
}

func TestConfigMigratesFlatModelsToDefaultGroup(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a", "b", "a"}
	cfg.ApplyDefaults()
	wantGroups := []ModelGroup{{Name: "Default", Provider: ProviderCodex, Models: []string{"a", "b"}}}
	if !reflect.DeepEqual(cfg.ModelGroups, wantGroups) {
		t.Fatalf("model groups = %#v, want %#v", cfg.ModelGroups, wantGroups)
	}
	if !reflect.DeepEqual(cfg.Models, []string{"a", "b"}) {
		t.Fatalf("models = %#v, want flat group order", cfg.Models)
	}
}

func TestConfigPreservesModelGroupsAndFlattensModels(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModelGroups = []ModelGroup{
		{Name: "fast", Provider: ProviderCodex, Models: []string{"a", "b"}},
		{Name: "slow", Provider: ProviderClaude, Models: []string{"b", "c", ""}},
	}
	cfg.ApplyDefaults()
	wantGroups := []ModelGroup{
		{Name: "fast", Provider: ProviderCodex, Models: []string{"a", "b"}},
		{Name: "slow", Provider: ProviderClaude, Models: []string{"c"}},
	}
	if !reflect.DeepEqual(cfg.ModelGroups, wantGroups) {
		t.Fatalf("model groups = %#v, want %#v", cfg.ModelGroups, wantGroups)
	}
	if !reflect.DeepEqual(cfg.Models, []string{"a", "b", "c"}) {
		t.Fatalf("models = %#v, want flattened groups", cfg.Models)
	}
}

func TestConfigFindsProviderFromModelGroup(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModelGroups = []ModelGroup{{Name: "Claude", Provider: ProviderClaude, Models: []string{"sonnet"}}}
	cfg.ApplyDefaults()
	if got := cfg.ProviderForModel("sonnet"); got != ProviderClaude {
		t.Fatalf("ProviderForModel = %q, want %q", got, ProviderClaude)
	}
}

func TestConfigPreservesZeroLoopCountAndClampsToTwoDigits(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models = []string{"a", "b", "c"}
	cfg.ModelLoopCounts = map[string]int{"a": 0, "b": 120, "c": -2}
	cfg.ApplyDefaults()
	if cfg.ModelLoopCounts["a"] != 0 || cfg.ModelLoopCounts["b"] != 99 || cfg.ModelLoopCounts["c"] != 1 {
		t.Fatalf("model loop counts = %#v, want a=0 b=99 c=1", cfg.ModelLoopCounts)
	}
	if got := cfg.LoopCountForModel("a"); got != 0 {
		t.Fatalf("LoopCountForModel(a) = %d, want 0", got)
	}
}
