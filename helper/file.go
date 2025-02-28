package helper

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// FileInfo 包含文件的基本信息
type FileInfo struct {
	Path string
	Info os.FileInfo
}

// WalkFunc 是一个回调函数类型，用于处理每个文件
type WalkFunc func(fileInfo FileInfo) error

// FilterFunc 是一个筛选函数类型，用于决定是否处理某个文件
type FilterFunc func(fileInfo FileInfo) bool

// WalkDirOptions 包含 WalkDir 的选项
type WalkDirOptions struct {
	DisableGitIgnore bool
	Extensions       []string
	Excludes         []string
}

// WalkDir 递归遍历目录，对每个文件调用回调函数，可选择性地使用筛选函数
func WalkDir(root string, callback WalkFunc, filter FilterFunc, options WalkDirOptions) error {
	gitignoreRules := make(map[string][]string)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileInfo := FileInfo{
			Path: path,
			Info: info,
		}

		// 如果是目录，检查是否有 .gitignore 文件
		if info.IsDir() && !options.DisableGitIgnore {
			rules, err := readGitignore(path)
			if err == nil && rules != nil {
				gitignoreRules[path] = rules
			}
		}

		// 应用筛选函数
		if filter != nil {
			if !filter(fileInfo) {
				return nil // 跳过不符合筛选条件的文件
			}
		}

		// 检查是否应该排除此文件
		if !options.DisableGitIgnore {
			excluded, _ := isPathExcludedByGitignore(path, root, gitignoreRules)
			if excluded {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		return callback(fileInfo)
	})
}

// FilterReadableFiles 使用 WalkDir 来过滤可读的文本文件
func FilterReadableFiles(root string, options WalkDirOptions) ([]string, error) {
	var files []string

	filter := func(fileInfo FileInfo) bool {
		// 如果是目录，允许继续遍历
		if fileInfo.Info.IsDir() {
			return true
		}

		// 检查文件是否为可读文本文件
		if !isReadableTextFile(fileInfo.Path) {
			return false
		}

		// 检查文件扩展名
		if len(options.Extensions) > 0 && !StringSliceContains(options.Extensions, "*") {
			ext := filepath.Ext(fileInfo.Path)
			if len(ext) > 0 {
				ext = ext[1:] // 移除开头的点
			}
			if !StringSliceContains(options.Extensions, ext) {
				return false
			}
		}

		// 检查文件是否应被排除
		return !IsPathExcluded(fileInfo.Path, options.Excludes, root)
	}

	err := WalkDir(root, func(fileInfo FileInfo) error {
		if !fileInfo.Info.IsDir() {
			files = append(files, fileInfo.Path)
		}
		return nil
	}, filter, options)

	return files, err
}

// isReadableTextFile 检查文件是否为可读的文本文件
func isReadableTextFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// 读取文件的前几千字节来判断文件类型
	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}
	buffer = buffer[:n]

	// 使用 http.DetectContentType 检测文件类型
	contentType := http.DetectContentType(buffer)

	// 检查是否为文本类型
	if strings.HasPrefix(contentType, "text/") {
		return true
	}

	// 对于一些常见的文本文件类型进行额外检查
	switch filepath.Ext(path) {
	case ".md", ".txt", ".log", ".json", ".xml", ".csv", ".yml", ".yaml",
		".go", ".py", ".js", ".ts", ".html", ".css", ".java", ".c", ".cpp", ".h",
		".rb", ".php", ".sh", ".bat", ".ps1", ".sql", ".r", ".scala", ".swift":
		return true
	}

	// 如果不是明确的文本类型，尝试检测是否为 UTF-8 编码
	return utf8.Valid(buffer)
}

// readGitignore 读取.gitignore文件并返回其中的规则
func readGitignore(dir string) ([]string, error) {
	gitignorePath := filepath.Join(dir, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var rules []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			rules = append(rules, line)
		}
	}

	return rules, scanner.Err()
}

// IsPathExcluded 检查给定路径是否应被排除
func IsPathExcluded(path string, excludes []string, rootDir string) bool {
	// 检查自定义排除规则
	for _, pattern := range excludes {
		if strings.Contains(path, pattern) {
			return true
		}
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}

	// 检查.gitignore规则
	gitignoreRules, err := readGitignore(rootDir)
	if err != nil {
		// 如果读取.gitignore出错，我们就忽略它，继续处理
		return false
	}

	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		// 如果无法获取相对路径，我们就忽略它，继续处理
		return false
	}

	for _, rule := range gitignoreRules {
		if strings.HasPrefix(rule, "/") {
			// 绝对路径规则
			if matched, _ := filepath.Match(rule[1:], relPath); matched {
				return true
			}
		} else {
			// 相对路径规则
			if matched, _ := filepath.Match(rule, filepath.Base(relPath)); matched {
				return true
			}
			if strings.Contains(relPath, rule) {
				return true
			}
		}
	}

	return false
}

// isPathExcludedByGitignore 检查给定路径是否被 .gitignore 规则排除
func isPathExcludedByGitignore(path, rootDir string, gitignoreRules map[string][]string) (bool, error) {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false, err
	}

	// 从当前目录向上遍历，检查每个目录的 .gitignore 规则
	currentDir := filepath.Dir(path)
	for {
		if rules, ok := gitignoreRules[currentDir]; ok {
			for _, rule := range rules {
				if matchGitignoreRule(relPath, rule) {
					return true, nil
				}
			}
		}

		if currentDir == rootDir {
			break
		}
		currentDir = filepath.Dir(currentDir)
	}

	return false, nil
}

// matchGitignoreRule 检查路径是否匹配 gitignore 规则
func matchGitignoreRule(path, rule string) bool {
	if strings.HasPrefix(rule, "/") {
		// 绝对路径规则
		matched, _ := filepath.Match(rule[1:], path)
		return matched
	} else {
		// 相对路径规则
		matched, _ := filepath.Match(rule, filepath.Base(path))
		if matched {
			return true
		}
		return strings.Contains(path, rule)
	}
}
