package analyzer

import (
	c "goMetrics/cache"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockFileSystem is a simple in-memory file system for testing.
type MockFileSystem struct {
	files map[string]string
}

func (mfs *MockFileSystem) Open(name string) (fs.File, error) {
	content, ok := mfs.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return &MockFile{name: name, content: content}, nil
}

func (mfs *MockFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	for path := range mfs.files {
		if filepath.Dir(path) == name {
			entries = append(entries, &MockDirEntry{name: filepath.Base(path)})
		}
	}
	return entries, nil
}

// MockFile implements fs.File.
type MockFile struct {
	name    string
	content string
	pos     int
}

func (mf *MockFile) Read(p []byte) (n int, err error) {
	if mf.pos >= len(mf.content) {
		return 0, io.EOF
	}
	n = copy(p, mf.content[mf.pos:])
	mf.pos += n
	return n, nil
}

func (mf *MockFile) Close() error {
	return nil
}

func (mf *MockFile) Stat() (fs.FileInfo, error) {
	return &MockFileInfo{name: mf.name, size: int64(len(mf.content))}, nil
}

// MockDirEntry implements fs.DirEntry.
type MockDirEntry struct {
	name string
}

func (mde *MockDirEntry) Name() string {
	return mde.name
}

func (mde *MockDirEntry) IsDir() bool {
	return false
}

func (mde *MockDirEntry) Type() fs.FileMode {
	return 0 // Simplified for this example
}

func (mde *MockDirEntry) Info() (fs.FileInfo, error) {
	return &MockFileInfo{name: mde.name}, nil
}

// MockFileInfo implements fs.FileInfo.
type MockFileInfo struct {
	name string
	size int64
}

func (mfi *MockFileInfo) Name() string {
	return mfi.name
}

func (mfi *MockFileInfo) Size() int64 {
	return mfi.size
}

func (mfi *MockFileInfo) Mode() fs.FileMode {
	return 0 // Simplified for this example
}

func (mfi *MockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (mfi *MockFileInfo) IsDir() bool {
	return false
}

func (mfi *MockFileInfo) Sys() interface{} {
	return nil
}

func mockFileSystemPath(mfs *MockFileSystem) (string, error) {
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		return "", err
	}

	for name, content := range mfs.files {
		err = os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		if err != nil {
			return "", err
		}
	}

	return tempDir, nil
}

func BenchmarkAnalyzeDirectoryParallel(b *testing.B) {
	cache := c.NewParsedFileCache()
	analyzer := &Analyzer{}

	mfs := &MockFileSystem{
		files: map[string]string{
			"file1.go": "package main\n\nfunc main() {\n\t// Hello, world!\n}",
			"file2.go": "package main\n\nfunc add(a, b int) int {\n\treturn a + b\n}",
			"file3.go": "package main\n\nfunc subtract(a, b int) int {\n\treturn a - b\n}\nfunc hello() {\n\t// Hello, world!\n}",
		},
	}

	tempDir, err := mockFileSystemPath(mfs)
	if err != nil {
		b.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Warm-up phase
	for i := 0; i < 5; i++ {
		if err = analyzer.AnalyzeDirectoryParallel(tempDir, cache); err != nil {
			b.Fatalf("Error during warm-up: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		if err = analyzer.AnalyzeDirectoryParallel(tempDir, cache); err != nil {
			b.Fatalf("Error during benchmark: %v", err)
		}
		elapsed := time.Since(start)
		b.Logf("Iteration %d: Time taken: %s", i, elapsed)
	}
}
