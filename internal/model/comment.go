package model

import (
	"strings"
	"time"
)

// 评论状态。
const (
	CommentPending  = "pending"  // 待审核
	CommentApproved = "approved" // 已发布
	CommentRejected = "rejected" // 已拒绝，终态
	CommentDeleted  = "deleted"  // 已删除，终态
)

// commentTransitions 评论状态机：pending -> approved/rejected，approved -> deleted。
var commentTransitions = map[string]map[string]bool{
	CommentPending:  {CommentApproved: true, CommentRejected: true},
	CommentApproved: {CommentDeleted: true},
}

// CanCommentTransition 判断评论状态流转是否合法。
func CanCommentTransition(from, to string) bool {
	if m, ok := commentTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Comment 评论，必须挂在某个内容下。
type Comment struct {
	ID         string    `json:"id"`
	ContentID  string    `json:"content_id"`
	UserID     string    `json:"user_id"`
	Body       string    `json:"body"`
	Status     string    `json:"status"`
	AutoFlag   bool      `json:"auto_flag"` // 是否被规则引擎自动标记
	CreatedAt  time.Time `json:"created_at"`
	ModeratedAt time.Time `json:"moderated_at"`
}

// Validate 校验评论字段。
func (c *Comment) Validate() error {
	c.ContentID = strings.TrimSpace(c.ContentID)
	c.UserID = strings.TrimSpace(c.UserID)
	c.Body = strings.TrimSpace(c.Body)
	if c.ContentID == "" {
		return NewValidationError("content_id", "内容 ID 不能为空")
	}
	if c.UserID == "" {
		return NewValidationError("user_id", "用户 ID 不能为空")
	}
	if c.Body == "" {
		return NewValidationError("body", "评论内容不能为空")
	}
	if len(c.Body) > 1000 {
		return NewValidationError("body", "评论内容不能超过 1000 个字符")
	}
	if c.Status == "" {
		c.Status = CommentPending
	}
	switch c.Status {
	case CommentPending, CommentApproved, CommentRejected, CommentDeleted:
	default:
		return NewValidationError("status", "评论状态不合法")
	}
	return nil
}

// CommentFilter 评论列表筛选条件。
type CommentFilter struct {
	ContentID string
	UserID    string
	Status    string
}

// Match 判断评论是否满足筛选条件。
func (f CommentFilter) Match(c *Comment) bool {
	if f.ContentID != "" && c.ContentID != f.ContentID {
		return false
	}
	if f.UserID != "" && c.UserID != f.UserID {
		return false
	}
	if f.Status != "" && c.Status != f.Status {
		return false
	}
	return true
}
