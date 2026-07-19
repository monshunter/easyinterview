#!/usr/bin/env python3
"""Install, inspect, upgrade, or repair the Project Arch v1 kernel.

The tool is intentionally self-contained. It discovers only bounded repository
facts, writes a fixed set of Arch-owned files, preserves unknown project files,
and rolls back the whole mutation set when a write fails.
"""

from __future__ import annotations

import argparse
from dataclasses import dataclass
import json
import os
from pathlib import Path
import re
import stat
import sys
import tempfile
from typing import Any


ARCH_MARKER_V0 = "<!-- project-arch: v0 -->"
ARCH_MARKER_V1 = "<!-- project-arch: v1 -->"
RESULT_SCHEMA = "project-arch.result/v1"
OWNER_PREFIX = "project-arch-owned:"
ROOT_BLOCK_BEGIN = "<!-- project-arch-owned:start arch.root@v1 -->"
ROOT_BLOCK_END = "<!-- project-arch-owned:end arch.root@v1 -->"
ENV_NAMES = (
    "env-setup.sh",
    "env-status.sh",
    "env-verify.sh",
    "env-redeploy.sh",
    "env-cleanup.sh",
)


@dataclass(frozen=True)
class Asset:
    path: str
    content: bytes
    mode: int = 0o644


@dataclass(frozen=True)
class FilePreimage:
    content: bytes
    mode: int
    mtime_ns: int


def _json_result(
    status: str,
    operation: str,
    *,
    changed: bool = False,
    inventory: dict[str, list[dict[str, str]]] | None = None,
    actions: list[dict[str, str]] | None = None,
    recovery: dict[str, Any] | None = None,
    reasons: list[str] | None = None,
) -> dict[str, Any]:
    result: dict[str, Any] = {
        "schema": RESULT_SCHEMA,
        "status": status,
        "operation": operation,
        "changed": changed,
        "inventory": inventory
        or {"absent": [], "compatible": [], "custom": [], "conflicting": []},
        "actions": actions or [],
    }
    if recovery is not None:
        result["recovery"] = recovery
    if reasons:
        result["reasons"] = reasons
    return result


def _emit(result: dict[str, Any], exit_code: int) -> int:
    print(json.dumps(result, ensure_ascii=True, sort_keys=True))
    return exit_code


def _safe_root(raw: str) -> Path:
    root = Path(raw).resolve(strict=True)
    if not root.is_dir():
        raise ValueError("repository root is not a directory")
    if root == Path(root.anchor) or root == Path.home().resolve():
        raise ValueError("refusing a broad repository root")
    return root


def _safe_path(root: Path, relative: str) -> Path:
    parts = Path(relative).parts
    if not parts or Path(relative).is_absolute() or ".." in parts:
        raise ValueError("unsafe managed path")
    current = root
    for part in parts:
        current = current / part
        try:
            metadata = current.lstat()
        except FileNotFoundError:
            continue
        if stat.S_ISLNK(metadata.st_mode):
            raise ValueError(f"managed path is a symlink: {relative}")
        if current != root / relative and not stat.S_ISDIR(metadata.st_mode):
            raise ValueError(f"managed parent is not a directory: {relative}")
    return root / relative


def _read_project_name(root: Path) -> str | None:
    pyproject = root / "pyproject.toml"
    if pyproject.is_file() and not pyproject.is_symlink():
        try:
            import tomllib

            data = tomllib.loads(pyproject.read_text(encoding="utf-8"))
            value = data.get("project", {}).get("name")
            if isinstance(value, str) and value.strip():
                return value.strip()
        except (OSError, UnicodeDecodeError, ValueError, TypeError):
            pass
    readme = root / "README.md"
    if readme.is_file() and not readme.is_symlink():
        try:
            for line in readme.read_text(encoding="utf-8").splitlines():
                if line.startswith("# ") and line[2:].strip():
                    return line[2:].strip()
        except (OSError, UnicodeDecodeError):
            pass
    return None


def _slug(value: str) -> str:
    return re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")[:64] or "project"


def _has_source(root: Path) -> bool:
    extensions = {".py", ".go", ".js", ".ts", ".rs", ".java", ".c", ".cpp"}
    for dirname in ("src", "cmd", "app", "lib"):
        directory = root / dirname
        if not directory.is_dir() or directory.is_symlink():
            continue
        if any(path.is_file() and path.suffix in extensions for path in directory.rglob("*")):
            return True
    return False


