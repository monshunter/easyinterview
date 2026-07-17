# Bug 模式库

> 从历史 Bug 中归纳的通用问题模式。Agent 在诊断新问题时应先查阅本文件。

<!-- 模式模板：

## 模式 N：[模式名称]

- **相关 Bug**：BUG-XXXX, BUG-YYYY
- **典型症状**：...
- **检查清单**：
  1. ...
  2. ...

-->

## 模式 1：Cleanup commit 漏同步 test consumer

- **相关 Bug**：BUG-0023, BUG-0129
- **典型症状**：truth source（`shared/*.yaml` / spec / generated 主体）已干净，但单包构建失败，报 `undefined: <RemovedConstant>`；`make codegen-check` 全绿但 `make test` 在某包失败；cleanup commit message 自述同步 owner 文档与 generated artifacts，未提及 test consumer。
- **检查清单**：
  1. 删除 enum value / JobType / capability / feature flag 等 cross-cutting 标识符前后，对仓库做一次 reverse-grep（`*.go` / `*.ts` / `*.tsx` / `*.yaml` / `*.json` / `*.tmpl` / `*.sql`），把命中点全部过一遍而不仅看 truth source。
  2. cleanup 类提交除 `make codegen-check` 之外，必须额外运行触达包的 `go test ./...` 与对应前端 `pnpm --filter ... test`，避免 generated 主体清理但 test 没跟上；若 lint 从 active spec 派生期望，也必须同步对应 spec row 后运行 focused lint。
  3. 把 `internal-only` / `forbidden` / `allowed` 等测试断言数组当作契约 anchor — 一旦上游 enum 集合调整，必须把数组中的对应元素同步删除或追加，禁止保留悬空引用。
  4. 落地新 lint 时优先考虑跨 yaml/code 的双向 enum 一致性（如 `lint-jobs-consumers`）：truth source 删除后，凡引用该常量的测试断言数组必须能被静态识别。

## 模式 2：入口 Skill 分支门禁缺少前置保护或语义命名约束

- **相关 Bug**：BUG-0035, BUG-0080
- **典型症状**：用户报 bug / 回归后，`/change-intake` 或其它入口 skill 先修改 spec / plan / checklist / docs，再进入 `/implement`；随后 `git status` 显示未提交改动落在 `main` 等默认父分支上；下游 `/implement` 的 branch resolution 已经无法防止前置文档改动污染父分支。另一类症状是已经创建了 feature branch，但第一段前缀是 `codex/`、`claude/`、`gemini/` 等工具身份，而不是 `fix/`、`docs/`、`design/`、`spec-design/` 等工作类型或业务域。
- **检查清单**：
  1. 任何可能写文件的入口 skill 在首次 `apply_patch`、formatter、codegen、doc creation、bug/report/journal 写入前，先运行 `git status --short --branch`。
  2. 若当前在默认父分支且工作区干净，先 fast-forward-only 更新父分支，再创建 feature branch；不得在父分支上做 spec / plan / checklist 原地修订。
  3. 若已经在默认父分支产生当前会话改动，先确认父分支与远端同步，再 `git switch -c <feature-branch>` 保留改动并报告恢复动作。
  4. 创建或恢复 feature branch 时，确认第一段前缀表达语义域：代码实施常用 `feat/`、`fix/`、`docs/`、`opt/`，docs/spec 设计派生常用 `design/` 或 `spec-design/`；禁止新建 `codex/`、`claude/`、`gemini/`、`agent/` 等工具名前缀。
  5. 若当前会话刚创建了工具名前缀分支且尚未推送 / 未被外部引用，先重命名为语义前缀再继续写文件；若已推送或可能被协作者依赖，停止并询问用户。
  6. 若 dirty 内容来源不明或可能属于用户，停止并询问用户；禁止擅自 `stash`、`reset`、`checkout` 或把不明改动提交进当前任务。

## 模式 3：Vite dev 默认 mock 路径没有覆盖真实预览语义

