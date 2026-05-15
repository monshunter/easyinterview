# 003 — Mode Policies and Provenance

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [backend-practice spec](../../spec.md) v1.7 §7 锁定的第三个 plan 范围落地，承接 002 的 5-kind `appendSessionEvent` state machine 与 strict-default `handleHintRequested` 占位（D-34），把 mode 策略（D-5）+ AssistantAction wire provenance（D-10）+ hint 路径 AI 失败语义（C-17 narrowing）完整闭合：

- `handleHintRequested(session, plan)` 在 002 strict-default outcome 之上按 `plan.mode` 拆两支：
  - `mode='assisted'`：调用 F3 `practice.turn.lightweight_observe` → A3 observed `AIClient.Complete` 生成 hint；返回 `200 + AssistantAction{type:'show_hint', hint, provenance}`；事务内 UPDATE `practice_turns.hint_text`；A3 observability decorator 自动写入 `ai_task_runs(task_type='hint_generate', ...)`；hint 路径不递增 `practice_sessions.turn_count` / 不计入 `question_budget` / 不发 `practice.turn.completed` outbox / 不写 `audit_events`。
  - `mode='strict'`：保留 002 已落地的 `409 PRACTICE_SESSION_CONFLICT` + `detail.policy='hint_disabled_in_mode'` + `detail.mode='strict'`；003 仅扩 mode-binding 单元测试矩阵 + cross-action provenance regression。
- mode × goal 正交：goal 取 `baseline / retry_current_round / next_round / debrief` 中任意值都不改变 hint 策略；hint 是否允许仅由 `practice_plans.mode` 决定（C-8b）。
- AssistantAction wire provenance 边界（C-12）：`show_hint` / `ask_question` / `ask_follow_up` / `session_wait` / `session_completed` 五种 action 的 `provenance` 严格只暴露 B2 `GenerationProvenance` 六个 wire 字段 (`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`)；runtime 字段 (`feature_key` / `model_profile_name` / provider / cost / latency) 仅写入 `ai_task_runs` typed columns 与 audit 摘要；002 已覆盖 ask_question / ask_follow_up，003 补 show_hint provenance + cross-action regression。
- hint 路径 AI 失败语义（D-36 plan-level 决策）：F3 `ResolveActive` 返回 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported`、A3 缺 provider secret、A3 capability mismatch、AI timeout、AI invalid output、parsed hint 为空 → **graceful degrade**：返回 `200 + AssistantAction{type:'session_wait', sessionStatus:'running'}`；service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` 只携带 sanitized B1 错误码摘要（非 wire-exposed，且不得写入 `audit_events.metadata`）；session 状态保持 `running`，**不**写 `practice_turns.hint_text`、**不**写 `failure_code`、**不**把 session 推进到 `failed`、**不**返回 502/503。F3 resolve 类失败统一映射为既有 B1 `AI_PROVIDER_CONFIG_INVALID`，不得私造 B1 之外的本地错误码。理由：hint 是 session-running 期间用户主动请求的辅助 AI；强制 fail-closed 会因临时 AI 故障中断答题循环，并与 002 D-19 follow_up fallback 模式（fallback 到 `ask_question` 占位、保 session running）形成不一致。
- 跨 spec 契约修订（按 D-30 integrator 模式延续）：
  - **backend-practice spec inline narrowing**：Phase 0 修订 `docs/spec/backend-practice/spec.md` Header `1.7 → 1.8`，§3.1 锁定决策表新增 D-36（hint AI graceful degrade narrowing）+ D-38（hint turn-lifecycle 边界）两行，§6 C-17 / §3.1 D-19 / §4.3 / §2.1 失败语义行用 inline 文字明确"session-survival AI（first_question / follow_up）" 与 "辅助 AI（hint / lightweight_observe）" 两条失败分支；同步 `history.md` 1.8 row 与 §7 row 3 描述（含 D-36 / D-37 / D-38 引用）。
  - **D-37 plan-level 决策**：B4 `migrations/000001_create_baseline.up.sql` `ai_task_runs.task_type` CHECK 扩值 `hint_generate`；`backend/internal/ai/aiclient/writers.go` 同步新增 `AITaskRunTaskHintGenerate AITaskRunCapability = "hint_generate"` 常量与 `allowedAITaskRunCapabilities` 集合；按 D-30 integrator 同步 `docs/spec/db-migrations-baseline/spec.md` Header bump + `history.md` append。Pre-launch 阶段直接改 baseline migration，与 D-21 / D-33 相同模式，不引入向后兼容 ALTER。
  - **D-38 plan-level 决策**：把 002 已隐含的 "hint 路径不写 outbox / 不递增 turn_count / 不改 turn.status" 边界提升到 spec §3.1 D-38 锁定行；003 在 assisted 接入后必须保持同一边界，并由 store/append_event 单元测试 + repository 集成测试 + `E2E.P0.048` 子断言固化。hint 仍写 `practice_session_events(kind='hint_requested')` 留痕。
  - F3 baseline `prompt-rubric-registry/001-baseline`（已 completed，不动）的 `practice.turn.lightweight_observe` 行 + `practice.turn_observe.default` model profile 在 Phase 0 preflight assert，且 Phase 2 单元测试用 fake registry 隔离真实 F3 baseline。

不含范围（留给后续 plan）：retry / next_round / debrief goal 派生与 B4 `source_debrief_id` 列（future `004-derived-plans-debrief`）；voice operation 入口（future `005-voice-turn-extension`）；DELETE /me CASCADE + 24h timeout sweep（future `006-privacy-cascade-and-cleanup`）。报告内容生成、证据评分、ReadinessTier 计算归 `backend-review` future owner。

## 2 背景

002 已交付（completed, 2026-05-13/14）：5-kind `appendSessionEvent` state machine、`SELECT FOR UPDATE` + `seq_no=MAX+1` 序列化、`completePracticeSession` 单事务 + idempotency middleware + D-35 双 key replay、B2 `PracticeTurn.status` wire enum 扩 5 值（D-33）、B3 `triggerEventSemantic: source_event_only`（D-32）、6 个 BDD 场景（`E2E.P0.038~043`）。其中 D-34 把 `hint_requested` 在 002 阶段固定为 strict-default 409，并显式标记 003 接手 assisted 分支。

003 在 spec v1.7 → v1.8 narrowing 框架内推进，不引入新的 spec D-* 主决策"创造层级"，但在 plan 层固化 3 项实施级子决策（D-36 / D-37 / D-38）并以 inline 文本 narrowing 修订 spec §6 C-17 / §3.1 D-19 / §4.3 / §2.1 失败语义文字，把"hint 是用户主动触发的辅助 AI，失败不阻断 session"以及"hint 路径在 turn 主表与事件出口上的边界"作为基本语义在 spec 中显式表达，避免让 spec §4.3 / C-17 / D-19 / §2.1 的"所有 AI 失败 → session=failed"读法把 hint 路径误推进 fail-closed，也避免让"hint 也要递增 turn_count / 发完成事件 / 写 audit"等隐式假设被引入实现。

