# 共享资源

本目录存放跨场景套件的共享脚本和工具。

## 1 约定路径

| 文件 | 用途 |
|------|------|
| `scripts/common.sh` | 通用 helper 函数 |
| `scripts/image-cache.sh` | 容器镜像预热脚本 |
| `scripts/frontend-real-backend-gate.sh` | frontend owner 场景前置 real-mode generated-client gate |
| `scripts/frontend-real-backend-verify.sh` | frontend owner 场景日志中的 real-mode gate 证据检查 |
| `scripts/local-dev-runtime.sh` | host-run backend/frontend 重启、PID、日志与本地调试摘要 helper |

这些路径只有在文件真实存在时才是可执行入口。缺失时不得杜撰命令。
