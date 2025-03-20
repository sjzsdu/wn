package project

import (
	"encoding/xml"
	"fmt"
	"os"
)

// XMLNode 表示 XML 文档中的节点
type XMLNode struct {
	XMLName  xml.Name  `xml:"node"`
	Name     string    `xml:"name,attr"`
	Type     string    `xml:"type,attr"`
	Content  *string   `xml:"content,omitempty"`
	Children []XMLNode `xml:"nodes>node,omitempty"`
}

// XMLDocument 表示整个 XML 文档
type XMLDocument struct {
	XMLName xml.Name `xml:"document"`
	Nodes   []XMLNode `xml:"nodes>node"`
}

// XMLExporter XML导出器
type XMLExporter struct {
	*BaseExporter
	doc XMLDocument
}

// NewXMLExporter 创建新的XML导出器
func NewXMLExporter(p *Project) *XMLExporter {
	return &XMLExporter{
		BaseExporter: NewBaseExporter(p),
		doc:         XMLDocument{},
	}
}

func (x *XMLExporter) ProcessDirectory(node *Node, path string, level int) error {
	xmlNode := XMLNode{
		Name: node.Name,
		Type: "directory",
		Children: make([]XMLNode, 0, len(node.Children)),
	}
	x.doc.Nodes = append(x.doc.Nodes, xmlNode)
	return nil
}

func (x *XMLExporter) ProcessFile(node *Node, path string, level int) error {
	content := string(node.Content)
	xmlNode := XMLNode{
		Name:    node.Name,
		Type:    "file",
		Content: &content,
	}
	x.doc.Nodes = append(x.doc.Nodes, xmlNode)
	return nil
}

func (x *XMLExporter) Export(outputPath string) error {
	if x.project.root == nil {
		return fmt.Errorf("project is empty")
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// 写入 XML 头
	file.WriteString(xml.Header)

	// 创建编码器
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	// 遍历节点并构建文档
	if err := x.TraverseNodes(x.project.root, "/", 0, x); err != nil {
		return err
	}

	// 编码并写入文件
	if err := encoder.Encode(x.doc); err != nil {
		return fmt.Errorf("failed to encode XML: %v", err)
	}

	return nil
}

// ExportToXML 将项目导出为 XML 文件
func (d *Project) ExportToXML(outputPath string) error {
	exporter := NewXMLExporter(d)
	return exporter.Export(outputPath)
}
