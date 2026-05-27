# E2E Scenarios P0 History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-26

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-27 | 1.5 | 将 `E2E.P0.100` 从独立 `manual-uat` companion 迁回标准 `e2e` 场景框架：AI Agent 先运行环境 preflight 与四段脚本，缺真实凭证/浏览器证据时输出 `MANUAL_REQUIRED`，人工或浏览器 Agent 在同一场景输出目录补证。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-26 | 1.4 | 对齐 local-dev-stack Mailpit revision：manual UAT 账号入口改为 synthetic 邮箱 + Mailpit magic-link，删除直接 session bootstrap 口径；继续保留 `test/scenarios` 只允许 shell/Python、不得新增 `backend/cmd` / Go helper 的边界。 | 002-manual-uat-real-provider-full-funnel + local-dev-stack/001 |
| 2026-05-26 | 1.3 | 过渡修正 002 manual UAT 账号边界：当时本地栈尚未包含 Mailpit / MailHog / SMTP mailbox，要求账号/session 辅助不得进入正式 `backend/cmd` / Go helper；该口径已被 v1.4 Mailpit magic-link 入口替代。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-26 | 1.2 | 新增真实 provider manual UAT 验收层：锁定 D-8~D-10，明确 002 不改写 001 的 stub-AI 自动化边界；新增 C-9~C-13，要求账号/session bootstrap、真实前后端、真实 AI provider、无 mock/stub 冒充、脱敏证据与完整材料包。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-24 | 1.1 | L1 plan-review 修订：校正 P0 场景实施前基线为 87 条切片场景（最高编号 `E2E.P0.097`），将 operation matrix 口径统一为 9 行（8 个主链必经 operation + `getJob` 备选轮询 / handler gate），明确 Playwright 全栈必须用 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL` 指向真后端，并把 legacy-negative 加固为 route-aware 旧 route / 独立 voice / `mode=debrief` 反查且避免误伤合法 `createPracticePlan` / `resumeAssetId`。 | 001-full-funnel-happy-journey |
| 2026-05-24 | 1.0 | 初始创建：定义 P0 完整漏斗端到端 journey owner subject；锁定 D-1~D-7（真后端全栈 + stub AI + happy 主干 + 两种 driver + 接续编号）；派生 `001-full-funnel-happy-journey`（`E2E.P0.098` API-level + `E2E.P0.099` Playwright 全栈） | 001-full-funnel-happy-journey |
