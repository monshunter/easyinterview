# Local Quality Gates Bootstrap Checklist

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: 入口 target 聚合（lint / test / build / docs-check / codegen-check）

- [x] 1.1 升级根 `Makefile` 的 `lint` target：串行调用 `lint-conventions` (B1) + `lint-config` (A4) + A3/F3/E1/runtime-topology local gates + `golangci-lint run ./...` (backend) + `pnpm --filter @easyinterview/frontend lint` (frontend)；验证: 2026-04-30 `make lint` exit 0，backend `golangci-lint run ./...` 报 `0 issues.`、frontend lint 进入 `pnpm --filter @easyinterview/frontend lint`；2026-07-10 frontend lint 改为 `tsc --noEmit`，不再 `exit 0` 假通过；2026-07-10 F1 未落地 helper 的 exit-zero target 已删除，聚合层只调用已落地 sub-target
- [x] 1.2 升级 `test` target：`cd backend && go test ./...` + `pnpm --filter @easyinterview/frontend test`；AI 单测严格走 stub / fixtures，不读取 `AI_PROVIDER_*` 真实 secret；验证: 2026-04-30 `make test` exit 0，顺序运行 backend `go test ./...` 与 frontend `vitest run`（10 files / 49 tests）；`grep -rn 'os.Getenv("AI_GATEWAY_' backend/internal | grep -v '_test.go\|stub\|fixture\|mock'` 不命中
- [x] 1.3 升级 `build` target：`cd backend && go build ./cmd/...` + `pnpm --filter @easyinterview/frontend build`；验证: 2026-04-30 `make build` exit 0，backend 执行 `go build ./cmd/...`；2026-07-10 frontend build 已是真实 `tsc --noEmit && vite build`
- [x] 1.4 新增 `docs-check` target：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + `python3 scripts/lint/check_md_links.py docs` 等价相对链接扫描；任一漂移 exit 非 0；验证: 临时在任一 spec Header 插入 drift 或在 docs/ md 中插入失链相对链接后 `make docs-check` exit 非 0；revert 后 exit 0
- [x] 1.5 升级 `codegen-check` target：`codegen-conventions` (B1) + `codegen-openapi` (B2) 顺序执行后 `git diff --exit-code` 守门；B1 diff 路径覆盖 `backend/internal/shared/types` / `backend/internal/shared/errors` / `backend/internal/shared/idx` / `frontend/src/lib/conventions` / `frontend/src/lib/ids` / `shared/conventions.yaml`；验证: `make codegen-check` 顺序执行 codegen-conventions / codegen-openapi 后 `git diff --exit-code` 覆盖 6 条 B1 路径 + 3 条 B2 路径（`backend/internal/api/generated` / `frontend/src/api/generated` / `openapi/openapi.yaml`）；任一漂移 exit 非 0
- [x] 1.6 L2 remediation X-L2：顶层 `make codegen-check` 纳入 B3 `codegen-events-check` / event-job drift gate，覆盖 `shared/events.yaml`、`shared/jobs.yaml`、Go/TS events/jobs、JSON Schema/ref/baseline 生成物；验证 `make -n codegen-check` 可见 B3 gate 且实际 `make codegen-check` exit 0；验证: 2026-04-30 `grep -n 'codegen-events-check' /tmp/easyinterview-codegen-check.dryrun` 与 `make codegen-check`

## Phase 2: 聚合层只调用已落地 sub-target

- [x] 2.1 在 Makefile 中执行已落地 sub-target，并让真实失败穿透 exit code；未来 owner 未暴露真实命令前不在 A5 保留 exit-zero target；验证: 已落地 sub-target 人工注入返回 1 时整体 `make` 必须 exit 非 0，2026-07-10 `make -n lint` 不再包含 future-owner exit-zero 命令
- [x] 2.2 登记当前已落地清单：`lint-config`（A4）已于 2026-04-30 落地切换为直接执行（`lint-getenv-boundary` + `lint-env-dict` + `lint-secrets-pattern`），B1 / B2 / B3 已落地直接执行；F1 metrics/log helper 未暴露本地命令，不进入根 `Makefile` 执行面；验证: `make -n lint-config` / `make -n lint-conventions` / `make -n codegen-conventions` / `make -n codegen-openapi` 均直接进入对应 owner 实现，focused grep 确认 F1 fake target 清零
- [x] 2.3 在 `make help` 中列出 5 个聚合入口与当前真实 sub-target owner，确保 owner subspec 改名或新增真实 helper 时聚合层可被发现；验证: 2026-07-10 `make help` 不列出 F1 fake lint target，仍列出 5 个本地质量门禁入口

## Phase 3: 文档与 CI deferral 守门

