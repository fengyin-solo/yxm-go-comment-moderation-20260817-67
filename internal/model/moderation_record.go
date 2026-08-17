package model

import (
	"strings"
	"time"
)

// 审核动作。
const (
	ModerationApprove = "approve" // 通过
	ModerationReject  = "reject"  // 拒绝
)

// 审核来源。
const (
	ModerationSourceAuto = "auto"  // 规则引擎自动审核
	ModerationSourceManual = "manual" // 人工审核
)

// ModerationRecord 审核记录，一次审核动作的不可变留痕。
type ModerationRecord struct {
	ID         string    `json:"id"`
	CommentID  string    `json:"comment_id"`
	Moderator  string    `json:"moderator"` // 人工审核为审核员 ID，自动审核为 rule:<规则ID>
	Action     string    `json:"action"`
	Source     string    `json:"source"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate 校验审核记录字段。
func (m *ModerationRecord) Validate() error {
	m.CommentID = strings.TrimSpace(m.CommentID)
	m.Moderator = strings.TrimSpace(m.Moderator)
	m.Reason = strings.TrimSpace(m.Reason)
	if m.CommentID == "" {
		return NewValidationError("comment_id", "评论 ID 不能为空")
	}
	if m.Moderator == "" {
		return NewValidationError("moderator", "审核人不能为空")
	}
	switch m.Action {
	case ModerationApprove, ModerationReject:
	default:
		return NewValidationError("action", "审核动作不合法")
	}
	if m.Source != ModerationSourceAuto && m.Source != ModerationSourceManual {
		return NewValidationError("source", "审核来源不合法")
	}
	if len(m.Reason) > 200 {
		return NewValidationError("reason", "审核理由不能超过 200 个字符")
	}
	if m.CreatedAt.IsZero() {
		return NewValidationError("created_at", "记录时间不能为空")
	}
	return nil
}

// ModerationFilter 审核记录筛选条件。
type ModerationFilter struct {
	CommentID string
	Action    string
	Source    string
}

// Match 判断审核记录是否满足筛选条件。
func (f ModerationFilter) Match(m *ModerationRecord) bool {
	if f.CommentID != "" && m.CommentID != f.CommentID {
		return false
	}
	if f.Action != "" && m.Action != f.Action {
		return false
	}
	if f.Source != "" && m.Source != f.Source {
		return false
	}
	return true
}
