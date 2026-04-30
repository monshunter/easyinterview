# ci-pipeline-baseline/001-local-quality-gates 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Claude (Opus 4.7) / monshunter

**关联计划**: [001-local-quality-gates](../spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md)

## 1 复盘范围与成功证据

### 1.1 交付范围

A5 `ci-pipeline-baseline` 的本地质量门禁聚合层。具体落地：

- 根 `Makefile` 5 个聚合入口：`make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`
- 新增 F1 `lint-observability` 占位 target（NOT-YET-LANDED owner 输出 `not implemented yet: F1 observability-stack` exit 0）
- 新增 `docs-check` target，调用 `sync-doc-index --check` + 新写的 `scripts/lint/check_md_links.py`
- 新增 `scripts/lint/check_md_links.py` 与契约测试（11 项），覆盖 markdown 相对链接、HTML 注释、inline code、TEMPLATES.md 占位的多重边界
- 根 `README.md` + `docs/development.md` 写入 5 个本地命令、CI deferral 声明、D-5 升级触发条件、secret 红线
- 5 个聚合入口的 `make help` 文案补齐 owner 标签（B1 / B2 / A4 / F1 / aggregator）
- A4 secrets-and-config spec §4.1 v1.7 → v1.8：边界 allow-list 扩展 `cmd/migrate/main.go`，`scripts/lint/getenv_boundary.go` `defaultAllowlist` 同步
- 关联的跨子规格文档修复：12 条 docs/ 内相对链接深度错误（4 个子规格 + 2 个 work-journal 文件）
- 关联的子规格 spec 修复：`local-dev-stack/spec.md` 把 `[test/scenarios/](../../../test/scenarios/)` 改为非链接引用，避免引用一个 W4 才会 spawn 的目录

### 1.2 成功证据

| 类别 | 证据 |
|------|------|
| Checklist 完整性 | 16/16 全部勾选；L2 remediation 后 plan / checklist Header 切到 `completed` v1.3 |
| 5 个聚合入口 exit 0 | 2026-04-30 L2 remediation 后顺序执行：`make lint` exit 0 / `make test` exit 0 / `make build` exit 0 / `make docs-check` exit 0 / `make codegen-check` exit 0；`make lint` 输出 `not implemented yet: F1 observability-stack`，backend `golangci-lint run ./...` 报 `0 issues.`；`make test` 真实运行 backend `go test ./...` 与 frontend `vitest run`（10 files / 49 tests）；`make build` 真实运行 backend `go build ./cmd/...`，frontend build 保留 D1-owned placeholder |
| L2 remediation 证据 | `BUG-0003` 记录本地质量 gate 曾跳过真实 backend/frontend 执行；修复后 `make -n lint/test/build` 均显示真实命令路径，`make help` owner 标签 grep 全部命中，secret 前缀 grep 在 `Makefile scripts/lint/*.py` 不命中 |
| Phase 4.3 codegen drift fail-injection | 在 `backend/internal/shared/types/enums.go` 第 4 行注入 `// INJECTED-DRIFT-FOR-A5-PHASE-4.3-VERIFICATION`，`make codegen-check` exit 2 + `FAIL: enum drift: shared/conventions.yaml -> backend/internal/shared/types/enums.go differs`；revert 后 exit 0 |
| Phase 4.4 lint fail-injection | 在 `shared/conventions.yaml:46` 注入 `auth_unauthorized_lower`，`make lint` exit 2 + `FAIL: error code must be UPPER_SNAKE_CASE, got 'auth_unauthorized_lower'`；revert 后 exit 0 |
| 远端 CI 文件零存在性 | `find .github/workflows -type f -name '*.yml' 2>/dev/null` 不命中（`.github/` 目录本身不存在）；`grep -r 'ci-pipeline-baseline' .github` 无命中 |
| INDEX 一致性 | `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 报告 zero drift；`docs/spec/ci-pipeline-baseline/plans/INDEX.md` 自动同步 plan 状态 `completed` |
| `check_md_links.py` 单元测试 | `python3 -m unittest scripts.lint.check_md_links_test` 11/11 通过 |

## 2 会话中的主要阻点/痛点

### 2.1 计划本身缺 `## 3 质量门禁分类` section

