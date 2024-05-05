# GoMetrics

This Go package provides a simple utility to count the total number of lines and functions in `.go` files within a specified directory. It's useful for getting a quick overview of the size and complexity of a Go project.

# Features

- Counts the total number of lines in `.go` files.
- Counts the total number of functions in `.go` files.
- Supports counting metrics for all `.go` files in a specified directory and its subdirectories.

# Usage

To use this package, run the `main.go` file with the path to the directory you want to analyze as an argument. If you want to analyze the current directory, use `.`.

```bash
go run main.go <path>
```
or
```bash
go run main.go .
```

## Build

You can build GoMetrics and use it in Linux by following these steps:

   ```bash
   go build -o gometrics main.go
   ```
Set executable permissions:

```bash
chmod +x gometrics
```
Run the executable with the desired path:

```bash
./goMetrics <path>
```
Replace `<path>` with the path to the directory you want to analyze. If you want to analyze the current directory, use ..

The **-t** flag can be used to measure the execution time of the analysis. Include it in the command-line arguments as follows:

```bash
./goMetrics -t <path>
```

## Output

The output will look something like this:
```
Project Name: myproject
-------------
Lines in main.go: 100; Functions: 20 
Lines in utils.go: 50; Functions: 10
-------------
Total lines: 150; Total functions: 30
```


## Requirements

- Go 1.22 or later
