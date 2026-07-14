# Frontend Resume Workshop Create Flow

> **版本**: 1.18
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本计划承接当前 Resume Workshop 创建路径：

- `resume_versions?flow=create` 渲染 `ResumeCreateFlow`。
- 创建入口只提供 upload / paste 两种输入。
- Upload 路径只接受 PDF / Markdown / TXT，先执行默认 2MiB 文件大小和扩展名校验，再通过 `createUploadPresign`、browser PUT 和 `registerResume` 完成注册；DOCX 不属于当前 Resume 上传支持范围。
- Paste 路径通过 `registerResume` 完成注册，但只提交中性来源标题；用户可见简历名称等待 backend parse 的 LLM-derived `displayName`，不得把原始文本第一行作为最终或列表名称。
- 注册成功后导航到 `resume_versions?resumeId=<id>`，由详情 route 展示解析等待态，直到 parse 成功后按来源格式展示 PDF 页面栈或 Markdown 只读详情，或失败后展示失败态。
- 创建流不渲染右侧“会保存什么 / 接下来”说明 rail、预览确认页或确认保存页；不在 create-flow 中轮询 `getResume` 或调用 `updateResume`。
- Home “1 分钟创建” CTA 进入当前 create flow。

本计划不实现 backend upload/resume handler、OpenAPI 契约、Resume Rewrites/Edit、PDF 导出或 Practice handoff。

## 2 背景

当前产品已采用 flat Resume IA：简历是平铺资产，创建流程只需要 upload / paste 输入和注册成功后的直接详情跳转。正式前端必须源级复刻 `ui-design/src/screen-resume-workshop.jsx` 的当前 CreateFlow 构图，同时通过 generated client 与 backend-resume / backend-upload 合同集成。

当前实现事实：

