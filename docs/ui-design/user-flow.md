# EasyInterview 目标用户流程

> **版本**: 2.11
> **状态**: active
> **更新日期**: 2026-06-29

## 1 文档目的

本文档定义当前静态 UI 原型对应的目标用户流程。当前流程按 D-22 收敛为三条主线：首页启动、模拟面试、简历管理；报告后的下一步只保留复练当前轮和进入下一轮。

## 2 主流程总览

```text
Home
  -> JD 导入
  -> Parse & Confirm
  -> Interview Session
  -> Generating
  -> Report Dashboard
     ├─ 复练当前轮
     └─ 进入下一轮

Resume
  -> 上传 / 粘贴创建
  -> Agent 解析
  -> 预览确认保存
  -> 在 Parse 或 Workspace 绑定

Mock Interview
  -> 回访当前面试规划
  -> 切换 JD / 简历 / 轮次
  -> 立即面试
```

## 3 首页启动流程

```text
Home
├─ 粘贴 JD
├─ 上传 JD
├─ URL 导入
└─ 还没有简历？1 分钟创建
```

首页不提供复盘辅助入口。用户有 JD 时进入解析确认；用户没有简历时进入简历创建。

## 4 JD 解析确认与启动

```text
Parse & Confirm
├─ 核对 JD 基础信息
├─ 核对必需项 / 加分项 / 隐性关注点
├─ 绑定简历
├─ 确认 InterviewRound
├─ 立即面试 -> Practice
└─ 仅保存规划 -> Workspace
```

首次导入链路只有一次全页确认。`workspace` 不再插在 parse 与 session 之间作为第二确认页。

## 5 模拟面试回访

```text
Workspace
├─ 当前面试规划
├─ JD / 简历 / 轮次
├─ 公司轻情报卡片
├─ 当前规划会话历史
└─ 立即面试
```

Workspace 服务回访、切换规划和再次发起面试，不是泛岗位资产管理中心。

## 6 面试与报告

```text
Practice
├─ 文本面试 / 语音面试
├─ 带提示练习 / 严格模拟
├─ 问题推进
└─ 结束并生成报告

Report Dashboard
├─ 准备度
├─ 维度详情
├─ 题目回顾
├─ 证据详情
├─ 复练当前轮
└─ 进入下一轮
```

报告必须隶属于 session。`复练当前轮` 和 `进入下一轮` 是唯一一对开练 CTA。

## 7 简历流程

```text
Resume
├─ 平铺简历列表
├─ 新建简历
│  ├─ 上传文件
│  ├─ 粘贴文本
│  ├─ Agent 解析
│  └─ 预览确认保存
└─ 简历详情
   ├─ 预览
   ├─ 改写建议（仅采纳）
   └─ 手动编辑
```

简历创建不提供轻量问答建档。改写建议采纳后必须先预览整份结果，再选择覆盖原简历或保存为新简历。

## 8 认证与设置

```text
登录触发点
├─ 立即面试
├─ 复练当前轮
├─ 进入下一轮
├─ 保存简历
└─ 设置与隐私

Auth
├─ auth_login
├─ auth_verify
├─ auth_profile_setup
└─ auth_logout
```

`auth_profile_setup` 是首次账号资料补全，用于登录后恢复 pendingAction 前的账号基础信息确认，不是用户画像。

## 9 已删除流程

以下流程不再作为目标用户流程、静态原型页面或正式前端入口：

- 真实面试复盘
- 复盘面试
- 用户画像查看和纠偏
- CandidateProfile / ExperienceCard 维护
- 岗位推荐
- 公司情报独立详情页
- 独立成长中心
- 独立错题本 / 单题 Drill
- 独立 Voice 页面

## 10 异常和边界

| 情况 | 处理 |
|------|------|
| 用户没有 JD | 首页提示粘贴、上传或 URL 导入 |
| 用户没有简历 | 首页和 Parse 提供创建简历入口 |
| 用户未登录执行写入动作 | 进入邮箱验证码登录，成功后恢复 pendingAction |
| `#route=debrief` / `#route=debrief_full` | 归一到 `home` |
| `#route=profile` | 归一到 `home` |
| 旧 `auth_reset` | 归一到 `auth_login` |
