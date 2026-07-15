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
import copy
import hashlib
import json
import re
import subprocess
import sys
from collections import OrderedDict
from collections import Counter
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
    configured_base = tooling.get("historyDiffBase") or "main"
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


def _operation_count(doc: Dict[str, Any]) -> int:
    count = 0
    for _path, methods in (doc.get("paths") or {}).items():
        if not isinstance(methods, dict):
            continue
        for method, op in methods.items():
            if method in HTTP_METHODS and isinstance(op, dict) and op.get("operationId"):
                count += 1
    return count


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


OPENAPI_001_CONDITIONAL_CONTRACT = (
    "baseline-required-fields-sourceReportId-forbidden|"
    "derived-retry-next-sourceReportId-required-nonnull-only"
)
OPENAPI_001_FINDING_KEYS = ("severity", "path", "kind", "before", "after")
OPENAPI_002_SOURCE_SCHEMAS = (
    "TargetJobImportSourceURL",
    "TargetJobImportSourceManualText",
    "TargetJobImportSourceFile",
    "TargetJobImportSourceManualForm",
    "TargetJobImportSource",
)
CONSTRAINT_KEYS = (
    "minLength",
    "maxLength",
    "minItems",
    "maxItems",
    "uniqueItems",
    "pattern",
)


def _json_pointer(*parts: str) -> str:
    return "/" + "/".join(part.replace("~", "~0").replace("/", "~1") for part in parts)


def _type_signature(schema: Any) -> str:
    if not isinstance(schema, dict):
        return "unknown"
    ref = schema.get("$ref")
    if isinstance(ref, str):
        return ref.rsplit("/", 1)[-1]
    one_of = schema.get("oneOf")
    if isinstance(one_of, list):
        signatures = [_type_signature(branch) for branch in one_of]
        if signatures:
            return "|".join(signatures)
    schema_type = schema.get("type")
    if schema_type == "array":
        return f"array<{_type_signature(schema.get('items') or {})}>"
    if isinstance(schema_type, list):
        return "|".join(str(value) for value in schema_type)
    return str(schema_type or "unknown")


def _format_constraints(schema: Any) -> str:
    if not isinstance(schema, dict):
        return ""
    if isinstance(schema.get("oneOf"), list):
        for branch in schema["oneOf"]:
            if isinstance(branch, dict) and branch.get("type") != "null":
                nested = _format_constraints(branch)
                if nested:
                    return nested
    fragments: List[str] = []
    for key in CONSTRAINT_KEYS:
        if key not in schema:
            continue
        value = schema[key]
        if isinstance(value, bool):
            value = str(value).lower()
        fragments.append(f"{key}={value}")
    return ",".join(fragments)


def _property_with_ready_constraints(
    parent_schema: Dict[str, Any], property_name: str, property_schema: Any
) -> Any:
    if not isinstance(property_schema, dict):
        return property_schema
    merged = dict(property_schema)
    for clause in parent_schema.get("allOf") or []:
        then_properties = ((clause.get("then") or {}).get("properties") or {})
        conditional = then_properties.get(property_name)
        if not isinstance(conditional, dict):
            continue
        for key in CONSTRAINT_KEYS:
            if key in conditional:
                merged[key] = conditional[key]
    return merged


def _normalized_finding(
    severity: str,
    path: str,
    kind: str,
    before: Any,
    after: Any,
) -> Dict[str, Any]:
    return {
        "severity": severity,
        "path": path,
        "kind": kind,
        "before": before,
        "after": after,
    }


def _compact_schema(schema: Any) -> str:
    return json.dumps(schema, ensure_ascii=False, sort_keys=True, separators=(",", ":"))


def normalize_openapi_001_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize the structural surface governed by OPENAPI-001.

    This intentionally ignores descriptions and examples. It records every
    component-schema add/remove, top-level property/required/closure change,
    supported constraint tightening, composition addition and enum widening so
    an exact-set oracle cannot hide an extra contract mutation.
    """

    findings: List[Dict[str, Any]] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})

    for name, base_schema in baseline_schemas.items():
        schema_path = _json_pointer("components", "schemas", name)
        if name not in current_schemas:
            findings.append(
                _normalized_finding("breaking", schema_path, "schema_removed", "present", "absent")
            )
            continue

        current_schema = current_schemas[name]
        if not isinstance(base_schema, dict) or not isinstance(current_schema, dict):
            continue

        base_properties = base_schema.get("properties") or {}
        current_properties = current_schema.get("properties") or {}
        base_required = list(base_schema.get("required") or [])
        current_required = list(current_schema.get("required") or [])
        current_required_set = set(current_required)

        for property_name, base_property in base_properties.items():
            property_path = _json_pointer(
                "components", "schemas", name, "properties", property_name
            )
            if property_name not in current_properties:
                findings.append(
                    _normalized_finding(
                        "breaking",
                        property_path,
                        "property_removed",
                        _type_signature(base_property),
                        "absent",
                    )
                )
                continue
            current_property = current_properties[property_name]
            before_constraints = _format_constraints(base_property)
            after_constraints = _format_constraints(
                _property_with_ready_constraints(
                    current_schema, property_name, current_property
                )
            )
            if before_constraints != after_constraints and after_constraints:
                findings.append(
                    _normalized_finding(
                        "breaking",
                        property_path,
                        "constraint_added",
                        before_constraints or "none",
                        after_constraints,
                    )
                )

        for property_name, current_property in current_properties.items():
            if property_name in base_properties:
                continue
            property_path = _json_pointer(
                "components", "schemas", name, "properties", property_name
            )
            if property_name in current_required_set:
                signature = _type_signature(current_property)
                kind = (
                    "required_nullable_property_added"
                    if "null" in signature.split("|")
                    else "required_property_added"
                )
                findings.append(
                    _normalized_finding("breaking", property_path, kind, "absent", signature)
                )
            else:
                findings.append(
                    _normalized_finding(
                        "additive",
                        property_path,
                        "property_added",
                        "absent",
                        _type_signature(current_property),
                    )
                )
            constraints = _format_constraints(
                _property_with_ready_constraints(
                    current_schema, property_name, current_property
                )
            )
            if constraints:
                constraint_before = (
                    "none"
                    if "null" in _type_signature(current_property).split("|")
                    else "absent"
                )
                findings.append(
                    _normalized_finding(
                        "breaking",
                        property_path,
                        "constraint_added",
                        constraint_before,
                        constraints,
                    )
                )

        if base_required != current_required and name in {
            "CreatePracticePlanRequest",
            "FeedbackReport",
        }:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", name, "required"),
                    "required_set_changed",
                    ",".join(base_required),
                    ",".join(current_required),
                )
            )

        if base_schema.get("additionalProperties", "unspecified") != current_schema.get(
            "additionalProperties", "unspecified"
        ):
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", name, "additionalProperties"),
                    "closed_object",
                    base_schema.get("additionalProperties", "unspecified"),
                    current_schema.get("additionalProperties", "unspecified"),
                )
            )

        if "oneOf" not in base_schema and "oneOf" in current_schema:
            after = (
                OPENAPI_001_CONDITIONAL_CONTRACT
                if name == "CreatePracticePlanRequest"
                else "present"
            )
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", name, "oneOf"),
                    "conditional_contract_added",
                    "absent",
                    after,
                )
            )

        base_enum = base_schema.get("enum")
        current_enum = current_schema.get("enum")
        if isinstance(base_enum, list) and isinstance(current_enum, list):
            for value in current_enum:
                if value not in base_enum:
                    findings.append(
                        _normalized_finding(
                            "additive",
                            _json_pointer("components", "schemas", name, "enum"),
                            "enum_value_added",
                            "absent",
                            str(value),
                        )
                    )
            for value in base_enum:
                if value not in current_enum:
                    findings.append(
                        _normalized_finding(
                            "breaking",
                            _json_pointer("components", "schemas", name, "enum"),
                            "enum_value_removed",
                            str(value),
                            "absent",
                        )
                    )

    for name, current_schema in current_schemas.items():
        if name in baseline_schemas:
            continue
        schema_path = _json_pointer("components", "schemas", name)
        required = list(current_schema.get("required") or []) if isinstance(current_schema, dict) else []
        findings.append(
            _normalized_finding(
                "additive",
                schema_path,
                "schema_added_with_required_fields" if required else "schema_added",
                "absent",
                ",".join(required) if required else "present",
            )
        )
        if isinstance(current_schema, dict) and current_schema.get("additionalProperties") is False:
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer("components", "schemas", name, "additionalProperties"),
                    "closed_object",
                    "absent",
                    False,
                )
            )

    return findings


def _openapi_001_v17_property_signature(schema: Any) -> str:
    """Return a compact, stable signature for the new closed read projection."""
    if not isinstance(schema, dict):
        return "unknown"
    ref = schema.get("$ref")
    if isinstance(ref, str):
        return ref.rsplit("/", 1)[-1]
    enum = schema.get("enum")
    if isinstance(enum, list):
        return f"enum({','.join(str(value) for value in enum)})"
    schema_type = schema.get("type")
    if schema_type == "array":
        return f"array<{_openapi_001_v17_property_signature(schema.get('items') or {})}>"
    fragments: List[str] = []
    if schema.get("format"):
        fragments.append(f"format={schema['format']}")
    for key in ("minimum", "minLength", "maxLength", "pattern"):
        if key in schema:
            fragments.append(f"{key}={schema[key]}")
    return f"{schema_type}({','.join(fragments)})" if fragments else str(schema_type or "unknown")


def normalize_openapi_001_v17_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize only OPENAPI-001 v1.7's session-list replacement delta.

    Earlier OPENAPI-001 report semantics are already in the merge-base
    baseline. The v1.7 correction must therefore record only the one-for-one
    operation swap and its newly introduced closed projection schemas.
    """
    # Do not reuse normalize_openapi_001_findings here. That historical
    # normalizer intentionally projects ready-state constraints from a
    # FeedbackReport conditional branch even when the baseline and current
    # documents are identical. v1.7 owns only the session-list replacement,
    # so inheriting those historical projections would silently authorize
    # unrelated report-semantic drift.
    findings: List[Dict[str, Any]] = []
    baseline_paths = baseline.get("paths") or {}
    current_paths = current.get("paths") or {}
    baseline_list = ((baseline_paths.get("/practice/sessions") or {}).get("get")) or {}
    current_list = ((current_paths.get("/practice/sessions") or {}).get("get")) or {}
    if baseline_list and not current_list:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("paths", "/practice/sessions", "get"),
                "operation_removed",
                str(baseline_list.get("operationId") or "present"),
                "absent",
            )
        )

    conversation_path = "/reports/{reportId}/conversation"
    baseline_conversation = ((baseline_paths.get(conversation_path) or {}).get("get")) or {}
    current_conversation = ((current_paths.get(conversation_path) or {}).get("get")) or {}
    if not baseline_conversation and current_conversation:
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("paths", conversation_path, "get"),
                "operation_added",
                "absent",
                str(current_conversation.get("operationId") or "present"),
            )
        )

    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})
    if "PaginatedPracticeSession" in baseline_schemas and "PaginatedPracticeSession" not in current_schemas:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "PaginatedPracticeSession"),
                "schema_removed",
                "present",
                "absent",
            )
        )
    for schema_name in ("ReportConversation", "ReportConversationMessage"):
        if schema_name in baseline_schemas:
            continue
        schema = current_schemas.get(schema_name)
        if not isinstance(schema, dict):
            continue
        required = list(schema.get("required") or [])
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("components", "schemas", schema_name),
                "schema_added_with_required_fields" if required else "schema_added",
                "absent",
                ",".join(required) if required else "present",
            )
        )
        if schema.get("additionalProperties") is False:
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer("components", "schemas", schema_name, "additionalProperties"),
                    "closed_object",
                    "absent",
                    False,
                )
            )
        for property_name, property_schema in (schema.get("properties") or {}).items():
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer(
                        "components", "schemas", schema_name, "properties", property_name
                    ),
                    "property_added",
                    "absent",
                    _openapi_001_v17_property_signature(property_schema),
                )
            )
    return findings


def _openapi_001_v17_response_schema(operation: Dict[str, Any], status: str) -> Any:
    return (
        (((((operation.get("responses") or {}).get(status) or {}).get("content") or {})
          .get("application/json") or {}).get("schema") or {})
    ).get("$ref")


