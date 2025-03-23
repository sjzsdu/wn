package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/sjzsdu/wn/share"
)

type Node struct {
	Name        string
	IsDir       bool
	Info        os.FileInfo
	Content     []byte
	LLMResponse string
	Children    map[string]*Node // 改回 map 类型
	Parent      *Node
	mu          sync.RWMutex
}

// Project 表示整个文档树
type Project struct {
	root     *Node
	rootPath string
	mu       sync.RWMutex
}

func (node *Node) SetLLMResponse(response string) *Node {
	node.LLMResponse = response
	return node
}

// CalculateHash 计算节点的哈希值
func (node *Node) CalculateHash() (string, error) {
	if node.IsDir {
		return node.calculateDirHash()
	}
	return node.calculateFileHash()
}

// calculateFileHash 计算文件内容的哈希值
func (node *Node) calculateFileHash() (string, error) {
	if node.Content == nil {
		return "", nil
	}
	hash := sha256.Sum256(node.Content)
	return hex.EncodeToString(hash[:]), nil
}

// calculateDirHash 计算目录的哈希值
func (node *Node) calculateDirHash() (string, error) {
	var hashes []string
	// 先对 Children 按名称排序
	sortedChildren := make([]*Node, len(node.Children))
	i := 0
	for _, child := range node.Children {
		sortedChildren[i] = child
		i++
	}
	sort.Slice(sortedChildren, func(i, j int) bool {
		return sortedChildren[i].Name < sortedChildren[j].Name
	})

	// 使用排序后的切片计算哈希
	for _, child := range sortedChildren {
		hash, err := child.CalculateHash()
		if err != nil {
			return "", err
		}
		hashes = append(hashes, hash)
	}

	combined := []byte(strings.Join(hashes, ""))
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:]), nil
}

// GetChildrenResponses 获取直接子节点的 LLMResponse 内容
func (node *Node) GetChildrenResponses() (string, error) {
	if !node.IsDir || len(node.Children) == 0 {
		return "", nil
	}

	var responses []string
	hasValidContent := false

	// 对子节点进行排序以保证顺序一致
	children := make([]*Node, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].Name < children[j].Name
	})

	for _, child := range children {
		// 跳过非程序文件的响应
		if child.LLMResponse == share.NOT_PROGRAM_TIP {
			continue
		}

		if child.LLMResponse != "" {
			if !child.IsDir && len(child.Content) == 0 {
				return "", fmt.Errorf("empty file content: %s", child.Name)
			}
			responses = append(responses, child.Name+":\n"+child.LLMResponse)
			hasValidContent = true
		}
	}

	if !hasValidContent || len(responses) == 0 {
		return "", fmt.Errorf("no valid content found in directory")
	}

	return strings.Join(responses, "\n\n"), nil
}
