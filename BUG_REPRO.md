# Bug 是什么

整改在 submitted 状态即可关闭，旧 revision 的关闭还能覆盖最新复核结果。

# 如何触发

保存 submitted 整改后直接调用关闭，再用两个基于 revision 2 的副本竞争保存，执行 `go test ./tests -run TestRemediationNeedsVerificationAndRejectsStaleClosure -count=20`。

# 错误信息

`submitted remediation closed without independent verification`
