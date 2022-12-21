package main

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strings"
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
	case "hash-object":
		HashObj(os.Args[2:])
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
		// todo: git allows sha prefixes, as long as they are unique
		fmt.Fprintf(os.Stderr, "SHA is not valid: '%s'\n", sha)
		os.Exit(1)
	}

	// todo: check that .git/objects exists
	path := fmt.Sprintf(".git/objects/%s/%s", sha[0:2], sha[2:])
	file, err := os.Open(path)
	check(err)

	// read and unzip file
	r, err := zlib.NewReader(file)
	check(err)
	buf := new(strings.Builder)
	io.Copy(buf, r)
	r.Close()

	// parse content -  "<type> <byte_size>\000<content>"
	parts := strings.Split(buf.String(), "\000")

	fmt.Print(parts[1])
}

func HashObj(args []string) {
	if len(args) != 2 && args[0] != "-w" {
		fmt.Fprint(os.Stderr, "usage: hash-object -w <file_path>\n")
		os.Exit(1)
	}

	path := args[1]

	contents, err := os.ReadFile(path)
	check(err)

	gitobj := []byte(fmt.Sprintf("blob %d\000%s", len(contents), contents))

	sha := sha1.Sum(gitobj)

	dirPath := fmt.Sprintf(".git/objects/%x", sha[0:1])
	err = os.MkdirAll(dirPath, 0755)
	check(err)

	outPath := fmt.Sprintf("%s/%x", dirPath, sha[1:])
	file, err := os.Create(outPath)
	check(err)
	defer file.Close()
	zlibWriter := zlib.NewWriter(file)
	zlibWriter.Write(gitobj)
	zlibWriter.Close()

	// print sha
	fmt.Printf("%x\n", sha)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
