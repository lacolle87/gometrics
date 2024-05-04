# GoMetrics

This Go package provides a simple utility to count the total number of lines and functions in `.go` files within a specified directory. It's useful for getting a quick overview of the size and complexity of a Go project.

## Features

- Counts the total number of lines in `.go` files.
- Counts the total number of functions in `.go` files.
- Supports counting metrics for all `.go` files in a specified directory and its subdirectories.

## Usage

To use this package, run the `main.go` file with the path to the directory you want to analyze as an argument.

```bash
go run main.go <path>
```

Replace `<path>` with the path to the directory you want to analyze. If you want to analyze the current directory, use `.`.

## Example

```bash
go run main.go .
```

This command will analyze the current directory and print the total number of lines and functions in all `.go` files found.

## Output

The output will look something like this:
```
Project Name: myproject
-------------
Lines in main.go: 100; Functions: 20 
Lines in utils.go: 50; Functions: 10
-------------
Total lines in .go files: 150 
Total functions in .go files: 30
```


## Requirements

- Go 1.16 or later