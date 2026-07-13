# Prompt Rubric Registry Spec

> **版本**: 2.41
> **状态**: active
> **更新日期**: 2026-07-13

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 F3 `prompt-rubric-registry` 定义为当前 active Quality / AI Governance spec（依赖 [A3 `ai-provider-and-model-routing`](./../ai-provider-and-model-routing/spec.md) 与 [B4 `db-migrations-baseline`](./../db-migrations-baseline/spec.md)）。它直接承接当前 AI 调用上下文的版本管理层。

当前 feature_key、prompt/rubric 坐标与 AI task 命名空间由本 spec、product-scope 当前范围、B4 与后续 `config/prompts` / `config/rubrics` 编码 truth source 决定。F3 独立承接 `feature_key`、prompt version、rubric version、language、template hash、model profile reference、Resolve 契约、LLM Judge 接口与 prompt/rubric lint gate。

[ADR-Q6 §3.6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 已锁定：F3 只持有 `(feature_key, prompt_version, rubric_version, model_profile_name)` 四元组，不持有 provider / model 字符串（后者归 A3 Provider Registry + Model Profile）。当前 6 个 baseline feature_key 的默认 `model_profile_name` 必须全部能在 A3 `config/ai-profiles.yaml` catalog 中解析到合法 capability profile。

本 spec 由 `engineering-roadmap/001-decompose-subspecs` 的 contract lock 创建；当前执行口径是固定 baseline prompt/rubric 的命名空间、文件落点、`feature_key + version` 坐标与 Resolve 调用契约。真实 baseline prompt / rubric 文件、loader 与 lint 由 F3 后续 plan 验证；未通过前，后续业务域不得 hardcode prompt 文本，也不得启动依赖 F3 的 AI 调用 implementation。

目标是：

1. **contract 就绪**：每个 P0 AI task 至少有稳定 `feature_key + version` 坐标与文件落点；真实 baseline prompt + rubric 文件由 F3 `001` plan 落地并验证后，后续业务域才能引用。
2. **跨语言、跨任务、跨灰度统一**：6 个当前 baseline feature_key（见 §3.1.1）共享同一 schema（feature_key / prompt_version / rubric_version / language / template_hash）；baseline 文件只维护 canonical `multi` 坐标，运行时 `language` 仍作为用户目标输出语言与 provenance 字段。
3. **评估升级路径**：F3 `004-real-model-profile-and-evals` 使用真实 judge Model Profile，并为当前 6 个 feature_key 维护每个至少 4 个 fixture case；新增/删除 feature_key 时同步调整。
4. **LLM Judge 接口**：本 spec 锁定 LLM Judge 接入契约，`004-real-model-profile-and-evals` 负责落地 `LLMJudge` 真实实现，并保留 `FailClosedJudge` 作为未配置 caller 的安全默认值。

本 spec 不实现具体业务调用现场（归各 C 域）、不实现 AIClient（归 [A3](./../ai-provider-and-model-routing/spec.md)）、不实现 DB 表（归 B4）。

## 2 范围

### 2.1 In Scope

- **prompt 真理源**：`config/prompts/<feature_key>/<version>.{yaml,md}` 表示 canonical `language: multi` baseline；P0 baseline 不再复制 `en` 等无语义差异语言版本。只有当某语言存在真实任务语义差异（例如地区法规、面试习惯、英文专项训练）并由 spec/plan 记录 rationale 时，才允许新增 `config/prompts/<feature_key>/<version>.<language>.{yaml,md}` override。YAML 元信息字段为 feature_key / version / language / template_hash / status / created_at，Markdown 模板正文与同名 YAML 成对存在。
- **rubric 真理源**：`config/rubrics/<feature_key>/<version>.yaml` 表示 canonical `language: multi` baseline；rubric 评估标准默认语言无关，不通过复制 `en` / `zh` rubric 文件实现用户可见本地化。只有当评估标准本身因语言/地区产生真实语义差异并由 spec/plan 记录 rationale 时，才允许新增 `config/rubrics/<feature_key>/<version>.<language>.yaml` override。schema：`feature_key` / `version` / `language` / `status=active|inactive` / `dimensions[]`（每个 dimension：`name` / `weight` / `score_levels[{label, threshold, description}]`）。发布后 rubric 内容不可变，只有 `status` activation metadata 可由最终激活 owner 修改。
- **output schema 真理源**：`config/prompts/<feature_key>/<version>.schema.json` 是该 feature_key 模型输出的**语言无关** JSON Schema 子集（校验关键字包含 `type` / `required` / `properties` / `additionalProperties` / `items` / `enum`、numeric bounds、string bounds/pattern 与 array bounds/uniqueness，允许 `description` 作为非校验注解），multi 与所有语言变体共用同一份（JSON key 与结构语言无关，不随 language 重复抄写）；`RegistryClient` 加载后随 `ResolveActive` 输出 `output_schema`，供 A3 `CompletePayload` 在调用返回后做 fail-close 校验。prompt body 中的输出契约段必须由 schema 渲染或被 lint 证明与 schema 一致，禁止手工维护第二份字段清单；其中 example 必须是包含 schema 声明 required + optional 字段的完整代表性 JSON output，并明确要求模型返回 JSON value 而不是 JSON Schema / OpenAPI schema。closed contract 的每个 object 必须显式 `additionalProperties: false`，且已锁定的长度、数量、模式与唯一性边界必须由 lint、loader 和共享 runtime validator 执行，不能只写在 prompt prose。schema 不持有 provider / model / endpoint / SDK 私有字符串（与 D-12 provider-neutral 边界一致）。语音（STT/TTS）feature 不产 JSON content，不在本真理源范围。
- **DB 表 schema 引用**：`prompt_versions` / `rubric_versions` 字段与 index 由本 spec 锁定；DB 落地由 B4。
- **加载器（`internal/ai/registry/`）**：
  - `RegistryClient.GetPrompt(featureKey, version, language) → (template, meta)`
  - `RegistryClient.GetRubric(featureKey, version, language) → (schema, meta)`
  - `RegistryClient.ResolveActive(featureKey, language) → (prompt_version, rubric_version, model_profile_name)`
  - 启动时从 `config/prompts/` + `config/rubrics/` + DB 同步；DB 是 staging / prod 真理源；本地 dev 直接读文件。
- **业务调用规约**：业务代码必须先 `Resolve(featureKey, ctx.Language)` 拿到三元组，然后传给 `AIClient.Complete(profileName, payload)`；payload 中携带 `prompt_version + rubric_version + feature_key`。
- **lint 规则**：禁止业务包出现 `prompt :=` 字面量字符串 / 多行字符串模板；当前由本地 lint gate 接入，远端 CI 仅在 A5 触发条件成立后再接入；任何 prompt 必须从 registry 加载。
- **contract 内容**：6 个当前 baseline feature_key 各 1 份 active baseline prompt + rubric 的坐标、schema 与落点在本 spec 中锁定；prompt 文本必须是可用文案，不写「TBD」。
- **LLM Judge 接口**：`Judge(featureKey, prompt_version, output, rubric_version) → ([]score, reasoning)`；返回值自 spec v2.8（D-9）演进为**逐 rubric dimension 一项**的 `[]Score`，实现归 `004-real-model-profile-and-evals`。
- **灰度策略**：同 `(feature_key, language)` 同时只允许 1 个 `is_active=true`；P0 baseline active language 坐标为 `multi`。灰度切换由 PostHog feature flag（[A4 D-4](../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典)）+ `Resolve` 内部分桶逻辑实现（后续接入）。

### 2.2 Out of Scope

- AI 调用本身：归 [A3](./../ai-provider-and-model-routing/spec.md)。
- 模型解码期强约束（向 provider 下发 `response_format` / `json_schema` mode）与 provider 侧结构化输出：归 [A3](./../ai-provider-and-model-routing/spec.md)；F3 只产 provider-neutral `output_schema`，由 A3 决定是否向具体 provider 下发，或仅在返回后做 fail-close 校验。
- 业务调用现场（C4-C7 / C9 在哪一行调用）：归各 C 域。
- LLM Judge 实现：归 F3 `004-real-model-profile-and-evals`。
- 离线评估集按当前 6 个 feature_key 每个至少 4 个 fixture case 维护。
- prompt / rubric 编辑 UI：当前 P0 不在范围。
- DB 表本身：归 B4。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含 6 个当前 baseline feature_key 字典）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 唯一标识 | 三元组 `(feature_key, version, language)` 是 prompt 与 rubric 的唯一坐标；version 用 SemVer（major.minor.patch）；P0 baseline canonical 坐标只使用 `language=multi` | DB 表 unique 约束已就位 |
| D-2 | 文件落点 | `config/prompts/<feature_key>/<version>.yaml`（meta）+ `<version>.md`（template）与 `config/rubrics/<feature_key>/<version>.yaml` 是 `language: multi` canonical baseline 文件。语言 override 统一追加 `.<language>` 后缀，但只有存在真实任务/评估语义差异并由 spec/plan 记录 rationale 时才允许新增；禁止为英语或其他 UI 语言复制无差异 baseline。`config/` 目录由 [A4](./../secrets-and-config/spec.md) 拥有，F3 在此命名空间 | 防止散落；避免隐藏配置层重复维护 |
| D-3 | template_hash | `sha256(template_body + meta_for_hash_canonical_json)`；`meta_for_hash` 是删除 `template_hash` 字段后的 YAML meta，避免 hash 自引用；自动计算，写入 yaml；本地 drift 校验 | – |
| D-4 | model_profile_name 引用 | `Resolve` 输出三元组 +「model_profile_name」（如 `practice.chat.default`），由 A3 Model Profile 定义 | F3 不持有 provider / model 字符串（与 ADR-Q6 一致） |
| D-5 | 业务调用契约 | 业务必须先 `Resolve(featureKey, ctx.Language)` 再调 `AIClient`；payload 中带三元组，便于 ai_task_runs 表写入 | 强制可追溯 |
| D-6 | language 兼容 | `language` 列允许 `multi` 表示语言无关；Resolve 优先匹配精确 language → fallback `multi`。请求语言（如 `en` / `zh-CN` / `fr`）是模型输出语言目标与 provenance，不要求 storage 中存在同名 baseline 文件 | – |
| D-7 | 灰度规则 | 同 `(feature_key, language)` 只允许 1 个 prompt active version + 1 个 rubric active version；A/B 由 PostHog flag + Resolve 内部分桶（后续实现）；P0 baseline 只在 `multi` coordinate active，不分桶 | – |
| D-8 | 6 个当前 baseline feature_key 字典 | 见 §3.1.1；增删必须 spec 修订 | – |
| D-9 | LLM Judge 接口 | 签名按 spec 版本演进：v2.8 起返回 `([]Score, Reasoning, error)`，`[]Score` 每个 rubric dimension 一项（`Score{Dimension, Value}`，`Value ∈ [0,1]`，可按 rubric `score_levels[].threshold` 映射 label）；`Reasoning` 保留 `Summary` + `EvidenceQuotes`；`TestJudgeSignature` 断言新签名；实现归 `004` | 多维度 rubric 评估需逐维度评分；接口演进不创建 sibling，原地随 spec 版本更新 |
| D-10 | 不入 log 明文 | template_body 不写入 log；只写 prompt_version + template_hash；与 [F1 D-6](../observability-stack/spec.md#31-已锁定决策含命名约定字典) 一致 | – |
| D-11 | A3 profile coverage | §3.1.1 中每个默认 `model_profile_name` 必须在 A3 `config/ai-profiles.yaml` catalog 中存在，并能解析到合法 `capability` / `provider_ref` / `status`；P1/P2 项可为 `status=disabled` / `status=unsupported`，但必须携带 `unsupported_reason` 且不得缺命名空间；本 gate 由 `make lint-ai-profile-coverage` 和顶层 `make lint` 触发 | 防止 F3 Resolve 输出悬空 profile |
| D-12 | Provider-neutral AI invocation hints | 后续 F3 编码 truth source 可为 Resolve 输出追加 `tools[]`、`output_schema`、`stream_wire`，供 A3 `CompletePayload` / `Stream` 消费；字段只表达业务 schema / wire preference，不表达 provider、model、API endpoint 或 SDK 私有字段 | 让 tools / structured output / streaming handoff 可治理，同时不破坏 A3 provider-neutral 边界 |
| D-13 | output_schema 契约 | 每个 chat feature_key 落地语言无关 JSON Schema；schema、prompt output contract、complete example 与 backend struct 必须一致并 fail-close。当前 6 个 chat feature_key 顶层均为 object。 | `practice.session.chat` 只允许 `messageText`；`report.generate` 只允许会话级 report shape |
| D-14 | language-coordinate simplification | Baseline prompt/rubric truth source 只维护 `multi` coordinate；`en` 等用户语言不再默认复制成隐藏配置文件。`ResolveActive(featureKey, requestedLanguage)` 保持 exact → `multi` fallback，prompt 通过 `{{language}}` 指令对齐用户目标语言；rubric 默认语言无关。新增 language override 必须说明真实语义差异并有 lint/seed/loader gate 覆盖 | 保留多语言用户体验，同时去掉 prompt/rubric 重复维护成本 |
| D-15 | 评估框架与 judge 决策（`004` 锁定） | Promptfoo 消费同一份 registry prompt；真实 provider 经 `EVAL_LIVE` opt-in；LLM Judge 使用隔离的 `judge.default`；当前 6 个 business prompt 均由 single-source registry 解析。 | 保证 single-source、确定性与 provider 隔离 |
| D-16 | Grounded report output owner | F3 `002` 唯一拥有 `report.generate/v0.2.0` prompt、output schema、registry 多版本解析与最终激活；它不创建或修改 v0.2 rubric 内容。closed schema 直接输出 summary/preparedness/code+label dimensions/dimensionCode evidence/actions/retryFocusDimensionCodes 与内部 sourceMessageSeqNos；runtime 不注入 judge rubric 或 numeric score。完整 JSON exemplar 必须保留，并与相邻 synthetic candidate input 配对；当前使用非阻断的 prioritization/tie-breaking gap，且 anti-copy 规则要求只学习 JSON shape/cross-field coherence、从当前 frozen context 重新生成全部事实。Prompt policy 只把 corrective `retry_current_round` 约束为转写 cited missing behavior；focused retry 首 label 必须对每个 selected focus code 至少命名一个直接引用 missing behavior，multi-focus 按 code 顺序使用分号短片段，English action `<=24` whitespace words、zh-CN `<=64` Unicode code points；targeted action-label repair prompt 使用内部生成目标18/52。umbrella term 无效。`next_round` 保持 readiness+hasNextRound gate；`review_evidence` 审计归 D-17，且 prompt 禁止它虚构 artifact/corrective gap/new scenario/transfer task。所有类型禁止未引用的新具体性。Focus policy 只允许两个 exact single-issue exception 使用 empty；其他 retry focus 精确复制完整 same-code needs-work issue codes 升序唯一集合。 | prompt/schema/struct/seed/example/paired-input/anti-copy 同步；unknown/old key fail closed；业务 cross-field validator 归 backend-review |
| D-17 | Context-aware report owner | 方案 A wire fuse200、semantic24/64、targeted18/52不变。Evalkit generation与judge维护独立max4-call budgets。Generation每轮完整validate并按当前violations选择action_labels/whole_report；最多3 retry。Judge仅对retryable provider或protocol/schema invalid重试；结构合法negative verdict是typed terminal content rejection，禁止重采样。 | attempt_count>4、错误scope、raw/truncate、protocol/content outcome混淆或valid-negative retry均失败 |
| D-18 | Coordinated version activation | registry 必须按 `(feature_key, language, version)` 保留 prompt/rubric 多版本，`GetPrompt` / `GetRubric` 精确读取指定版本，`ResolveActive` 各选择唯一 active prompt 与 rubric。`report.generate/v0.2.0` 与 `practice.session.chat/v0.2.0` 内容在发布后 immutable；只有 `REPORT_STORAGE_V18_PASS`、F3 `004` 的 rubric/eval GREEN 与 F3 `002` structural GREEN 同时成立后，`002` 才能修改 status metadata。dev 文件与 staging/prod DB 是互斥 truth substrate，不宣称跨文件系统与数据库原子：dev 一次切换两 feature 的 v0.1/v0.2 prompt+rubric 共 8 个 status，再由 loader 完整校验并 atomic snapshot swap；DB 由 `000019_activate_report_and_practice_prompt_rubric_v020` 单事务插入/激活两组 v0.2 pair。 | 两组 v0.1.0 均保留为可验证 rollback 坐标；dev rollback 同时还原 8 个 status 并 reload，DB down 同时恢复两组 v0.1；任一 marker 缺失、active 数量不是 1、version parity drift 或内容漂移均阻止 mutation |
| D-19 | Practice semantic-focus prompt pair | F3 `002` 拥有 immutable `practice.session.chat/v0.2.0` prompt/schema 与同版本 rubric pair。prompt 只消费 backend-practice 提供的结构化 `semanticFocus` / `{{semantic_focus_json}}`，不得消费 code-only focus 或旧 `focusCompetencies` token；schema 保持 closed `{messageText}`。v0.2 rubric 的 dimensions/weights/score levels 必须与 v0.1 byte-semantic 等价，仅 version/status coordinate 不同，避免为 token rename 发明新评估语义；发布前为 inactive，D-18 协调发布后与 prompt 一起 active。 | backend-practice/004 只拥有 runtime payload 与 exact-version test；F3/002 拥有文件、hash、version parity、activation/migration；v0.1 文件和 000002 不改且保留为 rollback coordinate |
| D-20 | Resolve provenance coordinate | `ResolveActive(report.generate)` 在 active pair 为 v0.2.0 时必须返回 `DataSourceVersion=report-context.v1`，使 backend-review generation `CallMetadata` / `AICallMeta` 与冻结输入坐标一致；不得沿用通用 registry coordinate。`practice.session.chat/v0.2.0` 仍返回现有 `registry.v1`。 | focused resolver test 同时断言 report/practice，防止条件扩散到其他 feature key |
| D-21 | Independent live report review | P0.100 真实矩阵中 generation/judge 同模型，judge PASS不能单独自证。每个case先在generation max4内完整validate/repair，再在独立judge max4内按typed retry policy执行；valid negative立即FAIL。最终 representative packet 再交blind Agent reviewer。 | manifest分别保留generation/judge attempt_count/retry_count/reason/scope、aggregate usage/latency与item/causal digest；valid negative retry、attempt5、protocol/content混淆或任何unsupported都失败 |
| D-22 | TargetJob raw-text-only prompt | `target.import.parse` 的唯一 JD 内容输入是 raw JD text（`{{jd_text}}`）；`{{target_language}}` 只指定输出语言。current prompt 不接收 URL、page heading、file 或 source metadata。 | prompt body/meta hash、baseline seed、registry resolved snapshot 与 TargetJob render tests 同步；旧 source token/wording 只允许出现在显式负测或合法历史证据中 |

#### 3.1.1 6 个当前 baseline feature_key 字典

| feature_key | 用途 | 关联业务域 | 关联 Model Profile（默认） |
|-------------|------|-----------|--------------------------|
| `target.import.parse` | JD 解析 | C4 | `target.import.default` |
| `practice.session.chat` | opening 与连续会话回复 | C5 | `practice.chat.default` |
| `report.generate` | 整轮报告生成 | C6 | `report.generate.default` |
| `resume.parse` | 简历解析 | C7 | `resume.parse.default` |
| `resume.tailor.gap_review` | 简历 gap review | C7 | `resume.tailor.default` |
| `resume.tailor.bullet_suggestions` | 简历 bullet 改写 | C7 | `resume.tailor.default`（共享） |

> 备注：当前不维护题目、专用 hint 或逐题 assessment feature key。Practice opening 与 reply 统一由 `practice.session.chat` 承载；报告只使用 `report.generate` 生成会话级维度与证据。

### 3.2 待确认事项

- 是否引入 prompt versioning 的语义化命名（如 `v1.0.0-baseline` / `v1.1.0-better-followup`）：默认是；由 `004` 评估迭代时按需启用。
- 评估工具选型：**已由 `004`（D-15）锁定为 Promptfoo**（pnpm Node），不引入 OpenAI Evals / `prompt-eng` 编辑器；Promptfoo 经 registry 解析消费同一份 prompt 真理源。
- LLM Judge 使用哪个 model profile：**已由 `004`（D-15）锁定为 `judge.default`**（与业务 chat profile 隔离）。
- rubric 维度名是否与 F1 / F3 AI 质量指标对齐：**已由 `004`（D-15）锁定为复用现有 rubric `dimensions[]`**（如 `followup_relevance` / `report_specificity` / `score_outlier` / `language_consistency`；report 业务仍可复用自身 `report_calibration` 维度），不新造同义维度。

## 4 设计约束

### 4.1 schema 约束

- prompt 元信息字段顺序固定（与 DB 表列顺序一致）：`feature_key / version / language / template_hash / status / created_at`；`status ∈ {draft, active}`。
- rubric `dimensions[].name` 必须使用 F1 / F3 推荐质量指标中定义的命名 +（业务域专有维度 by C 域 owner）；不允许重新发明同义维度。
- `version` 必须递增并使用 SemVer 字符串（baseline 从 `v0.1.0` 起）；同 `(feature_key, version, language)` 不允许覆盖（CI 拦截）。Baseline active 文件只要求 `language: multi`；language override 是例外路径，必须有业务 rationale。
- output schema 文件语言无关；当前 6 个 active chat feature_key 顶层 `type` 均为 `object`。schema、prompt body、complete example 和 backend struct 必须一致。
- schema enum 必须对齐 B1/shared enum、DB CHECK 与后端 consumer；当前不存在 question review enum。
- 受 output schema 约束的 runtime parser 只消费 schema canonical keys；范围外 alias 不属于 prompt 或 runtime contract，只能作为 negative-test 输入验证 fail-close / degrade 路径。
- `practice.session.chat` 输出只包含 canonical `messageText`；不得要求 question intent、generation kind、answer summary 或 hint cue。
- `practice.session.chat/v0.2.0` 输入只使用 JSON 编码的 `semanticFocus`；空值表示通用同轮复练，非空值是服务端解析后的 report-local code/label/issues 结构。旧 `focusCompetencies` 只允许保留在 immutable v0.1 rollback coordinate、000002 历史 seed 与显式负测，最终 active runtime 不得消费。
- `target.import.parse` 只渲染 `{{jd_text}}` 与输出语言指令 `{{target_language}}`；prompt/meta、active seed、resolved snapshot 与 runtime renderer 不得声明或消费 source URL、page heading、file metadata 或其它 source descriptor。
- `report.generate/v0.2.0` 从 root 到 nested item object 全部 closed；summary/dimension/evidence/action/focus 的已锁定 min/max/pattern/unique 约束必须同时通过 Python lint、registry loader 与 A3/F3 共享 `outputschema` runtime validator。
- `report.generate/v0.2.0` 保留完整 JSON exemplar，并在其前放置相邻 synthetic candidate input：候选人按 user impact 与 delivery effort 排序，同时明确没有解释 tie-breaking rule。示例的 highlight/issue/readiness/action 必须只引用该 synthetic seqNo；紧邻 anti-copy 规则明确示例只教授 JSON shape 与 cross-field coherence，不得复用示例事实、维度、准备度、措辞或行动。示例不得使用 unbounded load、未验证 rollback 或其他 unsafe current-round approach 充当 `needs_practice` 普通缺口。
- `report.generate/v0.2.0` action support 按 type 判定：corrective retry 只把 cited missing behavior 转成待补动作，focused retry 首 label 还必须对每个 selected focus code 至少命名一个 directly cited missing behavior，umbrella-only label 无效；review_evidence 可复核 cited positive 或 explicitly evidence-limited content，但不得虚构 artifact/corrective gap/new scenario/transfer task；next_round 只在 `hasNextRound=true` 且 readiness 为 basically/well 时支持。即使建议在领域上合理，任何类型新增未被 cited candidate messages 陈述的 mechanism、threshold、tool、sequence、framework 或 example，都超出 support 边界。
- Action schema `maxLength=200` code points只作fuse；English24 whitespace words/zh-CN64 Unicode code points为semantic/UX gate。Targeted action-label repair使用内部生成目标18/52，不改变UI边界。P0.099以desktop+390合法边界完整换行和typed-invalid/no-raw验收UI。
- Evalkit复用产品完整validator并以generation budget=4逐轮执行：纯label schema200/24/64选择`action_labels`，其它schema/semantic/mixed选择`whole_report`；每轮输出完整复验，最多3 retry。Judge使用独立budget=4；retryable provider或protocol/schema invalid可重试，valid unsupported/causal/zero-tolerance/critical negative直接typed content-rejected。两条链分别聚合usage/latency并记录脱敏attempt/retry/reason/scope。
- `report.generate/v0.2.0` focus decision table 使用 `I=len(issues)` 与 `F=所有 same-code needs-work issue dimension codes 的升序唯一集合`：无 retry 时 focus=`[]`；final exact single `answer_depth` brief 或 single `answer_relevance` control-only exception 时 focus=`[]`；其他 retry 必须 focus=`F` 且 `F` 非空。任何 subset/superset 失败，`I>=2` empty 无条件失败。
- `report.generate/v0.2.0` judge request 从 strict-decoded output 机械派生 ordered `expected_item_verdicts(path/kind)` 与 `expected_causal_dimension_codes`。Judge response 必须精确覆盖这些坐标：空 `highlights` / `issues` 不创建 `$.highlights` / `$.issues` 集合 verdict，`$.retryFocusDimensionCodes` 因 empty-versus-focused 本身是 advice 决策而始终保留为唯一数组整体 verdict。

### 4.2 边界约束

- F3 不直接调用 `AIClient.*`（避免循环依赖）；业务在 Resolve 之后自行调用 AIClient。
- F3 不持有 secret；DB 连接从 A4 注入。
- F3 不写入 metric / log（除自身加载状态）；AI 调用观测埋点由 A3 内部完成。
- F3 只产 provider-neutral `output_schema`，不向任何 provider 下发 `response_format` / `json_schema` 请求字段（解码期强约束归 A3）；schema 内不含 provider / model / endpoint / SDK 私有字符串。

### 4.3 性能约束

- `Resolve(featureKey, language)` P95 ≤ 5ms（内存 cache + 30s TTL）。
- 启动时全量预加载 ≤ 1s（6 × canonical `multi` baseline + 可选 language override）。
- 文件改动后 ≤ 30s 热加载（与 A3 Model Profile 同节奏）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `config/prompts/` + `config/rubrics/` | F3 | 真理源文件 |
| `config/prompts/<feature_key>/<version>.schema.json` | F3 | output schema 真理源（语言无关） |
| `config/evals/` | F3 `004` | 离线评估用例（当前 6-key 基线每个至少 4 题）+ LLM Judge 评分指令模板（eval 资产，非业务 prompt，不占用 §3.1.1 业务 feature_key 坐标） |
| `internal/ai/registry/` Go 包 | F3 | RegistryClient + Resolve 实现 |
| `validateOutputSchema`（模型输出 fail-close 校验 + 可选 `response_format` 下发） | A3（`aiclient` observability） | F3 提供 schema，A3 执行校验与下发决策 |
| `prompt_versions` / `rubric_versions` 表 schema | B4 | F3 提供字段名 |
| 业务调用现场 | 各 C 域 | 通过 Resolve 三元组 + AIClient |
| LLM Judge 实现 | F3 `004` | 接口 v2.8 演进为逐维度 `[]Score`；`LLMJudge` 经 `judge.default` 实现归 `004` |
| Model Profile | A3 | F3 引用 profile name |
| 灰度（PostHog flag） | F2 + F3 | F2 owns flag；F3 owns Resolve 分桶逻辑 |
| 离线评估集 ≥ 24（当前 6-key 基线） | F3 `004`（Promptfoo）+ 各 C 域 | 归 `004`；registry single-source；录制 fixture 默认 + `EVAL_LIVE` opt-in |
| 编辑 UI | 当前不在范围 | – |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 6 个 baseline 全集 | conversation/report rebase 完成 | 扫描 prompts/rubrics | 各输出 6 个 feature_key 目录；每个目录有 active `multi` baseline 与语言无关 output schema | F3 `001` + `002` + `003` |
| C-2 | template_hash 一致 | 修改 prompt template body 但忘改 hash | CI | `lint-prompts` 失败；提示重新生成 hash | F3 后续 001 + A5 |
| C-3 | Resolve 业务调用 | C5 调用 `registry.Resolve("practice.session.chat", "en")` | 单测 | 返回 chat prompt/rubric/profile 坐标 | F3 001 + C5 |
| C-4 | 业务不允许 hardcode prompt | 故意在 `internal/practice/` 中加 `prompt := "You are an interviewer..."` | CI | `lint-prompts-hardcode` 失败 | F3 后续 001 + A5 |
| C-5 | 灰度切换 | F3 自行 plan `is_active` 字段 | DB 直接修改 | 同 `(feature_key, language)` previous prompt → inactive DB row；selected prompt → active；Resolve 输出 selected version | F3 `005` |
| C-6 | 多 language fallback | 调 `Resolve("report.generate", "fr")`，`fr` baseline 不存在 | 加载逻辑 | 退化到 `multi` baseline；log warn | F3 后续 001 |
| C-7 | LLM Judge 接口（v2.8 演进） | 编译期 | F3 包 export `Judge` 接口 | 接口按 spec 版本演进：返回 `([]Score, Reasoning, error)`，`[]Score` 逐 rubric dimension；`TestJudgeSignature` 断言新签名；业务代码可 import 抽象 | F3 `001`（接口）+ `004`（演进+实现） |
| C-8 | F3 executable baseline handoff | 6 个 baseline 已落地 | 运行 loader/lint | prompt/rubric/schema/seed 坐标一致 | F3 001 |
| C-9 | DB 表写入闭环 | A3 调用产生 `ai_task_runs` 行 | 数据库 | `feature_key` + `prompt_version` + `rubric_version` + `feature_flag` + `data_source_version` typed 字段非空；其中 feature/prompt/rubric/data-source 与 Resolve / CallMetadata 输出一致，flag 无分桶时写 `none` | A3 + B4 + F3 |
| C-10 | 评估升级 | F3 `004` 维护每个 active feature key 至少 4 题离线评估集 + 真实 LLM Judge | 运行 `make eval-offline`（录制 fixture 默认）/ `EVAL_LIVE=1` opt-in | 当前 6-key 基线共 ≥ 24 题、`LLMJudge` 逐维度产出、`judge.default` active 与录制/live 执行模式均已验证 | F3 `004` |
| C-11 | A3 profile coverage | A3 003 完成 provider registry + capability profile catalog | 运行 `make lint-ai-profile-coverage` 或顶层 `make lint` | §3.1.1 的默认 `model_profile_name` 全部存在于 `config/ai-profiles.yaml`，且 capability / provider_ref / status 合法；`disabled` / `unsupported` profile 必须显式标记并携带 `unsupported_reason` | A3 003 + F3 后续 001 |
| C-12 | output_schema 契约闭环 | 6 个 schema 与 resolver 接线 | prompt/registry/aiclient gates | 每个 feature key schema/prompt/example/struct 一致；缺 required 或非法 enum fail-close | F3 002 |
| C-13 | language-coordinate 收敛 | canonical multi baseline | lint/registry/seed/migration | loader snapshot 6 个 coordinates；requested locale fallback 到 multi | F3 003 |
| C-14 | judge.default/profile coverage | profiles active/disabled 符合 owner | profile coverage lint | 6 个 business chat profile 可解析；voice profiles保持 disabled | F3 004 + A3 |
| C-15 | eval prompt single-source | F3 `004` Promptfoo 接线 | 运行 eval drift gate + `make lint-prompts-hardcode` | Promptfoo 经 `RegistryClient.ResolveActive` 消费同一份 prompt；无第二份 prompt 副本；prompt 漂移 → drift gate exit 1；`lint-prompts-hardcode` 仍 green | F3 `004` |
| C-16 | grounded report schema | report.generate v0.1 rollback coordinate 与 v0.2 draft prompt/schema | lint + focused multi-version parser tests | direct shape/anchors exact-match；旧 numeric/dimension/focus key 与 unknown field fail closed；GetPrompt 可精确读取两版 | F3 002 + backend-review 001 |
| C-17 | report reliability + UX | v0.2 +5 contexts | live eval+UI audit | 200只作fuse；24/64质量；18/52只作targeted-repair内部余量；P0.099 desktop+390/typed-invalid-no-raw | F3 004 + P0.100 + P0.099 |
| C-18 | report + practice v0.2 gated activation | `REPORT_STORAGE_V18_PASS`、F3 002 structural GREEN 与 F3 004 rubric/eval GREEN 均已产出 | dev 8-status activate/rollback/re-activate + DB `000019` independent PostgreSQL up/down/up | 两个 feature 的 ResolveActive 均唯一返回 prompt/rubric v0.2.0；exact getters 仍可读取两组 v0.1.0；file rollback/reload 与 DB down 同时恢复两组 v0.1.0，任一 marker 缺失时 mutation 前失败 | F3 002 + F3 004 + B4 + backend-practice 004 |
| C-19 | practice semantic focus pair | backend-practice 将 report-local code 解析为 label/issues 的 JSON payload | exact `GetPrompt/GetRubric(v0.2.0)` + prompt lint/hash + backend-practice fixture + final `ResolveActive` | v0.2 prompt 只含 `semanticFocus` / `{{semantic_focus_json}}`，closed schema 仍只输出 `messageText`；v0.2 rubric 内容/权重与 v0.1 相同；协调发布后 v0.2 pair active，v0.1 pair 保留为 exact rollback coordinate | F3 002 + backend-practice 004 |
| C-20 | report v0.2 frozen provenance | report/practice v0.2 pairs active | exact ResolveActive test | report returns `report-context.v1`; practice remains `registry.v1`; prompt/rubric versions are v0.2 on both | F3 002 + backend-review 001 |
| C-21 | independent live report audit | 5-case / 11-attempt matrix，generation/judge 当前同 provider/model | independent max4 generation/judge state machines → full validation / typed judge outcome → blind review | generation动态scope；judge invalid可重试、valid negative不重试；attempt/retry/reason/scope+usage/latency manifest与5-case evidence一致 | F3 004 + P0.100 |
| C-22 | TargetJob raw-text-only prompt | current prompt、seed 与 registry snapshot 已加载 | lint/hash/seed drift + TargetJob resolved-render tests | resolved prompt 只包含 raw JD text 与目标语言指令；无 URL/source metadata token 或说明；hash、seed 与 snapshot byte-semantic 一致 | F3 001 + backend-targetjob 001 |

## 7 关联计划

F3 当前 active impl plan 由 F3 自身的 plans 承接（[engineering-roadmap §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 保留该 active spec）：

- `001-baseline`：6 个 feature_key 的 baseline truth source + Resolve + lint。
- `002-output-schema-contract`（active）：6 个语言无关 schema + prompt/example/struct consistency gate；Phase 14 唯一拥有 report v0.2 与 practice semantic-focus v0.2 prompt/schema pair、多版本 registry、000019 与 marker-gated final activation。
- `003-language-coordinate-simplification`：删除默认 `en` prompt/rubric 副本，收敛到 canonical `multi` baseline；保留 runtime language fallback 与 output-language target。
- `004-real-model-profile-and-evals`（active）：真实 LLM Judge 实现（`Judge` 演进为逐维度 `[]Score`，`FailClosedJudge` 保留为安全默认）+ `judge.default` 激活 + runnable profile coverage 门禁；Phase 8 唯一拥有 report v0.2 rubric、context-aware judge 与当前 28-case suite，Phase 9 消费 A3 judge final-content marker并要求 P0.100 独立 Agent 第二视角。
- `005-grayscale-and-quality-feedback`（原 004）：PostHog 灰度分桶 + 报告页质量主观评分回流。

后续如需扩展（多模态 prompt / 函数调用 prompt schema）：递增 spec 版本，原地修订；不创建 sibling spec。
