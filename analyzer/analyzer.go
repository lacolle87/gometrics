package analyzer

import (
	"bufio"
	"bytes"
	"fmt"
	c "github.com/lacolle87/gometrics/cache"
	"github.com/lacolle87/gometrics/printer"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const (
	goFileExtension = ".go"
	workerPoolSize  = 8
)

type Analyzer struct {
	TotalLineCount     uint64
	TotalFunctionCount uint64
	Cache              *c.ParsedFileCache
}

func countFunctionsInAST(path string, fileContent []byte) uint {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, fileContent, parser.DeclarationErrors)
	if err != nil {
		fmt.Printf("Failed to parse file %s: %v", path, err)
		return 0
	}

	var funcCount uint
	ast.Inspect(node, func(n ast.Node) bool {
		if fdecl, ok := n.(*ast.FuncDecl); ok && fdecl.Name != nil {
			funcCount++
		}
		return true
	})

	return funcCount
}

func countLines(file []byte) uint {
	lineCount := uint(0)
	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		if line := scanner.Text(); len(line) > 0 && !isComment(line) {
			lineCount++
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file: %v\n", err)
	}
	return lineCount
}

func isComment(line string) bool {
	return len(line) > 0 && line[0] == '/'
}

func (a *Analyzer) analyzeFile(path string) error {
	fileContent, _ := a.Cache.Get(path)

	lineCount := countLines(fileContent)
	funcCount := countFunctionsInAST(path, fileContent)

	printer.PrintFileAnalysis(path, lineCount, funcCount)

	atomic.AddUint64(&a.TotalLineCount, uint64(lineCount))
	atomic.AddUint64(&a.TotalFunctionCount, uint64(funcCount))

	return nil
}

func (a *Analyzer) AnalyzeDirectoryParallel(dirPath string) error {
	filePaths, err := a.preload(dirPath)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	fileChan := make(chan string, workerPoolSize)

	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				if AnalyzeErr := a.analyzeFile(filePath); AnalyzeErr != nil {
					errChan <- AnalyzeErr
				}
			}
		}()
	}

	for _, filePath := range filePaths {
		fileChan <- filePath
	}
	close(fileChan)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Analyzer) preload(dirPath string) ([]string, error) {
	var filePaths []string
	goFileFound := false
	fileChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(fileChan)
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) == goFileExtension {
				goFileFound = true
				fileChan <- path
			}
			return nil
		})
		if err != nil {
			errChan <- err
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				src, err := os.ReadFile(path)
				if err != nil {
					errChan <- fmt.Errorf("failed to read file %s: %w", path, err)
					continue
				}
				a.Cache.Set(path, src)
				filePaths = append(filePaths, path)
			}
		}()
	}

	wg.Wait()

	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	if !goFileFound {
		return nil, fmt.Errorf("no Go files found in the given directory")
	}

	return filePaths, nil
}
