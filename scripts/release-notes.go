package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: release-notes CHANGELOG.md vX.Y.Z")
		os.Exit(2)
	}
	notes, err := releaseNotes(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(notes)
}

func releaseNotes(path, tag string) (string, error) {
	version := strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if version == "" {
		return "", fmt.Errorf("release tag is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	needle := "## " + version
	var out strings.Builder
	inSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			if inSection {
				break
			}
			if line == needle || strings.HasPrefix(line, needle+" ") {
				inSection = true
				out.WriteString(line)
				out.WriteByte('\n')
			}
			continue
		}
		if inSection {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if !inSection {
		return "", fmt.Errorf("release notes for %s not found in %s", tag, path)
	}
	notes := strings.TrimSpace(out.String()) + "\n"
	if notes == "\n" {
		return "", fmt.Errorf("release notes for %s are empty", tag)
	}
	return notes, nil
}
