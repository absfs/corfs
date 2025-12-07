package corfs_test

import (
	"testing"

	"github.com/absfs/absfs"
	"github.com/absfs/corfs"
	"github.com/absfs/fstesting"
	"github.com/absfs/memfs"
)

// TestCorFS_WrapperSuite runs the fstesting wrapper suite for corfs.
// corfs is a cache-on-read wrapper that combines a primary and cache filesystem.
func TestCorFS_WrapperSuite(t *testing.T) {
	// Create an in-memory base filesystem for testing
	baseFS, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	// Ensure temp directory exists (memfs doesn't create /tmp by default)
	if err := baseFS.MkdirAll(baseFS.TempDir(), 0755); err != nil {
		t.Fatal(err)
	}

	// Factory creates a corfs wrapper around the base filesystem
	factory := func(base absfs.FileSystem) (absfs.FileSystem, error) {
		// Create a separate cache filesystem
		cacheFS, err := memfs.NewFS()
		if err != nil {
			return nil, err
		}

		// Ensure temp directory exists in cache (memfs doesn't create /tmp by default)
		if err := cacheFS.MkdirAll(cacheFS.TempDir(), 0755); err != nil {
			return nil, err
		}

		// corfs.New returns a *corfs.FileSystem which implements absfs.Filer
		// We need to extend it to absfs.FileSystem
		corFilesystem := corfs.New(base, cacheFS)
		return absfs.ExtendFiler(corFilesystem), nil
	}

	suite := &fstesting.WrapperSuite{
		Factory:        factory,
		BaseFS:         baseFS,
		Name:           "corfs",
		TransformsData: false,   // corfs passes data through unchanged
		TransformsMeta: false,   // corfs preserves metadata
		ReadOnly:       false,   // corfs supports write operations
		TestDir:        "/test", // Use /test instead of /tmp to avoid path issues
	}

	// Ensure test directory exists
	if err := baseFS.MkdirAll("/test", 0755); err != nil {
		t.Fatal(err)
	}

	suite.Run(t)
}

// TestCorFS_Suite runs the full fstesting suite for corfs.
// This tests corfs as a complete filesystem implementation.
func TestCorFS_Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full suite in short mode")
	}

	// Create primary and cache filesystems
	primary, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	cache, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	// Ensure temp directory exists in both filesystems (memfs doesn't create /tmp by default)
	if err := primary.MkdirAll(primary.TempDir(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := cache.MkdirAll(cache.TempDir(), 0755); err != nil {
		t.Fatal(err)
	}

	// Create corfs and extend to FileSystem
	corFilesystem := corfs.New(primary, cache)
	fs := absfs.ExtendFiler(corFilesystem)

	// Configure features based on what corfs supports
	// corfs delegates to underlying filesystems (memfs in this case)
	// so it supports what memfs supports
	features := fstesting.Features{
		Symlinks:      false, // memfs doesn't support symlinks
		HardLinks:     false, // memfs doesn't support hard links
		Permissions:   true,  // memfs supports permissions
		Timestamps:    true,  // memfs supports timestamps
		CaseSensitive: true,  // memfs is case-sensitive
		AtomicRename:  true,  // memfs has atomic rename
		SparseFiles:   false, // memfs doesn't support sparse files
		LargeFiles:    true,  // memfs supports large files
	}

	suite := &fstesting.Suite{
		FS:       fs,
		Features: features,
		TestDir:  "/test", // Use /test instead of /tmp to avoid path issues
	}

	// Ensure test directory exists
	if err := fs.MkdirAll("/test", 0755); err != nil {
		t.Fatal(err)
	}

	suite.Run(t)
}

// TestCorFS_QuickCheck runs a quick sanity check.
func TestCorFS_QuickCheck(t *testing.T) {
	// Create primary and cache filesystems
	primary, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	cache, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	// Ensure temp directory exists in both filesystems (memfs doesn't create /tmp by default)
	if err := primary.MkdirAll(primary.TempDir(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := cache.MkdirAll(cache.TempDir(), 0755); err != nil {
		t.Fatal(err)
	}

	// Create corfs and extend to FileSystem
	corFilesystem := corfs.New(primary, cache)
	fs := absfs.ExtendFiler(corFilesystem)

	suite := &fstesting.Suite{
		FS: fs,
		Features: fstesting.Features{
			Permissions:   true,
			Timestamps:    true,
			CaseSensitive: true,
			AtomicRename:  true,
			LargeFiles:    true,
		},
		TestDir: "/test", // Use /test instead of /tmp to avoid path issues
	}

	// Ensure test directory exists
	if err := fs.MkdirAll("/test", 0755); err != nil {
		t.Fatal(err)
	}

	suite.QuickCheck(t)
}