def _has_tests(root: Path) -> bool:
    for dirname in ("tests", "test"):
        directory = root / dirname
        if not directory.is_dir() or directory.is_symlink():
            continue
        for path in directory.rglob("*"):
            name = path.name
            if path.is_file() and (
                name.startswith("test_")
                or name.endswith("_test.py")
                or name.endswith("_test.go")
                or name.endswith(".test.ts")
            ):
                return True
    return False


def _build_owner_exists(root: Path) -> bool:
    return any((root / name).is_file() for name in ("pyproject.toml", "go.mod", "package.json", "Cargo.toml"))


def _discover(root: Path, goal: str | None) -> tuple[str, str, list[str]]:
    project_name = _read_project_name(root) or "project"
    runtime_goal = bool(goal and re.search(r"\b(http|web|api|service|server|daemon)\b", goal, re.I))
    if goal is None and not _build_owner_exists(root):
        return project_name, "decision_required", ["project objective is not derivable from safe repository facts"]
    if runtime_goal:
        return project_name, "spec_required", ["runtime lifecycle requires an environment owner Spec and adapter"]
    if not (_build_owner_exists(root) and _has_source(root) and _has_tests(root)):
        return project_name, "spec_required", ["build, source, and owner-test evidence are incomplete"]
    return project_name, "ready", []


def _owned_markdown(role: str, title: str, body: str) -> bytes:
    return (
        f"<!-- {OWNER_PREFIX} {role}@v1 -->\n"
        f"<!-- project-arch-interface: {role}@v1 -->\n"
        f"# {title}\n\n{body.strip()}\n"
    ).encode("utf-8")


def _env_script(name: str) -> bytes:
    operation = name.removeprefix("env-").removesuffix(".sh")
    return (
        "#!/usr/bin/env sh\n"
        f"# {OWNER_PREFIX} arch.env.{operation}@v1\n"
        "# project-arch-env-adapter: v1\n"
        f"# interface: arch.env.{operation}\n"
        "set -eu\n"
        f"printf '%s\\n' '{{\"status\":\"NOT_CONFIGURED\",\"interface\":\"arch.env.{operation}\",\"handoff\":\"spec_required\"}}'\n"
        "exit 2\n"
    ).encode("utf-8")


def _context_yaml() -> bytes:
    return (
        "# project-arch-owned: arch.context@v1\n"
        "apiVersion: plancontext.agent.dev/v1alpha1\n"
        "kind: PlanContext\n"
        "metadata:\n"
        "  name: 001-bootstrap\n"
        "spec:\n"
        "  defaultTarget: project\n"
        "  targets:\n"
        "    project:\n"
        "      plan: ./plan.md\n"
        "      checklist: ./plan.md\n"
        "      spec: ../../spec.md\n"
    ).encode("utf-8")


