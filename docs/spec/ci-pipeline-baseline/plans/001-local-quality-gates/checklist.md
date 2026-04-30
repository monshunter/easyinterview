# Local Quality Gates Bootstrap Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-04-30

**关联计划**: [plan](./plan.md)

## Phase 1: 入口 target 聚合（lint / test / build / docs-check / codegen-check）

- [x] 1.1 升级根 `Makefile` 的 `lint` target：串行调用 `lint-conventions` (B1) + `lint-config` (A4) + `lint-observability` (F1 placeholder) + `golangci-lint run ./...` (backend) + `pnpm --filter @easyinterview/frontend lint` (frontend)；验证: 2026-04-30 `make lint` exit 0，输出 `not implemented yet: F1 observability-stack`、backend `golangci-lint run ./...` 报 `0 issues.`，frontend lint 进入 `pnpm --filter @easyinterview/frontend lint`；已落地 sub-target 失败不会被 placeholder 吞没
- [x] 1.2 升级 `test` target：`cd backend && go test ./...` + `pnpm --filter @easyinterview/frontend test`；AI 单测严格走 stub / fixtures，不读取 `AI_GATEWAY_*` 真实 secret；验证: 2026-04-30 `make test` exit 0，顺序运行 backend `go test ./...` 与 frontend `vitest run`（10 files / 49 tests）；`grep -rn 'os.Getenv("AI_GATEWAY_' backend/internal | grep -v '_test.go\|stub\|fixture\|mock'` 不命中
- [x] 1.3 升级 `build` target：`cd backend && go build ./cmd/...` + `pnpm --filter @easyinterview/frontend build`；未落地 cmd 输出 `TODO: implemented by <owner>` 并 `exit 0`；验证: 2026-04-30 `make build` exit 0，backend 执行 `go build ./cmd/...`，frontend 进入 `pnpm --filter @easyinterview/frontend build` 并输出 D1-owned placeholder `TODO: build implemented by D1 frontend-shell`
- [x] 1.4 新增 `docs-check` target：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + `python3 scripts/lint/check_md_links.py docs` 等价相对链接扫描；任一漂移 exit 非 0；验证: 临时在任一 spec Header 插入 drift 或在 docs/ md 中插入失链相对链接后 `make docs-check` exit 非 0；revert 后 exit 0
- [x] 1.5 升级 `codegen-check` target：`codegen-conventions` (B1) + `codegen-openapi` (B2) 顺序执行后 `git diff --exit-code` 守门；B1 diff 路径覆盖 `backend/internal/shared/types` / `backend/internal/shared/errors` / `backend/internal/shared/idx` / `frontend/src/lib/conventions` / `frontend/src/lib/ids` / `shared/conventions.yaml`；验证: `make codegen-check` 顺序执行 codegen-conventions / codegen-openapi 后 `git diff --exit-code` 覆盖 6 条 B1 路径 + 3 条 B2 路径（`backend/internal/api/generated` / `frontend/src/api/generated` / `openapi/openapi.yaml`）；任一漂移 exit 非 0

## Phase 2: 占位与缺位行为锁定（NOT-YET-LANDED owner 输出 + exit 0 边界）

- [x] 2.1 在 Makefile 中实现「未落地 sub-target → 打印 `not implemented yet: <owner>` 并 `exit 0`」「已落地 sub-target → 失败穿透 exit code」双轨判断逻辑；验证: 缺位 sub-target 时单行 `not implemented yet: <owner>` exit 0，已落地 sub-target 人工注入返回 1 时整体 `make` 必须 exit 非 0（占位逻辑不得吞没原始 exit code）
- [x] 2.2 登记当前 NOT-YET-LANDED 清单：仅 `lint-observability`（F1）为占位；`lint-config`（A4）已于 2026-04-30 落地切换为直接执行（`lint-getenv-boundary` + `lint-env-dict` + `lint-secrets-pattern`），其余（B1 / B2）已落地直接执行；验证: `make -n lint-observability` 输出 `not implemented yet: F1 observability-stack`；`make -n lint-config` / `make -n lint-conventions` / `make -n codegen-conventions` / `make -n codegen-openapi` 均直接进入对应 owner 实现
- [x] 2.3 在 `make help` 中列出 5 个聚合入口与每个 sub-target owner，确保 owner subspec 改名时聚合层可被发现；验证: 2026-04-30 `make help | grep -E ...` 分别命中 `lint-conventions (B1)` / `lint-config (A4)` / `lint-observability (F1)` / `codegen-conventions (B1)` / `codegen-openapi (B2)` 与 5 个聚合入口名

