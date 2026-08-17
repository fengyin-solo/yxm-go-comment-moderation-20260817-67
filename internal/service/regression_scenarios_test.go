package service

import (
	"testing"

	"commentmoderation/internal/model"
)

func TestKeywordRuleNormalizesCaseAndSpaces(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "关键词归一化")
	if _, err := svc.CreateRule(model.Rule{
		Name:      "营销词",
		Type:      model.RuleKeyword,
		TextValue: " 广告 , 代购 ",
	}); err != nil {
		t.Fatal(err)
	}

	comment, err := svc.PostComment(content.ID, "u1", "这个AD太硬了，像广告")
	if err != nil {
		t.Fatal(err)
	}
	if comment.Status != model.CommentRejected || !comment.AutoFlag {
		t.Fatalf("大小写和空格归一化后应自动拒绝，得到 status=%s auto=%v", comment.Status, comment.AutoFlag)
	}
}

func TestBatchApproveOnlyCountsPersistedPendingComments(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "批量审核计数")
	pendingA, _ := svc.PostComment(content.ID, "u1", "第一条正常评论")
	pendingB, _ := svc.PostComment(content.ID, "u2", "第二条正常评论")
	otherContent := setupContent(t, svc, "其他内容")
	_, _ = svc.PostComment(otherContent.ID, "u3", "别的内容不应该被通过")

	count, err := svc.BatchApprovePending(content.ID, "mod-batch", "批量通过")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("只应通过目标内容的 2 条待审评论，得到 %d", count)
	}
	for _, id := range []string{pendingA.ID, pendingB.ID} {
		got, err := svc.GetComment(id)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != model.CommentApproved {
			t.Fatalf("评论 %s 应为 approved，得到 %s", id, got.Status)
		}
	}
	gotContent, _ := svc.GetContent(content.ID)
	if gotContent.CommentCount != 2 {
		t.Fatalf("内容评论计数应同步为 2，得到 %d", gotContent.CommentCount)
	}
}

func TestResolveReportDeletesApprovedAndRejectsPending(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "举报处置")
	approved, _ := svc.PostComment(content.ID, "u1", "先通过再举报")
	if _, err := svc.ApproveComment(approved.ID, "mod1", "正常"); err != nil {
		t.Fatal(err)
	}
	pending, _ := svc.PostComment(content.ID, "u2", "待审核时被举报")

	reportApproved, _ := svc.CreateReport(approved.ID, "r1", "违规")
	if _, err := svc.ResolveReport(reportApproved.ID, "handler1"); err != nil {
		t.Fatal(err)
	}
	gotApproved, _ := svc.GetComment(approved.ID)
	if gotApproved.Status != model.CommentDeleted {
		t.Fatalf("已发布评论举报成立后应删除，得到 %s", gotApproved.Status)
	}

	reportPending, _ := svc.CreateReport(pending.ID, "r2", "违规")
	if _, err := svc.ResolveReport(reportPending.ID, "handler2"); err != nil {
		t.Fatal(err)
	}
	gotPending, _ := svc.GetComment(pending.ID)
	if gotPending.Status != model.CommentRejected {
		t.Fatalf("待审评论举报成立后应拒绝，得到 %s", gotPending.Status)
	}
	gotContent, _ := svc.GetContent(content.ID)
	if gotContent.CommentCount != 0 {
		t.Fatalf("举报处置后已发布计数应归零，得到 %d", gotContent.CommentCount)
	}
}

func TestReporterCanFileAgainAfterPreviousReportClosed(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "重复举报生命周期")
	comment, _ := svc.PostComment(content.ID, "u1", "这是一条被举报的评论")

	first, err := svc.CreateReport(comment.ID, "r1", "疑似广告")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateReport(comment.ID, "r1", "重复提交"); err == nil {
		t.Fatal("同一用户的未处理重复举报应被拒绝")
	}
	if _, err := svc.DismissReport(first.ID, "handler1"); err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateReport(comment.ID, "r1", "再次发现问题")
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == first.ID || second.Status != model.ReportOpen {
		t.Fatalf("关闭上一条后应创建新的 open 举报，得到 %+v", second)
	}
}

func TestTopModeratorsSortsByManualWork(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "审核员排行")
	first, _ := svc.PostComment(content.ID, "u1", "第一条正常评论")
	second, _ := svc.PostComment(content.ID, "u2", "第二条正常评论")
	third, _ := svc.PostComment(content.ID, "u3", "第三条正常评论")
	_, _ = svc.ApproveComment(first.ID, "bob", "ok")
	_, _ = svc.RejectComment(second.ID, "alice", "bad")
	_, _ = svc.ApproveComment(third.ID, "alice", "ok")

	list := svc.TopModerators(2)
	if len(list) != 2 {
		t.Fatalf("期望 2 个审核员，得到 %d", len(list))
	}
	if list[0].Moderator != "alice" || list[0].ApproveCount != 1 || list[0].RejectCount != 1 {
		t.Fatalf("alice 应按 2 次人工审核排第一，得到 %+v", list[0])
	}
	if list[1].Moderator != "bob" || list[1].ApproveCount != 1 || list[1].RejectCount != 0 {
		t.Fatalf("bob 应按 1 次人工审核排第二，得到 %+v", list[1])
	}
}
