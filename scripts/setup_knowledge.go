package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Process both repositories
	sources := []struct {
		path   string
		folder string
	}{
		{
			path:   "docs/typespec/website/src/content/docs/docs",
			folder: "typespec_docs",
		},
		{
			path:   "docs/typespec-azure/website/src/content/docs/docs",
			folder: "typespec_azure_docs",
		},
	}

	for _, source := range sources {
		sourceDir := source.path
		targetDir := "temp_docs/" + source.folder
		// Create target directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("Error creating target directory: %v\n", err)
			return
		}

		// Walk through all markdown files in source directory
		err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only process markdown files
			if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mdx")) {
				if err := processMarkdownFile(path, sourceDir, targetDir); err != nil {
					fmt.Printf("Error processing file %s: %v\n", path, err)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking through directory: %v\n", err)
			return
		}
	}
}

func processMarkdownFile(filePath, sourceDir, targetDir string) error {
	// Read source file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Generate new filename: replace path separators with underscores
	relPath, err := filepath.Rel(sourceDir, filePath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %w", err)
	}

	// Replace path separators with underscores to create unique filename
	newFileName := strings.ReplaceAll(relPath, string(os.PathSeparator), "_")
	targetPath := filepath.Join(targetDir, newFileName)

	// Create target file
	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("error creating target file: %w", err)
	}
	defer outFile.Close()

	// Process file content
	return convertMarkdown(file, outFile)
}

func convertMarkdown(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	var title string
	inFrontmatter := false
	foundTitle := false
	firstContentLine := true

	for scanner.Scan() {
		line := scanner.Text()

		// Process frontmatter
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				inFrontmatter = false
				continue
			}
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "title:") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				// Remove possible quotes
				title = strings.Trim(title, "\"'")
				foundTitle = true
			}
			continue
		}

		// Add title at the beginning of file content
		if !inFrontmatter && firstContentLine {
			if foundTitle {
				if _, err := writer.WriteString(fmt.Sprintf("# %s\n\n", title)); err != nil {
					return fmt.Errorf("error writing title: %w", err)
				}
			}
			firstContentLine = false
		}

		// Write non-empty lines
		if !inFrontmatter {
			if _, err := writer.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("error writing line: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %w", err)
	}

	return nil
}