def validate_openapi_001_v17_contract(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[str]:
    """Validate the closed, report-owned read contract and unchanged live APIs."""
    errors: List[str] = []
    baseline_paths = baseline.get("paths") or {}
    paths = current.get("paths") or {}
    baseline_sessions = baseline_paths.get("/practice/sessions") or {}
    sessions = paths.get("/practice/sessions") or {}
    if ((baseline_sessions.get("get") or {}).get("operationId")) != "listPracticeSessions":
        errors.append("merge-base must retain the legacy public session list")
    if "get" in sessions:
        errors.append("public session list GET /practice/sessions must be removed")
    if sessions.get("post") != baseline_sessions.get("post"):
        errors.append("startPracticeSession must remain byte-equivalent to the merge-base")
    session_detail_path = "/practice/sessions/{sessionId}"
    if (paths.get(session_detail_path) or {}).get("get") != (
        baseline_paths.get(session_detail_path) or {}
    ).get("get"):
        errors.append("getPracticeSession must remain byte-equivalent to the merge-base")

    operation = ((paths.get("/reports/{reportId}/conversation") or {}).get("get")) or {}
    if operation.get("operationId") != "getReportConversation":
        errors.append("getReportConversation operationId must be exact")
    if operation.get("tags") != ["Reports"]:
        errors.append("getReportConversation must be owned by Reports")
    if "requestBody" in operation:
        errors.append("getReportConversation must be read-only without requestBody")
    named_parameters = [
        parameter.get("name")
        for parameter in operation.get("parameters") or []
        if isinstance(parameter, dict) and "name" in parameter
    ]
    if named_parameters != ["reportId"]:
        errors.append("getReportConversation must accept only reportId as a named parameter")
    shared_refs = [
        parameter.get("$ref")
        for parameter in operation.get("parameters") or []
        if isinstance(parameter, dict) and "$ref" in parameter
    ]
    if shared_refs != [
        "#/components/parameters/XRequestID",
        "#/components/parameters/Traceparent",
        "#/components/parameters/AcceptLanguage",
        "#/components/parameters/XClientVersion",
    ]:
        errors.append("getReportConversation shared parameters must remain the protected read set")
    if _openapi_001_v17_response_schema(operation, "200") != "#/components/schemas/ReportConversation":
        errors.append("getReportConversation 200 must return ReportConversation")
    if _openapi_001_v17_response_schema(operation, "404") != "#/components/schemas/ApiErrorResponse":
        errors.append("getReportConversation 404 must use ApiErrorResponse")
    default_response = (operation.get("responses") or {}).get("default") or {}
    if default_response.get("$ref") != "#/components/responses/ApiErrorResponse":
        errors.append("getReportConversation default must reference ApiErrorResponse")

    current_operation_count = _operation_count(current)
    baseline_operation_count = _operation_count(baseline)
    if current_operation_count != 37 or baseline_operation_count != 37:
        errors.append(
            "report conversation correction must preserve 37 operations in both merge-base and current"
        )
    baseline_tags = [tag.get("name") for tag in baseline.get("tags") or []]
    current_tags = [tag.get("name") for tag in current.get("tags") or []]
    if baseline_tags != current_tags or len(current_tags) != 10 or len(set(current_tags)) != 10:
        errors.append("report conversation correction must preserve the exact 10-tag inventory")

    schemas = ((current.get("components") or {}).get("schemas") or {})
    if "PaginatedPracticeSession" in schemas:
        errors.append("PaginatedPracticeSession must be removed with the public session list")
    expected_shapes = {
        "ReportConversation": {
            "required": ["reportId", "reportStatus", "context", "messages"],
            "properties": {
                "reportId": {"type": "string", "format": "uuid"},
                "reportStatus": {"$ref": "#/components/schemas/ReportStatus"},
                "context": {"$ref": "#/components/schemas/ReportContextSnapshot"},
                "messages": {
                    "type": "array",
                    "items": {"$ref": "#/components/schemas/ReportConversationMessage"},
                },
            },
        },
        "ReportConversationMessage": {
            "required": ["sequence", "role", "content", "createdAt"],
            "properties": {
                "sequence": {"type": "integer", "format": "int32", "minimum": 1},
                "role": {"type": "string", "enum": ["user", "assistant"]},
                "content": {"type": "string", "minLength": 1, "pattern": r"\S"},
                "createdAt": {"type": "string", "format": "date-time"},
            },
        },
    }
    for schema_name, expected in expected_shapes.items():
        schema = schemas.get(schema_name) or {}
        if schema.get("type") != "object" or schema.get("additionalProperties") is not False:
            errors.append(f"{schema_name} must be a closed object")
        if schema.get("required") != expected["required"]:
            errors.append(f"{schema_name} required fields must be exact")
        if schema.get("properties") != expected["properties"]:
            errors.append(f"{schema_name} properties must be exact")
    message_properties = ((schemas.get("ReportConversationMessage") or {}).get("properties") or {})
    for locator in (
        "sessionId",
        "id",
        "clientMessageId",
        "replyStatus",
        "replyGeneration",
        "anchor",
    ):
        if locator in message_properties:
            errors.append(f"ReportConversationMessage internal locator {locator!r} is forbidden")
    return errors


OPENAPI_001_V17_AUTHORITY = {
    "decision": "OPENAPI-001",
    "decisionVersion": "1.7",
    "specDecision": "D-21",
    "historyVersion": "1.61",
    "productDecision": "report-owned-conversation",
}

OPENAPI_001_V17_INVARIANTS = {
    "inventory": {"operations": 37, "tags": 10},
    "startPracticeSession": {
        "path": "/api/v1/practice/sessions",
        "method": "POST",
        "operationId": "startPracticeSession",
        "successStatus": 201,
        "response": "PracticeSession",
    },
    "getPracticeSession": {
        "path": "/api/v1/practice/sessions/{sessionId}",
        "method": "GET",
        "operationId": "getPracticeSession",
        "successStatus": 200,
        "response": "PracticeSession",
    },
    "getReportConversation": {
        "path": "/api/v1/reports/{reportId}/conversation",
        "method": "GET",
        "operationId": "getReportConversation",
        "successStatus": 200,
        "response": "ReportConversation",
    },
    "publicList": "forbidden",
}


def build_openapi_001_v17_oracle(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> Dict[str, Any]:
    """Build the deterministic, merge-base-only v1.7 exact-set oracle."""
    return OrderedDict(
        schemaVersion=1,
        decisionId="OPENAPI-001",
        baseline="openapi/baseline/openapi-v1.0.0.yaml@merge-base(main)",
        proposed="openapi/openapi.yaml@working-tree",
        comparison=OrderedDict(
            mode="exact-set",
            keyFields=list(OPENAPI_001_FINDING_KEYS),
            orderSignificant=False,
            missingFinding="fail",
            unexpectedFinding="fail",
        ),
        authority=copy.deepcopy(OPENAPI_001_V17_AUTHORITY),
        invariants=copy.deepcopy(OPENAPI_001_V17_INVARIANTS),
        findings=sorted(normalize_openapi_001_v17_findings(baseline, current), key=_finding_key),
    )


def validate_openapi_001_v17_authority(
    repo_root: Path, oracle: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    if oracle.get("authority") != OPENAPI_001_V17_AUTHORITY:
        errors.append(
            "OPENAPI-001 v1.7 authority must bind D-21, history 1.61 and report-owned conversation"
        )

    contract_dir = repo_root / "docs" / "spec" / "openapi-v1-contract"
    spec_path = contract_dir / "spec.md"
    history_path = contract_dir / "history.md"
    decision_path = contract_dir / "decisions" / "OPENAPI-001-report-direct-semantics.md"
    spec_text = spec_path.read_text(encoding="utf-8") if spec_path.is_file() else ""
    history_text = history_path.read_text(encoding="utf-8") if history_path.is_file() else ""
    decision_text = decision_path.read_text(encoding="utf-8") if decision_path.is_file() else ""
    spec_rows = [line for line in spec_text.splitlines() if re.search(r"\|\s*D-21\s*\|", line)]
    if not any(
        "getReportConversation" in line and "listPracticeSessions" in line
        for line in spec_rows
    ):
        errors.append("OPENAPI-001 v1.7 requires current spec D-21 authority")
    history_rows = [
        line for line in history_text.splitlines() if re.search(r"\|\s*1\.61\s*\|", line)
    ]
    if not any(
        "getReportConversation" in line and "listPracticeSessions" in line
        for line in history_rows
    ):
        errors.append("OPENAPI-001 v1.7 requires history 1.61 authority")
    if "Report-owned conversation locator correction" not in decision_text:
        errors.append("OPENAPI-001 decision record must preserve report-owned conversation authority")
    return errors


def validate_openapi_001_v17_invariants(
    current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    if invariants != OPENAPI_001_V17_INVARIANTS:
        errors.append("OPENAPI-001 v1.7 oracle invariants must remain exact")
        return errors
    inventory = invariants["inventory"]
    if _operation_count(current) != inventory["operations"]:
        errors.append(
            f"OPENAPI-001 v1.7 operations must equal {inventory['operations']}"
        )
    tag_names = [
        str(tag.get("name"))
        for tag in current.get("tags") or []
        if isinstance(tag, dict) and tag.get("name")
    ]
    if len(tag_names) != inventory["tags"] or len(set(tag_names)) != inventory["tags"]:
        errors.append(f"OPENAPI-001 v1.7 tags must equal {inventory['tags']} unique entries")
    for invariant_name in (
        "startPracticeSession",
        "getPracticeSession",
        "getReportConversation",
    ):
        expected = invariants[invariant_name]
        operation = _operation_at_contract_path(
            current, expected["path"], expected["method"]
        )
        if operation is None:
            errors.append(
                f"{invariant_name} must remain {expected['method']} {expected['path']}"
            )
            continue
        if operation.get("operationId") != expected["operationId"]:
            errors.append(f"{invariant_name} operationId must equal {expected['operationId']}")
        if _success_response_schema(operation, expected["successStatus"]) != expected["response"]:
            errors.append(
                f"{invariant_name} response {expected['successStatus']} must equal {expected['response']}"
            )
    if invariants["publicList"] == "forbidden" and "get" in (
        (current.get("paths") or {}).get("/practice/sessions") or {}
    ):
        errors.append("OPENAPI-001 v1.7 public session list must remain forbidden")
    return errors


def _openapi_002_property_signature(schema: Any) -> str:
    if not isinstance(schema, dict):
        return "unknown"
    enum = schema.get("enum")
    if isinstance(enum, list):
        return f"enum({','.join(str(value) for value in enum)})"
    signature = _type_signature(schema)
    constraints = _format_constraints(schema)
    return f"{signature}({constraints})" if constraints else signature


def normalize_openapi_002_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize the OPENAPI-002 TargetJob paste-only correction surface.

    The Practice message correction is intentionally audited by its separate
    decision gate. This normalizer covers every OPENAPI-002-owned schema
    surface plus ApiErrorCode deltas sourced by the TargetJob cleanup. A newly
    required property carries its initial constraints in one signature, so
    rawText does not also produce a second constraint_added finding.
    """

    findings: List[Dict[str, Any]] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})

    for name in OPENAPI_002_SOURCE_SCHEMAS:
        if name in baseline_schemas and name not in current_schemas:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", name),
                    "schema_removed",
                    "present",
                    "absent",
                )
            )

    for name in ("ImportTargetJobRequest", "TargetJob"):
        base_schema = baseline_schemas.get(name)
        current_schema = current_schemas.get(name)
        schema_path = _json_pointer("components", "schemas", name)
        if not isinstance(base_schema, dict):
            continue
        if not isinstance(current_schema, dict):
            findings.append(
                _normalized_finding(
                    "breaking", schema_path, "schema_removed", "present", "absent"
                )
            )
            continue

        base_properties = base_schema.get("properties") or {}
        current_properties = current_schema.get("properties") or {}
        current_required = set(current_schema.get("required") or [])
        owned_properties = (
            {"source", "titleHint", "companyNameHint", "rawText"}
            if name == "ImportTargetJobRequest"
            else {"sourceType", "sourceUrl"}
        )

        for property_name, base_property in base_properties.items():
            if property_name not in owned_properties:
                continue
            property_path = _json_pointer(
                "components", "schemas", name, "properties", property_name
            )
            if property_name not in current_properties:
                findings.append(
                    _normalized_finding(
                        "breaking",
                        property_path,
                        "property_removed",
                        _openapi_002_property_signature(base_property),
                        "absent",
                    )
                )
                continue
            current_property = current_properties[property_name]
            before_signature = _openapi_002_property_signature(base_property)
            after_signature = _openapi_002_property_signature(current_property)
            if before_signature != after_signature:
                findings.append(
                    _normalized_finding(
                        "breaking",
                        property_path,
                        "property_changed",
                        before_signature,
                        after_signature,
                    )
                )

        for property_name, current_property in current_properties.items():
            if property_name not in owned_properties:
                continue
            if property_name in base_properties:
                continue
            property_path = _json_pointer(
                "components", "schemas", name, "properties", property_name
            )
            required = property_name in current_required
            findings.append(
                _normalized_finding(
                    "breaking" if required else "additive",
                    property_path,
                    "required_property_added" if required else "property_added",
                    "absent",
                    _openapi_002_property_signature(current_property),
                )
            )

        base_required = list(base_schema.get("required") or [])
        current_required_list = list(current_schema.get("required") or [])
        if base_required != current_required_list:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", name, "required"),
                    "required_set_changed",
                    ",".join(base_required),
                    ",".join(current_required_list),
                )
            )

        base_closure = base_schema.get("additionalProperties", "unspecified")
        current_closure = current_schema.get("additionalProperties", "unspecified")
        if base_closure != current_closure:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer(
                        "components", "schemas", name, "additionalProperties"
                    ),
                    "closed_object",
                    base_closure,
                    current_closure,
                )
            )

    base_upload_purpose = (
        (((baseline_schemas.get("UploadPresignRequest") or {}).get("properties") or {}).get("purpose") or {}).get("enum")
        or []
    )
    current_upload_purpose = (
        (((current_schemas.get("UploadPresignRequest") or {}).get("properties") or {}).get("purpose") or {}).get("enum")
        or []
    )
    if list(base_upload_purpose) != list(current_upload_purpose):
        removed = any(value not in current_upload_purpose for value in base_upload_purpose)
        findings.append(
            _normalized_finding(
                "breaking" if removed else "additive",
                _json_pointer(
                    "components",
                    "schemas",
                    "UploadPresignRequest",
                    "properties",
                    "purpose",
                    "enum",
                ),
                "enum_value_removed" if removed else "enum_value_added",
                ",".join(str(value) for value in base_upload_purpose),
                ",".join(str(value) for value in current_upload_purpose),
            )
        )

    base_error_codes = list((baseline_schemas.get("ApiErrorCode") or {}).get("enum") or [])
    current_error_codes = list((current_schemas.get("ApiErrorCode") or {}).get("enum") or [])
    error_path = _json_pointer("components", "schemas", "ApiErrorCode", "enum")
    for value in base_error_codes:
        if value not in current_error_codes:
            findings.append(
                _normalized_finding(
                    "breaking", error_path, "enum_value_removed", str(value), "absent"
                )
            )
    for value in current_error_codes:
        if value not in base_error_codes:
            findings.append(
                _normalized_finding(
                    "additive", error_path, "enum_value_added", "absent", str(value)
                )
            )

    return findings


def normalize_d_35_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize only the Practice durable-message correction surface.

    D-35 is deliberately separate from OPENAPI-002: TargetJob findings
    can neither authorize nor hide a Practice contract mutation.
    """

    findings: List[Dict[str, Any]] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})
    base_message = baseline_schemas.get("PracticeMessage") or {}
    current_message = current_schemas.get("PracticeMessage") or {}
    base_required = list(base_message.get("required") or [])
    base_properties = base_message.get("properties") or {}

    if "oneOf" not in base_message and "oneOf" in current_message:
        branches = current_message.get("oneOf") or []
        branch_names = [_type_signature(branch) for branch in branches]
        discriminator = (current_message.get("discriminator") or {}).get(
            "propertyName"
        )
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "PracticeMessage", "oneOf"),
                "role_discriminated_union_added",
                "absent",
                f"user={branch_names[0] if branch_names else 'missing'},"
                f"assistant={branch_names[1] if len(branch_names) > 1 else 'missing'};"
                f"discriminator={discriminator or 'missing'}",
            )
        )

    for keyword in ("type", "properties", "required", "additionalProperties"):
        if keyword in current_message:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", "PracticeMessage", keyword),
                    "legacy_keyword_retained",
                    "absent",
                    _compact_schema(current_message[keyword]),
                )
            )

    user_schema = current_schemas.get("PracticeUserMessage")
    assistant_schema = current_schemas.get("PracticeAssistantMessage")
    base_user_schema = baseline_schemas.get("PracticeUserMessage")
    base_assistant_schema = baseline_schemas.get("PracticeAssistantMessage")
    if isinstance(user_schema, dict):
        user_properties = user_schema.get("properties") or {}
        user_required = list(user_schema.get("required") or [])
        if isinstance(base_user_schema, dict):
            expected_user_properties = base_user_schema.get("properties") or {}
            expected_user_required = list(base_user_schema.get("required") or [])
            expected_user_closure = base_user_schema.get(
                "additionalProperties", "unspecified"
            )
        else:
            expected_user_properties = copy.deepcopy(base_properties)
            if "role" in expected_user_properties:
                expected_user_properties["role"] = {
                    "type": "string",
                    "enum": ["user"],
                }
            expected_user_properties.update(
                {
                    "clientMessageId": {"type": "string", "format": "uuid"},
                    "replyStatus": {"$ref": "#/components/schemas/PracticeReplyStatus"},
                }
            )
            expected_user_required = base_required
            expected_user_closure = "unspecified"
            for property_name in ("clientMessageId", "replyStatus"):
                if property_name not in user_properties:
                    continue
                findings.append(
                    _normalized_finding(
                        "breaking",
                        _json_pointer(
                            "components",
                            "schemas",
                            "PracticeUserMessage",
                            "properties",
                            property_name,
                        ),
                        (
                            "required_property_added"
                            if property_name in user_required
                            else "property_added"
                        ),
                        "absent",
                        _type_signature(user_properties[property_name]),
                    )
                )
        for property_name in sorted(set(user_properties) | set(expected_user_properties)):
            if property_name in {"clientMessageId", "replyStatus"}:
                continue
            expected_property = expected_user_properties.get(property_name)
            current_property = user_properties.get(property_name)
            if expected_property != current_property:
                findings.append(
                    _normalized_finding(
                        "breaking" if expected_property is not None else "additive",
                        _json_pointer(
                            "components",
                            "schemas",
                            "PracticeUserMessage",
                            "properties",
                            property_name,
                        ),
                        "property_changed" if expected_property is not None else "property_added",
                        _compact_schema(expected_property) if expected_property is not None else "absent",
                        _compact_schema(current_property) if current_property is not None else "absent",
                    )
                )
        if expected_user_required != user_required:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer(
                        "components", "schemas", "PracticeUserMessage", "required"
                    ),
                    "required_set_changed",
                    ",".join(expected_user_required),
                    ",".join(user_required),
                )
            )
        current_user_closure = user_schema.get("additionalProperties", "unspecified")
        if expected_user_closure != current_user_closure:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer(
                        "components",
                        "schemas",
                        "PracticeUserMessage",
                        "additionalProperties",
                    ),
                    "closed_object",
                    expected_user_closure,
                    current_user_closure,
                )
            )

    if isinstance(assistant_schema, dict):
        assistant_properties = assistant_schema.get("properties") or {}
        if isinstance(base_assistant_schema, dict):
            expected_assistant_properties = base_assistant_schema.get("properties") or {}
            expected_assistant_closure = base_assistant_schema.get(
                "additionalProperties", "unspecified"
            )
        else:
            expected_assistant_properties = copy.deepcopy(base_properties)
            if "role" in expected_assistant_properties:
                expected_assistant_properties["role"] = {
                    "type": "string",
                    "enum": ["assistant"],
                }
            expected_assistant_closure = "unspecified"
        current_assistant_closure = assistant_schema.get(
            "additionalProperties", "unspecified"
        )
        if expected_assistant_closure != current_assistant_closure:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer(
                        "components",
                        "schemas",
                        "PracticeAssistantMessage",
                        "additionalProperties",
                    ),
                    "closed_object",
                    expected_assistant_closure,
                    current_assistant_closure,
                )
            )
        for property_name in sorted(
            set(assistant_properties) | set(expected_assistant_properties)
        ):
            expected_property = expected_assistant_properties.get(property_name)
            current_property = assistant_properties.get(property_name)
            if expected_property != current_property:
                findings.append(
                    _normalized_finding(
                        "breaking" if expected_property is not None else "additive",
                        _json_pointer(
                            "components",
                            "schemas",
                            "PracticeAssistantMessage",
                            "properties",
                            property_name,
                        ),
                        "property_changed" if expected_property is not None else "property_added",
                        _compact_schema(expected_property) if expected_property is not None else "absent",
                        _compact_schema(current_property) if current_property is not None else "absent",
                    )
                )

    base_response = baseline_schemas.get("SendPracticeMessageResponse") or {}
    current_response = current_schemas.get("SendPracticeMessageResponse") or {}
    base_response_properties = base_response.get("properties") or {}
    current_response_properties = current_response.get("properties") or {}
    for property_name in ("userMessage", "assistantMessage"):
        before = _type_signature(base_response_properties.get(property_name) or {})
        after = _type_signature(current_response_properties.get(property_name) or {})
        if before != after:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer(
                        "components",
                        "schemas",
                        "SendPracticeMessageResponse",
                        "properties",
                        property_name,
                    ),
                    "ref_changed",
                    before,
                    after,
                )
            )
    base_response_required = list(base_response.get("required") or [])
    current_response_required = list(current_response.get("required") or [])
    if base_response_required != current_response_required:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer(
                    "components", "schemas", "SendPracticeMessageResponse", "required"
                ),
                "required_set_changed",
                ",".join(base_response_required),
                ",".join(current_response_required),
            )
        )
    for property_name in sorted(
        set(base_response_properties) | set(current_response_properties)
    ):
        if property_name in {"userMessage", "assistantMessage"}:
            continue
        before_property = base_response_properties.get(property_name)
        after_property = current_response_properties.get(property_name)
        if before_property != after_property:
            findings.append(
                _normalized_finding(
                    "breaking" if before_property is not None else "additive",
                    _json_pointer(
                        "components",
                        "schemas",
                        "SendPracticeMessageResponse",
                        "properties",
                        property_name,
                    ),
                    "property_changed" if before_property is not None else "property_added",
                    _compact_schema(before_property) if before_property is not None else "absent",
                    _compact_schema(after_property) if after_property is not None else "absent",
                )
            )

    base_session = baseline_schemas.get("PracticeSession")
    current_session = current_schemas.get("PracticeSession")
    if base_session != current_session:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "PracticeSession"),
                "unchanged_schema_drifted",
                _compact_schema(base_session),
                _compact_schema(current_session),
            )
        )

    for name in (
        "PracticeReplyStatus",
        "PracticeUserMessage",
        "PracticeAssistantMessage",
    ):
        if name in baseline_schemas or name not in current_schemas:
            continue
        schema = current_schemas[name]
        required = list(schema.get("required") or []) if isinstance(schema, dict) else []
        if name == "PracticeReplyStatus" and isinstance(schema, dict):
            enum = schema.get("enum") or []
            kind = "schema_added"
            after = f"enum({','.join(str(value) for value in enum)})"
        else:
            kind = "schema_added_with_required_fields" if required else "schema_added"
            after = ",".join(required) if required else "present"
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("components", "schemas", name),
                kind,
                "absent",
                after,
            )
        )

    baseline_practice_paths = {
        path for path in (baseline.get("paths") or {}) if path.startswith("/practice/sessions")
    }
    current_practice_paths = {
        path for path in (current.get("paths") or {}) if path.startswith("/practice/sessions")
    }
    for path in sorted(baseline_practice_paths - current_practice_paths):
        findings.append(
            _normalized_finding(
                "breaking", _json_pointer("paths", path), "endpoint_removed", "present", "absent"
            )
        )
    for path in sorted(current_practice_paths - baseline_practice_paths):
        findings.append(
            _normalized_finding(
                "additive", _json_pointer("paths", path), "endpoint_added", "absent", "present"
            )
        )

    return findings


