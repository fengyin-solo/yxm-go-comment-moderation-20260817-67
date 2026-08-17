# Bug 是什么

批量审核目标内容的待审评论时，服务会跳过目标内容或没有正确持久化通过状态，导致返回数量和内容评论计数不一致。

# 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestBatchApproveOnlyCountsPersistedPendingComments -count=20
```

# 错误信息

```text
只应通过目标内容的 2 条待审评论，得到 1
```
