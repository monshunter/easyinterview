# Resume Source Format Preview 交付复盘报告

> **日期**: 2026-07-07
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖简历来源格式分流：粘贴文本、Markdown 和 TXT 继续使用 Markdown renderer；upload-backed PDF 在详情页使用原件 PDF preview；DOCX 从简历上传白名单、presign/register、parse 路径、UI 文案和静态原型中退役；同时新增 `getResumeSource` 只读接口返回用户作用域内的 PDF 原件。

成功证据：

- Frontend focused gate: `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create/UploadTab.test.tsx src/app/screens/resume-workshop/adapters/resume.test.ts src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/api/devMockClient.test.ts src/api/mockTransport.test.ts`，5 个测试文件 30 项通过。
- Frontend type gate: `corepack pnpm --filter @easyinterview/frontend typecheck` 通过。
- Backend focused gate: `cd backend && go test ./internal/resume/... ./internal/upload/... ./cmd/api -count=1` 通过。
- OpenAPI gate: `make lint-openapi validate-fixtures openapi-diff` 通过，inventory / baseline / current 均为 36 operations，`getResumeSource` fixture 覆盖成功。
- Browser parity gate: `corepack pnpm --filter @easyinterview/frontend test:pixel-parity -- tests/pixel-parity/resume-workshop.spec.ts`，desktop/mobile 共 8 项通过，覆盖 Markdown 详情与 PDF source preview object。
- Local environment gate: `test/scenarios/env-redeploy.sh all` 与 `test/scenarios/env-verify.sh` 通过，本地前端 `http://127.0.0.1:5173/`、后端 `http://127.0.0.1:8080/api/v1` 可用。
- 截图验收：`.test-output/local-dev/resume-detail-markdown.png`、`.test-output/local-dev/resume-detail-pdf-source.png`、`.test-output/local-dev/resume-create-upload-formats.png` 均已生成；PDF 详情截图使用 headed Chromium，因为 headless Chromium 不启用内置 PDF viewer。
- 文档与格式 gate: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`make docs-check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- PDF preview 的自动截图不能依赖 headless Chromium。
  - **证据**：headless `<object>` 命中 PDF 分支但显示 fallback；headless `<iframe>` PDF 探针为空白；headed Chromium 探针和实际详情页能显示 PDF viewer。
  - **影响**：若只看 headless 截图，会误判 PDF renderer 没生效；最终需要 DOM/URL/layout 自动断言加 headed screenshot 双轨验收。
- 格式分流横跨多 owner 资产。
  - **证据**：本次同时更新 `frontend/` renderer 与上传入口、`backend/` presign/register/parse/source handler、`openapi/` operation/fixture/codegen、`docs/spec/` owner plans、`docs/ui-design/` 与 `ui-design/` 静态原型。
  - **影响**：只改前端详情页会留下后端仍接收 DOCX、fixture 不覆盖 source endpoint、静态原型与正式实现漂移等风险。
- DOCX 退役需要负向 gate。
  - **证据**：实现层需要同时覆盖 accept 文案、MIME/extension 校验、presign 前置校验、parse fallback 拒绝、文档和原型说明。
  - **影响**：如果没有 explicit negative tests，历史 DOCX 支持路径容易以 dead code 或旧文案残留。

## 3 根因归类

- Headless 浏览器能力差异。
  - **类别**：spec-plan
  - **说明**：现有 pixel parity gate 可验证 DOM、style、bounding box 和截图 smoke，但 PDF viewer 属于浏览器插件能力，headless 与 headed 行为不同。
- 来源格式合同没有单点 owner。
  - **类别**：spec-plan
  - **说明**：PDF/Markdown/TXT/DOCX 的合同同时存在于 UI、API、上传、解析和 fixture 层，需要在 owner plan 中显式列出 operation、consumer、handler、persistence、fixture 与 screenshot gate。
- 退役格式缺少固定搜索词。
  - **类别**：spec-plan
  - **说明**：DOCX 删除不能只依靠正向白名单测试，需要固定 `.docx`、DOCX MIME、`word/document.xml`、docx parser 等 zero-support 搜索项。

## 4 对流程资产的改进建议

- 在 `frontend-resume-workshop` 的 UI parity gate 中补充 PDF viewer 截图策略。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：PDF source preview 自动 gate 用 Playwright DOM/URL/layout 断言；真实视觉截图使用 headed Chromium，并在 checklist 记录 headless Chromium 不渲染内置 PDF viewer。
- 在 `backend-resume` / `openapi-v1-contract` plan 模板或当前 plan gate 中保留来源格式 operation matrix。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：每次变更上传格式时同时列出 presign、register、parse、source preview、fixture、generated client/server 和 UI consumer。
- 给格式退役类变更增加 negative search gate。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：除测试拒绝外，记录运行时代码、文案、静态原型、OpenAPI、fixtures 中允许保留和必须为零的旧格式命中。

## 5 建议优先级与后续动作

P1：将 PDF viewer headed screenshot 策略写入 `frontend-resume-workshop` 的后续 parity gate，避免未来复跑时把 headless fallback 当成产品问题。

P2：把来源格式 operation matrix 固化到 `backend-resume` 与 `openapi-v1-contract` 的当前 owner checklist，后续新增或删除格式时能一次性覆盖前后端合同。

P2：为 DOCX 退役追加一次 repo-wide negative search 审计，输出允许保留清单与必须清零清单；若仍有非目标命中，按 owner plan 原地清理。
