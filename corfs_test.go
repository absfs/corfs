package corfs

import (
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/absfs/absfs"
)

// mockFiler is a minimal mock implementation for testing
type mockFiler struct {
	files map[string]*mockFile
}

func newMockFiler() *mockFiler {
	return &mockFiler{files: make(map[string]*mockFile)}
}

func (m *mockFiler) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	if f, ok := m.files[name]; ok {
		return f, nil
	}
	f := &mockFile{name: name, data: []byte{}}
	m.files[name] = f
	return f, nil
}

func (m *mockFiler) Mkdir(name string, perm os.FileMode) error         { return nil }
func (m *mockFiler) Remove(name string) error                          { return nil }
func (m *mockFiler) Rename(oldpath, newpath string) error              { return nil }
func (m *mockFiler) Stat(name string) (os.FileInfo, error)             { return nil, os.ErrNotExist }
func (m *mockFiler) Chmod(name string, mode os.FileMode) error         { return nil }
func (m *mockFiler) Chtimes(name string, atime, mtime time.Time) error { return nil }
func (m *mockFiler) Chown(name string, uid, gid int) error             { return nil }
func (m *mockFiler) ReadDir(name string) ([]fs.DirEntry, error)        { return nil, nil }
func (m *mockFiler) ReadFile(name string) ([]byte, error)              { return nil, nil }
func (m *mockFiler) Sub(dir string) (fs.FS, error) {
	return absfs.FilerToFS(m, dir)
}

type mockFile struct {
	name   string
	data   []byte
	offset int64
}

func (f *mockFile) Name() string                                 { return f.name }
func (f *mockFile) Read(b []byte) (int, error)                   { return 0, nil }
func (f *mockFile) Write(b []byte) (int, error)                  { f.data = append(f.data, b...); return len(b), nil }
func (f *mockFile) Close() error                                 { return nil }
func (f *mockFile) Seek(offset int64, whence int) (int64, error) { return 0, nil }
func (f *mockFile) Stat() (os.FileInfo, error)                   { return nil, nil }
func (f *mockFile) Sync() error                                  { return nil }
func (f *mockFile) Readdir(n int) ([]os.FileInfo, error)         { return nil, nil }
func (f *mockFile) Readdirnames(n int) ([]string, error)         { return nil, nil }
func (f *mockFile) ReadAt(b []byte, off int64) (int, error)      { return 0, nil }
func (f *mockFile) WriteAt(b []byte, off int64) (int, error)     { return len(b), nil }
func (f *mockFile) WriteString(s string) (int, error)            { return len(s), nil }
func (f *mockFile) Truncate(size int64) error                    { return nil }
func (f *mockFile) ReadDir(n int) ([]fs.DirEntry, error)         { return nil, nil }

func TestNew(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()

	fs := New(primary, cache)
	if fs == nil {
		t.Fatal("New() returned nil")
	}
	if fs.primary != primary {
		t.Error("primary filesystem not set correctly")
	}
	if fs.cache != cache {
		t.Error("cache filesystem not set correctly")
	}
}

func TestOpenFile(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if f == nil {
		t.Fatal("OpenFile() returned nil file")
	}
	f.Close()
}

func TestMkdir(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	err := fs.Mkdir("/testdir", 0755)
	if err != nil {
		t.Errorf("Mkdir() error = %v", err)
	}
}

func TestRemove(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	err := fs.Remove("/test.txt")
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}
}

func TestRename(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	err := fs.Rename("/old.txt", "/new.txt")
	if err != nil {
		t.Errorf("Rename() error = %v", err)
	}
}

func TestStat(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	_, err := fs.Stat("/test.txt")
	if err == nil {
		t.Error("Stat() expected error for non-existent file")
	}
}

func TestChmod(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	err := fs.Chmod("/test.txt", 0644)
	if err != nil {
		t.Errorf("Chmod() error = %v", err)
	}
}

func TestChtimes(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	now := time.Now()
	err := fs.Chtimes("/test.txt", now, now)
	if err != nil {
		t.Errorf("Chtimes() error = %v", err)
	}
}

func TestChown(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	err := fs.Chown("/test.txt", 1000, 1000)
	if err != nil {
		t.Errorf("Chown() error = %v", err)
	}
}

func TestFileRead(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	buf := make([]byte, 100)
	_, err = f.Read(buf)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
}

