package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func processFile(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	lineCount := countLinesInFile(file)
	functionCount := countFunctionsInFile(path)

	return lineCount, functionCount, nil
}

func countLinesInFile(file *os.File) int {
	scanner := bufio.NewScanner(file)
	var lineCount int
	for scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return 0
	}
	return lineCount
}

func countFunctionsInFile(path string) int {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return 0
	}
	functionCount := 0
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			functionCount++
		}
		return true
	})
	return functionCount
}

func countLinesAndFunctions(path string) (int, int, error) {
	var totalLineCount, totalFunctionCount int
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".go" {
			lineCount, functionCount, err := processFile(path)
			if err != nil {
				return err
			}
			totalLineCount += lineCount
			totalFunctionCount += functionCount
			fmt.Printf("Lines in %s: %d; Functions: %d\n", filepath.Base(path), lineCount, functionCount)
		}
		return nil
	})
	return totalLineCount, totalFunctionCount, err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path>")
		return
	}
	path := os.Args[1]

	projectName := filepath.Base(path)
	fmt.Println("Project Name:", projectName)

	totalLineCount, totalFunctionCount, err := countLinesAndFunctions(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("-------------\n")
	fmt.Printf("Total lines in.go files: %d\n", totalLineCount)
	fmt.Printf("Total functions in.go files: %d\n", totalFunctionCount)
	fmt.Printf("-------------\n")
}
