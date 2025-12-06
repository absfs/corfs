package corfs

import (
	"os"

	"github.com/absfs/absfs"
)

// File wraps files from both primary and cache filesystems.
type File struct {
	primary absfs.File // Primary file handle
	cache   absfs.File // Cache file handle (may be nil)
	name    string
	fs      *FileSystem
	cached  bool // Track if we've cached the content
}

// Name returns the name of the file.
func (f *File) Name() string {
	return f.name
}

// Read reads from the primary file and caches content to the cache file.
func (f *File) Read(b []byte) (int, error) {
	n, err := f.primary.Read(b)

	// On successful read, try to cache the data
	if n > 0 && f.cache == nil && !f.cached {
		// Open cache file for writing if not already open
		if cacheFile, cacheErr := f.fs.cache.OpenFile(f.name, os.O_CREATE|os.O_WRONLY, 0644); cacheErr == nil {
			f.cache = cacheFile
		}
	}

	// Write to cache if available
	if n > 0 && f.cache != nil {
		f.cache.Write(b[:n])
	}

	return n, err
}

// ReadAt reads from the primary file at a specific offset.
func (f *File) ReadAt(b []byte, off int64) (int, error) {
	return f.primary.ReadAt(b, off)
}

// Write writes to both primary and cache files.
func (f *File) Write(b []byte) (int, error) {
	n, err := f.primary.Write(b)
	if n > 0 && f.cache != nil {
		f.cache.Write(b[:n])
	}
	return n, err
}

// WriteAt writes to both files at a specific offset.
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	n, err := f.primary.WriteAt(b, off)
	if n > 0 && f.cache != nil {
		f.cache.WriteAt(b[:n], off)
	}
	return n, err
}

// WriteString writes a string to both files.
func (f *File) WriteString(s string) (int, error) {
	n, err := f.primary.WriteString(s)
	if n > 0 && f.cache != nil {
		f.cache.WriteString(s)
	}
	return n, err
}

// Close closes both file handles.
func (f *File) Close() error {
	var err error
	if f.primary != nil {
		err = f.primary.Close()
	}
	if f.cache != nil {
		f.cache.Close()
	}
	return err
}

// Seek seeks in the primary file.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	ret, err := f.primary.Seek(offset, whence)
	if f.cache != nil {
		f.cache.Seek(offset, whence)
	}
	return ret, err
}

// Stat returns file info from the primary file.
func (f *File) Stat() (os.FileInfo, error) {
	return f.primary.Stat()
}

// Sync syncs both files.
func (f *File) Sync() error {
	err := f.primary.Sync()
	if f.cache != nil {
		f.cache.Sync()
	}
	return err
}

// Truncate truncates both files.
func (f *File) Truncate(size int64) error {
	err := f.primary.Truncate(size)
	if f.cache != nil {
		f.cache.Truncate(size)
	}
	return err
}

// Readdir reads directory entries from the primary file.
func (f *File) Readdir(n int) ([]os.FileInfo, error) {
	entries, err := f.primary.Readdir(n)
	if err != nil {
		return entries, err
	}

	// Filter out "." and ".." entries to match standard filesystem behavior
	filtered := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() != "." && entry.Name() != ".." {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// Readdirnames reads directory entry names from the primary file.
func (f *File) Readdirnames(n int) ([]string, error) {
	names, err := f.primary.Readdirnames(n)
	if err != nil {
		return names, err
	}

	// Filter out "." and ".." entries to match standard filesystem behavior
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if name != "." && name != ".." {
			filtered = append(filtered, name)
		}
	}

	return filtered, nil
}
