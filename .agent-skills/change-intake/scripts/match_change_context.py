#!/usr/bin/env python3
"""Rank plan targets for a free-form bugfix/feature-change request."""

from __future__ import annotations

import argparse
import json
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

TOKEN_PATTERN = re.compile(r"[a-z0-9][a-z0-9_./:-]*|[\u4e00-\u9fff]{2,}", re.IGNORECASE)
STATUS_PATTERN = re.compile(r"^\s*>\s*\*\*状态\*\*:\s*([^\n]+)$", re.MULTILINE)
SPLIT_PATTERN = re.compile(r"[/._:-]+")
STATUS_PRIORITY = {
    "active": 0,
    "draft": 1,
    "completed": 2,
    "unknown": 3,
}
HISTORICAL_STATUS_MAP = {
    "实施中": "active",
    "已完成": "completed",
}
FIELD_WEIGHTS = {
    "contextName": (4, 2),
    "displayName": (4, 2),
    "targetName": (3, 1),
    "planFile": (2, 1),
    "specFile": (2, 1),
}
STOP_WORDS = frozenset(
    {
        "a",
        "an",
        "and",
        "as",
        "at",
        "by",
        "for",
        "from",
        "in",
        "into",
        "is",
        "of",
        "on",
        "or",
        "the",
        "to",
        "with",
    }
)
ROUTING_GENERIC_TOKENS = frozenset(
    {
        "context",
        "yaml",
        "spec",
        "discovery",
        "metadata",
        "target",
        "targets",
        "reference",
        "references",
        "name",
        "plan",
        "checklist",
    }
)
SCENARIO_ID_PATTERN = re.compile(r"\b(?:e2e\.)?(p\d+\.\d+)\b", re.IGNORECASE)
SCENARIO_OWNER_PATTERN = re.compile(
    r"^\s*>\s*Owner:\s*\[[^\]]+\]\(([^)]+)\)",
    re.IGNORECASE | re.MULTILINE,
)
SCENARIO_OWNER_WEIGHT = 100
STATUS_MATCH_BONUS = {
    "active": 3,
    "draft": 1,
}


def token_variants(token: str) -> set[str]:
    """Return low-risk lexical variants for one normalized token."""
    if not token or token in STOP_WORDS:
        return set()

    variants = {token}
    if not token.isascii() or not token.isalpha():
        return variants

    if len(token) > 4 and token.endswith("ing"):
        stem = token[:-3]
        variants.update({stem, f"{stem}e"})
    if len(token) > 3 and token.endswith("ed"):
        stem = token[:-2]
        variants.update({stem, f"{stem}e", token[:-1]})
    if len(token) > 3 and token.endswith("s"):
        variants.add(token[:-1])
    return variants - STOP_WORDS


def tokenize(text: str) -> set[str]:
    """Tokenize identifiers and phrases while dropping low-information words."""
    if not text:
        return set()
    tokens = set()
    for raw in TOKEN_PATTERN.findall(text):
        normalized = raw.lower()
        tokens.update(token_variants(normalized))
        for part in SPLIT_PATTERN.split(normalized):
            if part:
                tokens.update(token_variants(part))
    return tokens


def find_repo_root(plan_root: str) -> str | None:
    """Find the nearest repository root that owns docs and scenario assets."""
    current = os.path.abspath(plan_root)
    while True:
        if os.path.isdir(os.path.join(current, "docs", "spec")) and os.path.isdir(
            os.path.join(current, "test", "scenarios")
        ):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return None
        current = parent


def find_scenario_owner_paths(plan_root: str, query_text: str) -> dict[str, set[str]]:
    """Resolve exact scenario IDs in a query to README-declared owner plans."""
    query_ids = {match.upper() for match in SCENARIO_ID_PATTERN.findall(query_text)}
    repo_root = find_repo_root(plan_root)
    if not query_ids or not repo_root:
        return {}

    owners: dict[str, set[str]] = {}
    scenario_root = os.path.join(repo_root, "test", "scenarios")
    for dirpath, _, files in os.walk(scenario_root):
        if "README.md" not in files:
            continue

        readme_path = os.path.join(dirpath, "README.md")
        with open(readme_path, "r", encoding="utf-8") as f:
            text = f.read()

        readme_ids = {
            match.upper() for match in SCENARIO_ID_PATTERN.findall(text)
        }
        matched_ids = query_ids & readme_ids
        if not matched_ids:
            continue

        owner_match = SCENARIO_OWNER_PATTERN.search(text)
        if not owner_match:
            continue
        owner_ref = owner_match.group(1).strip().strip("<>").split("#", 1)[0]
        if "://" in owner_ref:
            continue

        owner_path = os.path.realpath(os.path.join(dirpath, owner_ref))
        if not os.path.isfile(owner_path):
            continue
        owners.setdefault(owner_path, set()).update(matched_ids)

    return owners


