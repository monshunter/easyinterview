# Practice Phone Session Language 交付复盘报告

> **日期**: 2026-07-11
> **审查人**: Codex

**关联 Bug**: [BUG-0158](../bugs/BUG-0158.md)

**关联计划**: [practice-voice-mvp/001](../spec/practice-voice-mvp/plans/001-cascaded-stt-llm-tts/plan.md), [frontend-workspace-and-practice/002](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md), [backend-practice/002](../spec/backend-practice/plans/002-event-loop-and-completion/plan.md), [backend-practice/003](../spec/backend-practice/plans/003-mode-policies-and-provenance/plan.md)

## 1 复盘范围与成功证据

- 交付范围：删除电话模式 restart / `callEnded` 与分段/live/cut-off 旧契约；文本态使用单 handset 进入电话，电话态顶部 handset 与中央红色挂断共用同一退出路径并回到同一个文本 session。
- 会话完整性：前端用 call-scoped microphone/utterance recorder、VAD silence submit、TTS-ended re-arm、真实 speech-start barge-in 和 session-keyed remount锁定生命周期；后端文本与语音共用 server-owned canonical question generator、persisted session language、exactly-one repair 和 typed failure recovery。
- 数据真实性：正式 UI 通过 generated `getPracticeSession` / `getTargetJob` 显示后端 session/turn 与 TargetJob，公司、岗位和问题均为持久化中文事实；raw `questionIntent` 与 fixture dialogue 不进入用户界面。
- 自动化证据：前端 `144 files / 889 tests`、typecheck/build、UI contract `45/45`、Practice pixel parity `11 passed / 1 conditional skip`；后端 `go test ./... -count=1`、scoped staticcheck、prompt/eval、OpenAPI/fixture/codegen、migration、privacy/runtime-boundary gates 全部通过。
- BDD 证据：P0.007、P0.008、P0.009、P0.045、P0.046、P0.048、P0.050、P0.051 wrapper 通过；P0.038-P0.043 与 P0.048-P0.051 direct Go E2E 通过。
- 真实环境证据：host-run 环境验证通过；浏览器记录真实 `/api/v1/practice/sessions/{sessionId}` 和 `/api/v1/targets/{targetJobId}` 请求。三张 `1440x900` 截图位于 `.test-output/practice-phone-session-flow-0711/`；中央挂断前后 URL 保留同一 `sessionId`，数据库 session 仍为 `running`、turn 仍为 `asked`。
- 文档证据：四个 owner checklist 无遗留未勾选项；backend-practice/002 与 /003 恢复 `completed`；四个 context、Header/INDEX、Markdown link 与 diff whitespace gates 通过。

## 2 会话中的主要阻点/痛点

- 用户现象看起来是三个 UI/内容问题，但 artifact-level reconcile 暴露了跨层契约漂移。
  - **证据**：旧 restart/call-ended 口径同时存在于原型、正式前端、pixel parity、后端语言/生成路径、fixture 和 P0 场景；L2 复查还发现 voice context 证据、事件 metadata、answer summary persistence 和 session A→B state ownership 缺口。
  - **影响**：若只删两个按钮并替换图标，真实会话仍可能生成混合语言、接受客户端上下文、在挂断后继续 TTS 或用旧 turn 回调，形成表面修复。

- 历史 PASS 和旧真实截图曾把已过时的 restart surface 固化为成功证据。
  - **证据**：Phase 6 checklist/screenshot 明确把 captions + hang-up + restart 当作正向合同；本轮必须新开 Phase 7/10 并声明旧证据 superseded，才能防止 completed 状态继续掩盖当前需求。
  - **影响**：没有 current-state negative search 与真实截图刷新时，旧验证越完整，反而越容易阻止正确的产品修订。

- 真实浏览器验收的数据准备仍不稳定，且这是 2026-07-09 复盘已指出但尚未落地的重复成本。
  - **证据**：旧验收 URL 属于旧测试用户，新的合成账号得到正确的 404/session-lost；随后需要完成 Mailpit 登录、profile setup，并为新账号创建最小 TargetJob/Plan/Session/Turn/Event 数据才能继续截图。
  - **影响**：增加验收时间，也容易诱发复用旧 cookie/session、越权修改旧数据或把鉴权隔离误判成页面回归。

