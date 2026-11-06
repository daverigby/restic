package main

import (
	"path/filepath"
	"runtime"

	"github.com/restic/restic/internal/archiver"
	"github.com/restic/restic/internal/debug"
	"github.com/restic/restic/internal/errors"
	"github.com/restic/restic/internal/fs"
	"github.com/restic/restic/internal/repository"
)

// cacheOverlapsTargets reports whether a backup target might contain the
// restic cache, or be contained by it. If this cannot be ruled out, keep the
// normal per-item cache exclusion.
func cacheOverlapsTargets(cacheBase string, targets []string) bool {
	// Only Linux path semantics have been validated for this fast path.
	// Other systems and virtual stdin targets retain the original filter.
	if len(targets) == 0 || runtime.GOOS != "linux" || !filepath.IsAbs(cacheBase) {
		return true
	}
	// The per-item filter compares path names. Keep it for symlinked paths,
	// whose lexical and physical parent relationships can differ.
	resolvedCache, err := filepath.EvalSymlinks(cacheBase)
	if err != nil || resolvedCache != filepath.Clean(cacheBase) {
		return true
	}

	for _, target := range targets {
		absolute, err := filepath.Abs(target)
		if err != nil {
			return true
		}
		resolvedTarget, err := filepath.EvalSymlinks(absolute)
		if err != nil || resolvedTarget != absolute {
			return true
		}
		if fs.HasPathPrefix(absolute, cacheBase) || fs.HasPathPrefix(cacheBase, absolute) {
			return true
		}
	}
	return false
}

// rejectResticCache returns a RejectByNameFunc that rejects the restic cache
// directory (if set).
func rejectResticCache(repo *repository.Repository) (archiver.RejectByNameFunc, error) {
	if repo.Cache() == nil {
		return func(string) bool {
			return false
		}, nil
	}
	cacheBase := repo.Cache().BaseDir()

	if cacheBase == "" {
		return nil, errors.New("cacheBase is empty string")
	}

	return func(item string) bool {
		if fs.HasPathPrefix(cacheBase, item) {
			debug.Log("rejecting restic cache directory %v", item)
			return true
		}

		return false
	}, nil
}
