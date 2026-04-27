#!/usr/bin/env python3
"""sync-doc-index.py — Deterministic checker for document Header / INDEX drift.

Scans spec-centric docs for:
  - Header field violations (missing, wrong order, invalid values)
  - INDEX drift (Header vs INDEX mismatch)
  - Orphans (files not in INDEX, INDEX entries without files)

Usage:
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py               # default: --check
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check       # human-readable audit
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check --json  # JSON audit
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-header    # auto-fix headers
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-header --dry-run
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index     # auto-fix INDEX columns
  python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index --dry-run
"""

import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path

# ── Constants ──────────────────────────────────────────────────────────

VALID_STATUS = {"draft", "active", "completed", "superseded", "deprecated"}
LEGACY_STATUS = {"实施中": "active", "已完成": "completed", "废弃": "deprecated"}
EXEC_MODES = {"parallel", "sequential"}
DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}$")
HEADER_RE = re.compile(r"^>\s*\*\*(.+?)\*\*\s*:\s*(.+)$")
LINK_RE = re.compile(r"\[([^\]]*)\]\(([^)]+)\)")

SPEC_FIELDS = ["版本", "状态", "更新日期"]
PLAN_FIELDS = ["版本", "状态", "更新日期"]
CHECKLIST_FIELDS = ["版本", "状态", "更新日期"]
SKIP_FILES = {"README.md", "TEMPLATES.md", "INDEX.md"}

PLAN_GROUP_STATUS = {
    "进行中": "active",
    "Active": "active",
    "草稿": "draft",
    "Draft": "draft",
    "已完成": "completed",
    "Completed": "completed",
    "已取代": "superseded",
    "Superseded": "superseded",
}


# ── Header Parser ─────────────────────────────────────────────────────


def parse_header(filepath):
    """Extract header fields from a markdown file.

    Returns dict with 'fields' (ordered dict-like list of tuples),
    'standard' (dict of recognized fields), 'order' (list of field names),
    and 'adjacent' (bool: whether header immediately follows the title line).
    """
    fields = []
    in_header = False
    gap_seen = False
    title_seen = False
    lines_after_title = 0  # non-empty lines between title and first header field

    with open(filepath, "r", encoding="utf-8") as f:
        for line in f:
            stripped = line.strip()
            if stripped.startswith("# ") and not in_header and not fields:
                title_seen = True
                continue
            if title_seen and not in_header and not fields:
                # Between title and first header field
                if stripped and not HEADER_RE.match(stripped):
                    lines_after_title += 1
            m = HEADER_RE.match(stripped)
            if m:
                in_header = True
                gap_seen = False
                fields.append((m.group(1).strip(), m.group(2).strip()))
            elif in_header and stripped == "":
                gap_seen = True
            elif in_header and gap_seen and not stripped.startswith(">"):
                break
            elif in_header and not stripped.startswith(">") and stripped != "":
                break

    order = [name for name, _ in fields]
    standard = {name: val for name, val in fields if name in ("版本", "状态", "更新日期", "执行模式")}
    adjacent = lines_after_title == 0
    return {"fields": fields, "standard": standard, "order": order, "adjacent": adjacent}


# ── INDEX Parsers ──────────────────────────────────────────────────────


SEPARATOR_RE = re.compile(r"^\|[\s\-:|]+\|$")


def strip_html_comment_lines(lines):
    """Remove HTML comment blocks so examples do not count as live INDEX rows."""
    filtered = []
    in_comment = False
    for line in lines:
        cursor = line
        while cursor:
            if in_comment:
                end = cursor.find("-->")
                if end == -1:
                    cursor = ""
                    break
                cursor = cursor[end + 3 :]
                in_comment = False
                continue
            start = cursor.find("<!--")
            if start == -1:
                filtered.append(cursor)
                cursor = ""
            else:
                before = cursor[:start]
                if before:
                    filtered.append(before + "\n")
                cursor = cursor[start + 4 :]
                in_comment = True
    return filtered


def parse_spec_index(index_path):
    """Parse spec INDEX.md. Returns list of {name, file, version, status, date, non_standard}."""
    entries = []
    with open(index_path, "r", encoding="utf-8") as f:
        all_lines = strip_html_comment_lines(f.readlines())

    # Build set of table-header line indices (line immediately before a separator row)
    header_lines = set()
    for i, line in enumerate(all_lines):
        if SEPARATOR_RE.match(line.strip()):
            if i > 0:
                header_lines.add(i - 1)

    for i, line in enumerate(all_lines):
        stripped = line.strip()
        if not stripped.startswith("|"):
            continue
        # Skip separator rows and table header rows
        if SEPARATOR_RE.match(stripped) or i in header_lines:
            continue
        cells = [c.strip() for c in stripped.split("|")]
        cells = [c for c in cells if c]  # remove empty from leading/trailing |
        if len(cells) < 4:
            continue

        doc_cell, version, status, date = cells[0], cells[1], cells[2], cells[3]
        links = LINK_RE.findall(doc_cell)
        file_path = links[0][1] if links else None
        non_standard = (
            file_path is None
            or file_path.startswith("../")
            or file_path == ""
        )
        entries.append({
            "name": links[0][0] if links else doc_cell,
            "file": file_path,
            "version": version,
            "status": status,
            "date": date,
            "non_standard": non_standard,
        })
    return entries


