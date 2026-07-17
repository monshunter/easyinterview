# Harness 文档与技能体系全量迁移计划

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-07-17

## 1 背景与目标

依据 `../../spec.md`，在一个交付内将当前 spec-centric plan/context 体系切换为 Spec / Change / Decision 体系，并删除旧入口，不保留长期双轨兼容。

## 2 范围

### 2.1 In Scope

- 可重建、commit-aware 的文档索引与 owner 路由工具。
- 全部 `docs/spec/*` subject 的旧 plan/context/history/index 包装层清理。
- 根治理文档、docs 规则和相关 Skills 的轻量合同。
- R1、R2、R2/R3 review replay 与结构/成本/零残留 gate。

### 2.2 Out of Scope

- 改写产品业务语义或产品代码。
- 将历史 plan 逐份转写为 Change。
- 保留旧 context/plan 入口的兼容层。

## 3 质量门禁分类

- **Plan 类型**：tooling + docs architecture + governance。
- **TDD 策略**：索引、路由、迁移与 drift gate 均先写失败断言，再实现最小逻辑并回归。
- **BDD 策略**：BDD-N/A；本计划不引入产品用户行为。
- **替代验证 gate**：Python contract tests、repo structure lint、Skill contract tests、R1/R2/review replay、旧结构 zero-reference search。

## 4 实施阶段

### Phase 1: 冻结迁移合同

#### 1.1 建立 owner Spec、基线指标和迁移 fixture

#### 1.2 固化旧结构与必须保留质量边界的失败断言

### Phase 2: 生成式索引与精确路由

#### 2.1 实现从 Spec/Change/Decision、Markdown 链接、Git 与精确标识符生成索引

#### 2.2 实现 commit 失效、通用词降权、low-confidence 候选与可审计理由

### Phase 3: 文档结构全量迁移

#### 3.1 将当前迁移计划转换为冻结前的单一 Change

#### 3.2 删除所有 context、plan/checklist/普通 BDD、history 和分层 INDEX

#### 3.3 修复有效引用，保留 Spec、必要 Change/Decision、Bug/报告/迁移历史证据

### Phase 4: 治理与 Skills 收敛

#### 4.1 集中公共政策并按 R0-R3 风险加载

#### 4.2 将 Skills 收敛为 locate/design/execute/review/operate/closeout 能力，移除旧级联和全量预读

#### 4.3 更新初始化、文档、索引、场景和收尾合同及其测试

### Phase 5: Replay 与成本验收

#### 5.1 执行 R1 小 Bug replay

#### 5.2 执行 R2 跨层功能 replay

#### 5.3 执行 R2/R3 review replay

#### 5.4 对比读取文件/字节、工具调用、流程文件触碰量、路由准确率和缺陷发现

### Phase 6: 旧入口删除与全量收口

#### 6.1 删除旧 validator/matcher/template/Skill 兼容入口

#### 6.2 执行全量测试、build、lint、docs/links/zero-reference gate

#### 6.3 冻结 Change，确认 Spec 为唯一当前真理源

## 5 风险与恢复

- 路由能力退化：以精确 owner fixture、歧义 fixture 和 replay 比较阻止删除旧入口。
- 历史证据误删：只删除 Git 已保存且不再承接当前合同的包装层；Bug、报告、migration 历史保留。
- Skill 级联失效：为每个保留入口维护最小合同测试，并在同一提交内切换调用方。
- 一次性迁移中断：所有阶段保持 feature branch 可恢复，不向 main 自动合并。

## 6 完成标准

完成标准为 owner Spec A1-A12 全部由当前文件、测试、replay 指标和零残留搜索证明；窄测试或旧 PASS 不得替代全量验收。
