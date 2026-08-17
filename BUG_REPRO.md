# Bug 是什么

审计仓储接受引用错误 previous_hash 的分叉事件，并改变受控导出的链头结果。

# 如何触发

先追加合法首事件，再追加引用 stale-head 的第二事件，执行 `go test ./tests -run TestAuditTimelineRejectsForkedHashChain -count=20`。

# 错误信息

`forked audit event was accepted`
