package ai

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"bytes"
	"encoding/base64"
	"net/http"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/jung-kurt/gofpdf"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	case ".html":
		content, err = toHTML(conversations)
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
	pdf.SetAutoPageBreak(true, 10)

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

	// 添加字体 - 同时添加普通样式和粗体样式
	pdf.AddUTF8FontFromBytes(fontName, "", fontData)
	pdf.AddUTF8FontFromBytes(fontName, "B", fontData) // 添加粗体样式
	pdf.SetFont(fontName, "", 12)

	// 添加标题页
	pdf.AddPage()
	pdf.SetFont(fontName, "", 24) // 改用普通样式
	pdf.Cell(190, 10, "AI 对话记录")
	pdf.Ln(20)

	// 添加处理 Mermaid 的函数
	processMermaid := func(mermaidCode string) (string, error) {
		encoded := base64.URLEncoding.EncodeToString([]byte(mermaidCode))
		apiURL := "https://kroki.io/mermaid/svg/" + encoded

		resp, err := http.Get(apiURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to generate diagram: %s", resp.Status)
		}

		// 保存为临时文件
		tempFile := filepath.Join(os.TempDir(), "mermaid_"+encoded[:8]+".svg")
		f, err := os.Create(tempFile)
		if err != nil {
			return "", err
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", err
		}

		return tempFile, nil
	}

	for i, conv := range conversations {
		if i > 0 {
			pdf.AddPage()
		}

		// 添加角色标题
		pdf.SetFont(fontName, "B", 16)
		pdf.SetTextColor(0, 102, 204)
		pdf.CellFormat(190, 10, strings.Title(conv.Role), "", 1, "L", false, 0, "")
		pdf.Ln(5)

		// 重置颜色和字体
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 12)

		// 解析 Markdown
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		md := []byte(conv.Content)
		doc := markdown.Parse(md, p)

		// 配置 HTML 渲染器
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		// 渲染为 HTML
		htmlContent := markdown.Render(doc, renderer)

		// 清理 HTML 标签并处理特殊字符
		text := helper.StripHTMLTags(string(htmlContent))
		text = cleanText(text)

		var inMermaid bool
		var mermaidCode strings.Builder

		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			// 处理 Mermaid 代码块
			if strings.HasPrefix(line, "```mermaid") {
				inMermaid = true
				continue
			} else if inMermaid && strings.HasPrefix(line, "```") {
				inMermaid = false
				// 生成图表
				if svgFile, err := processMermaid(mermaidCode.String()); err == nil {
					// 在 PDF 中插入图表
					pdf.Image(svgFile, 10, pdf.GetY(), 190, 0, false, "", 0, "")
					pdf.Ln(10)
					// 清理临时文件
					os.Remove(svgFile)
				}
				mermaidCode.Reset()
				continue
			}

			if inMermaid {
				mermaidCode.WriteString(line + "\n")
				continue
			}

			// 处理普通文本
			if line != "" {
				pdf.MultiCell(190, 5, line, "", "L", false)
				pdf.Ln(2)
			} else {
				pdf.Ln(5)
			}
		}

		if i < len(conversations)-1 {
			pdf.Ln(5)
			pdf.SetDrawColor(200, 200, 200)
			pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
			pdf.Ln(5)
		}
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("生成 PDF 失败: %v", err)
	}
	return buf.Bytes(), nil
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

func toHTML(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder

	// HTML 头部
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI 对话记录</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/github-markdown-css@5.2.0/github-markdown.min.css">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        body {
            box-sizing: border-box;
            min-width: 200px;
            max-width: 980px;
            margin: 0 auto;
            padding: 45px;
            background-color: #f6f8fa;
        }
        .conversation {
            background-color: white;
            border-radius: 6px;
            padding: 20px;
            margin: 20px 0;
            box-shadow: 0 1px 3px rgba(0,0,0,0.12);
        }
        .role {
            color: #0366d6;
            font-size: 1.2em;
            font-weight: bold;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 1px solid #eaecef;
        }
        .content {
            margin-top: 15px;
        }
        pre {
            background-color: #f6f8fa;
            border-radius: 6px;
            padding: 16px;
        }
    </style>
</head>
<body>
    <h1>AI 对话记录</h1>`)

	// 对话内容
	for _, conv := range conversations {
		// 使用 base64 编码来安全地传递 markdown 内容
		encodedContent := base64.StdEncoding.EncodeToString([]byte(conv.Content))
		sb.WriteString(fmt.Sprintf(`
    <div class="conversation">
        <div class="role">%s</div>
        <div class="content markdown-body" data-markdown="%s"></div>
    </div>`,
			cases.Title(language.English).String(conv.Role),
			encodedContent))
	}

	// HTML 尾部和 JavaScript
	sb.WriteString(`
    <script>
        // 初始化 Mermaid
        mermaid.initialize({
            startOnLoad: true,  // 改回 true
            theme: 'default',
            securityLevel: 'loose'
        });

        // 处理 Markdown 内容
        document.querySelectorAll('[data-markdown]').forEach(async (element) => {
            const encodedMarkdown = element.getAttribute('data-markdown');
            try {
                const markdown = decodeURIComponent(escape(atob(encodedMarkdown)));
                element.removeAttribute('data-markdown');
                
                // 渲染 Markdown
                const html = marked.parse(markdown);
                element.innerHTML = html;
                
                // 强制重新渲染 Mermaid
                element.querySelectorAll('code.language-mermaid').forEach(code => {
                    const pre = code.parentElement;
                    const div = document.createElement('div');
                    div.className = 'mermaid';
                    div.textContent = code.textContent;
                    pre.parentElement.replaceChild(div, pre);
                });

                // 重新初始化 Mermaid
                mermaid.init();
            } catch (e) {
                console.error('Error parsing markdown:', e);
                element.innerHTML = '<pre>Error rendering content</pre>';
            }
        });
    </script>`)

	return []byte(sb.String()), nil
}
