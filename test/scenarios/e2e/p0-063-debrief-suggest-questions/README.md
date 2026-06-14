# E2E.P0.063 suggestDebriefQuestions Sync + AI Failure

> **场景 ID**: E2E.P0.063
> **关联需求**: backend-debrief C-9, C-10
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

用户在 debrief 输入前希望基于目标岗位、关联模拟面试和绑定简历生成推荐复盘问题。

## When

调用 `suggestDebriefQuestions` with `targetJobId` + optional `sessionId` + optional `resumeId`，并分别覆盖 successful AI、F3 failure、A3 timeout、invalid output。

## Then

成功时返回 suggestions，`sessionId` / `resumeId` 进入真实 handler → service → store context 并注入 prompt；失败时映射 B1 AI error code，并写入 `debrief_suggest_questions` task run 与 safe audit metadata。
