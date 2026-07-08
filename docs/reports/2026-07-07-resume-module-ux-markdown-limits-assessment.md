# Resume Module UX Markdown Limits 交付复盘报告

> **日期**: 2026-07-07
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付围绕简历模块体验闭环：删除列表页重复上传入口，新增可配置简历数量上限与 2MiB 默认上传限制，补齐列表删除操作，移除新建页右侧冗余说明栏，新增解析等待/失败状态页，并将解析后的正文统一走 LLM Markdown 结构化输出后在详情页渲染。

成功证据：

- Frontend focused gate: `corepack pnpm --filter @easyinterview/frontend test src/app/i18n/localeFiles.test.ts src/app/screens/resume-workshop/components/ResumeListView.test.tsx src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx src/app/screens/resume-workshop/create/UploadTab.test.tsx src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/screens/resume-workshop/ResumeWorkshopCssParity.test.ts`，7 个测试文件 40 项通过。
- Frontend resume-workshop wide gate: `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`，20 个测试文件 106 项通过。
- Backend resume/config/upload gate: `go test ./internal/resume/... ./internal/platform/config ./internal/upload/handler ./cmd/api -count=1` 通过，并用 focused tests 覆盖 active resume limit、idempotent replay、Markdown parse contract、2MiB resume upload config。
- Prompt/config gate: `make lint-config` 通过，`python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations` 返回 `prompt_lint: 9 files clean`。
- 文档与格式 gate: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 无漂移，`make docs-check` 通过，`git diff --check` 通过。
- 负向搜索确认重复 CTA、右侧说明栏、旧 10MiB resume purpose 配置已清理；仅保留测试中的缺失断言或非简历上传用途命中。

用户复核后追加修复证据：

- Markdown renderer gate: `react-markdown` + `remark-gfm` 接管 `ResumePreviewTab` 正文渲染，focused tests 断言 `strong` 与 `a[href]` DOM 存在，且不再显示 raw `**...**` / `[label](url)` 标记。
- PDF / AI failure fallback gate: `TestParseHandlerMarkdownFallbackSurvivesPDFAIOutputFailure` 覆盖 PDF 文本抽取成功但 AI 输出 invalid JSON 的路径，确认失败快照写入 `#` / `##` / bullet Markdown，并且 prompt 收到的是抽取后的 PDF 文本。
- Fixture/browser gate: `openapi/fixtures/Resumes/getResume.json` 默认样本补成 Markdown 正文，`ResumeDetailFixtureParity.test.tsx` 验证 fixture 页面渲染 inline strong/link；浏览器截图保存在 `.test-output/screenshots/resume-markdown-detail-mock-2026-07-07.png`，DOM 检查确认 heading / h2 / strong / link 存在且 raw Markdown 标记不可见。
- 追加 Bug 记录：[BUG-0139](../bugs/BUG-0139.md) 记录“详情页 plain text 渲染 + PDF AI 失败非 Markdown 快照”回归与修复证据。

## 2 会话中的主要阻点/痛点

本次范围横跨 `frontend-resume-workshop` 的列表/详情与新建流、`backend-resume` 的注册/解析链路、上传配置、prompt schema、seed migration、静态原型与 UI 文档。若只按单一 plan 推进，容易遗漏 checklist、BDD、INDEX 或跨层契约同步。本次通过原地重开并完成既有 owner plan，避免新增 sibling follow-up plan。

Prompt 输出从纯文本扩展为 required `markdownText` 时，隐藏耦合点较多。除了 prompt schema 和 handler test，还需要同步 seed migration、hash、`scripts/lint/prompt_lint.py` 的 feature contract，否则 prompt lint 会暴露迁移与 runtime contract 漂移。

“彻底删除 UI 冗余模块”不能只看可见 DOM。初始实现后仍需要负向搜索运行时代码、静态原型、i18n key 和 CSS 残留，才能确认右侧说明栏与重复 CTA 没有以 dead resource 形式留下。

计划中引用的 P0.081 create-flow 场景没有对应 frontend scenario Vitest 文件。本次用 focused component/CSS parity tests 覆盖相同契约，但这暴露了场景编号与可执行测试资产之间的可发现性不足。

等待/失败页的首版状态若没有返回动作，会形成失败态死路。本次在收尾前补齐返回列表入口，并把测试覆盖到 terminal failed state。

## 3 根因归类

- Cross-owner 需求：同一用户体验需求实际改动了前端视觉、API 行为、配置、prompt、migration 与文档治理，单 owner checklist 容易漏项。
- Prompt contract 分散：schema、prompt body、seed SQL、runtime decoder、fixture AI、lint contract 各自维护，没有一个单点能提示所有必改处。
- UI 删除 gate 不够显式：现有实现 gate 更容易验证“新行为存在”，不容易自动提示“旧文案、旧 testid、旧样式和旧 i18n 已不存在”。
- Scenario asset 命名可发现性不足：BDD 编号存在，但对应执行入口可能散落在 shell/scenario docs 或 component tests 中，容易让闭环验证说明变得含糊。
- Markdown 验收粒度不足：原先只检查正文文本存在，没有把 inline Markdown 语义 DOM 和 AI failure fallback Markdown 作为独立 gate。

## 4 对流程资产的改进建议

- Prompt-rubric 或相关 plan checklist 应新增 gate：当 prompt required output field 变化时，必须同步 schema、prompt body、seed migration body/hash、runtime decoder、fixture AI、`scripts/lint/prompt_lint.py` feature contract 与 focused negative test。
- Frontend UI removal checklist 应新增 gate：删除用户可见模块时，必须执行 runtime/prototype/i18n/CSS/testid/text 的 negative search，并在 checklist 中记录允许保留的测试断言命中。
- Resume create-flow BDD 或 scenario INDEX 应补充 P0.081 的可执行入口说明：如果没有独立 frontend scenario 文件，应明确替代 gate 是 focused component tests + CSS parity tests，避免后续误判为漏测。
- Terminal async state 的 UX gate 应明确要求恢复路径：等待、失败、超时等状态至少提供返回或重试入口，并有测试断言。
- Resume detail / parse gate 应新增 Markdown DOM 与 fallback 双轨断言：success path 验证 LLM `markdownText`，failure path 验证抽取文本的 Markdown fallback，前端验证 `strong` / `a` / heading / list DOM 而不是只看 textContent。

## 5 建议优先级与后续动作

P1：补强 prompt contract 变更 checklist，优先放入 prompt-rubric 或 backend-resume plan 模板。这类遗漏会直接影响运行时解析结果和 seed migration 一致性。

P1：给 UI 删除类需求增加 negative search gate，优先沉淀到 frontend-resume-workshop plan checklist 或 UI design parity gate。

P2：澄清 P0.081 的执行入口，将 scenario 编号、替代测试和实际命令写入对应 BDD 计划或 scenario INDEX。

P2：把异步等待/失败态恢复路径作为 resume-workshop 通用 UX gate，后续上传、解析、归档、重写等流程可复用。

P2：把 Markdown DOM 与 AI failure fallback 的双轨验收沉淀到 resume-workshop/backend-resume 后续 checklist，避免 future prompt 或 renderer 调整再次退化为 plain text。
