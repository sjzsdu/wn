package project

import (
	"fmt"
	"path/filepath"
	"sort"
	"sync"
)

// NodeVisitor 定义了节点访问器的接口
type NodeVisitor interface {
	// VisitDirectory 访问目录节点
	VisitDirectory(node *Node, path string, level int) error
	// VisitFile 访问文件节点
	VisitFile(node *Node, path string, level int) error
}

// TraverseOrder 定义遍历顺序
type TraverseOrder int

const (
	PreOrder  TraverseOrder = iota // 前序遍历
	PostOrder                      // 后序遍历
	InOrder                        // 中序遍历
)

// TraverseOption 定义遍历选项
type TraverseOption struct {
	ContinueOnError bool    // 遇到错误时是否继续
	Errors          []error // 记录所有错误
}

// TreeTraverser 提供了树遍历的基本功能
type TreeTraverser struct {
	project *Project
	order   TraverseOrder
	option  *TraverseOption
	wg      sync.WaitGroup  // 添加等待组
}

// SetOption 设置遍历选项
func (t *TreeTraverser) SetOption(option *TraverseOption) {
	t.option = option
}

// handleError 处理遍历过程中的错误
func (t *TreeTraverser) handleError(err error) error {
	if t.option != nil && t.option.ContinueOnError {
		t.option.Errors = append(t.option.Errors, err)
		return nil
	}
	return err
}

// NewTreeTraverser 创建一个树遍历器，默认使用前序遍历
func NewTreeTraverser(p *Project) *TreeTraverser {
	return &TreeTraverser{
		project: p,
		order:   PreOrder,
		option:  nil,
	}
}

// SetTraverseOrder 设置遍历顺序
func (t *TreeTraverser) SetTraverseOrder(order TraverseOrder) *TreeTraverser {
	t.order = order
	return t
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
		// 对子节点进行排序以保证遍历顺序一致
		children := make([]*Node, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, child)
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name < children[j].Name
		})

		// 创建错误通道
		errChan := make(chan error, len(children))
		
		switch t.order {
		case PreOrder:
			if err := visitor.VisitDirectory(node, path, level); err != nil {
				return err
			}
			for _, child := range children {
				childPath := filepath.Join(path, child.Name)
				if err := t.Traverse(child, childPath, level+1, visitor); err != nil {
					return err
				}
			}

		case PostOrder:
			// 并发处理子节点
			for _, child := range children {
				childPath := filepath.Join(path, child.Name)
				if !child.IsDir {
					t.wg.Add(1)
					go func(c *Node, p string) {
						defer t.wg.Done()
						if err := t.Traverse(c, p, level+1, visitor); err != nil {
							errChan <- err
						}
					}(child, childPath)
				} else {
					// 目录节点保持同步处理
					if err := t.Traverse(child, childPath, level+1, visitor); err != nil {
						return err
					}
				}
			}

			// 等待所有文件处理完成
			t.wg.Wait()
			
			// 检查是否有错误发生
			select {
			case err := <-errChan:
				return err
			default:
			}

			// 处理当前目录
			if err := visitor.VisitDirectory(node, path, level); err != nil {
				return err
			}

		case InOrder:
			mid := len(children) / 2

			// 遍历前半部分子节点
			for i := 0; i < mid; i++ {
				childPath := filepath.Join(path, children[i].Name)
				if err := t.Traverse(children[i], childPath, level+1, visitor); err != nil {
					return err
				}
			}

			// 访问当前节点
			if err := visitor.VisitDirectory(node, path, level); err != nil {
				return err
			}

			// 遍历后半部分子节点
			for i := mid; i < len(children); i++ {
				childPath := filepath.Join(path, children[i].Name)
				if err := t.Traverse(children[i], childPath, level+1, visitor); err != nil {
					return err
				}
			}
		}
	} else {
		if err := visitor.VisitFile(node, path, level); err != nil {
			fmt.Println("visit file:", err)
			return err
		}
	}

	return nil
}
