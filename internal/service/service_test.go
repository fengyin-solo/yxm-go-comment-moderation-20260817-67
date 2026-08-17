package service

import (
	"testing"

	"commentmoderation/internal/config"
	"commentmoderation/internal/model"
	"commentmoderation/internal/store"
	"commentmoderation/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

// setupContent 创建一个开放评论的内容，返回内容。
func setupContent(t *testing.T, svc *Service, title string) *model.Content {
	t.Helper()
	c, err := svc.CreateContent(model.Content{Title: title, Author: "作者", Category: "tech"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCreateContentValidation(t *testing.T) {
	svc := newTestService()
	cases := []struct {
		name  string
		input model.Content
	}{
		{"空标题", model.Content{Author: "a"}},
		{"空作者", model.Content{Title: "t"}},
		{"非法状态", model.Content{Title: "t", Author: "a", Status: "weird"}},
	}
	for _, c := range cases {
		if _, err := svc.CreateContent(c.input); err == nil {
			t.Errorf("%s: 期望校验失败但成功了", c.name)
		} else if !model.IsValidationError(err) {
			t.Errorf("%s: 期望 ValidationError，得到 %v", c.name, err)
		}
	}
	// 空类别应回落 general
	c, err := svc.CreateContent(model.Content{Title: "t", Author: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Category != "general" {
		t.Errorf("空类别应回落 general，得到 %s", c.Category)
	}
}

func TestContentUpdateAndDeleteGuard(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "删除保护")
	if _, err := svc.PostComment(content.ID, "u1", "这是一条正常评论"); err != nil {
		t.Fatal(err)
	}
	// 有未删除评论不能删内容
	if err := svc.DeleteContent(content.ID); err == nil {
		t.Fatal("有评论的内容删除应被拒绝")
	}
	// 更新状态
	updated, err := svc.UpdateContent(content.ID, "", "", model.ContentClosed)
	if err != nil || updated.Status != model.ContentClosed {
		t.Fatalf("更新失败: %v", err)
	}
	// 关闭评论区后不能发评论
	if _, err := svc.PostComment(content.ID, "u1", "再来一条"); err == nil {
		t.Fatal("关闭评论区后发评论应被拒绝")
	}
}

func TestPostCommentValidation(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "校验")
	if _, err := svc.PostComment("missing", "u1", "x"); err == nil {
		t.Fatal("内容不存在应报错")
	}
	if _, err := svc.PostComment(content.ID, "", "x"); err == nil {
		t.Fatal("空用户应被拒绝")
	}
	if _, err := svc.PostComment(content.ID, "u1", ""); err == nil {
		t.Fatal("空内容应被拒绝")
	}
}

func TestKeywordRuleAutoReject(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "关键词")
	if _, err := svc.CreateRule(model.Rule{
		Name: "敏感词", Type: model.RuleKeyword, TextValue: "广告,代购",
	}); err != nil {
		t.Fatal(err)
	}
	// 命中关键词 -> 自动拒绝
	c, err := svc.PostComment(content.ID, "u1", "加我微信做广告")
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != model.CommentRejected || !c.AutoFlag {
		t.Fatalf("命中关键词应自动拒绝，得到 %s auto=%v", c.Status, c.AutoFlag)
	}
	// 未命中 -> 待审核
	c2, err := svc.PostComment(content.ID, "u1", "写得真好")
	if err != nil {
		t.Fatal(err)
	}
	if c2.Status != model.CommentPending || c2.AutoFlag {
		t.Fatalf("未命中应待审核，得到 %s auto=%v", c2.Status, c2.AutoFlag)
	}
	// 自动拒绝应产生 auto 审核记录
	records, total, err := svc.ListModerationRecords(model.ModerationFilter{Source: model.ModerationSourceAuto}, 1, 10)
	if err != nil || total != 1 || records[0].CommentID != c.ID {
		t.Fatalf("期望 1 条自动审核记录，得到 total=%d err=%v", total, err)
	}
}

func TestMinLengthRuleAutoReject(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "长度")
	if _, err := svc.CreateRule(model.Rule{
		Name: "最短 5 字", Type: model.RuleMinLength, LimitValue: 5,
	}); err != nil {
		t.Fatal(err)
	}
	c, err := svc.PostComment(content.ID, "u1", "好")
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != model.CommentRejected || !c.AutoFlag {
		t.Fatalf("过短评论应自动拒绝，得到 %s", c.Status)
	}
	c2, err := svc.PostComment(content.ID, "u1", "这条评论足够长了")
	if err != nil {
		t.Fatal(err)
	}
	if c2.Status != model.CommentPending {
		t.Fatalf("正常长度应待审核，得到 %s", c2.Status)
	}
}

func TestMaxPerUserRuleAutoReject(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "限流")
	if _, err := svc.CreateRule(model.Rule{
		Name: "每人限 2 条", Type: model.RuleMaxPerUser, LimitValue: 2,
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		c, err := svc.PostComment(content.ID, "u1", "这是第几条正常评论内容啊")
		if err != nil {
			t.Fatal(err)
		}
		if c.Status != model.CommentPending {
			t.Fatalf("前 2 条应待审核，得到 %s", c.Status)
		}
	}
	c, err := svc.PostComment(content.ID, "u1", "这是第三条正常评论内容啊")
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != model.CommentRejected || !c.AutoFlag {
		t.Fatalf("超过上限应自动拒绝，得到 %s", c.Status)
	}
	// 其他用户不受影响
	c2, err := svc.PostComment(content.ID, "u2", "这是另一个用户的正常评论")
	if err != nil {
		t.Fatal(err)
	}
	if c2.Status != model.CommentPending {
		t.Fatalf("其他用户应待审核，得到 %s", c2.Status)
	}
}

func TestInactiveRuleNotApplied(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "停用规则")
	rule, err := svc.CreateRule(model.Rule{
		Name: "敏感词", Type: model.RuleKeyword, TextValue: "广告",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRule(rule.ID, "", model.RuleInactive, "", 0); err != nil {
		t.Fatal(err)
	}
	c, err := svc.PostComment(content.ID, "u1", "广告内容")
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != model.CommentPending {
		t.Fatalf("停用规则不应生效，得到 %s", c.Status)
	}
}

func TestContentScopedRule(t *testing.T) {
	svc := newTestService()
	c1 := setupContent(t, svc, "内容一")
	c2 := setupContent(t, svc, "内容二")
	// 规则只作用于内容一
	if _, err := svc.CreateRule(model.Rule{
		Name: "仅内容一", Type: model.RuleKeyword, ContentID: c1.ID, TextValue: "广告",
	}); err != nil {
		t.Fatal(err)
	}
	got1, _ := svc.PostComment(c1.ID, "u1", "广告")
	if got1.Status != model.CommentRejected {
		t.Fatalf("内容一应命中规则，得到 %s", got1.Status)
	}
	got2, _ := svc.PostComment(c2.ID, "u1", "广告")
	if got2.Status != model.CommentPending {
		t.Fatalf("内容二不应命中规则，得到 %s", got2.Status)
	}
	// 关联不存在内容的规则应被拒绝
	if _, err := svc.CreateRule(model.Rule{
		Name: "x", Type: model.RuleKeyword, ContentID: "missing", TextValue: "a",
	}); err == nil {
		t.Fatal("关联不存在内容的规则应被拒绝")
	}
}

func TestManualModerationFlow(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "人工审核")
	c, err := svc.PostComment(content.ID, "u1", "一条待审核的正常评论")
	if err != nil {
		t.Fatal(err)
	}
	// 通过
	approved, err := svc.ApproveComment(c.ID, "mod1", "内容合规")
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != model.CommentApproved {
		t.Fatalf("通过后应为 approved，得到 %s", approved.Status)
	}
	// 内容评论计数 +1
	got, _ := svc.GetContent(content.ID)
	if got.CommentCount != 1 {
		t.Fatalf("评论计数应为 1，得到 %d", got.CommentCount)
	}
	// 已通过的不能再通过
	if _, err := svc.ApproveComment(c.ID, "mod1", "再来"); err == nil {
		t.Fatal("已通过评论再通过应被拒绝")
	}
	// 删除已发布评论，计数 -1
	if err := svc.DeleteComment(c.ID); err != nil {
		t.Fatal(err)
	}
	got, _ = svc.GetContent(content.ID)
	if got.CommentCount != 0 {
		t.Fatalf("删除后评论计数应为 0，得到 %d", got.CommentCount)
	}
}

func TestManualRejectFlow(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "人工拒绝")
	c, _ := svc.PostComment(content.ID, "u1", "一条待审核的正常评论")
	// 拒绝必须填理由
	if _, err := svc.RejectComment(c.ID, "mod1", ""); err == nil {
		t.Fatal("拒绝不填理由应被拒绝")
	}
	rejected, err := svc.RejectComment(c.ID, "mod1", "违规")
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Status != model.CommentRejected {
		t.Fatalf("拒绝后应为 rejected，得到 %s", rejected.Status)
	}
	// 已拒绝的不能删除（仅 approved 可删）
	if err := svc.DeleteComment(c.ID); err == nil {
		t.Fatal("已拒绝评论删除应被拒绝")
	}
}

func TestReportFlow(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "举报")
	c, _ := svc.PostComment(content.ID, "u1", "一条待审核的正常评论")
	// 举报
	report, err := svc.CreateReport(c.ID, "u9", "垃圾广告")
	if err != nil {
		t.Fatal(err)
	}
	// 重复举报（未处理）应被拒绝
	if _, err := svc.CreateReport(c.ID, "u9", "再次举报"); err == nil {
		t.Fatal("重复未处理举报应被拒绝")
	}
	// 举报不存在的评论
	if _, err := svc.CreateReport("missing", "u9", "x"); err == nil {
		t.Fatal("举报不存在评论应报错")
	}
	// 举报成立 -> 待审核评论被拒绝
	resolved, err := svc.ResolveReport(report.ID, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status != model.ReportResolved || resolved.HandledBy != "admin" {
		t.Fatalf("举报处理结果不对: %+v", resolved)
	}
	got, _ := svc.GetComment(c.ID)
	if got.Status != model.CommentRejected {
		t.Fatalf("举报成立后待审核评论应被拒绝，得到 %s", got.Status)
	}
	// 已处理的举报不能再处理
	if _, err := svc.ResolveReport(report.ID, "admin"); err == nil {
		t.Fatal("已处理举报再处理应被拒绝")
	}
}

func TestReportDismissAndApprovedComment(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "驳回与删除")
	c, _ := svc.PostComment(content.ID, "u1", "一条待审核的正常评论")
	_ , _ = svc.ApproveComment(c.ID, "mod1", "ok")
	report, err := svc.CreateReport(c.ID, "u9", "引战")
	if err != nil {
		t.Fatal(err)
	}
	// 驳回
	dismissed, err := svc.DismissReport(report.ID, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if dismissed.Status != model.ReportDismissed {
		t.Fatalf("驳回后应为 dismissed，得到 %s", dismissed.Status)
	}
	// 评论不受影响
	got, _ := svc.GetComment(c.ID)
	if got.Status != model.CommentApproved {
		t.Fatalf("驳回举报后评论应保持 approved，得到 %s", got.Status)
	}
	// 再举报并成立 -> 已发布评论被删除
	report2, _ := svc.CreateReport(c.ID, "u8", "确实引战")
	if _, err := svc.ResolveReport(report2.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	got, _ = svc.GetComment(c.ID)
	if got.Status != model.CommentDeleted {
		t.Fatalf("举报成立后已发布评论应被删除，得到 %s", got.Status)
	}
}

func TestStatsOverviewAndGrouping(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "统计")
	if _, err := svc.CreateRule(model.Rule{
		Name: "敏感词", Type: model.RuleKeyword, TextValue: "广告",
	}); err != nil {
		t.Fatal(err)
	}
	// 1 条自动拒绝
	svc.PostComment(content.ID, "u1", "广告")
	// 2 条待审核
	c2, _ := svc.PostComment(content.ID, "u1", "第一条正常评论内容")
	c3, _ := svc.PostComment(content.ID, "u2", "第二条正常评论内容")
	// 1 条人工通过
	svc.ApproveComment(c2.ID, "mod1", "ok")
	// 1 条人工拒绝
	svc.RejectComment(c3.ID, "mod1", "不行")

	overview := svc.Stats()
	if overview.CommentCount != 3 || overview.PendingCount != 0 {
		t.Fatalf("概览数量不对: %+v", overview)
	}
	if overview.ApprovedCount != 1 || overview.RejectedCount != 2 {
		t.Fatalf("通过/拒绝计数不对: %+v", overview)
	}
	if overview.AutoRejected != 1 {
		t.Fatalf("自动拒绝应为 1，得到 %d", overview.AutoRejected)
	}
	// 自动率：2 拒绝中 1 自动 = 50%
	if overview.AutoRate != 50 {
		t.Fatalf("自动率期望 50%%，得到 %d", overview.AutoRate)
	}

	byContent := svc.StatsByContent()
	if len(byContent) != 1 || byContent[0].CommentCount != 1 || byContent[0].RejectedCount != 2 {
		t.Fatalf("按内容统计不对: %+v", byContent)
	}

	top := svc.TopModerators(10)
	if len(top) != 1 || top[0].Moderator != "mod1" || top[0].ApproveCount != 1 || top[0].RejectCount != 1 {
		t.Fatalf("审核员排行不对: %+v", top)
	}
}

func TestCommentListFilterAndPagination(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "筛选")
	for i := 0; i < 5; i++ {
		if _, err := svc.PostComment(content.ID, "u1", "这是第 N 条足够长的正常评论"); err != nil {
			t.Fatal(err)
		}
	}
	items, total, err := svc.ListComments(model.CommentFilter{Status: model.CommentPending}, 1, 10)
	if err != nil || total != 5 || len(items) != 5 {
		t.Fatalf("状态筛选失败: total=%d err=%v", total, err)
	}
	items, total, err = svc.ListComments(model.CommentFilter{UserID: "u1", ContentID: content.ID}, 2, 3)
	if err != nil || total != 5 || len(items) != 2 {
		t.Fatalf("分页期望 total=5 len=2，得到 total=%d len=%d", total, len(items))
	}
	_, total, err = svc.ListComments(model.CommentFilter{UserID: "nobody"}, 1, 10)
	if err != nil || total != 0 {
		t.Fatalf("无匹配应返回 0，得到 %d", total)
	}
}

