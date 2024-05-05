package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"time"
)

const goFileExtension = ".go"

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
	if scannerErr := scanner.Err(); err != nil {
		return 0, scannerErr
	}

	return lineCount, nil
}

func (a *Analyzer) countFunctionsInFile(path string) (uint, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return 0, err
	}

	var funcCount uint
	ast.Inspect(node, func(n ast.Node) bool {
		_, ok := n.(*ast.FuncDecl)
		if ok {
			funcCount++
		}
		return true
	})

	return funcCount, nil
}

func (a *Analyzer) analyzeFile(path string) error {
	lineCount, err := a.countLinesInFile(path)
	if err != nil {
		return err
	}

	funcCount, err := a.countFunctionsInFile(path)
	if err != nil {
		return err
	}

	fmt.Printf("Lines in %s: %d; Functions: %d\n", filepath.Base(path), lineCount, funcCount)

	a.TotalLineCount += lineCount
	a.TotalFunctionCount += funcCount

	return nil
}

func (a *Analyzer) analyzeDirectory(dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == goFileExtension {
			if analyzeErr := a.analyzeFile(path); err != nil {
				return analyzeErr
			}
		}
		return nil
	})
	return err
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

	analyzer.printProjectInfo(path)

	if err := analyzer.analyzeDirectory(path); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	elapsed := time.Since(start)

	analyzer.printAnalysisResults(elapsed)
}
