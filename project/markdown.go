package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExportToMarkdown 将项目导出为 Markdown 文件
func (d *Project) ExportToMarkdown(outputPath string) error {
	if d.root == nil || len(d.root.Children) == 0 {
		// 如果是空项目，创建一个空的 markdown 文件
		return os.WriteFile(outputPath, []byte("# 空项目\n"), 0644)
	}

	var content strings.Builder

	// 写入标题和目录标记
	content.WriteString("# 目录\n\n")

	// 处理节点的函数
	var processNode func(node *Node, path string, level int) error
	processNode = func(node *Node, path string, level int) error {
		if node == nil {
			return nil
		}

		node.mu.RLock()
		defer node.mu.RUnlock()

		if node.IsDir {
			if path != "/" {
				// 为目录添加标题
				indent := strings.Repeat("  ", level-1)
				content.WriteString(fmt.Sprintf("%s- [%s](#%s)\n",
					indent,
					node.Name,
					strings.ToLower(strings.ReplaceAll(node.Name, " ", "-"))))
			}

			for _, child := range node.Children {
				childPath := filepath.Join(path, child.Name)
				if err := processNode(child, childPath, level+1); err != nil {
					return err
				}
			}
		} else {
			// 为文件添加目录项
			indent := strings.Repeat("  ", level-1)
			content.WriteString(fmt.Sprintf("%s- [%s](#%s)\n",
				indent,
				node.Name,
				strings.ToLower(strings.ReplaceAll(node.Name, " ", "-"))))
		}

		return nil
	}

	// 处理目录树生成目录
	if err := processNode(d.root, "/", 1); err != nil {
		return err
	}

	content.WriteString("\n---\n\n") // 添加分隔线

	// 第二次遍历，添加实际内容
	var processContent func(node *Node, path string, level int) error
	processContent = func(node *Node, path string, level int) error {
		if node == nil {
			return nil
		}

		node.mu.RLock()
		defer node.mu.RUnlock()

		if node.IsDir {
			if path != "/" {
				// 添加目录标题
				content.WriteString(fmt.Sprintf("\n%s %s\n\n",
					strings.Repeat("#", level+1),
					node.Name))
			}

			for _, child := range node.Children {
				childPath := filepath.Join(path, child.Name)
				if err := processContent(child, childPath, level+1); err != nil {
					return err
				}
			}
		} else {
			// 添加文件标题和内容
			content.WriteString(fmt.Sprintf("\n%s %s\n\n",
				strings.Repeat("#", level+1),
				node.Name))
			content.WriteString(string(node.Content))
			content.WriteString("\n")
		}

		return nil
	}

	// 处理内容
	if err := processContent(d.root, "/", 1); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}
