package fs

import (
	"path/filepath"
	"strings"
)

// HasPathPrefix returns true if p is a subdir of (or a file within) base. It
// assumes a file system which is case sensitive. If the paths are not of the
// same type (one is relative, the other is absolute), false is returned.
func HasPathPrefix(base, p string) bool {
	volume := filepath.VolumeName(base)
	if volume != filepath.VolumeName(p) {
		return false
	}

	// handle case when base and p are not of the same type
	if filepath.IsAbs(base) != filepath.IsAbs(p) {
		return false
	}

	base = filepath.Clean(base)
	p = filepath.Clean(p)

	if base == p {
		return true
	}
	// A relative path's parent chain always reaches ".".
	if base == "." {
		return true
	}

	// Keep the original walk on Windows, where drive-relative and rooted paths
	// have additional volume rules. This branch is absent on Unix builds.
	if filepath.Separator != '/' {
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

	// After cleaning, a parent is an exact prefix ending at a path boundary.
	// Roots already end in a separator.
	if !strings.HasPrefix(p, base) {
		return false
	}
	if strings.HasSuffix(base, string(filepath.Separator)) {
		return true
	}
	return len(p) > len(base) && p[len(base)] == filepath.Separator
}
