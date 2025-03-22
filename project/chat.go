package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ChatVisitor 实现 NodeVisitor 接口，用于管理 AI 聊天数据
type ChatVisitor struct {
	cacheDir   string
	chatData   map[string]string // path -> AI response
	cacheMutex sync.RWMutex
}

func NewChatVisitor() (*ChatVisitor, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(homeDir, ".wn", "chat_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &ChatVisitor{
		cacheDir: cacheDir,
		chatData: make(map[string]string),
	}, nil
}

func (v *ChatVisitor) VisitDirectory(node *Node, path string, level int) error {
	// 计算目录下所有文件内容的组合哈希
	hash, err := v.calculateDirHash(node)
	if err != nil {
		return err
	}

	return v.loadOrCreateCache(path, hash)
}

func (v *ChatVisitor) VisitFile(node *Node, path string, level int) error {
	// 计算文件内容的哈希
	hash, err := v.calculateFileHash(node)
	if err != nil {
		return err
	}

	return v.loadOrCreateCache(path, hash)
}

func (v *ChatVisitor) calculateFileHash(node *Node) (string, error) {
	if node.Content == nil {
		return "", nil
	}
	hash := sha256.Sum256(node.Content) // 直接使用 Content，不需要间接引用
	return hex.EncodeToString(hash[:]), nil
}

func (v *ChatVisitor) calculateDirHash(node *Node) (string, error) {
	var hashes []string
	for _, child := range node.Children {
		if child.IsDir {
			hash, err := v.calculateDirHash(child)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, hash)
		} else {
			hash, err := v.calculateFileHash(child)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, hash)
		}
	}

	combined := []byte(strings.Join(hashes, ""))
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:]), nil
}

func (v *ChatVisitor) loadOrCreateCache(path, hash string) error {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()

	cachePath := filepath.Join(v.cacheDir, hash+".json")

	// 尝试从缓存文件加载
	data, err := os.ReadFile(cachePath)
	if err == nil {
		v.chatData[path] = string(data)
		return nil
	}

	// 如果缓存不存在，初始化为空字符串
	v.chatData[path] = ""
	return nil
}

// SetResponse 设置指定路径的 AI 响应
func (v *ChatVisitor) SetResponse(path, response string) error {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()

	hash := v.getHashForPath(path)
	if hash == "" {
		return fmt.Errorf("path not visited: %s", path)
	}

	v.chatData[path] = response

	// 保存到缓存文件
	cachePath := filepath.Join(v.cacheDir, hash+".json")
	return os.WriteFile(cachePath, []byte(response), 0644)
}

// GetResponse 获取指定路径的 AI 响应
func (v *ChatVisitor) GetResponse(path string) (string, error) {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()

	if response, exists := v.chatData[path]; exists {
		return response, nil
	}
	return "", fmt.Errorf("no response found for path: %s", path)
}

// getHashForPath 获取指定路径的哈希值
func (v *ChatVisitor) getHashForPath(path string) string {
	files, err := os.ReadDir(v.cacheDir)
	if err != nil {
		return ""
	}

	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()

	for _, file := range files {
		if !file.IsDir() {
			hash := strings.TrimSuffix(file.Name(), ".json")
			data, err := os.ReadFile(filepath.Join(v.cacheDir, file.Name()))
			if err != nil {
				continue
			}
			if string(data) == v.chatData[path] {
				return hash
			}
		}
	}
	return ""
}
