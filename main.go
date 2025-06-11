package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		recurse   = flag.Bool("r", false, "Recursively process directories")
		extension = flag.String("ext", "", "Only process files with this extension (e.g., .go, .txt)")
		help      = flag.Bool("h", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Get list of files to process
	files := flag.Args()

	// If no arguments provided, read from stdin (for xargs compatibility)
	if len(files) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			file := strings.TrimSpace(scanner.Text())
			if file != "" {
				files = append(files, file)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	// Process each file
	for _, file := range files {
		if err := processPath(file, *recurse, *extension); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
		}
	}
}

func processPath(path string, recurse bool, extension string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if !recurse {
			return fmt.Errorf("'%s' is a directory (use -r to recurse)", path)
		}
		return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && matchesExtension(p, extension) {
				return printFile(p)
			}
			return nil
		})
	}

	if matchesExtension(path, extension) {
		return printFile(path)
	}
	return nil
}

func matchesExtension(path, extension string) bool {
	if extension == "" {
		return true
	}
	// Ensure extension starts with a dot
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return strings.HasSuffix(strings.ToLower(path), strings.ToLower(extension))
}

func printFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Print delimiter with filename
	fmt.Printf("\n--- %s ---\n", path)

	// Copy file contents
	_, err = io.Copy(os.Stdout, file)
	if err != nil {
		return err
	}

	// Ensure we end with a newline
	fmt.Println()
	return nil
}

func showHelp() {
	fmt.Println("llm-cat - Display files in an LLM-friendly format")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  llm-cat [flags] [files...]")
	fmt.Println("  command | xargs llm-cat [flags]")
	fmt.Println("  find . -name '*.go' | llm-cat")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -r          Recursively process directories")
	fmt.Println("  -ext string Only process files with this extension")
	fmt.Println("  -h          Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  llm-cat file1.txt file2.go")
	fmt.Println("  llm-cat -r -ext .go src/")
	fmt.Println("  find . -type f -name '*.md' | llm-cat")
	fmt.Println("  ls *.py | xargs llm-cat")
	fmt.Println()
	fmt.Println("Output format:")
	fmt.Println("  --- filename.go ---")
	fmt.Println("  [file contents]")
	fmt.Println()
}
