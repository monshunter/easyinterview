# local-dev-stack/001-bootstrap 交付复盘报告

> **日期**: 2026-04-28
> **审查人**: Claude

**关联计划**: [local-dev-stack/001-bootstrap](../spec/local-dev-stack/plans/001-bootstrap/plan.md)

## 1 复盘范围与成功证据

本次交付覆盖 [local-dev-stack/001-bootstrap](../spec/local-dev-stack/plans/001-bootstrap/plan.md) 全部 4 个 phase / 20 个 checklist 项，把 spec D-1..D-9 落地为可在本机重复跑通的 docker compose + Make 生命周期。

成功证据（详见 [docs/work-journal/2026-04-28.md](../work-journal/2026-04-28.md)）：

- 4 个 phase commit 全部 ff-merge 回 dev：`f9143b7` / `920ffef` / `b18bc63` / `2623438`。
- C-1..C-9 AC 重跑：`make dev-up` exit 0 + `dev-doctor` summary `ok=3 degraded=0 down=0 total=3`；macOS Docker Desktop 双栈 IPv6 squatter 复现 5432 冲突 → `dev-up` exit 1 + doctor 报 `port conflict cmd=Python`；二次 `dev-up` 0s `already healthy`；写入 `ac_probe` → down → up → 数据完整；`DEV_RESET_FORCE=1` 删 3 卷后下一次 `dev-up` 表不存在；`select extname='vector'` 返回一行；`dev-logs` 工作；当前无 `/metrics` 与 AIClient 组件，doctor 已按 label 提前实现两条路径。
- `/sync-doc-index --check` zero drift；plans/INDEX 已迁到 §2 Completed。
- `dev-doctor.sh` 173 行（≤ 200 line 阈值）；POSIX sh + jq；schema 通过 jq 自校。

## 2 会话中的主要阻点/痛点

### 2.1 `docker compose up --wait` 与一次性 init 容器的语义冲突

- **证据**：Phase 1.5 自检阶段 `docker compose up -d --wait` 在 `minio-init exited (0)` 后整体返回非 0；该现象迫使 Phase 2 Makefile 把 `up` 拆成「按服务名 `--wait` 依赖」+「单独跑 init 后轮询退出码」两步。
- **影响**：增加了 Phase 2 ~30 行 Makefile 代码与一个 `_wait_init` 内部 target；后续 child 接入新的 init 容器（例如 db migration、seed）时，会再次踩到同一个坑。

### 2.2 macOS Docker Desktop 端口冲突复现需要双栈 IPv6 监听

- **证据**：3.4 第一轮用 `nc -l 5432` 与第二轮用纯 IPv4 Python 监听都未触发冲突，因 docker-desktop 走 IPv6 监听。第三轮显式 `socket(AF_INET6) + IPV6_V6ONLY=0` 才真正占住端口让 docker daemon 报 `bind: address already in use`。
- **影响**：在没有这个 platform-specific 知识时，C-2 验收看似「过了」实则没复现冲突；调试 ~10 分钟。已写入 `deploy/dev-stack/README.md` §6 故障排查。

### 2.3 `lsof` 默认 COMMAND 列截断到 9 字符导致 docker 进程过滤失效

- **证据**：第一版 `down_or_conflict` 把 docker-desktop 的 `com.docker.backend` 当作非 docker 占用，对 redis-dev / minio-dev 误报 `port conflict cmd=com.docke`；改用 `lsof +c 0` 取完整 cmd 后过滤命中。
- **影响**：让 Phase 3 的 doctor 报告一度「全员 DOWN」假阳性；如果不解决，C-2 验收的 reason 字段无法只把 postgres-dev 标 DOWN。

### 2.4 MinIO 不可变镜像 tag 在 spec 阶段未回填，由 plan 落地时挑选

- **证据**：spec D-2 写「`minio/minio:<immutable-release-tag>`，具体不可变 tag 在 001-bootstrap plan 落地时回填」。落地时只能挑一个「记忆中应该存在」的 RELEASE 字符串，靠 `docker pull` 兜底验证。
- **影响**：本次靠 `docker pull` 提前验证 tag 存在，未踩坑；但若挑了不存在的 tag，会到 Phase 1.5 自检阶段才暴露。

## 3 根因归类

| # | 根因 | 类别 |
|---|------|------|
| 2.1 | compose `--wait` 对 `restart: "no"` 的一次性 init 服务把 `exited(0)` 视为失败；spec / plan 未提示这条 wait-pattern | spec-plan + README |
| 2.2 | macOS Docker Desktop 用 IPv6 监听，常规 IPv4 squatter 不冲突；C-2 验收脚本未约束复现路径 | README（已落地）+ spec-plan（C-2 描述太抽象） |
| 2.3 | `lsof` 默认列宽截断属一次性实现细节，doctor 脚本里 +c 0 后已无重复风险 | 无需仓库改动 |
| 2.4 | spec 把 immutable tag 回填责任甩给 plan，但 plan 没要求事前 `docker pull` 验证 | spec-plan |

## 4 对流程资产的改进建议

- **建议 A**：在 `deploy/dev-stack/README.md` 新增「为 compose 添加新的 init / one-shot 容器」小节，文档化「`docker compose up -d --wait` 不要包含 `restart: "no"` 的 init service；按服务名分批 `--wait`，再轮询 init 退出码」这条 pattern。
  - **落点**：README（`deploy/dev-stack/README.md`）
  - **优先级**：medium
- **建议 B**：在 `local-dev-stack` spec C-2 验收行追加「macOS Docker Desktop 必须双栈 IPv6 监听才能真复现冲突」备注，避免后续修订把 C-2 误判为已通过。
  - **落点**：spec-plan（`docs/spec/local-dev-stack/spec.md` §6 C-2）
  - **优先级**：medium
- **建议 C**：把「新增镜像 tag 必须先 `docker pull` 验证存在」加入 `deploy/dev-stack/README.md` §7 升级与扩展（或 spec D-2 备注），降低后续 child 接入新 image 时挑错 tag 的概率。
  - **落点**：README + 可选 spec-plan
  - **优先级**：low
- **建议 D**：未来 `/tdd` 在显式 `--phase-commit` 模式下，可在 phase 收尾自动跑一次 `sync-doc-index --check` 而不是依赖人手；但本次手动调用没出错，属可选优化。
  - **落点**：skill（`.agent-skills/tdd/SKILL.md`）
  - **优先级**：low

## 5 建议优先级与后续动作

- **下一轮最值得做**：建议 A（init container wait pattern）+ 建议 B（C-2 复现路径备注）。两条直接对应本次踩过的坑，且未来 backend / worker 接入 compose 时高概率再次相关。
- **可延后**：建议 C 与建议 D，影响范围小，不阻塞当前 W1 后续 child。
- **不需要建 BUG**：本次没有 production-style bug，doctor 误报与 wait 误判都是新代码 first-run 调整，非回归故障；按 §3 不另开 `/bug-report`。
