# 认证与默认入口

> **版本**: 1.27
> **状态**: active
> **更新日期**: 2026-07-19

## 1 文档目的

本文档定义当前默认入口、认证触发点、pendingAction 接续和账号设置边界。未登录时显示登录入口；已登录时不再显示账号 chip 或用户菜单，只显示一个设置齿轮。

## 2 已确认决策

1. App 默认进入 `home`。
2. 登录方式只有邮箱验证码。
3. 登录是操作级拦截，不是默认前置页。
4. 登录成功后若 `profileCompletionRequired=true`，进入 `auth_profile_setup`；完成账号资料补全后接续 pendingAction。
5. `auth_profile_setup` 是账号资料补全，不是用户画像。
6. 已登录 TopBar 账号区只显示一个带可访问名称的设置齿轮，直接进入 `settings`；退出登录迁入设置页。
7. `用户画像` 不是 TopBar 或设置页入口。
8. `复盘` 不是业务入口或登录触发点。
9. `/reports?targetJobId=...` 是受保护的规划上下文页面；鉴权接续只保留合法 `targetJobId`，但它不进入 TopBar 一级导航。
10. `/workspace?targetJobId=...` 是受保护的只读规划详情；`/workspace` 是列表。`/parse?targetJobId=...` 只允许接续已创建的新导入 queued/processing 命令，ready 后 replace 到 Workspace 详情。
11. 受保护 route 等待或无法确认登录状态时继续使用统一 auth gate；其 eyebrow、标题和说明全部跟随当前语言，不得在中文模式显示硬编码英文。

## 3 默认入口

```text
打开 App
└─ home
   ├─ 在唯一文本框粘贴 JD
   ├─ 最近模拟面试
   └─ 创建简历入口
```

## 4 TopBar 账号区

```text
未登录
└─ 登录

已登录
└─ 设置齿轮 -> settings
```

设置入口必须使用可明确识别的齿轮图标和标准 button，具备 `aria-label="设置 / Settings"`、清晰 focus ring 和至少 40×40px 点击区域。desktop 与 mobile 都不得再出现头像、姓名、caret、backdrop 或账号 dropdown。

## 5 认证页面

```text
auth_login
  -> auth_verify
     -> profileCompletionRequired?
        ├─ true -> auth_profile_setup -> continue pendingAction
        └─ false -> continue pendingAction

auth_logout
  -> home
```

范围外 `auth_register` 和 `auth_reset` 归一到 `auth_login`。

## 6 Pending Action

可接续的 pendingAction 只覆盖当前核心业务动作：

| 动作 | 目标 route | 说明 |
|------|------------|------|
| 打开当前面试规划 | `workspace` | 从 ready Home/Workspace 卡片、Reports Back 或 Practice terminal recovery 发起；只携带 `targetJobId` |
| 查看新导入解析进度 | `parse` | 只接续已创建 TargetJob 的 queued/processing 命令；只携带 `targetJobId`，ready 后 replace 到 Workspace 详情 |
| 立即面试 | `practice` | 从 Workspace 详情或列表 quick-start 发起 |
| 打开当前规划报告 | `reports` | 从规划详情内容区入口或受保护深链发起；只携带 `targetJobId` |
| 复练当前轮 | `practice` | 从 report header 发起 |
| 进入下一轮 | `practice` | 从 report header 发起 |
| 保存简历 | `resume_versions` | 创建或编辑简历 |
| 打开设置 | `settings` | 已登录用户通过设置齿轮或受保护深链进入账号和隐私设置；route 不携带用户字段 |

不得创建 pendingAction 到 `debrief`、`debrief_full` 或 `profile`。

## 7 范围外入口归一

| 输入 | 处理 |
|------|------|
| `debrief` / `debrief_full` | 归一到 `home` |
| `profile` | 归一到 `home` |
| `auth_register` | 归一到 `auth_login` |
| `auth_reset` | 归一到 `auth_login` |

## 8 后续实现输入

