package project

import (
	"os"
	"sync"
)

type Node struct {
	Name     string
	IsDir    bool
	Info     os.FileInfo
	Content  []byte
	Children map[string]*Node  // 改回 map 类型
	Parent   *Node
	mu       sync.RWMutex
}

// Project 表示整个文档树
type Project struct {
	root *Node
	mu   sync.RWMutex
}
