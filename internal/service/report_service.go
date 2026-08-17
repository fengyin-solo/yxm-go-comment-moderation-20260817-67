package service

import (
	"sort"
	"time"

	"commentmoderation/internal/model"
	"commentmoderation/pkg/idgen"
)

// CreateReport 用户举报评论。同一用户对同一评论的未处理举报不可重复提交。
func (s *Service) CreateReport(commentID, reporterID, reason string) (*model.Report, error) {
	if _, err := s.store.GetComment(commentID); err != nil {
		return nil, err
	}
	for _, r := range s.store.ListReports() {
		if r.CommentID == commentID && r.ReporterID == reporterID {
			return nil, model.NewValidationError("report", "已有未处理的重复举报")
		}
	}
	report := &model.Report{
		ID:         idgen.Hex(),
		CommentID:  commentID,
		ReporterID: reporterID,
		Reason:     reason,
		Status:     model.ReportOpen,
		CreatedAt:  time.Now(),
	}
	if err := report.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateReport(report); err != nil {
		return nil, err
	}
	s.log.Infof("评论 %s 被用户 %s 举报", commentID, reporterID)
	return report, nil
}

// GetReport 查询举报详情。
func (s *Service) GetReport(id string) (*model.Report, error) {
	return s.store.GetReport(id)
}

// ListReports 分页查询举报列表，按时间倒序。
func (s *Service) ListReports(filter model.ReportFilter, page, size int) ([]*model.Report, int, error) {
	all := s.store.ListReports()
	matched := make([]*model.Report, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Report{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// ResolveReport 认定举报成立：举报置为 resolved，并删除被举报的已发布评论。
func (s *Service) ResolveReport(reportID, handler string) (*model.Report, error) {
	report, err := s.store.GetReport(reportID)
	if err != nil {
		return nil, err
	}
	if !model.CanReportTransition(report.Status, model.ReportResolved) {
		return nil, model.NewValidationError("status", "当前状态不可处理")
	}
	if handler == "" {
		return nil, model.NewValidationError("handled_by", "处理人不能为空")
	}
	report.Status = model.ReportResolved
	report.HandledBy = handler
	report.HandledAt = time.Now()
	if err := s.store.UpdateReport(report); err != nil {
		return nil, err
	}
	// 处置被举报评论：已发布的删除；待审核的直接拒绝
	if c, err := s.store.GetComment(report.CommentID); err == nil {
		switch c.Status {
		case model.CommentApproved:
			_ = s.DeleteComment(c.ID)
		case model.CommentPending:
			c.Status = model.CommentRejected
			c.ModeratedAt = time.Now()
			_ = s.store.UpdateComment(c)
			s.writeModeration(c.ID, handler, model.ModerationReject, model.ModerationSourceManual, "举报成立")
		}
	}
	s.log.Infof("举报 %s 处理成立", reportID)
	return report, nil
}

// DismissReport 认定举报不成立，驳回。
func (s *Service) DismissReport(reportID, handler string) (*model.Report, error) {
	report, err := s.store.GetReport(reportID)
	if err != nil {
		return nil, err
	}
	if !model.CanReportTransition(report.Status, model.ReportDismissed) {
		return nil, model.NewValidationError("status", "当前状态不可驳回")
	}
	if handler == "" {
		return nil, model.NewValidationError("handled_by", "处理人不能为空")
	}
	report.Status = model.ReportResolved
	report.HandledBy = handler
	report.HandledAt = time.Now()
	if err := s.store.UpdateReport(report); err != nil {
		return nil, err
	}
	s.log.Infof("举报 %s 已驳回", reportID)
	return report, nil
}
