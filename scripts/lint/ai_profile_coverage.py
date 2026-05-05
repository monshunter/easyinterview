#!/usr/bin/env python3
"""Validate A3/F3/Product-UI AI profile coverage.

The gate reads:

1. F3 `prompt-rubric-registry` baseline feature_key table.
2. A3 Product/UI AI Capability Catalog.
3. `config/ai-profiles.yaml` model profile catalog.
4. `config/ai-providers.yaml` provider registry.

Every documented default profile must exist and resolve to a legal capability,
provider ref, status, and unsupported_reason where required.
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path
from typing import Any

import yaml

PROFILE_RE = re.compile(r"`([a-z0-9_.-]+\.default)`")
ALLOWED_CAPABILITIES = {"chat", "embed", "stt", "realtime", "rerank", "judge"}
ALLOWED_STATUSES = {"active", "disabled", "unsupported"}


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def section_after(text: str, marker: str) -> str:
    if marker not in text:
        raise ValueError(f"missing section marker: {marker}")
    section = text.split(marker, 1)[1]
    next_header = re.search(r"^#{2,4}\s", section, re.MULTILINE)
    if next_header:
        section = section[: next_header.start()]
    return section


def documented_profiles(repo: Path) -> set[str]:
    f3 = read(repo / "docs/spec/prompt-rubric-registry/spec.md")
    a3 = read(repo / "docs/spec/ai-provider-and-model-routing/spec.md")
    profiles: set[str] = set()
    profiles.update(PROFILE_RE.findall(section_after(f3, "#### 3.1.1 12 个当前 baseline feature_key 字典")))
    profiles.update(PROFILE_RE.findall(section_after(a3, "### 4.5 Product/UI AI Capability Catalog")))
    bad = [p for p in profiles if "*" in p]
    if bad:
        raise ValueError("wildcard profile names are forbidden: " + ", ".join(sorted(bad)))
    return profiles


def load_yaml(path: Path) -> Any:
    with path.open(encoding="utf-8") as fh:
        return yaml.safe_load(fh) or {}


def load_profiles(repo: Path) -> dict[str, dict[str, Any]]:
    out: dict[str, dict[str, Any]] = {}
    profile_path = repo / "config/ai-profiles.yaml"
    doc = load_yaml(profile_path)
    profiles = doc.get("profiles")
    if not isinstance(profiles, list):
        raise ValueError(f"{profile_path}: missing profiles[]")
    for profile in profiles:
        if not isinstance(profile, dict):
            raise ValueError(f"{profile_path}: profile entry must be a mapping")
        name = str(profile.get("name", ""))
        if not name:
            raise ValueError(f"{profile_path}: profile entry missing name")
        if name in out:
            raise ValueError(f"duplicate profile name: {name}")
        out[name] = profile
    return out


def load_provider_registry(repo: Path) -> dict[str, dict[str, Any]]:
    doc = load_yaml(repo / "config/ai-providers.yaml")
    providers = doc.get("providers") or []
    out: dict[str, dict[str, Any]] = {}
    for provider in providers:
        name = str(provider.get("name", ""))
        if not name:
            raise ValueError("provider registry entry missing name")
        if name in out:
            raise ValueError(f"duplicate provider name: {name}")
        out[name] = {
            "protocol": str(provider.get("protocol", "")),
            "capabilities": {str(c) for c in provider.get("capabilities") or []},
        }
    return out


def validate(repo: Path) -> list[str]:
    problems: list[str] = []
    required = documented_profiles(repo)
    profiles = load_profiles(repo)
    providers = load_provider_registry(repo)

    missing = required - set(profiles)
    if missing:
        problems.append("missing profiles: " + ", ".join(sorted(missing)))

    for name, profile in sorted(profiles.items()):
        capability = str(profile.get("capability", ""))
        status = str(profile.get("status", ""))
        if capability not in ALLOWED_CAPABILITIES:
            problems.append(f"{name}: invalid capability {capability!r}")
        if status not in ALLOWED_STATUSES:
            problems.append(f"{name}: invalid status {status!r}")
        if status in {"disabled", "unsupported"} and not str(profile.get("unsupported_reason", "")).strip():
            problems.append(f"{name}: {status} profile missing unsupported_reason")
        default = profile.get("default") or {}
        provider_ref = str(default.get("provider_ref", ""))
        if not provider_ref:
            problems.append(f"{name}: missing default.provider_ref")
            continue
        if provider_ref not in providers:
            problems.append(f"{name}: provider_ref {provider_ref!r} not found")
            continue
        provider = providers[provider_ref]
        if capability and capability not in provider["capabilities"]:
            problems.append(
                f"{name}: capability not declared by provider {provider_ref!r}: {capability}"
            )
        if status == "active" and provider["protocol"] == "stub":
            problems.append(f"{name}: active profile must not use stub provider {provider_ref!r}")

    return problems


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate AI profile coverage")
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    args = parser.parse_args()
    repo = args.repo_root.resolve()
    try:
        problems = validate(repo)
    except Exception as exc:  # noqa: BLE001 - lint script should report cleanly
        print(f"ai_profile_coverage: {exc}", file=sys.stderr)
        return 2
    if problems:
        print("ai_profile_coverage: drift detected", file=sys.stderr)
        for problem in problems:
            print(f"  - {problem}", file=sys.stderr)
        return 1
    print("ai_profile_coverage: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
