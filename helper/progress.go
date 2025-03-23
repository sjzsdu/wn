package helper

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Progress struct {
	total     int
	current   int64
	width     int
	title     string
}

func NewProgress(title string, total int) *Progress {
	return &Progress{
		total:   total,
		width:   50, // 进度条宽度
		title:   title,
	}
}

func (p *Progress) Increment() {
	atomic.AddInt64(&p.current, 1)
	p.render()
}

func (p *Progress) render() {
	current := atomic.LoadInt64(&p.current)
	percentage := float64(current) / float64(p.total) * 100
	filled := int(float64(p.width) * float64(current) / float64(p.total))
	
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", p.width-filled)
	
	fmt.Printf("\r%s [%s] %.1f%% (%d/%d)", 
		p.title, bar, percentage, current, p.total)
	
	if current >= int64(p.total) {
		fmt.Println() // 完成时换行
	}
}