# Practice Voice MVP Spec

> **版本**: 1.15
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 背景与目标

当前 P0 先闭环连续文本面试、会话级报告与复练流程，电话模式暂不开放。用户界面保留一个不可点击的置灰电话图标，用于表达未来可能性；后端 voice operation 必须 fail-closed。通用 STT/TTS provider foundation 可以保留，但不得被 Practice 产品路径调用。

## 2 范围

### 2.1 In Scope

- `PracticeScreen` Top Bar 展示 disabled phone icon。
- disabled 控件具备 `disabled`、`aria-disabled=true` 和本地化“暂未开放”说明。
- `mode=phone` / `modality=phone` 等输入不得 materialize PhoneSurface，统一回到文本会话。
- `createPracticeVoiceTurn` 当前返回 typed `AI_UNSUPPORTED_CAPABILITY`。
- disabled 路径不读取/留存音频，不调用 STT / chat / TTS，不写 voice event、committed context 或 TTS metadata。
- `practice.voice.stt.default`、`practice.voice.tts.default`、realtime profile 保持 disabled / unsupported。
- 通用 speech adapters、capability vocabulary 和 provider tests继续作为基础设施存在。

### 2.2 Out of Scope

- 任何可用电话模式、麦克风权限、录音、VAD、TTS 播放、字幕、barge-in 或挂断交互。
- 真实 speech provider UAT。
- 独立 voice route 或 phone session state。
- 语音分析、录音留存与报告 communication modality。

## 3 用户决策

| ID | 决策 | 结论 | 理由 |
|----|------|------|------|
| D-1 | 当前是否开放电话模式 | 否 | 先跑通文本主流程，避免维护尚未验证的高成本分支 |
| D-2 | 是否删除通用 speech foundation | 否 | 保留可复用 provider 底座，不把它误当成已开放产品能力 |
| D-3 | 前端入口 | 保留置灰图标 | 明确当前不可用，不让用户进入残缺流程 |
| D-4 | 后端行为 | fail-closed | 防止手工 API 调用绕过 UI 并产生 provider/持久化副作用 |

## 4 设计约束

- disabled UI 必须先在 `frontend/src` 落地，再由正式前端按设计合同实现。
- disabled icon 不注册 click handler，不改变 route/context，不请求 voice API。
- backend fail-closed 必须发生在音频 decode、profile resolution、provider call 与 store mutation之前。
- 禁用期间删除/改写所有把 voice happy path 当作当前 P0 完成条件的正向文档与场景；保留一个 disabled negative scenario。
- 重新启用需要新的用户确认和 Product/UI/OpenAPI/Privacy/Provider 联合设计。

## 5 Operation Matrix

| operationId | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticeVoiceTurn` | none while disabled | existing Practice voice handler with leading disabled guard | none | none | handler/component contract tests + root `make test` |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | disabled icon | 文本会话 | 用户查看/键盘聚焦电话图标 | 图标置灰且不可触发模式变化 | 001 |
| C-2 | phone 参数负向 | URL 含 phone params | 页面加载 | 仍渲染同一文本聊天，无 PhoneSurface | 001 |
| C-3 | API fail-closed | voice endpoint 可寻址 | 提交任意请求 | typed unsupported，零 provider 调用、零写入 | 001 |
| C-4 | provider config | 应用启动 | 读取 profiles | STT/TTS/realtime 保持 disabled/unsupported | 001 |
| C-5 | 旧 surface 负向 | 全仓扫描 | 检查当前 UI/BDD | 无麦克风、字幕、VAD、TTS、barge-in、hangup 正向合同 | 001 |

## 7 关联计划

- [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md)

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.15 | 电话模式改为前端置灰、后端 fail-closed；当前 P0 只验证禁用边界，保留通用 speech foundation。 |