- **证据**：plan v1.1（2026-04-29）章节为 `## 1 目标` / `## 2 背景` / `## 3 实施步骤` / ...；2026-04-30 落地的「TDD/BDD 质量门禁分类」codification rule 要求每个 plan 包含 `## 3 质量门禁分类`。`/implement` Step 4.2 命中缺失 → 阻塞，必须先走 `/plan-review --fix`。
- **影响**：会话开头额外 1 个 plan-review 周期：插入 `## 3 质量门禁分类` 章节、renumber 后续章节、修复内部 §3.2 → §4.2 引用、checklist 16 项追加 `验证:` 子句、bump v1.1 → v1.2。

### 2.2 `scripts/lint/check_md_links.py` 在 plan 中被引用但未实现

- **证据**：plan §1.4 与 §3.2 明确写入 `python3 scripts/lint/check_md_links.py docs`，但仓库中文件不存在（`test -f scripts/lint/check_md_links.py` → ABSENT）。
- **影响**：本 plan 必须从零开始创建脚本 + 契约测试，并在迭代中追加 4 类边界（HTML 注释、code fence、inline code、`--ignore` glob 模式），加上对 docs/ 跑出 38 条 / 28 条 / 12 条逐次降级的 broken link 调查。脚本本身实现工作量约占本 plan 的 30%，超出原 §1.4 的「轻量相对链接检查」描述。

### 2.3 当前分支与工作树状态与会话起始 snapshot 严重不一致

- **证据**：`gitStatus` snapshot 显示分支 `feature/spec-init` + 工作树 `(clean)`；实际 `git branch --show-current` = `feat/event-and-outbox-contract-001-bootstrap-0430`，工作树有 12 个跨规格未提交修改（含 `Makefile`、`backend/go.mod`、3 个其它 subspec 的 plan/checklist/INDEX）+ 5 个未跟踪文件（`backend/cmd/migrate/`、`backend/internal/migrations/`、`scripts/lint/events_inventory.py` 等）。
- **影响**：`/implement` Step 4.5 `detect_session_branch.py` 正确报 `matchesSessionBranch=false`，会话被迫做出 user-driven 决策（用户选择「在当前分支实施即可」继续）。下游连锁：`/tdd` Step 9.5 phase-commit 被显式跳过（避免把跨规格脏内容批量提交进 A5 commit）；`make lint` 在 Phase 4.1 因 db-migrations 未提交代码触发 cmd/migrate 边界违规，又驱动了 A4 spec v1.7 → v1.8 的修订。

### 2.4 plan/checklist 2.2 的 NOT-YET-LANDED roster 已与现实漂移

- **证据**：checklist 2.2（写于 2026-04-29）声称 `lint-config (A4，待 A4 001 落地后切换为直接执行)`；但根 `Makefile` 的 `lint-config` target（lines 26-35）已直接调用 `lint-getenv-boundary` + `lint-env-dict` + `lint-secrets-pattern`，A4 001-bootstrap plan 状态为 `completed`（2026-04-30）。
- **影响**：本 plan 在 Phase 2 步骤里同时承担「记录占位状态」与「修订占位文本」两件事，2.2 字面文本被原地改写。这是跨 plan 时间漂移的直接体现，没有自动化机制。

### 2.5 A4 spec §4.1 边界 allow-list 没有跟上 B4 cmd/migrate 的引入

- **证据**：A4 spec v1.7 §4.1 允许 `os.Getenv` 出现在 `backend/cmd/{api,worker}/main.go`，未列出 `cmd/migrate`；B4 db-migrations-baseline 在 2026-04-30 创建了 `backend/cmd/migrate/main.go`（依赖 `os.Getenv` 读 `DATABASE_URL` / `APP_ENV` / `MIGRATE_*`）。Phase 4.1 跑 `make lint` 时 `lint-getenv-boundary` 把 `backend/internal/migrations/cli.go:28` 的 `processEnv.Getenv` 标为越界（虽然 cli.go 设计上是 DI，但 adapter 实现位置在 internal package）。
- **影响**：A5 plan 执行过程中触发了 A4 spec 的实质性修订（v1.7 → v1.8）。修订内容包括：spec §4.1 文本更新 + history.md 补行 + lint script `defaultAllowlist` 补条目。同时把 `processEnv` adapter 从 `backend/internal/migrations/cli.go` 物理迁移到 `backend/cmd/migrate/main.go`（B4 owner code）。这是典型的「实施 plan A 推动 plan B owner 文档/代码同步」事件，正常下应由 B4 owner 在创建 cmd/migrate 时主动同步 A4 spec，但 B4 plan 没有这个 cross-spec dependency 检查。

