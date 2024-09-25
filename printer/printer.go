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

func PrintAnalysisResults(elapsed time.Duration, tlc uint64, tfc uint64) {
	fmt.Printf("-------------\n")
	fmt.Printf("Total lines: %d; Total Functions: %d\n", tlc, tfc)
	if elapsed > 0 {
		fmt.Printf("Time taken: %.4f seconds\n", elapsed.Seconds())
	}
}
