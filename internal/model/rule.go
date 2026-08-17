package model

import (
	"strings"
	"time"
)

// 规则类型。
const (
	RuleKeyword     = "keyword"      // 命中关键词直接拒绝，TextValue 为关键词（逗号分隔多个）
	RuleMinLength   = "min_length"   // 评论长度下限，LimitValue 为字符数
	RuleMaxPerUser  = "max_per_user" // 单用户在同一内容下的评论数上限，LimitValue 为条数
)

// 规则状态。
const (
	RuleActive   = "active"
	RuleInactive = "inactive"
)

// Rule 评论审核规则，可作用于全局（ContentID 为空）或指定内容。
type Rule struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	ContentID  string    `json:"content_id"` // 为空表示全局规则
	TextValue  string    `json:"text_value"`
	LimitValue int64     `json:"limit_value"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate 校验规则字段。
func (r *Rule) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.ContentID = strings.TrimSpace(r.ContentID)
	r.TextValue = strings.TrimSpace(r.TextValue)
	if r.Name == "" {
		return NewValidationError("name", "规则名称不能为空")
	}
	switch r.Type {
	case RuleKeyword:
		if r.TextValue == "" {
			return NewValidationError("text_value", "关键词不能为空")
		}
	case RuleMinLength:
		if r.LimitValue <= 0 {
			return NewValidationError("limit_value", "长度下限必须大于 0")
		}
	case RuleMaxPerUser:
		if r.LimitValue <= 0 {
			return NewValidationError("limit_value", "评论数上限必须大于 0")
		}
	default:
		return NewValidationError("type", "规则类型不合法")
	}
	if r.Status == "" {
		r.Status = RuleActive
	}
	if r.Status != RuleActive && r.Status != RuleInactive {
		return NewValidationError("status", "规则状态不合法")
	}
	return nil
}

// AppliesTo 判断规则是否作用于指定内容（全局规则作用于所有内容）。
func (r *Rule) AppliesTo(contentID string) bool {
	if r.Status != RuleActive {
		return false
	}
	return r.ContentID == "" || r.ContentID == contentID
}

// Keywords 拆分关键词规则的关键词列表。
func (r *Rule) Keywords() []string {
	parts := strings.Split(r.TextValue, ",")
	list := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(p)
		if p != "" {
			list = append(list, p)
		}
	}
	return list
}

// RuleFilter 规则列表筛选条件。
type RuleFilter struct {
	Type      string
	Status    string
	ContentID string
}

// Match 判断规则是否满足筛选条件。
func (f RuleFilter) Match(r *Rule) bool {
	if f.Type != "" && r.Type != f.Type {
		return false
	}
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.ContentID != "" && r.ContentID != f.ContentID {
		return false
	}
	return true
}
