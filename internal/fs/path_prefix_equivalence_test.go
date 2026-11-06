package fs

import (
	"math/rand"
	"path/filepath"
	"testing"
)

// hasPathPrefixOriginal is the parent-walk implementation used before the
// optimized boundary check. Keep it as a reference for unusual path forms.
func hasPathPrefixOriginal(base, p string) bool {
	if filepath.VolumeName(base) != filepath.VolumeName(p) || filepath.IsAbs(base) != filepath.IsAbs(p) {
		return false
	}
	base, p = filepath.Clean(base), filepath.Clean(p)
	if base == p {
		return true
	}
	for {
		dir := filepath.Dir(p)
		if base == dir {
			return true
		}
		if p == dir {
			return false
		}
		p = dir
	}
}

func TestHasPathPrefixEquivalent(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	components := []string{"", ".", "..", "a", "ab", "b"}
	path := func() string {
		n := rnd.Intn(9)
		out := ""
		if rnd.Intn(2) == 0 {
			out = string(filepath.Separator)
		}
		for i := 0; i < n; i++ {
			if i > 0 {
				out += string(filepath.Separator)
			}
			out += components[rnd.Intn(len(components))]
		}
		if rnd.Intn(2) == 0 {
			out += string(filepath.Separator)
		}
		return out
	}
	for i := 0; i < 20000; i++ {
		base, p := path(), path()
		got, want := HasPathPrefix(base, p), hasPathPrefixOriginal(base, p)
		if got != want {
			t.Fatalf("HasPathPrefix(%q, %q) = %t, want %t", base, p, got, want)
		}
	}
}
