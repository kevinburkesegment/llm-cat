# llm-cat

A better way to show multiple files to LLMs.

## Why?

`cat` just dumps file contents. When you show multiple files to an LLM, it can't tell where one ends and another begins.

**With regular cat:**
```
$ cat file1.go file2.go
package main
func foo() {}package main
func bar() {}
```

**With llm-cat:**
```
$ llm-cat file1.go file2.go

--- file1.go ---
package main
func foo() {}

--- file2.go ---
package main
func bar() {}
```

Clear boundaries. No confusion.

## Install

```bash
go install github.com/kevinburkesegment/llm-cat@latest
```

Or build from source:
```bash
go build -o llm-cat main.go
sudo mv llm-cat /usr/local/bin/
```

## Usage

### Basic
```bash
llm-cat file1.go file2.py file3.js
```

### Recursive directories
```bash
llm-cat -r src/
```

### Filter by extension
```bash
llm-cat -r -ext .go ./
```

### With pipes
```bash
find . -name "*.md" | llm-cat
git ls-files | grep test | llm-cat
ls *.py | xargs llm-cat
```

## Examples

Show all Go files in current directory:
```bash
llm-cat *.go
```

Show all Python files recursively:
```bash
llm-cat -r -ext .py .
```

Show specific files from git:
```bash
git diff --name-only main | llm-cat
```

## Output Format

Each file is wrapped with clear delimiters:
```
--- path/to/file.ext ---
[file contents]
```

Perfect for copying into LLM chats.

## License

MIT