1. TopBar 已登录账号区只能有设置齿轮；不得包含头像/姓名 chip、dropdown、退出按钮或 `用户画像`。
2. 未登录保护路由不得把 `debrief` 或 `profile` 当业务目标。
3. 正式前端、URL fallback 和 scenario 都必须覆盖范围外入口负向。
4. 设置页为无 tab 单页，承载账号级主题偏好、真实账号字段、退出登录和当前可执行的隐私状态/动作；语言与暗色继续由 TopBar 控制，字体不再可配置。
5. `import_jd` pendingAction 只接续粘贴文本入口；路由只携带 opaque pending id 与业务 ID，不携带 JD 原文或导入类型。
6. `reports` pendingAction 只允许 `targetJobId`；`section`、report/status/round、原文和其他业务状态必须剔除。登录恢复后仍由受保护 API 校验规划归属。
7. `workspace` / `parse` pendingAction 只允许合法 `targetJobId`；不得接续 `planId`、`resumeId`、analysis status、动画步数或 auto-start 等业务事实。恢复后分别由受保护 API 判定只读详情或命令进度，ready Parse 必须 replace 到 Workspace detail。

## 9 设置页单页合同

```text
Settings
├─ Appearance
│  ├─ Theme: Ocean / Plum / Custom accent
│  ├─ Custom accent: hue + saturation（本地预览）
│  └─ Save: 单次 updateMe，成功响应直接刷新 runtime
├─ Account
│  ├─ Display name: /me.displayName（只读）
│  ├─ Login email: /me.email（完整账号邮箱，只读）
│  └─ Sign out -> auth_logout
└─ Privacy & data
   ├─ Export data: 暂不可用（不可触发伪成功）
   └─ Delete account -> destructive confirmation
```

- `SettingsScreen` 复用应用启动或认证恢复时已取得的 authenticated user，不为页面挂载或 route 切换重复调用 `getMe`；loading/error/unauthenticated 仍由统一 route guard 处理。
- 主题草稿在设置页本地预览；拖动 hue/chroma 不发网络请求。点击保存只发一次 `updateMe`，成功响应返回完整用户上下文并直接更新 runtime/display context，不追加 `getMe`。失败保留草稿和错误供重试，不覆盖最近一次服务端确认值；离开未保存页面恢复确认值。重新登录或其他平台启动后由首个 `getMe` 恢复同一账号主题。
- 不渲染 tab rail、手机号、界面语言行、时区、登录与安全、字体预设、产品信息、数据留存开关、数据概览、删除单次会话或删除所有练习数据等没有当前数据源/operation 的静态条目。
- 数据导出沿用 P0 `501 PRIVACY_EXPORT_NOT_AVAILABLE` 契约，默认显示禁用态和可读的“暂不可用”原因；不得展示为可触发按钮，也不得把静态文案或未发请求状态显示成已受理。
- 删除账号点击后打开二次确认；确认前无副作用。对话框使用 destructive title/description、初始焦点、焦点约束、Escape/取消和关闭后焦点归还；pending 时禁止关闭和重复提交。失败保留对话框、显示可恢复错误并允许重试；`401` 进入统一认证重探测，不继续展示可重试删除；成功收到 `202` 后调用现有 `refreshAuth()` 重探测 `/me`（预期 401），提交 unauthenticated 状态并 replace Home。一次确认生命周期内的网络重试复用同一 idempotency key，不新增第二套清 session 方法；重探测若遇网络/服务错误，诚实保留 auth error，不伪装成已确认退出。
- 退出登录继续进入既有 `auth_logout` 确认页，不在设置页复制第二套 logout mutation 或确认文案。

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.27 | 设置页更名为“设置”，新增账号级外观主题保存；锁定 bootstrap 单次读取、route 零重复读取和单次保存更新合同。 |
| 2026-07-16 | 1.26 | 明确 auth loading/error route gate 属于本地化 shell UI，中文模式不得出现英文硬编码。 |
| 2026-07-15 | 1.25 | 采用设置简化方案 A：已登录 TopBar 仅保留设置齿轮，设置页改为无 tab 的真实账号/隐私单页，并明确退出、导出不可用与账号删除状态机。 |
