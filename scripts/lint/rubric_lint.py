#!/usr/bin/env python3
"""scripts/lint/rubric_lint.py - F3 rubric registry linter.

Validates `config/rubrics/<feature_key>/<version>[.<language>].yaml`
against the schema and dimension allowlist fixed by
`config/rubrics/README.md` and `docs/spec/prompt-rubric-registry/spec.md` v2.1.

Run: `python3 scripts/lint/rubric_lint.py [--rubrics-dir DIR]`
Exit: 0 on success, 1 on any violation.
"""
from __future__ import annotations

import argparse
import pathlib
import re
import sys

import yaml

SEMVER_RE = re.compile(r"^v\d+\.\d+\.\d+(-[A-Za-z0-9\.-]+)?(\+[A-Za-z0-9\.-]+)?$")
LANGUAGE_RE = re.compile(r"^multi$|^[a-z]{2,3}$")
WEIGHT_TOLERANCE = 0.001
MIN_SCORE_LEVELS = 3

# F1/F3 quality metrics + business-domain dimension allowlist (mirrors
# config/rubrics/README.md §3). Adding a dimension requires updating both
# this constant and the README in the same change.
DIMENSION_ALLOWLIST = frozenset(
    {
        # F1/F3 quality metrics
        "followup_relevance",
        "report_specificity",
        "score_outlier",
        "language_consistency",
        # Practice family
        "practice_depth",
        "practice_dimension_coverage",
        "practice_signal_strength",
        "practice_clarity",
        # Report family
        "report_evidence",
        "report_action_quality",
        "report_calibration",
        # Resume family
        "resume_match",
        "resume_impact",
        "resume_clarity",
        "resume_truthfulness",
        # Target family
        "target_extraction_completeness",
        "target_field_accuracy",
        # JD Match family
        "relevance_to_profile",
        "risk_clarity",
        "actionability",
        "query_alignment",
        "diversity",
        "privacy_compliance",
        # Debrief family
        "debrief_recall_completeness",
        "debrief_lesson_specificity",
        "debrief_action_quality",
        # JD-Match family
        "relevance_to_profile",
        "risk_clarity",
        "actionability",
        "query_alignment",
        "diversity",
        "privacy_compliance",
    }
)

RETIRED_MODULE_RE = re.compile(r"\bmistakes\b|\bgrowth\b|\bdrill\b|mistake\.extract")


def _filename_language(yaml_path: pathlib.Path) -> str:
    parts = yaml_path.name.split(".")
    if len(parts) == 4:
        return "multi"
    if len(parts) == 5:
        return parts[3]
    return ""


def lint_rubric_yaml(yaml_path: pathlib.Path) -> list[str]:
    errors: list[str] = []
    text = yaml_path.read_text(encoding="utf-8")
    try:
        parsed = yaml.safe_load(text)
    except yaml.YAMLError as exc:
        return [f"{yaml_path}: yaml parse error: {exc}"]
    if not isinstance(parsed, dict):
        return [f"{yaml_path}: not a YAML mapping"]

    feature_key = parsed.get("feature_key")
    if feature_key != yaml_path.parent.name:
        errors.append(
            f"{yaml_path}: feature_key '{feature_key}' does not match parent dir "
            f"'{yaml_path.parent.name}'"
        )

    version = parsed.get("version")
    if not isinstance(version, str) or not SEMVER_RE.match(version):
        errors.append(f"{yaml_path}: version '{version}' is not a valid SemVer literal")

    language = parsed.get("language")
    if not isinstance(language, str) or not LANGUAGE_RE.match(language):
        errors.append(f"{yaml_path}: language '{language}' violates language rule")
    else:
        filename_lang = _filename_language(yaml_path)
        if filename_lang and language != filename_lang:
            errors.append(
                f"{yaml_path}: yaml language '{language}' does not match filename '{filename_lang}'"
            )

    dimensions = parsed.get("dimensions")
    if not isinstance(dimensions, list) or not dimensions:
        errors.append(f"{yaml_path}: dimensions must be a non-empty list")
        return errors

    weight_total = 0.0
    for idx, dim in enumerate(dimensions):
        prefix = f"{yaml_path}: dimensions[{idx}]"
        if not isinstance(dim, dict):
            errors.append(f"{prefix} must be a mapping")
            continue
        name = dim.get("name")
        if not isinstance(name, str) or not name:
            errors.append(f"{prefix}: name is required")
        elif name not in DIMENSION_ALLOWLIST:
            errors.append(
                f"{prefix}: name '{name}' not in allowlist (see config/rubrics/README.md §3)"
            )

        weight = dim.get("weight")
        if not isinstance(weight, (int, float)) or weight < 0:
            errors.append(f"{prefix}: weight must be a non-negative number, got {weight!r}")
        else:
            weight_total += float(weight)

        score_levels = dim.get("score_levels")
        if not isinstance(score_levels, list) or len(score_levels) < MIN_SCORE_LEVELS:
            errors.append(
                f"{prefix}: score_levels must be a list of at least {MIN_SCORE_LEVELS} entries"
            )
        else:
            for sidx, sl in enumerate(score_levels):
                sprefix = f"{prefix}.score_levels[{sidx}]"
                if not isinstance(sl, dict):
                    errors.append(f"{sprefix} must be a mapping")
                    continue
                for required_key in ("label", "threshold", "description"):
                    if required_key not in sl:
                        errors.append(f"{sprefix}: missing key '{required_key}'")

    if abs(weight_total - 1.0) > WEIGHT_TOLERANCE:
        errors.append(
            f"{yaml_path}: weight sum {weight_total} not within {WEIGHT_TOLERANCE} of 1.0"
        )

    if RETIRED_MODULE_RE.search(text):
        errors.append(f"{yaml_path}: retired-module name detected (mistakes/growth/drill/mistake.extract)")

    return errors


def lint_rubrics_directory(root: pathlib.Path) -> list[str]:
    if not root.exists():
        return [f"{root}: rubrics directory missing"]
    errors: list[str] = []
    for yp in sorted(p for p in root.rglob("*.yaml") if p.is_file()):
        errors.extend(lint_rubric_yaml(yp))
    return errors


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--rubrics-dir", default="config/rubrics")
    args = parser.parse_args(argv)

    root = pathlib.Path(args.rubrics_dir)
    errors = lint_rubrics_directory(root)

    if errors:
        for e in errors:
            print(f"rubric_lint: {e}", file=sys.stderr)
        return 1
    yaml_count = sum(1 for _ in root.rglob("*.yaml"))
    print(f"rubric_lint: {yaml_count} files clean")
    return 0


if __name__ == "__main__":
    sys.exit(main())
