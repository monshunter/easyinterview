# Local Quality Gates Bootstrap Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-29

**关联计划**: [plan](./plan.md)

## Phase 1: 入口 target 聚合（lint / test / build / docs-check / codegen-check）

- [ ] 1.1 升级根 `Makefile` 的 `lint` target：串行调用 `lint-conventions` (B1) + `lint-config` (A4) + `lint-observability` (F1 placeholder) + `golangci-lint run ./...` (backend) + `pnpm --filter @easyinterview/frontend lint` (frontend)
- [ ] 1.2 升级 `test` target：`cd backend && go test ./...` + `pnpm --filter @easyinterview/frontend test`；AI 单测严格走 stub / fixtures，不读取 `AI_GATEWAY_*` 真实 secret
- [ ] 1.3 升级 `build` target：`cd backend && go build ./cmd/...` + `pnpm --filter @easyinterview/frontend build`；未落地 cmd 输出 `TODO: implemented by <owner>` 并 `exit 0`
- [ ] 1.4 新增 `docs-check` target：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + `python3 scripts/lint/check_md_links.py docs` 等价相对链接扫描；任一漂移 exit 非 0
- [ ] 1.5 升级 `codegen-check` target：`codegen-conventions` (B1) + `codegen-openapi` (B2) 顺序执行后 `git diff --exit-code` 守门；B1 diff 路径覆盖 `backend/internal/shared/types` / `backend/internal/shared/errors` / `backend/internal/shared/idx` / `frontend/src/lib/conventions` / `frontend/src/lib/ids` / `shared/conventions.yaml`

## Phase 2: 占位与缺位行为锁定（NOT-YET-LANDED owner 输出 + exit 0 边界）

- [ ] 2.1 在 Makefile 中实现「未落地 sub-target → 打印 `not implemented yet: <owner>` 并 `exit 0`」「已落地 sub-target → 失败穿透 exit code」双轨判断逻辑
- [ ] 2.2 登记当前 NOT-YET-LANDED 清单：`lint-observability`（F1）+ `lint-config`（A4，待 A4 001 落地后切换为直接执行）；其余（B1 / B2）为已落地直接执行
- [ ] 2.3 在 `make help` 中列出 5 个聚合入口与每个 sub-target owner，确保 owner subspec 改名时聚合层可被发现

## Phase 3: 文档与 CI deferral 守门

- [ ] 3.1 在根 `README.md` 与 `docs/development.md` 写入 5 个本地命令；明确声明当前不存在远端 CI pipeline，文档禁用「CI 已启用」「PR required check」措辞
- [ ] 3.2 登记 spec D-5 升级触发条件（第二位长期贡献者 / 公开 release / 付费用户 / 自动发版 / 回归频率过高），并指明升级路径为原地修订 + 新增 plan（默认 `002-remote-ci`），不改 subject 路径，不把远端 CI scope 回填到 001
- [ ] 3.3 申明 secret 红线：本地门禁不读取业务 secret；未来接入 CI 必须先递增 spec / history 登记 runner secret 字典与权限边界

## Phase 4: Verification

- [ ] 4.1 跑 make lint / make test / make build / make docs-check / make codegen-check 全部 exit 0（NOT-YET-LANDED owner 占位输出可见）
- [ ] 4.2 grep .github/workflows/ 不存在 ci.yml / nightly.yml / dependabot.yml
- [ ] 4.3 临时把 B1 codegen-conventions 输出文件人为修改一行：make codegen-check 报错并 exit 非 0；revert 后恢复 clean
- [ ] 4.4 临时让 B1 lint 引入一个 lower_snake_case 错误码：make lint 报错并 exit 非 0；revert 后恢复 clean
- [ ] 4.5 把本 plan Header 从 active 切到 completed，运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` 同步 plans/INDEX.md；再运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 确认无 drift
