package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareTestFiles(t *testing.T) ([]string, string) {
	tempDir := t.TempDir()
	files := map[string]string{
		"file1.txt": "这是第一个文件的内容",
		"file2.md":  "# 这是第二个文件的内容",
		"file3.xml": "<root>这是第三个文件的内容</root>",
	}

	var filePaths []string
	for name, content := range files {
		path := filepath.Join(tempDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)
		filePaths = append(filePaths, path)
	}

	return filePaths, tempDir
}

func TestPackFunctions(t *testing.T) {
	filePaths, _ := prepareTestFiles(t)

	t.Run("PackToText", func(t *testing.T) {
		expected := "--- file1.txt ---\n这是第一个文件的内容\n\n--- file2.md ---\n# 这是第二个文件的内容\n\n--- file3.xml ---\n<root>这是第三个文件的内容</root>\n\n"
		output, err := packToText(filePaths)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(output))
	})

	t.Run("PackToMarkdown", func(t *testing.T) {
		expected := "## file1.txt\n\n```\n这是第一个文件的内容\n```\n\n## file2.md\n\n```md\n# 这是第二个文件的内容\n```\n\n## file3.xml\n\n```xml\n<root>这是第三个文件的内容</root>\n```\n\n"
		output, err := packToMarkdown(filePaths)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(output))
	})

	t.Run("PackToXML", func(t *testing.T) {
		expected := "<files>\n  <file name=\"file1.txt\">\n    <![CDATA[\n这是第一个文件的内容\n    ]]>\n  </file>\n  <file name=\"file2.md\">\n    <![CDATA[\n# 这是第二个文件的内容\n    ]]>\n  </file>\n  <file name=\"file3.xml\">\n    <![CDATA[\n<root>这是第三个文件的内容</root>\n    ]]>\n  </file>\n</files>\n"
		output, err := packToXML(filePaths)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(output))
	})
}
