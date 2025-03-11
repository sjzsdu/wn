package helper

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func WalkDir(root string, callback WalkFunc, filter FilterFunc, options WalkDirOptions) error {
	gitignoreRules := make(map[string][]string)
	root = filepath.Clean(root)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileInfo := FileInfo{
			Path: path,
			Info: info,
		}

		// 处理 .gitignore 规则
		if !options.DisableGitIgnore {
			if info.IsDir() {
				rules, err := readGitignore(path)
				if err == nil && rules != nil {
					gitignoreRules[path] = rules
				}
			}

			excluded, excludeErr := isPathExcludedByGitignore(path, root, gitignoreRules)
			if excludeErr != nil {
				return excludeErr
			}
			if excluded {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// 检查文件扩展名
		if len(options.Extensions) > 0 {
			ext := strings.ToLower(filepath.Ext(fileInfo.Path))
			if len(ext) > 0 {
				ext = ext[1:] // 移除开头的点
			}
			if !StringSliceContains(options.Extensions, ext) && !StringSliceContains(options.Extensions, "*") {
				return nil
			}
		}

		// 应用自定义筛选函数
		if filter != nil && !filter(fileInfo) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return callback(fileInfo)
	})
}

// FilterReadableFiles 使用 WalkDir 来过滤可读的文本文件
func FilterReadableFiles(root string, options WalkDirOptions) ([]string, error) {
	var files []string
	var count int

	startTime := time.Now()
	filter := func(fileInfo FileInfo) bool {
		count++
		if count%1000 == 0 {
			log.Printf("Processed %d files", count)
		}
		// 如果是目录，允许继续遍历，但排除 .git 目录
		if fileInfo.Info.IsDir() {
			return fileInfo.Info.Name() != ".git"
		}

		// 检查文件是否为可读文本文件
		if !isReadableTextFile(fileInfo.Path) {
			return false
		}

		// 检查文件扩展名
		if len(options.Extensions) > 0 {
			ext := strings.ToLower(filepath.Ext(fileInfo.Path))
			if len(ext) > 0 {
				ext = ext[1:] // 移除开头的点
			}
			if !StringSliceContains(options.Extensions, ext) && !StringSliceContains(options.Extensions, "*") {
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
	log.Printf("FilterReadableFiles completed: processed %d files, filtered %d readable files. Elapsed time: %v", count, len(files), time.Since(startTime))
	return files, err
}

var textExtensions = map[string]bool{
	".md": true, ".txt": true, ".log": true, ".json": true, ".xml": true, ".csv": true,
	".yml": true, ".yaml": true, ".go": true, ".py": true, ".js": true, ".ts": true,
	".html": true, ".css": true, ".java": true, ".c": true, ".cpp": true, ".h": true,
	".rb": true, ".php": true, ".sh": true, ".bat": true, ".ps1": true, ".sql": true,
	".r": true, ".scala": true, ".swift": true, ".mdx": true,
}

func isReadableTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if _, ok := textExtensions[ext]; ok {
		return true
	}

	// 只对未知扩展名的文件进行内容检查
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}
	buffer = buffer[:n]

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

func isPathExcludedByGitignore(path, rootDir string, gitignoreRules map[string][]string) (bool, error) {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false, err
	}

	// Check rules from all parent directories
	for dir := path; dir != rootDir && dir != "."; dir = filepath.Dir(dir) {
		if rules, ok := gitignoreRules[dir]; ok {
			for _, rule := range rules {
				if matchGitignoreRule(relPath, rule) {
					return true, nil
				}
			}
		}
	}

	// Check root directory rules last
	if rules, ok := gitignoreRules[rootDir]; ok {
		for _, rule := range rules {
			if matchGitignoreRule(relPath, rule) {
				return true, nil
			}
		}
	}

	return false, nil
}

func matchGitignoreRule(path, rule string) bool {
	// Skip empty rules
	if rule == "" {
		return false
	}

	// Handle directory-specific rules
	if strings.HasSuffix(rule, "/") {
		rule = rule[:len(rule)-1]
	}

	if strings.HasPrefix(rule, "/") {
		// Absolute path rule
		matched, _ := filepath.Match(rule[1:], path)
		return matched
	} else {
		// Relative path rule
		base := filepath.Base(path)
		matched, _ := filepath.Match(rule, base)
		if matched {
			return true
		}

		// Check if rule matches any path component
		components := strings.Split(path, string(filepath.Separator))
		for _, comp := range components {
			if matched, _ := filepath.Match(rule, comp); matched {
				return true
			}
		}
		return false
	}
}
