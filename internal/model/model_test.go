package model

import (
	"testing"
	"time"
)

func TestContentValidate(t *testing.T) {
	base := func() *Content {
		return &Content{Title: "标题", Author: "作者"}
	}
	cases := []struct {
		name    string
		mutate  func(*Content)
		wantErr bool
	}{
		{"合法", func(c *Content) {}, false},
		{"空标题", func(c *Content) { c.Title = "" }, true},
		{"标题过长", func(c *Content) { c.Title = string(make([]byte, 129)) }, true},
		{"空作者", func(c *Content) { c.Author = "" }, true},
		{"非法状态", func(c *Content) { c.Status = "weird" }, true},
	}
	for _, c := range cases {
		content := base()
		c.mutate(content)
		err := content.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestContentFilterMatch(t *testing.T) {
	c := &Content{Title: "Go 并发编程", Author: "张三", Category: "tech", Status: ContentOpen}
	if !(ContentFilter{}).Match(c) {
		t.Error("空筛选应匹配所有")
	}
	if !(ContentFilter{Category: "tech", Status: ContentOpen}).Match(c) {
		t.Error("类别+状态匹配应通过")
	}
	if (ContentFilter{Category: "life"}).Match(c) {
		t.Error("类别不匹配应失败")
	}
	if !(ContentFilter{Keyword: "并发"}).Match(c) {
		t.Error("关键词匹配标题应通过")
	}
	if !(ContentFilter{Keyword: "张三"}).Match(c) {
		t.Error("关键词匹配作者应通过")
	}
	if (ContentFilter{Keyword: "rust"}).Match(c) {
		t.Error("关键词不匹配应失败")
	}
}

func TestCommentValidateAndTransition(t *testing.T) {
	base := func() *Comment {
		return &Comment{ContentID: "c1", UserID: "u1", Body: "内容"}
	}
	cases := []struct {
		name    string
		mutate  func(*Comment)
		wantErr bool
	}{
		{"合法", func(c *Comment) {}, false},
		{"空内容ID", func(c *Comment) { c.ContentID = "" }, true},
		{"空用户", func(c *Comment) { c.UserID = "" }, true},
		{"空正文", func(c *Comment) { c.Body = "" }, true},
		{"正文过长", func(c *Comment) { c.Body = string(make([]byte, 1001)) }, true},
		{"非法状态", func(c *Comment) { c.Status = "weird" }, true},
	}
	for _, c := range cases {
		comment := base()
		c.mutate(comment)
		err := comment.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	transitions := []struct {
		from, to string
		want     bool
	}{
		{CommentPending, CommentApproved, true},
		{CommentPending, CommentRejected, true},
		{CommentPending, CommentDeleted, false},
		{CommentApproved, CommentDeleted, true},
		{CommentApproved, CommentRejected, false},
		{CommentRejected, CommentApproved, false},
		{CommentDeleted, CommentApproved, false},
	}
	for _, c := range transitions {
		if got := CanCommentTransition(c.from, c.to); got != c.want {
			t.Errorf("CanCommentTransition(%s,%s)=%v，期望 %v", c.from, c.to, got, c.want)
		}
	}
}

func TestCommentFilterMatch(t *testing.T) {
	c := &Comment{ContentID: "c1", UserID: "u1", Status: CommentPending}
	if !(CommentFilter{}).Match(c) {
		t.Error("空筛选应匹配所有")
	}
	if !(CommentFilter{ContentID: "c1", UserID: "u1", Status: CommentPending}).Match(c) {
		t.Error("全条件匹配应通过")
	}
	if (CommentFilter{Status: CommentApproved}).Match(c) {
		t.Error("状态不匹配应失败")
	}
	if (CommentFilter{UserID: "u2"}).Match(c) {
		t.Error("用户不匹配应失败")
	}
}

func TestModerationRecordValidate(t *testing.T) {
	base := func() *ModerationRecord {
		return &ModerationRecord{
			CommentID: "c1", Moderator: "mod1",
			Action: ModerationApprove, Source: ModerationSourceManual, CreatedAt: time.Now(),
		}
	}
	cases := []struct {
		name    string
		mutate  func(*ModerationRecord)
		wantErr bool
	}{
		{"合法", func(m *ModerationRecord) {}, false},
		{"空评论ID", func(m *ModerationRecord) { m.CommentID = "" }, true},
		{"空审核人", func(m *ModerationRecord) { m.Moderator = "" }, true},
		{"非法动作", func(m *ModerationRecord) { m.Action = "weird" }, true},
		{"非法来源", func(m *ModerationRecord) { m.Source = "weird" }, true},
		{"理由过长", func(m *ModerationRecord) { m.Reason = string(make([]byte, 201)) }, true},
		{"零值时间", func(m *ModerationRecord) { m.CreatedAt = time.Time{} }, true},
	}
	for _, c := range cases {
		m := base()
		c.mutate(m)
		err := m.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestRuleValidateAndHelpers(t *testing.T) {
	cases := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{"关键词规则合法", Rule{Name: "r", Type: RuleKeyword, TextValue: "a"}, false},
		{"关键词为空", Rule{Name: "r", Type: RuleKeyword}, true},
		{"长度下限合法", Rule{Name: "r", Type: RuleMinLength, LimitValue: 5}, false},
		{"长度下限为0", Rule{Name: "r", Type: RuleMinLength, LimitValue: 0}, true},
		{"限流合法", Rule{Name: "r", Type: RuleMaxPerUser, LimitValue: 3}, false},
		{"限流为0", Rule{Name: "r", Type: RuleMaxPerUser, LimitValue: 0}, true},
		{"非法类型", Rule{Name: "r", Type: "weird"}, true},
		{"空名称", Rule{Type: RuleKeyword, TextValue: "a"}, true},
		{"非法状态", Rule{Name: "r", Type: RuleKeyword, TextValue: "a", Status: "weird"}, true},
	}
	for _, c := range cases {
		r := c.rule
		err := r.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	// AppliesTo：全局规则匹配所有内容，定向规则只匹配指定内容
	global := &Rule{Status: RuleActive, ContentID: ""}
	if !global.AppliesTo("any") {
		t.Error("全局规则应作用于所有内容")
	}
	scoped := &Rule{Status: RuleActive, ContentID: "c1"}
	if !scoped.AppliesTo("c1") || scoped.AppliesTo("c2") {
		t.Error("定向规则作用域不对")
	}
	inactive := &Rule{Status: RuleInactive, ContentID: ""}
	if inactive.AppliesTo("any") {
		t.Error("停用规则不应生效")
	}
	// Keywords 拆分与清洗
	r := &Rule{TextValue: " A , b ,, c "}
	kws := r.Keywords()
	if len(kws) != 3 || kws[0] != "a" || kws[1] != "b" || kws[2] != "c" {
		t.Errorf("关键词拆分清洗不对: %v", kws)
	}
}

func TestReportValidateAndTransition(t *testing.T) {
	base := func() *Report {
		return &Report{CommentID: "c1", ReporterID: "u1", Reason: "理由"}
	}
	cases := []struct {
		name    string
		mutate  func(*Report)
		wantErr bool
	}{
		{"合法", func(r *Report) {}, false},
		{"空评论ID", func(r *Report) { r.CommentID = "" }, true},
		{"空举报人", func(r *Report) { r.ReporterID = "" }, true},
		{"空理由", func(r *Report) { r.Reason = "" }, true},
		{"理由过长", func(r *Report) { r.Reason = string(make([]byte, 501)) }, true},
		{"非法状态", func(r *Report) { r.Status = "weird" }, true},
	}
	for _, c := range cases {
		r := base()
		c.mutate(r)
		err := r.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	transitions := []struct {
		from, to string
		want     bool
	}{
		{ReportOpen, ReportResolved, true},
		{ReportOpen, ReportDismissed, true},
		{ReportResolved, ReportDismissed, false},
		{ReportDismissed, ReportResolved, false},
	}
	for _, c := range transitions {
		if got := CanReportTransition(c.from, c.to); got != c.want {
			t.Errorf("CanReportTransition(%s,%s)=%v，期望 %v", c.from, c.to, got, c.want)
		}
	}
}

func TestReportFilterMatch(t *testing.T) {
	r := &Report{CommentID: "c1", Status: ReportOpen}
	if !(ReportFilter{}).Match(r) {
		t.Error("空筛选应匹配所有")
	}
	if !(ReportFilter{CommentID: "c1", Status: ReportOpen}).Match(r) {
		t.Error("全条件匹配应通过")
	}
	if (ReportFilter{Status: ReportResolved}).Match(r) {
		t.Error("状态不匹配应失败")
	}
}
