package alive

import (
	"reflect"
	"testing"
)

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

func TestConfigDefaultsProviderCommandsAndRejectsUnknownProvider(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Provider = ""
	cfg.CodexCommand = ""
	cfg.ClaudeCommand = ""
	cfg.Models = []string{"a"}
	cfg.ApplyDefaults()
	if cfg.Provider != ProviderCodex || cfg.CodexCommand != "codex" || cfg.ClaudeCommand != "claude" {
		t.Fatalf("defaults = provider %q codex %q claude %q", cfg.Provider, cfg.CodexCommand, cfg.ClaudeCommand)
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

func TestConfigUsesGlobalProviderForLegacyGroups(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Provider = ProviderClaude
	cfg.Models = []string{"sonnet"}
	cfg.ApplyDefaults()
	wantGroups := []ModelGroup{{Name: "Default", Provider: ProviderClaude, Models: []string{"sonnet"}}}
	if !reflect.DeepEqual(cfg.ModelGroups, wantGroups) {
		t.Fatalf("model groups = %#v, want %#v", cfg.ModelGroups, wantGroups)
	}
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
