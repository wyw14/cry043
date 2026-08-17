# Bug 是什么

风险驾驶舱把同区域其他工序、旧版本确认和已过期例外混入当前筛选结果。

# 如何触发

在同时存在装配区焊接与喷涂规范时，给焊接班组写入旧版本确认和过期例外，再执行 `go test ./tests -run TestRiskBoardKeepsProcessAndVersionBoundaries -count=20`。

# 错误信息

`scope/version isolation failed: PendingConfirmations:1 ExpiringExceptions:1 ConfirmationCoverage:0.6666666666666666`