- `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx` 将 `flow=create` dispatch 到 `ResumeCreateFlow`。
- `frontend/src/app/screens/resume-workshop/create/` 包含 create-flow 容器、tabs、upload/paste hooks、2MiB upload validation 和 out-of-scope parsing/preview-confirm negative tests。
- P0.081 / P0.083 场景资产覆盖直接详情跳转、CTA handoff 和隐私边界；P0.082 现在只验证直接详情跳转和 out-of-scope 解析/预览面不存在，因为创建流不再暴露解析进度页。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `contract`
- **TDD 策略**: 当前实现已完成。修改 create-flow 或 Home 入口简历选择逻辑时，先更新 Vitest component / hook / adapter tests，再改组件或 generated-client adapter。
- **BDD 策略**: 适用。当前 BDD gate 为 E2E.P0.081、E2E.P0.083、E2E.P0.084；E2E.P0.082 是直接详情跳转与 out-of-scope 解析/预览面缺席 gate，不作为解析失败 UI 正向 gate。
- **替代验证 gate**:
  - `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create`
  - `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts`
  - `corepack pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeResumeSelection.test.tsx src/app/screens/parse/ParseResumeBinding.test.tsx`
  - browser screenshot proof for the Home existing resume picker showing selectable options from `listResumes`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-resume-workshop/plans/002-create-flow/context.yaml --target frontend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`

## 4 当前交付内容

### Phase 1: Create Flow Shell

`ResumeCreateFlow` owns the input state only. The input state shows upload and paste tabs, preserves route params through auth pending action without raw content, and keeps TopBar / route highlighting stable.

### Phase 2: Upload / Paste Registration

- Upload tab validates file extension (`.pdf` / `.md` / `.markdown` / `.txt`) and default 2MiB size, requests a presigned upload, uploads the file through browser `fetch`, then calls `registerResume`; `.docx` is rejected before presign.
- Paste tab validates non-empty text, sends a neutral source title, and calls `registerResume`; visible naming is owned by backend parse after LLM structured output.
- Side-effect calls use `Idempotency-Key`; `getResume` polling does not.
- Request headers carry language where required.

### Phase 3: Waiting/detail handoff

On successful `registerResume`, upload and paste paths call `navigate({ name: "resume_versions", params: { resumeId } })`. The detail route owns parser waiting / ready / failed states; the create flow does not poll `getResume` and does not show a confirm/save page.

### Phase 4: Parser / preview confirm absence

`ResumeParseFlow`, `ParsingStage`, `PreviewStage`, `ResumePreviewConfirm`, `useResumeParsingPolling`, `useResumeSave` create-flow usage and `mapParsedSummaryToStructuredProfileDraft` are out-of-scope surfaces. Tests and source negative scans must fail if they are imported by `ResumeCreateFlow` or rendered by the create route.

### Phase 5: CTA / i18n / a11y / Privacy

Home CTA paths enter `resume_versions?flow=create`. CreateFlow keeps current i18n keys, tablist semantics, focus behavior, and privacy redlines. The negative tests prevent out-of-scope CreateFlow inputs or prototype runtime imports from returning.

### Phase 6: Create page simplification

`ResumeCreateFlow` removes the right-side “会保存什么 / 接下来” sidebar from both static UI truth source and formal frontend implementation. The input card becomes the only main content surface.

### Phase 7: Resume upload source format support

Resume upload keeps the existing name generation and route handoff behavior. It only narrows the upload whitelist to PDF / Markdown / TXT, rejects DOCX before presign/register, and leaves renderer selection to the detail route based on the registered source format.

### Historical Phase 8: Home existing resume selection regression

Phase 8 originally inferred selectability from the then-full `listResumes` item (`parsedTextSnapshot` / `originalText` / structured profile). That list shape is superseded by active [001 Phase 19](../001-listing-routing-and-detail-readonly/plan.md): Home consumes closed `ResumeSummary` and uses only `parseStatus === ready || hasReadableContent`; Parse/Workspace detail does not call `listResumes`. The checked Phase 8 evidence below remains historical and is not a current contract gate.

### Phase 9: Zero-reference stage type removal

`ResumeCreateFlow` keeps the real `data-stage="input"` DOM contract but removes the exported `CreateStage` alias because no production or test consumer uses it. A source negative gate prevents the standalone declaration from returning; focused create-flow tests and typecheck preserve current behavior.

### Phase 10: Prototype create-flow call-surface pruning

The static `ResumeCreateFlow` uses `onBack` to return to the flat list and `onCreateResume` to open the created Resume detail. Remove its unread `nav` parameter and the matching `ResumeWorkshopScreen` child argument. Keep upload/paste input state and both callbacks unchanged; do not add a compatibility parameter or wrapper. The create-to-detail route handoff must preserve the locally created asset, explicitly exit create mode, and render the waiting/ready detail instead of remounting the Resume Workshop owner.

### Phase 11: Zero-consumer ghost CTA CSS pruning

删除 `screens.css` 中没有正式 CreateFlow DOM、静态原型或场景消费者的 `ei-resume-create-cta-ghost` base/variant/disabled branches；保留 upload/paste 当前使用的 accent CTA 及其 disabled state。BDD 不适用，因为 ghost CTA 不可达；替代 gate 为 create-flow source RED/GREEN、class inventory、focused CreateFlow/P0.081、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。

### Phase 12: Accent CTA rule consolidation

ghost variant 删除后，`ei-resume-create-cta-accent` 不再需要“共享基础规则 + 独立颜色规则”的两段声明。将 layout、typography、interaction、accent colors 与 border 合并为一个规则，disabled state 保持独立；最终 computed values 与 upload/paste DOM 不变。BDD 不适用；替代 gate 为 source RED/GREEN、focused CreateFlow、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。

## 5 验收标准

| ID | 场景 | Given | When | Then | 证据 |
|----|------|-------|------|------|------|
| C-1 | Create route | Authenticated user opens `resume_versions?flow=create` | App renders route | `ResumeCreateFlow` appears with upload tab active by default | focused Vitest |
| C-2 | Upload path | Valid PDF / Markdown / TXT file selected | Submit | DOCX and unsupported extensions are rejected before presign; 2MiB file limit is enforced before presign; valid file presign + PUT + register flow runs with IK and language headers | hook / component tests |
| C-3 | Paste path | Non-empty text | Submit | `registerResume` receives paste payload with a neutral source title, raw text is not used as a visible name, and direct detail navigation follows | hook / component tests |
| C-4 | Register recovery | Upload/register fails | User retries or returns | Input stays local and content does not leak | upload/paste tests |
| C-5 | Out-of-scope surfaces absent | Register succeeds | Route updates | Sidebar, preview confirm and create-flow `updateResume` save path do not render or run; waiting state and source-format renderer belong to detail route | negative tests |
| C-6 | CTA handoff | Home create CTA | Click | Route lands on current CreateFlow without raw data in pending action | integration tests |
| C-7 | BDD gates | P0.081 / P0.083 assets plus P0.082 parser/preview absence | Scenario verify | Direct-to-detail main path, out-of-scope parser/confirm absence and CTA handoff are covered | BDD docs + scenario scripts |
| C-8 | Home existing resume picker | `listResumes` returns closed `ResumeSummary` items with `parseStatus` / `hasReadableContent` | Home renders JD quick-start | Native select is enabled for `ready || hasReadableContent`, empty state is absent, and selected `resumeId` is carried only in the import body；full detail fields are not required | Active 001 Phase 19 + HomeResumeSelection Vitest |
| C-9 | Zero-reference type cleanup | `CreateStage` has no consumer | Source gate and create-flow regressions run | The alias is absent while `data-stage="input"` and create behavior remain unchanged | source negative + focused Vitest + typecheck |
| C-10 | Prototype call surface | Static create flow receives its owner callbacks | User switches mode, returns or creates a Resume | Only `onBack` / `onCreateResume` own child transitions; no unread `nav` prop remains, and the parent preserves the created asset through waiting/ready detail | UI contract + AST inventory + browser smoke |

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Generated client drift | Keep fixture parity and real-mode frontend owner tests before changing hook payloads |
| CreateFlow privacy regression | Preserve URL / pendingAction / storage / console tests for raw resume content |
| UI parity drift | Keep CreateFlow DOM anchors, tab roles and screenshot smoke coverage aligned with `ui-design/` |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.18 | Mark the old full-Resume Home selection inference as historical; current selection consumes ResumeSummary parseStatus/hasReadableContent under active plan 001 Phase 19. |
| 2026-07-10 | 1.17 | Consolidate the accent CTA declarations into one equivalent rule. |
| 2026-07-10 | 1.16 | Delete the zero-consumer CreateFlow ghost CTA CSS branches. |
| 2026-07-10 | 1.15 | Remove the unread ResumeCreateFlow navigation prop and caller argument; preserve the created detail handoff documented by BUG-0154. |
| 2026-07-10 | 1.14 | Remove the zero-reference CreateStage type while preserving the input-stage DOM contract. |
| 2026-07-10 | 1.13 | 将 create-flow parser / preview-confirm 负向 gate 表述统一为 out-of-scope 口径；行为不变。 |
| 2026-07-10 | 1.12 | 收敛 P0.083 handoff gate 到当前 Home CTA、auth pendingAction 与 direct-detail 验证。 |
| 2026-07-10 | 1.11 | 将 parser / preview-confirm 负向面表述统一为 out-of-scope wording；行为不变。 |
| 2026-07-10 | 1.10 | 重命名 P0.082 场景为直接详情跳转与 out-of-scope 解析/预览面缺席 gate，删除 active plan 中的范围外生命周期标签口径。 |
| 2026-07-08 | 1.9 | 修复首页已有简历下拉回归：有可读简历时不得显示空态或禁用选择。 |
| 2026-07-08 | 1.8 | 对齐详情 route 当前 PDF 页面栈合同，创建流仍只负责注册后跳转。 |
| 2026-07-07 | 1.5 | 修订未闭环命名回归：paste 创建不再提交原文首行作为标题，列表/详情名称等待 backend LLM-derived displayName。 |
| 2026-07-07 | 1.6 | 本轮优化：upload 默认 2MiB 校验、注册后交给详情等待态、删除右侧冗余说明 rail。 |
| 2026-07-07 | 1.7 | 本轮讨论决策：Resume upload 移出 DOCX 当前支持范围，仅支持 PDF / Markdown / TXT；详情渲染由来源格式自适应，create-flow 交互和名称生成逻辑保持不变。 |
