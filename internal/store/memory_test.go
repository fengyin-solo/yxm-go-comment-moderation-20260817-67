package store

import (
	"testing"
	"time"

	"commentmoderation/internal/model"
)

func newContent(id, title string) *model.Content {
	return &model.Content{
		ID:        id,
		Title:     title,
		Author:    "作者",
		Category:  "tech",
		Status:    model.ContentOpen,
		CreatedAt: time.Now(),
	}
}

func TestContentCRUD(t *testing.T) {
	s := NewMemoryStore()
	c := newContent("c1", "Go 入门")
	if err := s.CreateContent(c); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetContent("c1"); err != nil || got.Title != "Go 入门" {
		t.Fatalf("查询内容失败: %v", err)
	}
	if list := s.ListContents(); len(list) != 1 {
		t.Fatalf("期望 1 个内容，得到 %d", len(list))
	}
	c.Title = "Go 进阶"
	if err := s.UpdateContent(c); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteContent("c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetContent("c1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
	if err := s.DeleteContent("c1"); err != ErrNotFound {
		t.Fatalf("重复删除期望 ErrNotFound，得到 %v", err)
	}
	if err := s.UpdateContent(newContent("missing", "x")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func newComment(id, contentID, userID string) *model.Comment {
	return &model.Comment{
		ID:        id,
		ContentID: contentID,
		UserID:    userID,
		Body:      "评论内容",
		Status:    model.CommentPending,
		CreatedAt: time.Now(),
	}
}

func TestCommentCRUD(t *testing.T) {
	s := NewMemoryStore()
	c := newComment("cm1", "c1", "u1")
	if err := s.CreateComment(c); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetComment("cm1"); err != nil || got.UserID != "u1" {
		t.Fatalf("查询评论失败: %v", err)
	}
	if list := s.ListComments(); len(list) != 1 {
		t.Fatalf("期望 1 条评论，得到 %d", len(list))
	}
	c.Status = model.CommentApproved
	if err := s.UpdateComment(c); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteComment("cm1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetComment("cm1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
	if err := s.UpdateComment(newComment("missing", "c1", "u1")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func TestCountCommentsByUserAndContent(t *testing.T) {
	s := NewMemoryStore()
	_ = s.CreateComment(newComment("cm1", "c1", "u1"))
	_ = s.CreateComment(newComment("cm2", "c1", "u1"))
	_ = s.CreateComment(newComment("cm3", "c2", "u1"))
	_ = s.CreateComment(newComment("cm4", "c1", "u2"))
	// 已删除的不计数
	deleted := newComment("cm5", "c1", "u1")
	deleted.Status = model.CommentDeleted
	_ = s.CreateComment(deleted)

	if got := s.CountCommentsByUserAndContent("u1", "c1"); got != 2 {
		t.Fatalf("期望 2，得到 %d", got)
	}
	if got := s.CountCommentsByUserAndContent("u2", "c1"); got != 1 {
		t.Fatalf("期望 1，得到 %d", got)
	}
	if got := s.CountCommentsByUserAndContent("u1", "c9"); got != 0 {
		t.Fatalf("期望 0，得到 %d", got)
	}
}

func newModeration(id, commentID string) *model.ModerationRecord {
	return &model.ModerationRecord{
		ID:        id,
		CommentID: commentID,
		Moderator: "mod1",
		Action:    model.ModerationApprove,
		Source:    model.ModerationSourceManual,
		CreatedAt: time.Now(),
	}
}

func TestModerationRecordCreateAndList(t *testing.T) {
	s := NewMemoryStore()
	if err := s.CreateModerationRecord(newModeration("m1", "cm1")); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateModerationRecord(newModeration("m2", "cm1")); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetModerationRecord("m1"); err != nil || got.CommentID != "cm1" {
		t.Fatalf("查询审核记录失败: %v", err)
	}
	if _, err := s.GetModerationRecord("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListModerationRecords(); len(list) != 2 {
		t.Fatalf("期望 2 条记录，得到 %d", len(list))
	}
}

func newRule(id, name, ruleType string) *model.Rule {
	r := &model.Rule{
		ID:        id,
		Name:      name,
		Type:      ruleType,
		Status:    model.RuleActive,
		CreatedAt: time.Now(),
	}
	switch ruleType {
	case model.RuleKeyword:
		r.TextValue = "敏感词"
	case model.RuleMinLength, model.RuleMaxPerUser:
		r.LimitValue = 5
	}
	return r
}

func TestRuleCRUD(t *testing.T) {
	s := NewMemoryStore()
	r := newRule("r1", "关键词过滤", model.RuleKeyword)
	if err := s.CreateRule(r); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetRule("r1"); err != nil || got.Name != "关键词过滤" {
		t.Fatalf("查询规则失败: %v", err)
	}
	if list := s.ListRules(); len(list) != 1 {
		t.Fatalf("期望 1 条规则，得到 %d", len(list))
	}
	r.Status = model.RuleInactive
	if err := s.UpdateRule(r); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRule("r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetRule("r1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
	if err := s.UpdateRule(newRule("missing", "x", model.RuleKeyword)); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if err := s.DeleteRule("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func newReport(id, commentID string) *model.Report {
	return &model.Report{
		ID:         id,
		CommentID:  commentID,
		ReporterID: "u9",
		Reason:     "垃圾广告",
		Status:     model.ReportOpen,
		CreatedAt:  time.Now(),
	}
}

func TestReportCRUD(t *testing.T) {
	s := NewMemoryStore()
	r := newReport("rp1", "cm1")
	if err := s.CreateReport(r); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetReport("rp1"); err != nil || got.Reason != "垃圾广告" {
		t.Fatalf("查询举报失败: %v", err)
	}
	if list := s.ListReports(); len(list) != 1 {
		t.Fatalf("期望 1 条举报，得到 %d", len(list))
	}
	r.Status = model.ReportResolved
	if err := s.UpdateReport(r); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateReport(newReport("missing", "cm1")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}
