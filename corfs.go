// Package corfs implements a Cache-on-Read FileSystem that wraps two absfs.Filer
// implementations. It reads from the primary filesystem and caches content to
// the secondary filesystem on successful reads, providing a two-tier caching system.
package corfs

import (
	"os"
	"time"

	"github.com/absfs/absfs"
)

// FileSystem implements absfs.Filer with cache-on-read semantics.
// Reads are performed from the primary filesystem, with successful reads
// being cached to the secondary filesystem for future access.
type FileSystem struct {
	primary absfs.Filer // Primary filesystem to read from
	cache   absfs.Filer // Secondary filesystem for caching
}

// New creates a new CorFS that reads from primary and caches to cache.
func New(primary, cache absfs.Filer) *FileSystem {
	return &FileSystem{
		primary: primary,
		cache:   cache,
	}
}

// OpenFile opens a file from the primary filesystem and caches it to the cache
// filesystem on successful read operations.
func (fs *FileSystem) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	// Try to open from primary first
	primaryFile, primaryErr := fs.primary.OpenFile(name, flag, perm)

	// If we're creating or writing, try both filesystems
	if flag&(os.O_CREATE|os.O_WRONLY|os.O_RDWR) != 0 {
		if primaryErr != nil {
			return primaryFile, primaryErr
		}
		// Try to open/create in cache as well for write operations
		cacheFile, _ := fs.cache.OpenFile(name, flag, perm)
		return &File{
			primary: primaryFile,
			cache:   cacheFile,
			name:    name,
			fs:      fs,
		}, nil
	}

	// For read operations, return wrapped file
	if primaryErr != nil {
		// Try cache as fallback
		cacheFile, cacheErr := fs.cache.OpenFile(name, flag, perm)
		if cacheErr != nil {
			return nil, primaryErr // Return original error
		}
		return cacheFile, nil
	}

	return &File{
		primary: primaryFile,
		cache:   nil,
		name:    name,
		fs:      fs,
	}, nil
}

// Mkdir creates a directory in both filesystems.
func (fs *FileSystem) Mkdir(name string, perm os.FileMode) error {
	err := fs.primary.Mkdir(name, perm)
	fs.cache.Mkdir(name, perm) // Best effort for cache
	return err
}

// Remove removes a file from both filesystems.
func (fs *FileSystem) Remove(name string) error {
	err := fs.primary.Remove(name)
	fs.cache.Remove(name) // Best effort for cache
	return err
}

// Rename renames a file in both filesystems.
func (fs *FileSystem) Rename(oldpath, newpath string) error {
	err := fs.primary.Rename(oldpath, newpath)
	fs.cache.Rename(oldpath, newpath) // Best effort for cache
	return err
}

// Stat returns file info from the primary filesystem.
func (fs *FileSystem) Stat(name string) (os.FileInfo, error) {
	info, err := fs.primary.Stat(name)
	if err != nil {
		// Try cache as fallback
		return fs.cache.Stat(name)
	}
	return info, nil
}

// Chmod changes the mode in both filesystems.
func (fs *FileSystem) Chmod(name string, mode os.FileMode) error {
	err := fs.primary.Chmod(name, mode)
	fs.cache.Chmod(name, mode) // Best effort for cache
	return err
}

// Chtimes changes the access and modification times in both filesystems.
func (fs *FileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	err := fs.primary.Chtimes(name, atime, mtime)
	fs.cache.Chtimes(name, atime, mtime) // Best effort for cache
	return err
}

// Chown changes the owner and group in both filesystems.
func (fs *FileSystem) Chown(name string, uid, gid int) error {
	err := fs.primary.Chown(name, uid, gid)
	fs.cache.Chown(name, uid, gid) // Best effort for cache
	return err
}
