# Expected Outcome
- WorkspaceInsightCard 点击后仍停留在 workspace 并保留 targetJobId/jdId，不调用独立 insight API
- 当前规划记录 Empty placeholder, 点击不触发 nav("report")
- 不读取 TargetJob.recentSessions
- getFeedbackReport 调用 0
- 非当前 prototype testid + route alias grep 0
- Vitest Tests all passed
