package project

// ContentCollector 定义内容收集的接口
type ContentCollector interface {
	// AddTitle 添加标题
	AddTitle(title string, level int) error
	// AddContent 添加内容
	AddContent(content string) error
	// Render 渲染最终结果
	Render(outputPath string) error
}

// Exporter 定义了项目导出器的接口
type Exporter interface {
	NodeVisitor
	Export(outputPath string) error
}

// BaseExporter 提供了基本的导出功能
type BaseExporter struct {
	project   *Project
	collector ContentCollector
}

// NewBaseExporter 创建一个基本导出器
func NewBaseExporter(p *Project, collector ContentCollector) *BaseExporter {
	return &BaseExporter{
		project:   p,
		collector: collector,
	}
}

// Export 实现通用的导出逻辑
func (b *BaseExporter) Export(outputPath string) error {
	if b.project.root == nil || len(b.project.root.Children) == 0 {
		return b.collector.AddTitle("空项目", 1)
	}

	traverser := NewTreeTraverser(b.project)
	if err := traverser.Traverse(b.project.root, "/", 0, b); err != nil {
		return err
	}

	return b.collector.Render(outputPath)
}

// VisitDirectory 实现通用的目录访问逻辑
func (b *BaseExporter) VisitDirectory(node *Node, path string, level int) error {
	if path != "/" {
		return b.collector.AddTitle(node.Name, level)
	}
	return nil
}

// VisitFile 实现通用的文件访问逻辑
func (b *BaseExporter) VisitFile(node *Node, path string, level int) error {
	if err := b.collector.AddTitle(node.Name, level); err != nil {
		return err
	}
	return b.collector.AddContent(string(node.Content))
}
