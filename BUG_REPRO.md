# Bug 是什么

关键词审核规则在规则文本带空格、评论文本大小写混用时不能稳定命中，应该被自动拒绝的评论会留在 pending。

# 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestKeywordRuleNormalizesCaseAndSpaces -count=20
```

# 错误信息

```text
大小写和空格归一化后应自动拒绝，得到 status=pending auto=false
```
