# EasyInterview 目标用户流程

> **版本**: 2.23
> **状态**: active
> **更新日期**: 2026-07-12

## 1 文档目的

本文档定义当前静态 UI 原型对应的目标用户流程。当前流程按 D-22 收敛为三条主线：首页启动、面试、简历管理；报告后的下一步只保留复练当前轮和进入下一轮。

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
  -> 直接打开简历详情
  -> 只读简历详情
  -> 在 Parse 或 Workspace 绑定

Interview
  -> 浏览面试规划列表
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
├─ 查看已绑定简历
├─ 确认 InterviewRound
└─ 立即面试 -> Practice
```

首次导入链路只有一次全页确认。解析成功即代表规划已保存；`workspace` 是既有规划回访枢纽，不是 parse 与 session 之间的第二确认页。

## 5 面试规划回访

```text
Workspace
├─ 面试规划列表（无上下文一级入口）
│  ├─ 打开已有规划
│  └─ 从新 JD 创建规划 -> Home
└─ 当前面试规划（带 targetJobId / planId 上下文）
   ├─ JD / 简历 / 轮次
   ├─ 公司轻情报卡片
   ├─ 当前规划会话记录
   └─ 立即面试
```

Workspace 服务回访、切换规划和再次发起面试，不是泛岗位资产管理中心。

## 6 面试与报告

```text
Practice
├─ 连续文本聊天
├─ 电话入口置灰
├─ 用户与 AI 自然推进
└─ 结束并生成报告

Report Dashboard
├─ 准备度
├─ 维度详情
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
│  └─ 注册成功后直接打开详情
└─ 简历详情
   └─ 只读原始正文
```

简历创建不提供轻量问答建档，不展示解析动画或结构化草稿确认页。解析成功后，backend parse 负责根据 LLM `displayName` / 结构化结果生成可识别 `displayName`；若解析失败但可读正文已抽取，backend 必须写入非通用 fallback 名称。解析前或旧数据不得从原文第一行或文件名派生可见名称，只能显示中性 fallback 名称、来源信息或 LLM/结构化 headline。详情页只阅读原始简历正文，不提供导出、复制、编辑、改写或原件弹层。

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

`auth_profile_setup` 是首次账号资料补全，用于登录后接续 pendingAction 前的账号基础信息确认，不是用户画像。

## 9 范围外流程

以下流程不属于目标用户流程、静态原型页面或正式前端入口：

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
| 用户未登录执行写入动作 | 进入邮箱验证码登录，成功后接续 pendingAction |
| `#route=debrief` / `#route=debrief_full` | 归一到 `home` |
| `#route=profile` | 归一到 `home` |
| 范围外 `auth_reset` | 归一到 `auth_login` |
