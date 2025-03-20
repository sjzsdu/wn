package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MarkdownExporter struct {
	*BaseExporter
	content strings.Builder
}

func NewMarkdownExporter(p *Project) *MarkdownExporter {
	return &MarkdownExporter{
		BaseExporter: NewBaseExporter(p),
		content:     strings.Builder{},
	}
}

func (e *MarkdownExporter) ProcessDirectory(node *Node, path string, level int) error {
	if path != "/" {
		indent := strings.Repeat("  ", level-1)
		anchor := strings.ToLower(strings.ReplaceAll(node.Name, " ", "-"))
		e.content.WriteString(fmt.Sprintf("%s- [%s](#%s)\n",
			indent, node.Name, anchor))
	}
	return nil
}

func (e *MarkdownExporter) ProcessFile(node *Node, path string, level int) error {
	indent := strings.Repeat("  ", level-1)
	anchor := strings.ToLower(strings.ReplaceAll(node.Name, " ", "-"))
	e.content.WriteString(fmt.Sprintf("%s- [%s](#%s)\n",
		indent, node.Name, anchor))
	return nil
}

func (e *MarkdownExporter) writeContent() error {
	e.content.WriteString("\n---\n\n")

	var processContent func(node *Node, path string, level int) error
	processContent = func(node *Node, path string, level int) error {
		if node == nil {
			return nil
		}

		node.mu.RLock()
		defer node.mu.RUnlock()

		if node.IsDir {
			if path != "/" {
				e.content.WriteString(fmt.Sprintf("\n%s %s\n\n",
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
			e.content.WriteString(fmt.Sprintf("\n%s %s\n\n",
				strings.Repeat("#", level+1),
				node.Name))
			e.content.WriteString(string(node.Content))
			e.content.WriteString("\n")
		}

		return nil
	}

	return processContent(e.project.root, "/", 1)
}

func (e *MarkdownExporter) Export(outputPath string) error {
	if e.project.root == nil || len(e.project.root.Children) == 0 {
		return os.WriteFile(outputPath, []byte("# 空项目\n"), 0644)
	}

	// 写入标题和目录标记
	e.content.WriteString("# 目录\n\n")

	// 生成目录
	if err := e.TraverseNodes(e.project.root, "/", 1, e); err != nil {
		return err
	}

	// 写入内容
	if err := e.writeContent(); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(outputPath, []byte(e.content.String()), 0644)
}

func (d *Project) ExportToMarkdown(outputPath string) error {
	exporter := NewMarkdownExporter(d)
	return exporter.Export(outputPath)
}
