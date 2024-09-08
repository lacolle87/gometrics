package main

import (
	"flag"
	"fmt"
	a "github.com/lacolle87/gometrics/analyzer"
	c "github.com/lacolle87/gometrics/cache"
	"github.com/lacolle87/gometrics/help"
	"github.com/lacolle87/gometrics/printer"
	"time"
)

func main() {
	timed := flag.Bool("t", false, "Measure execution time")
	flagHelp := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *flagHelp {
		help.ShowHelp()
		return
	}

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
