// Package corfs implements a Cache-on-Read FileSystem that wraps two absfs.Filer
// implementations. It reads from the primary filesystem and caches content to
// the secondary filesystem on successful reads, providing a two-tier caching system.
package corfs

import (
	"io/fs"
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

// Truncate truncates a file to the specified size in both filesystems.
func (fs *FileSystem) Truncate(name string, size int64) error {
	// Open file for writing (but don't truncate with O_TRUNC)
	f, err := fs.OpenFile(name, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Truncate to the specified size
	return f.Truncate(size)
}

// RemoveAll removes a path and any children it contains in both filesystems.
func (fs *FileSystem) RemoveAll(path string) error {
	// Remove from primary first
	var err error
	if remover, ok := fs.primary.(interface{ RemoveAll(string) error }); ok {
		err = remover.RemoveAll(path)
	} else {
		err = removeAll(fs.primary, path)
	}

	// Best effort removal from cache
	if remover, ok := fs.cache.(interface{ RemoveAll(string) error }); ok {
		remover.RemoveAll(path)
	} else {
		removeAll(fs.cache, path)
	}

	return err
}

// ReadDir reads the named directory and returns a list of directory entries.
func (fs *FileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.primary.ReadDir(name)
	if err != nil {
		// Try cache as fallback
		return fs.cache.ReadDir(name)
	}
	return entries, nil
}

// ReadFile reads the named file and returns its contents.
func (fs *FileSystem) ReadFile(name string) ([]byte, error) {
	data, err := fs.primary.ReadFile(name)
	if err != nil {
		// Try cache as fallback
		return fs.cache.ReadFile(name)
	}

	// On successful read, cache the data
	if err == nil && len(data) > 0 {
		// Best effort cache write
		fs.cache.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if cacheFile, cacheErr := fs.cache.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); cacheErr == nil {
			cacheFile.Write(data)
			cacheFile.Close()
		}
	}

	return data, nil
}

// Sub returns an fs.FS corresponding to the subtree rooted at dir.
func (fs *FileSystem) Sub(dir string) (fs.FS, error) {
	return absfs.FilerToFS(fs, dir)
}

var ErrNotDir = os.ErrInvalid

// subCorFS wraps a corfs FileSystem for a subdirectory.
type subCorFS struct {
	primary absfs.Filer
	cache   absfs.Filer
	dir     string
}

func (s *subCorFS) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	primaryFile, primaryErr := s.primary.OpenFile(name, flag, perm)

	if flag&(os.O_CREATE|os.O_WRONLY|os.O_RDWR) != 0 {
		if primaryErr != nil {
			return primaryFile, primaryErr
		}
		var cacheFile absfs.File
		if s.cache != nil {
			cacheFile, _ = s.cache.OpenFile(name, flag, perm)
		}
		return &File{
			primary: primaryFile,
			cache:   cacheFile,
			name:    name,
			fs:      nil, // subCorFS doesn't need fs reference
		}, nil
	}

	if primaryErr != nil {
		if s.cache != nil {
			cacheFile, cacheErr := s.cache.OpenFile(name, flag, perm)
			if cacheErr != nil {
				return nil, primaryErr
			}
			return cacheFile, nil
		}
		return nil, primaryErr
	}

	return &File{
		primary: primaryFile,
		cache:   nil,
		name:    name,
		fs:      nil,
	}, nil
}

func (s *subCorFS) Mkdir(name string, perm os.FileMode) error {
	err := s.primary.Mkdir(name, perm)
	if s.cache != nil {
		s.cache.Mkdir(name, perm)
	}
	return err
}

func (s *subCorFS) Remove(name string) error {
	err := s.primary.Remove(name)
	if s.cache != nil {
		s.cache.Remove(name)
	}
	return err
}

func (s *subCorFS) Rename(oldpath, newpath string) error {
	err := s.primary.Rename(oldpath, newpath)
	if s.cache != nil {
		s.cache.Rename(oldpath, newpath)
	}
	return err
}

func (s *subCorFS) Stat(name string) (os.FileInfo, error) {
	info, err := s.primary.Stat(name)
	if err != nil && s.cache != nil {
		return s.cache.Stat(name)
	}
	return info, err
}

func (s *subCorFS) Chmod(name string, mode os.FileMode) error {
	err := s.primary.Chmod(name, mode)
	if s.cache != nil {
		s.cache.Chmod(name, mode)
	}
	return err
}

func (s *subCorFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	err := s.primary.Chtimes(name, atime, mtime)
	if s.cache != nil {
		s.cache.Chtimes(name, atime, mtime)
	}
	return err
}

func (s *subCorFS) Chown(name string, uid, gid int) error {
	err := s.primary.Chown(name, uid, gid)
	if s.cache != nil {
		s.cache.Chown(name, uid, gid)
	}
	return err
}

func (s *subCorFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := s.primary.ReadDir(name)
	if err != nil && s.cache != nil {
		return s.cache.ReadDir(name)
	}
	return entries, err
}

func (s *subCorFS) ReadFile(name string) ([]byte, error) {
	data, err := s.primary.ReadFile(name)
	if err != nil && s.cache != nil {
		return s.cache.ReadFile(name)
	}
	return data, err
}

func (s *subCorFS) Sub(dir string) (fs.FS, error) {
	return absfs.FilerToFS(s, dir)
}

// removeAll is a helper that recursively removes a path.
func removeAll(filer absfs.Filer, path string) error {
	// Open the file to check if it's a directory
	f, err := filer.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	// Get FileInfo
	info, err := f.Stat()
	closeErr := f.Close()
	if err != nil {
		return err
	}

	// If it's not a directory, just remove it
	if !info.IsDir() {
		return filer.Remove(path)
	}

	// For directories, recursively remove contents
	f, err = filer.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	// Read all directory entries and remove them recursively
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == "." || name == ".." {
			continue
		}
		fullPath := path + string(os.PathSeparator) + name
		if err := removeAll(filer, fullPath); err != nil {
			return err
		}
	}

	// Finally, remove the directory itself
	if err := filer.Remove(path); err != nil {
		return err
	}

	return closeErr
}
