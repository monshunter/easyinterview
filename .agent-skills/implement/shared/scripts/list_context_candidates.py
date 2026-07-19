#!/usr/bin/env python3
"""List candidate plans for /implement no-arg mode.

Current project scaffolds use spec-centric plan contexts:
  docs/spec/<subspec>/plans/<plan>/context.yaml
"""

from __future__ import annotations

import argparse
import os
import re
import sys

try:
    import yaml
except ImportError:
    print(
        "ERROR: PyYAML is not installed.\n"
        "Install it with: pip3 install PyYAML",
        file=sys.stderr,
    )
    sys.exit(2)


HEADER_STATUS_RE = re.compile(r"^\s*>\s*\*\*状态\*\*:\s*(.+?)\s*$", re.MULTILINE)
HEADER_DATE_RE = re.compile(r"^\s*>\s*\*\*更新日期\*\*:\s*(\d{4}-\d{2}-\d{2})\s*$", re.MULTILINE)
STATUS_ORDER = {"active": 0, "draft": 1, "completed": 2, "unknown": 3}


def find_contexts(root: str) -> list[str]:
    """Find spec-centric context.yaml files below a caller-provided root."""
    abs_root = os.path.abspath(root)
    candidates: list[str] = []

    roots: list[str]
    if os.path.basename(abs_root) == "plan":
        docs_root = os.path.dirname(abs_root)
        roots = [os.path.join(docs_root, "spec")]
    elif os.path.basename(abs_root) == "docs":
        roots = [os.path.join(abs_root, "spec")]
    else:
        roots = [os.path.join(abs_root, "docs", "spec"), abs_root]

    for search_root in roots:
        if not os.path.isdir(search_root):
            continue
        if os.path.isfile(os.path.join(search_root, "context.yaml")):
            candidates.append(os.path.join(search_root, "context.yaml"))
            continue
        for dirpath, _, files in os.walk(search_root):
            if "context.yaml" not in files:
                continue
            context = os.path.join(dirpath, "context.yaml")
            parts = os.path.normpath(dirpath).split(os.sep)
            if "plans" in parts:
                candidates.append(context)

    return sorted(set(candidates))


def read_manifest(context_path: str) -> dict | None:
    try:
        with open(context_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
    except (OSError, yaml.YAMLError):
        return None
    return data if isinstance(data, dict) else None


def plan_status_and_date(plan_path: str | None) -> tuple[str, str]:
    if not plan_path or not os.path.isfile(plan_path):
        return "unknown", "0000-00-00"
    with open(plan_path, "r", encoding="utf-8") as f:
        text = f.read()
    status_match = HEADER_STATUS_RE.search(text)
    date_match = HEADER_DATE_RE.search(text)
    status = status_match.group(1).strip() if status_match else "unknown"
    date = date_match.group(1).strip() if date_match else "0000-00-00"
    return status, date


def count_checklist_progress(checklist_path: str | None) -> tuple[int, int] | None:
    if not checklist_path or not os.path.isfile(checklist_path):
        return None
    checked = 0
    total = 0
    with open(checklist_path, "r", encoding="utf-8") as f:
        for line in f:
            stripped = line.strip()
            if stripped.startswith("- [x]") or stripped.startswith("- [X]"):
                checked += 1
                total += 1
            elif stripped.startswith("- [ ]"):
                total += 1
    return (checked, total) if total else None


def candidate_from_context(context_path: str) -> dict | None:
    data = read_manifest(context_path)
    if not data:
        return None
    spec = data.get("spec")
    metadata = data.get("metadata")
    if not isinstance(spec, dict) or not isinstance(metadata, dict):
        return None
    targets = spec.get("targets")
    if not isinstance(targets, dict):
        return None

    plan_dir = os.path.dirname(context_path)
    target_names = sorted(targets.keys())
    default_target = spec.get("defaultTarget")
    target = targets.get(default_target) if isinstance(default_target, str) else None
    if not isinstance(target, dict) and target_names:
        target = targets[target_names[0]]

    plan_path = None
    checklist_path = None
    if isinstance(target, dict) and isinstance(target.get("plan"), str):
        plan_path = os.path.normpath(os.path.join(plan_dir, target["plan"]))
    if isinstance(target, dict) and isinstance(target.get("checklist"), str):
        checklist_path = os.path.normpath(os.path.join(plan_dir, target["checklist"]))
    status, date = plan_status_and_date(plan_path)
    progress = count_checklist_progress(checklist_path)

    name = metadata.get("name") or os.path.basename(plan_dir)
    subject = os.path.basename(os.path.dirname(os.path.dirname(plan_dir)))
    display = f"{subject}/{name}"

    return {
        "display": display,
        "context": context_path,
        "status": status,
        "date": date,
        "targets": target_names,
        "progress": progress,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="List candidate plans for /implement")
    parser.add_argument("--plan-root", default="docs", help="Search root: docs, docs/spec, repo root, or spec-centric plan dir")
    parser.add_argument("--max-candidates", type=int, default=5)
    args = parser.parse_args()

    if args.max_candidates <= 0:
        print("ERROR: --max-candidates must be > 0", file=sys.stderr)
        sys.exit(2)

    candidates = [c for c in (candidate_from_context(p) for p in find_contexts(args.plan_root)) if c]
    candidates = [c for c in candidates if c["status"] in {"active", "draft"}]
    if not candidates:
        print(
            "No active/draft plan contexts found.\n"
            "Create docs/spec/<subspec>/plans/<plan>/context.yaml using the spec-centric v2 template."
        )
        sys.exit(0)

    candidates.sort(key=lambda c: (STATUS_ORDER.get(c["status"], 9), c["date"], c["display"]), reverse=False)
    recommended = candidates[: args.max_candidates]

    print("Available plans for /implement (latest recommendations):\n")
    for i, c in enumerate(recommended, 1):
        reasons = [f"Status: {c['status']}", f"Updated: {c['date']}", f"Targets: {', '.join(c['targets'])}"]
        if c["progress"]:
            done, total = c["progress"]
            pct = int(done / total * 100) if total else 0
            reasons.insert(2, f"Progress: {done}/{total} ({pct}%)")
        print(f"  {i}. {c['display']}")
        print(f"     Context: {c['context']}")
        print(f"     {' | '.join(reasons)}")
        print()

    print(f"Summary: recommended={len(recommended)} | candidates={len(candidates)} | contexts={len(find_contexts(args.plan_root))}")
    print(f"Enter a number (1-{len(recommended)}) to select a plan.")


if __name__ == "__main__":
    main()