- [x] 3.1 在根 `README.md` 与 `docs/development.md` 写入 5 个本地命令；明确声明当前不存在远端 CI pipeline，文档禁用「CI 已启用」「PR required check」措辞；验证: `grep -nE '(CI 已启用\|PR required check)' README.md docs/development.md` 不命中（除 deferred / out of scope 段落以否定语境出现）；5 个 `make` 命令名出现在 onboarding 文档
- [x] 3.2 登记 spec D-5 升级触发条件（第二位长期贡献者 / 公开 release / 付费用户 / 自动发版 / 回归频率过高），并指明升级路径为原地修订 + 新增 plan（默认 `002-remote-ci`），不改 subject 路径，不把远端 CI scope 回填到 001；验证: 本 plan 同级文档 grep 命中 D-5 五条触发条件全部 + 升级路径 `002-remote-ci`；与 spec §3.1 D-5 锁定值一致
- [x] 3.3 申明 secret 红线：本地门禁不读取业务 secret；未来接入 CI 必须先递增 spec / history 登记 runner secret 字典与权限边界；验证: 本 plan 文档出现 secret 红线段落；2026-04-30 `grep -nE '(AI_GATEWAY_\|DB_\|REDIS_\|POSTHOG_)' Makefile scripts/lint/*.py` 不命中；`make -n lint` / `make -n test` / `make -n build` 不引用 `.env` 业务 secret

## Phase 4: Verification

