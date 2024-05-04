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
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			var lineCount int
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lineCount++
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			totalLineCount += lineCount
			fmt.Printf("Lines in %s: %d; ", filepath.Base(path), lineCount)

			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			functionCount := 0
			ast.Inspect(node, func(n ast.Node) bool {
				switch n.(type) {
				case *ast.FuncDecl:
					functionCount++
				}
				return true
			})
			totalFunctionCount += functionCount
			fmt.Printf("Functions: %d\n", functionCount)
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
	totalLineCount, totalFunctionCount, err := countLinesAndFunctions(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Total lines in .go files: %d\n", totalLineCount)
	fmt.Printf("Total functions in .go files: %d\n", totalFunctionCount)
}
