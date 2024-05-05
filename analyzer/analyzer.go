package analyzer

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	c "goMetrics/cache"
	"goMetrics/printer"
	"os"
	"path/filepath"
	"sync"
)

const (
	goFileExtension = ".go"
	workerPoolSize  = 8
)

type Analyzer struct {
	TotalLineCount     uint
	TotalFunctionCount uint
}

func (a *Analyzer) countLinesInFile(path string) (uint, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	var lineCount uint
	buf := make([]byte, 4096)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, 0)

	for scanner.Scan() {
		lineCount++
	}
	if scannerErr := scanner.Err(); scannerErr != nil {
		return 0, scannerErr
	}

	return lineCount, nil
}

func (a *Analyzer) countFunctionsInFile(path string, cache *c.ParsedFileCache) (uint, error) {
	if file, ok := cache.Get(path); ok {
		return countFunctionsInAST(file), nil
	}

	fset := token.NewFileSet()
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	node, err := parser.ParseFile(fset, path, src, parser.DeclarationErrors)
	if err != nil {
		return 0, err
	}
	cache.Set(path, node)

	var funcCount uint
	for _, decl := range node.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok {
			if fdecl.Name != nil {
				funcCount++
			}
		}
	}
	return funcCount, nil
}

func countFunctionsInAST(node *ast.File) uint {
	var funcCount uint
	for _, decl := range node.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name != nil {
			funcCount++
		}
	}
	return funcCount
}

func (a *Analyzer) analyzeFile(path string, cache *c.ParsedFileCache) error {
	lineCount, err := a.countLinesInFile(path)
	if err != nil {
		return err
	}

	funcCount, err := a.countFunctionsInFile(path, cache)
	if err != nil {
		return err
	}

	printer.PrintFileAnalysis(path, lineCount, funcCount)
	a.updateTotals(lineCount, funcCount)

	return nil
}

func (a *Analyzer) updateTotals(lineCount uint, funcCount uint) {
	a.TotalLineCount += lineCount
	a.TotalFunctionCount += funcCount
}

func (a *Analyzer) AnalyzeDirectoryParallel(dirPath string, cache *c.ParsedFileCache) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	fileChan := make(chan []string, workerPoolSize)

	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for files := range fileChan {
				if err := a.analyzeFile(files[0], cache); err != nil {
					errChan <- err
				}
			}
		}()
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == goFileExtension {
			fileChan <- []string{path}
		}
		return nil
	})
	if err != nil {
		return err
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