def _assets(root: Path, project_name: str) -> tuple[Asset, ...]:
    subject = _slug(project_name)
    work_root = f"docs/spec/{subject}/plans/001-bootstrap"
    root_body = (
        "# Project documentation\n\n"
        f"{ARCH_MARKER_V1}\n\n"
        f"{ROOT_BLOCK_BEGIN}\n"
        "Project Arch v1 uses four layers: repository principles, workflow and engineering, "
        "current subject truth, and current work. INDEX files and minimal context.yaml files "
        "are rebuildable navigation projections, not semantic owners.\n"
        f"{ROOT_BLOCK_END}\n"
    ).encode("utf-8")
    spec = _owned_markdown(
        "arch.spec",
        f"{project_name} Project Spec",
        "## 1 Goal\n\nOwn the current project contract.\n\n"
        "## 2 Evidence\n\nBuild, source, and owner-test facts come from the repository.\n\n"
        "## 3 Acceptance\n\nProject-specific implementation and verification remain in their nearest owners.",
    )
    plan = _owned_markdown(
        "arch.plan",
        f"{project_name} Bootstrap Plan",
        "> **版本**: 1.0\n> **状态**: active\n> **更新日期**: 2026-07-19\n\n"
        "## 1 Goal\n\nBind the installed Project Arch interfaces to current project facts.\n\n"
        "## 2 Work\n\n- [ ] Confirm project-specific development, test, scenario, and environment owners.",
    )
    source = Path(__file__).read_bytes()
    values = [
        Asset("AGENTS.md", _owned_markdown("arch.principles", "Agent Instructions", "Use current repository owners, explicit evidence, safe changes, and recoverable execution.")),
        Asset("docs/README.md", root_body),
        Asset("docs/agent-workflow.md", _owned_markdown("arch.workflow", "Agent Workflow", "Resolve the current owner, implement the smallest authorized change, verify at the execution owner, and leave a precise recovery point.")),
        Asset("docs/development.md", _owned_markdown("arch.development", "Development Contract", "Source owners, focused test entry, and aggregate verification entry are derived from the project's build and test files.")),
        Asset("docs/spec/INDEX.md", _owned_markdown("arch.index", "Spec Index", f"- [{project_name}](./{subject}/spec.md)")),
        Asset(f"docs/spec/{subject}/spec.md", spec),
        Asset(f"docs/spec/{subject}/plans/INDEX.md", _owned_markdown("arch.plan-index", "Current Plans", "- [001-bootstrap](./001-bootstrap/plan.md)")),
        Asset(f"{work_root}/plan.md", plan),
        Asset(f"{work_root}/context.yaml", _context_yaml()),
        Asset("test/README.md", _owned_markdown("arch.test", "Test Contract", "Unit, contract, integration, and real-system tests remain distinct evidence layers.")),
        Asset("test/scenarios/README.md", _owned_markdown("arch.scenario", "Scenario Contract", "Real scenarios drive running systems and own setup, trigger, verification, evidence, and cleanup.")),
        Asset("test/scenarios/_shared/README.md", _owned_markdown("arch.scenario-shared", "Scenario Shared Contract", "Shared scenario helpers preserve isolation, redaction, cleanup, and truthful failure results.")),
        Asset("scripts/harness_arch.py", source, 0o755),
    ]
    values.extend(Asset(f"test/scenarios/{name}", _env_script(name), 0o755) for name in ENV_NAMES)
    return tuple(values)


def _installation(root: Path) -> str:
    path = root / "docs/README.md"
    if not path.exists():
        return "absent"
    if path.is_symlink() or not path.is_file():
        return "conflict"
    try:
        text = path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError):
        return "conflict"
    markers = re.findall(r"<!--\s*project-arch:\s*(v\d+)\s*-->", text)
    if len(markers) > 1:
        return "conflict"
    if markers == ["v1"]:
        return "v1"
    if markers == ["v0"]:
        return "v0"
    return "absent"


def _desired(asset: Asset, current: bytes | None, installation: str) -> bytes:
    if asset.path != "docs/README.md" or current is None:
        return asset.content
    try:
        text = current.decode("utf-8")
    except UnicodeDecodeError:
        return asset.content
    if installation in {"v0", "v1"}:
        managed = asset.content.decode("utf-8").split(ARCH_MARKER_V1, 1)[1].strip()
        text = text.replace(ARCH_MARKER_V0, ARCH_MARKER_V1, 1)
        pattern = re.compile(
            re.escape(ROOT_BLOCK_BEGIN) + r".*?" + re.escape(ROOT_BLOCK_END),
            re.DOTALL,
        )
        if pattern.search(text):
            text = pattern.sub(managed, text, count=1)
        else:
            text = text.rstrip() + "\n\n" + managed + "\n"
        return text.encode("utf-8")
    return asset.content


def _read_file(path: Path) -> tuple[bytes | None, int | None, str | None]:
    try:
        metadata = path.lstat()
    except FileNotFoundError:
        return None, None, None
    except OSError:
        return None, None, "cannot inspect managed path"
    if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISREG(metadata.st_mode):
        return None, None, "managed path is not a regular file"
    try:
        return path.read_bytes(), stat.S_IMODE(metadata.st_mode), None
    except OSError:
        return None, None, "cannot read managed file"


