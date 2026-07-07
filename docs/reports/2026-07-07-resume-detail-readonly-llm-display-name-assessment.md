# Resume Detail Readonly and LLM Display Name 交付复盘报告

> **日期**: 2026-07-07
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 Resume Workshop 创建和详情闭环：详情页只读展示原始简历正文；upload / paste 注册成功后直接打开 `resume_versions?resumeId=<id>`；删除解析动画页、预览确认页和确认保存页；列表/详情过滤通用“粘贴的简历 / 上传的简历”名称；backend `resume.parse` 成功后根据 LLM structured output 派生 `display_name`。

已通过的主要证据：

- `go test ./backend/internal/resume/jobs`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx src/app/screens/resume-workshop/create/UploadTab.test.tsx src/app/screens/resume-workshop/create/CreateFlowNonCurrentNegative.test.ts src/app/screens/resume-workshop/adapters/resume.test.ts src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/i18n/localeFiles.test.ts`
- `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts`
- `corepack pnpm --filter @easyinterview/frontend typecheck`
- `node --test ui-design/ui-design-contract.test.mjs`
- E2E.P0.037 / P0.081 / P0.082 / P0.083 / P0.084 `trigger -> verify`
- `sync-doc-index --check`, `make docs-check`, `git diff --check`
- `BUG-0137` 已建档并登记到 `docs/bugs/INDEX.md`

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

## 3 根因归类

- Runtime / spec-plan 漂移：
  - **类别**：spec-plan
  - **根因**：早期 create-flow 的“解析草稿 -> 预览确认 -> 保存”模型没有在产品合同变更后被彻底删除，计划完成态没有阻止旧状态机和旧场景继续存在。

- UI contract 漂移：
  - **类别**：README / spec-plan
  - **根因**：`ui-design` contract test 与场景 expected-outcome 没有随 `docs/ui-design` 和静态原型同步更新。

- 名称 fallback 漏洞：
  - **类别**：spec-plan
  - **根因**：计划只要求 LLM 生成完成态名称，但没有明确旧数据/解析前状态的 frontend generic-title negative fallback。

## 4 对流程资产的改进建议

- UI 删除类变更应在 plan gate 中固定“runtime source + locale + CSS + ui-design contract + scenario README/data/scripts”的零残留检查。
  - **落点**：spec-plan
  - **优先级**：high

- Resume 命名合同应明确两层：backend LLM-derived `displayName` 是完成态来源，frontend 必须对通用 source title 做负向过滤并从原文/文件名/结构化字段兜底。
  - **落点**：spec-plan
  - **优先级**：high

- P0 scenario verify 中涉及“retired UI”的场景，应优先断言不存在旧 DOM / 旧 imports，而不是保留旧目录名里的正向语义作为说明。
  - **落点**：README / scenario scripts
  - **优先级**：medium

## 5 建议优先级与后续动作

最高价值的下一步是把“UI 删除类零残留 gate”和“简历名称 generic-title negative fallback”固化到 frontend-resume-workshop 后续 owner checklist。它们直接对应本轮返工来源：旧页面和旧名称都不是单个组件问题，而是 runtime、UI truth、scenario 和 adapter contract 的联合漂移。

可以延后处理的是 P0.082/P0.083 目录 slug 的重命名；本轮已把内容和 verify 语义改为 retired/direct handoff，目录名若继续造成误解，再作为独立 test asset cleanup 处理。
