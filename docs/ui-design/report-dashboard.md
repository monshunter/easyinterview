# 报告仪表盘目标结构

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-07-12

## 1 目标

报告以整场 conversation 为分析单位，帮助用户判断准备度、理解能力表现、查看证据并选择复练当前轮或进入下一轮。报告不再按题目组织。

## 2 页面结构

```text
ReportDashboard(sessionId, reportId)
├─ Back
├─ Header
│  ├─ breadcrumb
│  ├─ title / subtitle
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ ContextStrip
│  ├─ session
│  ├─ target / round
│  └─ resume
├─ SummaryCards
│  ├─ 准备度
│  ├─ 能力维度
│  ├─ 会话证据
│  └─ 下一步
└─ DetailSurface
   ├─ 准备度
   ├─ 能力维度
   ├─ 会话证据
   └─ 下一步
```

## 3 生成态

生成态五阶段：

1. 整理会话上下文。
2. 提取能力证据。
3. 评估能力维度。
4. 归纳改进重点。
5. 生成行动建议。

不得出现逐题抽取、题目回顾、每题评分等旧文案。

## 4 Summary Cards

| Card | 内容 | 进入详情 |
|------|------|----------|
| 准备度 | readiness tier | Readiness tab |
| 能力维度 | dimension count / summary | Dimensions tab |
| 会话证据 | highlights + issues 数量 | Evidence tab |
| 下一步 | 推荐动作 | Next tab |

## 5 Detail Surface

### 5.1 Readiness

展示当前 tier、核心判断和最优先改进项，不展示精确通过率。

### 5.2 Dimensions

按能力维度展示 `status / confidence`，证据入口跳到 Evidence tab，不跳题目。

### 5.3 Evidence

展示 highlights / issues 的 dimension、evidence summary 和 confidence。不得复制完整 transcript，不按题号或 turn 分组。

### 5.4 Next

展示复练当前轮与进入下一轮的路径说明、能力重点和行动清单；CTA 仍只存在于 Header。

## 6 Replay

- `复练当前轮`：使用 report 的 `retryFocusCompetencyCodes` 创建新 plan/session。
- `进入下一轮`：从当前 `TargetJob.summary.interviewRounds[]` 按 `sequence` 排序后的列表中选择紧邻下一轮，使用该轮 id/name/duration 创建新 plan/session。
- 轮次列表产生重复派生 ID、当前轮是末轮/单轮、轮次为空、当前 `roundId` 未命中、TargetJob 仍在加载或加载失败时，`进入下一轮` disabled，不得回退第一轮、当前轮或固定默认轮次。
- 任一 replay/next start 进行中时两枚 CTA 都 disabled；重复点击最多创建一次 plan/session。
- 不传 `retryFocusTurnIds`、question IDs 或 per-question selection。

## 7 状态

- Missing session/report：专用空态。
- Queued/generating：留在 generating。
- Failed/not found/timeout：typed error + retry/back。
- Empty evidence：Evidence tab 显示空态，其他 tab 仍可用。

## 8 负向边界

当前 UI、fixtures、tests、scenarios 和文档中不得保留正向：

- Questions tab / 题目回顾页。
- questionAssessments / retryFocusTurnIds。
- 题数 summary card。
- per-question replay toggle。
- hint/practiceMode/phone modality context。
- 独立错题本、精确通过率或 timeline。

## 9 验收标准

| ID | Given | When | Then |
|----|-------|------|------|
| R-1 | ready report | 打开 report | 四卡四 tab，默认 readiness |
| R-2 | 有 dimensions/evidence | 切换 tabs | 展示会话级分析，无题目结构 |
| R-3 | needs practice | 点击复练当前轮 | competency focus 创建 fresh session |
| R-4 | next round available | 点击进入下一轮 | next-round fresh session |
| R-5 | desktop/mobile | parity gate | DOM、geometry、screenshot 与原型一致 |
| R-6 | final/single/empty/unknown/loading round state | 查看或点击进入下一轮 | CTA disabled 且不创建 plan/session；无 fallback |

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.16 | 下一轮只使用 TargetJob 有序结构化轮次的紧邻后一项；末轮、未知/缺失/加载失败和重复点击 fail closed。 |
| 2026-07-12 | 1.15 | 删除题目回顾和逐题 replay，报告收敛为 readiness/dimensions/evidence/next。 |
