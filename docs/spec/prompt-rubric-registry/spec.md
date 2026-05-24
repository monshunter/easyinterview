# Prompt Rubric Registry Spec

> **版本**: 2.9
> **状态**: active
> **更新日期**: 2026-05-24

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 F3 `prompt-rubric-registry` 定义为当前 active Quality / AI Governance spec（依赖 [A3 `ai-provider-and-model-routing`](./../ai-provider-and-model-routing/spec.md) 与 [B4 `db-migrations-baseline`](./../db-migrations-baseline/spec.md)）。它直接承接当前 AI 调用上下文的版本管理层。

当前 feature_key、prompt/rubric 坐标与 AI task 命名空间由本 spec、product-scope 当前范围、B4 与后续 `config/prompts` / `config/rubrics` 编码 truth source 决定。F3 独立承接 `feature_key`、prompt version、rubric version、language、template hash、model profile reference、Resolve 契约、LLM Judge 接口与 prompt/rubric lint gate。

[ADR-Q6 §3.6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 已锁定：F3 只持有 `(feature_key, prompt_version, rubric_version, model_profile_name)` 四元组，不持有 provider / model 字符串（后者归 A3 Provider Registry + Model Profile）。A3 v2.8 进一步要求 F3 13 个 baseline feature_key 的默认 `model_profile_name` 必须全部能在 A3 `config/ai-profiles.yaml` catalog 中解析到合法 capability profile。A3 v2.7 打开 Tools / provider-side streaming / STT 后，F3 可在后续编码 truth source 中为 Resolve 输出追加 provider-neutral `tools[]`、`output_schema` 与 `stream_wire` hints，但这些字段不得包含 provider/model 字符串。

本 spec 历史上由 `engineering-roadmap/001-decompose-subspecs` 的 contract lock 创建；当前执行口径是固定 baseline prompt/rubric 的命名空间、文件落点、`feature_key + version` 坐标与 Resolve 调用契约。真实 baseline prompt / rubric 文件、loader 与 lint 由 F3 后续 plan 验证；未通过前，后续业务域不得 hardcode prompt 文本，也不得启动依赖 F3 的 AI 调用 implementation。

目标是：

1. **contract 就绪**：每个 P0 AI task 至少有稳定 `feature_key + version` 坐标与文件落点；真实 baseline prompt + rubric 文件由 F3 `001` plan 落地并验证后，后续业务域才能引用。
2. **跨语言、跨任务、跨灰度统一**：13 个当前 baseline feature_key（见 §3.1.1）共享同一 schema（feature_key / prompt_version / rubric_version / language / template_hash）；baseline 文件只维护 canonical `multi` 坐标，运行时 `language` 仍作为用户目标输出语言与 provenance 字段。
3. **评估升级路径**：F3 `004-real-model-profile-and-evals` 切到真实 judge Model Profile + 落地 ≥ 50 题离线评估集；本 spec 锁定接口、边界与 gate，代码实施由该 plan 承接。
4. **LLM Judge 接口**：本 spec 锁定 LLM Judge 接入契约，`004-real-model-profile-and-evals` 负责把 `NotImplementedJudge` 替换为真实实现。

本 spec 不实现具体业务调用现场（归各 C 域）、不实现 AIClient（归 [A3](./../ai-provider-and-model-routing/spec.md)）、不实现 DB 表（归 B4）。

## 2 范围

### 2.1 In Scope

- **prompt 真理源**：`config/prompts/<feature_key>/<version>.{yaml,md}` 表示 canonical `language: multi` baseline；P0 baseline 不再复制 `en` 等无语义差异语言版本。只有当某语言存在真实任务语义差异（例如地区法规、面试习惯、英文专项训练）并由 spec/plan 记录 rationale 时，才允许新增 `config/prompts/<feature_key>/<version>.<language>.{yaml,md}` override。YAML 元信息字段为 feature_key / version / language / template_hash / status / created_at，Markdown 模板正文与同名 YAML 成对存在。
- **rubric 真理源**：`config/rubrics/<feature_key>/<version>.yaml` 表示 canonical `language: multi` baseline；rubric 评估标准默认语言无关，不通过复制 `en` / `zh` rubric 文件实现用户可见本地化。只有当评估标准本身因语言/地区产生真实语义差异并由 spec/plan 记录 rationale 时，才允许新增 `config/rubrics/<feature_key>/<version>.<language>.yaml` override。schema：`feature_key` / `version` / `language` / `dimensions[]`（每个 dimension：`name` / `weight` / `score_levels[{label, threshold, description}]`）。
- **output schema 真理源**：`config/prompts/<feature_key>/<version>.schema.json` 是该 feature_key 模型输出的**语言无关** JSON Schema 子集（校验关键字限于 `type` / `required` / `properties` / `items` / `enum`，允许 `description` 作为非校验注解），multi 与所有语言变体共用同一份（JSON key 与结构语言无关，不随 language 重复抄写）；`RegistryClient` 加载后随 `ResolveActive` 输出 `output_schema`，供 A3 `CompletePayload` 在调用返回后做 fail-close 校验。prompt body 中的输出契约段必须由 schema 渲染或被 lint 证明与 schema 一致，禁止手工维护第二份字段清单；其中 example 必须是包含 schema 声明 required + optional 字段的完整代表性 JSON output，并明确要求模型返回 JSON value 而不是 JSON Schema / OpenAPI schema。schema 不持有 provider / model / endpoint / SDK 私有字符串（与 D-12 provider-neutral 边界一致）。语音（STT/TTS）feature 不产 JSON content，不在本真理源范围。
- **DB 表 schema 引用**：`prompt_versions` / `rubric_versions` 字段与 index 由本 spec 锁定；DB 落地由 B4。
- **加载器（`internal/ai/registry/`）**：
  - `RegistryClient.GetPrompt(featureKey, version, language) → (template, meta)`
  - `RegistryClient.GetRubric(featureKey, version, language) → (schema, meta)`
  - `RegistryClient.ResolveActive(featureKey, language) → (prompt_version, rubric_version, model_profile_name)`
  - 启动时从 `config/prompts/` + `config/rubrics/` + DB 同步；DB 是 staging / prod 真理源；本地 dev 直接读文件。
- **业务调用规约**：业务代码必须先 `Resolve(featureKey, ctx.Language)` 拿到三元组，然后传给 `AIClient.Complete(profileName, payload)`；payload 中携带 `prompt_version + rubric_version + feature_key`。
- **lint 规则**：禁止业务包出现 `prompt :=` 字面量字符串 / 多行字符串模板；当前由本地 lint gate 接入，远端 CI 仅在 A5 触发条件成立后再接入；任何 prompt 必须从 registry 加载。
- **contract 内容**：13 个当前 baseline feature_key 各 1 份 v0.1.0 baseline prompt + rubric 的坐标、schema 与落点在本 spec 中锁定；实际 `config/prompts/` / `config/rubrics/` 文件由 F3 `001` plan 创建，prompt 文本必须是可用 baseline 文案，不写「TBD」占位。
- **LLM Judge 接口**：`Judge(featureKey, prompt_version, output, rubric_version) → ([]score, reasoning)`；返回值自 spec v2.8（D-9）演进为**逐 rubric dimension 一项**的 `[]Score`，实现归 `004-real-model-profile-and-evals`。
- **灰度策略**：同 `(feature_key, language)` 同时只允许 1 个 `is_active=true`；P0 baseline active language 坐标为 `multi`。灰度切换由 PostHog feature flag（[A4 D-4](../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典)）+ `Resolve` 内部分桶逻辑实现（后续接入）。

### 2.2 Out of Scope

- AI 调用本身：归 [A3](./../ai-provider-and-model-routing/spec.md)。
- 模型解码期强约束（向 provider 下发 `response_format` / `json_schema` mode）与 provider 侧结构化输出：归 [A3](./../ai-provider-and-model-routing/spec.md)；F3 只产 provider-neutral `output_schema`，由 A3 决定是否向具体 provider 下发，或仅在返回后做 fail-close 校验。
- 业务调用现场（C4-C7 / C9 在哪一行调用）：归各 C 域。
- LLM Judge 实现：归 F3 `004-real-model-profile-and-evals`。
- 离线评估集 ≥ 50 题：归 F3 `004-real-model-profile-and-evals`，并在相关 backend / release workstream 进入实现前重新确认。
- prompt / rubric 编辑 UI：当前 P0 不在范围。
- DB 表本身：归 B4。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含 13 个当前 baseline feature_key 字典）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 唯一标识 | 三元组 `(feature_key, version, language)` 是 prompt 与 rubric 的唯一坐标；version 用 SemVer（major.minor.patch）；P0 baseline canonical 坐标只使用 `language=multi` | DB 表 unique 约束已就位 |
| D-2 | 文件落点 | `config/prompts/<feature_key>/<version>.yaml`（meta）+ `<version>.md`（template）与 `config/rubrics/<feature_key>/<version>.yaml` 是 `language: multi` canonical baseline 文件。语言 override 统一追加 `.<language>` 后缀，但只有存在真实任务/评估语义差异并由 spec/plan 记录 rationale 时才允许新增；禁止为英语或其他 UI 语言复制无差异 baseline。`config/` 目录由 [A4](./../secrets-and-config/spec.md) 拥有，F3 在此命名空间 | 防止散落；避免隐藏配置层重复维护 |
| D-3 | template_hash | `sha256(template_body + meta_for_hash_canonical_json)`；`meta_for_hash` 是删除 `template_hash` 字段后的 YAML meta，避免 hash 自引用；自动计算，写入 yaml；本地 drift 校验 | – |
| D-4 | model_profile_name 引用 | `Resolve` 输出三元组 +「model_profile_name」（如 `practice.followup.default`），由 A3 Model Profile 定义 | F3 不持有 provider / model 字符串（与 ADR-Q6 一致） |
| D-5 | 业务调用契约 | 业务必须先 `Resolve(featureKey, ctx.Language)` 再调 `AIClient`；payload 中带三元组，便于 ai_task_runs 表写入 | 强制可追溯 |
| D-6 | language 兼容 | `language` 列允许 `multi` 表示语言无关；Resolve 优先匹配精确 language → fallback `multi`。请求语言（如 `en` / `zh-CN` / `fr`）是模型输出语言目标与 provenance，不要求 storage 中存在同名 baseline 文件 | – |
| D-7 | 灰度规则 | 同 `(feature_key, language)` 只允许 1 个 prompt active version + 1 个 rubric active version；A/B 由 PostHog flag + Resolve 内部分桶（后续实现）；P0 baseline 只在 `multi` coordinate active，不分桶 | – |
| D-8 | 13 个当前 baseline feature_key 字典 | 见 §3.1.1；新增必须 spec 修订 | – |
| D-9 | LLM Judge 接口 | 签名按 spec 版本演进：v2.8 起返回 `([]Score, Reasoning, error)`，`[]Score` 每个 rubric dimension 一项（`Score{Dimension, Value}`，`Value ∈ [0,1]`，可按 rubric `score_levels[].threshold` 映射 label）；`Reasoning` 保留 `Summary` + `EvidenceQuotes`；`TestJudgeSignature` 断言新签名；实现归 `004` | 多维度 rubric 评估需逐维度评分；接口演进不创建 sibling，原地随 spec 版本更新 |
| D-10 | 不入 log 明文 | template_body 不写入 log；只写 prompt_version + template_hash；与 [F1 D-6](../observability-stack/spec.md#31-已锁定决策含命名约定字典) 一致 | – |
| D-11 | A3 profile coverage | §3.1.1 中每个默认 `model_profile_name` 必须在 A3 `config/ai-profiles.yaml` catalog 中存在，并能解析到合法 `capability` / `provider_ref` / `status`；P1/P2 项可为 `status=disabled` / `status=unsupported`，但必须携带 `unsupported_reason` 且不得缺命名空间；本 gate 由 `make lint-ai-profile-coverage` 和顶层 `make lint` 触发 | 防止 F3 Resolve 输出悬空 profile |
| D-12 | JD-Match feature_key cross-owner additive | backend-jobs-recommendations/001 携带 F3 spec/history additive：§3.1.1 字典从 11 升至 13，新增 `jd_match.recommendation`（默认 profile `jd_match.recommendation.default`，由 `jd_match_agent_scan` job 内联调用，每次产出 `JobMatchRecommendation` JSON 数组 + `GenerationProvenance`）+ `jd_match.search`（默认 profile `jd_match.search.default`，30s sync 调用 + 输出 ranked recommendations 数组）；baseline prompt / rubric canonical 文件为 `config/prompts/jd_match.{recommendation,search}/v0.1.0.{yaml,md}` + `config/rubrics/jd_match.{recommendation,search}/v0.1.0.yaml`，不再维护无差异 `en` 副本；recommendation/search 输出必须保留内部 jobs 池的 `jobMatchId` 以便 search handler join 已有 `jd_match_recommendations`；且明确禁止 LLM 输出引用 LinkedIn / Boss / 脉脉 / 拉勾 等外部招聘平台或私人 PII。 | backend-jobs-recommendations/001-jd-match-real-backend-baseline Phase 0.5 + L2 hardening + `003-language-coordinate-simplification` |
| D-12 | Provider-neutral AI invocation hints | 后续 F3 编码 truth source 可为 Resolve 输出追加 `tools[]`、`output_schema`、`stream_wire`，供 A3 `CompletePayload` / `Stream` 消费；字段只表达业务 schema / wire preference，不表达 provider、model、API endpoint 或 SDK 私有字段 | 让 tools / structured output / streaming handoff 可治理，同时不破坏 A3 provider-neutral 边界 |
| D-13 | output_schema 契约落地（由 002 实施） | 每个 chat feature_key 落地 `config/prompts/<feature_key>/<version>.schema.json`（**语言无关**，JSON Schema 子集校验关键字 `type`/`required`/`properties`/`items`/`enum`，允许 `description` 作为非校验注解，描述后端实际反序列化契约）；`RegistryClient` 加载并随 `ResolveActive` 输出填充 `OutputSchema`；A3 `aiclient` observability `validateOutputSchema` 扩展支持 `enum` 并对模型输出做 fail-close 校验；prompt body 输出段统一为 schema 渲染/校验的「输出契约 + complete example JSON output」，schema 是唯一字段真理源；schema `required` 必须 ⊆ prompt body 声明的输出 key，且字段必须与后端反序列化 struct 的 json tag 对齐（drift → `make lint-prompts` exit 1）；example JSON 必须覆盖 schema 声明的 required + optional 字段、使用业务形态值而非 `string` / `1` 占位，并明确禁止模型返回 JSON Schema / OpenAPI schema。允许模型产出 schema 未声明的额外字段（向后兼容），但 prompt body 不应要求后端不消费且无评估价值的字段。`response_format` 解码期强约束属 A3 后续；语音（STT/TTS）feature 不产 JSON content，不在范围 | 把 D-12 规划的 `output_schema` 从「可追加」升级为可机器校验、可渲染、可评估的锁定契约，消除 prompt 形状 ↔ struct drift（BUG-0065 类），并降低 prompt 字段清单重复维护成本，同时让 prompt 内示例成为可执行的完整 JSON 输出样例而非 schema 文档。注：`jd_match.recommendation` / `jd_match.search` 顶层为 array，其余 11 个 chat feature_key 顶层为 object |
| D-14 | language-coordinate simplification | Baseline prompt/rubric truth source 只维护 `multi` coordinate；`en` 等用户语言不再默认复制成隐藏配置文件。`ResolveActive(featureKey, requestedLanguage)` 保持 exact → `multi` fallback，prompt 通过 `{{language}}` 指令对齐用户目标语言；rubric 默认语言无关。新增 language override 必须说明真实语义差异并有 lint/seed/loader gate 覆盖 | 保留多语言用户体验，同时去掉 prompt/rubric 重复维护成本 |
| D-15 | 评估框架与 judge 决策（`004` 锁定） | Promptfoo（pnpm Node）承载离线 eval runner，经 `RegistryClient.ResolveActive` 消费同一份 `config/prompts` 真理源（禁止复制第二份 prompt 正文）；录制 fixture 为默认执行模式，真实 provider 调用经 `EVAL_LIVE` 显式 opt-in 且不进 `make test` / 默认 `make eval-offline`；LLM Judge 走 `judge.default` profile（与业务 chat profile 隔离）；评估维度复用现有 rubric `dimensions[]`，不新造同义维度；被评估的 13 个 business prompt 经 `RegistryClient` single-source 消费，而 LLM Judge 自身的评分指令模板是 eval-harness 资产，落 `config/evals/`（F3 `004` 命名空间）由 judge 从文件加载，不内联字面量、不占用 §3.1.1 业务 feature_key 坐标 | 把 spec §3.2 defer 给评估 plan 的选型固化；保证 single-source、确定性与 provider 隔离，并区分「被评估的业务 prompt」与「judge 自身指令」两类资产 |

#### 3.1.1 13 个当前 baseline feature_key 字典

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
| `debrief.suggest_questions` | 真实面试复盘问题建议 | C9（P0；用于补齐用户真实面试回忆结构） | `debrief.suggest_questions.default` |
| `jd_match.recommendation` | JD-Match AI 推荐生成（per-user agent_scan 内联） | backend-jobs-recommendations（P0） | `jd_match.recommendation.default` |
| `jd_match.search` | JD-Match 自然语言岗位搜索（同步） | backend-jobs-recommendations（P0） | `jd_match.search.default` |

> 备注：当前删除 C11 资料检索类占位，个人开发阶段先不维护对应 prompt/rubric/profile 命名空间；未来如果产品确认资料规模与质量评估需求，再由新的设计重新引入。C9 已升格为 P0 真实面试复现 / 复盘文本流，感谢信草稿与完整跟进建议仍延后到 C9 P1 增强。报告内题目回顾与本轮复练由 `report.generate` / `report.question_assessment` 承载，不再保留独立 `mistake.extract`。

### 3.2 待确认事项

- 是否引入 prompt versioning 的语义化命名（如 `v1.0.0-baseline` / `v1.1.0-better-followup`）：默认是；由 `004` 评估迭代时按需启用。
- 评估工具选型：**已由 `004`（D-15）锁定为 Promptfoo**（pnpm Node），不引入 OpenAI Evals / `prompt-eng` 编辑器；Promptfoo 经 registry 解析消费同一份 prompt 真理源。
- LLM Judge 使用哪个 model profile：**已由 `004`（D-15）锁定为 `judge.default`**（与业务 chat profile 隔离）。
- rubric 维度名是否与 F1 / F3 AI 质量指标对齐：**已由 `004`（D-15）锁定为复用现有 rubric `dimensions[]`**（如 `followup_relevance` / `report_specificity` / `score_outlier` / `language_consistency`；report 业务仍可复用自身 `report_calibration` 维度），不新造同义维度。

## 4 设计约束

### 4.1 schema 约束

- prompt 元信息字段顺序固定（与 DB 表列顺序一致）：`feature_key / version / language / template_hash / status / created_at`；`status ∈ {draft, active, deprecated}`。
- rubric `dimensions[].name` 必须使用 F1 / F3 推荐质量指标中定义的命名 +（业务域专有维度 by C 域 owner）；不允许重新发明同义维度。
- `version` 必须递增并使用 SemVer 字符串（baseline 从 `v0.1.0` 起）；同 `(feature_key, version, language)` 不允许覆盖（CI 拦截）。Baseline active 文件只要求 `language: multi`；language override 是例外路径，必须有业务 rationale。
- output schema 文件 `config/prompts/<feature_key>/<version>.schema.json` **语言无关**（每个 `(feature_key, version)` 唯一一份，multi 与各 language 变体共用），不混入 per-language `template_hash`；允许的 JSON Schema 校验关键字限于 `type` / `required` / `properties` / `items` / `enum` 子集，允许 `description` 作为非校验注解，且必须与 A3 `aiclient` 的 `outputSchema` 校验器实现保持同一校验子集；顶层 `type` 为 `object` 或 `array`（array 仅用于 `jd_match.recommendation` / `jd_match.search`）。schema `required` key 集合必须 ⊆ 对应 prompt body 声明的输出 key，并与后端反序列化 struct 的 json tag 对齐；prompt body 输出段必须由 schema 渲染或 lint 校验，三者一致性由 `make lint-prompts` 静态校验，drift 即 exit 1；example JSON 必须是完整代表性 output 值，包含 schema 声明的 optional 字段，不得退化为 OpenAPI / JSON Schema 文档或 `string` / `1` 形式的最小占位。

### 4.2 边界约束

- F3 不直接调用 `AIClient.*`（避免循环依赖）；业务在 Resolve 之后自行调用 AIClient。
- F3 不持有 secret；DB 连接从 A4 注入。
- F3 不写入 metric / log（除自身加载状态）；AI 调用观测埋点由 A3 内部完成。
- F3 只产 provider-neutral `output_schema`，不向任何 provider 下发 `response_format` / `json_schema` 请求字段（解码期强约束归 A3）；schema 内不含 provider / model / endpoint / SDK 私有字符串。

### 4.3 性能约束

- `Resolve(featureKey, language)` P95 ≤ 5ms（内存 cache + 30s TTL）。
- 启动时全量预加载 ≤ 1s（13 × canonical `multi` baseline + 可选 language override）。
- 文件改动后 ≤ 30s 热加载（与 A3 Model Profile 同节奏）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `config/prompts/` + `config/rubrics/` | F3 | 真理源文件 |
| `config/prompts/<feature_key>/<version>.schema.json` | F3 | output schema 真理源（语言无关） |
| `config/evals/` | F3 `004` | 离线评估用例（≥50 题）+ LLM Judge 评分指令模板（eval 资产，非业务 prompt，不占用 §3.1.1 业务 feature_key 坐标） |
| `internal/ai/registry/` Go 包 | F3 | RegistryClient + Resolve 实现 |
| `validateOutputSchema`（模型输出 fail-close 校验 + 可选 `response_format` 下发） | A3（`aiclient` observability） | F3 提供 schema，A3 执行校验与下发决策 |
| `prompt_versions` / `rubric_versions` 表 schema | B4 | F3 提供字段名 |
| 业务调用现场 | 各 C 域 | 通过 Resolve 三元组 + AIClient |
| LLM Judge 实现 | F3 `004` | 接口 v2.8 演进为逐维度 `[]Score`；`LLMJudge` 经 `judge.default` 实现归 `004` |
| Model Profile | A3 | F3 引用 profile name |
| 灰度（PostHog flag） | F2 + F3 | F2 owns flag；F3 owns Resolve 分桶逻辑 |
| 离线评估集 ≥ 50 | F3 `004`（Promptfoo）+ 各 C 域 | 归 `004`；registry single-source；录制 fixture 默认 + `EVAL_LIVE` opt-in |
| 编辑 UI | 当前不在范围 | – |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 13 个 baseline 全集 | F3 后续 001 完成 + 003 收敛 | `find config/prompts -mindepth 1 -maxdepth 1 -type d -print` 与 `find config/rubrics -mindepth 1 -maxdepth 1 -type d -print` | 各输出 13 个 feature_key 目录；`README.md` 等根级说明文件不计入；每个目录恰有 canonical `v0.1.0` `multi` baseline（prompt yaml+md、rubric yaml）和 1 份语言无关 output schema；无默认 `v0.1.0.en.*` 副本 | F3 `001` + `002` + `003` |
| C-2 | template_hash 一致 | 修改 prompt template body 但忘改 hash | CI | `lint-prompts` 失败；提示重新生成 hash | F3 后续 001 + A5 |
| C-3 | Resolve 业务调用 | C5 调用 `registry.Resolve("practice.session.follow_up", "en")` | 单测 | 返回 `(prompt_version, rubric_version, model_profile_name)` 三元组 | F3 后续 001 + C5 |
| C-4 | 业务不允许 hardcode prompt | 故意在 `internal/practice/` 中加 `prompt := "You are an interviewer..."` | CI | `lint-prompts-hardcode` 失败 | F3 后续 001 + A5 |
| C-5 | 灰度切换 | F3 自行 plan `is_active` 字段 | DB 直接修改 | 同 `(feature_key, language)` 旧 prompt → deprecated；新 prompt → active；Resolve 输出新 version | F3 `005` |
| C-6 | 多 language fallback | 调 `Resolve("report.generate", "fr")`，`fr` baseline 不存在 | 加载逻辑 | 退化到 `multi` baseline；log warn | F3 后续 001 |
| C-7 | LLM Judge 接口（v2.8 演进） | 编译期 | F3 包 export `Judge` 接口 | 接口按 spec 版本演进：返回 `([]Score, Reasoning, error)`，`[]Score` 逐 rubric dimension；`TestJudgeSignature` 断言新签名；业务代码可 import 抽象 | F3 `001`（接口）+ `004`（演进+实现） |
| C-8 | F3 executable baseline handoff | 本 spec 的 contract lock 已完成，F3 后续 `001` 完成 baseline | active spec 关系已保留 | 13 个 baseline prompt / rubric 文件、loader 与 lint 均通过验证；依赖 F3 的后续 implementation 可启动；roadmap 只保留 active spec 关系，不单独冒充本项已通过 | F3 后续 `001` |
| C-9 | DB 表写入闭环 | A3 调用产生 `ai_task_runs` 行 | 数据库 | `feature_key` + `prompt_version` + `rubric_version` + `feature_flag` + `data_source_version` typed 字段非空；其中 feature/prompt/rubric/data-source 与 Resolve / CallMetadata 输出一致，flag 无分桶时写 `none` | A3 + B4 + F3 |
| C-10 | 评估升级 | F3 `004` 完成 ≥ 50 题离线评估集 + 真实 LLM Judge | 运行 `make eval-offline`（录制 fixture 默认）/ `EVAL_LIVE=1` opt-in | 评估集 ≥ 50、`LLMJudge` 逐维度产出、`judge.default` active 与录制/ live 执行模式均已验证 | F3 `004` |
| C-11 | A3 profile coverage | A3 003 完成 provider registry + capability profile catalog | 运行 `make lint-ai-profile-coverage` 或顶层 `make lint` | §3.1.1 的默认 `model_profile_name` 全部存在于 `config/ai-profiles.yaml`，且 capability / provider_ref / status 合法；`disabled` / `unsupported` profile 必须显式标记并携带 `unsupported_reason` | A3 003 + F3 后续 001 |
| C-12 | output_schema 契约闭环 | F3 002 完成 13 个 chat feature_key 的 `<version>.schema.json` 与 resolver 接线 | 运行 `make lint-prompts` + `go test ./backend/internal/ai/registry/...` + `go test ./backend/internal/ai/aiclient/...` | 每个 chat feature_key 有 1 份语言无关 schema；`ResolveActive` 输出非空 `OutputSchema`；prompt body 输出段可由 schema 重新渲染，且 complete example JSON output 覆盖 schema 声明的 required + optional 字段、使用业务形态值、通过 schema 校验、明确不是 JSON Schema / OpenAPI schema；故意让 prompt 输出 key 与 schema/struct 不一致 → `make lint-prompts` 失败；`validateOutputSchema` 对违反 `enum` 或缺 required 的模型输出 fail-close（`AI_OUTPUT_INVALID`） | F3 002 |
| C-13 | language-coordinate 收敛 | F3 003 完成 | 运行 prompt/rubric lint、registry tests、seed coverage、migration check 与 stale-contract grep | `config/prompts` / `config/rubrics` 无默认 `v0.1.0.en.*`；seed migration 无 active `en` rows；loader snapshot 13 个 baseline coordinates；`ResolveActive(featureKey, "en")` / unknown locale fallback 到 `multi` 且返回同一语言无关 schema；active README/spec/lint/test 不再正向要求 `multi + en` 或 `>=2 language coordinates` | F3 003 |
| C-14 | judge.default 激活 + profile coverage | F3 `004` 翻 judge.default active | 运行 `make lint-ai-profile-coverage` | `judge.default` `status=active` 且 default provider_ref 非 placeholder；§3.1.1 的 13 个 chat profile 全部解析到非 placeholder provider；负向 placeholder → gate fail；A3 已 active 业务 profile status 不被翻动 | F3 `004` + A3 |
| C-15 | eval prompt single-source | F3 `004` Promptfoo 接线 | 运行 eval drift gate + `make lint-prompts-hardcode` | Promptfoo 经 `RegistryClient.ResolveActive` 消费同一份 prompt；无第二份 prompt 副本；prompt 漂移 → drift gate exit 1；`lint-prompts-hardcode` 仍 green | F3 `004` |

## 7 关联计划

F3 当前 active impl plan 由 F3 自身的 plans 承接（[engineering-roadmap §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 保留该 active spec）：

- `001-baseline`：`internal/ai/registry/` + `config/prompts/` + `config/rubrics/` 13 个 feature_key 的 baseline truth source + Resolve 实现 + lint 规则。
- `002-output-schema-contract`：13 个 chat feature_key 各落地**语言无关** `config/prompts/<feature_key>/<version>.schema.json` + schema 渲染/校验的 prompt body 输出契约（字段表 + complete example JSON output，非 JSON Schema / OpenAPI schema）；`RegistryClient` 加载 schema 并接线 `ResolveActive` 的 `OutputSchema`；A3 `aiclient` `validateOutputSchema` 扩展支持 `enum`；新增 schema↔prompt↔struct 一致性 lint gate（D-13）。
- `003-language-coordinate-simplification`：删除默认 `en` prompt/rubric 副本，收敛到 canonical `multi` baseline；保留 runtime language fallback 与 output-language target。
- `004-real-model-profile-and-evals`（**当前 active plan**）：真实 LLM Judge 实现（`Judge` 演进为逐维度 `[]Score`）+ `judge.default` 激活 + profile coverage 非 placeholder 门禁 + ≥ 50 题离线评估集（Promptfoo，registry single-source，录制 fixture 默认 + `EVAL_LIVE` opt-in）；依赖 A3 `003-provider-registry-and-capability-profiles`（completed）提供完整 profile coverage 与 judge capability profile。
- `005-grayscale-and-quality-feedback`（原 004）：PostHog 灰度分桶 + 报告页质量主观评分回流。

后续如需扩展（多模态 prompt / 函数调用 prompt schema）：递增 spec 版本，原地修订；不创建 sibling spec。
