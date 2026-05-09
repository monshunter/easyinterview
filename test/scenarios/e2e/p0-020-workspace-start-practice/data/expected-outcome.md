# Expected Outcome
- createPracticePlan body 含 targetJobId/goal='baseline'/mode='assisted'/hintsEnabled 由 practiceMode 派生
- Idempotency-Key 双键稳定派生，retry 复用
- nav practice 携带完整 InterviewContext + PracticeDisplayContext
- pendingAction.params 仅含 IDs/route/PracticeDisplayContext/autoStartPractice，不含敏感字段
- Vitest Tests all passed