def parse_plan_index(index_path):
    """Parse plan INDEX.md. Returns list of entries with status group info."""
    entries = []
    current_status = None
    has_version_col = True  # Superseded group has no version/date columns

    with open(index_path, "r", encoding="utf-8") as f:
        all_lines = strip_html_comment_lines(f.readlines())

    # Build set of table-header line indices (line immediately before a separator row)
    header_lines = set()
    for i, line in enumerate(all_lines):
        if SEPARATOR_RE.match(line.strip()):
            if i > 0:
                header_lines.add(i - 1)

    for i, line in enumerate(all_lines):
        stripped = line.strip()

        # Detect group headings: ## 1 进行中（Active）
        section_match = re.match(r"^##\s+\d+\s+(.+)$", stripped)
        if section_match:
            section_name = section_match.group(1)
            current_status = None
            has_version_col = True
            for keyword, status in PLAN_GROUP_STATUS.items():
                if keyword in section_name:
                    current_status = status
                    if status == "superseded":
                        has_version_col = False
                    break
            continue

        if current_status is None or not stripped.startswith("|"):
            continue
        # Skip separator rows and table header rows
        if SEPARATOR_RE.match(stripped) or i in header_lines:
            continue

        cells = [c.strip() for c in stripped.split("|")]
        cells = [c for c in cells if c]
        if not cells:
            continue

        plan_cell = cells[0]
        files_cell = cells[1] if len(cells) > 1 else ""
        version = cells[2] if len(cells) > 2 and has_version_col else ""
        status_col = cells[3] if len(cells) > 3 and has_version_col else current_status
        date = cells[4] if len(cells) > 4 and has_version_col else (
            cells[3] if len(cells) > 3 and has_version_col else ""
        )
        is_sub = plan_cell.strip().startswith("↳")

        # Extract all linked files
        file_links = LINK_RE.findall(files_cell)
        linked_files = [url for _, url in file_links]

        entries.append({
            "name": plan_cell,
            "is_sub": is_sub,
            "expected_status": current_status,
            "linked_files": linked_files,
            "version": version,
            "status": status_col,
            "date": date,
            "has_version_col": has_version_col,
            "non_standard": not linked_files,
        })
    return entries


# ── Validators ─────────────────────────────────────────────────────────


def is_checklist(filepath):
    name = filepath.name
    return name.endswith("-checklist.md") or name == "checklist.md"


def validate_header(filepath, header, required_fields, root):
    """Validate header fields. Returns list of violation dicts."""
    violations = []
    rel = str(filepath.relative_to(root))
    std = header["standard"]
    order = header["order"]

    # Missing fields
    for f in required_fields:
        if f not in std:
            auto = f in ("更新日期",)
            violations.append({
                "file": rel,
                "type": "missing_field",
                "field": f,
                "auto_fixable": auto,
                "message": f"Missing required field: {f}",
            })

    # Field order
    present_standard = [f for f in order if f in set(required_fields)]
    expected_order = [f for f in required_fields if f in present_standard]
    if present_standard != expected_order:
        violations.append({
            "file": rel,
            "type": "wrong_order",
            "auto_fixable": True,
            "expected": required_fields,
            "actual": present_standard,
            "message": f"Wrong field order: {present_standard} (expected {expected_order})",
        })

    # Status validation
    if "状态" in std:
        s = std["状态"]
        if s in LEGACY_STATUS:
            violations.append({
                "file": rel,
                "type": "legacy_status",
                "field": "状态",
                "current": s,
                "normalized": LEGACY_STATUS[s],
                "auto_fixable": True,
                "message": f"Legacy status '{s}' → '{LEGACY_STATUS[s]}'",
            })
        elif s not in VALID_STATUS:
            violations.append({
                "file": rel,
                "type": "invalid_status",
                "value": s,
                "auto_fixable": False,
                "message": f"Invalid status: {s}",
            })

    # Date validation
    if "更新日期" in std and not DATE_RE.match(std["更新日期"]):
        violations.append({
            "file": rel,
            "type": "invalid_date",
            "value": std["更新日期"],
            "auto_fixable": False,
            "message": f"Invalid date format: {std['更新日期']}",
        })

    # 执行模式 validation
    if "执行模式" in std and std["执行模式"] not in EXEC_MODES:
        violations.append({
            "file": rel,
            "type": "invalid_exec_mode",
            "value": std["执行模式"],
            "auto_fixable": False,
            "message": f"Invalid 执行模式: {std['执行模式']}",
        })

    # No header at all
    if not header["fields"]:
        violations.append({
            "file": rel,
            "type": "no_header",
            "auto_fixable": False,
            "message": "No header block found",
        })

    # Header not adjacent to title (may have captured fields from body)
    if header["fields"] and not header.get("adjacent", True):
        violations.append({
            "file": rel,
            "type": "header_not_adjacent",
            "auto_fixable": False,
            "message": "Header block is not immediately after the title line (may be unreliable)",
        })

    return violations


# ── Drift Checker ──────────────────────────────────────────────────────


def check_spec_drift(spec_dir, root):
    """Check spec INDEX vs document headers. Returns (drifts, warnings)."""
    index_path = spec_dir / "INDEX.md"
    if not index_path.exists():
        return [], [{"entry": str(index_path), "reason": "INDEX.md not found"}]

    entries = parse_spec_index(index_path)
    drifts = []
    warnings = []

    for entry in entries:
        if entry["non_standard"]:
            warnings.append({"entry": entry["name"], "reason": "non-standard entry (external link or no link)"})
            continue

        doc_path = (spec_dir / entry["file"]).resolve()
        if not doc_path.exists():
            warnings.append({"entry": entry["name"], "file": entry["file"], "reason": "linked file does not exist"})
            continue

        header = parse_header(doc_path)
        std = header["standard"]
        rel = str(doc_path.relative_to(root))

        for field, idx_key in [("版本", "version"), ("状态", "status"), ("更新日期", "date")]:
            if field in std and std[field] != entry[idx_key]:
                drifts.append({
                    "file": rel,
                    "field": field,
                    "header_value": std[field],
                    "index_value": entry[idx_key],
                    "auto_fixable": True,
                })

    return drifts, warnings


