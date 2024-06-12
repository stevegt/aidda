package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Function to handle running tests
func runTests(args []string) {
	if len(args) < 1 {
		fmt.Println("Missing arguments for runTests")
		return
	}
	packagePath := args[0]
	result, err := runGoTests(packagePath)
	if err != nil {
		fmt.Println("Error running tests:", err)
	} else {
		fmt.Println(result)
	}
}

// Function to handle fetching lines from a file
func fetchLinesFromFile(args []string) {
	if len(args) < 3 {
		fmt.Println("Missing arguments for fetchLinesFromFile")
		return
	}
	filePath := args[0]
	startLine, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid start line:", args[1])
		return
	}
	endLine, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("Invalid end line:", args[2])
		return
	}
	result, err := readFile(filePath, startLine, endLine)
	if err != nil {
		fmt.Println("Error reading file:", err)
	} else {
		fmt.Println(result)
	}
}

// Function to handle querying gopls
func queryGopls(args []string) {
	result, err := runGoplsCommand(args...)
	if err != nil {
		fmt.Println("Error querying gopls:", err)
	} else {
		fmt.Println(result)
	}
}

// Main function to dispatch actions
func main() {
	if len(os.Args) < 2 {
		fmt.Println("No action specified")
		return
	}
	action := os.Args[1]
	args := os.Args[2:]

	switch action {
	case "runTests":
		runTests(args)
	case "fetchLinesFromFile":
		fetchLinesFromFile(args)
	case "queryGopls":
		queryGopls(args)
	// Add other actions here
	default:
		fmt.Println("Unknown action:", action)
	}
}

// Helper function to run go tests
func runGoTests(packagePath string) (string, error) {
	cmd := exec.Command("go", "test", "-v", packagePath)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper function to read specific lines from a file
func readFile(path string, startLine int, endLine int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		if lineNumber >= startLine && lineNumber <= endLine {
			result.WriteString(scanner.Text() + "\n")
		}
		if lineNumber > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}

// Helper function to run gopls command
func runGoplsCommand(args ...string) (string, error) {
	cmd := exec.Command("gopls", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