## Phase 3: 文档与 CI deferral 守门

- [x] 3.1 在根 `README.md` 与 `docs/development.md` 写入 5 个本地命令；明确声明当前不存在远端 CI pipeline，文档禁用「CI 已启用」「PR required check」措辞；验证: `grep -nE '(CI 已启用\|PR required check)' README.md docs/development.md` 不命中（除 deferred / out of scope 段落以否定语境出现）；5 个 `make` 命令名出现在 onboarding 文档
- [x] 3.2 登记 spec D-5 升级触发条件（第二位长期贡献者 / 公开 release / 付费用户 / 自动发版 / 回归频率过高），并指明升级路径为原地修订 + 新增 plan（默认 `002-remote-ci`），不改 subject 路径，不把远端 CI scope 回填到 001；验证: 本 plan 同级文档 grep 命中 D-5 五条触发条件全部 + 升级路径 `002-remote-ci`；与 spec §3.1 D-5 锁定值一致
- [x] 3.3 申明 secret 红线：本地门禁不读取业务 secret；未来接入 CI 必须先递增 spec / history 登记 runner secret 字典与权限边界；验证: 本 plan 文档出现 secret 红线段落；2026-04-30 `grep -nE '(AI_GATEWAY_\|DB_\|REDIS_\|POSTHOG_)' Makefile scripts/lint/*.py` 不命中；`make -n lint` / `make -n test` / `make -n build` 不引用 `.env` 业务 secret

## Phase 4: Verification

- [x] 4.1 跑 make lint / make test / make build / make docs-check / make codegen-check 全部 exit 0（NOT-YET-LANDED owner 占位输出可见）；验证: 2026-04-30 顺序执行 `make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check` 全部 exit 0；占位输出包括 `not implemented yet: F1 observability-stack`（lint）与 D1 frontend lint/build placeholder；backend lint/test/build 均真实执行并通过
- [x] 4.2 grep .github/workflows/ 不存在 ci.yml / nightly.yml / dependabot.yml；验证: `find .github/workflows -type f -name '*.yml' 2>/dev/null` 不命中（`.github/` 目录本身不存在）；`grep -r 'ci-pipeline-baseline' .github 2>/dev/null` 无命中
- [x] 4.3 临时把 B1 codegen-conventions 输出文件人为修改一行：make codegen-check 报错并 exit 非 0；revert 后恢复 clean；验证: 2026-04-30 在 `backend/internal/shared/types/enums.go` 第 4 行注入 `// INJECTED-DRIFT-FOR-A5-PHASE-4.3-VERIFICATION`，`make codegen-check` exit 2 并打印 `FAIL: enum drift: shared/conventions.yaml -> backend/internal/shared/types/enums.go differs; run make codegen-conventions`；revert 后再跑 exit 0
- [x] 4.4 临时让 B1 lint 引入一个 lower_snake_case 错误码：make lint 报错并 exit 非 0；revert 后恢复 clean；验证: 2026-04-30 把 `shared/conventions.yaml` 第 46 行的 `AUTH_UNAUTHORIZED` 改为 `auth_unauthorized_lower`，`make lint` exit 2 并打印 `FAIL: error code must be UPPER_SNAKE_CASE, got 'auth_unauthorized_lower'`；revert 后再跑 exit 0
- [x] 4.5 把本 plan Header 从 active 切到 completed，运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` 同步 plans/INDEX.md；再运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 确认无 drift；验证: 2026-04-30 plan / checklist Header 切到 `completed` v1.3；`docs/spec/ci-pipeline-baseline/plans/INDEX.md` 状态投影同步；`make docs-check` exit 0
