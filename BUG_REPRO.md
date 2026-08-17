# Bug 是什么

人工审核员排行没有正确保留和统计人工拒绝动作，导致审核员的通过/拒绝次数以及排行顺序不符合实际工作量。

# 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestTopModeratorsSortsByManualWork -count=20
```

# 错误信息

```text
alice 应按 2 次人工审核排第一，得到 &{Moderator:alice ApproveCount:1 RejectCount:0}
```
