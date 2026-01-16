package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startWebServer()
		return
	}

	// REPL mode
	db := NewDatabase()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("SimpleDB RDBMS v1.0")
	fmt.Println("Enter SQL commands (type 'exit' to quit)")
	fmt.Print("> ")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			fmt.Print("> ")
			continue
		}

		if strings.ToLower(line) == "exit" {
			break
		}

		result, err := db.Execute(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if result != nil {
			result.Print()
		}

		fmt.Print("> ")
	}
}