def normalize_status(raw_status: str) -> str:
    """Normalize a markdown Header status value."""
    if not raw_status:
        return "unknown"
    value = raw_status.strip()
    value = HISTORICAL_STATUS_MAP.get(value, value).lower()
    return value if value in STATUS_PRIORITY else "unknown"


def read_plan_status(plan_path: str) -> str:
    """Extract Header 状态 from a markdown plan file."""
    if not os.path.isfile(plan_path):
        return "unknown"

    with open(plan_path, "r", encoding="utf-8") as f:
        text = f.read()

    match = STATUS_PATTERN.search(text)
    if not match:
        return "unknown"
    return normalize_status(match.group(1))


def make_abs(plan_dir: str, rel_path: str | None) -> str | None:
    """Resolve a plan-relative file path to an absolute path."""
    if not rel_path:
        return None
    return os.path.normpath(os.path.join(plan_dir, rel_path))


def find_plan_dirs(plan_root: str) -> list[str]:
    """Find spec-centric plan directories that expose the required manifest."""
    abs_plan_root = os.path.abspath(plan_root)
    roots = []
    if os.path.basename(abs_plan_root) == "plan":
        docs_root = os.path.dirname(abs_plan_root)
        roots.append(os.path.join(docs_root, "spec"))
    elif os.path.basename(abs_plan_root) == "docs":
        roots.append(os.path.join(abs_plan_root, "spec"))
    else:
        roots.extend([os.path.join(abs_plan_root, "docs", "spec"), abs_plan_root])

    plan_dirs = []
    for root in roots:
        if not os.path.isdir(root):
            continue
        if os.path.isfile(os.path.join(root, "context.yaml")):
            plan_dirs.append(root)
            continue
        for dirpath, _, files in os.walk(root):
            if "context.yaml" not in files:
                continue
            parts = os.path.normpath(dirpath).split(os.sep)
            if "plans" in parts:
                plan_dirs.append(dirpath)
    return sorted(set(plan_dirs))


def iter_context_targets(plan_root: str):
    """Yield one record per target from required minimal link manifests."""
    for plan_dir in find_plan_dirs(plan_root):
        context_path = os.path.join(plan_dir, "context.yaml")
        entry = os.path.basename(plan_dir)
        try:
            with open(context_path, "r", encoding="utf-8") as f:
                data = yaml.safe_load(f)
        except (OSError, yaml.YAMLError):
            continue
        spec = data.get("spec") if isinstance(data, dict) else None
        metadata = data.get("metadata") if isinstance(data, dict) else None
        targets = spec.get("targets") if isinstance(spec, dict) else None
        if not isinstance(targets, dict) or not targets:
            continue

        context_name = metadata.get("name", entry) if isinstance(metadata, dict) else entry
        subject = os.path.basename(os.path.dirname(os.path.dirname(plan_dir)))
        display_name = f"{subject}/{context_name}"
        for target_name, target in sorted(targets.items()):
            if not isinstance(target, dict):
                continue

            plan_path = make_abs(plan_dir, target.get("plan"))
            spec_path = make_abs(plan_dir, target.get("spec"))
            status = read_plan_status(plan_path) if plan_path else "unknown"

            yield {
                "contextName": context_name,
                "displayName": display_name,
                "contextPath": context_path,
                "planDir": plan_dir,
                "target": target_name,
                "status": status,
                "files": {
                    "plan": plan_path,
                    "checklist": make_abs(plan_dir, target.get("checklist")),
                    "spec": spec_path,
                    "testPlan": make_abs(plan_dir, target.get("testPlan")),
                    "testChecklist": make_abs(plan_dir, target.get("testChecklist")),
                },
            }


def score_values(
    query_text: str,
    query_tokens: set[str],
    field_name: str,
    values: list[str],
) -> tuple[int, list[str]]:
    """Score one path-derived candidate field against the query."""
    exact_weight, partial_weight = FIELD_WEIGHTS[field_name]
    reasons = []
    seen = set()
    overlap_tokens = set()
    has_exact_hit = False

    for value in values:
        if not isinstance(value, str) or not value.strip():
            continue

        normalized = value.strip().lower()
        value_tokens = tokenize(normalized)
        exact_hit = len(normalized) >= 2 and normalized in query_text
        overlap = (query_tokens & value_tokens) - ROUTING_GENERIC_TOKENS
        if not overlap:
            continue
        new_overlap = overlap - overlap_tokens
        if not exact_hit and not new_overlap:
            continue

        has_exact_hit = has_exact_hit or exact_hit
        overlap_tokens.update(overlap)
        reason = f"{field_name}={value}"
        if reason not in seen:
            reasons.append(reason)
            seen.add(reason)

    score = partial_weight * len(overlap_tokens)
    if has_exact_hit:
        score = max(score, exact_weight)
    return score, reasons


