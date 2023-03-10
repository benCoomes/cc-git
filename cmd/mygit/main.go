package main

import (
	"compress/zlib"
	"crypto/sha1"
	"errors"
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
	case "ls-tree":
		ListTree((os.Args[2:]))
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

func ListTree(args []string) {
	if len(args) != 2 || args[0] != "--name-only" {
		fmt.Fprintf(os.Stderr, "usage: ls-tree --name-only <tree_sha>\n")
		os.Exit(1)
	}

	sha := args[1]
	gobj, err := readGitObject(sha)
	check(err)

	// internal structure from https://stackoverflow.com/questions/14790681/what-is-the-internal-format-of-a-git-tree-object
	// tree [content size]\0[entries]
	// each entry is:
	// [mode] [file/folder name]\0[SHA-1 of referencing blob or tree]
	// Note that entries are not seperated by anything. Ex: [mode] [name]\0[SHA][mode] [name]\0[SHA]
	// So, splitting entries on \0 splits across entries always puts name at end of substrings, with no name in last substring

	// todo: objects may contain nil bytes, so use length to read rather than splitting
	parts := strings.Split(gobj, "\000")
	entries := parts[1:]
	for i := 0; i < len(entries)-1; i += 1 {
		entry_parts := strings.SplitN(entries[i], " ", 2)
		obj_name := entry_parts[1]
		fmt.Println(obj_name)
	}
}

func CatFile(args []string) {
	if len(args) != 2 || args[0] != "-p" {
		fmt.Fprint(os.Stderr, "usage: cat-file -p <blob_sha>\n")
		os.Exit(1)
	}

	sha := args[1]
	gobj, err := readGitObject(sha)
	check(err)

	// parse content -  "<type> <byte_size>\000<content>"
	// todo: objects may contain nil bytes, so use length to read rather than splitting
	parts := strings.Split(gobj, "\000")
	for _, part := range parts[1:] {
		fmt.Print(part)
	}
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

func readGitObject(sha string) (string, error) {
	if utf8.RuneCountInString(sha) != 40 {
		// todo: git allows sha prefixes, as long as they are unique
		msg := fmt.Sprintf("SHA is not valid: '%s'\n", sha)
		return "", errors.New(msg)
	}

	// todo: check that .git/objects exists
	path := fmt.Sprintf(".git/objects/%s/%s", sha[0:2], sha[2:])
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	r, err := zlib.NewReader(file)
	if err != nil {
		return "", err
	}

	buf := new(strings.Builder)
	io.Copy(buf, r)
	r.Close()

	return buf.String(), nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
