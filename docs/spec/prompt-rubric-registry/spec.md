# Prompt Rubric Registry Spec

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-03

## 1 背景与目标

[engineering-roadmap spec §5.6](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 把 F3 `prompt-rubric-registry` 列为 Layer F · Quality 横切的第三份 child（依赖 [A3 `ai-gateway-and-model-routing`](./../ai-gateway-and-model-routing/spec.md) 与 [B4 `db-migrations-baseline`](./../db-migrations-baseline/spec.md)）。它从 [01-technical-architecture.md §10.1 Prompt Registry / Rubric Registry / Context Builder](../../../easyinterview-tech-docs/01-technical-architecture.md#10-ai-编排层设计) 与 [03-db-definition.md §5.8 prompt_versions / rubric_versions / ai_task_runs](../../../easyinterview-tech-docs/03-db-definition.md) 的有效历史 seed 中收敛当前 AI 调用上下文的版本管理层。

`easyinterview-tech-docs/01` 与 `03` 只保留为历史 prompt / rubric / DB 输入。当前 feature_key、prompt/rubric 坐标与 AI task 命名空间由本 spec、product-scope v1.4、B4 与后续 `config/prompts` / `config/rubrics` 编码 truth source 决定；独立 `mistake.extract`、旧错题本与旧 Growth 相关 prompt 不得作为 baseline 恢复。

[ADR-Q6 §3.6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) 已锁定：F3 只持有 `(feature_key, prompt_version, rubric_version, model_profile_name)` 四元组，不持有 provider / model 字符串（后者归 A3 Model Profile）。

本 spec 由 [001-decompose-subspecs Phase 3.5](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 锁定为 **W1 spec-contract lock**：parent phase 先固定 baseline prompt/rubric 的命名空间、文件落点、`feature_key + version` 坐标与 Resolve 调用契约。真实 baseline prompt / rubric 文件、loader 与 lint 由 F3 child `001` plan 验证；未通过前 W2 业务域不得 hardcode prompt 文本，也不得启动依赖 F3 的 AI 调用 implementation。

目标是：

1. **W1 contract 就绪**：每个 P0 AI task 至少有稳定 `feature_key + version` 坐标与文件落点；真实 baseline prompt + rubric 文件由 F3 child `001` plan 落地并验证后，W2 业务域才能引用。
2. **跨语言、跨任务、跨灰度统一**：12 个当前 baseline feature_key（见 §3.1.1）共享同一 schema（feature_key / prompt_version / rubric_version / language / template_hash）。
3. **W3 升级路径**：F3 在 W3 切到真实 Model Profile + 落地 ≥ 50 题离线评估集（不在本 spec 范围，但本 spec 锁接口）。
4. **LLM Judge 接口**：本 spec 锁定 LLM Judge 在 W3 接入的契约（不实现），让评估闭环由后续 plan 承接。

本 spec 不实现具体业务调用现场（归各 C 域）、不实现 AIClient（归 [A3](./../ai-gateway-and-model-routing/spec.md)）、不实现 DB 表（归 B4）。

## 2 范围

### 2.1 In Scope

- **prompt 真理源**：`config/prompts/<feature_key>/<version>.{yaml,md}`；YAML 元信息（feature_key / version / language / template_hash / status / created_at），Markdown 模板正文。
- **rubric 真理源**：`config/rubrics/<feature_key>/<version>.yaml`；schema：`feature_key` / `version` / `dimensions[]`（每个 dimension：`name` / `weight` / `score_levels[{label, threshold, description}]`）/ `language`。
- **DB 表 schema 引用**：`prompt_versions` / `rubric_versions` 与 [03 §5.8](../../../easyinterview-tech-docs/03-db-definition.md) 一致；本 spec 锁定字段 + index；DB 落地由 B4。
- **加载器（`internal/ai/registry/`）**：
  - `RegistryClient.GetPrompt(featureKey, version, language) → (template, meta)`
  - `RegistryClient.GetRubric(featureKey, version, language) → (schema, meta)`
  - `RegistryClient.ResolveActive(featureKey, language) → (prompt_version, rubric_version, model_profile_name)`
  - 启动时从 `config/prompts/` + `config/rubrics/` + DB 同步；DB 是 staging / prod 真理源；本地 dev 直接读文件。
- **业务调用规约**：业务代码必须先 `Resolve(featureKey, ctx.Language)` 拿到三元组，然后传给 `AIClient.Complete(profileName, payload)`；payload 中携带 `prompt_version + rubric_version + feature_key`。
- **lint 规则**：禁止业务包出现 `prompt :=` 字面量字符串 / 多行字符串模板；当前由本地 lint gate 接入，远端 CI 仅在 A5 触发条件成立后再接入；任何 prompt 必须从 registry 加载。
- **W1 contract 内容**：12 个当前 baseline feature_key 各 1 份 v0.1 baseline prompt + rubric 的坐标、schema 与落点在本 spec 中锁定；实际 `config/prompts/` / `config/rubrics/` 文件由 F3 child `001` plan 创建（prompt 文本可先是「TBD by W3 real model profile」，但 schema 必须就位）。
- **LLM Judge 接口**：`Judge(featureKey, prompt_version, output, rubric_version) → (score, reasoning)`；接口签名锁定，实现归 W3 plan。
- **灰度策略**：每个 feature_key 同时只允许 1 个 `is_active=true`；灰度切换由 PostHog feature flag（[A4 D-4](../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典)）+ `Resolve` 内部分桶逻辑实现（W3 接入）。

### 2.2 Out of Scope

- AI 调用本身：归 [A3](./../ai-gateway-and-model-routing/spec.md)。
- 业务调用现场（C4-C7 / C9 / C11 在哪一行调用）：归各 C 域。
- LLM Judge 实现：归 F3 后续 W3 plan。
- 离线评估集 ≥ 50 题：归 [001 Phase 5.5](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-5-wave-3核心业务域后端) + F3 后续 W3 plan。
- prompt / rubric 编辑 UI：当前 P0 不在范围。
- DB 表本身：归 B4。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含 12 个当前 baseline feature_key 字典）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 唯一标识 | 三元组 `(feature_key, version, language)` 是 prompt 与 rubric 的唯一坐标；version 用 SemVer（major.minor.patch） | DB 表 unique 约束已就位 |
| D-2 | 文件落点 | `config/prompts/<feature_key>/<version>.yaml`（meta） + `<version>.md`（template）；`config/rubrics/<feature_key>/<version>.yaml`；`config/` 目录由 [A4](./../secrets-and-config/spec.md) 拥有，F3 在此命名空间 | 防止散落 |
| D-3 | template_hash | `sha256(template_body + meta_canonical_json)`；自动计算，写入 yaml；本地 drift 校验 | – |
| D-4 | model_profile_name 引用 | `Resolve` 输出三元组 +「model_profile_name」（如 `practice.followup.default`），由 A3 Model Profile 定义 | F3 不持有 provider / model 字符串（与 ADR-Q6 一致） |
| D-5 | 业务调用契约 | 业务必须先 `Resolve(featureKey, ctx.Language)` 再调 `AIClient`；payload 中带三元组，便于 ai_task_runs 表写入 | 强制可追溯 |
| D-6 | language 兼容 | `language` 列允许 `multi` 表示语言无关；Resolve 优先匹配精确 language → fallback `multi` | – |
| D-7 | 灰度规则 | 同 feature_key 只允许 1 个 prompt + 1 个 rubric `is_active=true`；A/B 由 PostHog flag + Resolve 内部分桶（W3 实现）；P0 baseline 不分桶 | – |
| D-8 | 12 个当前 baseline feature_key 字典 | 见 §3.1.1；新增必须 spec 修订 | – |
| D-9 | LLM Judge 接口 | 签名锁定；W3 实现 | – |
| D-10 | 不入 log 明文 | template_body 不写入 log；只写 prompt_version + template_hash；与 [F1 D-6](../observability-stack/spec.md#31-已锁定决策含命名约定字典) / [05 §5](../../../easyinterview-tech-docs/05-logging-standard.md) 一致 | – |

#### 3.1.1 12 个当前 baseline feature_key 字典

| feature_key | 用途 | 关联业务域 | 关联 Model Profile（默认） |
|-------------|------|-----------|--------------------------|
| `target.import.parse` | JD 解析 | C4 | `target.import.default` |
| `practice.session.first_question` | 首题生成 | C5 | `practice.first_question.default` |
| `practice.session.follow_up` | 追问生成 | C5 | `practice.followup.default` |
| `practice.turn.lightweight_observe` | 同步轻量观察 | C5 | `practice.turn_observe.default` |
| `report.generate` | 整轮报告生成 | C6 | `report.generate.default` |
| `report.question_assessment` | 逐题维度评估 | C6 | `report.assessment.default` |
| `resume.parse` | 简历解析 | C7 | `resume.parse.default` |
| `resume.tailor.gap_review` | 简历 gap review | C7 | `resume.tailor.default` |
| `resume.tailor.bullet_suggestions` | 简历 bullet 改写 | C7 | `resume.tailor.default`（共享） |
| `debrief.generate` | 真实面试复现 / 复盘文本生成 | C9（P0；感谢信草稿与完整跟进建议为 C9 P1 增强） | `debrief.generate.default` |
| `embedding.upsert` | 向量 upsert | C11（P1） | `embedding.default` |
| `retrieval.rerank` | 检索 rerank | C11（P1） | `retrieval.rerank.default` |

> 备注：当前 12 项中仍包含 P1 child（C11）的 feature_key 占位，是为了让 F3 baseline 一次性把保留命名空间锁住，避免 P1 spawn 时再扩展 schema；C9 已升格为 P0 真实面试复现 / 复盘文本流，感谢信草稿与完整跟进建议仍延后到 C9 P1 增强。报告内题目回顾与本轮复练由 `report.generate` / `report.question_assessment` 承载，不再保留独立 `mistake.extract`。

### 3.2 待确认事项

- 是否在 W3 引入 prompt versioning 的语义化命名（如 `v1.0.0-baseline` / `v1.1.0-better-followup`）：默认是；具体由 W3 plan 决策。
- 是否引入 `prompt-eng` 工作流编辑器：默认 P0 不上；社区方案（Promptfoo / OpenAI Evals）由 W3 决策。
- LLM Judge 本身使用哪个 model profile：默认走 `judge.default` profile（与业务 profile 隔离），由 W3 plan 决策。
- rubric 维度名是否与 [04 §9 AI 质量指标](../../../easyinterview-tech-docs/04-metrics-observability.md#9-ai-质量指标) 对齐：默认是（追问相关率 / 报告空泛率 / 异常高分率 / 语言混乱率），由 W3 plan 决策。

## 4 设计约束

### 4.1 schema 约束

- prompt 元信息字段顺序固定（与 DB 表列顺序一致）：`feature_key / version / language / template_hash / status / created_at`；`status ∈ {draft, active, deprecated}`。
- rubric `dimensions[].name` 必须使用 [04 §9 推荐质量指标](../../../easyinterview-tech-docs/04-metrics-observability.md#92-推荐质量指标定义) 中定义的命名 +（业务域专有维度 by C 域 owner）；不允许重新发明同义维度。
- `version` 必须递增；同 `(feature_key, version)` 不允许覆盖（CI 拦截）。

### 4.2 边界约束

- F3 不直接调用 `AIClient.*`（避免循环依赖）；业务在 Resolve 之后自行调用 AIClient。
- F3 不持有 secret；DB 连接从 A4 注入。
- F3 不写入 metric / log（除自身加载状态）；AI 调用观测埋点由 A3 内部完成。

### 4.3 性能约束

- `Resolve(featureKey, language)` P95 ≤ 5ms（内存 cache + 30s TTL）。
- 启动时全量预加载 ≤ 1s（12 × 多 language baseline）。
- 文件改动后 ≤ 30s 热加载（与 A3 Model Profile 同节奏）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `config/prompts/` + `config/rubrics/` | F3 | 真理源文件 |
| `internal/ai/registry/` Go 包 | F3 | RegistryClient + Resolve 实现 |
| `prompt_versions` / `rubric_versions` 表 schema | B4 | F3 提供字段名 |
| 业务调用现场 | 各 C 域 | 通过 Resolve 三元组 + AIClient |
| LLM Judge 实现 | F3 后续 W3 plan | 接口已锁定 |
| Model Profile | A3 | F3 引用 profile name |
| 灰度（PostHog flag） | F2 + F3 | F2 owns flag；F3 owns Resolve 分桶逻辑 |
| 离线评估集 ≥ 50 | F3 后续 W3 plan + 各 C 域 | 当前不在范围 |
| 编辑 UI | 当前不在范围 | – |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 12 个 baseline 全集 | F3 后续 001 完成 | `ls config/prompts/ \| wc -l` 与 `ls config/rubrics/ \| wc -l` | 各 12 个目录；每个目录至少 1 份 v0.1 baseline | F3 后续 001 |
| C-2 | template_hash 一致 | 修改 prompt template body 但忘改 hash | CI | `lint-prompts` 失败；提示重新生成 hash | F3 后续 001 + A5 |
| C-3 | Resolve 业务调用 | C5 调用 `registry.Resolve("practice.session.follow_up", "en")` | 单测 | 返回 `(prompt_version, rubric_version, model_profile_name)` 三元组 | F3 后续 001 + C5 |
| C-4 | 业务不允许 hardcode prompt | 故意在 `internal/practice/` 中加 `prompt := "You are an interviewer..."` | CI | `lint-prompts-hardcode` 失败 | F3 后续 001 + A5 |
| C-5 | 灰度切换 | F3 自行 plan `is_active` 字段 | DB 直接修改 | 同 feature_key 旧 prompt → deprecated；新 prompt → active；Resolve 输出新 version | F3 后续 002（W3） |
| C-6 | 多 language fallback | 调 `Resolve("report.generate", "fr")`，`fr` baseline 不存在 | 加载逻辑 | 退化到 `multi` baseline；log warn | F3 后续 001 |
| C-7 | LLM Judge 接口锁定 | 编译期 | F3 包 export `Judge` 接口 | 接口签名固定（W3 实现）；业务代码可 import 抽象 | F3 后续 001 |
| C-8 | F3 executable baseline handoff | 本 spec 的 contract lock 已完成，F3 后续 `001` 完成 baseline | engineering-roadmap §5.7 W1 准入 | 12 个 baseline prompt / rubric 文件、loader 与 lint 均通过验证；依赖 F3 的 W2 implementation 可启动；parent Phase 3 只记录 spec-contract lock，不单独冒充本项已通过 | F3 后续 `001` |
| C-9 | DB 表写入闭环 | A3 调用产生 `ai_task_runs` 行 | 数据库 | `prompt_version` + `rubric_version` 字段非空，与 Resolve 输出一致 | A3 + B4 + F3 |
| C-10 | W3 升级 | F3 后续 002 完成 ≥ 50 题离线评估集 + LLM Judge | engineering-roadmap §5.7 W3 准入 | 标记 [001 Phase 5.5](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-5-wave-3核心业务域后端) 可勾选 | F3 后续 002（W3） |

## 7 关联计划

F3 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 F3 自身的 plans 承接（[engineering-roadmap §5.6](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 估算 3 plan）：

- `001-baseline`（W1 末 / W2 初）：`internal/ai/registry/` + `config/prompts/` + `config/rubrics/` 12 份 baseline + Resolve 实现 + lint 规则。
- `002-real-model-profile-and-evals`（W3）：切到真实 Model Profile + ≥ 50 题离线评估集 + LLM Judge 实现。
- `003-grayscale-and-quality-feedback`（W3 末 / W4）：PostHog 灰度分桶 + 报告页质量主观评分回流。

后续如需扩展（多模态 prompt / 函数调用 prompt schema）：递增 spec 版本，原地修订；不创建 sibling spec。
