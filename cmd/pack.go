package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: lang.T("Pack files"),
	Long:  lang.T("Pack files with specified extensions into a single output file"),
	Run:   runPack,
}

func init() {
	rootCmd.AddCommand(packCmd)
}

func runPack(cmd *cobra.Command, args []string) {
	var targetPath string

	if output == "" {
		fmt.Printf("Output is required")
		return
	}

	if gitURL != "" {
		// 创建临时目录
		tempDir, err := helper.CloneProject(gitURL)
		if err != nil {
			fmt.Printf("Error cloning repository: %v\n", err)
			return
		}
		targetPath = tempDir
	} else {
		targetPath = cmdPath
	}

	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}
	files, ferr := helper.FilterReadableFiles(targetPath, options)

	if ferr != nil {
		fmt.Printf("Error finding files: %v\n", ferr)
		return
	}

	outputExt := strings.ToLower(filepath.Ext(output))
	var content []byte
	var err error

	switch outputExt {
	case ".pdf":
		content, err = packToPDF(files)
	case ".md":
		content, err = packToMarkdown(files)
	case ".xml":
		content, err = packToXML(files)
	default:
		content, err = packToText(files)
	}

	if err != nil {
		fmt.Printf("Error packing files: %v\n", err)
		return
	}

	err = os.WriteFile(output, content, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	fmt.Printf("Successfully packed files into %s\n", output)
}

func packToText(filePaths []string) ([]byte, error) {
	var result strings.Builder
	for _, path := range filePaths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileName := filepath.Base(path)
		result.WriteString(fmt.Sprintf("--- %s ---\n%s\n\n", fileName, string(content)))
	}
	return []byte(result.String()), nil
}

func packToMarkdown(filePaths []string) ([]byte, error) {
	var result strings.Builder
	for _, path := range filePaths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileName := filepath.Base(path)
		ext := filepath.Ext(fileName)
		langHint := strings.TrimPrefix(ext, ".")
		if langHint == "txt" {
			langHint = ""
		}
		result.WriteString(fmt.Sprintf("## %s\n\n```%s\n%s\n```\n\n", fileName, langHint, string(content)))
	}
	return []byte(result.String()), nil
}

func packToXML(filePaths []string) ([]byte, error) {
	var result strings.Builder
	result.WriteString("<files>\n")
	for _, path := range filePaths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileName := filepath.Base(path)
		result.WriteString(fmt.Sprintf("  <file name=\"%s\">\n    <![CDATA[\n%s\n    ]]>\n  </file>\n", fileName, string(content)))
	}
	result.WriteString("</files>\n")
	return []byte(result.String()), nil
}

func packToPDF(files []string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	fontPath, err := helper.FindFont()
	if err != nil {
		return nil, fmt.Errorf("error finding suitable font: %v", err)
	}

	fontName := filepath.Base(fontPath)
	fontName = fontName[:len(fontName)-len(filepath.Ext(fontName))] // Remove extension
	fontName = strings.ReplaceAll(fontName, " ", "")                // Remove spaces from font name

	// Read font file
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("error reading font file: %v", err)
	}

	// Add font
	pdf.AddUTF8FontFromBytes(fontName, "", fontData)

	pdf.SetFont(fontName, "", 12)

	for _, file := range files {
		pdf.AddPage()

		// Add filename as title
		pdf.SetFont(fontName, "", 16) // Changed from "B" to ""
		pdf.Cell(40, 10, filepath.Base(file))
		pdf.Ln(10)

		// Reset font to normal size
		pdf.SetFont(fontName, "", 12)

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}

		// Add file content to PDF
		pdf.MultiCell(0, 10, string(content), "", "", false)
	}

	// Save PDF to memory buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("error generating PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// 修改 Pack 相关函数，确保文件按照固定顺序处理

// 添加一个新的结构体来保存文件信息
type FileEntry struct {
	Name    string
	Content string
}

// 修改 PackToText 函数
func PackToText(files map[string]string) string {
	// 将 map 转换为有序的切片
	var entries []FileEntry
	for name, content := range files {
		entries = append(entries, FileEntry{Name: name, Content: content})
	}

	// 按文件名排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var result strings.Builder
	for _, entry := range entries {
		result.WriteString(fmt.Sprintf("--- %s ---\n%s\n\n", entry.Name, entry.Content))
	}
	return result.String()
}

// 修改 PackToMarkdown 函数
func PackToMarkdown(files map[string]string) string {
	// 将 map 转换为有序的切片
	var entries []FileEntry
	for name, content := range files {
		entries = append(entries, FileEntry{Name: name, Content: content})
	}

	// 按文件名排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var result strings.Builder
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name)
		lang := ""
		if ext != "" {
			lang = ext[1:] // 去掉点号
		}
		result.WriteString(fmt.Sprintf("## %s\n\n```%s\n%s\n```\n\n", entry.Name, lang, entry.Content))
	}
	return result.String()
}

// 修改 PackToXML 函数
func PackToXML(files map[string]string) string {
	// 将 map 转换为有序的切片
	var entries []FileEntry
	for name, content := range files {
		entries = append(entries, FileEntry{Name: name, Content: content})
	}

	// 按文件名排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var result strings.Builder
	result.WriteString("<files>\n")
	for _, entry := range entries {
		result.WriteString(fmt.Sprintf("  <file name=\"%s\">\n    <![CDATA[\n%s\n    ]]>\n  </file>\n", entry.Name, entry.Content))
	}
	result.WriteString("</files>\n")
	return result.String()
}
