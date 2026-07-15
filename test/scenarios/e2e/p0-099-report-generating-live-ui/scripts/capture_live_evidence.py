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


SCHEMA_VERSION = "p0-099-live-capture.v3"
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
CONVERSATION_CONTEXT_KEYS = {
    "sourcePlanId",
    "targetJobTitle",
    "targetJobCompany",
    "resumeId",
    "resumeDisplayName",
    "roundId",
    "roundSequence",
    "roundName",
    "roundType",
    "language",
    "hasNextRound",
}
CONVERSATION_MESSAGE_KEYS = {"sequence", "role", "content", "createdAt"}
DATABASE_CONVERSATION_MESSAGE_KEYS = {"sequence", "role", "content", "created_at"}


class CaptureError(ValueError):
    pass


class ManualRequired(RuntimeError):
    pass


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req: Any, fp: Any, code: int, msg: str, headers: Any, newurl: str) -> None:
        return None


def fail(message: str) -> None:
    raise CaptureError(message)


def write_artifact(
    path: Path,
    run_id: str,
    result: str,
    reason_code: str,
    reports: list[dict[str, Any]],
    conversation: dict[str, Any] | None = None,
) -> None:
    artifact = {
        "schema_version": SCHEMA_VERSION,
        "scenario_id": "E2E.P0.099",
        "run_id": run_id,
        "method": "authenticated-live-http+read-only-postgres",
        "captured_at": datetime.now(timezone.utc).isoformat(timespec="microseconds").replace("+00:00", "Z"),
        "result": result,
        "reason_code": reason_code,
        "reports": reports,
        "conversation": conversation,
        "privacy": {
            "cookie_written": False,
            "database_url_written": False,
            "raw_api_written": False,
            "raw_db_written": False,
            "raw_frozen_context_written": False,
            "raw_conversation_content_written": False,
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


def require_uuid(value: Any, path: str) -> str:
    if not isinstance(value, str) or not value or value != value.strip():
        fail(f"{path}_invalid")
    try:
        uuid.UUID(value)
    except ValueError:
        fail(f"{path}_invalid")
    return value


def normalize_utc_timestamp(value: Any, path: str) -> str:
    if not isinstance(value, str) or not value or value != value.strip():
        fail(f"{path}_invalid")
    normalized = value[:-1] + "+00:00" if value.endswith("Z") else value
    try:
        parsed = datetime.fromisoformat(normalized)
    except ValueError:
        fail(f"{path}_invalid")
    if parsed.utcoffset() is None:
        fail(f"{path}_invalid")
    return parsed.astimezone(timezone.utc).isoformat(timespec="microseconds").replace("+00:00", "Z")


def project_conversation_context(value: Any, path: str) -> dict[str, Any]:
    if not isinstance(value, dict) or set(value) != CONVERSATION_CONTEXT_KEYS:
        fail(f"{path}_shape_invalid")
    for key in (
        "sourcePlanId",
        "targetJobTitle",
        "resumeId",
        "resumeDisplayName",
        "roundId",
        "roundName",
        "roundType",
        "language",
    ):
        if not isinstance(value[key], str) or not value[key].strip():
            fail(f"{path}_{key}_invalid")
    if not isinstance(value["targetJobCompany"], str):
        fail(f"{path}_targetJobCompany_invalid")
    if (
        not isinstance(value["roundSequence"], int)
        or isinstance(value["roundSequence"], bool)
        or value["roundSequence"] < 1
    ):
        fail(f"{path}_roundSequence_invalid")
    if not isinstance(value["hasNextRound"], bool):
        fail(f"{path}_hasNextRound_invalid")
    return {key: value[key] for key in sorted(CONVERSATION_CONTEXT_KEYS)}


def project_ordered_messages(
    value: Any,
    path: str,
    message_keys: set[str],
    timestamp_key: str,
) -> dict[str, Any]:
    if not isinstance(value, list):
        fail(f"{path}_not_array")
    projection: list[dict[str, Any]] = []
    previous_sequence = 0
    for index, message in enumerate(value):
        if not isinstance(message, dict) or set(message) != message_keys:
            fail(f"{path}_{index}_shape_invalid")
        sequence = message["sequence"]
        role = message["role"]
        content = message["content"]
        if (
            not isinstance(sequence, int)
            or isinstance(sequence, bool)
            or sequence < 1
            or sequence <= previous_sequence
        ):
            fail(f"{path}_{index}_sequence_invalid")
        if role not in {"user", "assistant"}:
            fail(f"{path}_{index}_role_invalid")
        if not isinstance(content, str) or not content.strip():
            fail(f"{path}_{index}_content_invalid")
        projection.append(
            {
                "sequence": sequence,
                "role": role,
                "content_digest": hashlib.sha256(content.encode("utf-8")).hexdigest(),
                "created_at": normalize_utc_timestamp(message[timestamp_key], f"{path}_{index}_{timestamp_key}"),
            }
        )
        previous_sequence = sequence
    return {
        "message_count": len(projection),
        "strict_sequence_digest": canonical_json_digest([message["sequence"] for message in projection]),
        "ordered_message_digest": canonical_json_digest(projection),
    }


def project_conversation(value: Any, requested_ref: str) -> dict[str, Any]:
    if not isinstance(value, dict) or set(value) != {"reportId", "reportStatus", "context", "messages"}:
        fail("conversation_api_shape_invalid")
    report_ref = require_uuid(value["reportId"], "conversation_api_report_ref")
    if report_ref != requested_ref:
        fail("conversation_api_report_ref_mismatch")
    report_status = value["reportStatus"]
    if report_status not in {"queued", "generating", "ready", "failed"}:
        fail("conversation_api_report_status_invalid")
    context = project_conversation_context(value["context"], "conversation_api_context")
    messages = project_ordered_messages(
        value["messages"],
        "conversation_api_messages",
        CONVERSATION_MESSAGE_KEYS,
        "createdAt",
    )
    return {
        "report_ref": report_ref,
        "report_status": report_status,
        "context_digest": canonical_json_digest(context),
        "internal_locator_exposed": False,
        **messages,
    }


def project_database_conversation_context(value: Any, session_ref: str) -> tuple[dict[str, Any], dict[str, Any]]:
    if not isinstance(value, dict):
        fail("database_conversation_frozen_context_invalid")
    try:
        public_context = {
            "sourcePlanId": value["plan"]["id"],
            "targetJobTitle": value["targetJob"]["title"],
            "targetJobCompany": value["targetJob"]["company"],
            "resumeId": value["resume"]["id"],
            "resumeDisplayName": value["resume"]["displayName"],
            "roundId": value["round"]["id"],
            "roundSequence": value["round"]["sequence"],
            "roundName": value["round"]["name"],
            "roundType": value["round"]["type"],
            "language": value["conversation"]["language"],
            "hasNextRound": value["hasNextRound"],
        }
        coordinate = value["conversation"]
    except (KeyError, TypeError):
        fail("database_conversation_frozen_context_invalid")
    context = project_conversation_context(public_context, "database_conversation_context")
    if not isinstance(coordinate, dict):
        fail("database_conversation_coordinate_invalid")
    if coordinate.get("sessionId") != session_ref:
        fail("database_conversation_session_binding_invalid")
    return context, coordinate


def project_database_conversation(value: Any, requested_ref: str) -> dict[str, Any]:
    expected_keys = {"report_ref", "session_ref", "status", "generation_context", "messages"}
    if not isinstance(value, dict) or set(value) != expected_keys:
        fail("database_conversation_shape_invalid")
    report_ref = require_uuid(value["report_ref"], "database_conversation_report_ref")
    session_ref = require_uuid(value["session_ref"], "database_conversation_session_ref")
    if report_ref != requested_ref:
        fail("database_conversation_report_ref_mismatch")
    report_status = value["status"]
    if report_status not in {"queued", "generating", "ready", "failed"}:
        fail("database_conversation_report_status_invalid")
    context, coordinate = project_database_conversation_context(value["generation_context"], session_ref)
    messages = project_ordered_messages(
        value["messages"],
        "database_conversation_messages",
        DATABASE_CONVERSATION_MESSAGE_KEYS,
        "created_at",
    )
    message_count = coordinate.get("messageCount")
    last_sequence = coordinate.get("lastMessageSeqNo")
    if (
        not isinstance(message_count, int)
        or isinstance(message_count, bool)
        or message_count != messages["message_count"]
        or not isinstance(last_sequence, int)
        or isinstance(last_sequence, bool)
        or (messages["message_count"] == 0 and last_sequence != 0)
        or (messages["message_count"] > 0 and last_sequence != value["messages"][-1]["sequence"])
    ):
        fail("database_conversation_coordinate_invalid")
    return {
        "report_ref": report_ref,
        "session_ref": session_ref,
        "report_status": report_status,
        "frozen_context_digest": canonical_json_digest(value["generation_context"]),
        "context_digest": canonical_json_digest(context),
        **messages,
    }


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


def database_conversation_query(report_ref: str) -> str:
    report_ref = require_uuid(report_ref, "database_conversation_query_report_ref")
    return f"""
select json_build_object(
  'report_ref', fr.id::text,
  'session_ref', fr.session_id::text,
  'status', fr.status,
  'generation_context', fr.generation_context,
  'messages', coalesce((
    select json_agg(json_build_object(
      'sequence', pm.seq_no,
      'role', pm.role,
      'content', pm.content,
      'created_at', to_char(pm.created_at at time zone 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"')
    ) order by pm.seq_no asc)
    from practice_messages pm
    where pm.session_id = fr.session_id
  ), '[]'::json)
)::text
from feedback_reports fr
join practice_sessions ps on ps.id = fr.session_id
where fr.id = '{report_ref}'::uuid;
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


def capture_database_conversation(report_ref: str, database_url: str) -> dict[str, Any]:
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
            input=database_conversation_query(report_ref),
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
        fail("database_conversation_projection_too_large")
    lines = [line for line in result.stdout.splitlines() if line.strip()]
    if len(lines) != 1:
        fail("database_conversation_projection_row_count_invalid")
    try:
        return project_database_conversation(json.loads(lines[0]), report_ref)
    except json.JSONDecodeError:
        fail("database_conversation_projection_json_invalid")


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


def fetch_report_conversation(api_base_url: str, report_ref: str, cookie_value: str) -> dict[str, Any]:
    url = f"{api_base_url.rstrip('/')}/reports/{urllib.parse.quote(report_ref, safe='')}/conversation"
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
        fail(f"conversation_live_http_status_{exc.code}")
    except (urllib.error.URLError, TimeoutError, OSError):
        raise ManualRequired("live_http_unavailable") from None
    if len(body) > MAX_RESPONSE_BYTES:
        fail("conversation_live_response_too_large")
    try:
        decoded = json.loads(body)
    except (UnicodeDecodeError, json.JSONDecodeError):
        fail("conversation_live_response_json_invalid")
    return project_conversation(decoded, report_ref)


def load_navigation_report_ref(path: Path, run_id: str) -> str:
    if not path.is_file():
        raise ManualRequired("conversation_navigation_missing")
    try:
        navigation = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        fail("conversation_navigation_invalid")
    if not isinstance(navigation, dict):
        fail("conversation_navigation_invalid")
    if (
        navigation.get("schema_version") != "p0-099-conversation-navigation.v1"
        or navigation.get("scenario_id") != "E2E.P0.099"
        or navigation.get("run_id") != run_id
        or navigation.get("method") != "real-browser-report-conversation-back"
    ):
        fail("conversation_navigation_identity_invalid")
    return require_uuid(navigation.get("report_ref"), "conversation_navigation_report_ref")


def bind_conversation_capture(
    database: dict[str, Any],
    api: dict[str, Any],
    report_binding: dict[str, Any],
) -> dict[str, Any]:
    comparable = (
        "report_ref",
        "report_status",
        "context_digest",
        "message_count",
        "strict_sequence_digest",
        "ordered_message_digest",
    )
    if any(database[key] != api[key] for key in comparable):
        fail("conversation_database_api_projection_mismatch")
    if database["report_status"] != "ready" or database["message_count"] < 1:
        fail("conversation_ready_ordered_messages_required")
    if (
        database["session_ref"] != report_binding["session_ref"]
        or database["report_status"] != report_binding["status"]
        or database["frozen_context_digest"] != report_binding["frozen_context_digest"]
    ):
        fail("conversation_report_session_context_binding_mismatch")
    return {
        "report_ref": database["report_ref"],
        "session_ref": database["session_ref"],
        "db": {
            "report_status": database["report_status"],
            "frozen_context_digest": database["frozen_context_digest"],
            "context_digest": database["context_digest"],
            "message_count": database["message_count"],
            "strict_sequence_digest": database["strict_sequence_digest"],
            "ordered_message_digest": database["ordered_message_digest"],
            "read_only": True,
            "ordered_by": "seq_no ASC",
        },
        "api": {
            "report_status": api["report_status"],
            "context_digest": api["context_digest"],
            "message_count": api["message_count"],
            "strict_sequence_digest": api["strict_sequence_digest"],
            "ordered_message_digest": api["ordered_message_digest"],
            "authenticated": True,
            "internal_locator_exposed": api["internal_locator_exposed"],
        },
    }


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
    parser.add_argument("--navigation", type=Path, required=True)
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
        navigation_report_ref = load_navigation_report_ref(args.navigation, args.run_id)
        if navigation_report_ref not in refs:
            fail("conversation_navigation_report_not_in_manifest")
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
        database_conversation = capture_database_conversation(
            navigation_report_ref, os.environ.get("P0_099_DATABASE_URL", "")
        )
        api_conversation = fetch_report_conversation(
            args.api_base_url, navigation_report_ref, cookie_value
        )
        conversation = bind_conversation_capture(
            database_conversation,
            api_conversation,
            database_reports[navigation_report_ref],
        )
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

    write_artifact(args.output, args.run_id, "PASS", "captured", reports, conversation)
    print("P0_099_LIVE_CAPTURE_PASS reports=3 conversation=1 privacy=redacted")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
