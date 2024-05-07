package analyzer

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	c "gometrics/cache"
	"gometrics/printer"
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
	Cache              *c.ParsedFileCache
}

func countFunctionsInAST(path string, fileContent []byte) uint {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, fileContent, parser.DeclarationErrors)
	if err != nil {
		fmt.Printf("Failed to parse file %s: %v", path, err)
	}

	var funcCount uint
	for _, decl := range node.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name != nil {
			funcCount++
		}
	}
	return funcCount
}

func countLines(file []byte) uint {
	lineCount := uint(0)
	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file: %v\n", err)
	}
	return lineCount
}

func (a *Analyzer) analyzeFile(path string) error {
	fileContent, _ := a.Cache.Get(path)

	lineCount := countLines(fileContent)
	funcCount := countFunctionsInAST(path, fileContent)

	printer.PrintFileAnalysis(path, lineCount, funcCount)

	a.updateTotals(lineCount, funcCount)

	return nil
}

func (a *Analyzer) updateTotals(lineCount uint, funcCount uint) {
	a.TotalLineCount += lineCount
	a.TotalFunctionCount += funcCount
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
					return
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
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == goFileExtension {
			goFileFound = true
			src, ReadFileErr := os.ReadFile(path)
			if ReadFileErr != nil {
				return fmt.Errorf("failed to read file %s: %w", path, ReadFileErr)
			}
			a.Cache.Set(path, src)
			filePaths = append(filePaths, path)
		}

		return nil
	})
	if !goFileFound {
		return nil, fmt.Errorf("no go files found in the given directory")
	}

	return filePaths, err
}
