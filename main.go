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

const GoFileExtension = ".go"

type Analyzer struct {
	TotalLineCount     int
	TotalFunctionCount int
}

func (a *Analyzer) countLinesInFile(file *os.File) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		a.TotalLineCount++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) countFunctionsInFile(path string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			a.TotalFunctionCount++
		}
		return true
	})
	return nil
}

func (a *Analyzer) processFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	if err := a.countLinesInFile(file); err != nil {
		return err
	}
	if err := a.countFunctionsInFile(path); err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) countLinesAndFunctions(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == GoFileExtension {
			if err := a.processFile(path); err != nil {
				return err
			}
			fmt.Printf("Lines in %s: %d; Functions: %d\n", filepath.Base(path), a.TotalLineCount, a.TotalFunctionCount)
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

	if err := analyzer.countLinesAndFunctions(path); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	elapsed := time.Since(start)

	analyzer.printAnalysisResults(elapsed)
}
