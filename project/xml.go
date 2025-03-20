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

// ExportToXML 将项目导出为 XML 文件
func (d *Project) ExportToXML(outputPath string) error {
	if d.root == nil {
		return fmt.Errorf("project is empty")
	}

	doc := XMLDocument{}

	// 转换节点的函数
	var convertNode func(node *Node) XMLNode
	convertNode = func(node *Node) XMLNode {
		node.mu.RLock()
		defer node.mu.RUnlock()

		xmlNode := XMLNode{
			Name: node.Name,
		}

		if node.IsDir {
			xmlNode.Type = "directory"
			xmlNode.Children = make([]XMLNode, 0, len(node.Children))
			for _, child := range node.Children {
				xmlNode.Children = append(xmlNode.Children, convertNode(child))
			}
		} else {
			xmlNode.Type = "file"
			content := string(node.Content)
			xmlNode.Content = &content
		}

		return xmlNode
	}

	// 转换整个文档树
	doc.Nodes = []XMLNode{convertNode(d.root)}

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

	// 编码并写入文件
	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf("failed to encode XML: %v", err)
	}

	return nil
}