- **相关 Bug**：BUG-0036, BUG-0078
- **典型症状**：前端 dev server 运行在 `5173`，页面请求 `/api/v1/...` 时 Network 面板显示目标也是 `localhost:5173`；后端未启动时大量真实 API 报错，已开发页面因 bootstrap data 失败而无法查看；组件测试能过，但真实 `main.tsx` 启动路径失败。另一类症状是页面确实走了 fixture-backed mock，但跨 operation 的 async job / derived handoff 没有状态推进，用户在默认 dev mock 下卡在 loading 或进入错误的 baseline fixture。
- **检查清单**：
  1. 检查 generated client 是否默认使用相对 `/api/v1`，以及 Vite config 是否有显式 `/api` proxy；没有 proxy 时相对 URL 会落到前端 origin。
  2. 检查 `main.tsx` 是否直接 `new EasyInterviewClient()`；正式 app bootstrap 必须通过可测试 factory 选择 dev mock / real backend 模式。
  3. Vite dev 默认应能在 backend absent 时展示 fixture-backed 页面；真实 backend 模式必须显式 opt-in，并指向 backend port 或 `VITE_EI_API_BASE_URL`。
  4. 对包含 `create* -> getJob -> get*` 或 `create* -> start*` 的流程，必须用默认 `createDevMockClient()` 写一条跨 operation regression，证明无显式 `Prefer` 时 fixture state 能推进到用户可见下一步。
  5. Playwright smoke 不应只靠 route mock；至少有一条 dev-preview smoke 断言真实页面加载期间没有意外 `/api/v1` network request，或有等价的 fixture-backed full-flow 测试覆盖正式 bootstrap mock 语义。

## 模式 4：Completed checklist 掩盖未执行的 runner gate

- **相关 Bug**：BUG-0064, BUG-0066, BUG-0068, BUG-0075, BUG-0082, BUG-0087, BUG-0090, BUG-0093, BUG-0100, BUG-0179
- **典型症状**：plan/checklist 标记 `completed`，但 test checklist / BDD checklist 仍有未勾选项；scenario `verify.sh` 只检查 spec 文件存在、历史说明或宽泛 `PASS` 字样；pixel parity / scenario wrapper 被写成 deferred 或外部运行，仍被计入完成证据；completed-plan 压缩删除仍被下游消费的 owner marker，或 verifier 仅凭失败说明中出现 marker 名称就误判通过。
- **检查清单**：
  1. 对 completed plan 先 `rg "\\[ \\]|deferred|pending|no tests|Playwright.*待|pixel parity 待"`，把空勾选、延期口径和 no-op 风险当作 blocking drift。
  2. 直接读取每个 scenario 的 `trigger.sh` / `verify.sh`，确认 `trigger.sh` 真正调用 runner；仅 `grep` 测试名不能作为通过证据，`verify.sh` 必须对命名测试要求 runner 的精确 pass marker（Go 为 `--- PASS: TestName`），并显式拒绝 `--- FAIL:`、package `FAIL`、skip 与 no-tests。
  3. 对 `go test -run` 证据必须加 `go test -list` 或源码反查，确认 focused test 名真实存在；如果测试可以 skip，verify 必须显式拒绝 `--- SKIP` / `testing: warning: no tests to run` / `[no tests to run]`。
  4. Pixel parity gate 必须证明浏览器 runner 执行过；不能只检查 Playwright spec 文件存在或在 README 中写“可手动运行”。
  5. 对 `pnpm` / package script wrapper，必须从 trigger log 反查最终 runner command：如果 package script 本身已经包含 `vitest run`，不得再用 `-- --run ...` 这类可能扩大范围或让 filter 失效的透传形式。
  6. Hash-route pixel parity harness 必须与 routeStore bootstrap 优先级一致；若 hash adapter 要生效，URL path/search 应保持 bare `/`，不能用 `?nonce=...#route=...` 让 canonical search 抢先解析。
  7. 文档收口时把证据 artifact 名称写成当前脚本真实产物，例如 `.test-output/e2e/<scenario>/trigger.log`，避免 checklist 引用不存在的 `*.evidence.log`。
  8. 对 `Ready` / `Verified` 场景先做结构 preflight，确认 `README.md`、`scripts/setup.sh`、`scripts/trigger.sh`、`scripts/verify.sh`、`scripts/cleanup.sh`、`data/seed-input.md`、`data/expected-outcome.md` 全部存在；缺任何一个文件都不能把 runner pass 当成完整 BDD 证据。
  9. 用户可见 route/context handoff 场景必须检查浏览器 URL query、目标 route DOM state 与 exact context key marker；component-level navigation spy 只能作为补充，不能替代 browser gate。
  10. 对强制 preflight / review gate，不要复用会在依赖缺失或配置加载失败时 `t.Skip` 的 live-scenario helper；必需 gate 的 loader / setup helper 应在契约失败时 `t.Fatal` / 非零退出，只有真正可选的 live integration 才允许 skip。
  11. 对 checklist owner marker，使用显式 `marker=<name>` 机器字段并增加负向测试：verified 失败说明、普通 evidence 和 unchecked item 即使提到 marker 名称也不得满足 gate；压缩 completed plan 前反查 marker 消费者并保留仍在用的 handoff。

