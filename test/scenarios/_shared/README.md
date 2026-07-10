# 共享资源

本目录存放跨场景套件的共享脚本和工具。

## 1 约定路径

| 文件 | 用途 |
|------|------|
| `scripts/frontend-real-backend-gate.sh` | frontend owner 场景前置 real-mode generated-client gate |
| `scripts/frontend-real-backend-verify.sh` | frontend owner 场景日志中的 real-mode gate、可配置 owner test marker 与 Vitest runner/summary 证据检查 |
| `scripts/local-dev-runtime.sh` | host-run backend/frontend 重启、PID、日志与本地调试摘要 helper |
| `scripts/resume-runtime-negative-gate.sh` | Resume P0.075-P0.080 共用的 old-mode / out-of-scope-module production negative gate |
| `scripts/scenario-evidence-setup.sh` | 从场景目录创建标准 `.test-output/e2e/<scenario>/setup.env` 证据 |
| `scripts/scenario-evidence-cleanup.sh` | 删除标准 `setup.env`，保留 trigger/verify 证据 |
