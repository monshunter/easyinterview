# E2E.P0.072 Practice derived source isolation privacy

> **场景 ID**: E2E.P0.072
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

验证 missing、cross-user、wrong-target、draft 和 empty source 均返回统一 validation envelope，且不泄露 source id 或 debrief raw text。
