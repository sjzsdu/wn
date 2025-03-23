package project

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sort"
	"strings"
	"sync"
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
func (node *Node) GetChildrenResponses() string {

	if !node.IsDir || len(node.Children) == 0 {
		return ""
	}

	var responses []string
	for name, child := range node.Children {
		if child.LLMResponse != "" {
			responses = append(responses, name+":\n"+child.LLMResponse)
		}
	}

	return strings.Join(responses, "\n\n")
}
