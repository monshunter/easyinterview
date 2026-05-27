# Data Index: E2E.P0.100 Real Provider Full Funnel Hybrid

> 这些是面向人工验收者的合成素材，人物 / 公司 / 经历均为虚构，**不含真实 PII**，可安全粘贴进 UI。
> Owner: `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel`

## 1 文件清单

| 文件 | 角色 | 语言 | 主要用途 |
|------|------|------|----------|
| [`account.md`](./account.md) | UAT account | N/A | Mailpit email-code 登录、cookie 检查与 cleanup 说明 |
| [`jd-backend-engineer.zh.md`](./jd-backend-engineer.zh.md) | 资深后端工程师 | 中文 | Home「粘贴 JD」导入 → 触发解析 |
| [`jd-backend-engineer.en.md`](./jd-backend-engineer.en.md) | 资深后端工程师 | English | Home「粘贴 JD」导入 → 触发解析 |
| [`resume-backend-engineer.zh.md`](./resume-backend-engineer.zh.md) | 资深后端工程师 | 中文 | 简历创建流程「粘贴」导入 / seeded 简历内容参照 |
| [`resume-backend-engineer.en.md`](./resume-backend-engineer.en.md) | 资深后端工程师 | English | 简历创建流程「粘贴」导入 / seeded 简历内容参照 |
| [`answer-sample-backend-engineer.zh.md`](./answer-sample-backend-engineer.zh.md) | 资深后端工程师 | 中文 | Practice 作答样例 |
| [`answer-sample-backend-engineer.en.md`](./answer-sample-backend-engineer.en.md) | Senior Backend Engineer | English | Practice answer sample |
| [`expected-observations.md`](./expected-observations.md) | UAT reviewer prompts | N/A | Parse / Workspace / Practice / Report / Next Round 期望观察点 |

JD 与简历角色匹配，便于走查时 practice / report 产出贴近真实匹配场景。

## 2 在漏斗中怎么用

| 漏斗步骤 | 用哪个材料 | 怎么用 |
|----------|------------|--------|
| 登录态 | `account.md` | 输入 synthetic 邮箱，打开 Mailpit 本地邮件，完成 email-code 登录 |
| Home → 导入 JD | `jd-backend-engineer.<lang>.md` 正文（`---` 分隔线以下整段） | 选「粘贴 JD」入口，整段粘贴，提交导入 |
| Parse → 确认 | —（系统解析上一步 JD） | 等待解析 ready，检查结构化字段，Confirm 进 Workspace |
| 简历 | `resume-backend-engineer.<lang>.md` 正文 | 通过真实 UI 创建 / 注册简历，或在登录后按 runbook 准备 ready resume |
| Practice → 作答 | `answer-sample-backend-engineer.<lang>.md` 或自由作答 | 结合简历经历回答面试问题，观察真实 provider follow-up |
| 全漏斗观察 | `expected-observations.md` | 作为人工 reviewer 的观察提示，不作为 mock response 或精确断言 |

## 3 语言走查约定

- UI 语言：用 TopBar 的 globe 下拉切换 `中文` / `English`，对应使用 zh / en 材料。
- 业务语言字段独立于 UI 语言；中英材料分别用于核对对应语言下的解析与练习产出。
- `targetLanguage` / practice language 是业务字段，不随 `Accept-Language` 改变；走查时如需切换，按 UI 内业务语言选择项操作。

## 4 注意

- 这些材料只是**输入样本**，不是 mock response、fixture response 或断言基线。验收判定以 [`../checklist.md`](../checklist.md) 的可观察结果为准。
- `E2E.P0.100` 是 `hybrid` 场景：AI Agent 先用脚本完成环境与材料 preflight，人工或浏览器 Agent 再用本目录材料完成真实 provider 观察和脱敏证据。