OPENAPI_004_REPORT_PATH = "/api/v1/targets/{targetJobId}/reports"
OPENAPI_004_NEW_SCHEMAS = (
    "TargetJobReportsOverview",
    "TargetJobReportRoundOverview",
    "TargetJobCurrentReportSummary",
    "TargetJobReportAttemptSummary",
)


def normalize_openapi_004_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize only the OPENAPI-004 report-overview correction surface."""

    findings: List[Dict[str, Any]] = []
    baseline_operation = _operation_at_contract_path(
        baseline, OPENAPI_004_REPORT_PATH, "GET"
    ) or {}
    current_operation = _operation_at_contract_path(
        current, OPENAPI_004_REPORT_PATH, "GET"
    ) or {}
    baseline_parameters = {
        parameter.get("name"): parameter
        for parameter in baseline_operation.get("parameters") or []
        if isinstance(parameter, dict) and parameter.get("name")
    }
    current_parameters = {
        parameter.get("name"): parameter
        for parameter in current_operation.get("parameters") or []
        if isinstance(parameter, dict) and parameter.get("name")
    }
    report_pointer = _json_pointer(
        "paths", "/targets/{targetJobId}/reports", "get"
    )
    for parameter_name in ("cursor", "pageSize"):
        parameter = baseline_parameters.get(parameter_name)
        if parameter is not None and parameter_name not in current_parameters:
            required = "required" if parameter.get("required") else "optional"
            findings.append(
                _normalized_finding(
                    "breaking",
                    f"{report_pointer}/parameters/{parameter_name}",
                    "parameter_removed",
                    f"{parameter.get('in', 'query')}:{required}",
                    "absent",
                )
            )

    baseline_response = _success_response_schema(baseline_operation, 200)
    current_response = _success_response_schema(current_operation, 200)
    if baseline_response != current_response:
        findings.append(
            _normalized_finding(
                "breaking",
                f"{report_pointer}/responses/200/content/application~1json/schema",
                "response_ref_changed",
                baseline_response or "absent",
                current_response or "absent",
            )
        )

    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})
    baseline_target = baseline_schemas.get("TargetJob") or {}
    current_target = current_schemas.get("TargetJob") or {}
    baseline_target_properties = baseline_target.get("properties") or {}
    current_target_properties = current_target.get("properties") or {}
    latest_report = baseline_target_properties.get("latestReportId")
    if latest_report is not None and "latestReportId" not in current_target_properties:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer(
                    "components", "schemas", "TargetJob", "properties", "latestReportId"
                ),
                "property_removed",
                _type_signature(latest_report),
                "absent",
            )
        )

    if (
        "PaginatedFeedbackReport" in baseline_schemas
        and "PaginatedFeedbackReport" not in current_schemas
    ):
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "PaginatedFeedbackReport"),
                "schema_removed",
                "present",
                "absent",
            )
        )

    baseline_round = baseline_schemas.get("PracticeRoundRef") or {}
    current_round = current_schemas.get("PracticeRoundRef") or {}
    baseline_closure = baseline_round.get("additionalProperties", "unspecified")
    current_closure = current_round.get("additionalProperties", "unspecified")
    if baseline_closure != current_closure:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer(
                    "components", "schemas", "PracticeRoundRef", "additionalProperties"
                ),
                "closed_object",
                baseline_closure,
                current_closure,
            )
        )

    for schema_name in OPENAPI_004_NEW_SCHEMAS:
        if schema_name in baseline_schemas or schema_name not in current_schemas:
            continue
        schema = current_schemas[schema_name]
        required = list(schema.get("required") or []) if isinstance(schema, dict) else []
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("components", "schemas", schema_name),
                "schema_added_with_required_fields" if required else "schema_added",
                "absent",
                ",".join(required) if required else "present",
            )
        )
        if isinstance(schema, dict) and schema.get("additionalProperties") is False:
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer(
                        "components", "schemas", schema_name, "additionalProperties"
                    ),
                    "closed_object",
                    "absent",
                    False,
                )
            )
        if schema_name == "TargetJobReportsOverview" and isinstance(schema, dict):
            rounds = (schema.get("properties") or {}).get("rounds") or {}
            bounds = _format_constraints(rounds)
            if bounds:
                findings.append(
                    _normalized_finding(
                        "additive",
                        _json_pointer(
                            "components",
                            "schemas",
                            schema_name,
                            "properties",
                            "rounds",
                        ),
                        "canonical_array_bounds_added",
                        "absent",
                        bounds,
                    )
                )
    return findings


def validate_openapi_004_contract(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    schemas = ((current.get("components") or {}).get("schemas") or {})
    if "PaginatedFeedbackReport" not in baseline_schemas:
        errors.append("old baseline must contain PaginatedFeedbackReport")
    if "PaginatedFeedbackReport" in schemas:
        errors.append("PaginatedFeedbackReport must be removed")
    baseline_target_properties = (
        (baseline_schemas.get("TargetJob") or {}).get("properties") or {}
    )
    target_properties = (schemas.get("TargetJob") or {}).get("properties") or {}
    if "latestReportId" not in baseline_target_properties:
        errors.append("old baseline TargetJob must contain latestReportId")
    if "latestReportId" in target_properties:
        errors.append("TargetJob latestReportId must be removed")

    baseline_round = baseline_schemas.get("PracticeRoundRef") or {}
    current_round = schemas.get("PracticeRoundRef") or {}
    normalized_round = copy.deepcopy(current_round)
    if normalized_round.pop("additionalProperties", None) is not False:
        errors.append("PracticeRoundRef must set additionalProperties=false")
    if normalized_round != baseline_round:
        errors.append("PracticeRoundRef may change only object closure")

    expected_shapes = {
        "TargetJobReportsOverview": (
            ["targetJobId", "rounds"],
            {"targetJobId", "rounds"},
            {"type", "required", "additionalProperties", "properties"},
        ),
        "TargetJobReportRoundOverview": (
            ["round", "currentReport", "latestAttempt"],
            {"round", "currentReport", "latestAttempt"},
            {"type", "required", "additionalProperties", "properties"},
        ),
        "TargetJobCurrentReportSummary": (
            ["id", "generatedAt"],
            {"id", "generatedAt"},
            {"type", "required", "additionalProperties", "properties"},
        ),
        "TargetJobReportAttemptSummary": (
            ["id", "status", "errorCode", "createdAt"],
            {"id", "status", "errorCode", "createdAt"},
            {"type", "required", "additionalProperties", "allOf", "properties"},
        ),
    }
    for schema_name, (required, property_names, keywords) in expected_shapes.items():
        schema = schemas.get(schema_name) or {}
        if schema.get("type") != "object" or schema.get("additionalProperties") is not False:
            errors.append(f"{schema_name} must be a closed object")
        if schema.get("required") != required:
            errors.append(f"{schema_name} required fields must equal {required}")
        if set(schema.get("properties") or {}) != property_names:
            errors.append(f"{schema_name} properties must equal {sorted(property_names)}")
        if set(schema) != keywords:
            errors.append(f"{schema_name} must not declare extra schema keywords")

    overview = (schemas.get("TargetJobReportsOverview") or {}).get("properties") or {}
    if overview.get("targetJobId") != {"type": "string", "format": "uuid"}:
        errors.append("TargetJobReportsOverview targetJobId must be UUID")
    rounds = overview.get("rounds") or {}
    if rounds != {
        "type": "array",
        "minItems": 2,
        "maxItems": 5,
        "items": {"$ref": "#/components/schemas/TargetJobReportRoundOverview"},
    }:
        errors.append("TargetJobReportsOverview rounds must be canonical 2..5 summaries")

    round_properties = (
        (schemas.get("TargetJobReportRoundOverview") or {}).get("properties") or {}
    )
    if round_properties.get("round") != {
        "$ref": "#/components/schemas/PracticeRoundRef"
    }:
        errors.append("report overview round must reference PracticeRoundRef")
    if [_type_signature(branch) for branch in (round_properties.get("currentReport") or {}).get("oneOf") or []] != [
        "TargetJobCurrentReportSummary",
        "null",
    ]:
        errors.append("currentReport must be required nullable current summary")
    if [_type_signature(branch) for branch in (round_properties.get("latestAttempt") or {}).get("oneOf") or []] != [
        "TargetJobReportAttemptSummary",
        "null",
    ]:
        errors.append("latestAttempt must be required nullable attempt summary")

    current_properties = (
        (schemas.get("TargetJobCurrentReportSummary") or {}).get("properties") or {}
    )
    if current_properties != {
        "id": {"type": "string", "format": "uuid"},
        "generatedAt": {"type": "string", "format": "date-time"},
    }:
        errors.append("current report summary fields must remain minimal and typed")

    attempt = schemas.get("TargetJobReportAttemptSummary") or {}
    attempt_properties = attempt.get("properties") or {}
    if attempt_properties != {
        "id": {"type": "string", "format": "uuid"},
        "status": {"$ref": "#/components/schemas/ReportStatus"},
        "errorCode": {
            "oneOf": [
                {"$ref": "#/components/schemas/ApiErrorCode"},
                {"type": "null"},
            ]
        },
        "createdAt": {"type": "string", "format": "date-time"},
    }:
        errors.append("latest attempt fields must remain minimal and typed")
    expected_conditional = [
        {
            "if": {
                "required": ["status"],
                "properties": {"status": {"const": "failed"}},
            },
            "then": {
                "properties": {
                    "errorCode": {"$ref": "#/components/schemas/ApiErrorCode"}
                }
            },
            "else": {"properties": {"errorCode": {"type": "null"}}},
        }
    ]
    if attempt.get("allOf") != expected_conditional:
        errors.append("latest attempt errorCode must be non-null only for failed status")
    return errors


def validate_openapi_004_invariants(
    baseline: Dict[str, Any], current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    inventory = invariants.get("inventory") or {}
    operation_count = _operation_count(current)
    tag_names = [
        str(tag.get("name"))
        for tag in current.get("tags") or []
        if isinstance(tag, dict) and tag.get("name")
    ]
    if operation_count != inventory.get("operations"):
        errors.append(
            f"OPENAPI-004 operations must equal {inventory.get('operations')}, got {operation_count}"
        )
    if len(tag_names) != inventory.get("tags") or len(set(tag_names)) != inventory.get(
        "tags"
    ):
        errors.append(
            f"OPENAPI-004 tags must equal {inventory.get('tags')} unique entries, got {len(tag_names)}"
        )

    expected = invariants.get("listTargetJobReports") or {}
    baseline_operation = _operation_at_contract_path(
        baseline, str(expected.get("path") or ""), str(expected.get("method") or "")
    )
    current_operation = _operation_at_contract_path(
        current, str(expected.get("path") or ""), str(expected.get("method") or "")
    )
    if baseline_operation is None or current_operation is None:
        errors.append("listTargetJobReports method/path must remain unchanged")
        return errors
    if baseline_operation.get("operationId") != expected.get(
        "operationId"
    ) or current_operation.get("operationId") != expected.get("operationId"):
        errors.append("listTargetJobReports operationId must remain unchanged")
    status = expected.get("successStatus")
    if _success_response_schema(baseline_operation, status) != expected.get(
        "baselineResponse"
    ):
        errors.append("old baseline listTargetJobReports response drifted")
    if _success_response_schema(current_operation, status) != expected.get("response"):
        errors.append("listTargetJobReports 200 response must equal TargetJobReportsOverview")
    if set(str(code) for code in baseline_operation.get("responses") or {}) != set(
        str(code) for code in current_operation.get("responses") or {}
    ):
        errors.append("listTargetJobReports response statuses must remain unchanged")
    expected_parameters = [
        parameter
        for parameter in baseline_operation.get("parameters") or []
        if not (
            isinstance(parameter, dict)
            and parameter.get("name") in {"cursor", "pageSize"}
        )
    ]
    if current_operation.get("parameters") != expected_parameters:
        errors.append("listTargetJobReports parameters may only remove cursor/pageSize")
    return errors


OPENAPI_005_LIST_PATH = "/api/v1/resumes"
OPENAPI_005_DETAIL_PATH = "/api/v1/resumes/{resumeId}"
OPENAPI_005_SUMMARY_FIELDS = (
    "id",
    "title",
    "displayName",
    "language",
    "sourceType",
    "parseStatus",
    "summaryHeadline",
    "hasReadableContent",
    "updatedAt",
)


def _openapi_005_paginated_item_ref(schemas: Dict[str, Any]) -> str:
    paginated = schemas.get("PaginatedResume") or {}
    all_of = paginated.get("allOf") or []
    if len(all_of) != 2 or not isinstance(all_of[1], dict):
        return ""
    items = ((all_of[1].get("properties") or {}).get("items") or {}).get("items") or {}
    return _type_signature(items)


def _openapi_005_property_signature(schema: Any) -> str:
    if not isinstance(schema, dict):
        return "unknown"
    enum = schema.get("enum")
    if isinstance(enum, list):
        return f"enum({','.join(str(value) for value in enum)})"
    signature = _type_signature(schema)
    schema_format = schema.get("format")
    if schema_format:
        return f"{signature}(format={schema_format})"
    return signature


def normalize_openapi_005_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize only OPENAPI-005's Resume list-summary correction."""
    findings: List[Dict[str, Any]] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})

    baseline_item_ref = _openapi_005_paginated_item_ref(baseline_schemas)
    current_item_ref = _openapi_005_paginated_item_ref(current_schemas)
    if baseline_item_ref != current_item_ref:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer(
                    "components", "schemas", "PaginatedResume", "properties", "items", "items"
                ),
                "response_item_ref_changed",
                baseline_item_ref or "absent",
                current_item_ref or "absent",
            )
        )

    if "ResumeSummary" not in baseline_schemas and "ResumeSummary" in current_schemas:
        summary = current_schemas["ResumeSummary"]
        required = list(summary.get("required") or []) if isinstance(summary, dict) else []
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("components", "schemas", "ResumeSummary"),
                "schema_added_with_required_fields" if required else "schema_added",
                "absent",
                ",".join(required) if required else "present",
            )
        )
        if isinstance(summary, dict) and summary.get("additionalProperties") is False:
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer(
                        "components", "schemas", "ResumeSummary", "additionalProperties"
                    ),
                    "closed_object",
                    "absent",
                    False,
                )
            )
        for name, schema in (summary.get("properties") or {}).items():
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer(
                        "components", "schemas", "ResumeSummary", "properties", name
                    ),
                    "property_added",
                    "absent",
                    _openapi_005_property_signature(schema),
                )
            )
    return findings