### 2.6 跨子规格的相对链接深度错误集中浮现

- **证据**：新写的 `make docs-check` 在 docs/ 下找到 38 条 broken link；过滤 `TEMPLATES.md` / HTML 注释 / inline code 后剩 12 条真实问题，分布在 4 个子规格 + 2 个 work-journal 历史条目：`shared-conventions-codified` plan/checklist（深度 4 → 5 ups 修正 3 条）、`secrets-and-config` plan（兄弟 subspec 深度 + tech-docs 深度 5 条）、`ai-gateway-and-model-routing` plan（1 条）、`local-dev-stack/spec.md` 引用一个不存在目录（1 条）、`work-journal/2026-04-26.md` 深度错位（1 条）、`work-journal/2026-04-27.md` 把 `[plans](...)` inline code 误识别为链接（修脚本而非修文档，1 条）。
- **影响**：用户在「修全部 vs. 加 allowlist vs. 暂停 plan」之间被迫做选择（最后选 A 修全部）。本 plan 的实施范围由「A5 聚合层」扩展到「A5 + 多个其它子规格相对路径修复」。

### 2.7 中途自我描述 option A1 时的逻辑错误

- **证据**：在向用户提出 option A1 时，描述包含「revert my cli.go edit (B4's design was right)」。实际正确路径是「保留 cli.go 修改 + 扩展 allow-list」（cli.go 修改把 os.Getenv 从 internal/migrations 移到 cmd/migrate，cmd/migrate 在新 allow-list 内）。错误执行了 revert，再跑 `make lint` 暴露还是 internal/migrations:28 越界，发现描述自相矛盾，又重新 re-apply。
- **影响**：1 个额外的 edit/revert/re-apply 周期。会话内一致性轻微下降，没有产生错误代码进入仓库。

### 2.8 `/tdd` Step 9.5 phase-commit 被显式跳过

- **证据**：会话决定不调用 `/work-journal --auto`，理由是工作树跨规格脏内容会被批量扫入 A5 commit（污染历史）。Step 9.5 在 Phase 1 末尾显式宣告「Skipping phase-commit due to cross-plan dirty tree」，Phase 2/3/4 同理。
- **影响**：`/tdd` 默认契约被绕过；会话结束时，整个 plan 16 项实施 + 跨规格修复仍未 commit；后续仍需用户手动决定 commit 策略（选择性 staging vs. 批量提交）。

### 2.9 L2 review 发现完成证据没有证明真实 gate 被执行

- **证据**：`/plan-code-review --fix` 后复跑 `make test` / `make build` 发现二者只打印 backend/frontend child subspec TODO；`make -n lint` 仍走不存在的 `backend/Makefile` / `frontend/Makefile` fallback；直接执行 `cd backend && golangci-lint run ./...` 失败。
- **影响**：plan/checklist 已标记 completed，但 Phase 1.1-1.3 / Phase 4.1 的通过证据不完整。已通过 v1.3 remediation 修复，并建档 [BUG-0003](../bugs/BUG-0003.md)。

## 3 根因归类

### 3.1 plan 在 quality-gate codification 之后没有自动重新校验

- **类别**：spec-plan
- **说明**：A5 plan v1.1 写于 2026-04-30 codification 之前，缺 `## 3 质量门禁分类` 是不可避免的时间漂移。`/implement` Step 4.2 已经能拦截，但拦截后用户必须显式走 `/plan-review --fix` 修补。无 repo defect，但暴露了「codification 落地后存量 plan 批量 retrofit」的工作流缺口。

### 3.2 plan 引用的辅助脚本未在仓库中存在 / 未在 plan 内被显式列为 deliverable

- **类别**：skill (`/plan-review`)
- **说明**：plan §1.4 引用 `scripts/lint/check_md_links.py` 时未声明它需要被本 plan 创建；`/plan-review` L1 没有「referenced helper scripts must exist or be flagged in scope」的检查项。

### 3.3 当前分支 + 工作树跨规格脏的状态没有在会话起始就被显著揭示

- **类别**：skill (`/implement`)
- **说明**：`gitStatus` snapshot 是会话起始时 harness 提供的；它的过期对话题影响非常大。`/implement` Step 4.5 已检测到 branch mismatch，但用户已经看了 snapshot 决定了「不切换分支」。Step 4.5 应在 Step 4.2 quality-gate check 之前先做 branch / dirty-tree 风险声明，或要求用户先确认才进入实质实施。

### 3.4 跨 plan 时间漂移（A4 已落地但 A5 plan 文本仍用过时描述）

- **类别**：no repo change needed
- **说明**：单次 transient drift；本 plan 已就地修订，下次 A4 升级时 A5 owner 再重核即可。无需自动化。

### 3.5 B4 plan 未触发 A4 spec 的 cross-spec dependency 修订

- **类别**：spec-plan + skill (`/plan-review`)
- **说明**：B4 db-migrations-baseline 在 2026-04-30 引入 `cmd/migrate` 时，A4 spec §4.1 的 boundary allow-list 应同步更新，但实际没有。`/plan-review` 没有「跨 spec 依赖检查」（plan A 修改了 spec B 拥有的接口/边界 → 必须同时引用 spec B 的修订路径）。这是一个真实 governance 缺口。

### 3.6 跨子规格相对链接深度错误集中存在

- **类别**：no repo change needed
- **说明**：本 plan 新建的 `make docs-check` gate 本身就是这个问题的修复；后续被它 lock 住。已修的 12 条不是流程缺陷，而是 W1 阶段累积的相对路径 hygiene 漂移。

### 3.7 一次性执行错误

- **类别**：no repo change needed
- **说明**：option A1 描述自相矛盾是单次 reasoning 错误；没有进入 repo；不是流程缺陷。

### 3.8 `/tdd` Step 9.5 在 dirty-cross-plan-tree 场景下没有 graceful 模式

- **类别**：skill (`/tdd`)
- **说明**：Step 9.5 当前只支持「auto-commit via /work-journal」一种路径；当工作树跨规格时只能显式跳过。可以引入「selective staging」模式（按 plan 显式声明的 deliverable 路径筛选 git diff 后再 commit），把跨规格污染挡在 commit 边界之外。

### 3.9 聚合 gate 证据缺少命令路径核对

- **类别**：spec-plan + skill (`/tdd`)
- **说明**：A5 checklist 写了 5 个 gate exit 0，但没有强制记录 `make -n` 或等价命令路径证据，导致 placeholder fallback 与真实 backend/frontend gate 混淆。L2 review 能发现问题，但更早的 `/tdd` verification 应该要求聚合 target 同时证明「会执行什么」和「实际执行结果」。

## 4 对流程资产的改进建议

### 4.1 `/plan-review` 增加「referenced helper exists or in scope」检查（high）

- **落点**：`.claude/skills/plan-review/SKILL.md` Step 6 (L1 semantic analysis) Baseline checks 部分
- **建议**：新增 `S-007` 检查：plan / checklist 中以代码块或行内代码引用的脚本路径（如 `python3 scripts/lint/check_md_links.py`）必须满足以下其一：(a) 文件存在；(b) 在某个 checklist item 显式声明「create」`scripts/lint/check_md_links.py` 作为本 plan deliverable。
- **预期效果**：避免「plan 引用未存在脚本」类未声明 deliverable，未来类似 §1.4 的隐藏工作量在 L1 阶段就被识别。

### 4.2 `/plan-review` 增加跨 spec 依赖检查（high）

- **落点**：`.claude/skills/plan-review/SKILL.md` Step 6 Baseline checks
- **建议**：新增 `S-008` 检查：plan 引入新的代码路径（特别是 `backend/cmd/<NEW_BINARY>/main.go`、`backend/internal/<NEW_PACKAGE>/`）时，必须同时检查并声明对其它 spec 中 boundary / allow-list / dictionary 类约束的影响：
  - A4 secrets-and-config §4.1 boundary allow-list
  - A4 secrets-and-config §3.1.1 env key dictionary（本次未中招，但同类）
  - B1 shared-conventions-codified `aiVocabulary` namespace（如适用）
  - B2 openapi-v1-contract `tags` / `operationIds`（如适用）
- **预期效果**：B4 plan 写到 `cmd/migrate` 时就被提示「需要 A4 spec §4.1 同步修订」，避免 A5 实施时才暴露。

### 4.3 `/implement` Step 4.5 把 dirty-cross-plan-tree 风险揭示提前到 Step 4.2 之前（medium）

- **落点**：`.claude/skills/implement/SKILL.md` Step 4 / Step 4.5
- **建议**：当前流程是 Step 4.2（quality gate）→ Step 4.5（branch resolution）。建议把「`detect_session_branch.py` + 跨 plan 脏工作树检测」提前到 Step 4 之后立即显示，并要求用户在见到完整脏文件清单后才决定是否继续。
- **预期效果**：用户在第一次决策（「不用切换分支」）时已经看到真实工作树状态，避免后续被 cmd/migrate 边界违规等连锁事件「再次惊讶」。

### 4.4 `/tdd` Step 9.5 引入「selective staging」模式（medium）

- **落点**：`.claude/skills/tdd/SKILL.md` Step 9.5
- **建议**：Step 9.5 在 dirty-cross-plan-tree 场景下，`--phase-commit` 可以从 plan 的 `context.yaml` 读 `commitPathFilter`（声明本 plan 允许 commit 的路径前缀），仅 stage 与之匹配的文件，再调用 `/work-journal --auto`。
- **预期效果**：A5 phase commit 只 stage A5 deliverable + 显式声明的跨规格修复，不再被迫整体跳过 phase-commit。

### 4.5 plan template 增加「跨 spec 影响声明」字段（low）

- **落点**：`docs/spec/TEMPLATES.md` `plan.md` 模板
- **建议**：在 `## 3 质量门禁分类` 之后新增 `## 3.1 跨 Spec 依赖与影响`（可选小节），列出本 plan 触发的其它 spec 修订路径（boundary、env key、API tag、enum、aiVocabulary 等）。配合 §4.2 的 S-008 检查使用。
- **预期效果**：plan 撰写时就明确声明「本 plan 会改 A4 §4.1」，把跨 spec 修订从「实施期才发现」前移到「设计期就声明」。

### 4.6 `/tdd` 对聚合 target 增加命令路径证据要求（medium）

- **落点**：`.claude/skills/tdd/SKILL.md` Step 7 verification guidance 或 plan checklist 模板说明
- **建议**：当 checklist item 是 Makefile / npm script / wrapper / aggregator target 时，验收证据必须包含 `make -n`、`npm run --if-present --dry-run` 或等价命令路径核对，再附真实执行结果。
- **预期效果**：避免 aggregator 只因 placeholder exit 0 被误判为真实 gate 通过。

## 5 建议优先级与后续动作

### 5.1 high 优先（值得在下一轮 implementation 之前落地）

- **§4.1 `/plan-review` S-007 referenced-helper 检查**：实施成本低（在现有 L1 review 增加一项 grep + 存在性检查），收益高（避免类似 check_md_links.py 隐藏工作量再次出现）。
- **§4.2 `/plan-review` S-008 跨 spec 依赖检查**：governance 价值最高，本次会话 60% 的额外工作（A4 spec 修订、cli.go 重构、cmd/migrate adapter 移位）都源于这一缺口。落地需要明确「关注哪些 spec 的哪些字段集」，建议以「A4 §4.1 allow-list」「A4 §3.1.1 env dict」「B1 aiVocabulary」「B2 tags / operationIds」为初始关注集，后续按 spec 演进追加。

### 5.2 medium 优先（可纳入下一个 governance plan）

- **§4.3 `/implement` 把 dirty-tree 风险揭示前置**：用户体验改进，不影响正确性。
- **§4.4 `/tdd` selective staging**：解决 phase-commit 在 dirty-cross-plan 场景下的尴尬；需要先在 `context.yaml` schema 上加 `commitPathFilter` 字段，工作量中等。
- **§4.6 `/tdd` 聚合 target 命令路径证据**：直接针对 BUG-0003，建议和下一次 TDD skill 规则修订一起做。

### 5.3 low 优先（可延后或合并到其它治理修订）

- **§4.5 plan template 增字段**：依赖 §4.2 S-008 落地后才有意义，单独做价值不大。

### 5.4 close-out 后续动作

会话结束时仍未完成的事项（user-decision required）：

- 工作树有跨 A5 / db-migrations / event-outbox 的合并修改 + 未跟踪文件，未提交。需用户决定 commit 策略：
  - **方案 1**：A5 deliverable 单独 commit（按 §4.4 思路手动 selective staging）；db-migrations / event-outbox 内容由各自 plan owner 单独 commit
  - **方案 2**：批量提交，承认这次 commit 跨 3 个 plan
- 本会话同时修订了 A4 spec v1.7 → v1.8，建议 A4 spec 修订单独 commit（独立 history table 行）以保留可追溯性
- 跨子规格的 12 条相对路径修复，建议单独 commit 或与 A4 spec 修订打包，避免混入 A5 主 commit
