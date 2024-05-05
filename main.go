package main

import (
	"flag"
	"fmt"
	a "goMetrics/analyzer"
	c "goMetrics/cache"
	"goMetrics/printer"
	"time"
)

func main() {
	timed := flag.Bool("t", false, "Measure execution time")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Usage: go run main.go [-t] <path>")
		return
	}

	path := args[0]

	if *timed && len(args) != 1 {
		fmt.Println("Usage: go run main.go -t <path>")
		return
	}

	analyzer := &a.Analyzer{}
	analyzer.Cache = c.NewParsedFileCache()

	printer.PrintProjectInfo(path)

	if err := analyzer.AnalyzeDirectoryParallel(path); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var elapsed time.Duration
	if *timed {
		elapsed = time.Since(time.Now())
	}
	printer.PrintAnalysisResults(elapsed, analyzer.TotalLineCount, analyzer.TotalFunctionCount)
}
