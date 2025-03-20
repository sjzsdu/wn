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

const (
	pageWidth      = 190.0  // 页面宽度
	lineHeight     = 5.0    // 每行高度
	titleHeight    = 10.0   // 标题高度
	pageHeight     = 297.0  // A4页面高度
	marginTop      = 30.0   // 上边距
	marginBottom   = 20.0   // 下边距
)

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

	// 先添加一个空白目录页
	pdf.AddPage()
	tocPage := pdf.PageNo() // 记录目录页页码

	var tocItems []tocItem

	// 修改processNode函数，记录实际页码
	var processNode func(node *Node, path string, level int) error
	processNode = func(node *Node, path string, level int) error {
		if node == nil {
			return nil
		}

		node.mu.RLock()
		defer node.mu.RUnlock()

		if node.IsDir {
			if path != "/" {
				// 记录目录项，页码与后续文件页码一致
				tocItems = append(tocItems, tocItem{
					title: node.Name,
					page:  pdf.PageNo() + 1, // 确保页码与后续文件页码一致
					level: level,
				})
			}

			for _, child := range node.Children {
				childPath := filepath.Join(path, child.Name)
				if err := processNode(child, childPath, level+1); err != nil {
					return err
				}
			}
		} else {
			// 记录当前页码作为文件起始页
			startPage := pdf.PageNo() + 1
			
			// 添加文件内容
			pdf.AddPage()
			pdf.SetFont(fontName, "", 14)
			pdf.CellFormat(190, 10, node.Name, "", 1, "L", false, 0, "")

			pdf.SetFont(fontName, "", 12)
			content := string(node.Content)
			pdf.MultiCell(190, 5, content, "", "L", false)

			// 更新目录项
			tocItems = append(tocItems, tocItem{
				title: node.Name,
				page:  startPage,
				level: level,
			})
		}

		return nil
	}

	// 处理整个文档树
	if err := processNode(d.root, "/", 0); err != nil {
		return err
	}

	// 返回到目录页填充内容
	pdf.SetPage(tocPage)
	pdf.SetY(30)
	pdf.SetFont(fontName, "", 16)
	pdf.CellFormat(190, 10, "目录", "", 1, "C", false, 0, "")
	pdf.SetFont(fontName, "", 12)

	// 添加目录项
	for _, item := range tocItems {
		indent := strings.Repeat(" ", item.level) // 减少缩进空格数量
		titleWidth := 150 - float64(item.level*5) // 调整标题宽度

		// 保存当前位置用于添加链接
		x, y := pdf.GetXY()

		// 添加目录项（使用链接功能）
		pdf.CellFormat(float64(item.level*5), 5, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(titleWidth, 5, indent+item.title, "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, fmt.Sprintf("%d", item.page), "", 1, "R", false, 0, "")

		// 使用AddLink和Link方法创建内部页面链接
		link := pdf.AddLink()
		pdf.SetLink(link, 0, item.page)
		pdf.Link(x, y, 190, 5, link)
	}

	return pdf.OutputFileAndClose(outputPath)
}
