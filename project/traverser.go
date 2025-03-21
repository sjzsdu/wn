package project

import "path/filepath"

// NodeVisitor 定义了节点访问器的接口
type NodeVisitor interface {
	// VisitDirectory 访问目录节点
	VisitDirectory(node *Node, path string, level int) error
	// VisitFile 访问文件节点
	VisitFile(node *Node, path string, level int) error
}

// TreeTraverser 提供了树遍历的基本功能
type TreeTraverser struct {
	project *Project
}

// NewTreeTraverser 创建一个树遍历器
func NewTreeTraverser(p *Project) *TreeTraverser {
	return &TreeTraverser{project: p}
}

// TraverseTree 遍历整个项目树
func (t *TreeTraverser) TraverseTree(visitor NodeVisitor) error {
	if t.project.root == nil {
		return nil
	}
	return t.Traverse(t.project.root, "/", 0, visitor)
}

// Traverse 遍历节点的通用方法
func (t *TreeTraverser) Traverse(node *Node, path string, level int, visitor NodeVisitor) error {
	if node == nil {
		return nil
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.IsDir {
		if err := visitor.VisitDirectory(node, path, level); err != nil {
			return err
		}

		for _, child := range node.Children {
			childPath := filepath.Join(path, child.Name)
			if err := t.Traverse(child, childPath, level+1, visitor); err != nil {
				return err
			}
		}
	} else {
		if err := visitor.VisitFile(node, path, level); err != nil {
			return err
		}
	}

	return nil
}