func TestFileWrite(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	n, err := f.Write([]byte("test data"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 9 {
		t.Errorf("Write() wrote %d bytes, expected 9", n)
	}
}

func TestFileReadAt(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	buf := make([]byte, 10)
	_, err = f.ReadAt(buf, 0)
	if err != nil {
		t.Errorf("ReadAt() error = %v", err)
	}
}

func TestFileWriteAt(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	n, err := f.WriteAt([]byte("test"), 0)
	if err != nil {
		t.Errorf("WriteAt() error = %v", err)
	}
	if n != 4 {
		t.Errorf("WriteAt() wrote %d bytes, expected 4", n)
	}
}

func TestFileWriteString(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	n, err := f.WriteString("test string")
	if err != nil {
		t.Errorf("WriteString() error = %v", err)
	}
	if n != 11 {
		t.Errorf("WriteString() wrote %d bytes, expected 11", n)
	}
}

func TestFileSeek(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	_, err = f.Seek(10, 0)
	if err != nil {
		t.Errorf("Seek() error = %v", err)
	}
}

func TestFileStat(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	_, err = f.Stat()
	if err != nil {
		t.Errorf("Stat() error = %v", err)
	}
}

func TestFileSync(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	err = f.Sync()
	if err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestFileTruncate(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	err = f.Truncate(100)
	if err != nil {
		t.Errorf("Truncate() error = %v", err)
	}
}

func TestFileReaddir(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/testdir", os.O_RDONLY, 0755)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	_, err = f.Readdir(-1)
	if err != nil {
		t.Errorf("Readdir() error = %v", err)
	}
}

func TestFileReaddirnames(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/testdir", os.O_RDONLY, 0755)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	_, err = f.Readdirnames(-1)
	if err != nil {
		t.Errorf("Readdirnames() error = %v", err)
	}
}

func TestFileName(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	if f.Name() != "/test.txt" {
		t.Errorf("Name() = %q, expected %q", f.Name(), "/test.txt")
	}
}

func TestCacheOnRead(t *testing.T) {
	primary := newMockFiler()
	cache := newMockFiler()

	// Add a file to primary
	primaryFile := &mockFile{name: "/cached.txt", data: []byte("cached content")}
	primary.files["/cached.txt"] = primaryFile

	fs := New(primary, cache)

	// Open and read from primary
	f, err := fs.OpenFile("/cached.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	buf := make([]byte, 100)
	_, err = f.Read(buf)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
}

func TestOpenFileFromCache(t *testing.T) {
	primary := &mockFilerWithError{err: os.ErrNotExist}
	cache := newMockFiler()

	// Add a file to cache
	cacheFile := &mockFile{name: "/cached.txt", data: []byte("cached content")}
	cache.files["/cached.txt"] = cacheFile

	fs := New(primary, cache)

	// Should fall back to cache when primary fails
	f, err := fs.OpenFile("/cached.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v, expected to read from cache", err)
	}
	if f == nil {
		t.Fatal("OpenFile() returned nil file")
	}
	f.Close()
}

// mockFilerWithError is a mock that returns errors
type mockFilerWithError struct {
	err error
}

func (m *mockFilerWithError) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	return nil, m.err
}

func (m *mockFilerWithError) Mkdir(name string, perm os.FileMode) error         { return m.err }
func (m *mockFilerWithError) Remove(name string) error                          { return m.err }
func (m *mockFilerWithError) Rename(oldpath, newpath string) error              { return m.err }
func (m *mockFilerWithError) Stat(name string) (os.FileInfo, error)             { return nil, m.err }
func (m *mockFilerWithError) Chmod(name string, mode os.FileMode) error         { return m.err }
func (m *mockFilerWithError) Chtimes(name string, atime, mtime time.Time) error { return m.err }
func (m *mockFilerWithError) Chown(name string, uid, gid int) error             { return m.err }
func (m *mockFilerWithError) ReadDir(name string) ([]fs.DirEntry, error)        { return nil, m.err }
func (m *mockFilerWithError) ReadFile(name string) ([]byte, error)              { return nil, m.err }
func (m *mockFilerWithError) Sub(dir string) (fs.FS, error) { return nil, m.err }

// Benchmarks

func BenchmarkOpenFile(b *testing.B) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := fs.OpenFile("/bench.txt", os.O_RDONLY, 0644)
		if f != nil {
			f.Close()
		}
	}
}

func BenchmarkFileRead(b *testing.B) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, _ := fs.OpenFile("/bench.txt", os.O_RDONLY, 0644)
	defer f.Close()

	buf := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Read(buf)
	}
}

func BenchmarkFileWrite(b *testing.B) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	f, _ := fs.OpenFile("/bench.txt", os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	data := []byte("benchmark data")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Write(data)
	}
}

func BenchmarkCacheHit(b *testing.B) {
	primary := &mockFilerWithError{err: os.ErrNotExist}
	cache := newMockFiler()

	// Populate cache
	cacheFile := &mockFile{name: "/cached.txt", data: []byte("cached content")}
	cache.files["/cached.txt"] = cacheFile

	fs := New(primary, cache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := fs.OpenFile("/cached.txt", os.O_RDONLY, 0644)
		if f != nil {
			f.Close()
		}
	}
}

func BenchmarkMkdir(b *testing.B) {
	primary := newMockFiler()
	cache := newMockFiler()
	fs := New(primary, cache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Mkdir("/benchdir", 0755)
	}
}
