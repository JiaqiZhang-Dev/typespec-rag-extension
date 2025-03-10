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
	// 设置源目录和目标目录
	sourceDir := "data"
	targetDir := "doc"

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating target directory: %v\n", err)
		return
	}

	// 遍历源目录下的所有markdown文件
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理markdown文件
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

func processMarkdownFile(filePath, sourceDir, targetDir string) error {
	// 读取源文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// 生成新的文件名：用路径替换斜杠为下划线
	relPath, err := filepath.Rel(sourceDir, filePath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %w", err)
	}
	
	// 将路径中的斜杠替换为下划线，创建唯一文件名
	newFileName := strings.ReplaceAll(relPath, string(os.PathSeparator), "_")
	targetPath := filepath.Join(targetDir, newFileName)

	// 创建目标文件
	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("error creating target file: %w", err)
	}
	defer outFile.Close()

	// 处理文件内容
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

		// 处理frontmatter
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
				// 移除可能的引号
				title = strings.Trim(title, "\"'")
				foundTitle = true
			}
			continue
		}

		// 在文件内容开始处添加标题
		if !inFrontmatter && firstContentLine {
			if foundTitle {
				if _, err := writer.WriteString(fmt.Sprintf("# %s\n\n", title)); err != nil {
					return fmt.Errorf("error writing title: %w", err)
				}
			}
			firstContentLine = false
		}

		// 写入非空行
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