def _inventory(root: Path, assets: tuple[Asset, ...], installation: str) -> dict[str, list[dict[str, str]]]:
    result: dict[str, list[dict[str, str]]] = {"absent": [], "compatible": [], "custom": [], "conflicting": []}
    managed = {asset.path for asset in assets}
    for asset in assets:
        try:
            path = _safe_path(root, asset.path)
        except ValueError as exc:
            result["conflicting"].append({"path": asset.path, "reason": str(exc)})
            continue
        current, mode, error = _read_file(path)
        if error:
            result["conflicting"].append({"path": asset.path, "reason": error})
        elif current is None:
            result["absent"].append({"path": asset.path, "reason": "required Arch interface is missing"})
        elif current == _desired(asset, current, installation) and mode == asset.mode:
            result["compatible"].append({"path": asset.path, "reason": "matches the Project Arch v1 contract"})
        elif asset.path == "docs/README.md" and installation == "v0":
            result["compatible"].append({"path": asset.path, "reason": "known v0 marker can be upgraded"})
        else:
            result["conflicting"].append({"path": asset.path, "reason": "canonical Arch-owned content is incompatible"})
    for dirname in ("docs", "test", "scripts"):
        base = root / dirname
        if not base.is_dir() or base.is_symlink():
            continue
        for path in base.rglob("*"):
            if path.is_file() and not path.is_symlink():
                relative = path.relative_to(root).as_posix()
                if relative not in managed:
                    result["custom"].append({"path": relative, "reason": "project-owned extension"})
    for entries in result.values():
        entries.sort(key=lambda item: item["path"])
    return result


def _capture_transaction(
    root: Path,
    planned: list[tuple[Asset, bytes]],
) -> tuple[dict[str, FilePreimage | None], dict[str, tuple[int, int]]]:
    """Capture only files that may be written and existing parent metadata."""
    files: dict[str, FilePreimage | None] = {}
    directories: dict[str, tuple[int, int]] = {}
    for asset, _ in planned:
        path = _safe_path(root, asset.path)
        try:
            metadata = path.lstat()
        except FileNotFoundError:
            files[asset.path] = None
        else:
            files[asset.path] = FilePreimage(
                path.read_bytes(),
                stat.S_IMODE(metadata.st_mode),
                metadata.st_mtime_ns,
            )
        parent = path.parent
        while parent != root.parent:
            if parent.exists():
                metadata = parent.lstat()
                relative = parent.relative_to(root).as_posix() if parent != root else "."
                directories.setdefault(
                    relative,
                    (stat.S_IMODE(metadata.st_mode), metadata.st_mtime_ns),
                )
            if parent == root:
                break
            parent = parent.parent
    return files, directories


def _restore_transaction(
    root: Path,
    planned: list[tuple[Asset, bytes]],
    files: dict[str, FilePreimage | None],
    directories: dict[str, tuple[int, int]],
) -> None:
    for asset, _ in reversed(planned):
        path = root / asset.path
        preimage = files[asset.path]
        if preimage is None:
            if path.exists() or path.is_symlink():
                path.unlink()
        else:
            _atomic_write(path, preimage.content, preimage.mode)
            os.utime(path, ns=(preimage.mtime_ns, preimage.mtime_ns), follow_symlinks=False)

    candidate_dirs = {
        parent
        for asset, _ in planned
        for parent in (root / asset.path).parents
        if parent != root and parent.is_relative_to(root)
    }
    for directory in sorted(candidate_dirs, key=lambda path: len(path.parts), reverse=True):
        relative = directory.relative_to(root).as_posix()
        if relative not in directories and directory.exists():
            directory.rmdir()

    for relative, (mode, mtime_ns) in sorted(
        directories.items(),
        key=lambda item: len(Path(item[0]).parts),
        reverse=True,
    ):
        path = root if relative == "." else root / relative
        path.chmod(mode)
        os.utime(path, ns=(mtime_ns, mtime_ns), follow_symlinks=False)