- “真实截图”与“真实音频 provider 验证”的边界需要显式说明。
  - **证据**：本地 STT/TTS profile 未启用，自动化浏览器没有麦克风设备；电话 shell、真实 API 数据和同会话切换可验收，但截图不能证明 provider 音频成功。
  - **影响**：若不写明，截图 PASS 可能被误读成完整 STT/TTS UAT；若用浏览器 mock 隐藏错误，又会制造假绿。

- 本轮工具操作曾让 ignored runtime secret 出现在工具输出中。
  - **证据**：本地环境检查阶段不必要地读取了 ignored `.env` 内容；后续所有数据库命令改为静默 `source`，并停止读取 cookie/state 内容。
  - **影响**：需要轮换已暴露的 provider key、session cookie secret 与 auth challenge pepper；任何文档、Bug、报告和日志都必须只记录变量名，不记录值。

## 3 根因归类

- 产品交互修订没有在旧 completed evidence 上建立“superseded + negative inventory + current screenshot”三联门禁。
  - **类别**：spec-plan

- 文本/语音问题生成和前端音频生命周期原先由不同局部测试覆盖，缺少以 session ownership 与语言为主轴的跨层不变量。
  - **类别**：spec-plan

- 本地真实 Practice 验收缺少 repo-tracked、幂等、按合成用户隔离且自带清理的最小 seed/bootstrap helper。
  - **类别**：README / scenario tooling

- 真实 phone shell 截图与启用 provider 的音频 UAT 没有被拆成两个清楚的验收层级。
  - **类别**：README / spec-plan

- secret 输出来自一次不必要的诊断操作；现有脱敏规范和 PATTERNS 已足够，当前不需要新增仓库规则，执行层面必须轮换并避免再次读取明文。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 为 `test/scenarios/` 增加一个真实 Practice screenshot bootstrap helper：创建唯一合成邮箱、完成 profile、以单事务或正式 API 建立最小 TargetJob/Plan/Session/Turn/Event，输出非敏感 ID，并只按“专属 ID + 合成用户 ownership”清理。
  - **落点**：`test/scenarios/README.md` / scenario tooling
  - **优先级**：high

- 在用户可见 phone 修订的 owner checklist 中长期保留 current-state 三联门禁：旧正向证据显式 superseded、removed symbol/copy exact negative search、真实 API 同-session screenshot + DB state assertion。
  - **落点**：spec-plan
  - **优先级**：high

- 把 phone 验收拆成两层：层 1 验证真实 API shell、权限失败、同 session 切换与截图；层 2 只在 STT/chat/TTS profile 启用时执行真实 provider/microphone UAT，并记录 provider/profile/result metadata 而不保存音频或秘密。
  - **落点**：`test/scenarios/README.md` / practice-voice spec-plan
  - **优先级**：medium

- 对 AI 用户可见文本统一增加“persisted session language wins + wrong-language negative fixture + exactly-one repair + no canned fallback”矩阵，文本、语音、hint 必须共用同一语言判定器或等价不变量。
  - **落点**：backend-practice spec-plan
  - **优先级**：high

- 在本地环境操作范例中只允许静默加载 `.env`，诊断变量是否存在时输出 `SET/UNSET`，禁止打印值；已暴露的本地密钥立即轮换。
  - **落点**：无需仓库改动（执行动作）
  - **优先级**：high

## 5 建议优先级与后续动作

- highest：轮换本地 AI provider key、session cookie secret 与 auth challenge pepper；这不改变仓库代码，但应在继续真实 provider UAT 前完成。
- high：用 `/change-intake` 命中现有 `test/scenarios` owner，落实幂等 synthetic Practice screenshot bootstrap helper；这项建议在两次真实会话复盘中重复出现，已从便利性优化升级为稳定验收前置能力。
- high：保持 BUG-0158 中的跨层语言矩阵和同-session negative gates，不再接受“隐藏控件 + 历史 PASS”作为电话修订完成证据。
- medium：密钥轮换且 STT/TTS profile 启用后，再通过 `/scenario-run` 执行独立真实 provider phone smoke；在此之前，本轮截图结论严格限定为 UI、真实 API 数据与会话切换已通过。
