#!/usr/bin/env python3
"""Capture bounded DB/API evidence and bind its machine facts into the manifest."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


SCHEMA_VERSION = "p0-099-live-capture.v2"
MAX_RESPONSE_BYTES = 1024 * 1024
MAX_DATABASE_BYTES = 4 * 1024 * 1024
CONTENT_KEYS = (
    "summary",
    "preparednessLevel",
    "dimensionAssessments",
    "highlights",
    "issues",
    "nextActions",
    "retryFocusDimensionCodes",
    "provenance",
)


class CaptureError(ValueError):
    pass


class ManualRequired(RuntimeError):
    pass


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req: Any, fp: Any, code: int, msg: str, headers: Any, newurl: str) -> None:
        return None


def fail(message: str) -> None:
    raise CaptureError(message)


def write_artifact(path: Path, run_id: str, result: str, reason_code: str, reports: list[dict[str, Any]]) -> None:
    artifact = {
        "schema_version": SCHEMA_VERSION,
        "scenario_id": "E2E.P0.099",
        "run_id": run_id,
        "method": "authenticated-live-http+read-only-postgres",
        "captured_at": datetime.now(timezone.utc).isoformat(timespec="microseconds").replace("+00:00", "Z"),
        "result": result,
        "reason_code": reason_code,
        "reports": reports,
        "privacy": {
            "cookie_written": False,
            "database_url_written": False,
            "raw_api_written": False,
            "raw_db_written": False,
            "raw_frozen_context_written": False,
            "prose_written": False,
        },
    }
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(f".{path.name}.tmp")
    temporary.write_text(
        json.dumps(artifact, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    temporary.chmod(0o600)
    temporary.replace(path)


def load_refs(path: Path, run_id: str) -> list[str]:
    if not path.is_file():
        raise ManualRequired("manifest_missing")
    try:
        manifest = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"manifest_invalid:{type(exc).__name__}")
    if not isinstance(manifest, dict):
        fail("manifest_invalid_shape")
    if manifest.get("scenario_id") != "E2E.P0.099" or manifest.get("run_id") != run_id:
        fail("manifest_run_mismatch")
    rows = manifest.get("screenshots")
    if not isinstance(rows, list) or len(rows) != 6:
        fail("manifest_screenshot_matrix_invalid")
    refs: list[str] = []
    for row in rows:
        if not isinstance(row, dict):
            fail("manifest_screenshot_row_invalid")
        report_ref = row.get("report_ref")
        if not isinstance(report_ref, str) or not report_ref or len(report_ref) > 128 or re.search(r"\s", report_ref):
            fail("manifest_report_ref_invalid")
        try:
            uuid.UUID(report_ref)
        except ValueError:
            fail("manifest_report_ref_not_uuid")
        if report_ref not in refs:
            refs.append(report_ref)
    if len(refs) != 3:
        fail("manifest_report_ref_set_invalid")
    return refs


def canonical_json_digest(value: Any) -> str:
    canonical = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def project_public_evidence(value: Any, path: str) -> list[dict[str, Any]]:
    if not isinstance(value, list):
        fail(f"{path}_not_array")
    projected = []
    expected_keys = {
        "dimensionCode",
        "evidence",
        "confidence",
        "sourceMessageSeqNos",
    }
    for index, item in enumerate(value):
        if not isinstance(item, dict) or set(item) != expected_keys:
            fail(f"{path}_{index}_shape_invalid")
        anchors = item["sourceMessageSeqNos"]
        if (
            not isinstance(anchors, list)
            or not anchors
            or any(
                not isinstance(anchor, int) or isinstance(anchor, bool) or anchor <= 0
                for anchor in anchors
            )
        ):
            fail(f"{path}_{index}_anchors_invalid")
        projected.append(
            {
                "dimensionCode": item["dimensionCode"],
                "evidence": item["evidence"],
                "confidence": item["confidence"],
            }
        )
    return projected


def postgres_environment(database_url: str) -> dict[str, str]:
    if not database_url:
        raise ManualRequired("database_url_missing")
    parsed = urllib.parse.urlsplit(database_url)
    if (
        parsed.scheme not in {"postgres", "postgresql"}
        or parsed.hostname not in {"127.0.0.1", "localhost", "::1"}
        or not parsed.username
        or not parsed.path.lstrip("/")
        or parsed.fragment
    ):
        raise ManualRequired("database_url_invalid")
    query = urllib.parse.parse_qs(parsed.query, keep_blank_values=True)
    if set(query) - {"sslmode"} or any(len(values) != 1 for values in query.values()):
        raise ManualRequired("database_url_invalid")

    env = {
        "PATH": os.environ.get("PATH", ""),
        "PGHOST": parsed.hostname,
        "PGPORT": str(parsed.port or 5432),
        "PGUSER": urllib.parse.unquote(parsed.username),
        "PGDATABASE": urllib.parse.unquote(parsed.path.lstrip("/")),
        "PGCONNECT_TIMEOUT": "5",
        "PGOPTIONS": "-c default_transaction_read_only=on",
    }
    if parsed.password is not None:
        env["PGPASSWORD"] = urllib.parse.unquote(parsed.password)
    if "sslmode" in query:
        env["PGSSLMODE"] = query["sslmode"][0]
    for key in ("HOME", "LANG", "LC_ALL", "TMPDIR"):
        if key in os.environ:
            env[key] = os.environ[key]
    return env


def database_query(refs: list[str]) -> str:
    quoted_refs = ",".join(f"'{ref}'::uuid" for ref in refs)
    return f"""
