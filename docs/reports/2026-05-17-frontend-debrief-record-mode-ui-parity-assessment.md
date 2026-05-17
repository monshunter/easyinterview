# Frontend Debrief Record Mode UI Parity 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 `frontend-debrief/001-debrief-screen-and-handoff` 复盘记录 Step 0 的文本/语音模式 UI parity drift，使正式前端重新对齐 `ui-design/src/screens-p1-depth.jsx::DebriefFullScreen` 中的 record workspace、文本引导卡片流、语音 intro card、点击开始后的连续对话态与右侧 `整体感受` card。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx src/app/screens/debrief/DebriefPickerRegression.test.tsx src/app/screens/debrief/components/DebriefContextStrip.test.tsx src/app/screens/debrief/components/DebriefHeader.test.tsx src/app/screens/debrief/components/DebriefStepper.test.tsx src/app/i18n/__tests__/debriefI18nCoverage.test.ts`
  - `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts`
  - `pnpm --filter @easyinterview/frontend build`
  - `python3 -m pytest scripts/lint/frontend_debrief_legacy_test.py -q`
  - `git diff --check`
  - 旧语音占位口径反向搜索返回无命中。

## 2 会话中的主要阻点/痛点

- Step 0 parity gate 之前过于宽泛。
  - **证据**：修复前新增 mode-specific Playwright 断言时，缺失 `debrief-record-workspace` 与 `debrief-voice-intro-card`，但旧 gate 没有捕获。
  - **影响**：文本模式和语音模式的主要结构可漂移为扁平占位态，仍可能通过宽泛页面锚点和 screenshot smoke。
- 语音模式旧占位口径容易残留。
  - **证据**：初次修复后隐藏节点仍保留 `debrief-voice-not-implemented` 和 `Voice debrief integration coming soon` / `语音复盘集成中` copy。
  - **影响**：即使 UI 不可见，也会让测试锚点和 copy 继续表达旧实现边界，降低后续 parity 审查可信度。
- 语音 CTA 的点击路径没有纳入第一轮 parity gate。
  - **证据**：用户复核后指出 `开始语音复盘对话` 点击无反应；补充 Playwright red case 后确认找不到 `debrief-voice-chat`。
  - **影响**：intro card 表面接近原型，但核心交互仍停留在静态 shell，无法达到 `ui-design` 中连续对话和实时抽卡状态。

## 3 根因归类

- 根因：`frontend-debrief` 的 pixel parity test 只覆盖页面级锚点、viewport、theme computed values 与 smoke screenshot，没有把文本/语音 mode-specific 原型层级作为完成 gate。
  - **类别**：spec-plan
- 根因：旧占位态没有被列入 zero-reference 负向搜索。
  - **类别**：spec-plan
- 根因：语音模式只断言 intro card 存在，未断言 CTA 后的 chat / extraction / review 状态。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 保持本次新增的 `debrief.spec.ts` mode-specific 断言作为后续复盘 UI gate，任何新增 tab/mode 必须至少覆盖 shared workspace、mode panel、primary CTA 点击后的状态、right rail/mobile stacked 几何关系。
  - **落点**：spec-plan
  - **优先级**：high
- 后续处理复盘语音模式真实 STT/LLM/TTS 接入时，继续保留 intro/chat/review phase 的 DOM 锚点和旧占位零残留搜索，避免把真实语音能力接入成新的 flat placeholder。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：继续用 `tests/pixel-parity/debrief.spec.ts` 作为复盘页面 UI truth gate；语音真实能力接入前要保留 start click、chat/extraction、review/save 的前端 parity。
- 可延后：若后续再出现 mode-level UI drift，再把 BUG-0071 追加到 `docs/bugs/PATTERNS.md` 模式 4；本次不直接修改模式库。
