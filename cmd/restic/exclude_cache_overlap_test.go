package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCacheOverlapsTargets(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("non-Linux systems always keep the per-item filter")
	}
	base := filepath.Join(t.TempDir(), "cache")
	parent := filepath.Dir(base)
	for _, dir := range []string{base, filepath.Join(base, "data"), filepath.Join(parent, "data", "subdir"), filepath.Join(parent, "other")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name    string
		targets []string
		want    bool
	}{
		{"disjoint", []string{filepath.Join(parent, "data")}, false},
		{"multiple-disjoint", []string{filepath.Join(parent, "data"), filepath.Join(parent, "other")}, false},
		{"ancestor", []string{parent}, true},
		{"inside", []string{filepath.Join(base, "data")}, true},
		{"equal", []string{base}, true},
		{"mixed", []string{filepath.Join(parent, "data"), parent}, true},
		{"no-targets", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := cacheOverlapsTargets(base, tc.targets); got != tc.want {
				t.Fatalf("cacheOverlapsTargets(%q, %q) = %t, want %t", base, tc.targets, got, tc.want)
			}
		})
	}
	if !cacheOverlapsTargets("relative/cache", []string{parent}) {
		t.Fatal("relative cache location must retain the filter")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, filepath.Join(parent, "data"))
	if err != nil {
		t.Fatal(err)
	}
	if cacheOverlapsTargets(base, []string{relative}) {
		t.Fatal("relative disjoint target must be resolved before overlap check")
	}
	linkTarget := filepath.Join(parent, "target-link")
	if err := os.Symlink(filepath.Join(parent, "data"), linkTarget); err != nil {
		t.Fatal(err)
	}
	if !cacheOverlapsTargets(base, []string{linkTarget}) ||
		!cacheOverlapsTargets(base, []string{filepath.Join(linkTarget, "subdir")}) {
		t.Fatal("symlinked target or ancestor must retain the filter")
	}
	linkCache := filepath.Join(parent, "cache-link")
	if err := os.Symlink(base, linkCache); err != nil {
		t.Fatal(err)
	}
	if !cacheOverlapsTargets(linkCache, []string{filepath.Join(parent, "data")}) {
		t.Fatal("symlinked cache path must retain the filter")
	}
}