def check_plan_drift(plan_dir, root):
    """Check plan INDEX vs document headers. Returns (drifts, warnings)."""
    index_path = plan_dir / "INDEX.md"
    if not index_path.exists():
        return [], [{"entry": str(index_path), "reason": "INDEX.md not found"}]

    entries = parse_plan_index(index_path)
    drifts = []
    warnings = []

    for entry in entries:
        if entry["non_standard"]:
            warnings.append({"entry": entry["name"], "reason": "non-standard entry"})
            continue

        # Check all linked files for status group drift
        for linked_file in entry["linked_files"]:
            doc_path = (plan_dir / linked_file).resolve()
            if not doc_path.exists():
                continue

            header = parse_header(doc_path)
            std = header["standard"]
            rel = str(doc_path.relative_to(root))

            # Status group drift
            if "状态" in std and std["状态"] != entry["expected_status"]:
                severity = "advisory" if entry["is_sub"] else "error"
                # Auto-fixable when both source and target sections share the same
                # column shape: active / draft / completed. Superseded uses a different
                # schema (no version/date columns), so leave it for human judgment.
                same_shape = {"active", "draft", "completed"}
                auto_fixable = (
                    not entry["is_sub"]
                    and entry["expected_status"] in same_shape
                    and std["状态"] in same_shape
                )
                drifts.append({
                    "file": rel,
                    "field": "状态(group)",
                    "header_value": std["状态"],
                    "index_value": entry["expected_status"],
                    "auto_fixable": auto_fixable,
                    "severity": severity,
                })
            if "状态" in std and entry.get("status") and std["状态"] != entry["status"]:
                drifts.append({
                    "file": rel,
                    "field": "状态",
                    "header_value": std["状态"],
                    "index_value": entry["status"],
                    "auto_fixable": True,
                })

        # Version/date drift: use first existing linked file
        for linked_file in entry["linked_files"]:
            doc_path = (plan_dir / linked_file).resolve()
            if not doc_path.exists():
                continue

            header = parse_header(doc_path)
            std = header["standard"]
            rel = str(doc_path.relative_to(root))

            # Version drift (skip for groups without version column, e.g. Superseded)
            if entry.get("has_version_col", True) and "版本" in std and entry["version"] and std["版本"] != entry["version"]:
                drifts.append({
                    "file": rel,
                    "field": "版本",
                    "header_value": std["版本"],
                    "index_value": entry["version"],
                    "auto_fixable": True,
                })

            # Date drift (skip for groups without date column)
            if entry.get("has_version_col", True) and "更新日期" in std and entry["date"] and std["更新日期"] != entry["date"]:
                drifts.append({
                    "file": rel,
                    "field": "更新日期",
                    "header_value": std["更新日期"],
                    "index_value": entry["date"],
                    "auto_fixable": True,
                })

            break  # Version/date columns represent the first file

    return drifts, warnings


# ── Orphan Detector ────────────────────────────────────────────────────


def detect_orphans(spec_dir, root):
    """Detect files not in INDEX and INDEX entries pointing to missing files."""
    missing_from_index = []
    dangling = []

    # Spec orphans
    spec_index_path = spec_dir / "INDEX.md"
    if spec_index_path.exists():
        entries = parse_spec_index(spec_index_path)
        indexed_files = {e["file"] for e in entries if e["file"]}

        spec_docs = list(spec_dir.glob("*.md")) + list(spec_dir.glob("*/spec.md"))
        for md in sorted(spec_docs):
            if md.name in SKIP_FILES:
                continue
            rel_to_spec = "./" + str(md.relative_to(spec_dir))
            if rel_to_spec not in indexed_files:
                missing_from_index.append(str(md.relative_to(root)))

        for e in entries:
            if e["file"] and not e["non_standard"]:
                full = (spec_dir / e["file"]).resolve()
                if not full.exists():
                    dangling.append(e["file"])

    # Per-subspec plan orphans
    for plans_dir in sorted(spec_dir.glob("*/plans")):
        plan_index_path = plans_dir / "INDEX.md"
        if not plan_index_path.exists():
            continue

        entries = parse_plan_index(plan_index_path)
        indexed_files = set()
        for entry in entries:
            for linked_file in entry["linked_files"]:
                indexed_files.add(linked_file)

        plan_docs = list(plans_dir.glob("*/*.md"))
        for md in sorted(plan_docs):
            if md.name in SKIP_FILES:
                continue
            if md.name not in {"plan.md", "checklist.md", "bdd-plan.md", "bdd-checklist.md"} and not md.name.endswith("-plan.md") and not md.name.endswith("-checklist.md"):
                continue
            rel_link = "./" + str(md.relative_to(plans_dir))
            if rel_link not in indexed_files:
                missing_from_index.append(str(md.relative_to(root)))

        for entry in entries:
            for linked_file in entry["linked_files"]:
                full = (plans_dir / linked_file).resolve()
                if not full.exists():
                    dangling.append(str((plans_dir / linked_file).relative_to(root)))

    return {"missing_from_index": missing_from_index, "dangling_index_entries": dangling}


# ── Full Check ─────────────────────────────────────────────────────────


