package helper

import "strings"

// UpdateOperation 定义更新操作的结构体
type UpdateOperation struct {
	Operation string // 操作类型：add, insert, update, delete, replace, replaceAll
	Position  string // 插入位置：before, after
	Target    string // 目标内容
	Content   string // 新增或更新的内容
}

// ApplyChanges 利用更新数组完成对原来文档的更新
func ApplyChanges(blogContent string, changes []UpdateOperation) string {
	for _, change := range changes {
		switch change.Operation {
		case "add":
			if change.Position == "end" {
				blogContent += "\n" + change.Content
			}
		case "insert":
			target := change.Target
			position := change.Position
			content := change.Content
			if position == "before" {
				blogContent = strings.Replace(blogContent, target, content+"\n"+target, 1)
			} else if position == "after" {
				blogContent = strings.Replace(blogContent, target, target+"\n"+content, 1)
			}
		case "update":
			target := change.Target
			content := change.Content
			blogContent = strings.Replace(blogContent, target, content, 1)
		case "delete":
			target := change.Target
			blogContent = strings.Replace(blogContent, target, "", 1)
		case "replace":
			target := change.Target
			content := change.Content
			blogContent = strings.Replace(blogContent, target, content, 1)
		case "replaceAll":
			target := change.Target
			content := change.Content
			blogContent = strings.ReplaceAll(blogContent, target, content)
		}
	}
	return blogContent
}
