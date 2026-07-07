# 工作日志记录规范

本目录用于记录项目的日常开发工作日志，便于团队成员回顾进展、追踪决策过程与问题解决思路。

## 1 目录结构

```
docs/work-journal/
├── README.md              # 本文件，记录规范说明
├── TEMPLATES.md           # 日志模板资产
├── INDEX.md               # 日志索引，按月份组织的摘要导航
└── YYYY-MM-DD.md          # 按日期命名的工作日志
```

## 2 核心文件说明

| 文件 | 用途 | 维护频率 |
|------|------|----------|
| `INDEX.md` | 总索引，快速定位工作记录 | 每次新增日志后更新 |
| `README.md` | 规范说明，指导如何记录 | 规范变更时更新 |
| `TEMPLATES.md` | 日志模板与示例 | 写新日志前参考 |
| `YYYY-MM-DD.md` | 具体日期的详细工作记录 | 当天工作完成后创建/追加 |

## 2.1 协作约定

协作前必须先阅读本目录 `README.md` 的记录规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。

## 3 日志记录流程

**核心理念**：代码变更与工作日志在同一次 commit 中提交，使用 commit message 作为关联标识。

### 3.1 Commit message 语言规则

所有 commit message 必须使用英文，且必须通过 ASCII-only 校验。

- 约束范围包括 commit subject、body、bullet、trailer，以及写入 `INDEX.md` 和日志 `### 关联 Commit` 的同一条 commit message。
- 日志正文可以继续使用中文；只有 commit message 字符串必须是英文。
- 不得直接把中文 plan 标题、phase 标题、checklist 文案或缺陷摘要复制进 commit message；必须翻译或概括为简洁英文。
- 提交前必须先校验完整 commit message：

```bash
LC_ALL=C perl -ne 'if (/[^\x00-\x7F]/) { print; exit 1 }' <message-file>
```

安装仓库 hook 后，`scripts/git-hooks/commit-msg` 会执行同一条 ASCII-only 拦截。

### 步骤 1：创建或追加日志文件

- 文件名格式：`YYYY-MM-DD.md`（如 `2025-01-15.md`）
- 如果当天文件已存在，在文件末尾追加新记录
- 如果当天文件不存在，创建新文件

### 步骤 2：记录工作内容

使用 [TEMPLATES.md](./TEMPLATES.md) 中的模板记录具体工作内容。

### 步骤 3：更新索引（重要）

在 `INDEX.md` 中添加索引条目，**每个 commit 独立一行**：

```markdown
| [2025-01-15](2025-01-15.md) | `feat(module): implement feature` | #tag |
```

### 步骤 4：一次性提交

将代码变更 + 日志文件 + 索引一起提交。

---

## 4 索引机制

### 索引结构

`INDEX.md` 采用按月分组的表格结构，**每个 commit 独立一行**：

```markdown
## 2025-01

| 日期 | Commit Message | 标签 |
|------|----------------|------|
| [2025-01-15](2025-01-15.md) | `feat(module): implement feature` | #feat |
```

**如何通过日志找代码？**

使用 `git log --grep` 检索：
```bash
git log --grep="feat(module): implement feature"
```

## 5 模板资产

- [日志模板](./TEMPLATES.md) — 推荐的日常记录结构与自动提交备注格式

## 6 检查清单

完成工作日志记录前，确认以下事项：

- [ ] 日志文件命名正确（`YYYY-MM-DD.md`）
- [ ] 包含完成事项和关联 Commit，且关联 Commit 为英文 / ASCII-only
- [ ] 已更新 `INDEX.md` 索引（每个 commit 独立一行）
- [ ] 标签选择恰当
- [ ] 已校验完整 commit message 不含非 ASCII 字符
- [ ] 代码和日志在同一次 commit 中提交

## 7 记录原则

1. **及时性**：当天的工作当天记录，避免遗忘细节
2. **完整性**：记录工作内容、技术决策、问题解决的完整过程
3. **可追溯**：关联相关的代码文件、PR、Issue 等引用
4. **简洁性**：使用简洁的语言，突出重点
5. **索引同步**：日志与索引保持同步更新