def run_check(root):
    """Run full check. Returns report dict."""
    spec_dir = root / "docs" / "spec"

    violations = []
    all_warnings = []

    # Spec headers: docs/spec/*/spec.md and docs/spec/*/history.md.
    spec_docs = list(spec_dir.glob("*/spec.md")) + list(spec_dir.glob("*/history.md"))
    for md in sorted(spec_docs):
        if md.name in SKIP_FILES:
            continue
        header = parse_header(md)
        violations.extend(validate_header(md, header, SPEC_FIELDS, root))

    # Plan headers: docs/spec/*/plans/*/*.md.
    plan_docs = list(spec_dir.glob("*/plans/*/*.md"))
    for md in sorted(plan_docs):
        if md.name in SKIP_FILES:
            continue
        rel = md.relative_to(spec_dir)
        if "plans" not in rel.parts:
            continue
        required = CHECKLIST_FIELDS if is_checklist(md) else PLAN_FIELDS
        header = parse_header(md)
        violations.extend(validate_header(md, header, required, root))

    # INDEX drifts
    spec_drifts, spec_warns = check_spec_drift(spec_dir, root)
    plan_drifts = []
    plan_warns = []
    for plans_dir in sorted(spec_dir.glob("*/plans")):
        drifts_for_dir, warnings_for_dir = check_plan_drift(plans_dir, root)
        plan_drifts.extend(drifts_for_dir)
        plan_warns.extend(warnings_for_dir)
    drifts = spec_drifts + plan_drifts
    all_warnings.extend(spec_warns)
    all_warnings.extend(plan_warns)

    # Orphans
    orphans = detect_orphans(spec_dir, root)

    auto_fixable = sum(1 for v in violations if v.get("auto_fixable")) + sum(1 for d in drifts if d.get("auto_fixable"))
    needs_llm = (
        sum(1 for v in violations if not v.get("auto_fixable"))
        + sum(1 for d in drifts if not d.get("auto_fixable"))
        + len(orphans["missing_from_index"])
        + len(orphans["dangling_index_entries"])
    )

    return {
        "header_violations": violations,
        "index_drifts": drifts,
        "orphans": orphans,
        "warnings": all_warnings,
        "summary": {
            "violations": len(violations),
            "drifts": len(drifts),
            "orphans": len(orphans["missing_from_index"]) + len(orphans["dangling_index_entries"]),
            "warnings": len(all_warnings),
            "auto_fixable": auto_fixable,
            "needs_llm": needs_llm,
        },
    }


# ── Auto-Fixers ────────────────────────────────────────────────────────


def git_last_modified(filepath):
    """Get last modified date from git log."""
    try:
        result = subprocess.run(
            ["git", "log", "--follow", "-1", "--format=%cs", str(filepath)],
            capture_output=True, text=True, timeout=10,
        )
        date = result.stdout.strip()
        if DATE_RE.match(date):
            return date
    except Exception:
        pass
    return None


def fix_header_file(filepath, header, required_fields, dry_run=False):
    """Fix header of a single file. Returns (fixes, skipped) tuple."""
    fixes = []
    skipped = []
    std = header["standard"]

    with open(filepath, "r", encoding="utf-8") as f:
        lines = f.readlines()

    modified = False

    # 1. Normalize legacy status
    for i, line in enumerate(lines):
        m = HEADER_RE.match(line.strip())
        if m and m.group(1).strip() == "状态":
            val = m.group(2).strip()
            if val in LEGACY_STATUS:
                new_val = LEGACY_STATUS[val]
                lines[i] = line.replace(f"**: {val}", f"**: {new_val}")
                fixes.append({"action": "normalize_status", "from": val, "to": new_val})
                modified = True

    # 2. Recover missing 更新日期 from git
    if "更新日期" in required_fields and "更新日期" not in std:
        date = git_last_modified(filepath)
        if date:
            insert_after = None
            for i, line in enumerate(lines):
                m = HEADER_RE.match(line.strip())
                if m and m.group(1).strip() in ("状态", "版本"):
                    insert_after = i
            if insert_after is not None:
                lines.insert(insert_after + 1, f"> **更新日期**: {date}\n")
                fixes.append({"action": "add_field", "field": "更新日期", "value": date, "source": "git_log"})
                modified = True
            else:
                skipped.append({"action": "add_field", "field": "更新日期", "reason": "no 状态/版本 line found as anchor"})
        else:
            skipped.append({"action": "add_field", "field": "更新日期", "reason": "git log returned no date"})

    # 3. Fix field order — always re-parse from current lines for consistency
    # Build a temporary file-like header from current lines (not the stale input header)
    current_fields = []
    for line in lines:
        m = HEADER_RE.match(line.strip())
        if m:
            current_fields.append((m.group(1).strip(), m.group(2).strip()))
    current_order = [name for name, _ in current_fields]
    present_standard = [f for f in current_order if f in set(required_fields)]
    expected_order = [f for f in required_fields if f in present_standard]

    if present_standard != expected_order:
        header_lines_list = []
        header_start = None
        header_end = None
        for i, line in enumerate(lines):
            m = HEADER_RE.match(line.strip())
            if m:
                if header_start is None:
                    header_start = i
                header_end = i
                header_lines_list.append((m.group(1).strip(), line))

        if header_start is not None:
            standard_lines = {name: line for name, line in header_lines_list if name in set(required_fields)}
            non_standard_lines = [(name, line) for name, line in header_lines_list if name not in set(required_fields)]

            reordered = []
            for f in required_fields:
                if f in standard_lines:
                    reordered.append(standard_lines[f])
            for _, line in non_standard_lines:
                reordered.append(line)

            for idx, i in enumerate(range(header_start, header_end + 1)):
                if idx < len(reordered):
                    lines[i] = reordered[idx]

            fixes.append({"action": "reorder_fields"})
            modified = True

    if modified and not dry_run:
        with open(filepath, "w", encoding="utf-8") as f:
            f.writelines(lines)

    return fixes, skipped


