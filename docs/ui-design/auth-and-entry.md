# 认证与默认入口

> **版本**: 1.18
> **状态**: active
> **更新日期**: 2026-06-29

## 1 文档目的

本文档定义当前默认入口、认证触发点、pendingAction 恢复和用户菜单边界。当前 UI 已按 D-22 删除复盘和用户画像入口。

## 2 已确认决策

1. App 默认进入 `home`。
2. 登录方式只有邮箱验证码。
3. 登录是操作级拦截，不是默认前置页。
4. 登录成功后若 `profileCompletionRequired=true`，进入 `auth_profile_setup`；完成账号资料补全后恢复 pendingAction。
5. `auth_profile_setup` 是账号资料补全，不是用户画像。
6. 已登录用户菜单只显示 `设置与隐私` 和 `退出登录`。
7. `用户画像` 不再是用户菜单入口。
8. `复盘` 不再是业务入口或登录触发点。

## 3 默认入口

```text
打开 App
└─ home
   ├─ JD 粘贴 / 上传 / URL 导入
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
        ├─ true -> auth_profile_setup -> restore pendingAction
        └─ false -> restore pendingAction

auth_logout
  -> home
```

旧 `auth_register` 和 `auth_reset` 归一到 `auth_login`。

## 6 Pending Action

可恢复的 pendingAction 只覆盖当前核心业务动作：

| 动作 | 目标 route | 说明 |
|------|------------|------|
| 立即面试 | `practice` | 从 parse 或 workspace 发起 |
| 复练当前轮 | `practice` | 从 report header 发起 |
| 进入下一轮 | `practice` | 从 report header 发起 |
| 保存简历 | `resume_versions` | 创建或编辑简历 |
| 打开设置 | `settings` | 账号和隐私设置 |

不得创建 pendingAction 到 `debrief`、`debrief_full` 或 `profile`。

## 7 旧入口归一

| 输入 | 处理 |
|------|------|
| `debrief` / `debrief_full` | 归一到 `home` |
| `profile` | 归一到 `home` |
| `auth_register` | 归一到 `auth_login` |
| `auth_reset` | 归一到 `auth_login` |

## 8 后续实现输入

1. TopBar 用户菜单不得恢复 `用户画像`。
2. 未登录保护路由不得把 `debrief` 或 `profile` 当业务目标。
3. 正式前端、URL fallback 和 scenario 都必须覆盖旧入口负向。
4. 设置页只承载账号、界面偏好和隐私数据控制。
