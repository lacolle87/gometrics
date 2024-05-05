package printer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func PrintFileAnalysis(path string, lineCount uint, funcCount uint) {
	fmt.Printf("Lines in %s: %d; Functions: %d\n", filepath.Base(path), lineCount, funcCount)
}

func PrintProjectInfo(path string) {
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

func PrintAnalysisResults(elapsed time.Duration, tlc uint, tfc uint) {
	fmt.Printf("-------------\n")
	fmt.Printf("Total lines in.go files: %d\n", tlc)
	fmt.Printf("Total functions in.go files: %d\n", tfc)
	fmt.Printf("Time taken: %s\n", elapsed)
}