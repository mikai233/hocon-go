package config

import (
	"encoding/json"
	"errors"
	"hocon-go/common"
	"hocon-go/parser"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConfigGoldenFiles(t *testing.T) {
	resBase := resourcesDir(t)
	opts := &parser.ConfigOptions{
		Classpath: []string{resBase},
	}
	tests := []struct {
		name     string
		expected string
		skip     bool
	}{
		{"empty", "empty.json", false},
		{"add_assign", "add_assign_expected.json", false},
		{"concat", "concat.json", false},
		{"concat2", "concat2.json", false},
		{"comment", "comment.json", false},
		{"substitution", "substitution.json", false},
		{"base", "base.json", true},       // TODO: path segments with leading whitespace.
		{"concat3", "concat3.json", true}, // TODO: complex concatenation semantics.
		{"concat4", "concat4.json", true},
		{"concat5", "concat5.json", true},
		{"include", "include.json", true}, // TODO: relative substitutions in included files.
		{"substitution3", "substitution3.json", true},
		{"self_referential", "self_referential.json", true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip("pending implementation parity")
			}
			confPath := filepath.Join(resBase, tc.name+".conf")
			actual, err := Load(confPath, opts)
			if err != nil {
				t.Fatalf("Load(%s) failed: %v", confPath, err)
			}
			expectedPath := filepath.Join(resBase, tc.expected)
			expected := readJSON(t, expectedPath)
			normalizedActual := normalizeValue(t, actual)
			if !reflect.DeepEqual(normalizedActual, expected) {
				t.Fatalf("resolved config mismatch for %s\nactual:   %+v\nexpected: %+v", tc.name, normalizedActual, expected)
			}
		})
	}
}

func TestConfigErrors(t *testing.T) {
	resBase := resourcesDir(t)
	opts := &parser.ConfigOptions{
		Classpath: []string{resBase},
	}
	t.Run("max_depth", func(t *testing.T) {
		_, err := Load(filepath.Join(resBase, "max_depth.conf"), opts)
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "depth") {
			t.Fatalf("expected depth error, got %v", err)
		}
	})

	t.Run("include_cycle", func(t *testing.T) {
		_, err := Load(filepath.Join(resBase, "include_cycle.conf"), opts)
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "cycle") {
			t.Fatalf("expected include cycle error, got %v", err)
		}
	})

	t.Run("substitution_cycle", func(t *testing.T) {
		_, err := Load(filepath.Join(resBase, "substitution_cycle.conf"), opts)
		var cycleErr *common.SubstitutionCycle
		if !errors.As(err, &cycleErr) {
			t.Fatalf("expected substitution cycle error, got %v", err)
		}
	})

	t.Run("substitution_not_found", func(t *testing.T) {
		t.Skip("relative substitution resolution pending")
	})
}

func readJSON(t *testing.T, path string) interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readJSON(%s): %v", path, err)
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", path, err)
	}
	return v
}

func normalizeValue(t *testing.T, v interface{}) interface{} {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var out interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return out
}

func resourcesDir(t *testing.T) string {
	t.Helper()
	base := filepath.Join("..", "resources")
	abs, err := filepath.Abs(base)
	if err != nil {
		t.Fatalf("resourcesDir: %v", err)
	}
	return abs
}
