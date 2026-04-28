#!/usr/bin/env python3
"""scripts/lint/openapi_diff.py — openapi-v1-contract breaking-change gate.

Compares the live `openapi/openapi.yaml` against a frozen baseline under
`openapi/baseline/` and classifies each finding per
[openapi-v1-contract spec §4.4](docs/spec/openapi-v1-contract/spec.md).

The wrapper IS the source of truth for the final exit code. Spec §3 risk row
mandates that any disagreement between an external tool and spec §4.4 must be
reclassified by this wrapper before exit. To keep the local gate runnable
without a network-dependent third-party CLI, this wrapper implements the
diff directly. A future revision may shell to OpenAPITools openapi-diff first
and reclassify against `openapi/diff-config.yaml`; the wrapper must remain
authoritative for exit codes.

CLI:

    --repo-root PATH                Default: parent of this script's grandparent.
    --baseline PATH                 Explicit baseline file. Wins over --baseline-version.
    --baseline-version vX.Y.Z       Pick `openapi/baseline/openapi-<ver>.yaml`.
    --current PATH                  Default: `openapi/openapi.yaml`.
    --config PATH                   Default: `openapi/diff-config.yaml`.
    --history PATH                  Default: `docs/spec/openapi-v1-contract/history.md`.
    --history-ref REF               Git ref to compare history.md against (default: HEAD).
    --fail-on-incompatible / --no-fail-on-incompatible
                                    Default: fail-on-incompatible.
    --output PATH                   Optional JSON output path; otherwise stdout.
    --print-default-baseline        Print the resolved default baseline path and exit 0.

Exit codes:

    0  — no breaking change (or all breaking changes were whitelisted AND the
         whitelist gate is satisfied).
    1  — at least one unwhitelisted breaking change, or whitelist gate failed
         (e.g., privacy 501→202 transition without a history.md increment).
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from collections import OrderedDict
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple

import yaml


WRAPPER_VERSION = "1.0.0"
HTTP_METHODS = {"get", "put", "post", "delete", "options", "head", "patch", "trace"}
COMPOSITION_KEYS = ("allOf", "oneOf", "anyOf")

SemVerTuple = Tuple[int, int, int]


# ---------- helpers -----------------------------------------------------------


def _yaml_load(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as f:
        return yaml.safe_load(f) or {}


def _resolve_baseline(repo_root: Path, version: Optional[str]) -> Path:
    baseline_dir = repo_root / "openapi" / "baseline"
    if not baseline_dir.is_dir():
        raise SystemExit(
            f"ERROR: baseline directory not found: {baseline_dir} "
            f"(expected per spec §3.1 D-10 / plan 003 Phase 1.1)"
        )
    if version:
        candidate = baseline_dir / f"openapi-{version}.yaml"
        if not candidate.is_file():
            raise SystemExit(f"ERROR: baseline file not found: {candidate}")
        return candidate
    candidates: List[Tuple[SemVerTuple, Path]] = []
    semver_re = re.compile(r"^openapi-v(\d+)\.(\d+)\.(\d+)\.yaml$")
    for entry in baseline_dir.iterdir():
        m = semver_re.match(entry.name)
        if m and entry.is_file():
            candidates.append(((int(m.group(1)), int(m.group(2)), int(m.group(3))), entry))
    if not candidates:
        raise SystemExit(
            f"ERROR: no baseline files in {baseline_dir} matching openapi-vX.Y.Z.yaml"
        )
    candidates.sort(key=lambda t: t[0])
    return candidates[-1][1]


def _git_show(repo_root: Path, ref: str, path: Path) -> Optional[str]:
    rel = path.relative_to(repo_root).as_posix()
    try:
        out = subprocess.run(
            ["git", "-C", str(repo_root), "show", f"{ref}:{rel}"],
            check=True,
            capture_output=True,
        )
        return out.stdout.decode("utf-8", errors="replace")
    except subprocess.CalledProcessError:
        return None


def _git_rev_parse(repo_root: Path, ref: str) -> Optional[str]:
    try:
        out = subprocess.run(
            ["git", "-C", str(repo_root), "rev-parse", "--verify", f"{ref}^{{commit}}"],
            check=True,
            capture_output=True,
        )
        return out.stdout.decode("utf-8", errors="replace").strip()
    except subprocess.CalledProcessError:
        return None


def _git_merge_base(repo_root: Path, left: str, right: str) -> Optional[str]:
    try:
        out = subprocess.run(
            ["git", "-C", str(repo_root), "merge-base", left, right],
            check=True,
            capture_output=True,
        )
        return out.stdout.decode("utf-8", errors="replace").strip()
    except subprocess.CalledProcessError:
        return None


def _resolve_history_ref(repo_root: Path, requested_ref: str, config: Dict[str, Any]) -> str:
    if requested_ref and requested_ref != "auto":
        return requested_ref

    tooling = config.get("tooling") or {}
    configured_base = tooling.get("historyDiffBase") or "dev"
    candidates: List[str] = []
    for ref in (
        configured_base,
        f"origin/{configured_base}" if configured_base else None,
        "main",
        "origin/main",
        "master",
        "origin/master",
    ):
        if ref and ref not in candidates:
            candidates.append(ref)

    for ref in candidates:
        if not _git_rev_parse(repo_root, ref):
            continue
        merge_base = _git_merge_base(repo_root, "HEAD", ref)
        if merge_base:
            return merge_base
    return "HEAD"


def _count_history_rows(text: str) -> int:
    """Count data rows in the first markdown table inside `text`.

    A data row is a line starting with `|`, whose first non-pipe cell is not
    blank, that is not the header row (which contains `日期`) and not the
    separator row (which is dashes-and-pipes only). The function is tolerant
    of trailing whitespace and other tables further down the document.
    """
    rows = 0
    in_table = False
    for raw in text.splitlines():
        line = raw.rstrip()
        stripped = line.strip()
        if not stripped.startswith("|"):
            if in_table:
                break
            continue
        cells = [c.strip() for c in stripped.strip("|").split("|")]
        if not cells:
            continue
        # Header row marker: contains `日期` cell.
        if "日期" in cells and "版本" in cells:
            in_table = True
            continue
        # Separator row: all cells made of dashes / colons.
        if in_table and all(re.fullmatch(r":?-+:?", c) for c in cells if c):
            continue
        if in_table and cells[0]:
            rows += 1
    return rows


def _load_diff_config(path: Path) -> Dict[str, Any]:
    if not path.is_file():
        return {}
    return _yaml_load(path)


# ---------- finding classifier -----------------------------------------------


def _new_finding(kind: str, severity: str, **fields: Any) -> Dict[str, Any]:
    finding: "OrderedDict[str, Any]" = OrderedDict()
    finding["kind"] = kind
    finding["severity"] = severity
    for k, v in fields.items():
        finding[k] = v
    return finding


def _is_method(method: str) -> bool:
    return method.lower() in HTTP_METHODS


def _operation_id(op: Dict[str, Any]) -> str:
    return str(op.get("operationId") or "")


def _ref_or_dict(node: Any) -> Tuple[Optional[str], Dict[str, Any]]:
    if isinstance(node, dict):
        return node.get("$ref"), node
    return None, {}


def diff_schema_node(
    loc: str,
    base: Any,
    cur: Any,
    visited: Optional[set] = None,
) -> List[Dict[str, Any]]:
    if visited is None:
        visited = set()
    findings: List[Dict[str, Any]] = []
    if not isinstance(base, dict) and not isinstance(cur, dict):
        return findings
    base = base or {}
    cur = cur or {}
    bref, _ = _ref_or_dict(base)
    cref, _ = _ref_or_dict(cur)
    if bref or cref:
        if bref != cref:
            findings.append(
                _new_finding("ref-changed", "breaking", where=loc, fromRef=bref, toRef=cref)
            )
        return findings

    base_compositions = [key for key in COMPOSITION_KEYS if key in base]
    cur_compositions = [key for key in COMPOSITION_KEYS if key in cur]
    for key in base_compositions:
        if key not in cur_compositions:
            findings.append(
                _new_finding("composition-removed", "breaking", where=loc, composition=key)
            )
    for key in cur_compositions:
        if key not in base_compositions:
            findings.append(
                _new_finding("composition-added", "breaking", where=loc, composition=key)
            )
    for key in base_compositions:
        if key not in cur_compositions:
            continue
        base_items = base.get(key) or []
        cur_items = cur.get(key) or []
        if not isinstance(base_items, list) or not isinstance(cur_items, list):
            if base_items != cur_items:
                findings.append(
                    _new_finding(
                        "composition-changed",
                        "breaking",
                        where=loc,
                        composition=key,
                    )
                )
            continue
        shared = min(len(base_items), len(cur_items))
        for i in range(shared):
            findings.extend(
                diff_schema_node(f"{loc}.{key}[{i}]", base_items[i], cur_items[i], visited)
            )
        for i in range(shared, len(base_items)):
            findings.append(
                _new_finding(
                    "composition-branch-removed",
                    "breaking",
                    where=f"{loc}.{key}[{i}]",
                    composition=key,
                )
            )
        for i in range(shared, len(cur_items)):
            findings.append(
                _new_finding(
                    "composition-branch-added",
                    "breaking",
                    where=f"{loc}.{key}[{i}]",
                    composition=key,
                )
            )

    btype = base.get("type")
    ctype = cur.get("type")
    if btype and ctype and btype != ctype:
        findings.append(
            _new_finding("type-changed", "breaking", where=loc, fromType=btype, toType=ctype)
        )
        return findings
    # nullable / format flips can be material; spec §4.4 doesn't enumerate them,
    # so we surface as informational rather than failing the gate.
    if base.get("format") and cur.get("format") and base["format"] != cur["format"]:
        findings.append(
            _new_finding(
                "format-changed",
                "informational",
                where=loc,
                fromFormat=base["format"],
                toFormat=cur["format"],
            )
        )
    # object: properties + required
    has_object_shape = (
        btype == "object" or ctype == "object" or "properties" in base or "properties" in cur
    )
    if has_object_shape:
        bprops = base.get("properties") or {}
        cprops = cur.get("properties") or {}
        breq = set(base.get("required") or [])
        creq = set(cur.get("required") or [])
        for p in bprops:
            child_loc = f"{loc}.{p}"
            if p not in cprops:
                findings.append(_new_finding("field-deleted", "breaking", where=child_loc))
            else:
                findings.extend(diff_schema_node(child_loc, bprops[p], cprops[p], visited))
        for p in cprops:
            if p not in bprops:
                child_loc = f"{loc}.{p}"
                if p in creq:
                    findings.append(
                        _new_finding("field-required-added", "breaking", where=child_loc)
                    )
                else:
                    findings.append(_new_finding("field-added", "additive", where=child_loc))
        for p in (creq - breq) & set(bprops):
            findings.append(
                _new_finding("field-promoted-required", "breaking", where=f"{loc}.{p}")
            )
    # array: items
    if btype == "array" and ctype == "array":
        findings.extend(
            diff_schema_node(f"{loc}.items", base.get("items"), cur.get("items"), visited)
        )
    # enum: per spec §4.4, removing a value is breaking, adding a string-typed
    # value is additive. Non-string enum widening is also surfaced as breaking
    # because the rule explicitly scopes additive enums to string-typed.
    benum = base.get("enum")
    cenum = cur.get("enum")
    if benum is not None and cenum is not None:
        bset = list(benum)
        cset = list(cenum)
        for v in bset:
            if v not in cset:
                findings.append(
                    _new_finding("enum-value-removed", "breaking", where=loc, value=v)
                )
        scoped_string = (btype == "string" or ctype == "string") and not (btype and btype != "string")
        for v in cset:
            if v not in bset:
                if scoped_string:
                    findings.append(
                        _new_finding("enum-value-added", "additive", where=loc, value=v)
                    )
                else:
                    findings.append(
                        _new_finding("enum-value-added-non-string", "breaking", where=loc, value=v)
                    )
    return findings


def diff_operation(
    path: str,
    method: str,
    base_op: Dict[str, Any],
    cur_op: Dict[str, Any],
) -> List[Dict[str, Any]]:
    findings: List[Dict[str, Any]] = []

    # parameters keyed by (name, in)
    def _key(p: Dict[str, Any]) -> Tuple[str, str]:
        return (str(p.get("name", "")), str(p.get("in", "query")))

    base_params = {_key(p): p for p in base_op.get("parameters", []) or []}
    cur_params = {_key(p): p for p in cur_op.get("parameters", []) or []}
    for key, base_p in base_params.items():
        if key not in cur_params:
            findings.append(
                _new_finding(
                    "parameter-removed",
                    "breaking",
                    path=path,
                    method=method,
                    parameter=key[0],
                    location=key[1],
                )
            )
            continue
        cur_p = cur_params[key]
        if cur_p.get("required") and not base_p.get("required"):
            findings.append(
                _new_finding(
                    "parameter-promoted-required",
                    "breaking",
                    path=path,
                    method=method,
                    parameter=key[0],
                    location=key[1],
                )
            )
        b_schema = base_p.get("schema") or {}
        c_schema = cur_p.get("schema") or {}
        findings.extend(
            diff_schema_node(
                f"{method.upper()} {path} param[{key[1]}:{key[0]}]",
                b_schema,
                c_schema,
            )
        )
    for key, cur_p in cur_params.items():
        if key not in base_params:
            severity = "breaking" if cur_p.get("required") else "additive"
            kind = (
                "parameter-required-added-new" if cur_p.get("required") else "parameter-added"
            )
            findings.append(
                _new_finding(
                    kind,
                    severity,
                    path=path,
                    method=method,
                    parameter=key[0],
                    location=key[1],
                )
            )

    # requestBody
    base_body = base_op.get("requestBody") or {}
    cur_body = cur_op.get("requestBody") or {}
    if base_body and not cur_body:
        findings.append(
            _new_finding("request-body-removed", "breaking", path=path, method=method)
        )
    elif cur_body and not base_body:
        if cur_body.get("required"):
            findings.append(
                _new_finding(
                    "request-body-required-added", "breaking", path=path, method=method
                )
            )
        else:
            findings.append(
                _new_finding("request-body-added", "additive", path=path, method=method)
            )
    elif base_body and cur_body:
        if base_body.get("required") and not cur_body.get("required"):
            findings.append(
                _new_finding(
                    "request-body-required-relaxed",
                    "informational",
                    path=path,
                    method=method,
                )
            )
        if cur_body.get("required") and not base_body.get("required"):
            findings.append(
                _new_finding(
                    "request-body-promoted-required",
                    "breaking",
                    path=path,
                    method=method,
                )
            )
        b_content = (base_body.get("content") or {})
        c_content = (cur_body.get("content") or {})
        for media_type in set(b_content) | set(c_content):
            b_media = b_content.get(media_type) or {}
            c_media = c_content.get(media_type) or {}
            if b_media and not c_media:
                findings.append(
                    _new_finding(
                        "request-content-removed",
                        "breaking",
                        path=path,
                        method=method,
                        mediaType=media_type,
                    )
                )
                continue
            if c_media and not b_media:
                findings.append(
                    _new_finding(
                        "request-content-added",
                        "additive",
                        path=path,
                        method=method,
                        mediaType=media_type,
                    )
                )
                continue
            findings.extend(
                diff_schema_node(
                    f"{method.upper()} {path} requestBody[{media_type}]",
                    b_media.get("schema") or {},
                    c_media.get("schema") or {},
                )
            )

    # responses
    base_resps = base_op.get("responses") or {}
    cur_resps = cur_op.get("responses") or {}
    for code, base_resp in base_resps.items():
        if code not in cur_resps:
            findings.append(
                _new_finding(
                    "response-status-removed",
                    "breaking",
                    path=path,
                    method=method,
                    status=str(code),
                )
            )
            continue
        cur_resp = cur_resps[code]
        b_content = (base_resp.get("content") if isinstance(base_resp, dict) else {}) or {}
        c_content = (cur_resp.get("content") if isinstance(cur_resp, dict) else {}) or {}
        for media_type in set(b_content) | set(c_content):
            b_media = b_content.get(media_type) or {}
            c_media = c_content.get(media_type) or {}
            if b_media and not c_media:
                findings.append(
                    _new_finding(
                        "response-content-removed",
                        "breaking",
                        path=path,
                        method=method,
                        status=str(code),
                        mediaType=media_type,
                    )
                )
                continue
            if c_media and not b_media:
                findings.append(
                    _new_finding(
                        "response-content-added",
                        "additive",
                        path=path,
                        method=method,
                        status=str(code),
                        mediaType=media_type,
                    )
                )
                continue
            findings.extend(
                diff_schema_node(
                    f"{method.upper()} {path} response[{code}][{media_type}]",
                    b_media.get("schema") or {},
                    c_media.get("schema") or {},
                )
            )
    for code in cur_resps:
        if code not in base_resps:
            findings.append(
                _new_finding(
                    "response-status-added",
                    "additive",
                    path=path,
                    method=method,
                    status=str(code),
                )
            )

    return findings


def diff_documents(baseline: Dict[str, Any], current: Dict[str, Any]) -> List[Dict[str, Any]]:
    findings: List[Dict[str, Any]] = []

    # tags
    base_tags = OrderedDict(
        (str(t.get("name")), t) for t in (baseline.get("tags") or [])
    )
    cur_tags = OrderedDict(
        (str(t.get("name")), t) for t in (current.get("tags") or [])
    )
    for name in base_tags:
        if name not in cur_tags:
            findings.append(_new_finding("tag-removed", "breaking", tag=name))
    for name in cur_tags:
        if name not in base_tags:
            findings.append(_new_finding("tag-added", "additive", tag=name))

    # paths and operations
    base_paths = baseline.get("paths") or {}
    cur_paths = current.get("paths") or {}
    for path, base_item in base_paths.items():
        if path not in cur_paths:
            findings.append(_new_finding("endpoint-removed", "breaking", path=path))
            continue
        cur_item = cur_paths[path]
        base_methods = {m for m in (base_item or {}) if _is_method(m)}
        cur_methods = {m for m in (cur_item or {}) if _is_method(m)}
        for m in base_methods - cur_methods:
            findings.append(
                _new_finding(
                    "method-removed",
                    "breaking",
                    path=path,
                    method=m.upper(),
                    operationId=_operation_id(base_item.get(m, {})),
                )
            )
        for m in cur_methods - base_methods:
            findings.append(
                _new_finding(
                    "method-added",
                    "additive",
                    path=path,
                    method=m.upper(),
                    operationId=_operation_id(cur_item.get(m, {})),
                )
            )
        for m in sorted(base_methods & cur_methods):
            findings.extend(
                diff_operation(path, m.upper(), base_item.get(m) or {}, cur_item.get(m) or {})
            )
    for path in cur_paths:
        if path not in base_paths:
            findings.append(_new_finding("endpoint-added", "additive", path=path))

    # components.schemas
    base_schemas = ((baseline.get("components") or {}).get("schemas")) or {}
    cur_schemas = ((current.get("components") or {}).get("schemas")) or {}
    for name in base_schemas:
        if name not in cur_schemas:
            findings.append(_new_finding("schema-removed", "breaking", schema=name))
            continue
        findings.extend(
            diff_schema_node(
                f"components.schemas.{name}",
                base_schemas[name],
                cur_schemas[name],
            )
        )
    for name in cur_schemas:
        if name not in base_schemas:
            findings.append(_new_finding("schema-added", "additive", schema=name))

    return findings


# ---------- whitelist ---------------------------------------------------------


def apply_response_status_whitelist(
    findings: List[Dict[str, Any]],
    config: Dict[str, Any],
) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]]]:
    """Downgrade matched (status-removed, status-added) pairs to whitelisted.

    Returns (mutated_findings, whitelist_matches). Each whitelist_match
    records which transition fired and the rule it matched.
    """
    rules = ((config.get("whitelist") or {}).get("responseStatusTransitions")) or []
    if not rules:
        return findings, []

    matches: List[Dict[str, Any]] = []
    out: List[Dict[str, Any]] = list(findings)

    for rule in rules:
        rule_path = rule.get("path")
        rule_method = (rule.get("method") or "").upper()
        rule_from = str(rule.get("from"))
        rule_to = str(rule.get("to"))
        # Privacy export whitelist is also resolvable when path includes /api/v1 prefix
        # vs. when servers strip it. Match by suffix.
        candidates_removed = [
            i
            for i, f in enumerate(out)
            if f["kind"] == "response-status-removed"
            and f.get("method") == rule_method
            and str(f.get("status")) == rule_from
            and (f.get("path") == rule_path or str(f.get("path", "")).endswith(rule_path))
        ]
        candidates_added = [
            i
            for i, f in enumerate(out)
            if f["kind"] == "response-status-added"
            and f.get("method") == rule_method
            and str(f.get("status")) == rule_to
            and (f.get("path") == rule_path or str(f.get("path", "")).endswith(rule_path))
        ]
        if not candidates_removed or not candidates_added:
            continue
        ri = candidates_removed[0]
        ai = candidates_added[0]
        for idx in (ri, ai):
            out[idx] = OrderedDict(out[idx])
            out[idx]["severity"] = "informational"
            out[idx]["whitelist"] = "responseStatusTransition"
            out[idx]["whitelistRule"] = {
                "path": rule_path,
                "method": rule_method,
                "from": rule_from,
                "to": rule_to,
                "reasonRef": rule.get("reasonRef", ""),
            }
        matches.append(
            OrderedDict(
                kind="response-status-transition",
                path=rule_path,
                method=rule_method,
                **{"from": rule_from, "to": rule_to},
                requireHistoryIncrement=bool(rule.get("requireHistoryIncrement", False)),
            )
        )
    return out, matches


def evaluate_history_gate(
    repo_root: Path,
    history_path: Path,
    history_ref: str,
    whitelist_matches: List[Dict[str, Any]],
) -> Optional[Dict[str, Any]]:
    """Return a finding if a whitelist requires a history.md increment but the
    working tree didn't add a row vs. `history_ref`. Returns None on pass or
    if no whitelist requires the gate.
    """
    requires = [m for m in whitelist_matches if m.get("requireHistoryIncrement")]
    if not requires:
        return None
    if not history_path.is_file():
        return _new_finding(
            "history-missing",
            "breaking",
            historyPath=str(history_path),
            note="history.md not found while whitelist requires increment",
        )
    cur_text = history_path.read_text(encoding="utf-8")
    cur_rows = _count_history_rows(cur_text)
    base_text = _git_show(repo_root, history_ref, history_path)
    base_rows = _count_history_rows(base_text) if base_text is not None else 0
    if cur_rows > base_rows:
        return None
    return _new_finding(
        "history-not-incremented",
        "breaking",
        historyPath=str(history_path),
        ref=history_ref,
        rowsAtRef=base_rows,
        rowsAtWorkingTree=cur_rows,
        whitelist=[m for m in requires],
    )


# ---------- entry -------------------------------------------------------------


def _format_summary(findings: Iterable[Dict[str, Any]]) -> Dict[str, int]:
    summary = {"breaking": 0, "additive": 0, "informational": 0}
    for f in findings:
        sev = f.get("severity", "informational")
        summary[sev] = summary.get(sev, 0) + 1
    return summary


def run(args: argparse.Namespace) -> int:
    repo_root = Path(args.repo_root).resolve()

    if args.print_default_baseline:
        path = _resolve_baseline(repo_root, None)
        sys.stdout.write(str(path) + "\n")
        return 0

    if args.baseline:
        baseline_path = Path(args.baseline).resolve()
        if not baseline_path.is_file():
            sys.stderr.write(f"ERROR: baseline file not found: {baseline_path}\n")
            return 1
    else:
        baseline_path = _resolve_baseline(repo_root, args.baseline_version)

    current_path = (
        Path(args.current).resolve()
        if args.current
        else repo_root / "openapi" / "openapi.yaml"
    )
    if not current_path.is_file():
        sys.stderr.write(f"ERROR: current file not found: {current_path}\n")
        return 1

    config_path = (
        Path(args.config).resolve() if args.config else repo_root / "openapi" / "diff-config.yaml"
    )
    history_path = (
        Path(args.history).resolve()
        if args.history
        else repo_root / "docs" / "spec" / "openapi-v1-contract" / "history.md"
    )

    sys.stderr.write(
        f"[openapi-diff] tool=wrapper-{WRAPPER_VERSION} "
        f"baseline={baseline_path.relative_to(repo_root) if baseline_path.is_relative_to(repo_root) else baseline_path} "
        f"current={current_path.relative_to(repo_root) if current_path.is_relative_to(repo_root) else current_path}\n"
    )

    baseline_doc = _yaml_load(baseline_path)
    current_doc = _yaml_load(current_path)
    config = _load_diff_config(config_path)
    history_ref = _resolve_history_ref(repo_root, args.history_ref, config)

    findings = diff_documents(baseline_doc, current_doc)
    findings, whitelist_matches = apply_response_status_whitelist(findings, config)

    if whitelist_matches:
        gate_finding = evaluate_history_gate(
            repo_root, history_path, history_ref, whitelist_matches
        )
        if gate_finding is not None:
            findings.append(gate_finding)

    summary = _format_summary(findings)

    output = OrderedDict(
        tool=f"wrapper-{WRAPPER_VERSION}",
        baseline=str(baseline_path),
        current=str(current_path),
        config=str(config_path) if config_path.is_file() else None,
        history=str(history_path) if history_path.is_file() else None,
        historyRef=history_ref,
        historyRefInput=args.history_ref,
        summary=summary,
        whitelistMatches=whitelist_matches,
        findings=findings,
    )

    payload = json.dumps(output, ensure_ascii=False, indent=2)
    if args.output:
        Path(args.output).write_text(payload + "\n", encoding="utf-8")
    sys.stdout.write(payload + "\n")

    if args.fail_on_incompatible and summary["breaking"] > 0:
        return 1
    return 0


def build_parser() -> argparse.ArgumentParser:
    here = Path(__file__).resolve()
    default_root = here.parent.parent.parent
    parser = argparse.ArgumentParser(description="openapi-v1-contract breaking-change gate")
    parser.add_argument("--repo-root", default=str(default_root))
    parser.add_argument("--baseline", default=None)
    parser.add_argument("--baseline-version", default=None)
    parser.add_argument("--current", default=None)
    parser.add_argument("--config", default=None)
    parser.add_argument("--history", default=None)
    parser.add_argument("--history-ref", default="auto")
    parser.add_argument(
        "--fail-on-incompatible",
        dest="fail_on_incompatible",
        action="store_true",
        default=True,
    )
    parser.add_argument(
        "--no-fail-on-incompatible",
        dest="fail_on_incompatible",
        action="store_false",
    )
    parser.add_argument("--output", default=None)
    parser.add_argument(
        "--print-default-baseline", action="store_true", default=False
    )
    return parser


def main(argv: Optional[List[str]] = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return run(args)


if __name__ == "__main__":
    raise SystemExit(main())
