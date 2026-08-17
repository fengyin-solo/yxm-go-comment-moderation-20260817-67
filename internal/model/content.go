package model

import (
	"strings"
	"time"
)

// 内容状态。
const (
	ContentOpen   = "open"   // 开放评论
	ContentClosed = "closed" // 关闭评论区
)

// Content 被评论的内容对象（文章/视频等）。
type Content struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Category     string    `json:"category"`
	CommentCount int       `json:"comment_count"` // 冗余：已发布评论数
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Validate 校验内容字段。
func (c *Content) Validate() error {
	c.Title = strings.TrimSpace(c.Title)
	c.Author = strings.TrimSpace(c.Author)
	c.Category = strings.TrimSpace(c.Category)
	if c.Title == "" {
		return NewValidationError("title", "标题不能为空")
	}
	if len(c.Title) > 128 {
		return NewValidationError("title", "标题不能超过 128 个字符")
	}
	if c.Author == "" {
		return NewValidationError("author", "作者不能为空")
	}
	if c.Category == "" {
		c.Category = "general"
	}
	if c.Status == "" {
		c.Status = ContentOpen
	}
	if c.Status != ContentOpen && c.Status != ContentClosed {
		return NewValidationError("status", "内容状态不合法")
	}
	return nil
}

// ContentFilter 内容列表筛选条件。
type ContentFilter struct {
	Category string
	Status   string
	Keyword  string
}

// Match 判断内容是否满足筛选条件。
func (f ContentFilter) Match(c *Content) bool {
	if f.Category != "" && c.Category != f.Category {
		return false
	}
	if f.Status != "" && c.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(c.Title), k) &&
			!strings.Contains(strings.ToLower(c.Author), k) {
			return false
		}
	}
	return true
}
