package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/sjzsdu/wn/helper"
)

type tocItem struct {
	title string
	page  int
	level int
}

const (
	pageWidth    = 190.0 // 页面宽度
	lineHeight   = 5.0   // 每行高度
	titleHeight  = 10.0  // 标题高度
	pageHeight   = 297.0 // A4页面高度
	marginTop    = 30.0  // 上边距
	marginBottom = 20.0  // 下边距
)

type PDFExporter struct {
	*BaseExporter
	pdf      *gofpdf.Fpdf
	tocItems []tocItem
	fontName string
	tocPage  int
}

func NewPDFExporter(p *Project) (*PDFExporter, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// 设置字体
	fontPath, err := helper.FindFont()
	if err != nil {
		return nil, fmt.Errorf("error finding suitable font: %v", err)
	}

	fontName := filepath.Base(fontPath)
	fontName = fontName[:len(fontName)-len(filepath.Ext(fontName))]
	fontName = strings.ReplaceAll(fontName, " ", "")

	// 读取字体文件
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("error reading font file: %v", err)
	}

	// 添加字体
	pdf.AddUTF8FontFromBytes(fontName, "", fontData)
	pdf.SetFont(fontName, "", 12)

	return &PDFExporter{
		BaseExporter: NewBaseExporter(p),
		pdf:         pdf,
		tocItems:    make([]tocItem, 0),
		fontName:    fontName,
	}, nil
}

func (e *PDFExporter) ProcessDirectory(node *Node, path string, level int) error {
	if path != "/" {
		e.tocItems = append(e.tocItems, tocItem{
			title: node.Name,
			page:  e.pdf.PageNo() + 1,
			level: level,
		})
	}
	return nil
}

func (e *PDFExporter) ProcessFile(node *Node, path string, level int) error {
	startPage := e.pdf.PageNo() + 1
	
	e.pdf.AddPage()
	e.pdf.SetFont(e.fontName, "", 14)
	e.pdf.CellFormat(190, 10, node.Name, "", 1, "L", false, 0, "")

	e.pdf.SetFont(e.fontName, "", 12)
	content := string(node.Content)
	e.pdf.MultiCell(190, 5, content, "", "L", false)

	e.tocItems = append(e.tocItems, tocItem{
		title: node.Name,
		page:  startPage,
		level: level,
	})
	
	return nil
}

func (e *PDFExporter) writeTOC() error {
	e.pdf.SetPage(e.tocPage)
	e.pdf.SetY(30)
	e.pdf.SetFont(e.fontName, "", 16)
	e.pdf.CellFormat(190, 10, "目录", "", 1, "C", false, 0, "")
	e.pdf.SetFont(e.fontName, "", 12)

	for _, item := range e.tocItems {
		indent := strings.Repeat(" ", item.level)
		titleWidth := 150 - float64(item.level*5)

		x, y := e.pdf.GetXY()

		e.pdf.CellFormat(float64(item.level*5), 5, "", "", 0, "L", false, 0, "")
		e.pdf.CellFormat(titleWidth, 5, indent+item.title, "", 0, "L", false, 0, "")
		e.pdf.CellFormat(40, 5, fmt.Sprintf("%d", item.page), "", 1, "R", false, 0, "")

		link := e.pdf.AddLink()
		e.pdf.SetLink(link, 0, item.page)
		e.pdf.Link(x, y, 190, 5, link)
	}

	return nil
}

func (e *PDFExporter) Export(outputPath string) error {
	if e.project.root == nil || len(e.project.root.Children) == 0 {
		e.pdf.AddPage()
		e.pdf.SetFont(e.fontName, "", 16)
		e.pdf.CellFormat(190, 10, "空项目", "", 1, "C", false, 0, "")
		return e.pdf.OutputFileAndClose(outputPath)
	}

	// 添加目录页
	e.pdf.AddPage()
	e.tocPage = e.pdf.PageNo()

	// 遍历节点
	if err := e.TraverseNodes(e.project.root, "/", 0, e); err != nil {
		return err
	}

	// 写入目录
	if err := e.writeTOC(); err != nil {
		return err
	}

	return e.pdf.OutputFileAndClose(outputPath)
}

func (d *Project) ExportToPDF(outputPath string) error {
	exporter, err := NewPDFExporter(d)
	if err != nil {
		return err
	}
	return exporter.Export(outputPath)
}
