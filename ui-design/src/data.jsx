// Mock data for EasyInterview prototype — 贴近 spec 的真实示例
const jdSampleInterviewRounds = [
  {
    sequence: 1,
    type: "hr",
    name: "HR 初筛",
    durationMinutes: 20,
    focus: "HR 会围绕求职动机、B 端产品兴趣和节奏稳定性追问",
  },
  {
    sequence: 2,
    type: "technical",
    name: "技术一面",
    durationMinutes: 45,
    focus: "技术一面会聚焦 Design System 推进、TS 类型设计和性能证据",
  },
  {
    sequence: 3,
    type: "technical",
    name: "技术二面",
    durationMinutes: 60,
    focus: "技术二面会追问 Monorepo / 微前端架构取舍与大型协作案例",
  },
  {
    sequence: 4,
    type: "manager",
    name: "经理面",
    durationMinutes: 40,
    focus: "经理面会关注跨团队影响力、冲突处理和目标一致性",
  },
];

window.EI_DATA = {
  user: {
    name: "林舟",
    email: "lin.zhou@example.com",
    avatar: "LZ",
    locale: "zh-CN",
    years: 5,
    title: "高级前端工程师",
  },

  targetJobs: [
    {
      id: "tj-1",
      title: "资深前端工程师",
      company: "星环科技",
      location: "上海 · 混合办公",
      language: "中文",
      status: "面试中",
      statusTone: "amber",
      level: "P6",
      updatedAt: "2 小时前",
      readiness: 3, // 0-3: 未就绪 / 基本可面 / 建议再练 / 较为充分
      readinessLabel: "建议再练",
      practiceProgress: {
        status: "in_progress",
        completedRounds: [
          { roundId: "round-1-hr", roundSequence: 1 },
          { roundId: "round-2-technical", roundSequence: 2 },
        ],
        currentRound: { roundId: "round-3-technical", roundSequence: 3 },
      },
      hits: ["React 深度", "性能优化", "可访问性"],
      gaps: ["大型协作案例", "Design System 落地故事"],
      practices: 4,
      match: 78,
    },
    {
      id: "tj-2",
      title: "Frontend Platform Engineer",
      company: "Lumen Labs",
      location: "远程 · 新加坡时区",
      language: "英文",
      status: "准备中",
      statusTone: "neutral",
      level: "Senior",
      updatedAt: "昨天",
      readiness: 2,
      readinessLabel: "基本可面",
      practiceProgress: {
        status: "not_started",
        completedRounds: [],
        currentRound: { roundId: "round-1-hr", roundSequence: 1 },
      },
      hits: ["TypeScript", "Monorepo"],
      gaps: ["英文表达节奏", "Platform 案例"],
      practices: 2,
      match: 64,
    },
    {
      id: "tj-3",
      title: "技术专家（Web 架构）",
      company: "云栖集团",
      location: "杭州",
      language: "中文",
      status: "草稿",
      statusTone: "muted",
      level: "P7",
      updatedAt: "3 天前",
      readiness: 1,
      readinessLabel: "未就绪",
      practiceProgress: {
        status: "completed",
        completedRounds: [
          { roundId: "round-1-hr", roundSequence: 1 },
          { roundId: "round-2-technical", roundSequence: 2 },
          { roundId: "round-3-technical", roundSequence: 3 },
          { roundId: "round-4-manager", roundSequence: 4 },
        ],
        currentRound: null,
      },
      hits: ["系统设计"],
      gaps: ["跨团队影响", "技术决策案例"],
      practices: 0,
      match: 52,
    },
  ],

  jdSample: {
    title: "资深前端工程师",
    company: "星环科技",
    mustHave: [
      "5 年以上大型 Web 应用开发经验",
      "精通 React / TypeScript 与现代构建工具链",
      "主导过性能优化与可访问性改造",
    ],
    nice: [
      "Design System / 组件库落地经验",
      "有跨团队技术推动经验",
      "熟悉可观测性与前端监控",
    ],
    hidden: [
      "岗位上下文偏向 B 端复杂表单与数据可视化",
      "团队近期在推 Monorepo + 微前端架构",
      "面试官团队倾向听到可量化的改进故事",
    ],
    interviewRounds: jdSampleInterviewRounds,
    rounds: jdSampleInterviewRounds,
  },

  sessionTranscript: [
    { role: "ai", text: "你好林舟，我会扮演星环科技的技术面试官。准备好后我们就开始。", t: "00:00" },
    { role: "user", text: "好的，我准备好了。", t: "00:12" },
    { role: "ai", text: "先介绍一下你自己，也聊聊为什么对这个岗位感兴趣。", t: "00:18" },
    { role: "user", text: "我叫林舟，在电商行业做了 5 年前端，主要做后台系统和复杂表单。我对这个岗位感兴趣是因为…", t: "00:35" },
    { role: "ai", text: "你提到「复杂表单」，能举一个具体场景吗？比如最棘手的那次。", t: "01:42" },
    { role: "user", text: "有一个就是我们的订单改价系统，涉及 40+ 字段…", t: "01:55" },
  ],

  experiences: [
    {
      id: "exp-1",
      title: "复杂表单平台性能治理",
      company: "星环科技",
      situation: "订单改价系统字段多、校验链路长，用户在高峰期频繁遇到卡顿。",
      task: "在不重写业务流程的前提下，把关键交互延迟降到可接受范围。",
      action: "拆分表单状态、延迟非关键校验、补齐性能埋点，并和后端约定批量校验接口。",
      result: "关键路径交互延迟下降 38%，客服反馈的卡顿工单两周内下降一半。",
      skills: ["React", "Performance", "Observability"],
      language: "zh-CN",
    },
    {
      id: "exp-2",
      title: "Design System 渐进式落地",
      company: "Acme",
      situation: "多个业务线组件风格不一致，发布节奏互相阻塞。",
      task: "建立可复用组件和迁移节奏，避免一次性大规模重构。",
      action: "先沉淀 token 与表单组件，再分批接入三个高频页面。",
      result: "新页面交付周期缩短 25%，设计走查返工次数明显减少。",
      skills: ["Design System", "TypeScript", "Collaboration"],
      language: "zh-CN",
    },
  ],

  reportGeneration: {
    id: "report-24",
    targetJobId: "tj-1",
    status: "generating",
    errorCode: null,
  },

  reportOverviewFixtures: {
    ready: {
      state: "ready",
      targetJobId: "tj-1",
      rounds: [
        {
          round: { roundId: "round-1-hr", roundSequence: 1 },
          currentReport: null,
          latestAttempt: null,
        },
        {
          round: { roundId: "round-2-technical", roundSequence: 2 },
          currentReport: { id: "report-21", generatedAt: "2026-07-13T14:20:00Z" },
          latestAttempt: { id: "report-22", status: "failed", errorCode: "AI_PROVIDER_TIMEOUT", createdAt: "2026-07-14T09:12:00Z" },
        },
        {
          round: { roundId: "round-3-technical", roundSequence: 3 },
          currentReport: null,
          latestAttempt: { id: "report-23", status: "generating", errorCode: null, createdAt: "2026-07-14T09:16:00Z" },
        },
        {
          round: { roundId: "round-4-manager", roundSequence: 4 },
          currentReport: { id: "report-24", generatedAt: "2026-07-14T09:20:00Z" },
          latestAttempt: { id: "report-24", status: "ready", errorCode: null, createdAt: "2026-07-14T09:18:00Z" },
        },
      ],
    },
    loading: { state: "loading" },
    error: { state: "error" },
  },

  reportConversation: {
    state: "ready",
    reportId: "report-24",
    reportStatus: "ready",
    context: {
      sourcePlanId: "plan-tj-1",
      targetJobTitle: "高级前端工程师",
      targetJobCompany: "星环科技",
      resumeId: "frontend-v3",
      resumeDisplayName: "前端工程师简历 · 第 3 版",
      roundId: "round-2-technical",
      roundSequence: 2,
      roundName: "技术一面 · 45m",
      roundType: "technical",
      language: "zh-CN",
      hasNextRound: true,
    },
    messages: [
      {
        sequence: 1,
        role: "assistant",
        content: "## 技术取舍追问\n\n请先说明约束、候选方案，以及最终选择的理由。",
        createdAt: "2026-07-15T08:00:00Z",
      },
      {
        sequence: 2,
        role: "user",
        content: "我会按 **STAR** 结构回答。\n\n| 阶段 | 成功指标 | 回滚条件 |\n| --- | --- | --- |\n| 灰度 | 核心链路成功率保持在 99.99% 且连续观察两个完整窗口 | 任一关键指标连续三个窗口低于基线立即回滚 |\n\n```ts\nconst rollout = \"baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-baseline-\";\n```",
        createdAt: "2026-07-15T08:00:18Z",
      },
    ],
  },

  report: {
    id: "report-24",
    targetJobId: "tj-1",
    sessionId: "session-24",
    status: "ready",
    language: "zh-CN",
    summary: "你能用具体项目说明自己的行动，结构也比较清楚；但关键技术取舍与量化结果仍不够完整，建议先补强这些证据再进入下一轮。",
    preparednessLevel: "needs_practice",
    context: {
      sourcePlanId: "plan-tj-1",
      targetJobTitle: "高级前端工程师",
      targetJobCompany: "星环科技",
      resumeId: "frontend-v3",
      resumeDisplayName: "前端工程师简历 · 第 3 版",
      roundId: "round-2-technical",
      roundSequence: 2,
      roundName: "技术一面 · 45m",
      roundType: "technical",
      language: "zh-CN",
      hasNextRound: true,
    },
    dimensionAssessments: [
      { code: "answer_structure", label: "回答结构", status: "strong", confidence: "high" },
      { code: "technical_tradeoffs", label: "技术取舍", status: "needs_work", confidence: "high" },
      { code: "role_relevance", label: "岗位相关性", status: "meets_bar", confidence: "medium" },
    ],
    highlights: [
      { dimensionCode: "answer_structure", evidence: "你按背景、行动和结果说明了订单改价系统的治理过程，面试官可以顺着叙述理解关键步骤。", confidence: "high" },
      { dimensionCode: "role_relevance", evidence: "你把复杂表单与设计系统经验直接对应到岗位要求中的中后台交付与组件治理。", confidence: "medium" },
    ],
    issues: [
      { dimensionCode: "technical_tradeoffs", evidence: "被追问替代方案时，你说明了最终方案，却没有比较其他方案的成本、风险与放弃原因。", confidence: "high" },
      { dimensionCode: "technical_tradeoffs", evidence: "性能优化案例提到了结果改善，但没有给出起点指标、结束指标与影响范围。", confidence: "high" },
    ],
    nextActions: [
      { type: "retry_current_round", label: "先复练当前轮：补充一个包含替代方案比较和量化结果的技术案例。" },
      { type: "review_evidence", label: "回看简历中的性能治理经历，整理可核对的指标与个人责任边界。" },
    ],
    retryFocusDimensionCodes: ["technical_tradeoffs"],
  }
};