frontend-workspace-and-practice 当前不消费 `appendSessionEvent` 的 hint 分支（hint UI 由未来 frontend plan 落地）；003 完成后 generated TS client 的 `AssistantAction.type='show_hint'` 与 `AssistantAction.hint?: string` 字段已可被前端消费，但本 plan 不要求前端 wire。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract（B4 `ai_task_runs.task_type` CHECK + writers.go enum + backend-practice spec inline narrowing）+ code-internal
- **TDD 策略**: Code plan requires TDD — 每个 implementation checklist 项 Red-Green-Refactor 入口在 `backend/internal/practice/`、`backend/internal/api/practice/`、`backend/internal/store/practice/`、`backend/internal/ai/aiclient/`、`backend/internal/ai/registry/` 下相应包；migration check 入口为 `cd backend && go test ./internal/migrations/...`（与 002 同模式：baseline 迁移契约测试已位于 `backend/internal/migrations/sql_contract_test.go`，hint_generate CHECK 用例新增到 `backend/internal/migrations/baseline_aitaskruns_check_test.go` 或扩展 `sql_contract_test.go`）+ `python3 scripts/lint/conventions_drift.py --repo-root .`；测试命令从 Go module 根执行（例如 `cd backend && go test ./internal/practice/... ./internal/api/practice/... ./internal/store/practice/... ./internal/ai/aiclient/... ./internal/migrations/... ./cmd/api/...`）；详细 phase / file / verification 映射见 [test-plan](./test-plan.md)
- **BDD 策略**: Feature plan requires BDD — 引用 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md) 中的 4 个场景 `E2E.P0.048` / `E2E.P0.049` / `E2E.P0.050` / `E2E.P0.051`；主 [checklist](./checklist.md) 在每个 user-visible behavior phase 末尾列 `BDD-Gate:` 项
- **替代验证 gate**: Phase 0 跨 spec 契约修订使用 contract test + drift check（B4 migration apply / `ai_task_runs.task_type` CHECK 接受 `hint_generate` / `migrations/enum-sources.yaml` 与 `migrations/lint.sh` 同步 / writers.go enum + `allowedAITaskRunCapabilities` 单元测试）+ legacy-negative grep（旧 hint global-disabled / 旧 mode 三值假设 / spec inline narrowing 前的"全 AI 一律 fail-closed"通用文字必须覆盖 §6 C-17 / §3.1 D-19 / §4.3 / §2.1 line 45 四处来源）+ F3 preflight assert + backend-practice / B4 spec history append 文件存在断言 + 新增 D-36 / D-38 行存在断言；Phase 4 隐私 / 观测使用 metric label allowlist + repo grep + redaction assertion 作为 gate

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `appendSessionEvent`（hint_requested 分支，复用 002 路由） | `openapi/fixtures/PracticeSessions/appendSessionEvent.json`：Phase 0 补齐 `hint-assisted-show`（assisted → 200 + show_hint）、`hint-assisted-ai-failed-degrade`（assisted + AI 失败 → 200 + session_wait）、`hint-strict-debrief-conflict`（mode=strict + goal=debrief → 409）；保留 002 已落地的 `hint-strict-conflict` / `default` / `follow-up` / `turn-skipped` / `pause-resume` / `replay` / `mismatch` / `completed` 不动 | 不要求前端实际消费 hint 分支；generated TS client schema 必须无 drift；`AssistantAction.type='show_hint'` + `AssistantAction.hint?: string` 字段在 003 完成后对前端就绪 | Phase 1-3：`backend/internal/practice.SessionEventService.handleHintRequested(session, plan)` 拆分；新增 `backend/internal/practice/hint_ai.go::applyHintAI(ctx, reservation, payload, outcome)`（沿用 002 follow_up 接入模式 + 显式 F3 / parse-failure ai_task_runs writer）；扩展 `backend/internal/store/practice/append_event.go::AppendSessionEvent` 增加 `hint_text` UPDATE 分支；补 strict / unknown 409 路径的 reserved `practice_session_events` finalization，必须写 sanitized conflict payload，禁止留下 `pending=true` stuck reservation；handler `backend/internal/api/practice/session_event_handlers.go::AppendSessionEvent` 不改入口、不挂 idempotency middleware（appendSessionEvent 仍是 clientEventId 双轨幂等） | `practice_session_events`（hint_requested kind 行；assisted success / degrade 必须 finalized；strict / unknown 409 必须 finalized 为 sanitized conflict payload，不能留下 stuck pending，且不得清理掉 D-38 要求的事件留痕）、`practice_turns.hint_text`（003 新增 UPDATE 分支；assisted 成功时填，degrade / strict 保持 NULL）、`ai_task_runs(task_type='hint_generate')`（A3 路径：A3 observability decorator 在 `AIClient.Complete` 调用后自动写 succeeded / failed 行；F3 resolve 与 parse-after-success 失败路径：`applyHintAI` 显式调用 `aiclient.AITaskRunWriter` 写一行 `validation_status='failed'` + sanitized B1 `error_code`，F3 resolve 统一用 `AI_PROVIDER_CONFIG_INVALID`，因为 Complete 未被调用或 callErr 不存在时 decorator 不会写失败行）；hint 路径不写 `audit_events`、不写 `outbox_events`（practice.turn.completed）、不递增 `practice_sessions.turn_count` / `turn_count` budget（D-38） | F3 `practice.turn.lightweight_observe`（assisted 分支，配 `practice.turn_observe.default` model profile）；strict 分支不调 AI；graceful degrade 分支：F3 失败不进 A3，A3 失败由 decorator 收尾，parse-after-success 失败由 `applyHintAI` 显式写 row，三类都进 D-36 degrade outcome | `E2E.P0.048`（assisted 主路径）、`E2E.P0.049`（strict 拒绝）、`E2E.P0.050`（provenance wire 边界）、`E2E.P0.051`（graceful degrade + 隐私 + legacy negative） |

## 3.5 Coverage Matrix