def fix_headers(root, dry_run=False):
    """Auto-fix all fixable header violations. Returns (all_fixes, all_skipped)."""
    spec_dir = root / "docs" / "spec"
    all_fixes = []
    all_skipped = []

    # Spec docs
    for md in sorted(spec_dir.glob("*.md")):
        if md.name in SKIP_FILES:
            continue
        header = parse_header(md)
        violations = validate_header(md, header, SPEC_FIELDS, root)
        fixable = [v for v in violations if v.get("auto_fixable")]
        if fixable:
            fixes, skipped = fix_header_file(md, header, SPEC_FIELDS, dry_run)
            rel = str(md.relative_to(root))
            if fixes:
                all_fixes.append({"file": rel, "fixes": fixes})
            if skipped:
                all_skipped.append({"file": rel, "skipped": skipped})

    # Plan docs inside spec-centric subjects
    for md in sorted(spec_dir.glob("*/plans/*/*.md")):
        if md.name in SKIP_FILES:
            continue
        required = CHECKLIST_FIELDS if is_checklist(md) else PLAN_FIELDS
        header = parse_header(md)
        violations = validate_header(md, header, required, root)
        fixable = [v for v in violations if v.get("auto_fixable")]
        if fixable:
            fixes, skipped = fix_header_file(md, header, required, dry_run)
            rel_str = str(md.relative_to(root))
            if fixes:
                all_fixes.append({"file": rel_str, "fixes": fixes})
            if skipped:
                all_skipped.append({"file": rel_str, "skipped": skipped})

    return all_fixes, all_skipped


def fix_index_columns(root, dry_run=False):
    """Auto-fix INDEX column values (version/date) to match headers, and migrate
    rows to the section that matches their underlying doc Header `状态`. Returns
    (fixes, skipped) — skipped covers transitions that share status mismatch but
    cannot be auto-migrated (e.g., superseded → active, where column shape differs).
    """
    fixes = []
    skipped = []

    # Spec INDEX (column-only fix; spec INDEX uses domain grouping not status grouping)
    spec_dir = root / "docs" / "spec"
    spec_index = spec_dir / "INDEX.md"
    if spec_index.exists():
        spec_fixes = _fix_spec_index(spec_dir, spec_index, root, dry_run)
        fixes.extend(spec_fixes)

    for plans_dir in sorted(spec_dir.glob("*/plans")):
        plan_index = plans_dir / "INDEX.md"
        if not plan_index.exists():
            continue
        # Pass 1: column values inside whichever section currently holds the row.
        plan_col_fixes = _fix_plan_index(plans_dir, plan_index, root, dry_run)
        fixes.extend(plan_col_fixes)
        # Pass 2: relocate rows whose doc 状态 doesn't match their containing section.
        # Pass 1 already updated the row's 状态 cell on disk (when not dry-run), so the
        # migration sees a consistent row whether or not column drift was repaired.
        mig_fixes, mig_skipped = _migrate_plan_index_rows(plans_dir, plan_index, root, dry_run)
        fixes.extend(mig_fixes)
        skipped.extend(mig_skipped)

    return fixes, skipped


def _new_plan_section(status, number):
    """Return (heading_line, table_header_line, separator_line) bytes for a brand-new
    plans/INDEX.md section matching the active/draft/completed conventions used by
    /init-docs `subspec-plans-index.md` and `docs/spec/TEMPLATES.md`.
    """
    label = {
        "active": "进行中（Active）",
        "completed": "已完成（Completed）",
        "draft": "草稿（Draft）",
    }.get(status, status)
    date_label = "完成日期" if status == "completed" else "更新日期"
    heading = f"## {number} {label}\n"
    table_header = f"| 计划 | 文件 | 版本 | 状态 | {date_label} |\n"
    separator = "|------|------|------|------|----------|\n"
    return heading, table_header, separator


