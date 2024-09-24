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

const version = "0.4.0"

func main() {
	flagHelp := flag.Bool("help", false, "Show help message")
	flagVersion := flag.Bool("version", false, "Show version information")
	flagVerbose := flag.Bool("v", false, "Show version information")
	flag.Parse()

	if *flagHelp {
		help.ShowHelp()
		return
	}

	if *flagVersion || *flagVerbose {
		fmt.Printf("Version: %s\n", version)
		return
	}

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Usage: gometrics <path> or gometrics -help for more information")
		return
	}

	path := args[0]

	analyzer := &a.Analyzer{}
	analyzer.Cache = c.NewParsedFileCache()

	printer.PrintProjectInfo(path)

	//startTime := time.Now()

	if err := analyzer.AnalyzeDirectory(path); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var elapsed time.Duration
	//elapsed = time.Since(startTime)
	printer.PrintAnalysisResults(elapsed, analyzer.TotalLineCount, analyzer.TotalFunctionCount)
}