func TestBatchApprovePending(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "批量审核")
	for i := 0; i < 3; i++ {
		if _, err := svc.PostComment(content.ID, "u1", "这是第 N 条足够长的正常评论"); err != nil {
			t.Fatal(err)
		}
	}
	// 内容不存在
	if _, err := svc.BatchApprovePending("missing", "mod1", "ok"); err == nil {
		t.Fatal("内容不存在应报错")
	}
	// 审核人为空
	if _, err := svc.BatchApprovePending(content.ID, "", "ok"); err == nil {
		t.Fatal("审核人为空应被拒绝")
	}
	count, err := svc.BatchApprovePending(content.ID, "mod1", "批量通过")
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("期望通过 3 条，得到 %d", count)
	}
	got, _ := svc.GetContent(content.ID)
	if got.CommentCount != 3 {
		t.Fatalf("评论计数应为 3，得到 %d", got.CommentCount)
	}
	// 再次批量应无可处理项
	count, err = svc.BatchApprovePending(content.ID, "mod1", "再来")
	if err != nil || count != 0 {
		t.Fatalf("二次批量期望 0，得到 %d err=%v", count, err)
	}
}

func TestRuleCRUDAndValidation(t *testing.T) {
	svc := newTestService()
	content := setupContent(t, svc, "规则CRUD")
	rule, err := svc.CreateRule(model.Rule{
		Name: "关键词", Type: model.RuleKeyword, ContentID: content.ID, TextValue: "a,b",
	})
	if err != nil {
		t.Fatal(err)
	}
	// 校验失败用例
	if _, err := svc.CreateRule(model.Rule{Name: "", Type: model.RuleKeyword, TextValue: "x"}); err == nil {
		t.Fatal("空名称应被拒绝")
	}
	if _, err := svc.CreateRule(model.Rule{Name: "x", Type: "weird"}); err == nil {
		t.Fatal("非法类型应被拒绝")
	}
	if _, err := svc.CreateRule(model.Rule{Name: "x", Type: model.RuleMinLength, LimitValue: 0}); err == nil {
		t.Fatal("长度下限为 0 应被拒绝")
	}
	// 关键词拆分
	kws := rule.Keywords()
	if len(kws) != 2 || kws[0] != "a" || kws[1] != "b" {
		t.Fatalf("关键词拆分不对: %v", kws)
	}
	// 更新
	updated, err := svc.UpdateRule(rule.ID, "", model.RuleInactive, "", 0)
	if err != nil || updated.Status != model.RuleInactive {
		t.Fatalf("更新规则失败: %v", err)
	}
	// 筛选
	items, total, err := svc.ListRules(model.RuleFilter{Type: model.RuleKeyword}, 1, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("规则筛选失败: total=%d err=%v", total, err)
	}
	if err := svc.DeleteRule(rule.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteRule(rule.ID); err == nil {
		t.Fatal("重复删除应报错")
	}
}
