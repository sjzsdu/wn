package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/sjzsdu/wn/helper"
)

// tocItem 表示目录项
type tocItem struct {
	title string
	page  int
	level int
}

func (d *Project) ExportToPDF(outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// 设置字体
	fontPath, err := helper.FindFont()
	if err != nil {
		return fmt.Errorf("error finding suitable font: %v", err)
	}

	fontName := filepath.Base(fontPath)
	fontName = fontName[:len(fontName)-len(filepath.Ext(fontName))] // Remove extension
	fontName = strings.ReplaceAll(fontName, " ", "")                // Remove spaces from font name

	// 读取字体文件
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return fmt.Errorf("error reading font file: %v", err)
	}

	// 添加字体
	pdf.AddUTF8FontFromBytes(fontName, "", fontData)
	pdf.SetFont(fontName, "", 12)

	// 确保至少有一个页面
	if d.root == nil || len(d.root.Children) == 0 {
		pdf.AddPage()
		pdf.SetFont(fontName, "", 16)
		pdf.CellFormat(190, 10, "空项目", "", 1, "C", false, 0, "")
		return pdf.OutputFileAndClose(outputPath)
	}

	// 添加目录页（第1页）
	pdf.AddPage()
	pdf.SetFont(fontName, "", 16)
	pdf.CellFormat(190, 10, "目录", "", 1, "C", false, 0, "")
	pdf.SetFont(fontName, "", 12)

	var tocItems []tocItem
	currentPage := 1 // 从1开始，因为目录页是第1页

	// 先生成内容页面
	var processNode func(node *Node, path string, level int) error
	processNode = func(node *Node, path string, level int) error {
		if node == nil {
			return nil
		}

		node.mu.RLock()
		defer node.mu.RUnlock()

		if node.IsDir {
			if path != "/" {
				tocItems = append(tocItems, tocItem{
					title: node.Name,
					page:  currentPage + 1, // 目录项指向下一页
					level: level,
				})
			}

			// 处理子节点
			for _, child := range node.Children {
				childPath := filepath.Join(path, child.Name)
				if err := processNode(child, childPath, level+1); err != nil {
					return err
				}
			}
		} else {
			// 为文件添加新页面
			pdf.AddPage()
			currentPage++

			// 添加到目录
			tocItems = append(tocItems, tocItem{
				title: node.Name,
				page:  currentPage,
				level: level,
			})

			// 添加文件内容
			pdf.SetFont(fontName, "", 14)
			pdf.CellFormat(190, 10, node.Name, "", 1, "L", false, 0, "")

			pdf.SetFont(fontName, "", 12)
			content := string(node.Content)
			pdf.MultiCell(190, 5, content, "", "L", false)
		}

		return nil
	}

	// 处理整个文档树
	if err := processNode(d.root, "/", 0); err != nil {
		return err
	}

	// 返回到目录页填充内容
	pdf.SetPage(1)
	pdf.SetY(30)

	// 添加目录项
	for _, item := range tocItems {
		indent := strings.Repeat("  ", item.level)
		titleWidth := 150 - float64(item.level*10)

		// 添加目录项（不使用链接功能）
		pdf.CellFormat(float64(item.level*10), 5, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(titleWidth, 5, indent+item.title, "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, fmt.Sprintf("%d", item.page), "", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(outputPath)
}
