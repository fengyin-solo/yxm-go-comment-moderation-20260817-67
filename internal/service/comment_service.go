package service

import (
	"sort"
	"strings"
	"time"

	"commentmoderation/internal/model"
	"commentmoderation/pkg/idgen"
)

// PostComment 用户发表评论。
// 流程：校验内容与评论区开放 -> 规则引擎自动审核 -> 命中规则直接拒绝，否则进入待审核。
func (s *Service) PostComment(contentID, userID, body string) (*model.Comment, error) {
	content, err := s.store.GetContent(contentID)
	if err != nil {
		return nil, err
	}
	if content.Status != model.ContentOpen {
		return nil, model.NewValidationError("content", "该内容评论区已关闭")
	}
	comment := &model.Comment{
		ID:        idgen.Hex(),
		ContentID: contentID,
		UserID:    userID,
		Body:      body,
		Status:    model.CommentPending,
		CreatedAt: time.Now(),
	}
	if err := comment.Validate(); err != nil {
		return nil, err
	}
	// 规则引擎自动审核
	if hitRule, reason := s.autoModerate(comment); hitRule != nil {
		comment.Status = model.CommentRejected
		comment.AutoFlag = true
		comment.ModeratedAt = time.Now()
		if err := s.store.CreateComment(comment); err != nil {
			return nil, err
		}
		s.writeModeration(comment.ID, "rule:"+hitRule.ID, model.ModerationReject,
			model.ModerationSourceAuto, reason)
		s.log.Infof("评论 %s 被规则 %s 自动拒绝: %s", comment.ID, hitRule.Name, reason)
		return comment, nil
	}
	if err := s.store.CreateComment(comment); err != nil {
		return nil, err
	}
	s.log.Infof("评论 %s 已提交，进入待审核", comment.ID)
	return comment, nil
}

// autoModerate 用生效规则检查评论，返回命中的规则与原因；未命中返回 nil。
func (s *Service) autoModerate(comment *model.Comment) (*model.Rule, string) {
	bodyLower := strings.ToLower(comment.Body)
	for _, r := range s.store.ListRules() {
		if !r.AppliesTo(comment.ContentID) {
			continue
		}
		switch r.Type {
		case model.RuleKeyword:
			for _, kw := range r.Keywords() {
				if strings.Contains(bodyLower, kw) {
					return r, "命中敏感关键词"
				}
			}
		case model.RuleMinLength:
			if int64(len([]rune(comment.Body))) < r.LimitValue {
				return r, "评论内容过短"
			}
		case model.RuleMaxPerUser:
			count := s.store.CountCommentsByUserAndContent(comment.UserID, comment.ContentID)
			if int64(count) >= r.LimitValue {
				return r, "超过单用户评论数上限"
			}
		}
	}
	return nil, ""
}

// GetComment 查询评论详情。
func (s *Service) GetComment(id string) (*model.Comment, error) {
	return s.store.GetComment(id)
}

// ListComments 分页查询评论列表，按创建时间倒序。
func (s *Service) ListComments(filter model.CommentFilter, page, size int) ([]*model.Comment, int, error) {
	all := s.store.ListComments()
	matched := make([]*model.Comment, 0, len(all))
	for _, c := range all {
		if filter.Match(c) {
			matched = append(matched, c)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Comment{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// ApproveComment 人工审核通过评论，同步内容评论计数。
func (s *Service) ApproveComment(commentID, moderator, reason string) (*model.Comment, error) {
	c, err := s.store.GetComment(commentID)
	if err != nil {
		return nil, err
	}
	if !model.CanCommentTransition(c.Status, model.CommentApproved) {
		return nil, model.NewValidationError("status", "当前状态不可通过")
	}
	c.Status = model.CommentApproved
	c.ModeratedAt = time.Now()
	if err := s.store.UpdateComment(c); err != nil {
		return nil, err
	}
	s.bumpContentCommentCount(c.ContentID, 1)
	s.writeModeration(commentID, moderator, model.ModerationApprove, model.ModerationSourceManual, reason)
	s.log.Infof("评论 %s 人工审核通过", commentID)
	return c, nil
}

// RejectComment 人工审核拒绝评论。
func (s *Service) RejectComment(commentID, moderator, reason string) (*model.Comment, error) {
	c, err := s.store.GetComment(commentID)
	if err != nil {
		return nil, err
	}
	if !model.CanCommentTransition(c.Status, model.CommentRejected) {
		return nil, model.NewValidationError("status", "当前状态不可拒绝")
	}
	if reason == "" {
		return nil, model.NewValidationError("reason", "拒绝必须填写理由")
	}
	c.Status = model.CommentRejected
	c.ModeratedAt = time.Now()
	if err := s.store.UpdateComment(c); err != nil {
		return nil, err
	}
	s.writeModeration(commentID, moderator, model.ModerationReject, model.ModerationSourceManual, reason)
	s.log.Infof("评论 %s 人工审核拒绝", commentID)
	return c, nil
}

// DeleteComment 删除已发布的评论，同步内容评论计数。
func (s *Service) DeleteComment(commentID string) error {
	c, err := s.store.GetComment(commentID)
	if err != nil {
		return err
	}
	if !model.CanCommentTransition(c.Status, model.CommentDeleted) {
		return model.NewValidationError("status", "仅已发布的评论可删除")
	}
	wasApproved := c.Status == model.CommentApproved
	c.Status = model.CommentDeleted
	if err := s.store.UpdateComment(c); err != nil {
		return err
	}
	if wasApproved {
		s.bumpContentCommentCount(c.ContentID, -1)
	}
	return nil
}

// BatchApprovePending 批量通过某内容下的全部待审核评论，返回通过数量。
// 仅处理 pending 状态的评论，其余状态跳过。
func (s *Service) BatchApprovePending(contentID, moderator, reason string) (int, error) {
	if _, err := s.store.GetContent(contentID); err != nil {
		return 0, err
	}
	if moderator == "" {
		return 0, model.NewValidationError("moderator", "审核人不能为空")
	}
	approved := 0
	for _, c := range s.store.ListComments() {
		if c.ContentID != contentID || c.Status != model.CommentPending {
			continue
		}
		c.Status = model.CommentApproved
		c.ModeratedAt = time.Now()
		if err := s.store.UpdateComment(c); err != nil {
			continue
		}
		s.bumpContentCommentCount(c.ContentID, 1)
		s.writeModeration(c.ID, moderator, model.ModerationApprove, model.ModerationSourceManual, reason)
		approved++
	}
	if approved > 0 {
		s.log.Infof("内容 %s 批量通过 %d 条待审核评论", contentID, approved)
	}
	return approved, nil
}

// bumpContentCommentCount 增减内容评论计数。
func (s *Service) bumpContentCommentCount(contentID string, delta int) {
	content, err := s.store.GetContent(contentID)
	if err != nil {
		return
	}
	content.CommentCount += delta
	if content.CommentCount < 0 {
		content.CommentCount = 0
	}
	_ = s.store.UpdateContent(content)
}

// writeModeration 写入一条审核记录。
func (s *Service) writeModeration(commentID, moderator, action, source, reason string) {
	record := &model.ModerationRecord{
		ID:        idgen.Hex(),
		CommentID: commentID,
		Moderator: moderator,
		Action:    action,
		Source:    source,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
	if err := record.Validate(); err != nil {
		s.log.Warnf("审核记录校验失败: %v", err)
		return
	}
	_ = s.store.CreateModerationRecord(record)
}
