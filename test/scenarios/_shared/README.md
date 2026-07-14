# 共享资源

本目录存放跨场景套件的共享脚本和工具。

## 1 约定路径

| 文件 | 用途 |
|------|------|
| `scripts/local-dev-runtime.sh` | host-run backend/frontend 重启、PID、日志与本地调试摘要 helper |

代码层单测、源码契约、lint 与 build 不在本目录提供 wrapper；前后端全量单测统一运行根 `make test`。
