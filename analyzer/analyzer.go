package analyzer

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/karrick/godirwalk"
	c "github.com/lacolle87/gometrics/cache"
	"github.com/lacolle87/gometrics/printer"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const (
	goFileExtension = ".go"
	workerPoolSize  = 16
	fileChanBuffer  = 100
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
	reader := bufio.NewReader(bytes.NewReader(file))
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading file: %v\n", err)
		}

		if len(line) > 0 && !isComment(string(line)) {
			lineCount++
		}
	}
	return lineCount
}

func isComment(line string) bool {
	return len(line) > 0 && (line[0] == '/' || line[0] == '#')
}

func (a *Analyzer) analyzeGoFile(path string) error {
	fileContent, _ := a.Cache.Get(path)

	if bytes.IndexByte(fileContent, 0) != -1 {
		return nil
	}

	if !bytes.HasPrefix(fileContent, []byte("package ")) {
		return nil
	}

	lineCount := countLines(fileContent)
	funcCount := countFunctionsInAST(path, fileContent)

	printer.PrintFileAnalysis(path, lineCount, funcCount)

	atomic.AddUint64(&a.TotalLineCount, uint64(lineCount))
	atomic.AddUint64(&a.TotalFunctionCount, uint64(funcCount))

	return nil
}

func (a *Analyzer) AnalyzeDirectory(dirPath string) error {
	fileChan := make(chan string, fileChanBuffer)
	errChan := make(chan error, 1)

	go a.preload(dirPath, fileChan, errChan)

	var wg sync.WaitGroup
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				if err := a.analyzeGoFile(filePath); err != nil {
					errChan <- err
				}
			}
		}()
	}

	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
	}
	return nil
}

func (a *Analyzer) preload(dirPath string, fileChan chan<- string, errChan chan<- error) {
	defer close(fileChan)

	var wg sync.WaitGroup
	filePaths := make(chan string, fileChanBuffer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := godirwalk.Walk(dirPath, &godirwalk.Options{
			Callback: func(fullPath string, de *godirwalk.Dirent) error {
				if de.IsDir() || filepath.Ext(fullPath) != goFileExtension {
					return nil
				}
				filePaths <- fullPath
				return nil
			},
			Unsorted: true,
		})
		if err != nil {
			errChan <- fmt.Errorf("error walking the path %s: %w", dirPath, err)
		}
		close(filePaths)
	}()

	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filePaths {
				if err := a.readAndCacheFile(path); err != nil {
					errChan <- err
				} else {
					fileChan <- path
				}
			}
		}()
	}
	wg.Wait()
}

func (a *Analyzer) readAndCacheFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	a.Cache.Set(path, src)
	return nil
}
