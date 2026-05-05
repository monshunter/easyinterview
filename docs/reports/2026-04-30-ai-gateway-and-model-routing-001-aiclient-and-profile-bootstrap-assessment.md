# ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Claude Opus 4.7

**关联计划**: [docs/spec/ai-gateway-and-model-routing/plans/001-aiclient-and-profile-bootstrap/plan.md](../spec/ai-provider-and-model-routing/plans/001-aiclient-and-profile-bootstrap/plan.md)

## 1 复盘范围与成功证据

本次交付收口 [ai-gateway-and-model-routing spec §6](../spec/ai-provider-and-model-routing/spec.md#6-验收标准) Plan 001 全部 5 个 Phase / 17 个 checklist item，落地 `backend/internal/ai/aiclient/` 完整 Go 包：

- AIClient interface（`Complete` / `Embed` / `Stream` 三方法）+ `AICallMeta` 17 字段固定顺序 + `metaBuilder`
- 两个 Provider：`stub`（deterministic + APP_ENV 显式注入门）+ `openai_compatible`（纯 net/http + encoding/json，零厂商 SDK）
- `profile/` Model Profile YAML loader + polling reloader（5s cadence，plan §2.3 允许的 fsnotify fallback 路径）+ `Reload(ctx) error` 测试入口
- `observability/` middleware decorator：7 metric family（`Registerer` 抽象）+ 4 类日志事件 + `ai_task_runs` 行 + `audit_events` 行（hash + length + profile name 白名单）
- `aiclient.New` 启动期 fail-fast 矩阵（spec §6 C-9）
- 包级 README（stub 激活矩阵、smoke 验证、polling/fsnotify 升级路径、cmd/api / cmd/worker 接入示例）
- 2 份 fixture profile：`config/ai-profiles/{practice.followup,review.report}.default.yaml`
- 离线 mock server helper 供 E1 mock-contract-suite 复用

**通过证据**：

- `cd backend && go test ./internal/ai/aiclient/... -count=1 -race` 全绿（5 packages：`aiclient` / `profile` / `observability` / `providers/stub` / `providers/openai_compatible`，含 race detector 100 reload × 8 readers）
- spec §6 C-1 / C-2 / C-3 / C-4 / C-5 / C-6 / C-7 / C-9 共 8 项 AC 通过 targeted go test：`TestComplete_NormalChatCompletion`、`TestComplete_DeterministicForSameInput`、`TestComplete_FallbackHeadersPopulateMeta`、`TestLoaderReloadConvergesUnderHotReloadSLA`、`TestDecorator_AllSevenMetricFamiliesRegistered`、`TestDecorator_SuccessIncrementsRunsAndLogsCompleted`、`TestPrivacy_NoPlaintextLeaksAnywhere`、`TestDecorator_OutputSchemaInvalidEmitsAIOutputInvalid`、`TestNew_FailFastMatrix`（C-8 W1 review gate 由 engineering-roadmap/001-decompose-subspecs Phase 3.2 关闭，不在本 plan 范畴）
- Phase 5.3 grep 红线全部清空：`grep '"github.com/(sashabaranov/go-openai|openai/openai-go|...)"' backend/`、`grep 'openai-go|anthropic-sdk-go|...' backend/go.mod`、`grep '(payload\.Messages|m\.Content|input\.Texts|resp\.Content|response\.Content)' observability/ providers/` 全部审查；明文路径仅出现在 hash / 长度 / wire 构造的合法位置
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` → "All documents are in sync. Zero drift detected."
- 5 个 phase commit 全部按 /tdd phase-commit 协议落地：`8c049d9` → `0ba51d1` → `9810254` → `5af73d7` → `12a27cd`，对应 5 条 work-journal 条目
- plan / checklist Header 切到 `状态: completed` / `更新日期: 2026-04-30`；`docs/spec/ai-gateway-and-model-routing/plans/INDEX.md` 把 001 移到 §2 Completed

## 2 会话中的主要阻点/痛点

### 2.1 Plan 002 governance 与 user 请求范围不一致

- **证据**：用户请求 "ai-gateway-and-model-routing 001 & 002 ... 始终在当前分支实施"，但 002 在 `plans/INDEX.md` §3 Draft / Blocked，§6 Activation governance 显式禁止 `/implement` 直接进入。Phase 1 执行前停下来与用户对齐，用户选定方案 A（仅实施 001）。
- **影响**：实际不阻塞，但需要在最早阶段做一次结构化方案选择。governance 设计本身工作良好（plan-level 自禁止 + INDEX 分组双层防护），用户感知时延 1 个 turn。

### 2.2 working tree 含 A4 secrets-and-config WIP 未提交变更

- **证据**：会话开始时 `git status` 干净，但 `go vet ./...` 暴露 `backend/internal/platform/config/validator_test.go:53` `loader.Validate undefined` 编译错误；首次写文件后 `git status` 显示 `backend/go.mod`、`backend/go.sum`、`backend/internal/platform/`、`config/{config,dev,prod,staging}.yaml`、`docs/spec/secrets-and-config/plans/001-bootstrap/checklist.md` 等多项与本 plan 无关的改动。
- **影响**：每次 `git add` 都需要逐个列出我的文件路径，避免误把 A4 in-flight 编进我的 commit；如果用 `git add .` 会导致 5 个 plan-001 commit 都污染 A4 域。这增加每次 phase-commit 的 cognitive load。

### 2.3 stub.go 与 stub_test.go 在 Phase 2 期间被 lint 改写

- **证据**：Phase 2 写 openai_compatible 期间收到 system-reminder "stub.go was modified" — 改动把 `os.Getenv("APP_ENV")` 读取删除，强制要求 `stub.WithAppEnv(env)` 显式注入；改动来自 secrets-and-config A4 boundary lint。lint 把我 Phase 1 写的 4 个测试用例也一并调整。
- **影响**：发生瞬间需要切上下文核对 lint 改写是否合理（合理）、是否破坏既有测试断言（未破坏）、是否需要补哪些位点（无）。增加 1 个 turn 的 verification 成本，但 lint 改写是正确的。

### 2.4 fsnotify 与 A4 WIP go.mod 间接依赖耦合

- **证据**：plan §2.3 默认 fsnotify watcher，但 `backend/go.mod` 当前仅有 `uuid` + `yaml.v3`；working tree 中 `backend/go.mod` 已被 A4 加上 fsnotify / koanf / mapstructure 等多个间接依赖。如果我用 fsnotify 直依赖会触发 `go mod tidy` 把 A4 的整条依赖链一并调整，污染本 plan 的 commit。
- **影响**：需要做一次决策（plan 允许 polling fallback，5s cadence 满足 ≤30s SLA），并在 README + 工作日志 + 5 个 commit 文案中明确记录偏离与升级路径。决策本身合理，但 Plan-level 没有显式给出"当依赖与其它 plan 耦合时如何处理"的指引。

### 2.5 Bash 子进程 cwd 漂移

- **证据**：会话中段 Bash session 的 cwd 漂移到 `/Users/tanzhangyu/Documents/my-opensources/easyinterview/backend`（应为根目录），导致 `git add backend/internal/ai/...` 失败 fatal: pathspec ...did not match any files。defensive 加 `cd /Users/tanzhangyu/Documents/my-opensources/easyinterview &&` 前缀解决。
- **影响**：单次 stage 失败 + 1 个 retry。不严重，但反映长会话 Bash session state 不可预期。

## 3 根因归类

| 阻点 | 根因 | 类别 |
|------|------|------|
| 2.1 plan 002 范围不一致 | Plan 002 自身 governance 本来就阻止误实施；用户指令未提及 002 是 draft 是因为人类视角无法 1:1 映射 plans/INDEX 分组 | 无需仓库改动（governance 已工作） |
| 2.2 A4 WIP 工作树污染 | 多 plan 并发开发期间，工作树常态化包含他域未提交改动；当前没有 skill 或 README 强调 staging 阶段必须按 plan 边界手动列文件 | skill (`/work-journal`) + AGENTS.md |
| 2.3 stub.go 被 lint 改写 | 这是预期机制（boundary lint 跨 plan 强制约束）；偶发瞬时的协作信号，不是流程缺陷 | 无需仓库改动 |
| 2.4 fsnotify 依赖耦合 | Plan §2.3 没有给出"依赖与他域 plan 共享"的取舍指引；这是 spec / plan 缺失而非 skill / README 问题 | spec-plan |
| 2.5 Bash cwd 漂移 | Tool quirk；与 plan / skill 无关 | 无需仓库改动 |

## 4 对流程资产的改进建议

### 4.1 在 `/work-journal --auto` 或 AGENTS.md 中增加多 plan 并发的 staging 边界提示

- 当前 `/work-journal --auto` 直接信任调用方传进来的 staged files，没有提示用户当前工作树包含其它 plan 的未提交改动
- 建议在 Step 4.5（drift check）之前，新增一个轻量提示：当 `git status` 中 `?? ` 或 `M ` 文件路径与本次 staged 文件路径不同 plan / 不同 owner 时，提醒人类核对边界
- **落点**：`.agent-skills/work-journal/SKILL.md` Step 4.5 之前；或 AGENTS.md §6 协作约束章节追加一条
- **优先级**：medium

### 4.2 在 plan 模板里追加"跨 plan 共享依赖如何处理"的 §2.3 标准段

- Plan §2.3 fsnotify 的指引隐含了"如果平台不支持则降级 polling"，但实际触发降级的还有"依赖已被他域 plan 抢先添加"这一常见场景
- 建议在 plan 模板的"实施步骤"章节范例里追加一段：当 plan 引入新的 backend/go.mod 依赖时，先检查 working tree 是否已被其它 plan 修改 go.mod；如已修改则原 plan 的实施 owner 或先合入依赖、或采用 plan 内列出的 fallback；偏离必须在 plan / commit / journal 三处同步说明
- **落点**：`docs/spec/TEMPLATES.md` 的 plan 模板（如存在）；或在 `docs/spec/README.md` 增加协作小节
- **优先级**：low（本次决策已通过 README + 工作日志 + 5 个 commit 文案完整记录，未来多 plan 并发时复用方便）

### 4.3 用户请求带 draft / blocked plan 时由 `/implement` 先做范围对齐

- 本次 `/implement ai-gateway-and-model-routing 001 & 002` 是命中 governance 的输入；`/implement` Step 1 已经识别 002 是 draft，但当前 prompt 未规定遇到 draft / blocked plan 时必须暂停并向用户出三方案
- 建议 `/implement` skill 在 Step 1 plan 解析后，遇到 `状态: draft` 且 `INDEX.md §3 Draft / Blocked` 双重命中时，必须按 §4.1 / §4.4 的"必须咨询用户"协议显式停下来，给出至少两方案
- **落点**：`.agent-skills/implement/SKILL.md` Step 1 / Step 4.1
- **优先级**：medium（本次靠 governance 兜底成功，但靠的是 plan 002 自己的 §6 Activation governance；如果某 draft plan 缺这一段就会被误实施）

## 5 建议优先级与后续动作

- **下一轮最值得实施**：4.3 让 `/implement` 显式识别 draft / blocked plan 并强制范围对齐 — 这是 governance 的双层保险，与 plan 002 自身 §6 Activation governance 形成防呆
- **可同步落地**：4.1 工作日志/AGENTS 多 plan 并发 staging 边界提示 — 当前的 cognitive load 真实存在，文档化后下一个 plan 直接受益
- **可延后**：4.2 plan 模板的"跨 plan 共享依赖如何处理"标准段 — 影响面小，等下一次出现同类决策时再补也不迟
- **不再处理**：2.3 lint 跨 plan 改写、2.5 Bash cwd 漂移 — 工作机制 / 工具 quirks，不是流程缺陷

`/retrospective` 不直接修改 skill / README / AGENTS.md / spec-plan；以上建议由后续 plan 或 user-triggered remediation 选择是否落地。
