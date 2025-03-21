package file

import (
	"github.com/sjzsdu/wn/project"
)

// Collector 定义内容收集的接口
type Collector interface {
	// AddTitle 添加标题
	AddTitle(title string, level int) error
	// AddContent 添加内容
	AddContent(content string) error
	// Render 渲染最终结果
	Render(outputPath string) error
}

// BaseExporter 提供基本的导出功能
type BaseExporter struct {
	project   *project.Project
	collector Collector
}

// NewBaseExporter 创建一个基本导出器
func NewBaseExporter(p *project.Project, collector Collector) *BaseExporter {
	return &BaseExporter{
		project:   p,
		collector: collector,
	}
}

// VisitDirectory 实现 project.NodeVisitor 接口
func (b *BaseExporter) VisitDirectory(node *project.Node, path string, level int) error {
	if path == "/" {
		return nil
	}

	// 尝试添加目录项（如果收集器支持的话）
	if tocCollector, ok := b.collector.(interface{ AddTOCItem(string, int) error }); ok {
		if err := tocCollector.AddTOCItem(node.Name, level); err != nil {
			return err
		}
	}

	return b.collector.AddTitle(node.Name, level)
}

// VisitFile 实现 project.NodeVisitor 接口
func (b *BaseExporter) VisitFile(node *project.Node, path string, level int) error {
	// 尝试添加目录项（如果收集器支持的话）
	if tocCollector, ok := b.collector.(interface{ AddTOCItem(string, int) error }); ok {
		if err := tocCollector.AddTOCItem(node.Name, level); err != nil {
			return err
		}
	}

	if err := b.collector.AddTitle(node.Name, level); err != nil {
		return err
	}
	return b.collector.AddContent(string(node.Content))
}

// Export 执行导出操作
func (b *BaseExporter) Export(outputPath string) error {
	if b.project == nil || b.project.IsEmpty() {
		return b.collector.AddTitle("空项目", 1)
	}

	// 创建一个树遍历器
	traverser := project.NewTreeTraverser(b.project)
	if err := traverser.TraverseTree(b); err != nil {
		return err
	}

	return b.collector.Render(outputPath)
}
