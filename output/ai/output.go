package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sjzsdu/wn/llm"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/jung-kurt/gofpdf"
	"github.com/sjzsdu/wn/helper"
)

func Output(output string, messages []llm.Message) error {
	if output == "" || len(messages) == 0 {
		return nil
	}

	outputExt := strings.ToLower(filepath.Ext(output))
	var content []byte
	var err error

	// 格式化对话内容
	var conversations []Conversation
	for _, msg := range messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			conversations = append(conversations, Conversation{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	switch outputExt {
	case ".pdf":
		content, err = toPDF(conversations)
	case ".md":
		content, err = toMarkdown(conversations)
	case ".xml":
		content, err = toXML(conversations)
	default:
		content, err = toText(conversations)
	}

	if err != nil {
		return fmt.Errorf("pack content error: %v", err)
	}

	// 确保输出目录存在
	if dir := filepath.Dir(output); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create output directory error: %v", err)
		}
	}

	// 写入文件
	if err := os.WriteFile(output, content, 0644); err != nil {
		return fmt.Errorf("write file error: %v", err)
	}

	return nil
}

type Conversation struct {
	Role    string
	Content string
}

func toPDF(conversations []Conversation) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// 使用嵌入式字体
	fontPath, err := helper.UseEmbeddedFont("")
	if err != nil {
		return nil, fmt.Errorf("error loading embedded font: %v", err)
	}

	fontName := helper.FONT_NAME

	// 读取字体文件
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("error reading font file: %v", err)
	}

	// 添加字体
	pdf.AddUTF8FontFromBytes(fontName, "", fontData)
	pdf.SetFont(fontName, "", 11)
	
	// 添加标题页
	pdf.AddPage()
	pdf.SetFont(fontName, "", 24)
	pdf.Cell(190, 10, "AI 对话记录")
	pdf.Ln(20)
	
	// 设置正文字体
	pdf.SetFont(fontName, "", 11)
	
	for i, conv := range conversations {
		if i > 0 {
			pdf.AddPage()
		}
		
		// 添加角色标题
		pdf.SetFont(fontName, "", 16)
		pdf.SetTextColor(0, 102, 204) // 蓝色
		pdf.CellFormat(190, 10, strings.Title(conv.Role), "", 1, "L", false, 0, "")
		pdf.Ln(5)
		
		// 重置颜色和字体
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 11)
		
		// 解析 Markdown
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		md := []byte(conv.Content)
		doc := markdown.Parse(md, p)
		
		// 将解析后的内容转换为纯文本
		text := string(markdown.Render(doc, markdown.NewPlainRenderer()))
		
		// 清理文本
		text = cleanText(text)
		
		// 分段显示内容
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				pdf.MultiCell(190, 5, line, "", "L", false)
				pdf.Ln(2)
			}
		}
		
		// 添加分隔线
		if i < len(conversations)-1 {
			pdf.Ln(5)
			pdf.SetDrawColor(200, 200, 200)
			pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
			pdf.Ln(5)
		}
	}
	
	// 将 PDF 转换为字节数组
	return pdf.Output(nil)
}

// 添加辅助函数用于清理文本
func cleanText(text string) string {
	runes := []rune(text)
	result := make([]rune, 0, len(runes))
	for _, r := range runes {
		if r <= 0xFFFF {
			result = append(result, r)
		}
	}
	return string(result)
}

func toMarkdown(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("# AI 对话记录\n\n")

	for _, conv := range conversations {
		sb.WriteString(fmt.Sprintf("## %s\n\n",
			strings.Title(conv.Role)))
		sb.WriteString(conv.Content)
		sb.WriteString("\n\n---\n\n")
	}

	return []byte(sb.String()), nil
}

func toXML(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<conversations>\n")

	for _, conv := range conversations {
		sb.WriteString(fmt.Sprintf("  <message role=\"%s\">\n",
			conv.Role))
		sb.WriteString(fmt.Sprintf("    <content><![CDATA[%s]]></content>\n", conv.Content))
		sb.WriteString("  </message>\n")
	}

	sb.WriteString("</conversations>")
	return []byte(sb.String()), nil
}

func toText(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder

	for _, conv := range conversations {
		sb.WriteString(fmt.Sprintf("%s:\n",
			strings.Title(conv.Role)))
		sb.WriteString(conv.Content)
		sb.WriteString("\n\n")
	}

	return []byte(sb.String()), nil
}
