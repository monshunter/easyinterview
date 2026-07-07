# Resume Detail Readonly and LLM Display Name 交付复盘报告

> **日期**: 2026-07-07
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 Resume Workshop 创建和详情闭环的二次修复：详情页只读展示原始简历正文；upload / paste 注册成功后直接打开 `resume_versions?resumeId=<id>`；删除解析动画页、预览确认页和确认保存页；列表/详情名称只能来自 LLM-derived `displayName` 或 LLM/结构化 headline，不能使用通用来源名、raw resume 第一行或文件名；上传 PDF / DOCX / Markdown / text 会提取可读正文作为 AI prompt input 和 `parsed_text_snapshot`。

第三轮补修覆盖真实 PDF 空白/乱码问题：resume.parse upload object 读取预算从 256KiB 提升到 8MiB，避免真实浏览器生成 PDF 的 xref / 字体映射被截断；PDF 抽取优先使用 `pdftotext -layout - -`，并给 Go parser / literal fallback 增加可读性 gate；`CompleteParseFailure` 支持保存已抽取的 `parsed_text_snapshot`，使原文预览不依赖 LLM 结构化解析成功；frontend adapter 继续过滤与来源 title 相同的 PDF 文件名 `displayName`。

第四轮补修覆盖 BUG-0138：`resume.parse` prompt/schema 明确 required `displayName`，backend 优先消费并验证 AI 生成名称；AI provider / output 失败但已有正文时，失败事务写入正文派生 fallback `display_name`；P0.037 增加 failed-with-snapshot 详情只请求一次的回归场景，防止同一详情 URL 在非 pending 状态重复请求。

已通过的主要证据：

- `go test ./internal/resume/jobs -run 'TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1`
- `go test ./internal/resume/store -run 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady|TestCreateWithParseJobInsertsResumeAndJobAtomically|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1`
- `go test ./internal/resume/... ./cmd/api -run 'TestResume|TestParse|TestCreateWithParseJob|TestCompleteParse|TestBuildResumeRuntime' -count=1`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`
- `go test ./internal/resume/jobs -run 'TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox|TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1`
- `go test ./internal/resume/jobs -run 'TestParseHandlerRejectsUnreadablePDFText|TestParseHandlerExtractsReadableUploadText' -count=1`
- `go test ./internal/resume/store -run 'TestCompleteParseFailureCanPersistExtractedTextSnapshot|TestCompleteParseFailureMarksFailedWithoutCompletedOutbox|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically|TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady' -count=1`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/adapters/resume.test.ts`
- local UAT：真实 PDF resume `019f3bfb-7d9d-76d2-a82d-08bcbac225ce` 重排后 `parse_status=failed` / `AI_OUTPUT_INVALID`，但 `parsed_text_snapshot` 长度 3083，开头为 `谭章毓 | AI Infra / Agent / 平台工程师` 等可读中文正文。
- E2E.P0.035 / P0.037 / P0.081 `trigger -> verify`
- `go test ./internal/resume/jobs ./internal/resume/store ./cmd/api -count=1`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`
- `make lint-prompts`
- `make lint-prompts-hardcode`
- `go test ./internal/ai/registry -count=1`
- `go run ./backend/cmd/evalkit drift-check`
- E2E.P0.035 / P0.037 `setup -> trigger -> verify -> cleanup`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` post-fix verification: zero drift
- `BUG-0137` 已修订并登记到 `docs/bugs/INDEX.md`
- `BUG-0138` 已创建并登记到 `docs/bugs/INDEX.md`

## 2 会话中的主要阻点/痛点

- 原始用户反馈暴露了 owner plan 完成态和当前 runtime 的语义漂移。
  - **证据**：截图显示“粘贴的简历”、解析动画页、预览确认页和结构化草稿；`rg` 命中 `ResumeParseFlow`、`PreviewStage`、`ResumePreviewConfirm`、旧 locale key、P0.081-P0.084 旧期望。
  - **影响**：如果只改一个入口跳转，场景和 UI contract 仍会继续把旧页面当作验收目标。

- UI contract test 仍锁定旧“preview / rewrites / source preview / export / confirm”合同。
  - **证据**：`node --test ui-design/ui-design-contract.test.mjs` 初次失败，期望 `openResume(r, tab="preview")`、`RewriteSaveConfirmModal`、`OriginalResumePreviewModal` 和 `onConfirm`。
  - **影响**：静态原型已更新后，contract test 反而成为旧设计的保护网，需要同步为新负向 gate。

- P0.037 场景仍期望结构化 headline 出现在详情正文。
  - **证据**：P0.037 初次失败，期望 `Senior frontend engineer for platform-heavy product teams`，实际渲染原文 `Original resume parsed text snapshot`。
  - **影响**：旧场景会误判正确的“原文优先”实现为失败。

- 第一轮修复后，命名合同仍被错误理解为“从原文第一行或文件名兜底”。
  - **证据**：用户截图显示列表名称为 `# 姓名 | AI / Infra / DevOps 平台工程师`，详情标题显示 PDF 文件名；代码中 `PasteTab` 调 `derivePasteTitle(rawText)`，adapter 的 `deriveDisplayName` 仍使用 `title` 和 `firstContentLine()`。
  - **影响**：用户看到的仍不是 LLM 生成的简历名，且 Markdown heading 会被误认为业务名称。

