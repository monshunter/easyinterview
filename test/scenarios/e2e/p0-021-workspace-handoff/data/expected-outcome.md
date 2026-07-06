# Expected Outcome
- CompanyIntelEmbed 不调 getCompanyIntel，点击后仍停留在 workspace 并保留 targetJobId/jdId
- sessionHistory EmptyHistory/disabled placeholder, 点击不触发 nav("report")
- 不读取 TargetJob.recentSessions
- getFeedbackReport 调用 0
- 旧 prototype testid + route alias grep 0
- Vitest Tests all passed
