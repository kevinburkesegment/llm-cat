package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const defaultMaxSize = 10 << 20 // 10 MiB

func main() {
	var (
		recurse   = flag.Bool("r", false, "Recursively process directories")
		extension = flag.String("ext", "", "Only process files with this extension (e.g., .go, .txt)")
		namesOnly = flag.Bool("n", false, "Only print file names, not their contents")
		maxSize   = flag.Int64("max-size", defaultMaxSize, "Maximum number of bytes to output (0 = unlimited)")
		help      = flag.Bool("h", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	files := flag.Args()
	if len(files) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if f := strings.TrimSpace(scanner.Text()); f != "" {
				files = append(files, f)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	for _, f := range files {
		if err := processPath(f, *recurse, *extension, *namesOnly, *maxSize); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", f, err)
		}
	}
}

func processPath(path string, recurse bool, extension string, namesOnly bool, maxSize int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if !recurse {
			return fmt.Errorf("'%s' is a directory (use -r to recurse)", path)
		}
		return filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !i.IsDir() && matchesExtension(p, extension) {
				return handleFile(p, namesOnly, maxSize)
			}
			return nil
		})
	}

	if matchesExtension(path, extension) {
		return handleFile(path, namesOnly, maxSize)
	}
	return nil
}

func handleFile(path string, namesOnly bool, maxSize int64) error {
	if namesOnly {
		fmt.Println(path)
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if maxSize > 0 && info.Size() > maxSize {
		fmt.Fprintf(os.Stderr, "Skipping %s (size %d bytes exceeds limit %d)\n", path, info.Size(), maxSize)
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Detect binary by sampling first 8 kB
	const sampleSize = 8 << 10
	sample := make([]byte, sampleSize)
	n, _ := io.ReadFull(file, sample)
	if isBinary(sample[:n]) {
		fmt.Fprintf(os.Stderr, "Skipping binary file %s\n", path)
		return nil
	}
	// Rewind after sampling
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Print delimiter and contents
	fmt.Printf("\n--- %s ---\n", path)
	if _, err := io.Copy(os.Stdout, file); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func isBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	nonPrintable := 0
	for _, b := range data {
		r := rune(b)
		if (r == '\n') || (r == '\r') || (r == '\t') {
			continue
		}
		if r == 0 || !unicode.IsPrint(r) {
			nonPrintable++
		}
	}
	// Heuristic: >10 % non-printable → treat as binary
	return nonPrintable*10 > len(data)
}

func matchesExtension(path, extension string) bool {
	if extension == "" {
		return true
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return strings.HasSuffix(strings.ToLower(path), strings.ToLower(extension))
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
	fmt.Println("  -r               Recursively process directories")
	fmt.Println("  -ext string      Only process files with this extension")
	fmt.Println("  -n               Only print file names, not contents")
	fmt.Println("  -max-size bytes  Maximum bytes to show (default 10485760, 0 = unlimited)")
	fmt.Println("  -h               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  llm-cat file1.txt file2.go")
	fmt.Println("  llm-cat -r -ext .go src/")
	fmt.Println("  llm-cat -n $(git ls-files)")
	fmt.Println("  find . -type f -size -20M | llm-cat")
	fmt.Println()
	fmt.Println("Output format when dumping:")
	fmt.Println("  --- filename.go ---")
	fmt.Println("  [file contents]")
	fmt.Println()
}
