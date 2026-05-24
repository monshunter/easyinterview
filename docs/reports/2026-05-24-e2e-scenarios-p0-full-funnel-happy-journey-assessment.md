# E2E Scenarios P0 Full Funnel Happy Journey 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`e2e-scenarios-p0/001-full-funnel-happy-journey`，覆盖 `E2E.P0.098` API-level 真后端 journey、`E2E.P0.099` Playwright 全栈 journey、场景目录登记与 plan/BDD checklist 收口。
- 关键实现证据：
  - `cd backend && DATABASE_URL='postgres://user:***@localhost/easyinterview?sslmode=disable' go test -v ./cmd/api -run '^TestE2EP0098' -count=1` 通过。
  - `EI_PLAYWRIGHT_OUTPUT_DIR="$PWD/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright" pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts` 通过。
  - 两个 scenario 顺序 wrapper 通过：`p0-098-full-funnel-import-to-next-round-journey` 与 `p0-099-full-funnel-fullstack-ui-journey` 均输出 `setup: ok`、`trigger: ok`、`verify: ok`、`cleanup: ok`。
  - `cd backend && go test -v ./cmd/api -run '^TestE2EP0OperationMatrixPreflight$' -count=1` 通过，9 个 operation 子用例全部 PASS。
  - `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check` 均通过。
- 相关 bug：已创建 [BUG-0103](../bugs/BUG-0103.md)，记录空 `focusCompetencyCodes` 导致 practice plan 写入 NULL 数组并触发 500 的问题。

## 2 会话中的主要阻点/痛点

- 空数组持久化问题只在真后端全栈场景暴露。
  - **证据**：`E2E.P0.099` next-round CTA 的 `createPracticePlan` 发送 `focusCompetencyCodes: []` 后触发真实后端 500；修复后新增 store regression 与真实 DB `TestE2EP0098CreatePracticePlanAcceptsEmptyFocusCodes`。
  - **影响**：全栈 journey 中断，必须先补后端 practice store 契约再继续场景收口。
- BDD 文档中的脚本职责与真实实现有轻微漂移。
  - **证据**：原 `099.C/Given` 写成 `setup.sh` 拉起真后端和前端；最终实现由 Playwright `webServer` 在 `trigger.sh` 阶段托管后端测试进程、前端 build/preview 和 health probe。
  - **影响**：收口前需要同步 `bdd-plan.md` / `bdd-checklist.md`，否则 plan 证据与脚本事实不一致。
- wrapper 示例使用 `status` 变量不适合 zsh 环境。
  - **证据**：第一次顺序 wrapper 在 zsh 中因 `status` 是只读变量而失败，未进入真实 scenario 执行。
  - **影响**：造成一次无效失败；已把文档示例改为 `overall`。
- legacy route 负向 grep 初始范围过宽。
  - **证据**：校验曾误报注释中的旧 route 词、测试文件中的 negative assertion，以及合法 props `experiences={[]}`。
  - **影响**：需要收窄 runtime route materialization regex，并让 P0.099 scoped scan 排除 test-only assertions，同时保留 canonical-token false-positive guard。

## 3 根因归类

- `focusCompetencyCodes` 空数组问题：
  - **类别**：no repo change needed beyond current fix
  - **原因**：store 测试此前未断言 nil slice 对 PostgreSQL `NOT NULL` array 的编码行为；已由 BUG-0103 和 regression 覆盖。
- P0.099 setup/trigger 职责漂移：
  - **类别**：spec-plan
  - **原因**：BDD plan 先假定 shell setup 负责拉起进程，后续工程实现选择 Playwright `webServer` 托管，文档未同步到最终结构。
- `status` wrapper 变量：
  - **类别**：spec-plan / README
  - **原因**：示例脚本没有考虑当前默认 shell 为 zsh，使用了 shell 保留变量。
- legacy-negative 误报：
  - **类别**：spec-plan / test
  - **原因**：负向 grep 把“旧 route 词汇存在”与“旧 route 可运行入口被物化”混为一类，且未区分 runtime surface 与 test-only assertion。

## 4 对流程资产的改进建议

- 在 `scenario-create` 模板或 `test/scenarios/README.md` 中统一 wrapper 变量命名为 `overall`，并标明示例命令应能在 zsh 下直接运行。
  - **落点**：skill / README
  - **优先级**：high
- 在 scenario 文档模板中允许 Playwright `webServer` 托管真实后端和前端进程，并要求 BDD `Given` 精确区分 `setup.sh` 的环境准备与 `trigger.sh` 的 runner-owned server lifecycle。
  - **落点**：skill / spec-plan
  - **优先级**：medium
- 为 route-aware legacy-negative gate 增加模板要求：必须列出 canonical false-positive samples，必须明确 scan scope，测试文件中的 negative assertions 不应作为 runtime drift 命中。
  - **落点**：skill / README
  - **优先级**：medium
- 对后端 store 中写入 PostgreSQL array / jsonb / nullable boundary 的字段，owner plan 的 regression gate 应覆盖 nil input、empty input 与真实 DB roundtrip 的差异。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 首要后续动作：更新 `scenario-create` 或场景 README 的 wrapper 示例，移除 `status` 变量，并补充 runner-owned server lifecycle 的写法。
- 次要后续动作：抽象本次 P0.099 的 legacy-negative scoped scan 经验，沉淀为 scenario verify 模板要求，降低后续旧口径负向搜索的误报成本。
- 可延后处理：把 PostgreSQL array nil/empty regression gate 推广到 backend-practice 后续 owner plan；本次已由 BUG-0103 和回归测试覆盖当前风险点。
