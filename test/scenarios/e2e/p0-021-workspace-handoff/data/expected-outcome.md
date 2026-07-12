# Expected Outcome
- workspace runtime 不调用独立 insight API 或 report API
- workspace runtime 不读取 TargetJob.recentSessions 或 prototype helper
- report replay handoff regression 由 report owner 测试覆盖
- shared start-practice 使用当前结构化轮次时长，拒绝未知轮次并不复用预算不匹配的旧 plan
- out-of-scope prototype testid + import grep 0
- Vitest Tests all passed
