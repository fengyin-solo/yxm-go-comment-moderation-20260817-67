package model

import (
	"strings"
	"time"
)

// 举报状态。
const (
	ReportOpen     = "open"     // 待处理
	ReportResolved = "resolved" // 已处理（认定违规并处置）
	ReportDismissed = "dismissed" // 已驳回（认定不违规）
)

// reportTransitions 举报状态机：open -> resolved / dismissed。
var reportTransitions = map[string]map[string]bool{
	ReportOpen: {ReportDismissed: true},
}

// CanReportTransition 判断举报状态流转是否合法。
func CanReportTransition(from, to string) bool {
	if m, ok := reportTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Report 用户对评论的举报。
type Report struct {
	ID         string    `json:"id"`
	CommentID  string    `json:"comment_id"`
	ReporterID string    `json:"reporter_id"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	HandledBy  string    `json:"handled_by"` // 处理人
	CreatedAt  time.Time `json:"created_at"`
	HandledAt  time.Time `json:"handled_at"`
}

// Validate 校验举报字段。
func (r *Report) Validate() error {
	r.CommentID = strings.TrimSpace(r.CommentID)
	r.ReporterID = strings.TrimSpace(r.ReporterID)
	r.Reason = strings.TrimSpace(r.Reason)
	r.HandledBy = strings.TrimSpace(r.HandledBy)
	if r.CommentID == "" {
		return NewValidationError("comment_id", "评论 ID 不能为空")
	}
	if r.ReporterID == "" {
		return NewValidationError("reporter_id", "举报人不能为空")
	}
	if r.Reason == "" {
		return NewValidationError("reason", "举报理由不能为空")
	}
	if len(r.Reason) > 500 {
		return NewValidationError("reason", "举报理由不能超过 500 个字符")
	}
	if r.Status == "" {
		r.Status = ReportOpen
	}
	switch r.Status {
	case ReportOpen, ReportResolved, ReportDismissed:
	default:
		return NewValidationError("status", "举报状态不合法")
	}
	return nil
}

// ReportFilter 举报列表筛选条件。
type ReportFilter struct {
	CommentID string
	Status    string
}

// Match 判断举报是否满足筛选条件。
func (f ReportFilter) Match(r *Report) bool {
	if f.CommentID != "" && r.CommentID != f.CommentID {
		return false
	}
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	return true
}