| 行 | 类别 | source | plan_phase | verification | negative_scope |
|----|------|--------|-----------|--------------|----------------|
| R1 | Primary | spec C-7（assisted hint 生效，写 hint_text） | Phase 2 | `SessionEventService.handleHintRequested(assisted)` 单元测试 + `applyHintAI` service 单元测试（fake F3 + fake AIClient 成功路径）+ repository 集成测试（UPDATE practice_turns.hint_text）+ `E2E.P0.048` | — |
| R2 | Primary | spec C-12（show_hint provenance 仅含 B2 wire 字段）+ cross-action regression | Phase 2 + Phase 3 | provenance JSON marshal 单元测试（show_hint / ask_question / ask_follow_up / session_wait / session_completed 五种 action 都断言只暴露 6 个 wire 字段）+ `E2E.P0.050` | runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency）不得出现在 wire JSON |
| R3 | Alternate | spec C-8（strict hint 拒绝，复用 002 outcome） | Phase 1 | `handleHintRequested(strict)` 单元测试（不调 AI）+ `E2E.P0.049` | strict 不静默通过 200；不写 hint_text |
| R4 | Alternate | spec C-8b + D-5（mode × goal 正交） | Phase 1 + Phase 2 | mode×goal 4×2 矩阵单元测试（baseline/retry_current_round/next_round/debrief × assisted/strict）+ `E2E.P0.048` 含 assisted+debrief 子项 + `E2E.P0.049` 含 strict+debrief 子项 | goal 不影响 hint 策略；不出现"debrief mode 等价于 hint disabled"或"baseline goal 才允许 hint"假设 |
| R5 | Failure/recovery | D-36 plan-level（hint AI 失败 graceful degrade）+ C-17 narrowing | Phase 3 | `applyHintAI` 单元测试覆盖 F3 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported` / A3 secret missing / A3 capability mismatch / A3 timeout / A3 invalid output / parsed hint empty 七路径 + handler 单元测试断言 200 + session_wait + `E2E.P0.051` | hint 失败不返回 502 / 503；session 不进入 `failed`；不写 `failure_code`；不写 `practice_turns.hint_text` |
| R6 | Boundary | spec §3.1 D-38 plan-level（hint turn-lifecycle 边界）+ spec §2.1 AssistantAction 决策树 | Phase 2 | repository 集成测试断言 hint 路径不递增 `practice_sessions.turn_count`、不发 `practice.turn.completed` outbox、不改 `practice_turns.status`、不写 `audit_events` + `E2E.P0.048` | hint 不计入 `question_budget`；不与 answer_submitted 路径在 turn 主表上互相覆盖 |
| R7 | Cross-layer contract | spec §3.1 D-37 plan-level（B4 `ai_task_runs.task_type` CHECK + `migrations/enum-sources.yaml` + writers.go enum + B4 spec history append）+ D-30 integrator mode | Phase 0 | migration apply test 断言 CHECK 接受 `hint_generate` 与拒绝 unknown 值（入口：`cd backend && go test ./internal/migrations/... -count=1`）+ `migrations/lint.sh` 断言 enum manifest 与 SQL CHECK 同步 + `backend/internal/ai/aiclient/writers_test.go` 断言 `AITaskRunTaskHintGenerate` 常量 + `allowedAITaskRunCapabilities` 集合包含 `hint_generate` + `python3 scripts/lint/conventions_drift.py --repo-root .` 通过 + B4 / backend-practice spec history append 文件存在断言 | B4 不引入向后兼容 ALTER；不增加 deprecated alias |
| R8 | Cross-layer contract | F3 `practice.turn.lightweight_observe` + `practice.turn_observe.default` profile + A3 Chat capability 匹配 | Phase 0 + Phase 2 | F3 preflight integration test（读取 baseline 行 + ResolveActive 验证）+ A3 capability mismatch 单元测试（fake profile 错配时返回 graceful degrade，不调 mock provider）+ `E2E.P0.048` 实际通过 fake AIClient 触发 F3 Resolve | F3 baseline drift（baseline 缺失或 model profile 不存在）必须由 F3 owner 先修订；不允许 003 私自 stub registry |
| R9 | Privacy/security | spec D-11 + C-16（hint_text / AI prompt / response 明文红线）+ D-37（ai_task_runs typed columns 隐私） | Phase 4 | outbox emitter unit test（hint 路径不发 outbox 自然零 hint_text）+ `practice_session_events.payload` redaction 单元测试 + ai_task_runs typed column redaction 单元测试（仅 token_count / latency_ms / model_profile_name 进入 typed columns；hint_text / prompt body / response body 零出现）+ structured log + A3 metric label allowlist + repo grep + `E2E.P0.051` | `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 在 log / metric label / audit / `practice_session_events.payload` / ai_task_runs 行 / structured log fields 中零出现 |
| R10 | Observability | spec §4.4 + A3/F1 已有边界（`feature_key` / `model_profile_name` / provider / cost / latency 仅入 ai_task_runs typed columns；metric label 命中 F1 allowlist） | Phase 2 + Phase 3 + Phase 4 | A3 调用路径：observed AIClient wiring + decorator 自动写 `ai_task_runs(task_type='hint_generate')` 行（成功 / failed 由 callErr 决定）。F3 resolve 失败路径（ResolveActive 返回 `registry.ErrPromptUnsupported` / `ErrLanguageUnsupported`）：A3 Complete 未被调用，decorator 不会触发，`applyHintAI` 必须显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun`（通过 `observability.AITaskRunRowFromMeta` 或等价 helper 构造行）写一行 `task_type='hint_generate', validation_status='failed', error_code='AI_PROVIDER_CONFIG_INVALID'`。parse-after-success 失败路径（parsed hint empty）同样必须由 `applyHintAI` 显式写 failed row，因为 decorator 已按 callErr=nil 记录了 Complete 成功行。Phase 4 metric label allowlist 单元测试 + `E2E.P0.050` 含 ai_task_runs 反查子断言 | metric label 不含 `feature_key` / prompt-rubric version / provider raw model id；ai_task_runs 行包含 `task_type='hint_generate'` + `validation_status` + `latency_ms` + `model_profile_name`；F3 resolve、A3 callErr、parse-after-success 失败路径都必须留下可区分行 |
| R11 | UX (API) | spec D-16 strict hint 拒绝码 + D-36 degrade 状态码 | Phase 1 + Phase 3 | handler 单元测试断言 strict → 409 + `ApiError{code:'PRACTICE_SESSION_CONFLICT', details:{policy:'hint_disabled_in_mode', mode}}`；assisted 成功 → 200 + `AssistantAction{type:'show_hint', hint, provenance}`；assisted 失败 → 200 + `AssistantAction{type:'session_wait', hint:null, provenance(non-AI default)}`（200 SessionEventResult 无 `details` 字段；degrade 原因仅写入 service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` + ai_task_runs.error_code，不进入 `audit_events` 或 wire envelope） | strict 不返回 422 / 500；strict 409 不留下 pending reservation；degrade 不返回 502 / 503；degrade 200 wire envelope 不携带 `details.policy` |
| R12 | Regression / legacy-negative | spec D-20 retired 术语 + 002 D-34 默认 strict fallback 不再无条件触发 + 旧 mode 三值假设 + 旧 "全 AI 一律 fail-closed" spec 文字 + `legacy debrief replay value` 取值 | Phase 4 | scoped legacy grep（扩展 `scripts/lint/backend_practice_legacy.py --phase all`）：实现 / runtime 输出范围（backend practice/API/store、openapi PracticePlans/PracticeSessions fixtures、scenario runtime assets、generated/runtime tests）对 `hint_disabled_globally` / `legacy_hint_policy` / `legacy_mode_assisted_value` / `legacy debrief replay value` / `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 零出现；负向测试文档与禁止性说明（本 plan / checklist / bdd/test docs、backend-practice spec D-20/D-21 prohibition rows）允许枚举这些字面量作为 gate 输入，但不得作为 active value / route / fixture schema 出现；backend-practice spec.md / history.md 中未区分 session-survival / 辅助 的 "全 AI 一律 session=failed" 通用文字零残留 | retired 术语不得在实现/runtime 输出中出现；文档仅允许在 negative-gate / prohibition context 中引用 |
| R13 | Out-of-scope boundary | 004 / 005 / 006 owner 范围不应被 003 实现 | Phase 1 + Phase 2 | hint mode 单元测试断言 003 不修改 goal-derived plan schema（不动 `practice_plans.source_*` 列）+ voice operation 不挂 router + CASCADE / sweep 不在 003 落地 | 003 不调用 `practice.session.first_question` / `practice.session.follow_up`（这两个 feature_key 仍由 001 / 002 路径触发） |

无 UI 视觉地理 parity 行（本 plan 不涉及 `ui-design/` 复刻；其消费方 frontend-workspace-and-practice / 未来 practice screen plan 分别承担 UI parity）。

## 3.6 L2 修订说明

待 L1 / L2 review pass 后追加。

## 4 实施步骤

### Phase 0: 跨 spec 前置修订 + Preflight

**目标**：把 D-36（hint AI graceful degrade）narrowing 落到 backend-practice spec 编码真理源；把 D-37（`hint_generate` capability）落到 B4 migration + writers.go；F3 baseline 仅 preflight assert。003 直接修订各 owner spec 编码真理源（按 D-30 Q1=A integrator 模式延续），同步更新 owner spec history.md / spec.md Header。

#### 0.1 backend-practice spec inline narrowing（D-36 + D-38 锁定 + 跨节通用文字 narrowing）

修订 `docs/spec/backend-practice/spec.md`：

1. Header `1.7 → 1.8` + 更新日期 `2026-05-14`。
2. §3.1 锁定决策表追加新行 D-36：`hint / lightweight_observe AI 失败 graceful degrade narrowing`，锁定值表述"辅助 AI（`practice.turn.lightweight_observe` / hint）失败 → `appendSessionEvent` 返回 200 + AssistantAction{type:'session_wait', hint:null, sessionStatus:'running'}；session 保持 running，不写 failure_code，不写 hint_text，不返回 502/503，不写 audit_events；service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` 只允许携带 sanitized B1 error_code，不进入 `audit_events.metadata` / wire envelope；F3 resolve 类失败映射为 `AI_PROVIDER_CONFIG_INVALID`；ai_task_runs.error_code 始终来自 B1 enum"；影响列引用 003 plan。
3. §3.1 锁定决策表追加新行 D-38：`hint turn-lifecycle 边界`，锁定值表述"hint_requested 路径在 practice_session_events 上写入 kind='hint_requested' 留痕；但不递增 practice_sessions.turn_count、不改 practice_turns.status/turn_index/follow_up_count、不发 practice.turn.completed outbox、不写 audit_events；assisted 成功路径 UPDATE practice_turns.hint_text，degrade/strict 路径 hint_text 保持 NULL；hint 不计入 question_budget"；影响列引用 003 plan。
4. §6 验收标准 C-17 行修订：把"整个 operation 返回 B1 错误（`AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_TIMEOUT` 等）并写 session=`failed` + `failure_code`；不静默回退 stub"改为"session-survival AI（first_question / follow_up）：整个 operation 返回 B1 错误并写 session=`failed` + `failure_code`；不静默回退 stub。辅助 AI（hint / lightweight_observe）：按 D-36 graceful degrade（200 + session_wait，session 保持 running，不写 failure_code，不返回 502/503），ai_task_runs 仍记录 failed row 供运维观测"。
5. §3.1 D-19 行修订：在锁定值末尾追加 "（hint / lightweight_observe AI 失败例外，按 D-36 graceful degrade）"，并把首句改为 "session-survival AI（first_question / follow_up）：..."。
6. §4.3 line 140 "AI 调用必须 fail-closed：..." 整行修订为按调用类别区分：session-survival AI 维持 fail-closed → session=`failed`；辅助 AI 按 D-36 graceful degrade，wire 200 + session_wait；运维通过 ai_task_runs(task_type='hint_generate', validation_status='failed') 观测真实失败，service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` 仅供 handler 内部 / 单元测试断言，不进入 `audit_events` 或 wire envelope。
7. §2.1 line 45 "失败语义与状态机退出：..." 整行修订：把"AI 失败 → session=`failed`"改为"session-survival AI（first_question / follow_up）失败 → session=`failed`；辅助 AI（hint / lightweight_observe）失败按 D-36 graceful degrade，session 保持 running 且不写 failure_code"。
8. §7 关联计划 row 3 描述追加 "+ D-36 plan-level（hint / lightweight_observe AI 失败 graceful degrade narrowing；同步 inline-narrow C-17 / D-19 / §4.3 / §2.1 失败语义文字）+ D-37 plan-level（B4 `ai_task_runs.task_type` CHECK 扩值 `hint_generate` pre-launch baseline rebase + `AITaskRunTaskHintGenerate` writers.go enum）+ D-38 plan-level（hint turn-lifecycle 边界）"，并把 plan 路径链接化。
9. `docs/spec/backend-practice/history.md` 追加 1.8 row：覆盖上述 narrowing 范围（§6 C-17 / §3.1 D-19 / §4.3 / §2.1）+ D-36 / D-38 新增决策行 + §7 row 3 链接与 D-36 / D-37 / D-38 引用 + 与 002 D-19 follow_up fallback 一致性说明。

验证：spec.md 与 history.md diff 通过 `make docs-check` / `sync-doc-index` 校验；spec.md 中"全 AI 一律 session=failed" / "所有 AI 调用必须 fail-closed" / "整个 operation 返回 B1 错误并把 session 置 failed"（不区分 session-survival / 辅助）通用文字在 §6 C-17 / §3.1 D-19 / §4.3 / §2.1 四个来源点零残留；§3.1 中 D-36 行与 D-38 行新增成功且 §7 row 3 描述包含 D-36 / D-37 / D-38 三个标签。

#### 0.2 B4 `ai_task_runs.task_type` CHECK 扩值 + writers.go enum（D-37）

1. 修订 `migrations/000001_create_baseline.up.sql:386` `ai_task_runs.task_type` CHECK：从 `IN ('jd_parse', 'resume_parse', 'question_generate', 'followup_generate', 'report_generate', 'resume_tailor', 'debrief_generate')` 改为 `IN ('jd_parse', 'resume_parse', 'question_generate', 'followup_generate', 'report_generate', 'resume_tailor', 'debrief_generate', 'hint_generate')`。与 D-21 / D-33 pre-launch baseline rebase 同模式，不引入 ALTER 路径。
2. 同步修订 `migrations/enum-sources.yaml` 中 `ai_task_runs.task_type` 的 values 与 checksum；`migrations/lint.sh` 必须能证明 SQL CHECK 与 enum manifest 无 drift。
3. `backend/internal/ai/aiclient/writers.go` 同步新增 `AITaskRunTaskHintGenerate AITaskRunCapability = "hint_generate"` 常量（按文件内现有顺序追加，与 `AITaskRunTaskQuestionGenerate` / `AITaskRunTaskFollowupGenerate` 同风格使用字符串字面量），并把 `AITaskRunTaskHintGenerate` 加入 `allowedAITaskRunCapabilities` 集合（文件内现有名称，非 `validTaskTypes`）。
4. 同步修订 `docs/spec/db-migrations-baseline/spec.md` Header bump（next minor version）+ `history.md` 追加一行 row："授权 backend-practice/003 Phase 0 `ai_task_runs.task_type` CHECK 扩值 `hint_generate`（pre-launch baseline rebase）；hint 路径 AI 调用持久化到 ai_task_runs typed columns，graceful degrade 时 `validation_status='failed'` 仍写一行（A3 callErr 路径由 decorator 写，F3 resolve / parse-after-success 路径由 applyHintAI 显式写）"。
5. 验证：
   - migration apply test：用 `cd backend && go test ./internal/migrations/... -count=1` 入口断言 CHECK 接受 `hint_generate` + 拒绝 `unknown_task`；具体测试文件名为 `backend/internal/migrations/baseline_aitaskruns_check_test.go`（新增）或在既有 `sql_contract_test.go` 中扩展。
   - migration lint：`migrations/lint.sh` 通过，且 negative fixture / unit test 证明漏改 `migrations/enum-sources.yaml` 会失败。
   - writers.go 单元测试：`cd backend && go test ./internal/ai/aiclient -run TestAITaskRunCapability -count=1` 断言 `AITaskRunTaskHintGenerate` 在 `allowedAITaskRunCapabilities` 集合中、`Validate()` 接受该取值。
   - `python3 scripts/lint/conventions_drift.py --repo-root .` 通过。

#### 0.3 PracticeSessions fixtures 扩展

按 §3.1 Operation Matrix 在 `openapi/fixtures/PracticeSessions/appendSessionEvent.json` 补齐 3 个命名场景，保留 002 已落地的所有命名场景不动：

- `hint-assisted-show`：`mode=assisted` 任意 goal，response = 200 + `assistantAction.type='show_hint'` + `assistantAction.hint` 非空 + `assistantAction.provenance` 含 6 个 wire 字段（`promptVersion='1.0'` / `rubricVersion='not_applicable'` / `modelId='model-profile:practice.turn_observe.default'` / `language='en'` / `featureFlag='none'` / `dataSourceVersion='1.0'`）。
- `hint-assisted-ai-failed-degrade`：`mode=assisted` 任意 goal，response = 200 + `assistantAction.type='session_wait'` + `assistantAction.hint=null` + `assistantAction.provenance` 含 non-AI default 6 个字段；`session.status='running'` 保持不变。
- `hint-strict-debrief-conflict`：`mode=strict` + `goal=debrief`，response = 409 + `error.code='PRACTICE_SESSION_CONFLICT'` + `error.details.policy='hint_disabled_in_mode'` + `error.details.mode='strict'`。

验证：`make validate-fixtures`（或 `python3 scripts/lint/validate_fixtures.py --repo-root .`）通过 + contract test 通过。

#### 0.4 F3 baseline preflight

读取 `docs/spec/prompt-rubric-registry/spec.md` v2.1（或当前版本）+ `docs/spec/prompt-rubric-registry/plans/001-baseline/checklist.md` 的 `状态: completed` 与 work-journal commit `docs(prompt-rubric-registry): close 001-baseline lifecycle and record ac self-check` 行；断言：

- `practice.turn.lightweight_observe` baseline 行存在
- 该 feature_key 关联的 model_profile_name = `practice.turn_observe.default`
- `practice.turn_observe.default` profile 在 A3 provider registry 中存在且 capability = `chat`
- F3 `RegistryClient.ResolveActive(ctx, "practice.turn.lightweight_observe", "en")` 返回非空 ResolvedPrompt（focused integration test，复用 `backend/internal/ai/registry/backend_practice_preflight_test.go` 已有 fixture）

003 不修改 F3 真理源。

#### 0.5 Phase 0 收口 gate

- backend-practice `spec.md` Header `1.7 → 1.8` + `history.md` 1.8 row 齐备；§3.1 新增 D-36 行 + D-38 行；§6 C-17 / §3.1 D-19 / §4.3 line 140 / §2.1 line 45 四处通用 fail-closed 文字均按 "session-survival AI / 辅助 AI" 拆分 narrowing；§7 row 3 描述包含 D-36 / D-37 / D-38 三个标签并把 plan 路径链接化；spec 中"全 AI 一律 session=failed" / "所有 AI 调用必须 fail-closed" / "整个 operation 返回 B1 错误并把 session 置 failed"（不区分 session-survival / 辅助）通用文字在以上 4 个来源点零残留
- B4 `db-migrations-baseline` `spec.md` Header bump + `history.md` append 齐备；`migrations/000001_create_baseline.up.sql` 中 `ai_task_runs.task_type` CHECK 已含 `hint_generate`；`migrations/enum-sources.yaml` 对应 values/checksum 已同步且 `migrations/lint.sh` 通过
- `backend/internal/ai/aiclient/writers.go` `AITaskRunTaskHintGenerate` 常量与 `allowedAITaskRunCapabilities` 集合扩值齐备；`Validate()` 接受 `hint_generate`
- `openapi/fixtures/PracticeSessions/appendSessionEvent.json` 3 个新命名场景验证通过
- F3 preflight 断言通过；本 plan 状态保持 `active`
- 收口命令：`make codegen-check`、`make validate-fixtures`、`migrations/lint.sh`、`python3 scripts/lint/conventions_drift.py --repo-root .`、`cd backend && go build ./...`、`cd backend && go test ./internal/migrations/... -count=1` 全部通过

### Phase 1: handleHintRequested mode-binding & strict 边界覆盖

**目标**：把 002 strict-default outcome 重构为 mode-binding 的纯函数式 outcome，独立可单元测试；不接入 AI / IO；不改 store 行为。

#### 1.1 handleHintRequested mode dispatch

修改 `backend/internal/practice/session_event.go::handleHintRequested(session, plan)`：

1. 读取 `plan.Mode`（B1 `sharedtypes.PracticeMode` enum：`assisted` / `strict`）。
2. `plan.Mode == sharedtypes.PracticeModeStrict` 或 unrecognized → 沿用 002 现有 409 PRACTICE_SESSION_CONFLICT outcome（`detail.policy='hint_disabled_in_mode'` + `detail.mode=plan.Mode`）；session 状态不变；`AssistantAction` 为空（沿用 002 outcome 结构）。
3. `plan.Mode == sharedtypes.PracticeModeAssisted` → 返回新的 pending outcome：`Acknowledged: true`，`AssistantAction.Type='show_hint'`，`AssistantAction.RequiresAI=true`，`AssistantAction.Hint=""`（待 service 层填入），`AssistantAction.Provenance` 为零值（待 service 层填入），`NextSessionStatus = session.Status`（保持 running），`OutboxRecord=nil`，`NextTurn=nil`，`AuditMetadata={ "event_kind": "hint_requested", "mode": "assisted" }`。
4. unknown / empty mode → strict 处理（fail-safe）；新增单元测试覆盖该兜底路径。
5. service/store 必须修复 002 先 `ReserveSessionEvent` 再返回 `outcome.Error` 的 pending reservation 风险：strict / unknown 409 返回前，reserved `practice_session_events` 必须 finalized 为 sanitized conflict payload（可 replay 同一 error），禁止留下 `payload.pending=true` 的 stuck row，且不得清理掉 D-38 要求的 `hint_requested` 事件留痕。实现方式可复用 002 replay payload 结构扩展 error variant，但不得把 request payload 或 hint 文本写进 row。

#### 1.2 strict / mode-binding 单元测试矩阵

`backend/internal/practice/session_event_test.go` 新增 `TestHandleHintRequestedModeMatrix`：

- `plan.Mode='strict'` × goal in `{baseline, retry_current_round, next_round, debrief}` 4 组 → 全部 409 + outcome.AssistantAction 为空
- `plan.Mode='assisted'` × goal in `{baseline, retry_current_round, next_round, debrief}` 4 组 → 全部 outcome.AssistantAction.RequiresAI = true 且 outcome.NextSessionStatus = session.Status
- `plan.Mode=''`（unknown）→ 409（fail-safe 默认 strict）
- `plan.Mode='legacy debrief replay value'`（旧三值假设）→ 409
- 断言 strict 分支永远不调 AI：测试用 fake registry，断言 `ResolveActive` 调用次数为 0（fake registry 内部计数器）

#### 1.3 D-38 turn lifecycle 边界单元测试

`session_event_test.go` 新增 `TestHandleHintRequestedTurnLifecycle`：

- assisted hint outcome：`OutboxRecord == nil`（不发 practice.turn.completed）
- `NextTurn == nil` 或 `NextTurn.Status == latestTurn.Status`（不改 turn status）
- `NextSessionStatus == session.Status`（session 状态不变）
- `AuditMetadata` 仅含 `event_kind` / `mode` 字段；不包含 hint_text / answer_text / prompt 明文

#### 1.4 strict 409 reservation finalization

`backend/internal/practice/append_session_event_service_test.go` 与 `backend/internal/store/practice/append_complete_test.go` 新增：

- `TestAppendSessionEventHintStrictDoesNotLeavePendingReservation`：strict / unknown 409 返回后，SQL / fake store 中同 `(session_id, client_event_id)` 存在 finalized `kind='hint_requested'` 行，payload 仅含 sanitized conflict envelope，且不存在 `payload.pending=true` stuck row。
- 同 payload + same `clientEventId` 重试仍返回同一个 sanitized 409，不重复 reserve，不调 AI；mismatch 仍返回 conflict 且不泄露首次 payload。

#### 1.5 Phase 1 收口 gate

- `cd backend && go test ./internal/practice/... -count=1` 全部通过
- handleHintRequested 单元测试覆盖率包含 8 个 goal × mode 组合 + 2 个 unknown mode 兜底
- strict / unknown 409 no-pending gate 通过
- BDD-Gate: 验证 `E2E.P0.049` 通过（strict 分支已可由 002 + Phase 1 的 strict outcome 落地；Phase 2 落地 assisted 后整体场景 048/049 串联通过）

### Phase 2: assisted hint AI 接入 + practice_turns.hint_text 写入

**目标**：实现 assisted 分支完整 AI 接入路径 + AssistantAction wire provenance + hint_text 持久化 + ai_task_runs typed column 持久化，闭合 C-7 / C-8b / C-12（show_hint 部分） / D-37 / D-38。

#### 2.1 lightweight_observe AI invocation

新增 `backend/internal/practice/hint_ai.go`：

- 新增 const `hintFeatureKey = "practice.turn.lightweight_observe"`
- 新增函数 `applyHintAI(ctx, reservation, payload, outcome)`：复用 002 `applyFollowUpAI` 的 F3 Resolve + A3 Complete + parse 模式
- prompt user message 组装：
  - 引用 `reservation.Plan.TargetJobID` / `reservation.Plan.Goal` / `reservation.Plan.InterviewerPersona`（与 follow_up 同模板渲染入口）
  - 当前 turn 的 question / partial answer 仅按 length / count 摘要（不传 question_text 明文进 prompt body 也不可能，因为 prompt body 本身就是 AI 输入；但 user message 中仅追加 `current question length: N`，不在 prompt body 中拼接 question_text 明文）
- AI 调用 metadata：
  ```go
  TaskRun: aiclient.AITaskRunContext{
      UserID:       reservation.UserID,
      Capability:   aiclient.AITaskRunTaskHintGenerate,
      ResourceType: aiclient.AITaskRunResourceTargetJob,
      ResourceID:   reservation.Plan.TargetJobID,
  }
  ```
- AI 成功：把 `outcome.AssistantAction.Hint` 写为 parsed hint 文本；`outcome.AssistantAction.Provenance` 填入 F3 resolution 字段，但 `RubricVersion` 按 D-10 "非评分动作 `rubricVersion='not_applicable'`" 硬编码为 `"not_applicable"`（不沿用 `resolution.RubricVersion`，因为 hint / show_hint 是非评分辅助动作；002 follow_up 作为评分动作仍维持 `fallbackString(resolution.RubricVersion, "not_applicable")` 不变）；其余 wire 字段沿用 resolution + fallback（与 002 `applyFollowUpAI` 同模式：`PromptVersion=fallbackString(resolution.PromptVersion, "not_applicable")`、`ModelID=modelID`、`Language=fallbackString(reservation.Session.Language, "en")`、`FeatureFlag=fallbackString(resolution.FeatureFlag, "none")`、`DataSourceVersion=fallbackString(resolution.DataSourceVersion, "not_applicable")`）；`outcome.AssistantAction.RequiresAI=false`。`ai_task_runs(task_type='hint_generate', validation_status='succeeded')` 行由 A3 observability decorator 在 `AIClient.Complete` 调用后自动写入。
- AI 失败：按 D-36 graceful degrade →（详见 Phase 3 实现）。**F3 resolve 失败路径（ResolveActive 返回 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported`）发生在 A3 Complete 调用之前，observability decorator 不会触发，applyHintAI 必须显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun`（通过 `backend/internal/ai/aiclient/observability.AITaskRunRowFromMeta` 构造行）写一行 `task_type='hint_generate', validation_status='failed', error_code='AI_PROVIDER_CONFIG_INVALID'`。A3 callErr 路径（secret_missing / timeout / invalid output / capability mismatch）由 decorator 自动写一行 `validation_status='failed' + error_code`。**parse-after-success 路径（parsed hint empty）发生在 Complete 成功之后，decorator 已写 success row，applyHintAI 必须再显式写一行 failed row 或把该解析校验前移为 `OutputSchema` validation gate；当前 A3 schema validator 不支持 `minLength`，因此本 plan 默认选择 applyHintAI 显式写 failed row。**Phase 2 已在 service 层调用 `applyHintAI`，Phase 3 在 `applyHintAI` 内部完成 degrade 路径 + 显式 writer 接入并补全单元测试矩阵。

#### 2.2 service layer integration

修改 `backend/internal/practice/append_session_event_service.go::AppendSessionEvent`：

- 在 `router.Route` 返回 outcome 之后、`s.applyFollowUpAI` 之前，检查 `outcome.AssistantAction.Type == "show_hint" && outcome.AssistantAction.RequiresAI` → 调用 `s.applyHintAI(ctx, reservation, payload, &outcome)`
- 与 follow_up 分支互斥（hint outcome 永不与 ask_follow_up 同时为 RequiresAI=true）
- 若 `s.registry == nil || s.ai == nil`（与 follow_up 同模式），调用 `s.applyHintAI` 内部 fallback 路径（Phase 3 实现）触发 graceful degrade

#### 2.3 store hint_text persistence

修改 `backend/internal/store/practice/append_event.go::AppendSessionEvent`：

- 在事务内，检测 `Outcome.AssistantAction.Type == "show_hint" && Outcome.AssistantAction.Hint != ""`：UPDATE `practice_turns SET hint_text = $1, updated_at = now() WHERE id = $2`（current turn ID 来自 `Outcome.AssistantAction.TurnID` 或 reservation.LatestTurn.ID；user_id 通过已 SELECT FOR UPDATE 的 session row 间接校验）
- hint 路径不改 `practice_turns.status`、不改 `practice_turns.turn_index`、不递增 `practice_sessions.turn_count`、不写 `outbox_events`（与 002 hint 默认行为一致：002 strict 路径返回 outcome.Error，直接经由 service.AppendSessionEvent error 退出，根本不走 store 写入路径；003 assisted 路径成功时走到 store 但 outcome.OutboxRecord = nil，store 不写 outbox 行）
- `practice_session_events` 行：与 002 一致，事务内 INSERT hint_requested kind 行（`seq_no = MAX + 1`），payload 来自 request payload；003 不改 store INSERT 路径
- graceful degrade outcome（type=session_wait）path：不写 hint_text（DB 保持 NULL）；但 hint_requested kind 的 `practice_session_events` 行仍然写入（用户事件已发生，必须留痕）

#### 2.4 wire provenance regression cross-action

`backend/internal/practice/session_event_test.go` / `backend/internal/api/practice/session_event_handlers_test.go` 新增覆盖：

- 用反射或 `encoding/json.Marshal` 把 AssistantActionProvenance 序列化；断言导出的 JSON object 严格只有 6 个 key（与 B2 `GenerationProvenance` 字段一致），且 key 名严格相等：`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`
- 任何 runtime 字段（feature_key、model_profile_name、provider、cost、latency、capability）出现在 wire JSON → 测试失败
- 测试 5 种 action type（`ask_question` / `ask_follow_up` / `show_hint` / `session_wait` / `session_completed`）的 provenance 字段集严格一致（C-12 cross-action regression）

#### 2.5 BDD-Gate Phase 2

- BDD-Gate: 验证 `E2E.P0.048` 通过（assisted hint 主路径 + goal × mode 矩阵 assisted 子集；assisted+baseline / assisted+retry / assisted+next_round / assisted+debrief 共 4 个子断言）
- BDD-Gate: 验证 `E2E.P0.049` 通过（strict hint 拒绝 + goal × mode 矩阵 strict 子集；strict+baseline / strict+retry / strict+next_round / strict+debrief 共 4 个子断言）

### Phase 3: AI failure graceful degrade & provenance wire boundary

**目标**：固化 D-36 graceful degrade 7 路径 + C-12 wire-only provenance + C-17 narrowed 边界 + handler error mapping。

#### 3.1 graceful degrade 7 路径单元测试

`backend/internal/practice/append_session_event_service_test.go::TestApplyHintAIGracefulDegradeMatrix` 覆盖 7 路径：

| 路径 | fake setup | 断言 |
|------|------------|------|
| F3 ResolveActive 返回 `registry.ErrPromptUnsupported` | fake registry 返回 unsupported / baseline 行缺失 | outcome.AssistantAction.Type = `session_wait`；hint 字段空；session 保持 running；service-local `AuditMetadata["hint_degrade_reason"]` 含 sanitized B1 `AI_PROVIDER_CONFIG_INVALID` |
| F3 ResolveActive 返回 `registry.ErrLanguageUnsupported` | fake registry 返回 language unsupported | 同上 |
| A3 Complete 返回 `AI_PROVIDER_SECRET_MISSING` | fake AIClient 返回 sharederrors.CodeAiProviderSecretMissing | 同上；ai_task_runs 行 `validation_status='failed'` |
| A3 Complete 返回 `AI_PROVIDER_TIMEOUT` | fake AIClient 模拟 timeout | 同上 |
| A3 Complete 返回 invalid JSON content | fake AIClient 返回 malformed Content | parse 失败 → 同上 |
| A3 Complete 返回 capability mismatch | fake profile 错配 | 同上；不调用 mock provider request body |
| parsed hint 为空字符串 | fake AIClient 返回 `Content: {hint: ""}` | 同上 |

每路径断言：
- `outcome.NextSessionStatus == session.Status`（保持 running）
- `outcome.AssistantAction.Type == "session_wait"`
- `outcome.AssistantAction.Hint == ""`
- `outcome.AssistantAction.Provenance` 是 non-AI default（`PromptVersion='not_applicable'` / `RubricVersion='not_applicable'` / `ModelID='model-profile:static'` / `Language=session.Language` / `FeatureFlag='none'` / `DataSourceVersion='static'`）
- store layer 在该 outcome 下不调用 hint_text UPDATE（用 fake store 计数器断言）
- ai_task_runs 行按路径分桶：
  - **A3 callErr 失败路径（secret_missing / timeout / invalid output / capability mismatch）**：由 A3 observability decorator 在包装的 `AIClient.Complete` 调用返回 callErr 时自动写一行 `task_type='hint_generate', validation_status='failed', error_code` 来自 sharederrors B1 enum；用 fake decorator-backed AIClient + fake AITaskRunWriter 断言行被写。
  - **F3 resolve 失败路径（ResolveActive 返回 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported`）**：A3 Complete 未被调用，decorator 不会触发；`applyHintAI` 必须显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun` 写一行 `task_type='hint_generate', validation_status='failed', error_code='AI_PROVIDER_CONFIG_INVALID'`。fake AITaskRunWriter 计数器断言此行被写。
  - **parse-after-success 失败路径（parsed hint empty）**：decorator 已按 Complete 成功写 succeeded row；`applyHintAI` 必须额外显式写 failed row（`error_code='AI_OUTPUT_INVALID'`）或在未来 A3 schema validator 支持非空字符串约束后改为 schema-validation callErr。本 plan 的默认实现与测试采用显式 failed row。
  - 两类失败行都不得含 raw provider message / prompt body / response body / hint_text 明文。

#### 3.2 wire provenance boundary regression

`backend/internal/practice/provenance_test.go`（新增，或扩展 `session_event_test.go`）：

- 用 `reflect.TypeOf(AssistantActionProvenance{}).NumField()` 与 `reflect.StructTag` 断言 6 个字段且 json tag 严格匹配 B2 GenerationProvenance
- 用 `json.Marshal(provenance)` 后用 `json.Unmarshal(into map[string]any)` 断言 keys 集合 == `{promptVersion, rubricVersion, modelId, language, featureFlag, dataSourceVersion}`
- 反射或代码 grep 断言 `AssistantActionProvenance` struct 中没有 `feature_key` / `model_profile_name` / `provider` / `cost` / `latency` / `capability` 字段
- 5 种 AssistantAction.Type 各跑一遍 marshal 测试（regression：未来若引入 voice 等新 type 也必须通过）

#### 3.3 handler error mapping（graceful degrade 边界）

`backend/internal/api/practice/session_event_handlers_test.go` 新增覆盖：

- assisted hint AI 失败 → handler 返回 200 + `SessionEventResult{acknowledged: true, session: {status: 'running', ...}, assistantAction: {type: 'session_wait', hint: null, provenance: <wire-only non-AI default>, sessionStatus: 'running'}}`；不是 502 / 503；200 envelope 内**不**携带 `details.policy`（200 SessionEventResult schema 无 details 字段）；degrade 原因仅写入 service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` + ai_task_runs.error_code 供运维观测，不进入 `audit_events` 或 wire envelope
- assisted hint AI 成功 → handler 返回 200 + `assistantAction: {type: 'show_hint', hint: <non-empty>, provenance: <wire-only>, sessionStatus: 'running'}`
- strict hint → handler 返回 409 + `ApiError{code: 'PRACTICE_SESSION_CONFLICT', details: {policy: 'hint_disabled_in_mode', mode: 'strict'}}`（409 ApiError envelope 才有 details）
- cross-user assisted hint 仍然 404 PRACTICE_SESSION_NOT_FOUND（cross-user isolation 与 mode 正交）
- assisted hint AI 失败时 handler 不写新 audit_events 行（append 路径不写 audit）

