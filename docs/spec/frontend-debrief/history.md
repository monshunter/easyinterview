# Frontend Debrief History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-16

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-16 | 1.1 | 修正 frontend-debrief 的复盘面试 handoff 口径：`debrief` 是 `PracticeGoal`，不是 `PracticeMode`；frontend nav 仍只传 `practiceGoal='debrief'`，后续由 frontend-workspace-and-practice / backend-practice 使用合法 `mode IN ('assisted','strict')` 启动 session。同步修订 plan 001 Phase 0 依赖验证，避免 `mode='debrief'` 旧口径回流。 | [001-debrief-screen-and-handoff](./plans/001-debrief-screen-and-handoff/plan.md) |
| 2026-05-16 | 1.0 | 初始创建 Frontend Debrief owner spec：承接 engineering-roadmap §5.2 Debrief workstream 的前端业务域；锁定 18 条决策（D-1~D-18）覆盖正式 `debrief` route + 历史 `debrief_full` normalize alias / UI 真理源源级复刻 / 三步骤 stepper / 3 个 in-page picker modal / 文本模式 AI 推荐 (suggestDebriefQuestions) / 语音模式 UI shell (无真实 STT, P0 限定) / 跨模式共享 entries / createDebrief 提交 / 双轨 polling (getJob + getDebrief) / 失败态三态 (Failure/Missing/Timeout) / 复盘面试 handoff (nav practice with practiceGoal=debrief) / InterviewContext reducer 增量扩展 / DOM testid 命名 / 隐私红线 / 旧口径负向；Operation Matrix 包含 createDebrief / getDebrief / suggestDebriefQuestions / getJob / listTargetJobs / listResumes / listResumeVersions(resumeAssetId) / listPracticeSessions(Phase 0 addendum) / getTargetJob / getResumeVersion / getPracticeSession + createPracticePlan/startPracticeSession 负向断言；§6 验收标准 C-1~C-18 覆盖默认渲染 / 3 picker / AI 推荐 / 失败降级 / voice UI shell / createDebrief / polling happy / 三种 failure state / 分析渲染 / 复盘面试 handoff / source parity / visual parity / 旧口径负向 / 隐私红线 / BDD；派 plan `001-debrief-screen-and-handoff` v1.0 active；保留编号建议 `002-debrief-voice-integration-and-history` / `003-debrief-export-and-share`；scenario 编号占用 E2E.P0.065-069（5 个）。 | 001-debrief-screen-and-handoff |
