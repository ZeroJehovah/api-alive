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