select json_build_object(
  'report_ref', fr.id::text,
  'session_ref', fr.session_id::text,
  'status', fr.status,
  'preparedness_level', fr.preparedness_level,
  'report_created_at', to_char(fr.created_at at time zone 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"'),
  'session_created_at', to_char(ps.created_at at time zone 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"'),
  'generation_context', fr.generation_context,
  'summary', fr.summary,
  'dimension_assessments', fr.dimension_assessments,
  'highlights', fr.highlights,
  'issues', fr.issues,
  'next_actions', fr.next_actions,
  'retry_focus_dimension_codes', fr.retry_focus_dimension_codes,
  'prompt_version', fr.prompt_version,
  'rubric_version', fr.rubric_version,
  'model_id', fr.model_id,
  'language', fr.language,
  'feature_flag', fr.feature_flag,
  'data_source_version', fr.data_source_version
)::text
from feedback_reports fr
join practice_sessions ps on ps.id = fr.session_id
where fr.id in ({quoted_refs})
order by fr.id;
"""


def project_database_report(raw: Any) -> dict[str, Any]:
    if not isinstance(raw, dict):
        fail("database_row_not_object")
    expected_keys = {
        "report_ref",
        "session_ref",
        "status",
        "preparedness_level",
        "report_created_at",
        "session_created_at",
        "generation_context",
        "summary",
        "dimension_assessments",
        "highlights",
        "issues",
        "next_actions",
        "retry_focus_dimension_codes",
        "prompt_version",
        "rubric_version",
        "model_id",
        "language",
        "feature_flag",
        "data_source_version",
    }
    if set(raw) != expected_keys:
        fail("database_row_shape_invalid")
    report_ref = raw["report_ref"]
    session_ref = raw["session_ref"]
    try:
        uuid.UUID(report_ref)
        uuid.UUID(session_ref)
    except (TypeError, ValueError):
        fail("database_row_identity_invalid")
    status = raw["status"]
    preparedness = raw["preparedness_level"]
    if status not in {"queued", "generating", "ready", "failed"}:
        fail("database_row_status_invalid")
    if status == "ready" and not isinstance(preparedness, str):
        fail("database_row_preparedness_missing")
    if status != "ready" and preparedness is not None:
        fail("database_row_nonready_preparedness_invalid")
    if not isinstance(raw["generation_context"], dict) or not raw["generation_context"]:
        fail("database_row_frozen_context_invalid")
    for key in ("report_created_at", "session_created_at"):
        if not isinstance(raw[key], str) or not raw[key]:
            fail("database_row_timestamp_invalid")

    report_digest = None
    if status == "ready":
        for key in (
            "summary",
            "dimension_assessments",
            "highlights",
            "issues",
            "next_actions",
            "retry_focus_dimension_codes",
            "prompt_version",
            "rubric_version",
            "model_id",
            "language",
            "feature_flag",
            "data_source_version",
        ):
            if raw[key] is None:
                fail("database_row_ready_content_incomplete")
        projection = {
            "summary": raw["summary"],
            "preparednessLevel": preparedness,
            "dimensionAssessments": raw["dimension_assessments"],
            "highlights": project_public_evidence(
                raw["highlights"], "database_row_highlights"
            ),
            "issues": project_public_evidence(
                raw["issues"], "database_row_issues"
            ),
            "nextActions": raw["next_actions"],
            "retryFocusDimensionCodes": raw["retry_focus_dimension_codes"],
            "provenance": {
                "promptVersion": raw["prompt_version"],
                "rubricVersion": raw["rubric_version"],
                "modelId": raw["model_id"],
                "language": raw["language"],
                "featureFlag": raw["feature_flag"],
                "dataSourceVersion": raw["data_source_version"],
            },
        }
        report_digest = canonical_json_digest(projection)

    return {
        "report_ref": report_ref,
        "session_ref": session_ref,
        "status": status,
        "preparedness_level": preparedness,
        "report_created_at": raw["report_created_at"],
        "session_created_at": raw["session_created_at"],
        "frozen_context_digest": canonical_json_digest(raw["generation_context"]),
        "canonical_report_content_digest": report_digest,
    }


def capture_database_reports(refs: list[str], database_url: str) -> dict[str, dict[str, Any]]:
    psql = shutil.which("psql", path=os.environ.get("PATH", ""))
    if psql is None:
        raise ManualRequired("psql_missing")
    try:
        result = subprocess.run(
            [
                psql,
                "--no-psqlrc",
                "--quiet",
                "--tuples-only",
                "--no-align",
                "--set=ON_ERROR_STOP=1",
            ],
            input=database_query(refs),
            env=postgres_environment(database_url),
            text=True,
            capture_output=True,
            timeout=10,
            check=False,
        )
    except (OSError, subprocess.TimeoutExpired):
        raise ManualRequired("database_unavailable") from None
    if result.returncode != 0:
        raise ManualRequired("database_unavailable")
    if len(result.stdout.encode("utf-8")) > MAX_DATABASE_BYTES:
        fail("database_projection_too_large")
    lines = [line for line in result.stdout.splitlines() if line.strip()]
    if len(lines) != 3:
        fail("database_projection_row_count_invalid")
    projected: dict[str, dict[str, Any]] = {}
    for line in lines:
        try:
            row = project_database_report(json.loads(line))
        except json.JSONDecodeError:
            fail("database_projection_json_invalid")
        report_ref = row.pop("report_ref")
        if report_ref in projected:
            fail("database_projection_duplicate_report")
        projected[report_ref] = row
    if set(projected) != set(refs):
        fail("database_projection_report_set_invalid")
    return projected


def contains_source_anchors(value: Any) -> bool:
    if isinstance(value, dict):
        return any(
            key in {"sourceMessageSeqNos", "source_message_seq_nos"} or contains_source_anchors(child)
            for key, child in value.items()
        )
    if isinstance(value, list):
        return any(contains_source_anchors(child) for child in value)
    return False


def require_list(report: dict[str, Any], key: str) -> list[Any]:
    value = report.get(key)
    if not isinstance(value, list):
        fail(f"live_response_{key}_invalid")
    return value


def canonical_report_digest(report: dict[str, Any]) -> str:
    missing = [key for key in CONTENT_KEYS if key not in report]
    if missing:
        fail("live_response_ready_content_incomplete")
    projection = {key: report[key] for key in CONTENT_KEYS}
    canonical = json.dumps(projection, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def action_label_audit(report: dict[str, Any], ready: bool) -> dict[str, Any]:
    if not ready:
        return {"language": "not_applicable", "unit": "not_applicable", "limit": 0, "counts": []}
    context = report.get("context")
    provenance = report.get("provenance")
    language = context.get("language") if isinstance(context, dict) else None
    if not isinstance(language, str) and isinstance(provenance, dict):
        language = provenance.get("language")
    if not isinstance(language, str) or not language:
        fail("live_response_language_missing")
    labels = []
    for action in require_list(report, "nextActions"):
        if not isinstance(action, dict) or not isinstance(action.get("label"), str):
            fail("live_response_action_label_invalid")
        labels.append(action["label"])
    if not 1 <= len(labels) <= 2:
        fail("live_response_action_count_invalid")
    if language.lower().startswith("zh"):
        counts = [len(label) for label in labels]
        audit = {"language": "zh-CN", "unit": "code_points", "limit": 64, "counts": counts}
    else:
        counts = [len(label.split()) for label in labels]
        audit = {"language": "en", "unit": "words", "limit": 24, "counts": counts}
    if any(count < 1 or count > audit["limit"] for count in counts):
        fail("live_response_action_label_limit_invalid")
    return audit


def project_report(report: Any, requested_ref: str) -> dict[str, Any]:
    if not isinstance(report, dict):
        fail("live_response_not_object")
    if contains_source_anchors(report):
        fail("live_response_source_anchors_exposed")
    report_ref = report.get("id")
    session_ref = report.get("sessionId")
    status = report.get("status")
    preparedness = report.get("preparednessLevel")
    if report_ref != requested_ref:
        fail("live_response_report_ref_mismatch")
    if not isinstance(session_ref, str) or not session_ref or len(session_ref) > 128 or re.search(r"\s", session_ref):
        fail("live_response_session_ref_invalid")
    if status not in {"queued", "generating", "ready", "failed"}:
        fail("live_response_status_invalid")
    ready = status == "ready"
    if ready and not isinstance(preparedness, str):
        fail("live_response_preparedness_missing")
    if not ready and preparedness is not None:
        fail("live_response_nonready_preparedness_invalid")

    dimensions = require_list(report, "dimensionAssessments")
    highlights = require_list(report, "highlights")
    issues = require_list(report, "issues")
    actions = require_list(report, "nextActions")
    retry_focus = require_list(report, "retryFocusDimensionCodes")
    if not ready and any((dimensions, highlights, issues, actions, retry_focus)):
        fail("live_response_nonready_content_invalid")

    return {
        "report_ref": report_ref,
        "session_ref": session_ref,
        "status": status,
        "preparedness_level": preparedness,
        "canonical_report_content_digest": canonical_report_digest(report) if ready else None,
        "content_shape": {
            "dimension_assessment_count": len(dimensions),
            "highlight_count": len(highlights),
            "issue_count": len(issues),
            "next_action_count": len(actions),
            "retry_focus_count": len(retry_focus),
        },
        "action_label_audit": action_label_audit(report, ready),
    }


def fetch_report(api_base_url: str, report_ref: str, cookie_value: str) -> dict[str, Any]:
    url = f"{api_base_url.rstrip('/')}/reports/{urllib.parse.quote(report_ref, safe='')}"
    request = urllib.request.Request(
        url,
        headers={"Accept": "application/json", "Cookie": f"ei_session={cookie_value}"},
        method="GET",
    )
    try:
        opener = urllib.request.build_opener(NoRedirect)
        with opener.open(request, timeout=10) as response:
            body = response.read(MAX_RESPONSE_BYTES + 1)
    except urllib.error.HTTPError as exc:
        if exc.code in {401, 403}:
            raise ManualRequired("session_cookie_rejected") from None
        fail(f"live_http_status_{exc.code}")
    except (urllib.error.URLError, TimeoutError, OSError):
        raise ManualRequired("live_http_unavailable") from None
    if len(body) > MAX_RESPONSE_BYTES:
        fail("live_response_too_large")
    try:
        decoded = json.loads(body)
    except (UnicodeDecodeError, json.JSONDecodeError):
        fail("live_response_json_invalid")
    return project_report(decoded, report_ref)


def bind_manifest(path: Path, run_id: str, reports: list[dict[str, Any]]) -> None:
    try:
        manifest = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        fail("manifest_bind_read_failed")
    if not isinstance(manifest, dict) or manifest.get("run_id") != run_id:
        fail("manifest_bind_identity_invalid")
    rows = manifest.get("screenshots")
    if not isinstance(rows, list) or len(rows) != 6:
        fail("manifest_bind_matrix_invalid")
    reports_by_ref = {report["report_ref"]: report for report in reports}
    if len(reports_by_ref) != 3:
        fail("manifest_bind_report_set_invalid")

    audit_keys = {
        "fact_to_judgment_to_action",
        "item_verdict_count",
        "unsupported_count",
        "irrelevant_advice_count",
        "causal_mismatch_count",
        "action_label_audit",
    }
    for index, row in enumerate(rows):
        if not isinstance(row, dict):
            fail(f"manifest_bind_row_{index}_invalid")
        report = reports_by_ref.get(row.get("report_ref"))
        evidence = row.get("evidence")
        if report is None or not isinstance(evidence, dict):
            fail(f"manifest_bind_row_{index}_evidence_invalid")
        audit = evidence.get("content_audit")
        if not isinstance(audit, dict) or set(audit) != audit_keys:
            fail(f"manifest_bind_row_{index}_audit_invalid")
        db = report["db"]
        content_digest = report["canonical_report_content_digest"]
        shape = report["content_shape"]
        row["evidence"] = {
            "collection": {
                "run_id": run_id,
                "method": "trusted-current-run-db-api-capture",
                "report_ref": report["report_ref"],
                "session_ref": report["session_ref"],
                "frozen_context_digest": db["frozen_context_digest"],
                "report_content_digest": content_digest,
                "screenshot_sha256": row.get("screenshot_sha256"),
            },
            "db": {
                "status": db["status"],
                "preparedness_level": db["preparedness_level"],
                "frozen_context_digest": db["frozen_context_digest"],
                "report_content_digest": db["canonical_report_content_digest"],
            },
            "api": {
                "status": report["status"],
                "preparedness_level": report["preparedness_level"],
                "report_content_digest": content_digest,
                "source_message_seq_nos_exposed": False,
            },
            "content_audit": {
                "fact_to_judgment_to_action": audit["fact_to_judgment_to_action"],
                "item_verdict_count": (
                    shape["dimension_assessment_count"]
                    + shape["highlight_count"]
                    + shape["issue_count"]
                ),
                "unsupported_count": audit["unsupported_count"],
                "irrelevant_advice_count": audit["irrelevant_advice_count"],
                "causal_mismatch_count": audit["causal_mismatch_count"],
                "action_label_audit": report["action_label_audit"],
            },
        }

    temporary = path.with_name(f".{path.name}.bind.tmp")
    try:
        temporary.write_text(
            json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        temporary.chmod(0o600)
        temporary.replace(path)
    except OSError:
        temporary.unlink(missing_ok=True)
        fail("manifest_bind_write_failed")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--run-id", required=True)
    parser.add_argument("--api-base-url", required=True)
    parser.add_argument("--bind-manifest", action="store_true")
    args = parser.parse_args()

    parsed_base_url = urllib.parse.urlsplit(args.api_base_url)
    if (
        parsed_base_url.scheme != "http"
        or parsed_base_url.hostname not in {"127.0.0.1", "localhost", "::1"}
        or parsed_base_url.username is not None
        or parsed_base_url.password is not None
        or parsed_base_url.query
        or parsed_base_url.fragment
    ):
        write_artifact(args.output, args.run_id, "FAIL", "api_base_url_invalid", [])
        print("P0.099 live capture failed: api_base_url_invalid", file=sys.stderr)
        return 1

    cookie_value = os.environ.get("P0_099_SESSION_COOKIE", "")
    if not cookie_value:
        write_artifact(args.output, args.run_id, "MANUAL_REQUIRED", "session_cookie_missing", [])
        print("P0_099_LIVE_CAPTURE_MANUAL_REQUIRED reason=session_cookie_missing")
        return 2
    if cookie_value != cookie_value.strip() or re.search(r"[;\r\n]", cookie_value):
        write_artifact(args.output, args.run_id, "FAIL", "session_cookie_input_invalid", [])
        print("P0.099 live capture failed: session_cookie_input_invalid", file=sys.stderr)
        return 1

    try:
        refs = load_refs(args.manifest, args.run_id)
        database_reports = capture_database_reports(
            refs, os.environ.get("P0_099_DATABASE_URL", "")
        )
        reports = []
        for report_ref in refs:
            report = fetch_report(args.api_base_url, report_ref, cookie_value)
            database_report = database_reports[report_ref]
            if (
                database_report["session_ref"] != report["session_ref"]
                or database_report["status"] != report["status"]
                or database_report["preparedness_level"] != report["preparedness_level"]
                or database_report["canonical_report_content_digest"]
                != report["canonical_report_content_digest"]
            ):
                fail("database_api_projection_mismatch")
            report["db"] = {
                key: value
                for key, value in database_report.items()
                if key != "session_ref"
            }
            reports.append(report)
        if args.bind_manifest:
            bind_manifest(args.manifest, args.run_id, reports)
    except ManualRequired as exc:
        write_artifact(args.output, args.run_id, "MANUAL_REQUIRED", str(exc), [])
        print(f"P0_099_LIVE_CAPTURE_MANUAL_REQUIRED reason={exc}")
        return 2
    except CaptureError as exc:
        write_artifact(args.output, args.run_id, "FAIL", str(exc), [])
        print(f"P0.099 live capture failed: {exc}", file=sys.stderr)
        return 1

    write_artifact(args.output, args.run_id, "PASS", "captured", reports)
    print("P0_099_LIVE_CAPTURE_PASS reports=3 privacy=redacted")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
