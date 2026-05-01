# UI Canvas Runner Path Fix 交付复盘报告

> **日期**: 2026-05-01
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复 `easyinterview-ui/run.sh` 启动后 `easyinterview-canvas.html` 空白的问题，覆盖画板包装脚本路径、iframe 原型入口路径、默认启动入口文件名。
- 成功证据：
  - `curl -I http://127.0.0.1:5173/easyinterview-canvas.html` 返回 `HTTP/1.0 200 OK`。
  - `curl -I http://127.0.0.1:5173/design-canvas.jsx` 返回 `HTTP/1.0 200 OK`，覆盖截图中 404 的画板脚本。
  - `curl -I http://127.0.0.1:5173/EasyInterview.html` 返回 `HTTP/1.0 200 OK`，覆盖画板 iframe 的原型入口。
  - `./run.sh -f easyinterview-canvas.html --no-open` 在 5174 端口启动后，`easyinterview-canvas.html`、`design-canvas.jsx`、`EasyInterview.html` 均返回 200。
  - `./run.sh --no-open` 使用默认入口 `EasyInterview.html` 启动并返回 200。
  - `git diff --check -- easyinterview-ui/easyinterview-canvas.html easyinterview-ui/run.sh` 通过。

## 2 会话中的主要阻点/痛点

- 画板 HTML 使用了与 `run.sh` 服务根目录不一致的资源路径。
  - **证据**：`easyinterview-canvas.html` 曾引用 `easyinterview-ui/design-canvas.jsx` 与 `easyinterview-ui/easyInterview.html`，但 `run.sh` 会先 `cd` 到 `easyinterview-ui/` 再启动静态服务器。
  - **影响**：浏览器请求多出一层 `easyinterview-ui/`，导致画板脚本 404，页面只剩背景空白。
- 默认入口文件名与真实文件大小写不一致。
  - **证据**：`run.sh` 默认值为 `easyInterview.html`，仓库真实文件为 `EasyInterview.html`。
  - **影响**：默认启动脚本会找不到入口，用户需要手动传 `-f` 才能绕过。
- 当前 change-intake 无法稳定命中 UI 原型运行入口。
  - **证据**：matcher 将本次查询推荐到 `openapi-v1-contract/002-fixtures-and-mock-source`，与 canvas runner 不相关。
  - **影响**：问题定位需要手动回到 `easyinterview-ui` 文件路径，无法由 plan/context 自动承接。
- 浏览器级验证环境有额外限制。
  - **证据**：in-app browser backend 未发现可用 IAB；本机 headless Chrome 在 macOS 后台网络/更新器日志中超时，未产出可用 DOM。
  - **影响**：本次以 HTTP 资源状态和 `run.sh` 启动验证作为交付证据，没有生成新的视觉截图。

## 3 根因归类

- `no repo change needed`：根因是静态资源路径和文件名大小写笔误，代码层修复即可。
- `spec-plan`：UI 原型运行入口缺少可检索的轻量 plan/context，change-intake 对这类问题会误路由。
- `README`：`run.sh` 的默认入口和 canvas 启动方式没有在脚本注释之外形成可验证说明。

## 4 对流程资产的改进建议

- 为 `easyinterview-ui` 增加一个轻量 smoke 验证脚本，至少检查 `EasyInterview.html`、`easyinterview-canvas.html`、`design-canvas.jsx` 和 iframe 入口均可访问。
  - **落点**：README / test
  - **优先级**：medium
- 为 UI 原型运行入口补充可被 change-intake 检索的实现映射，避免 canvas / run.sh 问题误路由到 OpenAPI 计划。
  - **落点**：spec-plan
  - **优先级**：medium
- 若后续继续依赖画板页评审 UI，补一条稳定的 browser smoke 方案，明确在 in-app browser 不可用时的替代验证命令。
  - **落点**：README
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最值得补的是 UI runner smoke，直接防止路径和大小写回归。
- change-intake 检索映射可以和后续 UI 原型 plan 一起补，避免为临时文件结构单独固化规则。
- 本次 bug 属于明显路径/大小写配置笔误，未创建独立 BUG 记录；若后续出现同类多次回归，再沉淀到 `docs/bugs/PATTERNS.md`。
