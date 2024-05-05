package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	c "goMetrics/cache"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	goFileExtension = ".go"
	WorkerPoolSize  = 4
)

type Analyzer struct {
	TotalLineCount     uint
	TotalFunctionCount uint
}

func (a *Analyzer) countLinesInFile(path string) (uint, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	var lineCount uint
	scanner := bufio.NewScanner(file)
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
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return 0, err
	}
	cache.Set(path, node)

	return countFunctionsInAST(node), nil
}

func countFunctionsInAST(node *ast.File) uint {
	var funcCount uint
	ast.Inspect(node, func(n ast.Node) bool {
		_, ok := n.(*ast.FuncDecl)
		if ok {
			funcCount++
		}
		return true
	})
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

	a.printFileAnalysis(path, lineCount, funcCount)
	a.updateTotals(lineCount, funcCount)

	return nil
}

func (a *Analyzer) printFileAnalysis(path string, lineCount uint, funcCount uint) {
	fmt.Printf("Lines in %s: %d; Functions: %d\n", filepath.Base(path), lineCount, funcCount)
}

func (a *Analyzer) updateTotals(lineCount uint, funcCount uint) {
	a.TotalLineCount += lineCount
	a.TotalFunctionCount += funcCount
}

func (a *Analyzer) analyzeDirectoryParallel(dirPath string, cache *c.ParsedFileCache) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	fileChan := make(chan []string, WorkerPoolSize)

	for i := 0; i < WorkerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for files := range fileChan {
				if err := a.processBatch(files, cache); err != nil {
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

func (a *Analyzer) processBatch(files []string, cache *c.ParsedFileCache) error {
	for _, file := range files {
		if err := a.analyzeFile(file, cache); err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) printProjectInfo(path string) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	projectName := filepath.Base(path)
	if path == "." {
		projectName = filepath.Base(currentDir)
	}
	fmt.Println("Project Name:", projectName)
	fmt.Printf("-------------\n")
}

func (a *Analyzer) printAnalysisResults(elapsed time.Duration) {
	fmt.Printf("-------------\n")
	fmt.Printf("Total lines in.go files: %d\n", a.TotalLineCount)
	fmt.Printf("Total functions in.go files: %d\n", a.TotalFunctionCount)
	fmt.Printf("Time taken: %s\n", elapsed)
}

func main() {
	start := time.Now()
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path>")
		return
	}
	path := os.Args[1]

	analyzer := &Analyzer{}
	cache := c.NewParsedFileCache()

	analyzer.printProjectInfo(path)

	if err := analyzer.analyzeDirectoryParallel(path, cache); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	elapsed := time.Since(start)

	analyzer.printAnalysisResults(elapsed)
}