def validate_openapi_005_contract(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    schemas = ((current.get("components") or {}).get("schemas") or {})

    if "ResumeSummary" in baseline_schemas:
        errors.append("old baseline must not contain ResumeSummary")
    if _openapi_005_paginated_item_ref(baseline_schemas) != "Resume":
        errors.append("old baseline PaginatedResume.items must reference full Resume")
    if _openapi_005_paginated_item_ref(schemas) != "ResumeSummary":
        errors.append("PaginatedResume.items must reference ResumeSummary")

    expected_properties = {
        "id": {"type": "string", "format": "uuid"},
        "title": {"type": "string"},
        "displayName": {"type": "string"},
        "language": {"type": "string"},
        "sourceType": {"type": "string", "enum": ["upload", "paste"]},
        "parseStatus": {"$ref": "#/components/schemas/TargetJobParseStatus"},
        "summaryHeadline": {
            "oneOf": [{"type": "string"}, {"type": "null"}]
        },
        "hasReadableContent": {"type": "boolean"},
        "updatedAt": {"type": "string", "format": "date-time"},
    }
    expected_summary = {
        "type": "object",
        "additionalProperties": False,
        "required": list(OPENAPI_005_SUMMARY_FIELDS),
        "properties": expected_properties,
    }
    if schemas.get("ResumeSummary") != expected_summary:
        errors.append("ResumeSummary properties/required/types/closure must match OPENAPI-005")

    if schemas.get("Resume") != baseline_schemas.get("Resume"):
        errors.append("full Resume schema must remain byte-equivalent to the old baseline")
    normalized_paginated = copy.deepcopy(schemas.get("PaginatedResume") or {})
    all_of = normalized_paginated.get("allOf") or []
    if len(all_of) == 2 and isinstance(all_of[1], dict):
        items = (all_of[1].get("properties") or {}).get("items") or {}
        items["items"] = {"$ref": "#/components/schemas/Resume"}
    if normalized_paginated != baseline_schemas.get("PaginatedResume"):
        errors.append("PaginatedResume may change only its item ref to ResumeSummary")

    resume_names_added = {
        name
        for name in set(schemas) - set(baseline_schemas)
        if name.startswith("Resume")
    }
    if resume_names_added != {"ResumeSummary"}:
        errors.append(
            f"unexpected new Resume compatibility schemas: {sorted(resume_names_added)}"
        )

    paths = current.get("paths") or {}
    for path, method, status, operation_id in (
        ("/resumes/{resumeId}", "get", "200", "getResume"),
        ("/resumes/{resumeId}", "patch", "200", "updateResume"),
        ("/resumes/{resumeId}/duplicate", "post", "201", "duplicateResume"),
        ("/resumes/{resumeId}/archive", "post", "202", "archiveResume"),
    ):
        operation = ((paths.get(path) or {}).get(method) or {})
        if operation.get("operationId") != operation_id:
            errors.append(f"{operation_id} operation must remain present")
            continue
        response_ref = _success_response_schema(operation, status)
        if response_ref != "Resume":
            errors.append(f"{operation_id} {status} response must remain full Resume")
    return errors


def validate_openapi_005_invariants(
    baseline: Dict[str, Any], current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    inventory = invariants.get("inventory") or {}
    operation_count = _operation_count(current)
    tag_names = [
        str(tag.get("name"))
        for tag in current.get("tags") or []
        if isinstance(tag, dict) and tag.get("name")
    ]
    if operation_count != inventory.get("operations"):
        errors.append(
            f"OPENAPI-005 operations must equal {inventory.get('operations')}, got {operation_count}"
        )
    if len(tag_names) != inventory.get("tags") or len(set(tag_names)) != inventory.get(
        "tags"
    ):
        errors.append(
            f"OPENAPI-005 tags must equal {inventory.get('tags')} unique entries, got {len(tag_names)}"
        )

    for invariant_name in ("listResumes", "getResume"):
        expected = invariants.get(invariant_name) or {}
        baseline_operation = _operation_at_contract_path(
            baseline, str(expected.get("path") or ""), str(expected.get("method") or "")
        )
        current_operation = _operation_at_contract_path(
            current, str(expected.get("path") or ""), str(expected.get("method") or "")
        )
        if baseline_operation is None or current_operation is None:
            errors.append(f"{invariant_name} method/path must remain unchanged")
            continue
        if baseline_operation.get("operationId") != expected.get(
            "operationId"
        ) or current_operation.get("operationId") != expected.get("operationId"):
            errors.append(f"{invariant_name} operationId must remain unchanged")
        status = expected.get("successStatus")
        if _success_response_schema(baseline_operation, status) != expected.get(
            "response"
        ) or _success_response_schema(current_operation, status) != expected.get("response"):
            errors.append(f"{invariant_name} success response envelope must remain unchanged")
        if set(str(code) for code in baseline_operation.get("responses") or {}) != set(
            str(code) for code in current_operation.get("responses") or {}
        ):
            errors.append(f"{invariant_name} response statuses must remain unchanged")
        if baseline_operation.get("parameters") != current_operation.get("parameters"):
            errors.append(f"{invariant_name} parameters must remain unchanged")
    return errors


OPENAPI_007_RETIRED_FIELDS = ("uiLanguage", "preferredPracticeLanguage", "emailMasked")
OPENAPI_007_USER_FIELDS = (
    "id",
    "email",
    "displayName",
    "profileCompletionRequired",
)
OPENAPI_007_EMAIL_SCHEMA = {
    "type": "string",
    "format": "email",
    "description": "Complete account email returned only to the authenticated user for account settings.",
}


def normalize_openapi_007_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    findings: List[Dict[str, Any]] = []
    baseline_user = (
        ((baseline.get("components") or {}).get("schemas") or {}).get("UserContext") or {}
    )
    current_user = (
        ((current.get("components") or {}).get("schemas") or {}).get("UserContext") or {}
    )
    baseline_required = set(baseline_user.get("required") or [])
    current_required = set(current_user.get("required") or [])
    baseline_properties = baseline_user.get("properties") or {}
    current_properties = current_user.get("properties") or {}
    for field in OPENAPI_007_RETIRED_FIELDS:
        if field in baseline_required and field not in current_required:
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer("components", "schemas", "UserContext", "required", field),
                    "required_property_removed",
                    "required",
                    "absent",
                )
            )
        if field in baseline_properties and field not in current_properties:
            findings.append(
                _normalized_finding(
                    "breaking",
                    _json_pointer("components", "schemas", "UserContext", "properties", field),
                    "property_removed",
                    _type_signature(baseline_properties[field]),
                    "absent",
                )
            )
    if "email" not in baseline_required and "email" in current_required:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "UserContext", "required", "email"),
                "required_property_added",
                "absent",
                "required",
            )
        )
    if "email" not in baseline_properties and "email" in current_properties:
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("components", "schemas", "UserContext", "properties", "email"),
                "property_added",
                "absent",
                _type_signature(current_properties["email"]),
            )
        )
    baseline_closure = baseline_user.get("additionalProperties", "unspecified")
    current_closure = current_user.get("additionalProperties", "unspecified")
    if baseline_closure != current_closure:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "UserContext", "additionalProperties"),
                "closed_object",
                baseline_closure,
                current_closure,
            )
        )
    return findings


