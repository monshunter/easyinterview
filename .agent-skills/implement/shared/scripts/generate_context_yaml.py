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
import copy
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


LINK_PATTERN = re.compile(r"\[[^\]]+\]\(([^)]+)\)")
SEQUENCE_PATTERN = re.compile(r"^(\d{3})-(.+)$")
REQUIRED_API_VERSION = "plancontext.agent.dev/v1alpha1"


def normalize_target_config(config: dict) -> dict:
    """Deduplicate first-class references while preserving target metadata."""
    targets = config.get("targets", {})
    for target in targets.values():
        refs = target.get("references")
        if not refs:
            continue

        first_class_paths = {
            target.get("plan", ""),
            target.get("checklist", ""),
            target.get("spec", ""),
            target.get("testPlan", ""),
            target.get("testChecklist", ""),
            target.get("bddPlan", ""),
            target.get("bddChecklist", ""),
        }
        first_class_paths.discard("")

        seen = set()
        normalized = []
        for ref_path in refs:
            if not ref_path or ref_path in seen or ref_path in first_class_paths:
                continue
            seen.add(ref_path)
            normalized.append(ref_path)

        if normalized:
            target["references"] = sorted(normalized)
        else:
            target.pop("references", None)

    return config


def load_existing_manifest(context_path: str) -> dict | None:
    """Load an existing context manifest for preservation-only merge."""
    if not os.path.isfile(context_path):
        return None
    with open(context_path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)
    return data if isinstance(data, dict) else None


def merge_preserved_discovery(config: dict, existing_manifest: dict | None) -> dict:
    """Preserve manually maintained discovery and branch metadata."""
    if not existing_manifest:
        return config

    existing_metadata = existing_manifest.get("metadata")
    if isinstance(existing_metadata, dict):
        metadata = config.setdefault("metadata", {})
        for field_name in ("baseBranch", "branch"):
            value = existing_metadata.get(field_name)
            if isinstance(value, str):
                metadata[field_name] = value
        for field_name in ("supersedes", "specVersion"):
            value = existing_metadata.get(field_name)
            if value is not None:
                metadata[field_name] = copy.deepcopy(value)

    existing_spec = existing_manifest.get("spec")
    if not isinstance(existing_spec, dict):
        return config

    existing_discovery = existing_spec.get("discovery")
    if isinstance(existing_discovery, dict):
        config["discovery"] = copy.deepcopy(existing_discovery)

    existing_targets = existing_spec.get("targets")
    if not isinstance(existing_targets, dict):
        return config

    for target_name, target in config.get("targets", {}).items():
        existing_target = existing_targets.get(target_name)
        if not isinstance(existing_target, dict):
            continue
        target_discovery = existing_target.get("discovery")
        if isinstance(target_discovery, dict):
            preserved = copy.deepcopy(target_discovery)
            preserved.pop("commands", None)
            target["discovery"] = preserved

    return config


def to_plan_relative_path(plan_dir_path: str, abs_path: str) -> str:
    """Convert absolute path to normalized relative path from plan dir."""
    rel_path = os.path.relpath(abs_path, plan_dir_path).replace("\\", "/")
    if not rel_path.startswith("."):
        rel_path = f"./{rel_path}"
    return rel_path


def collect_association_links(
    plan_dir_path: str,
    docs_root: str,
    doc_rel_paths: list[str],
    excluded_rel_paths: set[str],
) -> list[str]:
    """Collect markdown references from association lines in plan/checklist docs."""
    refs = []
    abs_docs_root = os.path.abspath(docs_root)
    for doc_rel in doc_rel_paths:
        doc_abs = os.path.abspath(os.path.join(plan_dir_path, doc_rel))
        if not os.path.isfile(doc_abs):
            continue
        with open(doc_abs, "r", encoding="utf-8") as f:
            lines = f.readlines()

        for line in lines:
            stripped = line.strip()
            if "关联" not in line and "相关文档导航" not in line and not stripped.startswith("> - "):
                continue
            for link in LINK_PATTERN.findall(line):
                if link.startswith(("http://", "https://", "#")):
                    continue
                resolved = os.path.abspath(os.path.join(os.path.dirname(doc_abs), link))
                if not resolved.endswith(".md"):
                    continue
                if not (resolved == abs_docs_root or resolved.startswith(abs_docs_root + os.sep)):
                    continue
                if not os.path.isfile(resolved):
                    continue
                rel_path = to_plan_relative_path(plan_dir_path, resolved)
                if rel_path not in excluded_rel_paths:
                    refs.append(rel_path)
    return refs


def infer_subspec_and_sequence(plan_dir_path: str, dir_name: str) -> dict:
    """Infer spec-centric metadata from a plan directory path."""
    parts = os.path.normpath(plan_dir_path).split(os.sep)
    subspec = None
    if "plans" in parts:
        plan_index = len(parts) - 1 - list(reversed(parts)).index("plans")
        if plan_index > 0:
            subspec = parts[plan_index - 1]

    sequence = None
    match = SEQUENCE_PATTERN.match(dir_name)
    if match:
        sequence = int(match.group(1))

    metadata = {
        "subspec": subspec,
        "name": dir_name,
        "sequence": sequence,
        "supersedes": [],
        "specVersion": {"from": None, "to": 1.0},
    }
    return {k: v for k, v in metadata.items() if v is not None}


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


def scan_directory_targets(
    plan_dir_path: str, dir_name: str, spec_dir: str, docs_root: str
) -> dict | None:
    """Scan one plan directory and determine context targets."""
    if not os.path.isdir(plan_dir_path):
        return None

    files = set(os.listdir(plan_dir_path))
    targets: dict[str, dict] = {}
    spec_ref = find_spec_reference(plan_dir_path, spec_dir, dir_name)

    if "plan.md" in files and "checklist.md" in files:
        target_name = infer_target_name(dir_name)
        target = {
            "plan": "./plan.md",
            "checklist": "./checklist.md",
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

    for target in targets.values():
        excluded = {target["plan"], target["checklist"]}
        for key in ("spec", "testPlan", "testChecklist", "bddPlan", "bddChecklist"):
            if key in target:
                excluded.add(target[key])
        refs = collect_association_links(
            plan_dir_path=plan_dir_path,
            docs_root=docs_root,
            doc_rel_paths=[target["plan"], target["checklist"]],
            excluded_rel_paths=excluded,
        )
        if refs:
            target["references"] = refs

    default_target = "backend" if "backend" in targets else sorted(targets.keys())[0]
    metadata = infer_subspec_and_sequence(plan_dir_path, dir_name)
    if metadata.get("subspec"):
        discovery = {
            "aliases": [metadata["subspec"], dir_name],
            "keywords": [],
        }
    else:
        discovery = {"aliases": [dir_name], "keywords": []}

    return {
        "metadata": metadata,
        "defaultTarget": default_target,
        "discovery": discovery,
        "targets": targets,
    }


def format_yaml(dir_name: str, config: dict) -> str:
    """Format context.yaml content with deterministic key order."""
    metadata_block = copy.deepcopy(config.get("metadata") or {})
    metadata_block.setdefault("name", dir_name)

    spec_block = {"defaultTarget": config["defaultTarget"]}
    if isinstance(config.get("discovery"), dict):
        spec_block["discovery"] = config["discovery"]

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
            "references",
            "discovery",
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
        config = normalize_target_config(config)
        context_path = os.path.join(plan_dir_path, "context.yaml")
        config = merge_preserved_discovery(config, load_existing_manifest(context_path))
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
