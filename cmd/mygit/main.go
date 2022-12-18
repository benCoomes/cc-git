package main

import (
	"fmt"
	"os"
	"unicode/utf8"
)

// Usage: your_safe_git.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		Init()
	case "cat-file":
		CatFile(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func Init() {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/master\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
	}

	fmt.Println("Initialized git directory")
}

func CatFile(args []string) {
	if len(args) != 2 || args[0] != "-p" {
		fmt.Fprint(os.Stderr, "usage: cat-file -p <blob_sha>\n")
		os.Exit(1)
	}

	sha := args[1]

	if utf8.RuneCountInString(sha) != 40 {
		fmt.Fprintf(os.Stderr, "SHA is not valid: '%s'", sha)
	}

	// demo zlib use
	// var b bytes.Buffer
	// w := zlib.NewWriter(&b)
	// w.Write([]byte("hello, world\n"))
	// w.Close()
	// r, _ := zlib.NewReader(&b)
	// io.Copy(os.Stdout, r)
	// r.Close()

	path := fmt.Sprintf(".git/objects/%s/%s", sha[0:2], sha[2:])
	// todo:
	// * read file at path, throw if DNE
	// * unzip read bytes using zlib
	// * parse content - should be "blob <byte_size>\0<content>"
	// * print out content
}
