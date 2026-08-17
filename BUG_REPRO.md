# Bug 是什么

处理成立的举报时，举报状态流转和被举报评论处置没有保持一致，已发布评论、待审核评论和内容评论数都会出现错误结果。

# 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestResolveReportDeletesApprovedAndRejectsPending -count=20
```

# 错误信息

```text
status: 当前状态不可处理
```
