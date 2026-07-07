# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.21 | active | 2026-07-07 | 三入口信息架构、TopBar、用户菜单、核心 JD / 简历 -> 模拟面试 -> 报告 -> 复练 / 下一轮闭环、首页新建规划把上传 JD / URL source actions 整合进 JD 输入卡底部，简历下拉框与创建入口同排，简历详情只读原始正文，非当前 route 输入归一和当前边界 |
| [目标用户流程](./user-flow.md) | 2.17 | active | 2026-07-07 | 首页启动、解析确认、模拟面试回访、面试与报告、简历创建后直接打开详情、LLM-derived displayName、禁止 raw 第一行/文件名命名、只读原始简历详情、认证设置和非当前流程边界 |
| [目标模块地图](./module-map.md) | 2.10 | active | 2026-07-07 | 三入口模块、会话级页面、简历 LLM-derived displayName、禁止 raw 第一行/文件名命名、设置认证、非当前 route 输入归一和当前目标数据依赖 |
| [认证与默认入口](./auth-and-entry.md) | 1.20 | active | 2026-07-07 | 默认进首页后的邮箱验证码登录、首次账号资料补全、pending action 接续、三入口用户菜单和非当前 `debrief` / `profile` 负向 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.18 | active | 2026-07-07 | 模拟面试一级模块作为回访枢纽：JD、简历、轮次、InterviewContext、公司情报嵌入卡片、当前规划记录与立即面试入口；Home 最近模拟面试更多入口进入本页；首次导入主路径在 Home 输入卡内选择 JD source，再下拉选择已有简历后进入 parse 核对 |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.16 | active | 2026-07-07 | 完整模拟面试、统一文本/语音外层骨架、语音面试 practice 显式参数入口、带提示 / 严格模拟、固定结束生成报告、语音表达 Surface、语音转文字、面试官角色、复盘非当前边界和报告 CTA 直接进入面试 session 的仪表盘闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.12 | active | 2026-07-06 | session-scoped 报告、Summary Cards、二级详情、失败/缺 session 状态、Header 唯一一对复练当前轮与进入下一轮 CTA（D-19）、题目回顾加入本轮复练为标记动作 |
| [简历一级模块](./resume-module.md) | 2.7 | active | 2026-07-07 | 平铺简历列表、上传/粘贴创建后直接打开详情、LLM-derived displayName、禁止 raw 第一行/文件名命名、上传文件正文提取、只读原始正文、无导出/复制/编辑/改写/原件弹层和非当前入口边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 2.9 | active | 2026-07-07 | 多 JD、平铺简历资产、首页新建模拟面试规划在同一 JD 输入卡内整合粘贴 / 上传 / URL source actions、下拉选择已有 ready 简历、创建入口同排、最近模拟面试 3 条 + 更多、`resumeId` 绑定关系、LLM-derived displayName、只读原始详情正文和模拟面试规划绑定关系（D-17/D-20） |
| [首次无简历引导](./resume-onboarding.md) | 1.10 | active | 2026-07-07 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和非当前 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。前端视觉实现只以本目录文档和 `ui-design/` 静态原型为真理源；外部品牌设计系统不是项目 UI 参考。