## 模式 5：Domain service 已实现但 runtime caller 未接入

- **相关 Bug**：BUG-0083, BUG-0084, BUG-0087, BUG-0098, BUG-0105, BUG-0106
- **典型症状**：service 层已有 deletion / generator / outbox / prompt helper，单测也通过，但 `cmd/api` startup path、background runner 或 HTTP handler 没有实际调用；AI prompt contract 有字段名，真实 payload 却是 `{}` / `[]` / 空 RawMessage；error code 常量存在，但响应 envelope 与 retryable 元数据未按 generated contract 返回。
- **检查清单**：
  1. 对每个 cross-owner domain service，从 production caller 反查一次：`main.go` / runtime builder / drainer handler / HTTP route 是否真实注入并调用该 service。
  2. 对 AI generator / search adapter，不只检查 prompt body；在 focused 或 live test 中捕获 `AIClient.Complete` payload，断言业务关键 JSON 字段非空并包含 join key（如 `jobMatchId`）。
  3. 对 privacy delete、profile delete、domain cascade delete，必须通过 runner/handler 层测试证明 async job 调用了所有 domain deleter，并反查目标身份/域数据的关键残留字段；不能只断言 request/job terminal status。
  4. 对账号删除类 cleanup，request/job completed 还必须证明 `users.email` 原值不可查询、`users.deleted_at` 受理期已设置、所有 sessions 已撤销，且 `privacy_requests` tombstone 不会被用户 hard delete 级联删除。
  5. 对 API error response，优先解码 generated `ApiErrorResponse`，并断言 `error.retryable` 来自 shared registry 而不是 HTTP status 推断。
  6. 对 AI output schema 的 `required` 字段，反查生产 prompt input、consumer struct optionality 与 persistence optionality；若 caller 不提供该信息且 consumer 可接受缺失，就不能把字段标成 required。
  7. 对真实 provider UAT，必须捕获 production caller 的 prompt payload 摘要或 focused test，断言关键业务上下文非空（如 rubric dimensions、target/resume/session join keys）；prompt/schema 文件存在不等于运行时 payload 正确。

## 模式 6：Frontend-first handoff 完成后未回填真实 backend gate

