# TargetJob Parse Company Fallback 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复截图中的 JD 解析失败。关联 Bug：[BUG-0146](../bugs/BUG-0146.md)，owner plan：[backend-targetjob/001-targetjob-import-and-parse-bootstrap](../spec/backend-targetjob/plans/001-targetjob-import-and-parse-bootstrap/plan.md)。
- 成功证据：
  - `cd backend && go test ./internal/ai/aiclient/outputschema -run TestValidate -count=1` 通过。
  - `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_HappyPathAcceptsFencedJSON|TestParseExecutor_HappyPathCoalescesMissingCompanyName|TestParseExecutor_AIOutputInvalid_WhenRequirementsAreSemanticallyInvalid|TestParseExecutorAITaskRuns' -count=1` 通过。
  - `cd backend && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresDrainerAndAIClient|TestBuildTargetJobRuntimeWrapsParseAIWithObservability' -count=1` 通过。
  - `cd backend && go test ./internal/targetjob -count=1` 通过。
  - `test/scenarios/env-redeploy.sh backend` 与 `test/scenarios/env-verify.sh` 通过。
  - Real local API smoke：同款 JD 重新导入后返回 `analysisStatus='ready'`、title `AI 应用技术负责人`、company `未提供`、14 条 requirements。
  - Real browser smoke：authenticated `/parse?...targetJobId=019f44a1-b43e-754f-ba0b-3cd9ed11ce1f` 渲染 `route-parse`，未出现 `JD 解析失败`，Title/Company 输入框值分别为 `AI 应用技术负责人` / `未提供`。

## 2 会话中的主要阻点/痛点

- TargetJob parse path 缺少 `ai_task_runs` 证据。
  - **证据**：截图 target 的 `async_jobs` / outbox 能看到 `AI_OUTPUT_INVALID`，但修复前 `ai_task_runs` 没有对应 jd_parse 记录。
  - **影响**：无法直接确认 provider 返回的是格式错误、schema drift，还是业务字段空值；需要先补 observability 才能定位真实原因。
- `companyName` 被当成硬必填 identity。
  - **证据**：补 observability 后，同款 JD 的真实 provider 输出有 title 和 requirements，但 `companyName` 为空；旧 parser 将其判为 `AI_OUTPUT_INVALID`。
  - **影响**：合法但未披露公司名的 JD 被整单标记 failed，用户只能看到失败页。
- JSON 容错边界需要精确化。
  - **证据**：provider 可能返回完整 fenced JSON；同时 BUG-0095 已证明不能接受 trailing prose。
  - **影响**：需要共享 normalize 行为，既避免误杀完整 fenced JSON，又不能恢复宽松解析。

## 3 根因归类

- 根因：BUG-0145 后的 identity contract 没区分必需岗位名和可选公司展示字段。
  - **类别**：spec-plan。
- 根因：TargetJob `cmd/api` runtime 没有沿用 A3 observability decorator。
  - **类别**：spec-plan。
- 根因：strict JSON validation 与 provider fenced JSON 的兼容边界没有被 owner gate 明确覆盖。
  - **类别**：spec-plan。

## 4 对流程资产的改进建议

- 对 AI parse owner plan，凡字段来自 JD 原文且可能缺失，应标明“hard identity required”还是“display fallback allowed”。
  - **落点**：spec-plan
  - **优先级**：high
- 对所有真实 provider runtime，closeout gate 应检查 `ai_task_runs` 是否覆盖该 task type，避免只有部分业务路径有 A3 observability。
  - **落点**：spec-plan
  - **优先级**：high
- 对 provider JSON normalization，统一在 A3 shared helper 中维护，业务 decoder 不应另写一套更宽松的字符串裁剪逻辑。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值后续动作：对其它 AI parse / generation runtime 做一次窄扫，确认每个 task type 都有 `ai_task_runs` 证据，并检查可选展示字段是否被误设为硬失败条件。
- 可延后动作：把 fenced JSON normalization 的允许/拒绝样例加入 A3 输出 schema contract 文档，减少后续业务包重复实现。
