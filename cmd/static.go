package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var staticCmd = &cobra.Command{
	Use:   "static",
	Short: lang.T("Static files"),
	Long:  lang.T("Static files with specified extensions into a single output file"),
	Run:   runStatics,
}

var printToConsole bool

func init() {
	rootCmd.AddCommand(staticCmd)
	staticCmd.Flags().BoolVar(&printToConsole, "print", false, lang.T("Print results to console"))
}

func runStatics(cmd *cobra.Command, args []string) {
	var targetPath string

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

	// 创建一个map来存储每个扩展名的统计信息
	stats := make(map[string]struct {
		count int
		lines int
	})

	// 遍历所有文件
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			continue
		}

		ext := strings.ToLower(filepath.Ext(path))
		lineCount := len(bytes.Split(content, []byte{'\n'}))

		// 更新统计信息
		fileStats := stats[ext]
		fileStats.count++
		fileStats.lines += lineCount
		stats[ext] = fileStats
	}

	// 准备输出内容
	var result strings.Builder
	result.WriteString("File statistics:\n")
	for ext, stat := range stats {
		result.WriteString(fmt.Sprintf("Extension: %s, Files: %d, Total Lines: %d\n", ext, stat.count, stat.lines))
	}
	result.WriteString(fmt.Sprintf("\nSuccessfully analyzed files from %s\n", targetPath))

	// 如果指定了输出文件，写入文件
	if output != "" {
		err := os.WriteFile(output, []byte(result.String()), 0644)
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
			return
		}
		fmt.Printf("Results written to %s\n", output)
	}

	// 如果需要打印到控制台，则打印
	if printToConsole {
		fmt.Print(result.String())
	} else if output == "" {
		// 如果既没有指定输出文件，也没有要求打印到控制台，至少打印成功信息
		fmt.Printf("Successfully analyzed files from %s\n", targetPath)
	}
}
