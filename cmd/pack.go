package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/sjzsdu/wn/helper"
	"github.com/spf13/cobra"
)

var (
	exts     []string
	output   string
	excludes []string
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: L("Pack files"),
	Long:  "Pack files with specified extensions into a single output file",
	Run:   runPack,
}

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().StringSliceVarP(&exts, "exts", "e", []string{"py", "ts", "js", "html", "less"}, "File extensions to include")
	packCmd.Flags().StringVarP(&output, "output", "o", "output.pdf", "Output file name")
	packCmd.Flags().StringSliceVarP(&excludes, "excludes", "x", []string{}, "Glob patterns to exclude")
}

func runPack(cmd *cobra.Command, args []string) {
	files, ferr := helper.FilterFiles(cmdPath, exts, excludes)
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
	pdf.SetFont("Arial", "", 12)

	for _, file := range files {
		pdf.AddPage()

		// 添加文件名作为标题
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, filepath.Base(file))
		pdf.Ln(10)

		// 重置字体为正常大小
		pdf.SetFont("Arial", "", 12)

		// 读取文件内容
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}

		// 将文件内容添加到 PDF
		pdf.MultiCell(0, 10, string(content), "", "", false)
	}

	// 保存 PDF 到内存缓冲区
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("error generating PDF: %v", err)
	}

	return buf.Bytes(), nil
}