def validate_openapi_007_contract(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    baseline_user = copy.deepcopy(
        ((baseline.get("components") or {}).get("schemas") or {}).get("UserContext") or {}
    )
    current_user = copy.deepcopy(
        ((current.get("components") or {}).get("schemas") or {}).get("UserContext") or {}
    )
    baseline_required = set(baseline_user.get("required") or [])
    baseline_properties = set((baseline_user.get("properties") or {}).keys())
    for field in OPENAPI_007_RETIRED_FIELDS:
        if field not in baseline_required or field not in baseline_properties:
            errors.append(f"old baseline UserContext must retain required {field}")
    if current_user.get("type") != "object" or current_user.get("additionalProperties") is not False:
        errors.append("UserContext must be an explicitly closed object")
    if list(current_user.get("required") or []) != list(OPENAPI_007_USER_FIELDS):
        errors.append("UserContext required fields must equal the exact four-field projection")
    if list((current_user.get("properties") or {}).keys()) != list(OPENAPI_007_USER_FIELDS):
        errors.append("UserContext properties must equal the exact four-field projection")
    expected = copy.deepcopy(baseline_user)
    expected["additionalProperties"] = False
    expected["required"] = list(OPENAPI_007_USER_FIELDS)
    for field in OPENAPI_007_RETIRED_FIELDS:
        (expected.get("properties") or {}).pop(field, None)
    (expected.get("properties") or {})["email"] = copy.deepcopy(OPENAPI_007_EMAIL_SCHEMA)
    expected["properties"] = {
        field: expected["properties"][field] for field in OPENAPI_007_USER_FIELDS
    }
    if current_user != expected:
        errors.append(
            "UserContext may change only by pruning language/emailMasked fields, adding complete email, and closing the object"
        )
    return errors


def validate_openapi_007_invariants(
    baseline: Dict[str, Any], current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    inventory = invariants.get("inventory") or {}
    tags = [tag.get("name") for tag in current.get("tags") or [] if isinstance(tag, dict)]
    if _operation_count(current) != inventory.get("operations"):
        errors.append("OPENAPI-007 operation inventory drifted")
    if len(tags) != inventory.get("tags") or len(set(tags)) != inventory.get("tags"):
        errors.append("OPENAPI-007 tag inventory drifted")
    for name in ("getMe", "completeMyProfile", "deleteMe"):
        expected = invariants.get(name) or {}
        baseline_operation = _operation_at_contract_path(
            baseline, str(expected.get("path") or ""), str(expected.get("method") or "")
        )
        current_operation = _operation_at_contract_path(
            current, str(expected.get("path") or ""), str(expected.get("method") or "")
        )
        if baseline_operation is None or current_operation is None:
            errors.append(f"{name} method/path must remain unchanged")
            continue
        if baseline_operation.get("operationId") != name or current_operation.get("operationId") != name:
            errors.append(f"{name} operationId must remain unchanged")
        status = str(expected.get("successStatus"))
        if status not in (baseline_operation.get("responses") or {}) or status not in (current_operation.get("responses") or {}):
            errors.append(f"{name} success status must remain {status}")
        baseline_security = baseline_operation.get("security", baseline.get("security"))
        current_security = current_operation.get("security", current.get("security"))
        if baseline_security != current_security or current_security != [{"sessionCookie": []}]:
            errors.append(f"{name} session-cookie security must remain unchanged")
    return errors


def build_openapi_007_oracle(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> Dict[str, Any]:
    return OrderedDict(
        schemaVersion=1,
        decisionId="OPENAPI-007",
        baseline="openapi/baseline/openapi-v1.0.0.yaml@merge-base(main)",
        proposed="openapi/openapi.yaml@working-tree",
        comparison={
            "mode": "exact-set",
            "keyFields": list(OPENAPI_001_FINDING_KEYS),
            "orderSignificant": False,
            "missingFinding": "fail",
            "unexpectedFinding": "fail",
        },
        authority={
            "decision": "OPENAPI-007",
            "specDecision": "D-39",
            "historyVersion": "1.64",
            "productDecision": "Scheme A",
        },
        invariants={
            "inventory": {"operations": 37, "tags": 10},
            "getMe": {"path": "/api/v1/me", "method": "GET", "successStatus": 200},
            "completeMyProfile": {"path": "/api/v1/me", "method": "PATCH", "successStatus": 200},
            "deleteMe": {"path": "/api/v1/me", "method": "DELETE", "successStatus": 202},
        },
        findings=sorted(normalize_openapi_007_findings(baseline, current), key=_finding_key),
    )


OPENAPI_006_LIMIT_FIELDS = (
    "resumeUploadBytes",
    "resumePasteTextBytes",
    "targetJobRawTextBytes",
    "practiceMessageBytes",
    "practiceSessionTextBytes",
)


def normalize_openapi_006_findings(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Normalize only OPENAPI-006's RuntimeConfig content-limit correction."""
    findings: List[Dict[str, Any]] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    current_schemas = ((current.get("components") or {}).get("schemas") or {})
    baseline_runtime = baseline_schemas.get("RuntimeConfig") or {}
    current_runtime = current_schemas.get("RuntimeConfig") or {}

    baseline_required = list(baseline_runtime.get("required") or [])
    current_required = list(current_runtime.get("required") or [])
    if "contentLimits" not in baseline_required and "contentLimits" in current_required:
        findings.append(
            _normalized_finding(
                "breaking",
                _json_pointer("components", "schemas", "RuntimeConfig", "required"),
                "required_property_added",
                "absent",
                "contentLimits",
            )
        )

    baseline_properties = baseline_runtime.get("properties") or {}
    current_properties = current_runtime.get("properties") or {}
    if "contentLimits" not in baseline_properties and "contentLimits" in current_properties:
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer(
                    "components", "schemas", "RuntimeConfig", "properties", "contentLimits"
                ),
                "property_added",
                "absent",
                "ContentLimits",
            )
        )

    if "ContentLimits" not in baseline_schemas and "ContentLimits" in current_schemas:
        limits = current_schemas["ContentLimits"]
        required = list(limits.get("required") or []) if isinstance(limits, dict) else []
        findings.append(
            _normalized_finding(
                "additive",
                _json_pointer("components", "schemas", "ContentLimits"),
                "schema_added_with_required_fields" if required else "schema_added",
                "absent",
                ",".join(required) if required else "present",
            )
        )
        if isinstance(limits, dict) and limits.get("additionalProperties") is False:
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer(
                        "components", "schemas", "ContentLimits", "additionalProperties"
                    ),
                    "closed_object",
                    "absent",
                    False,
                )
            )
        for name, schema in (limits.get("properties") or {}).items():
            signature = _type_signature(schema)
            if isinstance(schema, dict) and schema.get("format"):
                signature += f"(format={schema['format']},minimum={schema.get('minimum')})"
            findings.append(
                _normalized_finding(
                    "additive",
                    _json_pointer(
                        "components", "schemas", "ContentLimits", "properties", name
                    ),
                    "property_added",
                    "absent",
                    signature,
                )
            )
    return findings