- **相关 Bug**：BUG-0085, BUG-0086, BUG-0089
- **典型症状**：frontend plan 在 backend owner 未完成时以 fixture-backed UI variants 交付，后续 backend owner 已落地真实 handler / scenarios，但 completed frontend plan 的 operation matrix、spec Out of Scope、scenario docs 仍写 `not-yet-implemented` / future backend / fixture-first；frontend scenarios 只跑 mock/fixture UI 子用例，未证明 `VITE_EI_API_MODE=real` 下 production generated client 指向真实 backend base URL。
- **检查清单**：
  1. 对 completed frontend-first plan 做 L2 review 时，先反查同 subspec 或相邻 backend owner plan 是否已经完成；不要把完成时的 fixture-first wording 当成当前事实。
  2. 对每个 operationId 建 operation matrix：fixture、frontend consumer、backend route/handler、persistence/AI owner、frontend BDD scenario、backend live/focused scenario，缺一项就不得宣称真实联调闭环。
  3. Frontend scenario trigger 应前置一个 `VITE_EI_API_MODE=real` generated-client gate，断言 base URL、`credentials: "include"`、无 fixture `Prefer` header、side-effect `Idempotency-Key` 与关键 response provenance；然后再跑 fixture-backed UI variants。
  4. Scenario verify 必须检查 real-mode marker 和测试文件名，防止 fixture UI PASS 单独满足完成条件。
  5. 若同一 subspec 已出现一次 handoff drift，立即 sweep sibling/completed plans 的相同模式，避免只修首个被用户指出的 plan。

## 模式 7：Secret lint 扫描范围误纳入 ignored runtime files

- **相关 Bug**：BUG-0094
- **典型症状**：`make lint` 或 `lint-secrets-pattern` 在带本地 `.env` 的开发机失败；失败路径位于 `.gitignore` 排除的 runtime config、缓存或临时证据目录；CI 或干净 checkout 不复现。
- **检查清单**：
  1. 先用 `git ls-files --cached --others --exclude-standard` 确认 secret scanner 的候选集，只允许 tracked files 与未忽略新增文件进入 repo quality gate。
  2. 对外部 scanner 使用临时 mirror 或等价输入文件列表，避免把整棵工作树作为 source。
  3. 修复后在带 ignored `.env` 的本机运行对应 lint，并确认 `.env` 不再被读取；同时保留未忽略文件中的 secret pattern negative gate。
  4. Bug / report / 日志中只能记录脱敏路径与变量名，禁止写入真实 secret 值或可还原片段。

## 模式 8：Schema validation 只校验首个 JSON value

- **相关 Bug**：BUG-0095
- **典型症状**：AI output schema validator 对 `{"field":"valid"} trailing prose` 或两个连续 JSON value 返回通过；观测层没有产生 `AI_OUTPUT_INVALID`，但后续业务 parser 可能才报错。
- **检查清单**：
  1. 对所有 runtime JSON schema / contract validator，确认解析后继续读取输入流并只接受 EOF；不能只调用一次 `Decode` 后立即认为响应是 strict JSON。
  2. focused negative tests 必须覆盖 schema-valid JSON 后追加非空 prose / token，并断言返回契约错误码、validation status、metric / log 失败信号。
  3. 保留既有 invalid JSON、missing required、enum mismatch、array top-level valid 等测试，避免 trailing-token 修复破坏正常 schema 校验。

## 模式 9：Seed migration gate 只验证已存在行

- **相关 Bug**：BUG-0097
- **典型症状**：config truth source、runtime parser、lint contract 已新增 feature_key / enum / baseline asset，但 seed migration 静态测试仍写死历史 row count 或历史 allowlist；hash lint 只检查已经出现在 SQL 里的 row，完整缺失的 feature_key 不报错；DB-backed runtime 首次依赖 baseline seed 时才暴露缺行。
- **检查清单**：
  1. 对 prompt/rubric/config/fixture seed gate，从当前 truth source 反推完整期望集合，再与 migration / fixture / generated artifact 做 exact set compare；不要只断言历史固定数量。
  2. 对 seed row 做 missing / extra / duplicate 三类断言；hash / checksum drift 只能作为附加检查，不能替代存在性检查。
  3. 若 seed 分散在多个 migration，测试应扫描当前 owner 命名规则下的全部相关 migration 文件，而不是只读最早的 baseline migration。
  4. L2 review completed plan 时，凡看到“seed migration hash clean”或“row count clean”，必须追问缺行是否会失败，并至少用一个新 active truth-source 坐标反查 SQL 命中。

## 模式 10：SPA 同路由参数切换复用旧 screen state

