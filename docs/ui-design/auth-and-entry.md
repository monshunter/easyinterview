# 认证与默认入口

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-14

## 1 文档目的

本文档定义当前默认入口、认证触发点、pendingAction 接续和用户菜单边界。当前用户菜单只包含设置与隐私、退出登录和未登录入口。

## 2 已确认决策

1. App 默认进入 `home`。
2. 登录方式只有邮箱验证码。
3. 登录是操作级拦截，不是默认前置页。
4. 登录成功后若 `profileCompletionRequired=true`，进入 `auth_profile_setup`；完成账号资料补全后接续 pendingAction。
5. `auth_profile_setup` 是账号资料补全，不是用户画像。
6. 已登录用户菜单只显示 `设置与隐私` 和 `退出登录`。
7. `用户画像` 不是用户菜单入口。
8. `复盘` 不是业务入口或登录触发点。
9. `/reports?targetJobId=...` 是受保护的规划上下文页面；鉴权接续只保留合法 `targetJobId`，但它不进入 TopBar 一级导航。

## 3 默认入口

```text
打开 App
└─ home
   ├─ 在唯一文本框粘贴 JD
   ├─ 最近模拟面试
   └─ 创建简历入口
```

## 4 用户菜单

```text
未登录
└─ 登录

已登录
├─ 设置与隐私 -> settings
└─ 退出登录 -> auth_logout
```

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
| 立即面试 | `practice` | 从 parse 或 workspace 发起 |
| 打开当前规划报告 | `reports` | 从规划详情内容区入口或受保护深链发起；只携带 `targetJobId` |
| 复练当前轮 | `practice` | 从 report header 发起 |
| 进入下一轮 | `practice` | 从 report header 发起 |
| 保存简历 | `resume_versions` | 创建或编辑简历 |
| 打开设置 | `settings` | 账号和隐私设置 |

不得创建 pendingAction 到 `debrief`、`debrief_full` 或 `profile`。

## 7 范围外入口归一

| 输入 | 处理 |
|------|------|
| `debrief` / `debrief_full` | 归一到 `home` |
| `profile` | 归一到 `home` |
| `auth_register` | 归一到 `auth_login` |
| `auth_reset` | 归一到 `auth_login` |

## 8 后续实现输入

1. TopBar 用户菜单不得包含 `用户画像`。
2. 未登录保护路由不得把 `debrief` 或 `profile` 当业务目标。
3. 正式前端、URL fallback 和 scenario 都必须覆盖范围外入口负向。
4. 设置页只承载账号、界面偏好和隐私数据控制。
5. `import_jd` pendingAction 只接续粘贴文本入口；路由只携带 opaque pending id 与业务 ID，不携带 JD 原文或导入类型。
6. `reports` pendingAction 只允许 `targetJobId`；`section`、report/status/round、原文和其他业务状态必须剔除。登录恢复后仍由受保护 API 校验规划归属。
