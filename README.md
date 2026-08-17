# 生产作业规范变更与合规校验平台

离线管理“规范起草 → 影响模拟 → 审批/定时生效 → 班组确认 → 现场校验/抽查 → 整改复核 → 版本替代或回滚”的可追溯闭环，防止车间继续沿用旧要求。

## 结构

- `internal/domain`：规范版本状态机、作用域、四类规则、不可豁免安全约束、确认、抽查和整改关闭条件。
- `internal/application`：冲突模拟、定时激活、例外审批、当前版本确认、合规评估与整改编排。
- `internal/repository`：并发安全内存仓储与 pgx/PostgreSQL 仓储，使用幂等键、唯一约束和乐观 revision。
- `internal/service`：本地时钟、ID 与可验证定时任务适配器；通知、文件和回调均不依赖外部服务。
- `internal/transport/http`、`internal/middleware`：`/api/v1`、角色权限、超时、request_id、结构化错误、安全头和 panic 恢复。
- `migrations`、`api/openapi`：可重复迁移、幂等演示数据、OpenAPI 3.0。
- `web`：Vue 3 + TypeScript + Vite + Pinia 的规范库、规则、影响、确认、抽查、整改、比较和审计页面。

## 启动

Go 1.24、Node 22、PostgreSQL 17；配置见 `.env.example`。

```sh
docker compose up -d postgres
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_init.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/002_seed.sql
go run ./cmd/server
cd web && npm ci && npm run dev
```

或 `docker compose up --build`。`/healthz` 检查进程，`/readyz` 检查数据库。

## 演示与状态规则

种子包含装配区焊接工序的护目镜强制安全规则。角色为 `author`、`approver`、`worker`、`inspector`、`reviewer`，通过 `X-Actor-ID`、`X-Actor-Role`、`X-Team-ID` 演示：创建新版本、冲突模拟、计划生效、个人确认、例外申请、现场抽查、整改提交/复核/关闭。

- `effective`/`replaced` 规则不可就地修改，必须 `NewVersion`；旧版保留替代关系和变更说明。
- 例外必须有未来到期时间、审批人和补偿控制；`MandatorySafety` 永远不能被例外绕过。
- 个人确认必须精确绑定当前生效版本，旧版本确认不计入待确认统计。
- 抽查必须带受控照片证据；问题按级别计算整改期限，只有提交证据并复核通过的整改才能关闭。
- 状态为 `draft → simulated → approved → scheduled → effective → replaced/rolled_back`，冲突或非法跳转返回稳定错误码。

所有列表按区域、工序、待确认、即将失效和逾期风险分页筛选，排序字段采用白名单。错误包含 `code`、`message`、`field_errors`、`request_id`；日志不记录附件正文、凭据或敏感字段。数据库种子使用 `ON CONFLICT DO NOTHING`，不覆盖用户数据。

## 验证

```sh
go build ./...
go test ./...
go test -race ./...
go vet ./...
cd web && npm ci && npm test && npm run build
```

2026-08-17 基线验证包含有效版本不可变、安全例外、组合校验、确认版本、整改关闭、冲突模拟、HTTP 权限和可选 PostgreSQL 集成测试。应用完全离线，不调用第三方 API、CDN、云存储、在线模型或真实消息服务。