- **相关 Bug**：BUG-0101
- **典型症状**：初始进入某个路由、刷新或重新解析测试都通过，但 SPA 内从 `/route?id=A` 切到 `/route?id=B` 时组件没有 remount，旧 entity 的 title/form/edit state 仍显示；network 已请求新 id，却因为本地 `stage` / pending state / hydration effect 条件不满足而无法渲染新实体。
- **检查清单**：
  1. 对以 route param 作为 owner identity 的 screen（如 `targetJobId`、`sessionId`、`resumeVersionId`），在 component test 中用同一个 mounted instance `rerender` 新 params，不能只测 initial mount。
  2. route-param switch regression 必须断言旧 DOM 消失、新 loading/skeleton/error boundary 出现，以及最终 hydrate 新 entity；只断言新请求发出不够。
  3. screen 内 editable fields、temporary toggles、pending ready job、error state、polling timeout 和 in-flight UI stage 必须以 owner identity 变化作为 reset boundary。
  4. 如果 App route table 不给 screen 加 `key`，不要假设 React 会 remount；需要生产代码显式 reset 或在 route composition 层引入明确 key，并用测试锁定。

## 模式 11：退役 schema 列残留在 active SQL 和 mock 列集合

- **相关 Bug**：BUG-0142
- **典型症状**：当前迁移 / DDL 已删除某列或模块，但真实后端接口返回 500，错误链里出现 `pq: column "<column>" does not exist`；sqlmock 单测继续通过，因为 mock rows 和 production SQL 一起保留了旧列；OpenAPI / fixture / codegen gate 全绿但真实 DB-backed 详情、列表、runner 读取失败。
- **检查清单**：
  1. 删除表列、模块 foreign key、profile/resume/session 等 owner 字段后，对 active SQL 做 scoped negative grep：handler、store、runner、generated fixture、migration、shared contract 都要覆盖，但允许历史 Bug/report/work-journal 记录保留脱敏说明。
  2. sqlmock 列集合不能作为 schema truth source；同一 owner 至少保留一条 `-tags=integration` 或等价真实 DB-backed gate，直接插入当前 DDL 允许的最小 row 后调用 production store read/write path。
  3. 对业务失败态对象（如 parse failed / async failed / no child rows），详情接口仍应有 focused 或 integration regression，断言返回可展示状态而不是把 store error 包装成业务操作失败。
  4. 完成退役列修复时，把 plan/checklist gate 写成当前 schema 对齐和 zero active reference 搜索，避免后续只运行 mock-focused tests 又把旧列带回。
  5. 若真实 DB-backed gate 被 owner plan 或 Bug 记录列为强制防线，缺少 `DATABASE_URL`、DB 不可达或 focused test 未执行时必须 fail fast；不得用 `t.Skip` 让 `go test` 整体 PASS，verify 证据也要显式拒绝 `--- SKIP` 和 `no tests to run`。

## 模式 12：持久化终态早于业务输出终态

- **相关 Bug**：BUG-0185
- **典型症状**：AI observability / task-run decorator 已把调用记录为 `success/ok`，但业务 handler 随后因 decoder、normalizer 或领域约束返回 `AI_OUTPUT_INVALID`；async job 失败且没有业务副作用，DB task run 却显示成功。
- **检查清单**：
  1. 明确唯一持久化 writer 的 terminal boundary；删除重复 writer 后，不能假设 provider 成功或通用 schema 通过就等于业务输出已被接受。
  2. 对比 output schema 与下游 decoder 的完整接受集合，尤其检查空数组、空对象、空白字符串、被过滤后为空和别名字段等“结构合法但业务不可用”输入。
  3. 优先把可声明的业务约束前移到共享 output schema，使 decorator 在落 success 前 fail closed；不要让 handler 通过第二个 task-run writer 修正状态。
  4. 回归测试必须同时断言 `ai_task_runs.status/validation_status`、业务 `JobOutcome` 和 ready/outbox/persistence 副作用，避免只测 error code。
  5. 若领域约束无法由现有 schema 表达，先设计一个仍由 observability owner 控制的 finalize seam；不得在业务层临时追加非原子的状态覆盖。