- upload path 没有正文提取 gate。
  - **证据**：`backend/internal/resume/jobs/parse.go` upload 分支直接 `raw = string(body)`；新增红测 `TestParseHandlerExtractsReadableUploadText` 在 PDF / DOCX case 中先失败，snapshot 包含 `%PDF` / `PK` / XML 片段。
  - **影响**：上传 PDF/DOCX 后详情页无法展示真实原文，LLM prompt 也可能消费不可读二进制。

- 真实 PDF 样本暴露了小型合成 PDF gate 不足。
  - **证据**：用户复测 PDF 详情仍为空；本地 DB 中失败 resume 对应 file_object size 为 554631 bytes，旧 parse handler 默认只读 262144 bytes；从 MinIO 取回 PDF 后 `pdftotext -layout` 能提取正文，证明文件可读但 runtime 读取被截断。
  - **影响**：之前的合成 PDF unit test 只证明 parser 分支存在，不能证明真实 PDF 不会因 xref / 字体映射缺失而空白。

- PDF fallback 缺少可读性 gate。
  - **证据**：扩大读取预算后，同一真实 PDF 曾得到 3285 字符 snapshot，但内容开头是 `%¡À` 和控制字符，不是简历正文；旧逻辑只检查“非空”，没有拒绝 PDF literal / binary 乱码。
  - **影响**：页面不再空白但仍不可读，等价于没有展示原始简历。

- 原文快照写入依赖 LLM 结构化成功。
  - **证据**：旧 `CompleteParseFailure` 只写 `parse_status='failed' + error_code`，不接受 `parsed_text_snapshot`；只要 AI provider 或 output validation 失败，详情页仍可能空白。
  - **影响**：用户核心诉求是“预览就是原始简历内容”，不应被 LLM 名称/profile 结构化成功率牵连。

- `resume.parse` prompt 输出合同没有把 UI 名称当成 required 字段。
  - **证据**：`config/prompts/resume.parse/v0.1.0.schema.json` 和 prompt body 缺少 required `displayName`；修改后还需要同步 seed migration、prompt lint contract 和 `config/evals/resolved-prompts.json`，否则 drift-check 会落后。
  - **影响**：即使 LLM 正常返回结构化内容，也没有强约束生成“正常有意义”的简历名称；离线 eval 单源导出也容易在后续阶段才暴露 drift。

- P0.037 只验证 pending upload 轮询，没有覆盖 failed-with-snapshot 停止轮询。
  - **证据**：新增 `ResumeDetailView` regression 与 P0.037 场景后，测试明确断言 failed + readable snapshot + backend generated `displayName` 时 `getResume` 只调用一次。
  - **影响**：同一详情 URL 重复请求的性能风险缺少场景级保护，容易被后续 hook 改动重新引入。

- store SQL mock 测试跟随新失败事务合同时暴露滞后。
  - **证据**：P0.035 trigger 初次失败于 `TestCompleteParseFailureMarksFailedWithoutCompletedOutbox`，旧 mock 仍期望 5 个 update 参数；新实现已经写入 `parsed_text_snapshot` 和 `display_name` 两个可选字段。
  - **影响**：这类测试失败是有价值的合同提醒，但也说明新增失败态字段时要同步老失败路径单测，而不能只新增 happy regression。

## 3 根因归类

- Runtime / spec-plan 漂移：
  - **类别**：spec-plan
  - **根因**：早期 create-flow 的“解析草稿 -> 预览确认 -> 保存”模型没有在产品合同变更后被彻底删除，计划完成态没有阻止旧状态机和旧场景继续存在。

- UI contract 漂移：
  - **类别**：README / spec-plan
  - **根因**：`ui-design` contract test 与场景 expected-outcome 没有随 `docs/ui-design` 和静态原型同步更新。

