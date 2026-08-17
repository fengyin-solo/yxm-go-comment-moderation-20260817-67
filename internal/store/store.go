// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"commentmoderation/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// 内容
	CreateContent(c *model.Content) error
	GetContent(id string) (*model.Content, error)
	ListContents() []*model.Content
	UpdateContent(c *model.Content) error
	DeleteContent(id string) error

	// 评论
	CreateComment(c *model.Comment) error
	GetComment(id string) (*model.Comment, error)
	ListComments() []*model.Comment
	UpdateComment(c *model.Comment) error
	DeleteComment(id string) error
	CountCommentsByUserAndContent(userID, contentID string) int

	// 审核记录
	CreateModerationRecord(m *model.ModerationRecord) error
	GetModerationRecord(id string) (*model.ModerationRecord, error)
	ListModerationRecords() []*model.ModerationRecord

	// 规则
	CreateRule(r *model.Rule) error
	GetRule(id string) (*model.Rule, error)
	ListRules() []*model.Rule
	UpdateRule(r *model.Rule) error
	DeleteRule(id string) error

	// 举报
	CreateReport(r *model.Report) error
	GetReport(id string) (*model.Report, error)
	ListReports() []*model.Report
	UpdateReport(r *model.Report) error
}