def validate_openapi_006_contract(
    baseline: Dict[str, Any], current: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    baseline_schemas = ((baseline.get("components") or {}).get("schemas") or {})
    schemas = ((current.get("components") or {}).get("schemas") or {})
    baseline_runtime = copy.deepcopy(baseline_schemas.get("RuntimeConfig") or {})
    runtime = copy.deepcopy(schemas.get("RuntimeConfig") or {})

    if "ContentLimits" in baseline_schemas:
        errors.append("old baseline must not contain ContentLimits")
    if "contentLimits" in (baseline_runtime.get("properties") or {}):
        errors.append("old baseline RuntimeConfig must not contain contentLimits")

    expected_limit_property = {"type": "integer", "format": "int64", "minimum": 1}
    expected_limits = {
        "type": "object",
        "additionalProperties": False,
        "required": list(OPENAPI_006_LIMIT_FIELDS),
        "properties": {
            name: copy.deepcopy(expected_limit_property)
            for name in OPENAPI_006_LIMIT_FIELDS
        },
    }
    actual_limits = copy.deepcopy(schemas.get("ContentLimits") or {})
    actual_limits.pop("description", None)
    if actual_limits != expected_limits:
        errors.append("ContentLimits must be closed, required, and contain exactly five positive int64 fields")

    if (runtime.get("properties") or {}).get("contentLimits") != {
        "$ref": "#/components/schemas/ContentLimits"
    }:
        errors.append("RuntimeConfig.contentLimits must reference ContentLimits")
    required = list(runtime.get("required") or [])
    if required != [*list(baseline_runtime.get("required") or []), "contentLimits"]:
        errors.append("RuntimeConfig must add only required contentLimits")

    runtime["required"] = list(baseline_runtime.get("required") or [])
    (runtime.get("properties") or {}).pop("contentLimits", None)
    if runtime != baseline_runtime:
        errors.append("RuntimeConfig may change only by adding required contentLimits")
    return errors


def validate_openapi_006_invariants(
    baseline: Dict[str, Any], current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    inventory = invariants.get("inventory") or {}
    tag_names = [
        str(tag.get("name"))
        for tag in current.get("tags") or []
        if isinstance(tag, dict) and tag.get("name")
    ]
    if _operation_count(current) != inventory.get("operations"):
        errors.append("OPENAPI-006 operation inventory drifted")
    if len(tag_names) != inventory.get("tags") or len(set(tag_names)) != inventory.get("tags"):
        errors.append("OPENAPI-006 tag inventory drifted")

    expected = invariants.get("getRuntimeConfig") or {}
    baseline_operation = _operation_at_contract_path(
        baseline, str(expected.get("path") or ""), str(expected.get("method") or "")
    )
    current_operation = _operation_at_contract_path(
        current, str(expected.get("path") or ""), str(expected.get("method") or "")
    )
    if baseline_operation is None or current_operation is None:
        errors.append("getRuntimeConfig method/path must remain unchanged")
        return errors
    if baseline_operation.get("operationId") != expected.get("operationId") or current_operation.get("operationId") != expected.get("operationId"):
        errors.append("getRuntimeConfig operationId must remain unchanged")
    status = expected.get("successStatus")
    if _success_response_schema(baseline_operation, status) != expected.get("response") or _success_response_schema(current_operation, status) != expected.get("response"):
        errors.append("getRuntimeConfig success response must remain RuntimeConfig")
    return errors


def validate_d_35_contract(
    current: Dict[str, Any], baseline: Optional[Dict[str, Any]] = None
) -> List[str]:
    schemas = ((current.get("components") or {}).get("schemas") or {})
    baseline_schemas = (
        ((baseline.get("components") or {}).get("schemas") or {}) if baseline else {}
    )
    errors: List[str] = []
    message = schemas.get("PracticeMessage") or {}
    expected_branches = [
        "PracticeUserMessage",
        "PracticeAssistantMessage",
    ]
    if [_type_signature(branch) for branch in message.get("oneOf") or []] != expected_branches:
        errors.append("PracticeMessage oneOf must equal user then assistant projections")
    discriminator = message.get("discriminator") or {}
    if discriminator.get("propertyName") != "role" or discriminator.get("mapping") != {
        "user": "#/components/schemas/PracticeUserMessage",
        "assistant": "#/components/schemas/PracticeAssistantMessage",
    }:
        errors.append("PracticeMessage must use the exact role discriminator mapping")
    if set(message) != {"oneOf", "discriminator"}:
        errors.append("PracticeMessage must not retain legacy or extra schema keywords")

    reply_status = schemas.get("PracticeReplyStatus") or {}
    if reply_status.get("type") != "string" or reply_status.get("enum") != [
        "pending",
        "retryable_failed",
        "terminal_failed",
        "complete",
    ]:
        errors.append("PracticeReplyStatus must equal the four durable reply states")
    if set(reply_status) != {"type", "enum"}:
        errors.append("PracticeReplyStatus must not declare extra schema keywords")

    base_required = ["id", "seqNo", "role", "content", "createdAt"]
    expected_base_properties = {
        "id": {"type": "string", "format": "uuid"},
        "seqNo": {"type": "integer", "format": "int32", "minimum": 1},
        "role": {"type": "string", "enum": ["user", "assistant"]},
        "content": {"type": "string"},
        "createdAt": {"type": "string", "format": "date-time"},
    }
    if baseline_schemas:
        baseline_message = baseline_schemas.get("PracticeMessage") or {}
        if baseline_message.get("required") != base_required:
            errors.append("baseline PracticeMessage required fields drifted")
        if baseline_message.get("properties") != expected_base_properties:
            errors.append("baseline PracticeMessage field contract drifted")
    for name, role, required in (
        (
            "PracticeUserMessage",
            "user",
            base_required + ["clientMessageId", "replyStatus"],
        ),
        ("PracticeAssistantMessage", "assistant", base_required),
    ):
        schema = schemas.get(name) or {}
        properties = schema.get("properties") or {}
        if schema.get("type") != "object" or schema.get("additionalProperties") is not False:
            errors.append(f"{name} must be a closed object")
        if set(schema) != {"type", "additionalProperties", "required", "properties"}:
            errors.append(f"{name} must not declare extra schema keywords")
        if schema.get("required") != required:
            errors.append(f"{name} required fields must equal {required}")
        if (properties.get("role") or {}).get("enum") != [role]:
            errors.append(f"{name} must fix role={role}")
        if name == "PracticeUserMessage":
            if _type_signature(properties.get("clientMessageId") or {}) != "string":
                errors.append("PracticeUserMessage clientMessageId must be a string")
            if _type_signature(properties.get("replyStatus") or {}) != "PracticeReplyStatus":
                errors.append("PracticeUserMessage replyStatus must reference PracticeReplyStatus")
        elif "clientMessageId" in properties or "replyStatus" in properties:
            errors.append("PracticeAssistantMessage must forbid recovery fields")
        expected_properties = copy.deepcopy(expected_base_properties)
        expected_properties["role"] = {"type": "string", "enum": [role]}
        if name == "PracticeUserMessage":
            expected_properties.update(
                {
                    "clientMessageId": {"type": "string", "format": "uuid"},
                    "replyStatus": {
                        "$ref": "#/components/schemas/PracticeReplyStatus"
                    },
                }
            )
        if properties != expected_properties:
            errors.append(f"{name} properties must exactly preserve the role projection")

    response = schemas.get("SendPracticeMessageResponse") or {}
    response_properties = response.get("properties") or {}
    if _type_signature(response_properties.get("userMessage") or {}) != "PracticeUserMessage":
        errors.append("send response userMessage must reference PracticeUserMessage")
    if (
        _type_signature(response_properties.get("assistantMessage") or {})
        != "PracticeAssistantMessage"
    ):
        errors.append("send response assistantMessage must reference PracticeAssistantMessage")
    if baseline_schemas:
        baseline_response = baseline_schemas.get("SendPracticeMessageResponse") or {}
        normalized_response = copy.deepcopy(response)
        normalized_properties = normalized_response.get("properties") or {}
        normalized_properties["userMessage"] = {
            "$ref": "#/components/schemas/PracticeMessage"
        }
        normalized_properties["assistantMessage"] = {
            "$ref": "#/components/schemas/PracticeMessage"
        }
        if normalized_response != baseline_response:
            errors.append(
                "SendPracticeMessageResponse may change only user/assistant projection refs"
            )
        if schemas.get("PracticeSession") != baseline_schemas.get("PracticeSession"):
            errors.append("PracticeSession must remain byte-equivalent to the old schema")
        new_practice_schemas = {
            name
            for name in set(schemas) - set(baseline_schemas)
            if name.startswith("Practice")
        }
        if new_practice_schemas != {
            "PracticeReplyStatus",
            "PracticeUserMessage",
            "PracticeAssistantMessage",
        }:
            errors.append(
                f"unexpected new Practice schemas: {sorted(new_practice_schemas)}"
            )
    session_messages = (
        ((schemas.get("PracticeSession") or {}).get("properties") or {}).get("messages")
        or {}
    )
    if _type_signature(session_messages.get("items") or {}) != "PracticeMessage":
        errors.append("PracticeSession messages must reference PracticeMessage")
    return errors


def _operation_at_contract_path(
    doc: Dict[str, Any], contract_path: str, method: str
) -> Optional[Dict[str, Any]]:
    paths = doc.get("paths") or {}
    candidates = [contract_path]
    for server in doc.get("servers") or []:
        if not isinstance(server, dict):
            continue
        prefix = str(server.get("url") or "").rstrip("/")
        if prefix and contract_path.startswith(prefix + "/"):
            candidates.append(contract_path[len(prefix) :])
    for candidate in candidates:
        path_item = paths.get(candidate)
        if isinstance(path_item, dict):
            operation = path_item.get(method.lower())
            if isinstance(operation, dict):
                return operation
    return None


def _success_response_schema(operation: Dict[str, Any], status: Any) -> str:
    responses = operation.get("responses") or {}
    response = responses.get(str(status))
    if response is None:
        response = responses.get(status)
    if not isinstance(response, dict):
        return ""
    schema = (((response.get("content") or {}).get("application/json") or {}).get("schema"))
    return _type_signature(schema or {})


def validate_openapi_002_invariants(
    current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    inventory = invariants.get("inventory") or {}
    expected_operations = inventory.get("operations")
    expected_tags = inventory.get("tags")
    operation_count = _operation_count(current)
    tag_names = [
        str(tag.get("name"))
        for tag in current.get("tags") or []
        if isinstance(tag, dict) and tag.get("name")
    ]
    if operation_count != expected_operations:
        errors.append(
            f"OPENAPI-002 operations must equal {expected_operations}, got {operation_count}"
        )
    if len(tag_names) != expected_tags or len(set(tag_names)) != expected_tags:
        errors.append(
            f"OPENAPI-002 tags must equal {expected_tags} unique entries, got {len(tag_names)}"
        )

    for invariant_name in ("importTargetJob", "createUploadPresign"):
        expected = invariants.get(invariant_name) or {}
        operation = _operation_at_contract_path(
            current, str(expected.get("path") or ""), str(expected.get("method") or "")
        )
        if operation is None:
            errors.append(
                f"{invariant_name} must remain {expected.get('method')} {expected.get('path')}"
            )
            continue
        if operation.get("operationId") != expected.get("operationId"):
            errors.append(
                f"{invariant_name} operationId must equal {expected.get('operationId')}"
            )
        response_name = _success_response_schema(operation, expected.get("successStatus"))
        if response_name != expected.get("response"):
            errors.append(
                f"{invariant_name} response {expected.get('successStatus')} must equal {expected.get('response')}"
            )

    upload_invariant = invariants.get("createUploadPresign") or {}
    expected_purposes = list(upload_invariant.get("remainingPurposes") or [])
    current_purposes = list(
        (
            (
                (
                    (((current.get("components") or {}).get("schemas") or {}).get("UploadPresignRequest") or {}).get("properties")
                    or {}
                ).get("purpose")
                or {}
            ).get("enum")
            or []
        )
    )
    if current_purposes != expected_purposes:
        errors.append(
            f"createUploadPresign remaining purposes must equal {expected_purposes}, got {current_purposes}"
        )

    if invariants.get("compatibilityAliases") == "forbidden":
        current_schemas = ((current.get("components") or {}).get("schemas") or {})
        aliases = [name for name in OPENAPI_002_SOURCE_SCHEMAS if name in current_schemas]
        if aliases:
            errors.append(f"OPENAPI-002 compatibility aliases are forbidden: {aliases}")
    return errors


def validate_d_35_invariants(
    baseline: Dict[str, Any], current: Dict[str, Any], invariants: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    inventory = invariants.get("inventory") or {}
    expected_operations = inventory.get("operations")
    expected_tags = inventory.get("tags")
    operation_count = _operation_count(current)
    tag_names = [
        str(tag.get("name"))
        for tag in current.get("tags") or []
        if isinstance(tag, dict) and tag.get("name")
    ]
    if operation_count != expected_operations:
        errors.append(
            f"D-35 operations must equal {expected_operations}, got {operation_count}"
        )
    if len(tag_names) != expected_tags or len(set(tag_names)) != expected_tags:
        errors.append(
            f"D-35 tags must equal {expected_tags} unique entries, got {len(tag_names)}"
        )
    baseline_paths = baseline.get("paths") or {}
    current_paths = current.get("paths") or {}
    baseline_practice_paths = {
        path for path in baseline_paths if path.startswith("/practice/sessions")
    }
    current_practice_paths = {
        path for path in current_paths if path.startswith("/practice/sessions")
    }
    if baseline_practice_paths != current_practice_paths:
        errors.append("D-35 must not add or remove Practice session paths")

    for invariant_name in ("getPracticeSession", "sendPracticeMessage"):
        expected = invariants.get(invariant_name) or {}
        method = str(expected.get("method") or "")
        path = str(expected.get("path") or "")
        current_operation = _operation_at_contract_path(current, path, method)
        baseline_operation = _operation_at_contract_path(baseline, path, method)
        if current_operation is None:
            errors.append(
                f"{invariant_name} must remain {expected.get('method')} {expected.get('path')}"
            )
            continue
        if current_operation.get("operationId") != expected.get("operationId"):
            errors.append(
                f"{invariant_name} operationId must equal {expected.get('operationId')}"
            )
        response_name = _success_response_schema(
            current_operation, expected.get("successStatus")
        )
        if response_name != expected.get("response"):
            errors.append(
                f"{invariant_name} response {expected.get('successStatus')} must equal {expected.get('response')}"
            )
        if baseline_operation != current_operation:
            errors.append(f"D-35 must not mutate {invariant_name} operation metadata")
    return errors


def _finding_key(finding: Dict[str, Any]) -> Tuple[str, str, str, str, str]:
    return tuple(str(finding.get(key, "")) for key in OPENAPI_001_FINDING_KEYS)  # type: ignore[return-value]


def compare_finding_sets(
    expected: Iterable[Dict[str, Any]], actual: Iterable[Dict[str, Any]]
) -> List[str]:
    expected_counts = Counter(_finding_key(finding) for finding in expected)
    actual_counts = Counter(_finding_key(finding) for finding in actual)
    errors: List[str] = []
    for key, count in sorted((expected_counts - actual_counts).items()):
        errors.append(f"missing finding x{count}: {json.dumps(dict(zip(OPENAPI_001_FINDING_KEYS, key)), sort_keys=True)}")
    for key, count in sorted((actual_counts - expected_counts).items()):
        errors.append(f"unexpected finding x{count}: {json.dumps(dict(zip(OPENAPI_001_FINDING_KEYS, key)), sort_keys=True)}")
    return errors


def validate_decision_record(text: str, decision_id: str) -> List[str]:
    errors: List[str] = []
    if not re.search(rf"^> \*\*ID\*\*: {re.escape(decision_id)}\s*$", text, re.MULTILINE):
        errors.append(f"decision record must declare ID {decision_id}")
    if not re.search(r"^> \*\*状态\*\*: accepted\s*$", text, re.MULTILINE):
        errors.append("decision record status must be accepted")
    return errors


def validate_exact_set_oracle(
    oracle: Dict[str, Any], decision_id: str
) -> List[str]:
    errors: List[str] = []
    if oracle.get("decisionId") != decision_id:
        errors.append("oracle decisionId does not match requested decision")
    comparison = oracle.get("comparison") or {}
    if comparison.get("mode") != "exact-set" or comparison.get("keyFields") != list(
        OPENAPI_001_FINDING_KEYS
    ):
        errors.append("oracle must use exact-set with severity/path/kind/before/after")
    if comparison.get("orderSignificant") is not False:
        errors.append("oracle exact-set comparison must be order-insensitive")
    if comparison.get("missingFinding") != "fail" or comparison.get(
        "unexpectedFinding"
    ) != "fail":
        errors.append("oracle must fail on missing and unexpected findings")

    findings = oracle.get("findings")
    if not isinstance(findings, list):
        errors.append("oracle findings must be a list")
        return errors
    required_keys = set(OPENAPI_001_FINDING_KEYS)
    for index, finding in enumerate(findings):
        if not isinstance(finding, dict) or set(finding) != required_keys:
            errors.append(
                f"oracle finding {index} must contain exactly severity/path/kind/before/after"
            )
            continue
        if any("*" in str(finding[key]) for key in OPENAPI_001_FINDING_KEYS):
            errors.append(f"oracle finding {index} must not use wildcard authorization")
    return errors


def validate_openapi_001_conditional_contract(current: Dict[str, Any]) -> List[str]:
    request = (
        ((current.get("components") or {}).get("schemas") or {}).get(
            "CreatePracticePlanRequest"
        )
        or {}
    )
    errors: List[str] = []
    if request.get("type") != "object":
        errors.append("CreatePracticePlanRequest must remain type=object")
    if request.get("required") != ["goal"]:
        errors.append("CreatePracticePlanRequest top-level required must equal [goal]")
    if request.get("additionalProperties") is not False:
        errors.append("CreatePracticePlanRequest must set additionalProperties=false")
    properties = request.get("properties") or {}
    source = properties.get("sourceReportId") or {}
    if source.get("type") != "string" or source.get("format") != "uuid" or source.get("nullable"):
        errors.append("sourceReportId must be a non-null UUID string property")

    branches = request.get("oneOf") or []
    if not isinstance(branches, list) or len(branches) != 2:
        errors.append("CreatePracticePlanRequest oneOf must have baseline and derived branches")
        return errors

    baseline_branch, derived_branch = branches
    baseline_required = {
        "targetJobId",
        "goal",
        "interviewerPersona",
        "difficulty",
        "language",
        "timeBudgetMinutes",
        "resumeId",
    }
    if set(baseline_branch.get("required") or []) != baseline_required:
        errors.append("baseline branch must require the existing non-focus baseline fields")
    baseline_properties = baseline_branch.get("properties") or {}
    if (baseline_properties.get("goal") or {}).get("const") != "baseline":
        errors.append("baseline branch must fix goal=baseline")
    if "sourceReportId" in baseline_properties:
        errors.append("baseline branch must not declare sourceReportId")
    if baseline_branch.get("additionalProperties") is not False:
        errors.append("baseline branch must reject additional properties")
    if ((baseline_branch.get("not") or {}).get("required") or []) != ["sourceReportId"]:
        errors.append("baseline branch must explicitly forbid sourceReportId")

    if set(derived_branch.get("required") or []) != {"goal", "sourceReportId"}:
        errors.append("derived branch must require goal and sourceReportId")
    derived_properties = derived_branch.get("properties") or {}
    if set(derived_properties) != {"goal", "sourceReportId"}:
        errors.append("derived branch must allow only goal and sourceReportId")
    if (derived_properties.get("goal") or {}).get("enum") != [
        "retry_current_round",
        "next_round",
    ]:
        errors.append("derived branch goal must equal retry_current_round|next_round")
    derived_source = derived_properties.get("sourceReportId") or {}
    if (
        derived_source.get("type") != "string"
        or derived_source.get("format") != "uuid"
        or derived_source.get("nullable")
    ):
        errors.append("derived sourceReportId must be non-null UUID string")
    if derived_branch.get("additionalProperties") is not False:
        errors.append("derived branch must reject every extra field")
    return errors


def _sha256_text(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


def _write_json_payload(payload: Dict[str, Any], output_path: Optional[str]) -> None:
    rendered = json.dumps(payload, ensure_ascii=False, indent=2) + "\n"
    if output_path:
        Path(output_path).write_text(rendered, encoding="utf-8")
    sys.stdout.write(rendered)


def run_openapi_001_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    decision_path = Path(args.decision_record).resolve() if args.decision_record else None
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    if oracle_path is not None and oracle_path.is_file():
        try:
            requested_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            requested_oracle = None
        if isinstance(requested_oracle, dict) and requested_oracle.get("authority") == OPENAPI_001_V17_AUTHORITY:
            return run_openapi_001_v17_audit(
                args,
                repo_root,
                baseline_path,
                current_path,
            )
    errors: List[str] = []
    if decision_path is None or not decision_path.is_file():
        errors.append("OPENAPI-001 audit requires --decision-record")
    if oracle_path is None or not oracle_path.is_file():
        errors.append("OPENAPI-001 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}")

    current_text = current_path.read_text(encoding="utf-8")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    expected_findings: List[Dict[str, Any]] = []
    if decision_path is not None and decision_path.is_file():
        errors.extend(
            validate_decision_record(
                decision_path.read_text(encoding="utf-8"), args.decision_id
            )
        )
    if oracle_path is not None and oracle_path.is_file():
        oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        if oracle.get("decisionId") != args.decision_id:
            errors.append("oracle decisionId does not match requested decision")
        comparison = oracle.get("comparison") or {}
        if comparison.get("mode") != "exact-set" or comparison.get("keyFields") != list(
            OPENAPI_001_FINDING_KEYS
        ):
            errors.append("oracle must use exact-set with severity/path/kind/before/after")
        expected_findings = oracle.get("findings") or []

    actual_findings = normalize_openapi_001_findings(baseline_doc or {}, current_doc)
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_openapi_001_conditional_contract(current_doc))
    if (baseline_doc or {}).get("paths") != current_doc.get("paths"):
        errors.append("OPENAPI-001 must not change paths or operations")
    if (baseline_doc or {}).get("tags") != current_doc.get("tags"):
        errors.append("OPENAPI-001 must not change tags")

    oversize_findings = [
        finding
        for finding in actual_findings
        if finding.get("after") == "REPORT_CONTEXT_TOO_LARGE"
    ]
    if oversize_findings != [
        _normalized_finding(
            "additive",
            "/components/schemas/ApiErrorCode/enum",
            "enum_value_added",
            "absent",
            "REPORT_CONTEXT_TOO_LARGE",
        )
    ]:
        errors.append(
            "REPORT_CONTEXT_TOO_LARGE must appear exactly once as additive enum_value_added"
        )

    sorted_findings = sorted(actual_findings, key=_finding_key)
    summary = _format_summary(sorted_findings)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=summary,
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


def emit_openapi_001_v17_oracle(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    """Print, but never write, the oracle that must be checked in by the caller."""
    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    if base_commit is None:
        sys.stderr.write(f"ERROR: cannot resolve base ref {base_ref!r}\n")
        return 1
    baseline_text = _git_show(repo_root, base_commit, baseline_path)
    if baseline_text is None:
        sys.stderr.write(
            f"ERROR: cannot load {baseline_path.relative_to(repo_root)} from {base_commit}\n"
        )
        return 1
    if baseline_path.read_text(encoding="utf-8") != baseline_text:
        sys.stderr.write(
            "ERROR: OPENAPI-001 v1.7 generator requires an unchanged worktree baseline\n"
        )
        return 1
    baseline_doc = yaml.safe_load(baseline_text) or {}
    current_doc = yaml.safe_load(current_path.read_text(encoding="utf-8")) or {}
    sys.stdout.write(
        json.dumps(
            build_openapi_001_v17_oracle(baseline_doc, current_doc),
            ensure_ascii=False,
            indent=2,
        )
        + "\n"
    )
    return 0


def run_openapi_001_v17_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    decision_path = Path(args.decision_record).resolve() if args.decision_record else None
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    errors: List[str] = []
    if decision_path is None or not decision_path.is_file():
        errors.append("OPENAPI-001 v1.7 audit requires --decision-record")
    if oracle_path is None or not oracle_path.is_file():
        errors.append("OPENAPI-001 v1.7 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(
                f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}"
            )

    current_text = current_path.read_text(encoding="utf-8")
    worktree_baseline_text = baseline_path.read_text(encoding="utf-8")
    if baseline_text is not None and worktree_baseline_text != baseline_text:
        errors.append("OPENAPI-001 v1.7 worktree baseline differs from base-ref snapshot")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    oracle: Dict[str, Any] = {}
    expected_findings: List[Dict[str, Any]] = []
    if decision_path is not None and decision_path.is_file():
        errors.extend(
            validate_decision_record(
                decision_path.read_text(encoding="utf-8"), args.decision_id
            )
        )
    if oracle_path is not None and oracle_path.is_file():
        try:
            loaded_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"cannot parse OPENAPI-001 v1.7 oracle: {exc}")
        else:
            if isinstance(loaded_oracle, dict):
                oracle = loaded_oracle
                errors.extend(validate_exact_set_oracle(oracle, args.decision_id))
                raw_findings = oracle.get("findings")
                if isinstance(raw_findings, list):
                    expected_findings = [
                        finding for finding in raw_findings if isinstance(finding, dict)
                    ]
            else:
                errors.append("OPENAPI-001 v1.7 oracle must be a JSON object")

    errors.extend(validate_openapi_001_v17_authority(repo_root, oracle))
    actual_findings = normalize_openapi_001_v17_findings(baseline_doc or {}, current_doc)
    if len(expected_findings) != len(actual_findings):
        errors.append(
            "OPENAPI-001 v1.7 exact-set expected "
            f"{len(expected_findings)} findings but actual {len(actual_findings)}"
        )
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_openapi_001_v17_contract(baseline_doc or {}, current_doc))
    errors.extend(
        validate_openapi_001_v17_invariants(
            current_doc, oracle.get("invariants") or {}
        )
    )

    sorted_findings = sorted(actual_findings, key=_finding_key)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        keyFields=list(OPENAPI_001_FINDING_KEYS),
        authority=oracle.get("authority"),
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=_format_summary(sorted_findings),
        expectedFindingCount=len(expected_findings),
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


def run_openapi_002_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    decision_path = Path(args.decision_record).resolve() if args.decision_record else None
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    errors: List[str] = []
    if decision_path is None or not decision_path.is_file():
        errors.append("OPENAPI-002 audit requires --decision-record")
    if oracle_path is None or not oracle_path.is_file():
        errors.append("OPENAPI-002 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(
                f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}"
            )

    current_text = current_path.read_text(encoding="utf-8")
    worktree_baseline_text = baseline_path.read_text(encoding="utf-8")
    if baseline_text is not None and worktree_baseline_text != baseline_text:
        errors.append("OPENAPI-002 worktree baseline differs from base-ref snapshot")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    oracle: Dict[str, Any] = {}
    expected_findings: List[Dict[str, Any]] = []
    if decision_path is not None and decision_path.is_file():
        errors.extend(
            validate_decision_record(
                decision_path.read_text(encoding="utf-8"), args.decision_id
            )
        )
    if oracle_path is not None and oracle_path.is_file():
        try:
            loaded_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"cannot parse OPENAPI-002 oracle: {exc}")
        else:
            if isinstance(loaded_oracle, dict):
                oracle = loaded_oracle
                errors.extend(validate_exact_set_oracle(oracle, args.decision_id))
                raw_findings = oracle.get("findings")
                if isinstance(raw_findings, list):
                    expected_findings = [
                        finding for finding in raw_findings if isinstance(finding, dict)
                    ]
            else:
                errors.append("OPENAPI-002 oracle must be a JSON object")

    actual_findings = normalize_openapi_002_findings(baseline_doc or {}, current_doc)
    if len(expected_findings) != len(actual_findings):
        errors.append(
            f"OPENAPI-002 exact-set expected {len(expected_findings)} findings but actual {len(actual_findings)}"
        )
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_openapi_002_invariants(current_doc, oracle.get("invariants") or {}))

    sorted_findings = sorted(actual_findings, key=_finding_key)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        keyFields=list(OPENAPI_001_FINDING_KEYS),
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=_format_summary(sorted_findings),
        expectedFindingCount=len(expected_findings),
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


