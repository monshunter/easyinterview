# 工作日志模板

## 1 简洁模板

```markdown
## HH:MM 工作记录

### 完成事项

- 具体完成的工作描述

### 关联 Commit

- `feat(scope): implement feature gate`

### 备注

其他需要记录的内容（可选）。
```

## 2 自动提交备注模板

```markdown
### 备注

Auto-committed by /tdd phase-commit, plan: {plan-name}
```

## 3 Commit message 规则

`### 关联 Commit` 中的 commit message 必须与真实 git commit 完全一致，并且必须为英文 / ASCII-only。日志正文可以使用中文。
