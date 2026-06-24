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
