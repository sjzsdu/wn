package project

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Export 根据输出文件的扩展名选择相应的导出方法
func (d *Project) Export(output string) error {
	outputExt := strings.ToLower(filepath.Ext(output))

	switch outputExt {
	case ".pdf":
		return d.ExportToPDF(output)
	case ".md":
		return d.ExportToMarkdown(output)
	case ".xml":
		return d.ExportToXML(output)
	default:
		return fmt.Errorf("unsupported output format: %s (only support pdf, md, xml)", outputExt)
	}
}
