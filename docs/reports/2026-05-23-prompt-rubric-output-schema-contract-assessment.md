# Prompt Rubric Output Schema Contract 交付复盘报告

> **日期**: 2026-05-23
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：实施 `prompt-rubric-registry/002-output-schema-contract`，为 13 个 chat feature_key 建立语言无关 output schema，生成 26 个 prompt body schema contract block，同步 prompt hash、seed migration、registry metadata、A3 schema validator enum 与各业务 caller 的 `CallMetadata.OutputSchema`。
- Prompt / lint 证据：`python3 -m pytest scripts/lint/prompt_lint_test.py -q` 通过；`make lint-prompts` 通过，`prompt_lint: 26 files clean`。
- Backend 证据：`go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/targetjob/... ./backend/internal/resume/jobs/... ./backend/internal/review/... ./backend/internal/practice/... ./backend/internal/debrief/... ./backend/internal/jdmatch/... -race` 通过。
- Repo gate 证据：`make lint`、`git diff --check`、`python3 .agent-skills/implement/scripts/validate_context.py --target backend`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`make docs-check` 均通过。
- 红线证据：语音 STT/TTS feature_key 无 schema；`backend/internal` 未引入 provider request 层 `response_format` / `json_schema` 字段。
- 关联 Bug：[BUG-0094](../bugs/BUG-0094.md)。

## 2 会话中的主要阻点/痛点

- Prompt 输出契约覆盖面宽，手工维护 prompt body 与 schema 容易漂移。
  - **证据**：同一批 feature 需要同时维护 13 个 schema、26 个语言 prompt body、YAML hash 与 seed migration。
  - **影响**：若没有 schema-rendered contract block 和 prompt lint negative fixture，后续很容易出现 prompt 文案与 schema 不一致但 hash 仍更新的 false-green。
- 本地 `make lint` 被 ignored runtime `.env` 阻塞。
  - **证据**：`lint-secrets-pattern` 通过 gitleaks 二层扫描读取了 `.gitignore` 下的 `deploy/dev-stack/.env`。
  - **影响**：带真实本地配置的开发机无法稳定复现 repo-wide lint，且 secret scanner 输入集超出质量门禁应覆盖的仓库候选文件范围。
- 多域 caller metadata 需要逐包反查。
  - **证据**：targetjob、resume/jobs、debrief、review、practice、jdmatch 都有独立 PromptResolution adapter 或 call metadata 组装路径。
  - **影响**：只修 registry/resolver 会造成 schema metadata 停在 registry 层，A3 validator 和业务 caller 仍无法拿到契约。

## 3 根因归类

- Prompt contract 缺少 schema-first drift gate。
  - **类别**：spec-plan / lint
  - **说明**：计划已收敛为 schema-first 后，本次用 `prompt_lint` renderer 与 negative fixtures 固化该约束。
- Secret lint 输入集未限定为 git candidate set。
  - **类别**：README / test gate
  - **说明**：脚本层已修复，并沉淀为 [BUG-0094](../bugs/BUG-0094.md) 与 `docs/bugs/PATTERNS.md` 模式 7。
- Caller metadata 是跨域 contract，不是单点 registry 改动。
  - **类别**：spec-plan
  - **说明**：Phase 7 需要显式列出每个 caller owner，并用包级测试断言 `OutputSchema` 进入 AI call metadata。

## 4 对流程资产的改进建议

- 保持 schema-rendered prompt block 作为唯一 prompt 输出契约呈现方式，禁止手写字段表回流。
  - **落点**：`scripts/lint/prompt_lint.py` / prompt-rubric-registry plans
  - **优先级**：high
- 为 `scripts/lint/gitleaks.sh` 增加脚本级 regression test，构造 ignored `.env` 与未忽略 secret fixture，证明扫描范围与失败范围同时正确。
  - **落点**：`scripts/lint/`
  - **优先级**：medium
- 后续 `response_format` 计划应复用本次 registry `OutputSchema`，但继续保持 provider request wiring 与 schema authoring 分离。
  - **落点**：prompt-rubric-registry 后续 A3 plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：提交当前 output schema contract 实施补丁，并在 commit message 中关联 `BUG-0094`。
- 下一轮建议：进入 prompt-rubric-registry 后续 A3 provider request wiring plan，把 `OutputSchema` 从 metadata 进一步接到 provider `response_format`，同时保留 provider-request 红线直到该计划正式开始。
- 可延后处理：给 gitleaks scan mirror 增加自动化测试，降低未来 secret-lint 输入集回退风险。
