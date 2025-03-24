package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sjzsdu/wn/data"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var (
	cacheFile string
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: lang.T("Cache manager"),
	Long:  lang.T("Cache manager for project"),
	Run:   runCache,
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	// 添加 file 标志
	cacheCmd.PersistentFlags().StringVarP(&cacheFile, "file", "f", "", "指定缓存文件路径")
}

func runCache(cmd *cobra.Command, args []string) {
	var rawPath string

	// 确定原始路径
	switch {
	case len(args) > 1:
		fmt.Println("参数过多，只需要指定一个路径")
		return
	case len(args) == 1:
		rawPath = args[0]
	case cacheFile != "":
		rawPath = cacheFile
	}
	// 获取绝对路径
	path, err := helper.GetTargetPath(rawPath, "")
	if err != nil {
		fmt.Printf("获取目标路径失败: %v\n", err)
		return
	}

	cache := data.GetDefaultCacheManager()
	defer cache.Close()

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("读取路径失败: %v\n", err)
		return
	}

	if fileInfo.IsDir() {
		// 处理目录
		records, err := cache.GetAllRecords()
		if err != nil {
			fmt.Printf("获取缓存记录失败: %v\n", err)
			return
		}

		fmt.Printf("目录: %s\n", path)
		fmt.Println(strings.Repeat("-", 50))
		for _, record := range records {
			if strings.HasPrefix(record.Path, path) {
				fmt.Printf("文件: %s\n", record.Path)
				fmt.Printf("内容:\n%s\n", record.Content)
				fmt.Println(strings.Repeat("-", 50))
			}
		}
	} else {
		// 处理单个文件
		record, err := cache.FindByPath(path)
		if err != nil {
			fmt.Printf("查找缓存记录失败: %v\n", err)
			return
		}
		if record == nil {
			fmt.Printf("未找到文件 %s 的缓存记录\n", path)
			return
		}

		fmt.Printf("文件: %s\n", record.Path)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("内容:\n%s\n", record.Content)
	}
}
