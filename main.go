package main

import (
	"fmt"
	a "goMetrics/analyzer"
	c "goMetrics/cache"
	"goMetrics/printer"
	"os"
	"time"
)

func main() {
	start := time.Now()
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path>")
		return
	}
	path := os.Args[1]

	analyzer := &a.Analyzer{}
	cache := c.NewParsedFileCache()

	printer.PrintProjectInfo(path)

	if err := analyzer.AnalyzeDirectoryParallel(path, cache); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	elapsed := time.Since(start)

	printer.PrintAnalysisResults(elapsed, analyzer.TotalLineCount, analyzer.TotalFunctionCount)
}