- 名称 fallback 漏洞：
  - **类别**：spec-plan
  - **根因**：计划只要求 LLM 生成完成态名称，但第一轮没有明确禁止 raw resume 第一行 / 文件名进入可见名称链路；旧数据/解析前状态应显示中性占位或 LLM/结构化 headline，而不是从正文或来源标题猜测名称。

- 上传正文提取缺口：
  - **类别**：spec-plan
  - **根因**：backend-resume D-14 只覆盖 displayName，未把 upload 白名单格式和 `parsed_text_snapshot` 的 readable body extraction 写成 D-15 gate。

- 真实样本覆盖不足：
  - **类别**：test-gate
  - **根因**：PDF gate 使用极小 synthetic PDF，未断言 object reader 的读取预算，也未用真实失败样本大小建立下限。

- PDF 可读性验证不足：
  - **类别**：test-gate
  - **根因**：旧 PDF extraction fallback 只以非空作为成功标准，未检查 replacement/control characters，也没有验证真实样本的正文开头。

- 原文预览与 LLM 结构化耦合：
  - **类别**：runtime-contract
  - **根因**：parse success 事务同时承担结构化 profile、displayName 和原文快照写入；failure path 没有保留已抽取的原文快照。

- Prompt 输出合同不完整：
  - **类别**：spec-plan
  - **根因**：UI 可见名称被当成后端派生细节，而不是 prompt schema、prompt body、backend validator 和 committed eval export 共同维护的合同字段。

- Failed-with-snapshot 轮询缺口：
  - **类别**：test-gate
  - **根因**：原场景只覆盖 pending polling to readable body，没有把 failed/ready/has-readable-body 作为停止轮询的不变量写进 BDD。

## 4 对流程资产的改进建议

- UI 删除类变更应在 plan gate 中固定“runtime source + locale + CSS + ui-design contract + scenario README/data/scripts”的零残留检查。
  - **落点**：spec-plan
  - **优先级**：high

- Resume 命名合同应明确两层：backend LLM-derived `displayName` 是完成态来源；frontend 必须过滤通用 source title，禁止从 raw 第一行或文件名兜底，只能使用 LLM/结构化 headline 或中性占位。
  - **落点**：spec-plan
  - **优先级**：high

- Upload 白名单格式必须有 readable text extraction gate，至少覆盖 PDF / DOCX / Markdown / text 的 prompt input 与 `parsed_text_snapshot` 一致性。
  - **落点**：spec-plan / scenario scripts
  - **优先级**：high

- PDF gate 必须断言真实文件读取预算，并保留至少一个接近真实浏览器导出大小的 fixture 下限；小型 synthetic PDF 只能作为 parser smoke。
  - **落点**：unit test / scenario scripts
  - **优先级**：high

- PDF gate 必须断言 extracted text 可读；非空但含大量不可打印字符、replacement char 或 PDF literal 乱码时必须失败，不能进入 AI prompt / snapshot。
  - **落点**：unit test / backend-resume D-15
  - **优先级**：high

- 原文快照写入应先于或独立于 LLM 结构化成功；failure path 已抽取正文时必须保存 snapshot，completed event 仍仅 ready 成功发出。
  - **落点**：backend-resume store/job contract
  - **优先级**：high

- Prompt schema 新增 UI 可见字段时，必须同步 prompt body、hash、seed migration、prompt lint contract、resolved export 和 drift-check 证据。
  - **落点**：prompt-rubric / backend owner checklist
  - **优先级**：high

- 详情轮询 gate 应把 `queued|processing && no readable body` 作为唯一正向轮询条件，并把 failed-with-snapshot 单次请求写入组件测试和 BDD 场景。
  - **落点**：frontend owner checklist / scenario scripts
  - **优先级**：high

- P0 scenario verify 中涉及“retired UI”的场景，应优先断言不存在旧 DOM / 旧 imports，而不是保留旧目录名里的正向语义作为说明。
  - **落点**：README / scenario scripts
  - **优先级**：medium

## 5 建议优先级与后续动作

最高价值的下一步是把“简历名称不得来自 raw 第一行 / 文件名”“prompt 必须生成 required `displayName`”“失败态已有正文时也要写 fallback `display_name`”和“详情只在 truly pending 且无正文时轮询”保留在 backend/frontend/scenario 三层 owner gate 中。它们直接对应本轮返工来源：用户可见名称、原文展示和详情请求频率不是单个组件问题，而是 backend parse、frontend hook、UI truth 和 scenario verify 的联合合同。

可以延后处理的是 P0.082/P0.083 目录 slug 的重命名；本轮已把内容和 verify 语义改为 retired/direct handoff，目录名若继续造成误解，再作为独立 test asset cleanup 处理。
