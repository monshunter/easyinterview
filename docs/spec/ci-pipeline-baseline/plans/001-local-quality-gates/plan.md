# Local Quality Gates Bootstrap

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-05-04

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ci-pipeline-baseline spec §2.1](../../spec.md#21-in-scope) 锁定的 5 个本地入口 target 在仓库根 `Makefile` 上接齐：`make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`。本 plan 只是聚合层，不重写已有 owner 的 lint / generator / config check 实现，只把它们组织成单人开发阶段可重复的本地质量门禁，并显式记录从本地门禁升级为远端 CI 的 [D-5 触发条件](../../spec.md#31-已锁定决策)。

2026-05-04 原地修订范围：在已完成的本地质量门禁上补强 `make docs-check`，把 docs/spec Markdown fragment anchor 审计纳入固定 gate。该修订只覆盖 Markdown heading anchor drift，不改变 `lint` / `test` / `build` / `codegen-check` 既有语义，也不引入远端 CI。

本 plan 是 `ci-pipeline-baseline` 当前唯一的 active plan，只负责本地质量门禁聚合。当 D-5 触发条件出现（第二位长期贡献者、公开 release、付费用户上线、自动发版、回归频率过高）时，在本 spec 原地修订并新增 `002-remote-ci` 或等价 plan；远端 CI workflow、branch protection、artifact、runner secret 不得塞回本 plan。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 把 A5 列为 Layer A · Foundation 的最后一份 child；本 plan 也是当前 Foundation 收口的最后一公里。聚合层依赖：

- [A1 repo-scaffold spec D-3](../../../repo-scaffold/spec.md#31-已锁定决策) 在根 `Makefile` 上锁定的 10 个 phony target（含 `lint` / `test` / `build`），本 plan 把其中已落地占位升级为聚合实现。
- [B1 shared-conventions-codified spec §2.1](../../../shared-conventions-codified/spec.md#21-in-scope) 提供的 `make codegen-conventions` / `make lint-conventions`（错误码 / 枚举命名 / 共享类型 generator drift）。
- [B2 openapi-v1-contract plan §3 Phase 2.3](../../../openapi-v1-contract/plans/001-bootstrap/plan.md#23-make-入口) 提供的 `make codegen-openapi` / `make codegen-check`。
- [A4 secrets-and-config spec §2.1](../../../secrets-and-config/spec.md#21-in-scope) 提供的 `make lint-config`（env key 与 `.env.example` drift）。
- [F1 observability-stack 后续 plan](../../../observability-stack/spec.md) 计划提供的 metric / log lint helper（当前未落地，只保留 placeholder hook）。

每个 phase 都是可独立验证的纵向切片：Phase 1 把入口 target 串通；Phase 2 锁定 NOT-YET-LANDED owner 的占位行为边界；Phase 3 收口文档与 CI deferral 守门；Phase 4 跑 spec [§6 验收标准](../../spec.md#6-验收标准) C-1..C-7 自检并贴日志。本 plan 不引入 BDD 资产、不创建 `.github/workflows/*.yml`，也不修改任何 owner subspec 的规则语义。

## 3 质量门禁分类

- **Plan 类型**: `tooling + contract + code-internal`。本 plan 在仓库根 `Makefile` 上聚合 5 个本地质量入口 target（`make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`），调用 B1 / B2 / A4 / F1 owner 已暴露的 lint / generator / config check 与轻量脚本（`scripts/lint/check_md_links.py`）。2026-05-04 修订把 docs/spec heading fragment anchor drift 纳入 `docs-check`。属于本地工具链与契约聚合层，不引入用户可感知 UI、HTTP API 行为、业务工作流或端到端功能。
- **TDD 策略**: 必须通过 `/tdd --file docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/checklist.md --references docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md,docs/spec/ci-pipeline-baseline/spec.md --phase-commit ci-pipeline-baseline/001-local-quality-gates` 顺序执行。每个 checklist item 以本 checklist 内的 `验证:` 子句作为 Red-Green-Refactor 断言来源；涉及 sub-target 接入或占位行为的 item 必须先在缺位 / 失败状态下复现 expected output（占位 `not implemented yet:` exit 0 或 fail-fast exit 非 0），再最小实现并复跑 focused command。Phase 5 的 Red 来源是 `scripts/lint/check_md_links_test.py` 对缺失 fragment、GitHub-style slug、重复标题后缀、纯页内 anchor 和非 fragment 相对链接兼容性的断言。
- **BDD 策略**: BDD 不适用。本 plan 只在 Makefile / 文档 / 自检脚本层操作，不产生浏览器 UI、外部 API、业务工作流或端到端场景测试可观察行为，因此不创建 `bdd-plan.md` / `bdd-checklist.md`，主 checklist 也不设置 `BDD-Gate:`。
- **替代验证 gate**: 使用本地 lint + drift + smoke 组合代替 BDD：5 个聚合入口端到端跑通（Phase 4.1）；NOT-YET-LANDED 占位 vs 已落地失败穿透双向 fail-injection 自检（Phase 4.3 / 4.4）；`grep -r '\.github/workflows' .` 远端 CI 文件零存在性自检（Phase 4.2）；`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` Header / INDEX drift gate（Phase 4.5）；`python3 scripts/lint/check_md_links.py docs/spec --ignore '**/TEMPLATES.md' --check-fragments` docs/spec fragment anchor drift gate（Phase 5）。

## 4 实施步骤

### Phase 1: 入口 target 聚合（lint / test / build / docs-check / codegen-check）

#### 1.1 `make lint` 聚合

把根 `Makefile` 现有 `lint` 占位升级为聚合 target，按下列顺序串行调用：

1. `$(MAKE) lint-conventions`（B1 owner，错误码 `UPPER_SNAKE_CASE` / 枚举 `lower_snake_case` / `camelCase` JSON tag）。
2. `$(MAKE) lint-config`（A4 owner，`.env.example` 与代码 `Get*` 调用一致性）。
3. F1 metrics / log lint hook：当前未落地，调用名暂定 `lint-observability`；如目标缺失则按 [§4.2 Phase 2](#phase-2-占位与缺位行为锁定not-yet-landed-owner-输出--exit-0-边界) 输出 `not implemented yet: F1 observability-stack` 并 `exit 0`。
4. Go lint：`cd backend && golangci-lint run ./...`。
5. TS lint：`pnpm --filter @easyinterview/frontend lint`。

任一已存在 sub-target 失败必须以非 0 退出（spec D-1 / [§4.1](../../spec.md#41-本地门禁约束)）。

#### 1.2 `make test` 聚合

`test` target 调用：

1. Go 单元测试：`cd backend && go test ./...`。
2. TS 单元测试：`pnpm --filter @easyinterview/frontend test`。

AI 单元测试必须走 stub / fixtures provider（[B1 spec §2.1](../../../shared-conventions-codified/spec.md#21-in-scope) + [A3 ai-provider-and-model-routing spec](../../../ai-provider-and-model-routing/spec.md) 共同约定）；`AI_PROVIDER_*` 真实 secret 不读取，`APP_ENV=test` 路径才允许 stub（spec [§4.2](../../spec.md#42-安全与权限约束)）。

#### 1.3 `make build` 聚合

`build` target 调用：

1. 后端：`cd backend && go build ./cmd/...`。
2. 前端：`pnpm --filter @easyinterview/frontend build`。

后端 / 前端 cmd 入口未落地的子组件按 [§4.2 Phase 2](#phase-2-占位与缺位行为锁定not-yet-landed-owner-输出--exit-0-边界) 输出 `TODO: implemented by <owner>` 并 `exit 0`，与 [A1 plan §3.2.1 占位规则](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 保持一致。

#### 1.4 `make docs-check` 聚合

新增根 `docs-check` target，串行执行：

1. `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`（spec [§4.1 文档约束](../../spec.md#41-本地门禁约束) 强制项）。
2. 轻量相对链接检查：调用 `python3 scripts/lint/check_md_links.py docs`（由本 plan 新增）或等价脚本扫描 `docs/**/*.md` 内的相对链接。
3. docs/spec fragment anchor 检查：调用 `python3 scripts/lint/check_md_links.py docs/spec --ignore '**/TEMPLATES.md' --check-fragments` 或等价脚本，确认 `docs/spec/**/*.md` 中每个本地 Markdown fragment 都解析到目标文件内的 GitHub-style heading anchor；缺失 anchor 必须 exit 非 0。

任一步骤报告 drift / 失链必须 exit 非 0（spec C-5）。

#### 1.5 `make codegen-check` 聚合

把 [A1 plan §3.2.1](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 的 `codegen` placeholder 升级为聚合 codegen drift gate：

1. `$(MAKE) codegen-conventions`（B1）+ `git diff --exit-code -- backend/internal/shared/types backend/internal/shared/errors backend/internal/shared/idx frontend/src/lib/conventions frontend/src/lib/ids shared/conventions.yaml`。
2. `$(MAKE) codegen-openapi`（B2）+ `git diff --exit-code -- backend/internal/api/generated frontend/src/api/generated openapi/openapi.yaml`。

执行顺序为 `codegen-conventions` → `codegen-openapi`，与 [B2 plan §3 Phase 2.3](../../../openapi-v1-contract/plans/001-bootstrap/plan.md#23-make-入口) 保持一致；任一 generator 漂移即失败（spec C-6）。

#### 1.6 L2 remediation: `codegen-check` 纳入 B3 events/jobs drift

修复 plan-code-review finding X-L2：B3 `event-and-outbox-contract` 已落地事件/任务 generator 与 `codegen-events-check` 后，A5 顶层 `make codegen-check` 必须把 B3 event/job drift 纳入同一聚合门禁，避免开发者只跑标准本地 gate 时漏掉 `shared/events.yaml`、`shared/jobs.yaml` 及其 Go/TS/JSON Schema/baseline 生成物漂移。验证包括 `make -n codegen-check` 可见 B3 gate、实际 `make codegen-check` 通过。

### Phase 2: 占位与缺位行为锁定（NOT-YET-LANDED owner 输出 + exit 0 边界）

#### 2.1 区分「未落地」与「已落地但失败」

聚合层必须明确区分两种状态：

- **未落地**（owner subspec 的 lint / generator 当前不存在）：打印单行 `not implemented yet: <owner>` 并 `exit 0` 仅对该 sub-target 生效，后续 sub-target 继续执行。spec [§4.1](../../spec.md#41-本地门禁约束) 禁止假装通过。
- **已落地但失败**（sub-target 存在且返回非 0）：聚合层必须立即 fail-fast，整体 `make` 命令 exit 非 0；不允许吞掉错误码。

#### 2.2 当前 NOT-YET-LANDED 清单

当前 baseline 的占位映射：

| Sub-target | Owner | 当前状态 | 占位输出 |
|------------|-------|----------|----------|
| `lint-observability` | F1 observability-stack | 未落地 | `not implemented yet: F1 observability-stack` |
| `codegen-openapi` | B2 openapi-v1-contract | 已落地（参见 [B2 001-bootstrap](../../../openapi-v1-contract/plans/001-bootstrap/plan.md)） | 直接执行 |
| `codegen-conventions` | B1 shared-conventions-codified | 已落地（参见 [B1 001-bootstrap](../../../shared-conventions-codified/plans/001-bootstrap/plan.md)） | 直接执行 |
| `lint-conventions` | B1 shared-conventions-codified | 已落地 | 直接执行 |
| `lint-config` | A4 secrets-and-config | 已落地（[A4 001-bootstrap](../../../secrets-and-config/plans/001-bootstrap/plan.md)，2026-04-30 切换为直接执行） | 直接执行 `lint-getenv-boundary` + `lint-env-dict` + `lint-secrets-pattern` |

#### 2.3 缺位检测实现

在根 `Makefile` 中使用 `$(MAKE) -n <sub-target> >/dev/null 2>&1 || echo 'not implemented yet: <owner>'` 等价语义判断 sub-target 是否存在；存在但执行失败必须穿透原始 exit code，不能被占位逻辑吞没。

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

在干净仓库下依次执行 `make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`，记录每条命令 exit code 与 NOT-YET-LANDED owner 占位输出；输出贴入工作日志。spec C-2 / C-3 / C-4 / C-5 / C-6 全部成立。

#### 4.2 远端 CI 文件零存在性自检

`grep -r '\.github/workflows' .` 不应命中由本 plan 创建的 `ci.yml` / `nightly.yml` / `dependabot.yml`；文档中未出现「CI 已启用」描述（spec C-1 / C-7）。

#### 4.3 「已落地但失败」边界自检

人工注入两次错误后 revert：

- 把 B1 任一 generator 输出文件改一行：`make codegen-check` 必须 exit 非 0，提示 generated drift；revert 后 gate 恢复 clean。
- 在 B1 lint 输入中插入一个 `lower_snake_case` 错误码：`make lint` 必须 exit 非 0，提示错误码命名违规；revert 后 gate 恢复 clean。

证明聚合层不会因「占位逻辑」吞掉真实失败（spec C-2 / C-6）。

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

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 至 C-7 全部成立；占位 / drift / 失败 / 守门 4 类边界由 Phase 4 命令日志佐证；docs/spec fragment anchor drift 由 Phase 5 命令日志佐证。
- 本 plan checklist 全部勾选；Phase 4.1 与 Phase 4.3 关键命令日志贴入工作日志。
- 不出现 `.github/workflows/*.yml` 由本 plan 创建；文档不声称项目已有远端 CI。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 聚合层把 NOT-YET-LANDED owner 当作通过，掩盖了已落地 sub-target 的真实失败 | Phase 2.1 强制区分两类状态；Phase 4.3 通过「人工注入 → 命令必须失败 → revert」的双向自检，证明占位逻辑不会吞掉非 0 退出码 |
| 业务 secret（AI / DB / Redis / PostHog）被聚合 target 误读，写入日志或缓存 | spec D-4 + Phase 3.3 红线再申明；`make test` 显式约束 `APP_ENV=test` 路径走 stub provider；任何新增 sub-target 必须经 spec 修订后再接入 |
| A5 聚合层与各 owner subspec 的 Make target 命名 / 行为漂移（owner 把 `lint-conventions` 改名后 A5 没跟上） | Phase 1 在 `make help` 中列出 5 个聚合入口与对应 sub-target；任何 owner 修改 sub-target 名称必须递增本 spec / plan，并同步更新聚合层；CI deferral 期内由原作者 + 本 plan owner review 控制 |
| 因 D-5 条件未触发就提前创建 `.github/workflows/*.yml` | Phase 4.2 把「不存在 ci.yml / nightly.yml / dependabot.yml」纳入收口自检；spec [§3.2](../../spec.md#32-待确认事项) 把升级路径锁定在原地修订；任何 PR 引入 workflow 文件而未先修订 spec 必须被 owner 拒绝 |
| `make docs-check` 在 macOS / Linux 不同 shell 行为不一致导致 `/sync-doc-index --check` 误报 | Phase 1.4 强制聚合在仓库根；首次落地后在 macOS zsh 与 Linux bash 各跑一次；调用 skill 时显式 set `LC_ALL=C.UTF-8` 避免本地化输出差异 |
| Fragment anchor slug 规则与 GitHub 渲染锚点存在细微差异导致误报 | Phase 5 单元测试锁定本仓库实际使用的中文标题、标点、多 hyphen 与重复 heading 场景；实现保持最小 GitHub-style slugger，不引入 Markdown 结构格式化检查 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-04 | 1.5 | 原地修订 `make docs-check`：新增 docs/spec heading fragment anchor audit gate，补充 TDD 测试与 Makefile 集成要求。 | implement remediation |
| 2026-04-30 | 1.4 | L2 code-review remediation：顶层 `make codegen-check` 纳入 B3 event/job drift gate。 | plan-code-review --fix |
| 2026-04-29 | 1.1 | 收口 plan-review：docs-check 改为可执行 sync-doc-index 脚本 + `scripts/lint/check_md_links.py`；B1 codegen diff 覆盖 errors / idx / frontend ids；远端 CI 明确由 future `002-remote-ci` 承接。 | plan-review remediation |
| 2026-04-30 | 1.2 | 补齐 `## 3 质量门禁分类`：Plan 类型 / TDD 策略 / BDD 不适用声明 / 替代验证 gate；renumber 后续章节并修复内部 §3.2→§4.2 引用。同步 checklist 16 项 `验证:` 子句。 | implement gate remediation |
| 2026-04-30 | 1.3 | L2 code review remediation：reopen 真实 backend/frontend gate、Go lint、help owner 标签与 secret grep 证据漂移，修复后重新验证。 | plan-code-review remediation |