#### 3.4 BDD-Gate Phase 3

- BDD-Gate: 验证 `E2E.P0.050` 通过（AssistantAction provenance wire 边界 + ai_task_runs runtime 字段验证：show_hint / ask_question / ask_follow_up / session_wait / session_completed 五种 action 的 provenance 字段集严格一致 + ai_task_runs row 包含 `task_type='hint_generate'` + typed columns 完整）
- BDD-Gate: 验证 `E2E.P0.051` 通过（graceful degrade 7 路径 + 隐私红线 + legacy-negative grep；E2E 层验证主路径覆盖，子路径细分仍以单元测试为主）

### Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

**目标**：固化 D-11 / D-36 / D-37 的反查 gate；为 004 / 005 / 006 / backend-review handoff 留干净接口。

#### 4.1 Privacy / observability gates

- redaction 单元测试断言：
  - assisted hint 路径 `practice_session_events.payload` json 不含 hint_text 明文（仅 turnId 与 client 时间）
  - `ai_task_runs(task_type='hint_generate')` typed columns 仅含 `prompt_token_count` / `completion_token_count` / `latency_ms` / `validation_status` / `cost_micros` / `model_profile_name` 等已登记字段；不含 `hint_text` / AI prompt body / response body / provider secret
  - structured log fields（service / handler）不含 hint_text / AI prompt body / response body / provider secret；仅含 sessionId / turnId / mode / policy / sanitized error_code
  - A3 `ai_task_*` metric label 命中 F1 allowlist；不含 `feature_key` / prompt-rubric version / provider raw model id（A3 已固化 allowlist，003 不扩展 label surface）
  - hint 路径不写 `audit_events`（与 002 append-path 红线一致；hint 不产生 audit metadata 行）
