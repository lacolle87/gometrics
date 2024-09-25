package analyzer

import (
	c "github.com/lacolle87/gometrics/cache"
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

func mockFileSystem(files map[string]string) (string, error) {
	tempDir, err := os.MkdirTemp("", "preload_benchmark")
	if err != nil {
		return "", err
	}

	for name, content := range files {
		err = os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		if err != nil {
			return "", err
		}
	}

	return tempDir, nil
}

func BenchmarkAnalyzeDirectoryParallel(b *testing.B) {
	analyzer := &Analyzer{}
	analyzer.Cache = c.NewParsedFileCache()

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		if err = analyzer.AnalyzeDirectory(tempDir); err != nil {
			b.Fatalf("Error during benchmark: %v", err)
		}
		elapsed := time.Since(start)
		b.Logf("Iteration %d: Time taken: %s", i, elapsed)
	}
}

func BenchmarkPreload(b *testing.B) {
	analyzer := &Analyzer{
		Cache: c.NewParsedFileCache(),
	}

	mockFiles := map[string]string{
		"file1.go": "package main\n\nfunc main() {\n\t// Hello, world!\n}",
		"file2.go": "package main\n\nfunc add(a, b int) int {\n\treturn a + b\n}",
		"file3.go": "package main\n\nfunc subtract(a, b int) int {\n\treturn a - b\n}",
	}

	tempDir, err := mockFileSystem(mockFiles)
	if err != nil {
		b.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileChan := make(chan string)
		errChan := make(chan error, 1)

		go func() {
			analyzer.preload(tempDir, fileChan, errChan)
			close(errChan)
		}()

		done := false
		for !done {
			select {
			case path, ok := <-fileChan:
				if !ok {
					fileChan = nil
				} else {
					b.Logf("Processed file: %s", path)
				}
			case err := <-errChan:
				if err != nil {
					b.Fatalf("Error during preload: %v", err)
				}
				done = true
			}
		}
	}
}
