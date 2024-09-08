package help

import "fmt"

func ShowHelp() {
	fmt.Println(`Usage: gometrics [-t] <path>

GoMetrics counts the total number of lines and functions in .go files within a specified directory.

Options:
  -t        Measure execution time
  --help    Show this help message and exit

Examples:
  gometrics <path>
  gometrics .    # To analyze the current directory

Output:
  The output displays the total lines and functions in each .go file within the specified directory, along with the total for the entire project.`)
}