def _migrate_plan_index_rows(plans_dir, index_path, root, dry_run):
    """Move plans/INDEX.md rows to the section whose group `状态` matches the row's
    underlying plan doc Header `状态`. Same-shape transitions (active / draft /
    completed) are auto-fixed; superseded transitions are returned as `skipped`
    because their section drops the version/date columns and the migration would
    have to fabricate or drop column data.

    Returns (fixes, skipped). Each entry in `fixes` describes the migration; each
    entry in `skipped` matches the schema used by `format_fix_result`'s Skipped
    section (`{file, skipped: [{action, field, reason}]}`).
    """
    fixes = []
    skipped = []

    with open(index_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Walk lines to build a list of sections and classify each line. We track HTML
    # comment ranges so example tables inside `<!-- -->` blocks don't count as
    # data rows.
    in_comment = False
    sections = []  # ordered list of section dicts
    current = None

    for i, line in enumerate(lines):
        stripped = line.strip()

        if not in_comment and "<!--" in stripped:
            # Single-line comment? if "-->" follows on the same line, it's contained.
            comment_open_idx = stripped.find("<!--")
            after_open = stripped[comment_open_idx + 4 :]
            if "-->" not in after_open:
                in_comment = True
                continue
        if in_comment:
            if "-->" in stripped:
                in_comment = False
            continue

        m = re.match(r"^##\s+(\d+)\s+(.+)$", stripped)
        if m:
            section_number = int(m.group(1))
            section_name = m.group(2)
            status = None
            for keyword, st in PLAN_GROUP_STATUS.items():
                if keyword in section_name:
                    status = st
                    break
            current = {
                "heading_line": i,
                "number": section_number,
                "name": section_name,
                "status": status,
                "table_header_line": None,
                "separator_line": None,
                "row_lines": [],
            }
            sections.append(current)
            continue

        if current is None:
            continue

        if SEPARATOR_RE.match(stripped):
            current["separator_line"] = i
            if i > 0 and lines[i - 1].strip().startswith("|"):
                current["table_header_line"] = i - 1
            continue

        if (
            stripped.startswith("|")
            and current.get("separator_line") is not None
        ):
            current["row_lines"].append(i)

    # Identify migrations: rows whose doc 状态 doesn't match the section status.
    same_shape = {"active", "draft", "completed"}
    delete_lines = set()
    pending_inserts = {}  # target_status -> list[row_text]

    for sec_idx, sec in enumerate(sections):
        if sec["status"] is None:
            continue
        for row_line_idx in sec["row_lines"]:
            row_text = lines[row_line_idx]
            row_stripped = row_text.strip()
            cells = [c.strip() for c in row_stripped.split("|") if c.strip()]
            if len(cells) < 2:
                continue
            # cells[0] = plan cell, cells[1] = files cell
            file_links = LINK_RE.findall(cells[1])
            if not file_links:
                continue
            first_file = file_links[0][1]
            doc_path = (plans_dir / first_file).resolve()
            if not doc_path.exists():
                continue

            header = parse_header(doc_path)
            doc_status = header["standard"].get("状态")
            if doc_status is None or doc_status == sec["status"]:
                continue

            rel = str(doc_path.relative_to(root))

            # Sub-rows (`↳` indented continuations of the previous plan) are
            # advisory drift — leave them alone so we don't split an entry from
            # its parent.
            if cells[0].startswith("↳"):
                skipped.append({
                    "file": rel,
                    "skipped": [{
                        "action": "migrate_row",
                        "field": "状态(group)",
                        "reason": f"sub-row continuation; manual placement required (header={doc_status}, section={sec['status']})",
                    }],
                })
                continue

            if sec["status"] not in same_shape or doc_status not in same_shape:
                skipped.append({
                    "file": rel,
                    "skipped": [{
                        "action": "migrate_row",
                        "field": "状态(group)",
                        "reason": f"superseded sections use a different column shape; migrate manually (header={doc_status}, section={sec['status']})",
                    }],
                })
                continue

            delete_lines.add(row_line_idx)
            pending_inserts.setdefault(doc_status, []).append(row_text)
            fixes.append({
                "index": str(index_path.relative_to(root)),
                "file": str(doc_path.relative_to(plans_dir)),
                "action": "migrate_row",
                "from_section": sec["name"],
                "to_status": doc_status,
            })

    if not delete_lines and not pending_inserts:
        return fixes, skipped

    # Compute per-section row insertion anchors. The anchor is the original
    # line index where new rows should be inserted (before that line in the
    # original file). For a section that already has surviving rows, we
    # insert directly after the last surviving row. For a section whose rows
    # are all being moved out (or that has none), we insert immediately after
    # the separator line.
    insertion_point = {}  # status -> original line idx where new rows go before
    for sec in sections:
        if sec["status"] is None or sec["status"] not in same_shape:
            continue
        if sec["status"] not in pending_inserts:
            continue
        survivors = [r for r in sec["row_lines"] if r not in delete_lines]
        if survivors:
            insertion_point[sec["status"]] = survivors[-1] + 1
        elif sec["separator_line"] is not None:
            insertion_point[sec["status"]] = sec["separator_line"] + 1
        # If neither rows nor separator exist, the section is malformed; we
        # treat it as missing and append a fresh section at EOF below.

    # Walk original lines and build new file content.
    new_lines = []
    pending_remaining = {st: list(rows) for st, rows in pending_inserts.items()}
    next_section_number = max((sec["number"] for sec in sections), default=0)

    for i, line in enumerate(lines):
        if i in delete_lines:
            continue
        new_lines.append(line)
        next_orig_idx = i + 1
        for status, point_idx in list(insertion_point.items()):
            if point_idx == next_orig_idx and pending_remaining.get(status):
                rows = pending_remaining.pop(status)
                # Make sure the row text ends with a newline for clean append.
                for row in rows:
                    new_lines.append(row if row.endswith("\n") else row + "\n")

    # Any target status that has no existing section gets a new section at EOF.
    new_section_status_in_order = [
        st for st in pending_remaining if pending_remaining.get(st)
    ]
    if new_section_status_in_order:
        if new_lines and not new_lines[-1].endswith("\n"):
            new_lines[-1] = new_lines[-1] + "\n"
        # Ensure a blank line before the first newly created section.
        if new_lines and new_lines[-1].strip() != "":
            new_lines.append("\n")
        for status in new_section_status_in_order:
            rows = pending_remaining[status]
            if not rows:
                continue
            next_section_number += 1
            heading, table_header, separator = _new_plan_section(status, next_section_number)
            new_lines.append(heading)
            new_lines.append("\n")
            new_lines.append(table_header)
            new_lines.append(separator)
            for row in rows:
                new_lines.append(row if row.endswith("\n") else row + "\n")
            new_lines.append("\n")

    if (delete_lines or new_section_status_in_order) and not dry_run:
        with open(index_path, "w", encoding="utf-8") as f:
            f.writelines(new_lines)

    return fixes, skipped


def _fix_spec_index(spec_dir, index_path, root, dry_run):
    """Fix spec INDEX column values."""
    fixes = []
    with open(index_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Build set of table-header line indices
    header_lines = set()
    for i, line in enumerate(lines):
        if SEPARATOR_RE.match(line.strip()):
            if i > 0:
                header_lines.add(i - 1)

    modified = False
    for i, line in enumerate(lines):
        stripped = line.strip()
        if not stripped.startswith("|"):
            continue
        if SEPARATOR_RE.match(stripped) or i in header_lines:
            continue
        cells = [c.strip() for c in stripped.split("|")]
        cells_raw = [c for c in cells if c]
        if len(cells_raw) < 4:
            continue

        links = LINK_RE.findall(cells_raw[0])
        if not links:
            continue
        file_path = links[0][1]
        if file_path.startswith("../"):
            continue

        doc_path = (spec_dir / file_path).resolve()
        if not doc_path.exists():
            continue

        header = parse_header(doc_path)
        std = header["standard"]
        idx_ver, idx_status, idx_date = cells_raw[1], cells_raw[2], cells_raw[3]
        new_ver = std.get("版本", idx_ver)
        new_status = std.get("状态", idx_status)
        new_date = std.get("更新日期", idx_date)

        if new_ver != idx_ver or new_status != idx_status or new_date != idx_date:
            # Rebuild only the canonical Header projection cells, preserving
            # extra columns such as the spec INDEX `Plans` link.
            new_cells = [cells_raw[0], new_ver, new_status, new_date, *cells_raw[4:]]
            new_line = "| " + " | ".join(new_cells) + " |\n"
            lines[i] = new_line
            modified = True
            fixes.append({
                "index": "docs/spec/INDEX.md",
                "file": file_path,
                "changes": {
                    k: {"from": old, "to": new}
                    for k, old, new in [("版本", idx_ver, new_ver), ("状态", idx_status, new_status), ("更新日期", idx_date, new_date)]
                    if old != new
                },
            })

    if modified and not dry_run:
        with open(index_path, "w", encoding="utf-8") as f:
            f.writelines(lines)

    return fixes


def _fix_plan_index(plans_dir, index_path, root, dry_run):
    """Fix per-subspec plans/INDEX.md column values."""
    fixes = []
    with open(index_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    header_lines = set()
    for i, line in enumerate(lines):
        if SEPARATOR_RE.match(line.strip()) and i > 0:
            header_lines.add(i - 1)

    modified = False
    for i, line in enumerate(lines):
        stripped = line.strip()
        if not stripped.startswith("|"):
            continue
        if SEPARATOR_RE.match(stripped) or i in header_lines:
            continue
        cells = [c.strip() for c in stripped.split("|")]
        cells_raw = [c for c in cells if c]
        if len(cells_raw) < 5:
            continue

        file_links = LINK_RE.findall(cells_raw[1])
        if not file_links:
            continue

        first_file = file_links[0][1]
        doc_path = (plans_dir / first_file).resolve()
        if not doc_path.exists():
            continue

        header = parse_header(doc_path)
        std = header["standard"]
        idx_ver, idx_status, idx_date = cells_raw[2], cells_raw[3], cells_raw[4]
        new_ver = std.get("版本", idx_ver)
        new_status = std.get("状态", idx_status)
        new_date = std.get("更新日期", idx_date)

        if new_ver != idx_ver or new_status != idx_status or new_date != idx_date:
            new_line = f"| {cells_raw[0]} | {cells_raw[1]} | {new_ver} | {new_status} | {new_date} |\n"
            lines[i] = new_line
            modified = True
            fixes.append({
                "index": str(index_path.relative_to(root)),
                "file": first_file,
                "changes": {
                    k: {"from": old, "to": new}
                    for k, old, new in [
                        ("版本", idx_ver, new_ver),
                        ("状态", idx_status, new_status),
                        ("更新日期", idx_date, new_date),
                    ]
                    if old != new
                },
            })

    if modified and not dry_run:
        with open(index_path, "w", encoding="utf-8") as f:
            f.writelines(lines)

    return fixes


# ── Output Formatters ──────────────────────────────────────────────────


def format_human(report):
    """Format report as human-readable text."""
    lines = []

    lines.append("## Header Violations")
    if report["header_violations"]:
        for v in report["header_violations"]:
            auto = " [auto-fixable]" if v.get("auto_fixable") else " [needs LLM]"
            lines.append(f"  - {v['file']}: {v['message']}{auto}")
    else:
        lines.append("  (none)")

    lines.append("")
    lines.append("## INDEX Drifts")
    if report["index_drifts"]:
        for d in report["index_drifts"]:
            auto = " [auto-fixable]" if d.get("auto_fixable") else " [needs LLM]"
            lines.append(f"  - {d['file']}: {d['field']} header={d['header_value']} index={d['index_value']}{auto}")
    else:
        lines.append("  (none)")

    lines.append("")
    lines.append("## Orphans")
    orphans = report["orphans"]
    if orphans["missing_from_index"] or orphans["dangling_index_entries"]:
        for f in orphans["missing_from_index"]:
            lines.append(f"  - [not in INDEX] {f}")
        for f in orphans["dangling_index_entries"]:
            lines.append(f"  - [dangling] {f}")
    else:
        lines.append("  (none)")

    lines.append("")
    lines.append("## Warnings")
    if report["warnings"]:
        for w in report["warnings"]:
            lines.append(f"  - {w.get('entry', '')}: {w['reason']}")
    else:
        lines.append("  (none)")

    lines.append("")
    s = report["summary"]
    if s["violations"] == 0 and s["drifts"] == 0 and s["orphans"] == 0:
        lines.append("All documents are in sync. Zero drift detected.")
    else:
        lines.append(f"Summary: {s['violations']} violations, {s['drifts']} drifts, {s['orphans']} orphans, {s['warnings']} warnings")
        lines.append(f"  auto-fixable: {s['auto_fixable']}, needs LLM: {s['needs_llm']}")

    return "\n".join(lines)


def format_fix_result(fixes, skipped, mode, dry_run):
    """Format fix results as structured human-readable text with 3 sections."""
    lines = []
    prefix = "[DRY RUN] " if dry_run else ""

    # Section 1: Applied fixes
    lines.append(f"## {prefix}Applied ({mode})")
    if fixes:
        lines.append(f"  {len(fixes)} file(s) {'would be ' if dry_run else ''}modified:")
        for fix in fixes:
            if "file" in fix and "fixes" in fix:
                lines.append(f"  {fix['file']}:")
                for f in fix["fixes"]:
                    lines.append(f"    - {f.get('action', 'unknown')}: {json.dumps({k: v for k, v in f.items() if k != 'action'}, ensure_ascii=False)}")
            elif fix.get("action") == "migrate_row":
                lines.append(f"  {fix['index']} → {fix.get('file', '?')}:")
                lines.append(f"    - migrate_row: from \"{fix.get('from_section', '?')}\" → status={fix.get('to_status', '?')}")
            elif "index" in fix:
                lines.append(f"  {fix['index']} → {fix.get('file', '?')}:")
                for field, change in fix.get("changes", {}).items():
                    lines.append(f"    - {field}: {change['from']} → {change['to']}")
    else:
        lines.append("  (none)")

    # Section 2: Skipped (auto-fix attempted but failed)
    lines.append("")
    lines.append(f"## {prefix}Skipped (needs LLM)")
    if skipped:
        lines.append(f"  {len(skipped)} file(s) with unresolved issues:")
        for item in skipped:
            lines.append(f"  {item['file']}:")
            for s in item["skipped"]:
                lines.append(f"    - {s.get('action', 'unknown')}: field={s.get('field', '?')} reason={s.get('reason', '?')}")
    else:
        lines.append("  (none)")

    return "\n".join(lines)


# ── Main ───────────────────────────────────────────────────────────────


def find_root():
    """Find project root by looking for docs/ directory."""
    cwd = Path.cwd()
    for candidate in [cwd, *cwd.parents]:
        if (candidate / "docs" / "spec").is_dir():
            return candidate
    print("Error: Cannot find project root (no docs/spec/ found)", file=sys.stderr)
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Check and fix document Header / INDEX drift")
    group = parser.add_mutually_exclusive_group()
    group.add_argument("--check", action="store_true", default=True, help="Audit mode (default)")
    group.add_argument("--fix-header", action="store_true", help="Auto-fix header violations")
    group.add_argument("--fix-index", action="store_true", help="Auto-fix INDEX column drifts")
    parser.add_argument("--json", action="store_true", help="JSON output (with --check)")
    parser.add_argument("--dry-run", action="store_true", help="Preview changes without writing (with --fix-*)")
    args = parser.parse_args()

    root = find_root()

    if args.fix_header:
        fixes, skipped = fix_headers(root, dry_run=args.dry_run)
        print(format_fix_result(fixes, skipped, "--fix-header", args.dry_run))
        # Section 3: Post-fix verification
        print("")
        print("## Post-fix Verification")
        report = run_check(root)
        s = report["summary"]
        if s["violations"] == 0 and s["drifts"] == 0 and s["orphans"] == 0:
            print("  All documents are in sync. Zero drift detected.")
        else:
            print(f"  {s['violations']} violations, {s['drifts']} drifts, {s['orphans']} orphans, {s['warnings']} warnings")
            print(f"  auto-fixable: {s['auto_fixable']}, needs LLM: {s['needs_llm']}")
            for v in report["header_violations"]:
                if not v.get("auto_fixable"):
                    print(f"  - [header] {v['file']}: {v['message']}")
            for d in report["index_drifts"]:
                if not d.get("auto_fixable"):
                    print(f"  - [drift] {d['file']}: {d['field']} header={d['header_value']} index={d['index_value']}")
            for f in report["orphans"]["missing_from_index"]:
                print(f"  - [orphan] {f}")
    elif args.fix_index:
        fixes, skipped = fix_index_columns(root, dry_run=args.dry_run)
        print(format_fix_result(fixes, skipped, "--fix-index", args.dry_run))
        # Section 3: Post-fix verification
        print("")
        print("## Post-fix Verification")
        report = run_check(root)
        s = report["summary"]
        if s["violations"] == 0 and s["drifts"] == 0 and s["orphans"] == 0:
            print("  All documents are in sync. Zero drift detected.")
        else:
            print(f"  {s['violations']} violations, {s['drifts']} drifts, {s['orphans']} orphans, {s['warnings']} warnings")
            print(f"  auto-fixable: {s['auto_fixable']}, needs LLM: {s['needs_llm']}")
            for d in report["index_drifts"]:
                if not d.get("auto_fixable"):
                    print(f"  - [drift] {d['file']}: {d['field']} header={d['header_value']} index={d['index_value']}")
            for f in report["orphans"]["missing_from_index"]:
                print(f"  - [orphan] {f}")
    else:
        report = run_check(root)
        if args.json:
            print(json.dumps(report, ensure_ascii=False, indent=2))
        else:
            print(format_human(report))


if __name__ == "__main__":
    main()
