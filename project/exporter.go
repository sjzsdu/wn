package project

import "path/filepath"

// Exporter 定义了项目导出器的接口
type Exporter interface {
	// Export 执行导出操作
	Export(outputPath string) error
	// ProcessDirectory 处理目录节点
	ProcessDirectory(node *Node, path string, level int) error
	// ProcessFile 处理文件节点
	ProcessFile(node *Node, path string, level int) error
}

// BaseExporter 提供了基本的导出功能
type BaseExporter struct {
	project *Project
}

// NewBaseExporter 创建一个基本导出器
func NewBaseExporter(p *Project) *BaseExporter {
	return &BaseExporter{project: p}
}

// TraverseNodes 遍历节点的通用方法
func (b *BaseExporter) TraverseNodes(node *Node, path string, level int, e Exporter) error {
	if node == nil {
		return nil
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.IsDir {
		if err := e.ProcessDirectory(node, path, level); err != nil {
			return err
		}

		for _, child := range node.Children {
			childPath := filepath.Join(path, child.Name)
			if err := b.TraverseNodes(child, childPath, level+1, e); err != nil {
				return err
			}
		}
	} else {
		if err := e.ProcessFile(node, path, level); err != nil {
			return err
		}
	}

	return nil
}
