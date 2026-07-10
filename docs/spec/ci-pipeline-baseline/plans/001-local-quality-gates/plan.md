# Local Quality Gates Bootstrap

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ci-pipeline-baseline spec §2.1](../../spec.md#21-in-scope) 锁定的 5 个本地入口 target 在仓库根 `Makefile` 上接齐：`make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`。本 plan 只是聚合层，不重写已有 owner 的 lint / generator / config check 实现，只把它们组织成单人开发阶段可重复的本地质量门禁，并显式记录从本地门禁升级为远端 CI 的 [D-5 触发条件](../../spec.md#31-已锁定决策)。

2026-07-10 原地修订范围：删除前端未接入 ESLint 配置，把 `pnpm --filter @easyinterview/frontend lint` 从 `exit 0` 假通过改为现有 TypeScript 静态检查（`tsc --noEmit`）。本修订不引入新工具链依赖，不改变远端 CI deferral，仅让已落地 frontend package lint 不再假通过。

2026-07-10 追加修订范围：删除根 `Makefile` 的 F1 observability exit-zero 假 target。A5 聚合层只调用当前已落地的本地 gate；F1 metrics / log lint helper 由 F1 owner 暴露真实命令后再接入，不在 A5 中保留未来 owner 的 exit-zero 假入口。

2026-07-10 二次追加修订范围：删除 `make build` 文档中的旧 frontend build exit-zero 口径。当前 build gate 真实执行 `cd backend && go build ./cmd/...` 与 `pnpm --filter @easyinterview/frontend build`；A5 不再记录 frontend build 的 exit-zero 输出。

2026-07-10 三次追加修订范围：根 `make test` 补入当前已存在但未聚合的 `scripts/` 与 `.agent-skills/` Python contracts，并新增根 `requirements-dev.txt` 显式声明 `pytest` / `PyYAML`。本修订同时修正全量 Python suite 暴露的一条 work-journal 陈旧断言，不改变 skill 当前 commit message 规则。

2026-07-10 四次追加修订范围：把唯一根级 Node test `ui-design/ui-design-contract.test.mjs` 纳入 `make test` 首段，使 UI 真理源的 45 项静态契约不再只由单个场景显式触发。场景专用 tests 继续由各场景 owner 运行。

本 plan 是 `ci-pipeline-baseline` 当前唯一的 active plan，只负责本地质量门禁聚合。当 D-5 触发条件出现（第二位长期贡献者、公开 release、付费用户上线、自动发版、回归频率过高）时，在本 spec 原地修订并新增 `002-remote-ci` 或等价 plan；远端 CI workflow、branch protection、artifact、runner secret 不得塞回本 plan。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 把 A5 列为 Layer A · Foundation 的最后一份 child；本 plan 也是当前 Foundation 收口的最后一公里。聚合层依赖：

- [A1 repo-scaffold spec D-3](../../../repo-scaffold/spec.md#31-已锁定决策) 在根 `Makefile` 上锁定的 10 个 phony target（含 `lint` / `test` / `build`），本 plan 把其中已落地入口升级为聚合实现。
- [B1 shared-conventions-codified spec §2.1](../../../shared-conventions-codified/spec.md#21-in-scope) 提供的 `make codegen-conventions` / `make lint-conventions`（错误码 / 枚举命名 / 共享类型 generator drift）。
- [B2 openapi-v1-contract plan §3 Phase 2.3](../../../openapi-v1-contract/plans/001-bootstrap/plan.md#23-make-入口) 提供的 `make codegen-openapi` / `make codegen-check`。
- [A4 secrets-and-config spec §2.1](../../../secrets-and-config/spec.md#21-in-scope) 提供的 `make lint-config`（env key 与 `.env.example` drift）。
- [F1 observability-stack](../../../observability-stack/spec.md) 承接 metric / log lint helper；当前 F1 未暴露独立本地命令，A5 不制造 exit-zero 假 target。

每个 phase 都是可独立验证的纵向切片：Phase 1 把入口 target 串通；Phase 2 锁定聚合层只调用已落地 sub-target 的边界；Phase 3 收口文档与 CI deferral 守门；Phase 4 跑 spec [§6 验收标准](../../spec.md#6-验收标准) C-1..C-7 自检并贴日志。本 plan 不引入 BDD 资产、不创建 `.github/workflows/*.yml`，也不修改任何 owner subspec 的规则语义。

## 3 质量门禁分类

- **Plan 类型**: `tooling + contract + code-internal`。本 plan 在仓库根 `Makefile` 上聚合 5 个本地质量入口 target（`make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`），调用 B1 / B2 / A4 / A3 / F3 / E1 等 owner 已暴露的 lint / generator / config check 与轻量脚本（`scripts/lint/check_md_links.py`）。2026-05-04 修订把 docs/spec heading fragment anchor drift 纳入 `docs-check`；2026-07-10 修订把 frontend package lint 从 exit-zero script 改为 typecheck-backed gate，删除 F1 未落地 helper 的 exit-zero Make target，并把 build gate 文档更新为真实 backend/frontend build。属于本地工具链与契约聚合层，不引入用户可感知 UI、HTTP API 行为、业务工作流或端到端功能。
- **TDD 策略**: 必须通过 `/tdd --file docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/checklist.md --references docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md,docs/spec/ci-pipeline-baseline/spec.md --phase-commit ci-pipeline-baseline/001-local-quality-gates` 顺序执行。每个 checklist item 以本 checklist 内的 `验证:` 子句作为 Red-Green-Refactor 断言来源；涉及 sub-target 接入或删除的 item 必须先复现当前命令输出或失败状态，再最小实现并复跑 focused command。Phase 5 的 Red 来源是 `scripts/lint/check_md_links_test.py` 对缺失 fragment、GitHub-style slug、重复标题后缀、纯页内 anchor 和非 fragment 相对链接兼容性的断言。Phase 7 的 Red 来源是根 lint 聚合中仍存在的 F1 exit-zero 假 target。Phase 9 的 Red 来源是全量 Python suite 暴露的 work-journal contract failure，以及 Makefile contract 对 Python suite / dependency declaration 缺失的断言。Phase 10 的 Red 来源是 Makefile contract 对唯一根级 Node test 缺失的断言。
- **BDD 策略**: BDD 不适用。本 plan 只在 Makefile / 文档 / 自检脚本层操作，不产生浏览器 UI、外部 API、业务工作流或端到端场景测试可观察行为，因此不创建 `bdd-plan.md` / `bdd-checklist.md`，主 checklist 也不设置 `BDD-Gate:`。
- **替代验证 gate**: 使用本地 lint + drift + smoke 组合代替 BDD：5 个聚合入口端到端跑通（Phase 4.1）；已落地失败穿透双向 fail-injection 自检（Phase 4.3 / 4.4）；`grep -r '\.github/workflows' .` 远端 CI 文件零存在性自检（Phase 4.2）；`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` Header / INDEX drift gate（Phase 4.5）；`python3 scripts/lint/check_md_links.py docs/spec --ignore '**/TEMPLATES.md' --check-fragments` docs/spec fragment anchor drift gate（Phase 5）；frontend package lint 由 `pnpm --filter @easyinterview/frontend lint` / `typecheck` 证明（Phase 6）；F1 假 target 删除由 `make -n lint` 和 focused grep 证明（Phase 7）；真实 build gate 由 `make -n build` 与 `make build` 证明（Phase 8）。

## 4 实施步骤

### Phase 1: 入口 target 聚合（lint / test / build / docs-check / codegen-check）

#### 1.1 `make lint` 聚合

把根 `Makefile` 现有 `lint` 入口接入为聚合 target，按下列顺序串行调用：

1. `$(MAKE) lint-conventions`（B1 owner，错误码 `UPPER_SNAKE_CASE` / 枚举 `lower_snake_case` / `camelCase` JSON tag）。
2. `$(MAKE) lint-config`（A4 owner，`.env.example` 与代码 `Get*` 调用一致性）。
3. Go lint：`cd backend && golangci-lint run ./...`。
4. Frontend static gate：`pnpm --filter @easyinterview/frontend lint`（当前执行 `tsc --noEmit`，不保留未接入 ESLint 配置）。

任一已存在 sub-target 失败必须以非 0 退出（spec D-1 / [§4.1](../../spec.md#41-本地门禁约束)）。

#### 1.2 `make test` 聚合

`test` target 调用：

1. UI prototype contract：`node --test ui-design/ui-design-contract.test.mjs`。
2. Python tooling / skill contracts：`python3 -m pytest scripts .agent-skills -q`。
3. Go 单元测试：`cd backend && go test ./...`。
4. TS 单元测试：`pnpm --filter @easyinterview/frontend test`。

AI 单元测试必须走 stub / fixtures provider（[B1 spec §2.1](../../../shared-conventions-codified/spec.md#21-in-scope) + [A3 ai-provider-and-model-routing spec](../../../ai-provider-and-model-routing/spec.md) 共同约定）；`AI_PROVIDER_*` 真实 secret 不读取，`APP_ENV=test` 路径才允许 stub（spec [§4.2](../../spec.md#42-安全与权限约束)）。

#### 1.3 `make build` 聚合

`build` target 调用：

1. 后端：`cd backend && go build ./cmd/...`。
2. 前端：`pnpm --filter @easyinterview/frontend build`。

该 gate 只列出当前真实 build 命令；任一 build 命令失败必须让 `make build` 返回非 0。

#### 1.4 `make docs-check` 聚合

新增根 `docs-check` target，串行执行：

1. `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`（spec [§4.1 文档约束](../../spec.md#41-本地门禁约束) 强制项）。
2. 轻量相对链接检查：调用 `python3 scripts/lint/check_md_links.py docs`（由本 plan 新增）或等价脚本扫描 `docs/**/*.md` 内的相对链接。
3. docs/spec fragment anchor 检查：调用 `python3 scripts/lint/check_md_links.py docs/spec --ignore '**/TEMPLATES.md' --check-fragments` 或等价脚本，确认 `docs/spec/**/*.md` 中每个本地 Markdown fragment 都解析到目标文件内的 GitHub-style heading anchor；缺失 anchor 必须 exit 非 0。

任一步骤报告 drift / 失链必须 exit 非 0（spec C-5）。

#### 1.5 `make codegen-check` 聚合

把 [A1 plan §3.2.1](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 的 `codegen` handoff 升级为聚合 codegen drift gate：

1. `$(MAKE) codegen-conventions`（B1）+ `git diff --exit-code -- backend/internal/shared/types backend/internal/shared/errors backend/internal/shared/idx frontend/src/lib/conventions frontend/src/lib/ids shared/conventions.yaml`。
2. `$(MAKE) codegen-openapi`（B2）+ `git diff --exit-code -- backend/internal/api/generated frontend/src/api/generated openapi/openapi.yaml`。

执行顺序为 `codegen-conventions` → `codegen-openapi`，与 [B2 plan §3 Phase 2.3](../../../openapi-v1-contract/plans/001-bootstrap/plan.md#23-make-入口) 保持一致；任一 generator 漂移即失败（spec C-6）。

#### 1.6 L2 remediation: `codegen-check` 纳入 B3 events/jobs drift

修复 plan-code-review finding X-L2：B3 `event-and-outbox-contract` 已落地事件/任务 generator 与 `codegen-events-check` 后，A5 顶层 `make codegen-check` 必须把 B3 event/job drift 纳入同一聚合门禁，避免开发者只跑标准本地 gate 时漏掉 `shared/events.yaml`、`shared/jobs.yaml` 及其 Go/TS/JSON Schema/baseline 生成物漂移。验证包括 `make -n codegen-check` 可见 B3 gate、实际 `make codegen-check` 通过。

### Phase 2: 聚合层只调用已落地 sub-target

#### 2.1 区分「未落地 future owner」与「已落地但失败」

聚合层必须明确区分两种状态：

- **未落地 future owner**（owner subspec 的 lint / generator 当前不存在）：不得在 A5 `Makefile` 中保留 exit-zero 执行 target；只在 owner spec/plan 中保留未来 scope，等真实 helper 暴露后再接入。
- **已落地但失败**（sub-target 存在且返回非 0）：聚合层必须立即 fail-fast，整体 `make` 命令 exit 非 0；不允许吞掉错误码。

#### 2.2 当前已落地 sub-target 清单

当前 baseline 的 lint / codegen 映射：

| Sub-target | Owner | 当前状态 | 执行 |
|------------|-------|----------|------|
| `codegen-openapi` | B2 openapi-v1-contract | 已落地（参见 [B2 001-bootstrap](../../../openapi-v1-contract/plans/001-bootstrap/plan.md)） | 直接执行 |
| `codegen-conventions` | B1 shared-conventions-codified | 已落地（参见 [B1 001-bootstrap](../../../shared-conventions-codified/plans/001-bootstrap/plan.md)） | 直接执行 |
| `lint-conventions` | B1 shared-conventions-codified | 已落地 | 直接执行 |
| `lint-config` | A4 secrets-and-config | 已落地（[A4 001-bootstrap](../../../secrets-and-config/plans/001-bootstrap/plan.md)，2026-04-30 切换为直接执行） | 直接执行 `lint-getenv-boundary` + `lint-env-dict` + `lint-secrets-pattern` |

Frontend package lint 已落地：`pnpm --filter @easyinterview/frontend lint` 当前执行 `tsc --noEmit`，`.eslintrc.cjs` 在未安装 ESLint 依赖前不保留。

F1 metrics / log lint helper 当前未暴露本地命令，因此不进入根 `Makefile` 的执行面。

#### 2.3 执行面发现性

在 `make help` 中列出当前真实聚合入口与已落地 sub-target owner。owner subspec 改名或新增真实 helper 时，必须递增本 spec / plan 并同步更新聚合层；不得用 exit-zero target 代替 owner 实现。

### Phase 3: 文档与 CI deferral 守门

#### 3.1 根 `README.md` 与 `docs/development.md` 更新

按 spec [§4.3 文档约束](../../spec.md#43-文档约束) 把 5 个本地命令写入 onboarding 文档；明确声明项目当前不存在远端 CI pipeline，避免误导外部贡献者去找 `.github/workflows/`。文档不得使用「CI 已启用」「PR required check」等措辞。

#### 3.2 D-5 升级触发条件登记

在本 plan 同级文档中写明：当满足任一 [spec D-5](../../spec.md#31-已锁定决策) 触发条件时，新增 `002-remote-ci` 或等价 plan 承接远端 CI：

- 第二位长期贡献者加入。
- 公开 release branch 出现。
- 付费用户上线。
- 需要自动发版 / release gate。
- 回归频率过高、本地门禁不足以控制。

升级路径为 spec [§7 关联计划](../../spec.md#7-关联计划) 已声明的「原地修订 + 新增 plan」（默认 `002-remote-ci`），不改 subject 路径，不创建 sibling spec，也不把远端 CI scope 回填到 001。

#### 3.3 secret 红线再申明

当前不创建 runner / workflow / artifact 上传通道；任何 AI provider / DB / Redis / PostHog secret 全部留在本地 `.env`，不进入聚合 target。未来如接入 CI，必须先递增 spec / history 登记 runner secret 字典与权限边界（spec [§4.2](../../spec.md#42-安全与权限约束)）。

### Phase 4: Verification

#### 4.1 5 个本地入口端到端跑通

在干净仓库下依次执行 `make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`，记录每条命令 exit code；输出贴入工作日志。spec C-2 / C-3 / C-4 / C-5 / C-6 全部成立。

#### 4.2 远端 CI 文件零存在性自检

`grep -r '\.github/workflows' .` 不应命中由本 plan 创建的 `ci.yml` / `nightly.yml` / `dependabot.yml`；文档中未出现「CI 已启用」描述（spec C-1 / C-7）。

#### 4.3 「已落地但失败」边界自检

人工注入两次错误后 revert：

- 把 B1 任一 generator 输出文件改一行：`make codegen-check` 必须 exit 非 0，提示 generated drift；revert 后 gate 恢复 clean。
- 在 B1 lint 输入中插入一个 `lower_snake_case` 错误码：`make lint` 必须 exit 非 0，提示错误码命名违规；revert 后 gate 恢复 clean。

证明聚合层不会因「exit-zero 逻辑」吞掉真实失败（spec C-2 / C-6）。

#### 4.4 文档与 INDEX 同步

- 本 plan checklist 全部勾选；Phase 4 命令日志贴入工作日志。
- 把本 plan Header 从 `active` 切到 `completed`，运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` 同步 [ci-pipeline-baseline/plans/INDEX.md](../INDEX.md)。
- 运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 确认 `docs/spec/INDEX.md` 与 `plans/INDEX.md` 与 spec / plan Header 无 drift。

### Phase 5: docs/spec heading anchor gate hardening

#### 5.1 Red: fragment anchor contract tests

在 `scripts/lint/check_md_links_test.py` 中补充 Red 测试，覆盖：

- `target.md#missing-heading` 必须被报告为 broken fragment。
- `#local-heading` 纯页内 anchor 在 `--check-fragments` 下必须被验证。
- GitHub-style heading slug 必须支持中文、标点移除后保留多 hyphen、重复 heading 的 `-1` 后缀。
- 未启用 `--check-fragments` 时保留既有相对链接检查兼容行为。

#### 5.2 Green: implement fragment-aware markdown link check

扩展 `scripts/lint/check_md_links.py`，新增 `--check-fragments` 参数。启用时，脚本在确认目标 Markdown 文件存在后读取 heading 列表，生成 GitHub-style anchor 集合，验证 fragment 是否命中；未启用时保持当前只检查路径存在的行为。外部链接、代码块、inline code、HTML comment、`TEMPLATES.md` ignore 规则沿用既有语义。

#### 5.3 Integrate with `make docs-check`

把 docs/spec fragment anchor pass 接入根 `Makefile` 的 `docs-check` target。`make docs-check` 必须先跑 Header / INDEX drift，再跑全 `docs/` 相对链接检查，最后跑 docs/spec fragment anchor 检查；任何一步失败均返回非 0。

#### 5.4 Verification and lifecycle close

完成实现后执行 focused unit test、docs/spec fragment audit、`make docs-check`、context validation、`sync-doc-index --check` 与 `git diff --check`。全部通过后把本 plan / checklist Header 切回 `completed`，同步 plans INDEX，并记录 work journal。

### Phase 6: Frontend lint gate cleanup

#### 6.1 Replace frontend package lint no-op

将 `frontend/package.json` 的 `lint` script 从 `echo ... && exit 0` 改为 `tsc --noEmit`，删除未接入且无依赖支撑的 `frontend/.eslintrc.cjs`。本阶段不引入 ESLint dependency，不新增规则配置；未来若需要 ESLint，必须由对应 owner 在当前工具链基础上显式添加依赖、规则和 focused tests。

#### 6.2 Verification and docs sync

执行 `pnpm --filter @easyinterview/frontend lint` 与 `pnpm --filter @easyinterview/frontend typecheck` 证明 lint/typecheck 入口一致且真实失败可穿透；执行 focused grep（排除本 plan 与历史 work-journal）确认旧 frontend lint exit-zero wording 和未接入 ESLint 配置已从执行面删除；执行 `make docs-check`、context validation、`sync-doc-index --check` 与 `git diff --check` 收口文档索引和 whitespace gate。

### Phase 7: Observability lint fake target deletion

#### 7.1 Delete F1 fake lint target

删除根 `Makefile` 中的 F1 fake phony target，并从 `make lint` 依赖链移除。当前 F1 observability 相关 Go contract tests 继续由 `make test` 覆盖；A5 不把测试伪装成 metrics / log lint helper。F1 后续若暴露真实 `lint-metrics` / `lint-logs` 或等价命令，再通过本 spec / plan 原地修订接回 `make lint`。

#### 7.2 Verification and docs sync

执行 `make -n lint` 证明根 lint 聚合不再调用 F1 假 target；执行 focused grep 确认旧 F1 fake target 名称、旧 exit-zero 输出和旧剩余 exit-zero 描述不再出现在当前执行面和 A5 owner 文档；执行 `make lint`、context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 收口。

### Phase 8: Build gate wording cleanup

#### 8.1 Sync build gate contract

更新 A5 spec / plan / checklist 中的 build gate 描述：`make build` 当前只执行真实 backend cmd build 与 frontend Vite build，不记录 frontend build 的旧 exit-zero 输出，也不把 future component scope 塞入当前 A5 聚合层。

#### 8.2 Verification and docs sync

执行 `make -n build` 证明 build 聚合只调用 backend/frontend 真实命令；执行 `make build` 证明当前 gate 通过；执行 focused grep 确认旧 build exit-zero 文本不再出现在 A5 owner 文档；执行 context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 收口。

### Phase 9: Python tooling and skill contract aggregation

#### 9.1 Repair the stale work-journal contract assertion

保持 `/work-journal` 当前“移除 Phase 前缀后翻译/概括为简洁英文，并在自然情况下小写”的规则，修正测试中只接受旧 lowercase remainder 文本的陈旧断言。Focused test 与全量 Python suite 都必须通过。

#### 9.2 Declare Python dev dependencies and wire the suite

新增根 `requirements-dev.txt`，只声明当前 Python tooling/tests 实际需要的 `pytest` 与 `PyYAML`。在 `make test` 中执行 `python3 -m pytest scripts .agent-skills -q`，失败必须阻止后续 gate；保留既有 Go 与 frontend test 命令，不增加平行 test target。

#### 9.3 Verification and docs sync

执行 focused work-journal/Makefile contract tests、全量 Python suite、`make test`、`make lint`、A5/product contexts、docs/index/diff/pruning gates。更新 README 与 `docs/development.md` 的 test gate 和依赖安装说明；全部通过后恢复本 plan completed。

### Phase 10: UI prototype Node contract aggregation

#### 10.1 Wire the existing prototype contract

扩展 Makefile contract，先证明 `make test` 未执行 `ui-design/ui-design-contract.test.mjs`，再把 `node --test` 命令加到现有 test target 首段。保留 Python → Go → frontend 后续顺序，不创建新 target。

#### 10.2 Verification and docs sync

执行 focused Makefile contract、UI contract 45 tests、完整 `make test`、A5/product contexts、README/development 与 docs/index/diff/pruning gates；确认场景专用 Python contract 不被扩大到根单元测试聚合。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 至 C-7 全部成立；drift / 失败 / 守门边界由 Phase 4 命令日志佐证；docs/spec fragment anchor drift 由 Phase 5 命令日志佐证；exit-zero future-owner target 清零由 Phase 7 命令日志佐证；真实 build gate 由 Phase 8 命令日志佐证；Python contracts 与依赖声明由 Phase 9 佐证；UI prototype contract 聚合由 Phase 10 佐证。
- 本 plan checklist 全部勾选；Phase 4.1 与 Phase 4.3 关键命令日志贴入工作日志。
- 不出现 `.github/workflows/*.yml` 由本 plan 创建；文档不声称项目已有远端 CI。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 聚合层把 future owner 当作通过，掩盖了已落地 sub-target 的真实失败 | Phase 2.1 强制 A5 只调用已落地 sub-target；Phase 4.3 通过「人工注入 → 命令必须失败 → revert」的双向自检，证明聚合层不会吞掉非 0 退出码 |
| 业务 secret（AI / DB / Redis / PostHog）被聚合 target 误读，写入日志或缓存 | spec D-4 + Phase 3.3 红线再申明；`make test` 显式约束 `APP_ENV=test` 路径走 stub provider；任何新增 sub-target 必须经 spec 修订后再接入 |
| A5 聚合层与各 owner subspec 的 Make target 命名 / 行为漂移（owner 把 `lint-conventions` 改名后 A5 没跟上） | Phase 1 在 `make help` 中列出 5 个聚合入口与对应 sub-target；任何 owner 修改 sub-target 名称必须递增本 spec / plan，并同步更新聚合层；CI deferral 期内由原作者 + 本 plan owner review 控制 |
| 因 D-5 条件未触发就提前创建 `.github/workflows/*.yml` | Phase 4.2 把「不存在 ci.yml / nightly.yml / dependabot.yml」纳入收口自检；spec [§3.2](../../spec.md#32-待确认事项) 把升级路径锁定在原地修订；任何 PR 引入 workflow 文件而未先修订 spec 必须被 owner 拒绝 |
| `make docs-check` 在 macOS / Linux 不同 shell 行为不一致导致 `/sync-doc-index --check` 误报 | Phase 1.4 强制聚合在仓库根；首次落地后在 macOS zsh 与 Linux bash 各跑一次；调用 skill 时显式 set `LC_ALL=C.UTF-8` 避免本地化输出差异 |
| Fragment anchor slug 规则与 GitHub 渲染锚点存在细微差异导致误报 | Phase 5 单元测试锁定本仓库实际使用的中文标题、标点、多 hyphen 与重复 heading 场景；实现保持最小 GitHub-style slugger，不引入 Markdown 结构格式化检查 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-10 | 1.11 | Add the repository's only root Node test, the UI prototype contract, to the existing test aggregator. | tech-debt pruning |
| 2026-07-10 | 1.10 | Add Python tooling/skill contracts and explicit dev dependencies to the root test gate; repair one stale work-journal assertion. | tech-debt pruning |
| 2026-07-10 | 1.9 | 将 A5 001 当前 plan / checklist 中的旧 scaffold wording 收敛为 exit-zero 假 target / 真实 gate 术语。 | tech-debt pruning |
| 2026-07-10 | 1.8 | 删除 `make build` 文档中的旧 frontend build exit-zero 口径；当前 build gate 执行真实 backend cmd build 与 frontend Vite build。 | tech-debt pruning |
| 2026-07-10 | 1.7 | 删除 F1 observability exit-zero lint target；A5 聚合层只调用已落地 gate，F1 helper 暴露真实命令后再接入。 | tech-debt pruning |
| 2026-07-10 | 1.6 | 删除 frontend package lint 的 `exit 0` no-op 和未接入 ESLint 配置，改为 typecheck-backed local gate。 | tech-debt pruning |
| 2026-05-04 | 1.5 | 原地修订 `make docs-check`：新增 docs/spec heading fragment anchor audit gate，补充 TDD 测试与 Makefile 集成要求。 | implement remediation |
| 2026-04-30 | 1.4 | L2 code-review remediation：顶层 `make codegen-check` 纳入 B3 event/job drift gate。 | plan-code-review --fix |
| 2026-04-29 | 1.1 | 收口 plan-review：docs-check 改为可执行 sync-doc-index 脚本 + `scripts/lint/check_md_links.py`；B1 codegen diff 覆盖 errors / idx / frontend ids；远端 CI 明确由 future `002-remote-ci` 承接。 | plan-review remediation |
| 2026-04-30 | 1.2 | 补齐 `## 3 质量门禁分类`：Plan 类型 / TDD 策略 / BDD 不适用声明 / 替代验证 gate；renumber 后续章节并修复内部 §3.2→§4.2 引用。同步 checklist 16 项 `验证:` 子句。 | implement gate remediation |
| 2026-04-30 | 1.3 | L2 code review remediation：reopen 真实 backend/frontend gate、Go lint、help owner 标签与 secret grep 证据漂移，修复后重新验证。 | plan-code-review remediation |
