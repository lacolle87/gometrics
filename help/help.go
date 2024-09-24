package help

import "fmt"

func ShowHelp() {
	fmt.Println(`Usage: gometrics <path>

GoMetrics counts the total number of lines and functions in .go files within a specified directory.

Options:
  -help    Show this help message and exit
  -v        Show version information

Examples:
  gometrics <path>
  gometrics .    # To analyze the current directory

Output:
  The output displays the total lines and functions in each .go file within the specified directory, along with the total for the entire project.`)
}