- metric label allowlist 单元测试：A3 observed AIClient 把 `task_type='hint_generate'` 当作合法 label 值；ai_task_runs Prometheus exporter（若已实现）只 emit 已登记 label

#### 4.2 Scoped legacy-negative grep

scoped legacy grep gate（在测试或 lint 中执行，可复用 002 已有的 `scripts/lint/backend_practice_legacy.py` 扩展，命令使用 `--phase all` 或 003-specific phase）：

- 实现 / runtime 输出范围（backend practice/API/store、openapi PracticePlans/PracticeSessions fixtures、scenario runtime assets、generated/runtime tests）中 `hint_disabled_globally` / `legacy_hint_policy` / `legacy_mode_assisted_value` / `legacy debrief replay value` 零出现；`legacy debrief replay value` 若只出现在 negative fixture/test input 名称或禁止性文档说明中不算失败，但不得作为 active `PracticeMode` / schema value / route 再出现
- backend-practice spec.md / history.md 中 "全 AI 一律 session=failed" / "所有 AI 调用必须 fail-closed" 等未区分 session-survival / 辅助 的通用文字零残留（Phase 0 已 inline narrowing）；本 plan / checklist / bdd/test docs 可在 negative-gate context 中引用这些短语作为禁止性断言
- 003 实现 / runtime 输出范围中 `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 零出现；本 plan、BDD、test-plan、negative test allowlist 可枚举这些字面量作为 grep 输入

#### 4.3 Handoff doc 更新

修订 `backend/internal/api/practice/README.md`：

- 在 `## 002 Event Loop Endpoints` 段之后新增 `## 003 Mode Policies and Provenance` 段，记录：
  - `handleHintRequested(session, plan)` mode dispatch（assisted → `applyHintAI` + show_hint；strict → 409）
  - 新增 const `hintFeatureKey = "practice.turn.lightweight_observe"` 与 `AITaskRunTaskHintGenerate` 常量
  - D-36 graceful degrade 行为（assisted hint AI 失败 → 200 + `AssistantAction{type:'session_wait', hint:null, sessionStatus:'running'}`；200 SessionEventResult envelope 不携带 `details.policy`；degrade 原因仅写入 service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` + `ai_task_runs(task_type='hint_generate', validation_status='failed').error_code`，不进入 `audit_events` 或 wire envelope）
  - `practice_turns.hint_text` 写入边界（仅 assisted 成功路径写）
  - hint 路径不写 outbox / audit / 不递增 turn_count
- 修订 `## Handoff Boundaries` 段：把 003 行 "owned by 003" 改为 "delivered in 003: handleHintRequested mode dispatch + practice.turn.lightweight_observe wiring + AITaskRunTaskHintGenerate"，保留 004 / 005 / 006 / backend-review 行不变

