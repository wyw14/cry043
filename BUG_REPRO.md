# Bug 是什么

强制安全规则可被带补偿措施的例外批准，且对应 critical 违规被校验隐藏。

# 如何触发

为 lock-pin 强制安全规则申请并审批限时例外，再执行 `go test ./tests -run TestMandatorySafetyRuleSurvivesExceptionRoundTrip -count=20`。

# 错误信息

`mandatory safety exception was approved: <nil>`