def score_candidate(query_text: str, query_tokens: set[str], candidate: dict) -> tuple[int, list[str]]:
    """Score one target candidate."""
    score = 0
    reasons = []

    fallback_values = {
        "contextName": [candidate["contextName"]],
        "displayName": [candidate.get("displayName", candidate["contextName"])],
        "targetName": [candidate["target"]],
        "planFile": [],
        "specFile": [],
    }

    plan_path = candidate["files"].get("plan")
    if plan_path:
        fallback_values["planFile"].append(os.path.splitext(os.path.basename(plan_path))[0])

    spec_path = candidate["files"].get("spec")
    if spec_path:
        fallback_values["specFile"].append(os.path.splitext(os.path.basename(spec_path))[0])

    for field_name, values in fallback_values.items():
        field_score, field_reasons = score_values(
            query_text, query_tokens, field_name, values
        )
        score += field_score
        reasons.extend(field_reasons)

    # Status remains a bounded tie-breaker; exact owner evidence is applied later.
    if score > 0:
        score += STATUS_MATCH_BONUS.get(candidate["status"], 0)

    deduped_reasons = []
    seen = set()
    for reason in reasons:
        if reason in seen:
            continue
        deduped_reasons.append(reason)
        seen.add(reason)

    return score, deduped_reasons


def assess_confidence(candidates: list[dict]) -> str:
    """Return high / medium / low / none based on top score separation."""
    if not candidates or candidates[0]["score"] <= 0:
        return "none"

    top_score = candidates[0]["score"]
    second_score = candidates[1]["score"] if len(candidates) > 1 else 0
    delta = top_score - second_score

    if top_score >= 12 and delta >= 3:
        return "high"
    if top_score >= 7 and delta >= 2:
        return "medium"
    return "low"


def match_change_contexts(plan_root: str, query: str, limit: int = 3) -> dict:
    """Score all plan targets and return the best candidates."""
    normalized_query = query.strip().lower()
    query_tokens = tokenize(normalized_query)
    scenario_owners = find_scenario_owner_paths(plan_root, normalized_query)
    scored = []

    for candidate in iter_context_targets(plan_root):
        score, reasons = score_candidate(normalized_query, query_tokens, candidate)
        plan_path = candidate["files"].get("plan")
        owner_scenarios = (
            scenario_owners.get(os.path.realpath(plan_path), set()) if plan_path else set()
        )
        if owner_scenarios:
            score += SCENARIO_OWNER_WEIGHT
            reasons = [
                *(f"scenarioOwner=E2E.{scenario_id}" for scenario_id in sorted(owner_scenarios)),
                *reasons,
            ]
        if score <= 0:
            continue

        scored.append(
            {
                "plan": os.path.basename(candidate["planDir"]),
                "displayPlan": candidate.get("displayName", os.path.basename(candidate["planDir"])),
                "target": candidate["target"],
                "status": candidate["status"],
                "score": score,
                "reviseInPlace": candidate["status"] == "completed",
                "contextPath": candidate["contextPath"],
                "files": candidate["files"],
                "reasons": reasons,
            }
        )

    scored.sort(
        key=lambda item: (
            -item["score"],
            STATUS_PRIORITY.get(item["status"], STATUS_PRIORITY["unknown"]),
            item["plan"],
            item.get("displayPlan", ""),
            item["target"],
        )
    )
    limited = scored[:limit]
    confidence = assess_confidence(limited)
    recommended = limited[0] if limited else None

    return {
        "query": query,
        "confidence": confidence,
        "recommended": recommended,
        "candidates": limited,
    }


def main():
    parser = argparse.ArgumentParser(description="Match a free-form issue to plan targets")
    parser.add_argument("--plan-root", default="docs", help="Search root: docs, docs/spec, repo root, or spec-centric plan dir")
    parser.add_argument("--query", required=True, help="Free-form bug/change description")
    parser.add_argument("--limit", type=int, default=3, help="Max candidates to return")
    args = parser.parse_args()

    result = match_change_contexts(
        plan_root=args.plan_root,
        query=args.query,
        limit=max(args.limit, 1),
    )
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
