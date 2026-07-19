#!/usr/bin/env python3
"""Generate or reconcile context.yaml files for spec-centric plan directories.

Current scaffold:
    docs/spec/<subspec>/plans/<NNN-plan>/context.yaml

Usage:
    python3 generate_context_yaml.py --plan-root docs [--docs-root docs] [--dry-run] [--write]
    python3 generate_context_yaml.py --plan-root docs/spec/<subspec>/plans/<NNN-plan> --write

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
    sys.exit(1)


SEQUENCE_PATTERN = re.compile(r"^(\d{3})-(.+)$")
REQUIRED_API_VERSION = "plancontext.agent.dev/v1alpha1"
ALLOWED_TARGET_FIELDS = (
    "plan",
    "checklist",
    "spec",
    "testPlan",
    "testChecklist",
    "bddPlan",
    "bddChecklist",
)


def normalize_target_config(config: dict) -> dict:
    """Return config for the exact minimal manifest contract."""
    config["metadata"] = {"name": config.get("metadata", {}).get("name")}
    config["metadata"] = {k: v for k, v in config["metadata"].items() if v}
    allowed = set(ALLOWED_TARGET_FIELDS)
    for target in config.get("targets", {}).values():
        for field_name in set(target) - allowed:
            target.pop(field_name)
    return config


def reconcile_existing_targets(config: dict, context_path: str) -> dict:
    """Retain target identity while reconciling links from current files."""
    if not os.path.isfile(context_path):
        return config
    with open(context_path, "r", encoding="utf-8") as f:
        existing = yaml.safe_load(f)
    if not isinstance(existing, dict):
        return config
    existing_spec = existing.get("spec")
    if not isinstance(existing_spec, dict):
        return config
    existing_targets = existing_spec.get("targets")
    if not isinstance(existing_targets, dict) or not existing_targets:
        return config

    scanned_targets = config.get("targets", {})
    targets = {}
    for target_name, target in existing_targets.items():
        if not isinstance(target_name, str) or not isinstance(target, dict):
            continue
        sanitized = {
            field_name: target[field_name]
            for field_name in ALLOWED_TARGET_FIELDS
            if isinstance(target.get(field_name), str) and target[field_name]
        }
        if "plan" in sanitized and "checklist" in sanitized:
            matching_scan = next(
                (
                    scanned
                    for scanned in scanned_targets.values()
                    if scanned.get("plan") == sanitized["plan"]
                    and scanned.get("checklist") == sanitized["checklist"]
                ),
                None,
            )
            if matching_scan:
                for field_name in (
                    "testPlan",
                    "testChecklist",
                    "bddPlan",
                    "bddChecklist",
                ):
                    sanitized.pop(field_name, None)
                    if matching_scan.get(field_name):
                        sanitized[field_name] = matching_scan[field_name]
                if "spec" not in sanitized and matching_scan.get("spec"):
                    sanitized["spec"] = matching_scan["spec"]
            targets[target_name] = sanitized
    if not targets:
        return config

    default_target = existing_spec.get("defaultTarget")
    if not isinstance(default_target, str) or default_target not in targets:
        default_target = sorted(targets)[0]
    return {
        "metadata": config.get("metadata", {}),
        "defaultTarget": default_target,
        "targets": targets,
    }


def infer_metadata(dir_name: str) -> dict:
    """Infer the only retained metadata field from the plan directory."""
    return {"name": dir_name}


def infer_target_name(dir_name: str) -> str:
    """Infer target name from NNN-kebab plan directory."""
    match = SEQUENCE_PATTERN.match(dir_name)
    stem = match.group(2) if match else dir_name
    known = {
        "frontend",
        "backend",
        "mock-contract",
        "integration",
        "unit-test",
        "api-contract",
        "foundation",
    }
    return stem if stem in known else "default"


def find_spec_reference(plan_dir_path: str, spec_dir: str, dir_name: str) -> str | None:
    """Return the owning spec.md reference for a spec-centric plan."""
    owning_spec = os.path.abspath(os.path.join(plan_dir_path, "../../spec.md"))
    if os.path.isfile(owning_spec):
        return "../../spec.md"
    return None


def has_inline_progress(plan_dir_path: str) -> bool:
    """Return whether plan.md explicitly owns checkbox progress."""
    plan_path = os.path.join(plan_dir_path, "plan.md")
    try:
        with open(plan_path, "r", encoding="utf-8") as f:
            return any(
                line.lstrip().startswith(("- [ ]", "- [x]", "- [X]"))
                for line in f
            )
    except OSError:
        return False


def scan_directory_targets(
    plan_dir_path: str, dir_name: str, spec_dir: str, docs_root: str
) -> dict | None:
    """Scan one plan directory and determine context targets."""
    if not os.path.isdir(plan_dir_path):
        return None

    files = set(os.listdir(plan_dir_path))
    targets: dict[str, dict] = {}
    spec_ref = find_spec_reference(plan_dir_path, spec_dir, dir_name)

    if "plan.md" in files and (
        "checklist.md" in files or has_inline_progress(plan_dir_path)
    ):
        target_name = infer_target_name(dir_name)
        target = {
            "plan": "./plan.md",
            "checklist": "./checklist.md" if "checklist.md" in files else "./plan.md",
        }
        if spec_ref:
            target["spec"] = spec_ref
        if "test-plan.md" in files:
            target["testPlan"] = "./test-plan.md"
        if "test-checklist.md" in files:
            target["testChecklist"] = "./test-checklist.md"
        if "bdd-plan.md" in files:
            target["bddPlan"] = "./bdd-plan.md"
        if "bdd-checklist.md" in files:
            target["bddChecklist"] = "./bdd-checklist.md"
        targets[target_name] = target
    if not targets:
        return None

    default_target = "backend" if "backend" in targets else sorted(targets.keys())[0]
    return {
        "metadata": infer_metadata(dir_name),
        "defaultTarget": default_target,
        "targets": targets,
    }


def format_yaml(dir_name: str, config: dict) -> str:
    """Format context.yaml content with deterministic key order."""
    metadata_block = {"name": dir_name}
    spec_block = {"defaultTarget": config["defaultTarget"]}

    ordered_targets = {}
    for target_name in sorted(config["targets"].keys()):
        target = config["targets"][target_name]
        target_block = {
            "plan": target["plan"],
            "checklist": target["checklist"],
        }
        for key in (
            "spec",
            "testPlan",
            "testChecklist",
            "bddPlan",
            "bddChecklist",
        ):
            value = target.get(key)
            if value:
                target_block[key] = value
        ordered_targets[target_name] = target_block
    spec_block["targets"] = ordered_targets

    payload = {
        "apiVersion": REQUIRED_API_VERSION,
        "kind": "PlanContext",
        "metadata": metadata_block,
        "spec": spec_block,
    }
    return yaml.safe_dump(payload, sort_keys=False, allow_unicode=True)


def discover_plan_dirs(plan_root: str, docs_root: str) -> list[str]:
    """Find spec-centric plan directories."""
    abs_root = os.path.abspath(plan_root)
    dirs: list[str] = []

    if os.path.isfile(os.path.join(abs_root, "plan.md")):
        return [abs_root]

    roots: list[str]
    if os.path.basename(abs_root) == "docs":
        roots = [os.path.join(abs_root, "spec")]
    elif os.path.basename(abs_root) == "spec":
        roots = [abs_root]
    else:
        roots = [abs_root]

    for root in roots:
        if not os.path.isdir(root):
            continue
        for dirpath, _, files in os.walk(root):
            parts = os.path.normpath(dirpath).split(os.sep)
            if "plans" not in parts:
                continue
            if "plan.md" in files:
                dirs.append(dirpath)

    return sorted(set(dirs))


def infer_docs_root(plan_root: str, docs_root: str | None) -> str:
    if docs_root:
        return os.path.abspath(docs_root)

    cursor = os.path.abspath(plan_root)
    while True:
        if os.path.basename(cursor) == "docs":
            return cursor
        parent = os.path.dirname(cursor)
        if parent == cursor:
            break
        cursor = parent
    return os.path.abspath("docs")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Generate context.yaml for spec-centric plan directories"
    )
    parser.add_argument(
        "--plan-root",
        default="docs",
        help="Search root: docs, docs/spec, a subject dir, or a spec-centric plan dir",
    )
    parser.add_argument("--docs-root", help="Path to docs/ directory")
    mode_group = parser.add_mutually_exclusive_group()
    mode_group.add_argument("--dry-run", action="store_true", help="Print generated YAML")
    mode_group.add_argument("--write", action="store_true", help="Write context.yaml files")
    args = parser.parse_args()

    docs_root = infer_docs_root(args.plan_root, args.docs_root)
    spec_dir = os.path.join(docs_root, "spec")
    plan_dirs = discover_plan_dirs(args.plan_root, docs_root)
    write_mode = args.write

    print(f"Found {len(plan_dirs)} plan directories.")
    print(f"Mode: {'write' if write_mode else 'dry-run'}\n")

    generated = 0
    created = 0
    updated = 0
    reconciled = 0
    skipped_no_targets = 0

    for plan_dir_path in plan_dirs:
        dir_name = os.path.basename(plan_dir_path)
        config = scan_directory_targets(plan_dir_path, dir_name, spec_dir, docs_root)
        if not config or not config.get("targets"):
            print(f"SKIP (no targets): {os.path.relpath(plan_dir_path, docs_root)}/")
            skipped_no_targets += 1
            continue
        context_path = os.path.join(plan_dir_path, "context.yaml")
        config = reconcile_existing_targets(config, context_path)
        config = normalize_target_config(config)
        yaml_content = format_yaml(dir_name, config)
        generated += 1

        rel_context = os.path.relpath(context_path, os.getcwd()).replace("\\", "/")
        if write_mode:
            old_content = None
            if os.path.isfile(context_path):
                with open(context_path, "r", encoding="utf-8") as f:
                    old_content = f.read()
            with open(context_path, "w", encoding="utf-8") as f:
                f.write(yaml_content)
            if old_content is None:
                created += 1
                print(f"CREATED: {rel_context}")
            elif old_content == yaml_content:
                reconciled += 1
                print(f"RECONCILED: {rel_context} (unchanged)")
            else:
                updated += 1
                print(f"UPDATED: {rel_context}")
        else:
            print(f"--- {rel_context} ---")
            print(yaml_content)

    if write_mode:
        print(
            f"\nSummary: generated={generated}, created={created}, updated={updated}, "
            f"reconciled={reconciled}, skipped_no_targets={skipped_no_targets}"
        )
    else:
        print(f"\nSummary: generated={generated}, skipped_no_targets={skipped_no_targets}")


if __name__ == "__main__":
    main()