- [x] 4.1 跑 make lint / make test / make build / make docs-check / make codegen-check 全部 exit 0；验证: 2026-04-30 顺序执行 `make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check` 全部 exit 0；2026-07-10 frontend lint fake script 已删除并由 `tsc --noEmit` 承接；2026-07-10 F1 fake lint target 已删除；backend lint/test/build 均真实执行并通过
- [x] 4.2 grep .github/workflows/ 不存在 ci.yml / nightly.yml / dependabot.yml；验证: `find .github/workflows -type f -name '*.yml' 2>/dev/null` 不命中（`.github/` 目录本身不存在）；`grep -r 'ci-pipeline-baseline' .github 2>/dev/null` 无命中
- [x] 4.3 临时把 B1 codegen-conventions 输出文件人为修改一行：make codegen-check 报错并 exit 非 0；revert 后恢复 clean；验证: 2026-04-30 在 `backend/internal/shared/types/enums.go` 第 4 行注入 `// INJECTED-DRIFT-FOR-A5-PHASE-4.3-VERIFICATION`，`make codegen-check` exit 2 并打印 `FAIL: enum drift: shared/conventions.yaml -> backend/internal/shared/types/enums.go differs; run make codegen-conventions`；revert 后再跑 exit 0
- [x] 4.4 临时让 B1 lint 引入一个 lower_snake_case 错误码：make lint 报错并 exit 非 0；revert 后恢复 clean；验证: 2026-04-30 把 `shared/conventions.yaml` 第 46 行的 `AUTH_UNAUTHORIZED` 改为 `auth_unauthorized_lower`，`make lint` exit 2 并打印 `FAIL: error code must be UPPER_SNAKE_CASE, got 'auth_unauthorized_lower'`；revert 后再跑 exit 0
- [x] 4.5 把本 plan Header 从 active 切到 completed，运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` 同步 plans/INDEX.md；再运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 确认无 drift；验证: 2026-04-30 plan / checklist Header 切到 `completed` v1.3；`docs/spec/ci-pipeline-baseline/plans/INDEX.md` 状态投影同步；`make docs-check` exit 0

## Phase 5: docs/spec heading anchor gate hardening

- [x] 5.1 Red: 在 `scripts/lint/check_md_links_test.py` 补充 fragment anchor contract tests，覆盖 missing fragment、pure in-page anchor、GitHub-style 中文/标点/multi-hyphen slug、duplicate heading `-1` 与未启用 fragment 检查的兼容行为；验证: 2026-05-04 修改测试后、实现前运行 `python3 -m unittest scripts/lint/check_md_links_test.py` exit 1，3 个新 fragment 测试因 `scan_directory() got an unexpected keyword argument 'check_fragments'` 失败，确认当前脚本缺少 fragment 校验能力
- [x] 5.2 Green: 扩展 `scripts/lint/check_md_links.py`，新增 `--check-fragments`，在目标 Markdown 文件存在后验证 fragment 是否命中 GitHub-style heading anchor；验证: 2026-05-04 `python3 -m unittest scripts/lint/check_md_links_test.py` exit 0，15 个测试通过，原有相对链接检查测试仍通过
- [x] 5.3 集成 `make docs-check`：保留全 `docs/` 相对链接检查，并新增 docs/spec fragment anchor pass；验证: 2026-05-04 `make -n docs-check` 可见 `scripts/lint/check_md_links.py` 分别以 `docs` 与 `docs/spec --check-fragments` 执行
- [x] 5.4 Verification and lifecycle close：执行 `python3 -m unittest scripts/lint/check_md_links_test.py`、`python3 scripts/lint/check_md_links.py docs/spec --ignore '**/TEMPLATES.md' --check-fragments`、`make docs-check`、`python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/context.yaml --target repo`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`git diff --check`；全部通过后把 plan/checklist Header 切回 completed 并同步 plans INDEX；验证: 2026-05-04 focused unit test exit 0（15 tests），docs/spec fragment audit exit 0，`make docs-check` exit 0，context validation exit 0，`sync-doc-index --check` zero drift，`git diff --check` exit 0；实施过程中修复 3 个 `event-and-outbox-contract` 旧 fragment 链接中的 `job_type` 下划线漂移

## Phase 6: Frontend lint gate cleanup

- [x] 6.1 删除 frontend package lint 假通过：`frontend/package.json` 的 `lint` 改为 `tsc --noEmit`，删除未接入 ESLint 依赖的 `frontend/.eslintrc.cjs`；验证: `pnpm --filter @easyinterview/frontend lint` 与 `pnpm --filter @easyinterview/frontend typecheck` 均运行真实 TypeScript 静态检查
- [x] 6.2 同步本地门禁文档并确认旧 fake lint 口径清零；验证: focused grep（排除本 plan 与历史 work-journal）确认旧 frontend lint exit-zero wording 和未接入 ESLint 配置已从执行面删除，`make docs-check`、context validation、`sync-doc-index --check` 与 `git diff --check` 通过

## Phase 7: Observability lint fake target deletion

- [x] 7.1 删除根 `Makefile` 的 F1 fake lint target：移除对应 phony target，并从 `make lint` 依赖链删除；验证: Red dry-run 只暴露旧 exit-zero 输出，Green 后 `make -n lint` 不再包含该命令，`make help` 不再列出该 target
- [x] 7.2 同步 A5 spec / plan / checklist / onboarding 文档并确认旧口径清零；验证: focused grep 确认旧 F1 fake target 名称、旧 exit-zero 输出和旧剩余 exit-zero 描述不再出现在当前执行面和 A5 owner 文档；`make lint`、context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 通过

## Phase 8: Build gate wording cleanup

- [x] 8.1 同步 build gate 当前合同：A5 spec / plan / checklist 改为 `make build` 真实执行 backend cmd build 与 frontend Vite build，不记录旧 frontend exit-zero 输出；验证: `make -n build` 只显示 `go build ./cmd/...` 与 `pnpm --filter @easyinterview/frontend build`
- [x] 8.2 验证 build gate 与文档清零；验证: `make build` 通过；focused grep 确认旧 build exit-zero 文本不再出现在 A5 owner 文档；context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 通过

## Phase 9: Python tooling and skill contract aggregation

- [x] 9.1 修正 work-journal auto-mode subject derivation 的陈旧合同断言，保持当前英文翻译/概括与自然小写规则；验证: focused test 与 full Python suite
- [x] 9.2 新增根 `requirements-dev.txt` 声明 `pytest` / `PyYAML`，并把 `python3 -m pytest scripts .agent-skills -q` 接入既有 `make test`；验证: Makefile contract RED/GREEN 与 failure propagation
- [x] 9.3 完整收口；验证: full Python suite、`make test`、`make lint`、A5/product contexts、README/development、docs/index/diff/pruning gates
  <!-- red: 2026-07-10 method=python-contract-aggregation-contract evidence="The existing full Python run failed 1/463 on a stale work-journal literal. The added focused Makefile contract and the stale assertion then failed 2/2, proving both the missing suite/dependency declaration and the contract drift." -->
  <!-- verified: 2026-07-10 method=root-python-contract-test-aggregation evidence="Focused tests pass 2/2; Python passes 464 tests plus 4269 subtests. Root make test also passes all backend packages and frontend 136 files/836 tests. A zero-match PYTEST_ADDOPTS probe exits at the Python step before Go/Vitest. Root lint, requirements temporary-target dry-run, A5/product contexts, BUG-0156, retrospective, docs/index/diff/pruning gates pass." -->

## Phase 10: UI prototype Node contract aggregation

- [x] 10.1 把 `node --test ui-design/ui-design-contract.test.mjs` 接入既有 `make test` 首段，不新增 target；验证: Makefile contract RED/GREEN 与 UI contract 45 tests
- [x] 10.2 完整收口；验证: `make test` 保留 Python/backend/frontend 全量 gate，A5/product contexts、README/development、docs/index/diff/pruning gates
  <!-- red: 2026-07-10 method=root-node-test-aggregation-contract evidence="The extended Makefile contract failed only because the repository's sole root Node test was absent from the test recipe; the UI contract itself already passed 45/45." -->
  <!-- verified: 2026-07-10 method=ui-prototype-node-contract-aggregation evidence="Focused Makefile contract passes 1/1 and UI contract passes 45/45. Root make test then runs UI, Python 464 tests plus 4269 subtests, all backend packages and frontend 136 files/836 tests. A5/product contexts, onboarding docs, docs/index/diff/pruning gates pass. Scenario-only contracts remain outside the unit aggregator. No new Bug or retrospective report was needed." -->
