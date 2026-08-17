# Bug 是什么

已生效规范可被原地改写，基于旧 revision 的并发编辑也能覆盖最新草稿。

# 如何触发

对生效规范调用规则替换，并让两个副本按相同期望 revision 保存，执行 `go test ./tests -run TestEffectiveSpecificationAndRevisionCannotBeOverwritten -count=20`。

# 错误信息

`effective rules were rewritten in place: <nil>`
