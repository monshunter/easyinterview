# UI Design Voice Mode Regression 交付复盘报告

> **日期**: 2026-05-02
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖用户指出的语音模式回归：上一轮把语音模式折回 `PracticeScreen` 时误删了语音专属呈现。本次修复将波形、标注波形、实时转写、表达层指标、现场提示和音频留存说明恢复到 `PracticeScreen` 的语音 Surface，同时保持顶部结束 / 暂停、题目地图、面试形式切换和 session 上下文统一。

已通过的验证：

- `node --test ui-design/ui-design-contract.test.mjs`，13 项通过。
- `npx --yes esbuild ui-design/src/*.jsx --outdir=/tmp/easyinterview-ui-check --format=iife`，所有 JSX 入口解析通过。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`，Header / INDEX zero drift。
- `python3 scripts/lint/check_md_links.py docs/ui-design`，通过。
- `git diff --check`，通过。
- `bash -n ui-design/run.sh`，退出码为 0，仅有本机 locale warning。
- `bash ui-design/run.sh --no-open` 后通过 `curl` 验证 `index.html`、`src/screen-practice.jsx`、`canvas.html` 可访问，并命中 `VoiceSessionSurface`、`表达层指标`、`实时转写` 与 `route="practice" mode="voice"`。
- bundled Playwright + 本机 Chrome 打开 `index.html#route=practice&mode=voice&modality=voice&sessionId=session-24`，确认实际页面出现 `表达层指标`、`实时转写`、`本轮题目`，且无页面级运行时错误。
- 已建立回归记录 [BUG-0004](../bugs/BUG-0004.md)。

## 2 会话中的主要阻点/痛点

1. “统一外层骨架”在上一轮被误解为只需要 route 折回，导致语音模式专属主体没有进入实现。
   - **证据**：用户明确反馈“不是把整个语音模式的页面删除”，并提供原始语音模式截图作为参照。
   - **影响**：需要二次返工，补回语音波形、实时转写和表达指标。

2. 之前的 UI 契约测试没有覆盖模式专属 Surface 的可见元素。
   - **证据**：新增测试前，`PracticeScreen` 只保留 `activeMode` 和切换按钮逻辑也能通过原有 11 项测试。
   - **影响**：route 和上下文正确并不等于用户可见体验完整。

3. `/change-intake` 再次低置信度匹配到 OpenAPI plan，而不是 `ui-design` 运行时原型主题。
   - **证据**：本轮 matcher 输出仍推荐 `openapi-v1-contract/002-fixtures-and-mock-source`。
   - **影响**：需要人工忽略错误 plan，直接以 `ui-design/src` 和 `docs/ui-design` 为 owner。

## 3 根因归类

1. 语义理解层没有把“统一骨架”和“保留变体主体”拆成两个独立验收点。
   - **类别**：no repo change needed

2. `docs/ui-design` 对语音模式 route 折回后的运行时边界描述不够具体，未明确旧 `voice` route 是兼容入口，目标渲染应发生在 `PracticeScreen` 内。
   - **类别**：spec-plan

3. `change-intake` 对静态 UI backlog / 原型修订的候选发现仍偏弱。
   - **类别**：skill

## 4 对流程资产的改进建议

1. 后续 UI 模式合并类修订，先列出“共用骨架保留项”和“模式专属主体保留项”，再写契约测试。
   - **落点**：no repo change needed
   - **优先级**：high

2. 在 `docs/ui-design` 的模块文档中持续保持 route 折回、兼容入口和实际 Surface 的三层边界，避免把兼容 hash 当成目标页面。
   - **落点**：spec-plan
   - **优先级**：medium

3. 增强 `change-intake` 对 `ui-design`、`canvas.html`、`PracticeScreen`、`VoiceSessionSurface`、`EasyInterview_UI_Revision_Backlog` 等关键词的 owner 识别。
   - **落点**：skill
   - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是把 UI 契约测试从“route 是否正确”扩展到“关键 Surface 是否仍可见”。本次已补入语音模式的回归断言，后续如果继续统一其它模式，应按同样方式先写可见元素契约。

其次是补强 `change-intake` 的 UI 原型路由识别，减少后续 UI 修订再次误投到 OpenAPI 或后端计划的概率。