def _atomic_write(path: Path, content: bytes, mode: int) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, temporary = tempfile.mkstemp(prefix=".harness-arch-", dir=path.parent)
    try:
        with os.fdopen(descriptor, "wb") as stream:
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        os.chmod(temporary, mode)
        os.replace(temporary, path)
        directory = os.open(path.parent, os.O_RDONLY | getattr(os, "O_DIRECTORY", 0))
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def _actions(
    mode: str,
    root: Path,
    assets: tuple[Asset, ...],
    installation: str,
    inventory: dict[str, list[dict[str, str]]],
) -> tuple[list[tuple[Asset, bytes]], list[str]]:
    conflicting = {item["path"] for item in inventory["conflicting"]}
    absent = {item["path"] for item in inventory["absent"]}
    blockers: list[str] = []
    if installation == "conflict":
        return [], ["Project Arch version marker is incompatible"]
    if mode == "upgrade" and installation not in {"v0", "v1"}:
        return [], ["upgrade requires a known v0 installation"]
    if mode == "repair" and installation != "v1":
        return [], ["repair requires Project Arch v1"]
    if mode == "init" and installation == "v0":
        return [], ["known v0 installations must use upgrade"]
    if mode == "check":
        if absent or conflicting or installation != "v1":
            blockers.append("Project Arch v1 is missing or drifted")
        return [], blockers

    planned: list[tuple[Asset, bytes]] = []
    for asset in assets:
        path = root / asset.path
        current, _, _ = _read_file(path)
        desired = _desired(asset, current, installation)
        if asset.path in conflicting:
            if mode == "repair" and installation == "v1":
                planned.append((asset, desired))
            else:
                blockers.append(f"{asset.path} conflicts with the canonical Arch owner")
        elif asset.path in absent:
            if mode == "init" and installation == "v1":
                blockers.append(f"{asset.path} is missing; use repair")
            else:
                planned.append((asset, desired))
        elif asset.path == "docs/README.md" and installation == "v0":
            planned.append((asset, desired))
    planned.sort(key=lambda item: (item[0].path == "docs/README.md", item[0].path))
    return planned, blockers


def run(mode: str, root: Path, goal: str | None, apply: bool) -> tuple[dict[str, Any], int]:
    project_name, readiness, reasons = _discover(root, goal)
    if readiness != "ready":
        return _json_result(readiness, mode, reasons=reasons), 2
    installation = _installation(root)
    assets = _assets(root, project_name)
    inventory = _inventory(root, assets, installation)
    planned, blockers = _actions(mode, root, assets, installation, inventory)
    public_actions = [{"operation": "write", "path": asset.path} for asset, _ in planned]
    if blockers:
        return _json_result("conflict", mode, inventory=inventory, actions=public_actions, reasons=blockers), 2
    if mode == "check":
        return _json_result("ready", mode, inventory=inventory), 0
    if not apply:
        return _json_result("ready", mode, inventory=inventory, actions=public_actions), 0
    if not planned:
        return _json_result("ready", mode, inventory=inventory), 0

    file_preimages, directory_preimages = _capture_transaction(root, planned)
    fail_after_raw = os.environ.get("HARNESS_ARCH_TEST_FAIL_AFTER_WRITES")
    fail_after = int(fail_after_raw) if fail_after_raw and fail_after_raw.isdigit() else None
    writes = 0
    try:
        for asset, desired in planned:
            path = _safe_path(root, asset.path)
            _atomic_write(path, desired, asset.mode)
            writes += 1
            if fail_after is not None and writes >= fail_after:
                raise OSError("injected write failure")
        verified = _inventory(root, assets, "v1")
        if verified["absent"] or verified["conflicting"]:
            raise OSError("post-write verification failed")
    except Exception:
        try:
            _restore_transaction(
                root,
                planned,
                file_preimages,
                directory_preimages,
            )
            rolled_back = True
        except Exception:
            rolled_back = False
        return (
            _json_result(
                "conflict",
                mode,
                inventory=inventory,
                actions=public_actions,
                recovery={
                    "rolled_back": rolled_back,
                    "resume_condition": "retry after the filesystem is stable" if rolled_back else "restore the repository checkpoint before retrying",
                },
                reasons=["Project Arch transaction failed"],
            ),
            2,
        )
    return _json_result("ready", mode, changed=True, inventory=verified, actions=public_actions), 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Project Arch v1 bootstrap tool")
    parser.add_argument("mode", choices=("init", "check", "upgrade", "repair"))
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--goal")
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    try:
        root = _safe_root(args.repo_root)
        result, exit_code = run(args.mode, root, args.goal, args.apply)
    except (OSError, ValueError):
        result = _json_result(
            "conflict",
            args.mode,
            recovery={"rolled_back": False, "resume_condition": "provide a safe readable repository root"},
            reasons=["repository preflight failed"],
        )
        exit_code = 2
    return _emit(result, exit_code)


if __name__ == "__main__":
    sys.exit(main())
