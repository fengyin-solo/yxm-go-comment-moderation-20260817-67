# Bug 是什么

同一举报人在上一条举报已经关闭后，仍然无法对同一评论重新提交新的举报。

# 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestReporterCanFileAgainAfterPreviousReportClosed -count=20
```

# 错误信息

```text
report: 已有未处理的重复举报
```
