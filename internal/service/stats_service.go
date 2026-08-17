package service

import (
	"sort"

	"commentmoderation/internal/model"
)

// OverviewStats 全局概览统计。
type OverviewStats struct {
	ContentCount    int `json:"content_count"`
	CommentCount    int `json:"comment_count"`
	PendingCount    int `json:"pending_count"`
	ApprovedCount   int `json:"approved_count"`
	RejectedCount   int `json:"rejected_count"`
	AutoRejected    int `json:"auto_rejected"`    // 规则引擎自动拒绝数
	ReportCount     int `json:"report_count"`
	OpenReportCount int `json:"open_report_count"`
	AutoRate        int `json:"auto_rate"`        // 自动拒绝占全部拒绝的百分比
}

// Stats 全局概览统计。
func (s *Service) Stats() *OverviewStats {
	comments := s.store.ListComments()
	reports := s.store.ListReports()
	stats := &OverviewStats{
		ContentCount: len(s.store.ListContents()),
		CommentCount: len(comments),
		ReportCount:  len(reports),
	}
	rejected := 0
	for _, c := range comments {
		switch c.Status {
		case model.CommentPending:
			stats.PendingCount++
		case model.CommentApproved:
			stats.ApprovedCount++
		case model.CommentRejected:
			stats.RejectedCount++
			rejected++
			if c.AutoFlag {
				stats.AutoRejected++
			}
		}
	}
	for _, r := range reports {
		if r.Status == model.ReportOpen {
			stats.OpenReportCount++
		}
	}
	if rejected > 0 {
		stats.AutoRate = (stats.AutoRejected*100 + rejected/2) / rejected
	}
	return stats
}

// ContentStat 单内容统计。
type ContentStat struct {
	ContentID     string `json:"content_id"`
	Title         string `json:"title"`
	CommentCount  int    `json:"comment_count"`
	PendingCount  int    `json:"pending_count"`
	RejectedCount int    `json:"rejected_count"`
	ReportCount   int    `json:"report_count"`
}

// StatsByContent 按内容分组统计评论与举报情况，按评论数降序。
func (s *Service) StatsByContent() []*ContentStat {
	byContent := make(map[string]*ContentStat)
	for _, c := range s.store.ListContents() {
		byContent[c.ID] = &ContentStat{ContentID: c.ID, Title: c.Title}
	}
	for _, c := range s.store.ListComments() {
		st, ok := byContent[c.ContentID]
		if !ok {
			continue
		}
		switch c.Status {
		case model.CommentPending:
			st.PendingCount++
		case model.CommentApproved:
			st.CommentCount++
		case model.CommentRejected:
			st.RejectedCount++
		}
	}
	for _, r := range s.store.ListReports() {
		if c, err := s.store.GetComment(r.CommentID); err == nil {
			if st, ok := byContent[c.ContentID]; ok {
				st.ReportCount++
			}
		}
	}
	list := make([]*ContentStat, 0, len(byContent))
	for _, st := range byContent {
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].CommentCount != list[j].CommentCount {
			return list[i].CommentCount > list[j].CommentCount
		}
		return list[i].ContentID < list[j].ContentID
	})
	return list
}

// ModeratorStat 审核员工作量统计。
type ModeratorStat struct {
	Moderator    string `json:"moderator"`
	ApproveCount int    `json:"approve_count"`
	RejectCount  int    `json:"reject_count"`
}

// TopModerators 按审核动作次数取 TOP N（仅统计人工审核）。
func (s *Service) TopModerators(n int) []*ModeratorStat {
	if n <= 0 {
		n = 10
	}
	byModerator := make(map[string]*ModeratorStat)
	for _, m := range s.store.ListModerationRecords() {
		if m.Source != model.ModerationSourceManual {
			continue
		}
		st, ok := byModerator[m.Moderator]
		if !ok {
			st = &ModeratorStat{Moderator: m.Moderator}
			byModerator[m.Moderator] = st
		}
		switch m.Action {
		case model.ModerationApprove:
			st.ApproveCount++
		case model.ModerationReject:
			st.RejectCount++
		}
	}
	list := make([]*ModeratorStat, 0, len(byModerator))
	for _, st := range byModerator {
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool {
		ci := list[i].ApproveCount + list[i].RejectCount
		cj := list[j].ApproveCount + list[j].RejectCount
		if ci != cj {
			return ci > cj
		}
		return list[i].Moderator < list[j].Moderator
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}