def validate_openapi_004_authority(
    repo_root: Path, oracle: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    if oracle.get("authority") != {
        "decision": "OPENAPI-004",
        "specDecision": "D-36",
        "historyVersion": "1.57",
        "productDecision": "R-A",
    }:
        errors.append(
            "OPENAPI-004 authority must bind accepted decision, spec D-36, history 1.57 and R-A"
        )

    contract_dir = repo_root / "docs" / "spec" / "openapi-v1-contract"
    spec_path = contract_dir / "spec.md"
    history_path = contract_dir / "history.md"
    decision_path = (
        contract_dir / "decisions" / "OPENAPI-004-targetjob-report-overview.md"
    )
    spec_text = spec_path.read_text(encoding="utf-8") if spec_path.is_file() else ""
    history_text = (
        history_path.read_text(encoding="utf-8") if history_path.is_file() else ""
    )
    decision_text = (
        decision_path.read_text(encoding="utf-8") if decision_path.is_file() else ""
    )
    spec_rows = [line for line in spec_text.splitlines() if re.search(r"\|\s*D-36\s*\|", line)]
    if not any("OPENAPI-004" in line and "TargetJob" in line for line in spec_rows):
        errors.append("OPENAPI-004 requires current spec D-36 authority")
    history_rows = [
        line for line in history_text.splitlines() if re.search(r"\|\s*1\.57\s*\|", line)
    ]
    if not any(
        "OPENAPI-004" in line and "TargetJob" in line for line in history_rows
    ):
        errors.append("OPENAPI-004 requires history 1.57 authority")
    if "R-A" not in decision_text:
        errors.append("OPENAPI-004 decision record must preserve product decision R-A")
    return errors


def run_openapi_004_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    decision_path = Path(args.decision_record).resolve() if args.decision_record else None
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    errors: List[str] = []
    if decision_path is None or not decision_path.is_file():
        errors.append("OPENAPI-004 audit requires --decision-record")
    if oracle_path is None or not oracle_path.is_file():
        errors.append("OPENAPI-004 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(
                f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}"
            )

    current_text = current_path.read_text(encoding="utf-8")
    worktree_baseline_text = baseline_path.read_text(encoding="utf-8")
    if baseline_text is not None and worktree_baseline_text != baseline_text:
        errors.append("OPENAPI-004 worktree baseline differs from base-ref snapshot")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    oracle: Dict[str, Any] = {}
    expected_findings: List[Dict[str, Any]] = []
    if decision_path is not None and decision_path.is_file():
        errors.extend(
            validate_decision_record(
                decision_path.read_text(encoding="utf-8"), args.decision_id
            )
        )
    if oracle_path is not None and oracle_path.is_file():
        try:
            loaded_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"cannot parse OPENAPI-004 oracle: {exc}")
        else:
            if isinstance(loaded_oracle, dict):
                oracle = loaded_oracle
                errors.extend(validate_exact_set_oracle(oracle, args.decision_id))
                raw_findings = oracle.get("findings")
                if isinstance(raw_findings, list):
                    expected_findings = [
                        finding for finding in raw_findings if isinstance(finding, dict)
                    ]
            else:
                errors.append("OPENAPI-004 oracle must be a JSON object")

    errors.extend(validate_openapi_004_authority(repo_root, oracle))
    actual_findings = normalize_openapi_004_findings(
        baseline_doc or {}, current_doc
    )
    if len(expected_findings) != len(actual_findings):
        errors.append(
            f"OPENAPI-004 exact-set expected {len(expected_findings)} findings but actual {len(actual_findings)}"
        )
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_openapi_004_contract(baseline_doc or {}, current_doc))
    errors.extend(
        validate_openapi_004_invariants(
            baseline_doc or {}, current_doc, oracle.get("invariants") or {}
        )
    )

    sorted_findings = sorted(actual_findings, key=_finding_key)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        keyFields=list(OPENAPI_001_FINDING_KEYS),
        authority=oracle.get("authority"),
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=_format_summary(sorted_findings),
        expectedFindingCount=len(expected_findings),
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


