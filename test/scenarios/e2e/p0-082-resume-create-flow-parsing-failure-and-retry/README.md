# E2E.P0.082 Resume Create Flow Parsing Failure / Timeout / Cancel-and-Return

> **场景 ID**: E2E.P0.082
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Resumes/registerResume.json default`、`Resumes/getResume.json default`
- Mock harness 使用 attempt-aware `parseStatus` 序列模拟：`processing → failed (AI_TIMEOUT_RETRYABLE)` 与 `processing × 8 → PARSE_TIMEOUT`
- 用户已登录 lang=zh-CN

## 2 When

- Paste tab → submit → ParseFlow → polling 模拟 failed
- 点击 "重试解析" → polling 模拟 ready
- Paste tab → submit → polling 模拟 8 attempt processing → 超过上限触发 PARSE_TIMEOUT
- 任一 ParseFlow 阶段点击 "取消并返回修改" → 验证输入保留

## 3 Then

- `resume-parse-failed-state` testid 命中 failed 路径
- errorCode 映射：`AI_TIMEOUT_RETRYABLE`、`PARSE_TIMEOUT` 文案
- "重试解析" 触发 polling 重启（不重新 registerResume）
- "取消并返回修改" 回到 `resume-create-flow` stage='input' 且原 rawText / pickedFile 保留
- 隐私：失败 toast 内容不含 raw text / parsedSummary / parsedTextSnapshot

## 4 Verification Entry

`scripts/trigger.sh` 调用：

- `src/app/screens/resume-workshop/create/hooks/useResumeParsingPolling.test.tsx`
- `src/app/screens/resume-workshop/create/ParsingStage.test.tsx`
- `src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx`

`scripts/verify.sh` 校验 trigger.log 内：

- `Test Files +\d+ passed` 匹配
- 关联 test file 名称命中
- failed-state / cancel-and-return / parseTimeout 关键 case 在 log 中出现

## 5 fixture / mock baseline

- 当前 fixture 缺少 `processing`/`failed`/`queued` 多态；本场景使用 attempt-aware mock client stepping 模拟
- retrospective 中提议 backend-resume followup 补 fixture（plan §6 R2）

## 6 baseline

- ParseFlow DOM anchors + failed-state branches + cancel preservation 在 Vitest 测试中断言

## 7 离线限制 / mock-first 标注

- 不真实启动 dev stack；所有 fetch 通过 mock transport
- `method=mock-fixture-client` 明确标注
