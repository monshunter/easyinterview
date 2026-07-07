# Resume Detail Readonly and LLM Display Name 交付复盘报告

> **日期**: 2026-07-07
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 Resume Workshop 创建和详情闭环的二次修复：详情页只读展示原始简历正文；upload / paste 注册成功后直接打开 `resume_versions?resumeId=<id>`；删除解析动画页、预览确认页和确认保存页；列表/详情名称只能来自 LLM-derived `displayName` 或 LLM/结构化 headline，不能使用通用来源名、raw resume 第一行或文件名；上传 PDF / DOCX / Markdown / text 会提取可读正文作为 AI prompt input 和 `parsed_text_snapshot`。

已通过的主要证据：

- `go test ./internal/resume/jobs -run 'TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1`
- `go test ./internal/resume/store -run 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady|TestCreateWithParseJobInsertsResumeAndJobAtomically|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1`
- `go test ./internal/resume/... ./cmd/api -run 'TestResume|TestParse|TestCreateWithParseJob|TestCompleteParse|TestBuildResumeRuntime' -count=1`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`
- E2E.P0.035 / P0.037 / P0.081 `trigger -> verify`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` post-fix verification: zero drift
- `BUG-0137` 已修订并登记到 `docs/bugs/INDEX.md`

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

- P0 scenario verify 中涉及“retired UI”的场景，应优先断言不存在旧 DOM / 旧 imports，而不是保留旧目录名里的正向语义作为说明。
  - **落点**：README / scenario scripts
  - **优先级**：medium

## 5 建议优先级与后续动作

最高价值的下一步是把“简历名称不得来自 raw 第一行 / 文件名”和“上传文件正文提取必须覆盖 PDF/DOCX/Markdown/text”保留在 backend/frontend/scenario 三层 owner gate 中。它们直接对应本轮返工来源：用户可见名称和原文展示不是单个组件问题，而是 backend parse、frontend adapter、UI truth 和 scenario verify 的联合合同。

可以延后处理的是 P0.082/P0.083 目录 slug 的重命名；本轮已把内容和 verify 语义改为 retired/direct handoff，目录名若继续造成误解，再作为独立 test asset cleanup 处理。