#### 4.4 BDD-Gate Phase 4

- BDD-Gate: 验证 `E2E.P0.051` 私有 / legacy / regression 子断言全部通过（包含 ai_task_runs typed columns 隐私反查、log/metric/audit 红线、scoped legacy grep）

#### 4.5 收口 gate

- `docs/spec/backend-practice/plans/INDEX.md` 在 active 行新增 003 行（本 plan 编写阶段先保持 `active`；plan 真正完成由后续 retrospective / closure plan-review skill / sync-doc-index 推进到 completed）
- `cd backend && go test ./... -count=1` 全绿；`make codegen-check`、`make validate-fixtures`、`make lint-events`、`make codegen-events-check`、`python3 scripts/lint/conventions_drift.py --repo-root .` 全通过

## 5 验收标准

- Phase 0 ~ Phase 4 checklist 全部勾选
- 关联 BDD 场景 `E2E.P0.048` / `E2E.P0.049` / `E2E.P0.050` / `E2E.P0.051` 均由 `backend/cmd/api/practice_http_scenario_test.go` 的 HTTP scenario tests 执行通过
- backend-practice spec.md Header 1.7 → 1.8 + history.md 1.8 row 齐备；§3.1 新增 D-36 行 + D-38 行；§6 C-17 / §3.1 D-19 / §4.3 line 140 / §2.1 line 45 四处通用 fail-closed 文字按 "session-survival AI / 辅助 AI" 拆分完毕；§7 row 3 描述含 D-36 / D-37 / D-38 三标签 + plan 路径链接化
- B4 `migrations/000001_create_baseline.up.sql` `ai_task_runs.task_type` CHECK 已包含 `hint_generate`；B4 spec history append 与 Header bump commit 在 Phase 1 实施前完成
- `backend/internal/ai/aiclient/writers.go` `AITaskRunTaskHintGenerate` 常量与 `allowedAITaskRunCapabilities` 集合扩值齐备
- `cd backend && go test ./...` / `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` 全绿
- 003 范围内代码与文档中无 §3.5 R12 列出的 legacy 术语 / 旧 fail-closed 通用文字 / 旧 mode 三值假设；ai_task_runs hint_generate 行在 A3 失败路径与 F3 失败路径都被覆盖

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| backend-practice spec C-17 / D-19 inline narrowing 与 frontend / scenarios / 已有测试代码对 "AI 失败 = session=failed" 旧假设冲突 | Phase 0.1 完成后立即在 `frontend/` / `test/scenarios/` / `backend/` 反查 `session=failed.*hint` / `failure_code.*hint` 类型假设；如发现，记录到 follow-up handoff 文档并请相关 owner 修复，003 不因前端 / scenarios drift 阻塞实施；002 的 `E2E.P0.024`（AI failure 路径）针对 first_question，不与 hint 路径冲突 |
| D-36 graceful degrade 隐藏 AI 配置错误，运营难以发现 hint AI 长期失败 | Phase 2.1 / Phase 4.1 在 `ai_task_runs(task_type='hint_generate')` 行写 `validation_status='failed'` + sanitized B1 `error_code`；F1 可基于 ai_task_runs typed columns / 已登记 metric label 观察；wire 仍不暴露细节，保证用户隐私与 graceful UX |
| F3 `practice.turn.lightweight_observe` baseline drift 导致 Phase 2 测试不稳定 | Phase 0.4 preflight 在每次 003 实施 commit 前断言；Phase 2 单元测试用 fake registry 隔离真实 F3 baseline；任何 baseline drift 必须由 F3 owner 先修订再继续 003；F3 baseline 已 completed 且 002 已用相同 preflight pattern 验证稳定 |
| `practice_turns.hint_text` UPDATE 与 002 append_event 的 SELECT FOR UPDATE 并发顺序产生死锁 | Phase 2.3 在同一事务内复用 002 已 SELECT FOR UPDATE 的 `practice_sessions` 行锁；hint 路径与 answer_submitted 路径在 `practice_turns` 上写不同列（`hint_text` vs `status`/`answer_text`）；不引入额外锁；repository 集成测试用 multi-goroutine + 真实 Postgres 验证 |
| assisted hint success 若只把 hint 存在可变的 `practice_turns.hint_text`，同一 turn 多次 hint 会破坏首个 `clientEventId` replay | L2 follow-up 将 `practice_session_events.payload`（redacted event payload）与内部 `replay_payload`（client-visible result snapshot）分离；`TestSQLRepositoryReserveSessionEventReplaysOriginalHintSnapshot` 与 `TestE2EP0048PracticeHintAssistedAcrossGoals` 覆盖先请求 hint A、再请求 hint B、再 replay A 仍返回 A |
| `ai_task_runs.task_type` CHECK 扩值在 pre-launch baseline rebase 与已有部署冲突 | 当前 pre-launch 阶段直接修订 baseline migration（与 D-21 / D-33 同模式）；如本地 dev DB 已 applied baseline，需要重跑 migration apply test（开发者按 `test/scenarios/README.md` 执行 env-cleanup + env-setup）；launch 后扩值由 B4 future plan 与 ops owner 协作，本 plan 不为未来兼容路径设计 |
| handleHintRequested mode-binding 重构破坏 002 现有 `TestHandleHintRequestedDefaultsToStrict409` 等测试 | Phase 1.2 单元测试矩阵显式包含 002 已有断言（strict / unknown → 409）；测试矩阵作为 mode-binding 的 superset，确保旧测试通过 |
| AssistantAction provenance wire-only regression 测试需要 5 种 action type 反复构造 outcome，测试代码冗长 | Phase 2.4 / Phase 3.2 抽取共享 builder helper（`buildShowHintOutcome` / `buildAskQuestionOutcome` 等）；测试用表驱动方式覆盖 5 种 type；不引入新的产品代码 helper（仅测试代码） |
| F3 / parse-after-success 失败路径下 ai_task_runs 行可能因 A3 observability decorator 仅包装 `AIClient.Complete` 而漏写或误标成功，导致运维仪表盘看不到真实失败 | Phase 2.1 / 3.1 / 3.2 显式要求 `applyHintAI` 在 F3 resolve 失败时直接调用 `aiclient.AITaskRunWriter.WriteAITaskRun`（通过 `observability.AITaskRunRowFromMeta` 构造 row），并在 parsed hint empty 时额外显式写 failed row（`AI_OUTPUT_INVALID`）；checklist Phase 3.2 用 `TestApplyHintAIGracefulDegradeMatrix` 覆盖 F3 / A3 / parse 失败分桶，并由 `TestTaskRunWriterInsertsTypedColumns` 与 `E2E.P0.050` / `E2E.P0.051` 验证 typed-column writer 与 observed harness |
| show_hint 的 `rubricVersion` 若沿用 `fallbackString(resolution.RubricVersion, "not_applicable")` 会从 F3 baseline 拿到 `v0.1.0`，与 spec D-10 "非评分动作 `rubricVersion='not_applicable'`" 文字冲突 | Phase 2.1 显式硬编码 show_hint 的 `RubricVersion='not_applicable'`，不沿用 resolution；002 follow_up（评分动作）保留 fallback 不变；`provenance_test.go` 增 assertion 锁定 show_hint 此字段值 |

## 7 L2 修订记录

- 2026-05-15 `plan-code-review --fix`: 修复 strict / unknown hint 409 replay payload 的 SQL 持久化边界，只保留 sanitized `requestFingerprint + error` envelope；补充 003 scenario runtime assets 的 scoped legacy grep 覆盖；加固 `E2E.P0.048`-`E2E.P0.051` shell gate，避免 `tee` 或弱 verify 造成假绿。
- 2026-05-15 L2 follow-up: 修复 SQL-backed strict 409 replay error 优先级，以及 assisted 多 hint 同 turn 的 per-event replay 漂移；新增 B4 `practice_session_events.replay_payload` 内部 snapshot，让 `payload` 继续保持隐私红线而 `show_hint` replay 不再读取可变的 `practice_turns.hint_text`。
