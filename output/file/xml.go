package file

import (
	"encoding/xml"
	"os"

	"github.com/sjzsdu/wn/project"
)

type XMLNode struct {
	XMLName  xml.Name  `xml:"node"`
	Name     string    `xml:"name,attr"`
	Type     string    `xml:"type,attr"`
	Content  *string   `xml:"content,omitempty"`
	Children []XMLNode `xml:"nodes>node,omitempty"`
}

type XMLDocument struct {
	XMLName xml.Name  `xml:"document"`
	Nodes   []XMLNode `xml:"nodes>node"`
}

type XMLCollector struct {
	doc XMLDocument
}

func NewXMLCollector() *XMLCollector {
	return &XMLCollector{
		doc: XMLDocument{},
	}
}

func (x *XMLCollector) AddTitle(title string, level int) error {
	node := XMLNode{
		Name: title,
		Type: "directory",
	}
	x.doc.Nodes = append(x.doc.Nodes, node)
	return nil
}

func (x *XMLCollector) AddContent(content string) error {
	node := XMLNode{
		Type:    "file",
		Content: &content,
	}
	x.doc.Nodes = append(x.doc.Nodes, node)
	return nil
}

// AddTOCItem 实现可选的目录项添加
func (x *XMLCollector) AddTOCItem(title string, level int) error {
	// XML 格式不需要特殊的目录处理，直接返回 nil
	return nil
}

func (x *XMLCollector) Render(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(x.doc)
}

type XMLExporter struct {
	*BaseExporter
}

func NewXMLExporter(p *project.Project) *XMLExporter {
	return &XMLExporter{
		BaseExporter: NewBaseExporter(p, NewXMLCollector()),
	}
}