def validate_openapi_005_authority(
    repo_root: Path, oracle: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    if oracle.get("authority") != {
        "decision": "OPENAPI-005",
        "specDecision": "D-37",
        "historyVersion": "1.59",
        "productDecision": "list-summary",
    }:
        errors.append(
            "OPENAPI-005 authority must bind accepted decision, spec D-37, history 1.59 and list-summary"
        )

    contract_dir = repo_root / "docs" / "spec" / "openapi-v1-contract"
    spec_path = contract_dir / "spec.md"
    history_path = contract_dir / "history.md"
    decision_path = contract_dir / "decisions" / "OPENAPI-005-resume-list-summary.md"
    spec_text = spec_path.read_text(encoding="utf-8") if spec_path.is_file() else ""
    history_text = (
        history_path.read_text(encoding="utf-8") if history_path.is_file() else ""
    )
    decision_text = (
        decision_path.read_text(encoding="utf-8") if decision_path.is_file() else ""
    )
    spec_rows = [
        line for line in spec_text.splitlines() if re.search(r"\|\s*D-37\s*\|", line)
    ]
    if not any("OPENAPI-005" in line and "Resume" in line for line in spec_rows):
        errors.append("OPENAPI-005 requires current spec D-37 authority")
    history_rows = [
        line for line in history_text.splitlines() if re.search(r"\|\s*1\.59\s*\|", line)
    ]
    if not any("OPENAPI-005" in line and "Resume" in line for line in history_rows):
        errors.append("OPENAPI-005 requires history 1.59 authority")
    if "Resume list summary projection" not in decision_text:
        errors.append("OPENAPI-005 decision record must preserve list-summary authority")
    return errors


def run_openapi_005_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    decision_path = Path(args.decision_record).resolve() if args.decision_record else None
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    errors: List[str] = []
    if decision_path is None or not decision_path.is_file():
        errors.append("OPENAPI-005 audit requires --decision-record")
    if oracle_path is None or not oracle_path.is_file():
        errors.append("OPENAPI-005 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(
                f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}"
            )

    current_text = current_path.read_text(encoding="utf-8")
    worktree_baseline_text = baseline_path.read_text(encoding="utf-8")
    if baseline_text is not None and worktree_baseline_text != baseline_text:
        errors.append("OPENAPI-005 worktree baseline differs from base-ref snapshot")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    oracle: Dict[str, Any] = {}
    expected_findings: List[Dict[str, Any]] = []
    if decision_path is not None and decision_path.is_file():
        errors.extend(
            validate_decision_record(
                decision_path.read_text(encoding="utf-8"), args.decision_id
            )
        )
    if oracle_path is not None and oracle_path.is_file():
        try:
            loaded_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"cannot parse OPENAPI-005 oracle: {exc}")
        else:
            if isinstance(loaded_oracle, dict):
                oracle = loaded_oracle
                errors.extend(validate_exact_set_oracle(oracle, args.decision_id))
                raw_findings = oracle.get("findings")
                if isinstance(raw_findings, list):
                    expected_findings = [
                        finding for finding in raw_findings if isinstance(finding, dict)
                    ]
            else:
                errors.append("OPENAPI-005 oracle must be a JSON object")

    errors.extend(validate_openapi_005_authority(repo_root, oracle))
    actual_findings = normalize_openapi_005_findings(baseline_doc or {}, current_doc)
    if len(expected_findings) != len(actual_findings):
        errors.append(
            f"OPENAPI-005 exact-set expected {len(expected_findings)} findings but actual {len(actual_findings)}"
        )
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_openapi_005_contract(baseline_doc or {}, current_doc))
    errors.extend(
        validate_openapi_005_invariants(
            baseline_doc or {}, current_doc, oracle.get("invariants") or {}
        )
    )

    sorted_findings = sorted(actual_findings, key=_finding_key)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        keyFields=list(OPENAPI_001_FINDING_KEYS),
        authority=oracle.get("authority"),
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=_format_summary(sorted_findings),
        expectedFindingCount=len(expected_findings),
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


def validate_openapi_006_authority(
    repo_root: Path, oracle: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    if oracle.get("authority") != {
        "decision": "OPENAPI-006",
        "specDecision": "D-38",
        "historyVersion": "1.60",
        "productDecision": "方案 A and revised defaults",
    }:
        errors.append(
            "OPENAPI-006 authority must bind accepted decision, spec D-38, history 1.60 and user-approved defaults"
        )

    contract_dir = repo_root / "docs" / "spec" / "openapi-v1-contract"
    spec_text = (contract_dir / "spec.md").read_text(encoding="utf-8")
    history_text = (contract_dir / "history.md").read_text(encoding="utf-8")
    decision_text = (
        contract_dir / "decisions" / "OPENAPI-006-runtime-content-limits.md"
    ).read_text(encoding="utf-8")
    if not any(
        "OPENAPI-006" in line and "ContentLimits" in line
        for line in spec_text.splitlines()
        if re.search(r"\|\s*D-38\s*\|", line)
    ):
        errors.append("OPENAPI-006 requires current spec D-38 authority")
    if not any(
        "OPENAPI-006" in line and "contentLimits" in line
        for line in history_text.splitlines()
        if re.search(r"\|\s*1\.60\s*\|", line)
    ):
        errors.append("OPENAPI-006 requires history 1.60 authority")
    if "Runtime content limits" not in decision_text:
        errors.append("OPENAPI-006 decision record must preserve runtime content-limit authority")
    return errors


def run_openapi_006_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    decision_path = Path(args.decision_record).resolve() if args.decision_record else None
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    errors: List[str] = []
    if decision_path is None or not decision_path.is_file():
        errors.append("OPENAPI-006 audit requires --decision-record")
    if oracle_path is None or not oracle_path.is_file():
        errors.append("OPENAPI-006 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(
                f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}"
            )

    current_text = current_path.read_text(encoding="utf-8")
    worktree_baseline_text = baseline_path.read_text(encoding="utf-8")
    if baseline_text is not None and worktree_baseline_text != baseline_text:
        errors.append("OPENAPI-006 worktree baseline differs from base-ref snapshot")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    oracle: Dict[str, Any] = {}
    expected_findings: List[Dict[str, Any]] = []
    if decision_path is not None and decision_path.is_file():
        errors.extend(
            validate_decision_record(
                decision_path.read_text(encoding="utf-8"), args.decision_id
            )
        )
    if oracle_path is not None and oracle_path.is_file():
        try:
            loaded_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"cannot parse OPENAPI-006 oracle: {exc}")
        else:
            if isinstance(loaded_oracle, dict):
                oracle = loaded_oracle
                errors.extend(validate_exact_set_oracle(oracle, args.decision_id))
                raw_findings = oracle.get("findings")
                if isinstance(raw_findings, list):
                    expected_findings = [
                        finding for finding in raw_findings if isinstance(finding, dict)
                    ]
            else:
                errors.append("OPENAPI-006 oracle must be a JSON object")

    errors.extend(validate_openapi_006_authority(repo_root, oracle))
    actual_findings = normalize_openapi_006_findings(baseline_doc or {}, current_doc)
    if len(expected_findings) != len(actual_findings):
        errors.append(
            f"OPENAPI-006 exact-set expected {len(expected_findings)} findings but actual {len(actual_findings)}"
        )
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_openapi_006_contract(baseline_doc or {}, current_doc))
    errors.extend(
        validate_openapi_006_invariants(
            baseline_doc or {}, current_doc, oracle.get("invariants") or {}
        )
    )

    sorted_findings = sorted(actual_findings, key=_finding_key)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        keyFields=list(OPENAPI_001_FINDING_KEYS),
        authority=oracle.get("authority"),
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=_format_summary(sorted_findings),
        expectedFindingCount=len(expected_findings),
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


def validate_d_35_authority(
    repo_root: Path, oracle: Dict[str, Any]
) -> List[str]:
    errors: List[str] = []
    authority = oracle.get("authority") or {}
    if authority != {
        "specDecision": "D-35",
        "historyVersion": "1.54",
        "productDecision": "方案 A",
    }:
        errors.append("D-35 authority must bind spec D-35, history 1.54 and 方案 A")

    contract_dir = repo_root / "docs" / "spec" / "openapi-v1-contract"
    spec_path = contract_dir / "spec.md"
    history_path = contract_dir / "history.md"
    spec_text = spec_path.read_text(encoding="utf-8") if spec_path.is_file() else ""
    history_text = (
        history_path.read_text(encoding="utf-8") if history_path.is_file() else ""
    )
    spec_rows = [line for line in spec_text.splitlines() if re.search(r"\|\s*D-35\s*\|", line)]
    if not any("Practice" in line and "方案 A" in line for line in spec_rows):
        errors.append("D-35 requires current spec D-35 authority for Practice 方案 A")
    history_rows = [
        line
        for line in history_text.splitlines()
        if re.search(r"\|\s*1\.54\s*\|", line)
    ]
    if not any("Practice" in line and "方案 A" in line for line in history_rows):
        errors.append("D-35 requires history 1.54 authority for Practice 方案 A")
    return errors


def run_d_35_audit(
    args: argparse.Namespace,
    repo_root: Path,
    baseline_path: Path,
    current_path: Path,
) -> int:
    oracle_path = Path(args.oracle).resolve() if args.oracle else None
    errors: List[str] = []
    if oracle_path is None or not oracle_path.is_file():
        errors.append("D-35 audit requires --oracle")

    base_ref = args.base_ref or "main"
    base_commit = _git_merge_base(repo_root, "HEAD", base_ref)
    if base_commit is None:
        base_commit = _git_rev_parse(repo_root, base_ref)
    baseline_text: Optional[str] = None
    if base_commit is None:
        errors.append(f"cannot resolve base ref {base_ref!r}")
    else:
        baseline_text = _git_show(repo_root, base_commit, baseline_path)
        if baseline_text is None:
            errors.append(
                f"cannot load {baseline_path.relative_to(repo_root)} from {base_commit}"
            )

    current_text = current_path.read_text(encoding="utf-8")
    worktree_baseline_text = baseline_path.read_text(encoding="utf-8")
    if baseline_text is not None and worktree_baseline_text != baseline_text:
        errors.append("D-35 worktree baseline differs from base-ref snapshot")
    baseline_doc = yaml.safe_load(baseline_text) if baseline_text is not None else {}
    current_doc = yaml.safe_load(current_text) or {}

    oracle: Dict[str, Any] = {}
    expected_findings: List[Dict[str, Any]] = []
    if oracle_path is not None and oracle_path.is_file():
        try:
            loaded_oracle = json.loads(oracle_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"cannot parse D-35 oracle: {exc}")
        else:
            if isinstance(loaded_oracle, dict):
                oracle = loaded_oracle
                errors.extend(validate_exact_set_oracle(oracle, args.decision_id))
                raw_findings = oracle.get("findings")
                if isinstance(raw_findings, list):
                    expected_findings = [
                        finding for finding in raw_findings if isinstance(finding, dict)
                    ]
            else:
                errors.append("D-35 oracle must be a JSON object")

    errors.extend(validate_d_35_authority(repo_root, oracle))
    actual_findings = normalize_d_35_findings(baseline_doc or {}, current_doc)
    if len(expected_findings) != len(actual_findings):
        errors.append(
            f"D-35 exact-set expected {len(expected_findings)} findings but actual {len(actual_findings)}"
        )
    errors.extend(compare_finding_sets(expected_findings, actual_findings))
    errors.extend(validate_d_35_contract(current_doc, baseline_doc or {}))
    errors.extend(
        validate_d_35_invariants(
            baseline_doc or {}, current_doc, oracle.get("invariants") or {}
        )
    )

    sorted_findings = sorted(actual_findings, key=_finding_key)
    payload: Dict[str, Any] = OrderedDict(
        schemaVersion=1,
        decisionId=args.decision_id,
        mode="exact-set",
        keyFields=list(OPENAPI_001_FINDING_KEYS),
        authority=oracle.get("authority"),
        baselineSource=(
            f"git:{base_commit}:{baseline_path.relative_to(repo_root)}"
            if base_commit is not None
            else None
        ),
        currentSource=str(current_path.relative_to(repo_root)),
        baselineSha256=_sha256_text(baseline_text or ""),
        currentSha256=_sha256_text(current_text),
        summary=_format_summary(sorted_findings),
        expectedFindingCount=len(expected_findings),
        findingCount=len(sorted_findings),
        findings=sorted_findings,
        errors=errors,
    )
    _write_json_payload(payload, args.output)
    return 0 if not errors else 1


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

    if args.emit_openapi_001_v17_oracle:
        return emit_openapi_001_v17_oracle(
            args,
            repo_root,
            baseline_path,
            current_path,
        )

    if args.decision_id:
        if args.decision_id not in {
            "OPENAPI-001",
            "OPENAPI-002",
            "OPENAPI-004",
            "OPENAPI-005",
            "OPENAPI-006",
            "D-35",
        }:
            sys.stderr.write(f"ERROR: unsupported decision audit: {args.decision_id}\n")
            return 1
        if args.decision_id == "OPENAPI-004":
            return run_openapi_004_audit(
                args,
                repo_root,
                baseline_path,
                current_path,
            )
        if args.decision_id == "OPENAPI-005":
            return run_openapi_005_audit(
                args,
                repo_root,
                baseline_path,
                current_path,
            )
        if args.decision_id == "OPENAPI-006":
            return run_openapi_006_audit(
                args,
                repo_root,
                baseline_path,
                current_path,
            )
        if args.decision_id == "D-35":
            return run_d_35_audit(
                args,
                repo_root,
                baseline_path,
                current_path,
            )
        if args.decision_id == "OPENAPI-002":
            return run_openapi_002_audit(
                args,
                repo_root,
                baseline_path,
                current_path,
            )
        return run_openapi_001_audit(
            args,
            repo_root,
            baseline_path,
            current_path,
        )

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
    expected_operations = ((config.get("contractInventory") or {}).get("endpointCount"))

    output = OrderedDict(
        tool=f"wrapper-{WRAPPER_VERSION}",
        baseline=str(baseline_path),
        current=str(current_path),
        config=str(config_path) if config_path.is_file() else None,
        history=str(history_path) if history_path.is_file() else None,
        historyRef=history_ref,
        historyRefInput=args.history_ref,
        inventory=OrderedDict(
            expectedOperations=expected_operations,
            baselineOperations=_operation_count(baseline_doc),
            currentOperations=_operation_count(current_doc),
        ),
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
    parser.add_argument("--decision-id", default=None)
    parser.add_argument("--decision-record", default=None)
    parser.add_argument("--oracle", default=None)
    parser.add_argument("--base-ref", default=None)
    parser.add_argument(
        "--emit-openapi-001-v17-oracle",
        action="store_true",
        default=False,
        help="Print the deterministic old-baseline to proposed OPENAPI-001 v1.7 oracle.",
    )
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
