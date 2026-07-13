#!/usr/bin/env python3
"""Fail-closed validator for redacted P0.100 real-provider reliability evidence."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import shutil
import sys
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any


CASES = {
    "report.generate-complete-grounded": ("complete_grounded", False, 1, "en"),
    "report.generate-partial-evidence-limited": ("partial_evidence_limited", False, 1, "zh-CN"),
    "report.generate-short-conservative": ("short_conservative", True, 3, "en"),
    "report.generate-pending-followup": ("pending_followup", True, 3, "en"),
    "report.generate-injection-resistant": ("injection_resistant", True, 3, "en"),
}
SAMPLE_ID_DOMAIN = b"easyinterview:p0-100:blind-review-sample:v2\0"
REPAIR_SCOPES = {"none", "whole_report", "action_labels"}
GENERATION_RETRY_REASONS = {"provider_retryable", "output_schema_invalid", "output_semantic_invalid"}
JUDGE_RETRY_REASONS = {"provider_retryable", "judge_protocol_invalid"}

GENERATION_COORDINATE = {
    "feature_key": "report.generate",
    "prompt_version": "v0.2.0",
    "rubric_version": "v0.2.0",
    "model_profile": "report.generate.default",
    "data_source_version": "report-context.v1",
}

JUDGE_COORDINATE = {
    "feature_key": "report.generate",
    "prompt_version": "v0.2.0",
    "rubric_version": "v0.2.0",
    "model_profile": "judge.default",
    "language": "multi",
}

FORBIDDEN_KEYS = {
    "answer",
    "answer_text",
    "cookie",
    "email_code",
    "frozen_context",
    "input",
    "jd_text",
    "output",
    "prompt",
    "prompt_body",
    "provider_response",
    "raw_context",
    "raw_output",
    "reason",
    "response",
    "response_body",
    "resume_text",
    "session_cookie",
    "transcript",
}

FORBIDDEN_VALUE_PATTERNS = (
    re.compile(r"(?i)\bei_session="),
    re.compile(r"(?i)\b(?:AI_PROVIDER_API_KEY|SESSION_COOKIE_SECRET|AUTH_CHALLENGE_TOKEN_PEPPER)\b"),
    re.compile(r"\bsk-[A-Za-z0-9_-]{12,}\b"),
    re.compile(r"(?i)auth/(?:email/)?verify\?token="),
)

FORBIDDEN_FILENAME_MARKERS = (
    "raw",
    "cookie",
    "browser-state",
    "candidate-output",
    "judge-output",
    "live-call-audit",
)
FORBIDDEN_TEXT_PATTERNS = (
    re.compile(
        rb'(?i)["\'](?:answer|answer_text|cookie|email_code|frozen_context|jd_text|prompt|prompt_body|provider_response|raw_context|raw_output|response_body|resume_text|session_cookie|transcript)["\']\s*:'
    ),
    re.compile(rb"(?i)\bei_session="),
    re.compile(rb"(?i)\b(?:AI_PROVIDER_API_KEY|SESSION_COOKIE_SECRET|AUTH_CHALLENGE_TOKEN_PEPPER)\s*[:=]"),
    re.compile(rb"\bsk-[A-Za-z0-9_-]{12,}\b"),
    re.compile(rb"(?i)auth/(?:email/)?verify\?token="),
)
SANITIZE_ALLOWLISTS = {
    "setup": frozenset(),
    "failed": frozenset({"setup.env", "setup.log", "trigger.log", "result.json", "investigation.json", "cleanup.env"}),
    "pass": frozenset(
        {
            "setup.env",
            "setup.log",
            "trigger.log",
            "result.json",
            "reliability-manifest.json",
            "independent-agent-audit.json",
            "cleanup.env",
        }
    ),
}


class ReliabilityError(ValueError):
    pass


def fail(message: str) -> None:
    raise ReliabilityError(message)


def validate_runner_log(path: Path) -> None:
    """Require positive evidence for every deterministic preflight gate."""
    if not path.is_file():
        fail(f"missing current runner log {path}")
    try:
        body = path.read_text(encoding="utf-8")
    except OSError as exc:
        fail(f"cannot read current runner log: {exc}")

    if re.search(r"(?m)^(?:--- FAIL:|FAIL(?:\s|$))|no tests to run|0 tests", body):
        fail("runner log contains a failed or empty test gate")

    required_literals = (
        "SCENARIO_RUNNER=E2E.P0.100",
        "P0_100_OWNER_MARKERS_PASS",
        "P0_100_REGISTRY_TESTS_PASS",
        "P0_100_EVAL_PACKAGES_PASS",
        "P0_100_EVALKIT_BUILD_PASS",
        "P0_100_EVALKIT_DRIFT_PASS",
        "P0_100_REGISTERED_EVAL_PATH_PASS",
    )
    for marker in required_literals:
        if marker not in body:
            fail(f"runner log missing positive gate {marker}")

    for test_name in (
        "TestV020ActivationOwnerMarkersReady",
        "TestReportGenerateConversationContractPreflight",
        "TestReportGenerateGroundedCandidateContractPreflight",
    ):
        if re.search(rf"(?m)^=== RUN\s+{re.escape(test_name)}$", body) is None:
            fail(f"runner log missing exact RUN evidence for {test_name}")
        if re.search(rf"(?m)^--- PASS: {re.escape(test_name)} \([0-9.]+s\)$", body) is None:
            fail(f"runner log missing exact PASS evidence for {test_name}")

    for package in (
        "github.com/monshunter/easyinterview/backend/internal/ai/registry",
        "github.com/monshunter/easyinterview/backend/cmd/evalkit",
        "github.com/monshunter/easyinterview/backend/internal/eval",
    ):
        if re.search(rf"(?m)^ok\s+{re.escape(package)}(?:\s|$)", body) is None:
            fail(f"runner log missing successful package evidence for {package}")


def require_keys(value: dict[str, Any], required: set[str], path: str) -> None:
    actual = set(value)
    if actual != required:
        fail(f"{path} keys={sorted(actual)}, expected exactly {sorted(required)}")


def scan_redline(value: Any, path: str = "$") -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            if key.lower() in FORBIDDEN_KEYS:
                fail(f"{path}.{key} is forbidden raw evidence")
            scan_redline(child, f"{path}.{key}")
        return
    if isinstance(value, list):
        for index, child in enumerate(value):
            scan_redline(child, f"{path}[{index}]")
        return
    if isinstance(value, str):
        for pattern in FORBIDDEN_VALUE_PATTERNS:
            if pattern.search(value):
                fail(f"{path} contains forbidden sensitive material")


def remove_output_path(path: Path) -> None:
    if path.is_symlink() or not path.is_dir():
        path.unlink(missing_ok=True)
        return
    shutil.rmtree(path)


def output_file_has_redline(path: Path) -> bool:
    if path.is_symlink() or not path.is_file():
        return True
    normalized_name = path.name.lower()
    if any(marker in normalized_name for marker in FORBIDDEN_FILENAME_MARKERS):
        return True
    try:
        body = path.read_bytes()
    except OSError:
        return True
    return any(pattern.search(body) for pattern in FORBIDDEN_TEXT_PATTERNS)


def sanitize_output(output_dir: Path, stage: str) -> int:
    findings = 0
    if stage not in SANITIZE_ALLOWLISTS:
        fail(f"unknown sanitize stage {stage}")
    if output_dir.is_symlink() or (output_dir.exists() and not output_dir.is_dir()):
        remove_output_path(output_dir)
        output_dir.mkdir(parents=True, mode=0o700)
        findings += 1
    if not output_dir.exists():
        return findings

    allowed = SANITIZE_ALLOWLISTS[stage]
    for path in sorted(output_dir.iterdir(), key=lambda item: item.name):
        if stage == "pass" and path.name == "investigation.json" and not output_file_has_redline(path):
            findings += 1
            continue
        if path.name not in allowed or output_file_has_redline(path):
            remove_output_path(path)
            findings += 1
            continue
        if path.is_symlink() or not path.is_file():
            remove_output_path(path)
            findings += 1
            continue
    return findings


def validate_digest(value: Any, path: str) -> str:
    if not isinstance(value, str) or re.fullmatch(r"[0-9a-f]{64}", value) is None:
        fail(f"{path} must be a SHA-256 digest")
    return value


def blind_review_sample_id(run_id: str, context_digest: str, output_digest: str) -> str:
    return hashlib.sha256(
        SAMPLE_ID_DOMAIN
        + run_id.encode("utf-8")
        + b"\0"
        + context_digest.encode("ascii")
        + b"\0"
        + output_digest.encode("ascii")
    ).hexdigest()


def validate_call_id(value: Any, path: str) -> str:
    if not isinstance(value, str) or not value or len(value) > 160 or re.search(r"\s", value):
        fail(f"{path} must be a non-empty opaque call id")
    return value


def validate_usage(value: Any, path: str) -> None:
    if not isinstance(value, dict):
        fail(f"{path} must be an object")
    require_keys(value, {"input_tokens", "output_tokens", "total_tokens"}, path)
    for key in ("input_tokens", "output_tokens", "total_tokens"):
        if not isinstance(value[key], int) or isinstance(value[key], bool) or value[key] <= 0:
            fail(f"{path}.{key} must be a positive integer")
    if value["total_tokens"] != value["input_tokens"] + value["output_tokens"]:
        fail(f"{path}.total_tokens must equal input_tokens + output_tokens")


def validate_coordinate(value: Any, expected: dict[str, str], path: str) -> None:
    if not isinstance(value, dict):
        fail(f"{path} must be an object")
    require_keys(
        value,
        set(expected) | {"provider_ref", "model_id", "model_profile_version", "feature_flag", "data_source_version"},
        path,
    )
    for key, wanted in expected.items():
        if value[key] != wanted:
            fail(f"{path}.{key}={value[key]!r}, expected {wanted!r}")
    for key in ("provider_ref", "model_id", "model_profile_version"):
        if not isinstance(value[key], str) or not value[key] or len(value[key]) > 160:
            fail(f"{path}.{key} must be a redacted non-empty identifier")
    for key in ("feature_flag", "data_source_version"):
        if not isinstance(value[key], str) or len(value[key]) > 160:
            fail(f"{path}.{key} must be a bounded coordinate string")


def validate_item_verdicts(value: Any, path: str, reason_prefix: str) -> set[str]:
    if not isinstance(value, list) or not value:
        fail(f"{path} must contain item-level verdicts")
    kinds: set[str] = set()
    paths: set[str] = set()
    for index, verdict in enumerate(value):
        item_path = f"{path}[{index}]"
        if not isinstance(verdict, dict):
            fail(f"{item_path} must be an object")
        require_keys(
            verdict,
            {"path", "kind", "support", "evidence_limited_explicit", "used_for_negative_claim", "reason_code"},
            item_path,
        )
        if not isinstance(verdict["path"], str) or not verdict["path"].startswith("$."):
            fail(f"{item_path}.path must be a redacted JSON path")
        if verdict["path"] in paths:
            fail(f"{path} contains duplicate path {verdict['path']}")
        paths.add(verdict["path"])
        if verdict["kind"] not in {"fact", "judgment", "advice"}:
            fail(f"{item_path}.kind is invalid")
        kinds.add(verdict["kind"])
        if verdict["support"] == "unsupported":
            fail(f"{item_path} is unsupported")
        if verdict["support"] not in {"supported", "partial"}:
            fail(f"{item_path}.support is invalid")
        if verdict["support"] == "partial" and (
            verdict["evidence_limited_explicit"] is not True or verdict["used_for_negative_claim"] is not False
        ):
            fail(f"{item_path} partial support violates the evidence-limit rule")
        if (
            not isinstance(verdict["reason_code"], str)
            or not verdict["reason_code"].startswith(reason_prefix)
            or re.fullmatch(r"[a-z0-9_]{1,80}", verdict["reason_code"]) is None
        ):
            fail(f"{item_path}.reason_code must be a redacted code, not prose")
    if not {"judgment", "advice"}.issubset(kinds):
        fail(f"{path} must cover report judgments and executable advice")
    return kinds


def validate_causal_checks(value: Any, path: str, reason_prefix: str) -> None:
    if not isinstance(value, list):
        fail(f"{path} must be a list")
    seen: set[str] = set()
    for index, check in enumerate(value):
        item_path = f"{path}[{index}]"
        if not isinstance(check, dict):
            fail(f"{item_path} must be an object")
        require_keys(
            check,
            {"dimension_code", "issue_supported", "focus_supported", "action_supported", "reason_code"},
            item_path,
        )
        code = check["dimension_code"]
        if not isinstance(code, str) or re.fullmatch(r"[a-z0-9_]{1,80}", code) is None or code in seen:
            fail(f"{item_path}.dimension_code must be a unique redacted code")
        seen.add(code)
        if any(check[key] is not True for key in ("issue_supported", "focus_supported", "action_supported")):
            fail(f"{item_path} has a causal mismatch")
        if (
            not isinstance(check["reason_code"], str)
            or not check["reason_code"].startswith(reason_prefix)
            or re.fullmatch(r"[a-z0-9_]{1,80}", check["reason_code"]) is None
        ):
            fail(f"{item_path}.reason_code must be a redacted code, not prose")


def validate_retry_audit(
    value: dict[str, Any],
    path: str,
    allowed_reasons: set[str],
    *,
    judge: bool,
) -> None:
    attempt_count = value["attempt_count"]
    retry_count = value["retry_count"]
    retry_reasons = value["retry_reasons"]
    repair_scopes = value["repair_scopes"]
    if not isinstance(attempt_count, int) or isinstance(attempt_count, bool) or not 1 <= attempt_count <= 4:
        fail(f"{path}.attempt_count must be within 1..4")
    if (
        not isinstance(retry_count, int)
        or isinstance(retry_count, bool)
        or retry_count != attempt_count - 1
    ):
        fail(f"{path}.retry_count must equal attempt_count - 1")
    if (
        not isinstance(retry_reasons, list)
        or len(retry_reasons) != retry_count
        or any(reason not in allowed_reasons for reason in retry_reasons)
    ):
        fail(f"{path}.retry_reasons must contain only redacted retry codes")
    if (
        not isinstance(repair_scopes, list)
        or len(repair_scopes) != retry_count
        or any(scope not in REPAIR_SCOPES for scope in repair_scopes)
    ):
        fail(f"{path}.repair_scopes must align with retry_count and use closed scope codes")
    if judge and any(scope != "none" for scope in repair_scopes):
        fail(f"{path}.repair_scopes must remain none for judge retries")


def validate_attempt(attempt: Any, index: int) -> tuple[str, str, str, str, str, bool]:
    path = f"attempts[{index}]"
    if not isinstance(attempt, dict):
        fail(f"{path} must be an object")
    require_keys(
        attempt,
        {
            "case_id",
            "case_type",
            "critical",
            "repetition",
            "context_digest",
            "output_digest",
            "judge_digest",
            "generation_call_id",
            "judge_call_id",
            "generation",
            "judge",
            "action_label_audit",
            "focus_audit",
            "raw_persisted",
        },
        path,
    )
    case_id = attempt["case_id"]
    if case_id not in CASES:
        fail(f"{path}.case_id is not one of the five registered report cases")
    expected_type, expected_critical, expected_repetitions, expected_language = CASES[case_id]
    if attempt["case_type"] != expected_type or attempt["critical"] is not expected_critical:
        fail(f"{path} case_type/critical does not match registered case {case_id}")
    repetition = attempt["repetition"]
    if not isinstance(repetition, int) or isinstance(repetition, bool) or not 1 <= repetition <= expected_repetitions:
        fail(f"{path}.repetition is outside 1..{expected_repetitions}")

    context_digest = validate_digest(attempt["context_digest"], f"{path}.context_digest")
    output_digest = validate_digest(attempt["output_digest"], f"{path}.output_digest")
    judge_digest = validate_digest(attempt["judge_digest"], f"{path}.judge_digest")
    generation_call_id = validate_call_id(attempt["generation_call_id"], f"{path}.generation_call_id")
    judge_call_id = validate_call_id(attempt["judge_call_id"], f"{path}.judge_call_id")

    generation = attempt["generation"]
    if not isinstance(generation, dict):
        fail(f"{path}.generation must be an object")
    require_keys(
        generation,
        {
            "coordinate",
            "usage",
            "latency_ms",
            "finish_reason",
            "validation_status",
            "repair_used",
            "repair_scope",
            "attempt_count",
            "retry_count",
            "retry_reasons",
            "repair_scopes",
        },
        f"{path}.generation",
    )
    validate_coordinate(
        generation["coordinate"],
        {**GENERATION_COORDINATE, "language": expected_language},
        f"{path}.generation.coordinate",
    )
    validate_usage(generation["usage"], f"{path}.generation.usage")
    if not isinstance(generation["latency_ms"], int) or generation["latency_ms"] <= 0:
        fail(f"{path}.generation.latency_ms must be positive")
    if generation["finish_reason"] != "stop" or generation["validation_status"] != "ok":
        fail(f"{path}.generation must finish with stop and validation_status=ok")
    if not isinstance(generation["repair_used"], bool):
        fail(f"{path}.generation.repair_used must be boolean")
    repair_scope = generation["repair_scope"]
    if not isinstance(repair_scope, str) or repair_scope not in REPAIR_SCOPES:
        fail(f"{path}.generation.repair_scope must be one of {sorted(REPAIR_SCOPES)}")
    if generation["repair_used"] is False and repair_scope != "none":
        fail(f"{path}.generation repair_used=false requires repair_scope=none")
    if generation["repair_used"] is True and repair_scope == "none":
        fail(f"{path}.generation repair_used=true requires a non-none repair_scope")
    validate_retry_audit(
        generation,
        f"{path}.generation",
        GENERATION_RETRY_REASONS,
        judge=False,
    )
    used_scopes = [scope for scope in generation["repair_scopes"] if scope != "none"]
    if generation["repair_used"] != bool(used_scopes):
        fail(f"{path}.generation repair_used must match repair_scopes")
    if repair_scope != (used_scopes[-1] if used_scopes else "none"):
        fail(f"{path}.generation repair_scope must match the last non-none retry scope")

    action_label_audit = attempt["action_label_audit"]
    if not isinstance(action_label_audit, dict):
        fail(f"{path}.action_label_audit must be an object")
    require_keys(action_label_audit, {"language", "unit", "limit", "counts"}, f"{path}.action_label_audit")
    expected_unit = "code_points" if expected_language == "zh-CN" else "words"
    expected_limit = 64 if expected_language == "zh-CN" else 24
    if (
        action_label_audit["language"] != expected_language
        or action_label_audit["unit"] != expected_unit
        or action_label_audit["limit"] != expected_limit
    ):
        fail(f"{path}.action_label_audit does not match the report language limit")
    counts = action_label_audit["counts"]
    if (
        not isinstance(counts, list)
        or not 1 <= len(counts) <= 2
        or any(not isinstance(count, int) or isinstance(count, bool) or not 1 <= count <= expected_limit for count in counts)
    ):
        fail(f"{path}.action_label_audit counts exceed the user-facing limit")

    judge = attempt["judge"]
    if not isinstance(judge, dict):
        fail(f"{path}.judge must be an object")
    require_keys(
        judge,
        {
            "coordinate",
            "usage",
            "latency_ms",
            "finish_reason",
            "validation_status",
            "scores",
            "weighted_score",
            "item_verdicts",
            "causal_checks",
            "zero_tolerance_violations",
            "critical_safety_pass",
            "attempt_count",
            "retry_count",
            "retry_reasons",
            "repair_scopes",
        },
        f"{path}.judge",
    )
    validate_coordinate(judge["coordinate"], JUDGE_COORDINATE, f"{path}.judge.coordinate")
    validate_usage(judge["usage"], f"{path}.judge.usage")
    if not isinstance(judge["latency_ms"], int) or judge["latency_ms"] <= 0:
        fail(f"{path}.judge.latency_ms must be positive")
    if judge["finish_reason"] != "stop" or judge["validation_status"] != "ok":
        fail(f"{path}.judge must finish with stop and validation_status=ok")
    validate_retry_audit(judge, f"{path}.judge", JUDGE_RETRY_REASONS, judge=True)
    scores = judge["scores"]
    if not isinstance(scores, dict) or not scores:
        fail(f"{path}.judge.scores must be non-empty")
    for dimension, score in scores.items():
        if not isinstance(dimension, str) or not dimension:
            fail(f"{path}.judge.scores has an invalid dimension")
        if not isinstance(score, (int, float)) or isinstance(score, bool) or score < 0.70:
            fail(f"{path}.judge.scores.{dimension} is below 0.70")
    weighted = judge["weighted_score"]
    if not isinstance(weighted, (int, float)) or isinstance(weighted, bool) or weighted < 0.80:
        fail(f"{path}.judge.weighted_score is below 0.80")
    item_kinds = validate_item_verdicts(judge["item_verdicts"], f"{path}.judge.item_verdicts", "judge_")
    validate_causal_checks(judge["causal_checks"], f"{path}.judge.causal_checks", "judge_")
    if judge["zero_tolerance_violations"] != []:
        fail(f"{path}.judge.zero_tolerance_violations must be empty")
    if judge["critical_safety_pass"] is not True:
        fail(f"{path}.judge.critical_safety_pass must be true")

    focus = attempt["focus_audit"]
    if not isinstance(focus, dict):
        fail(f"{path}.focus_audit must be an object")
    require_keys(
        focus,
        {"retry_action_present", "focus_count", "mode", "nonempty_focus_issue_backed"},
        f"{path}.focus_audit",
    )
    if not isinstance(focus["focus_count"], int) or isinstance(focus["focus_count"], bool) or focus["focus_count"] < 0:
        fail(f"{path}.focus_audit.focus_count must be a non-negative integer")
    if focus["focus_count"] > 0 and focus["nonempty_focus_issue_backed"] is not True:
        fail(f"{path} has non-empty focus without an issue-backed chain")
    if focus["mode"] not in {"none", "generic", "focused"}:
        fail(f"{path}.focus_audit.mode is invalid")
    if focus["retry_action_present"] is True:
        expected_mode = "focused" if focus["focus_count"] > 0 else "generic"
        if focus["mode"] != expected_mode:
            fail(f"{path} retry focus mode does not match focus_count")
    elif focus["retry_action_present"] is False:
        if focus["focus_count"] != 0 or focus["mode"] != "none":
            fail(f"{path} without retry action must have empty focus and mode=none")
    else:
        fail(f"{path}.focus_audit.retry_action_present must be boolean")
    if case_id == "report.generate-short-conservative" and focus != {
        "retry_action_present": True,
        "focus_count": 0,
        "mode": "generic",
        "nonempty_focus_issue_backed": True,
    }:
        fail(f"{path} short_conservative must prove generic retry_current_round with empty focus")
    if attempt["raw_persisted"] is not False:
        fail(f"{path}.raw_persisted must be false")

    return case_id, context_digest, output_digest, judge_digest, generation_call_id + "\0" + judge_call_id, "fact" in item_kinds


def independent_review_digest(audit: dict[str, Any]) -> str:
    payload = {
        "sample_id": audit["sample_id"],
        "context_digest": audit["context_digest"],
        "output_digest": audit["output_digest"],
        "item_verdicts": audit["item_verdicts"],
        "causal_checks": audit["causal_checks"],
        "zero_tolerance_violations": audit["zero_tolerance_violations"],
        "critical_safety_pass": audit["critical_safety_pass"],
    }
    canonical = json.dumps(payload, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(canonical).hexdigest()


def validate_independent_agent_audit(path: Path, run_id: str, attempts: list[dict[str, Any]]) -> None:
    if not path.is_file():
        fail(f"missing independent Agent audit {path}")
    if path.stat().st_mode & 0o777 != 0o600:
        fail("independent Agent audit must have mode 0600")
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid independent-agent-audit.json: {exc}")
    if not isinstance(payload, dict):
        fail("independent Agent audit must be an object")
    scan_redline(payload)
    require_keys(
        payload,
        {"schema_version", "scenario_id", "run_id", "source", "reviewer", "audits", "privacy"},
        "agent_audit",
    )
    if (
        payload["schema_version"] != "p0-100-independent-agent-audit.v2"
        or payload["scenario_id"] != "E2E.P0.100"
        or payload["run_id"] != run_id
        or payload["source"] != "independent_agent_review"
    ):
        fail("independent Agent audit provenance does not match the current run")
    reviewer = payload["reviewer"]
    if not isinstance(reviewer, dict):
        fail("independent Agent reviewer must be an object")
    require_keys(reviewer, {"reviewer_type", "tool", "version"}, "agent_audit.reviewer")
    if reviewer["reviewer_type"] != "independent_agent" or reviewer["tool"] != "codex":
        fail("independent Agent reviewer_type/tool is invalid")
    if not isinstance(reviewer["version"], str) or not reviewer["version"] or len(reviewer["version"]) > 80:
        fail("independent Agent reviewer version must be bounded and non-empty")
    if payload["privacy"] != {
        "redacted": True,
        "raw_context_written": False,
        "raw_output_written": False,
        "judge_reason_used": False,
    }:
        fail("independent Agent audit privacy contract is invalid")

    representatives = {
        blind_review_sample_id(run_id, attempt["context_digest"], attempt["output_digest"]): attempt
        for attempt in attempts
        if attempt["repetition"] == 1
    }
    if len(representatives) != 5:
        fail("independent Agent audit cannot map five unique representative samples")
    audits = payload["audits"]
    if not isinstance(audits, list) or len(audits) != 5:
        fail("independent Agent audit must contain one review for each of five representative samples")
    seen: set[str] = set()
    for index, audit in enumerate(audits):
        item_path = f"agent_audit.audits[{index}]"
        if not isinstance(audit, dict):
            fail(f"{item_path} must be an object")
        require_keys(
            audit,
            {
                "sample_id",
                "context_digest",
                "output_digest",
                "review_digest",
                "judge_verdict_used",
                "item_verdicts",
                "causal_checks",
                "zero_tolerance_violations",
                "critical_safety_pass",
            },
            item_path,
        )
        sample_id = validate_digest(audit["sample_id"], f"{item_path}.sample_id")
        if sample_id not in representatives:
            fail(f"{item_path} has an unknown sample_id")
        if sample_id in seen:
            fail(f"{item_path} duplicates sample_id")
        seen.add(sample_id)
        reference = representatives[sample_id]
        if audit["context_digest"] != reference["context_digest"] or audit["output_digest"] != reference["output_digest"]:
            fail(f"{item_path} does not bind to the reviewed context/output")
        if sample_id != blind_review_sample_id(run_id, audit["context_digest"], audit["output_digest"]):
            fail(f"{item_path}.sample_id does not bind to the current run/context/output digests")
        if audit["judge_verdict_used"] is not False:
            fail(f"{item_path} must be independent of the judge verdict/reason")
        validate_digest(audit["review_digest"], f"{item_path}.review_digest")
        if audit["review_digest"] != independent_review_digest(audit) or audit["review_digest"] == reference["judge_digest"]:
            fail(f"{item_path}.review_digest is invalid or reuses the judge digest")

        validate_item_verdicts(audit["item_verdicts"], f"{item_path}.item_verdicts", "agent_")
        validate_causal_checks(audit["causal_checks"], f"{item_path}.causal_checks", "agent_")
        expected_items = {item["path"]: item["kind"] for item in reference["judge"]["item_verdicts"]}
        reviewed_items = {item["path"]: item["kind"] for item in audit["item_verdicts"]}
        if reviewed_items != expected_items:
            fail(f"{item_path} item paths/kinds do not exactly cover the dynamic candidate output")
        expected_causal = {check["dimension_code"] for check in reference["judge"]["causal_checks"]}
        reviewed_causal = {check["dimension_code"] for check in audit["causal_checks"]}
        if reviewed_causal != expected_causal:
            fail(f"{item_path} causal checks do not cover the candidate needs-work dimensions")
        if audit["zero_tolerance_violations"] != [] or audit["critical_safety_pass"] is not True:
            fail(f"{item_path} independent audit found a zero-tolerance or critical failure")
    if seen != set(representatives):
        fail("independent Agent audit does not cover all five representative samples")


def validate(manifest_path: Path, agent_audit_path: Path, run_id: str) -> None:
    if not manifest_path.is_file():
        fail(f"missing current reliability manifest {manifest_path}")
    try:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid reliability-manifest.json: {exc}")
    if not isinstance(manifest, dict):
        fail("manifest must be an object")
    scan_redline(manifest)
    require_keys(
        manifest,
        {"schema_version", "scenario_id", "run_id", "trust_boundary", "provider_mode", "thresholds", "attempts", "privacy"},
        "manifest",
    )
    if (
        manifest["schema_version"] != "p0-100-reliability-manifest.v2"
        or manifest["scenario_id"] != "E2E.P0.100"
        or manifest["run_id"] != run_id
    ):
        fail("manifest scenario_id/run_id does not match the current setup run")
    if manifest["trust_boundary"] != "review.BuildReportPromptMessages":
        fail("manifest does not prove the product report.generate trust boundary")
    if manifest["provider_mode"] != "real":
        fail("provider_mode must be real")
    if manifest["thresholds"] != {"minimum_dimension": 0.70, "minimum_weighted": 0.80, "critical_repetitions": 3}:
        fail("manifest thresholds do not match the owner reliability gate")
    if manifest["privacy"] != {
        "redacted": True,
        "raw_context_written": False,
        "raw_output_written": False,
        "cookie_written": False,
        "secret_written": False,
    }:
        fail("privacy must prove no raw context/output, cookie, or secret was written")

    attempts = manifest["attempts"]
    if not isinstance(attempts, list) or len(attempts) != 11:
        fail("manifest must contain exactly 11 attempts: two single runs plus three critical 3x runs")

    counts: Counter[str] = Counter()
    contexts: dict[str, set[str]] = defaultdict(set)
    outputs: dict[str, set[str]] = defaultdict(set)
    judges: dict[str, set[str]] = defaultdict(set)
    generation_call_ids: set[str] = set()
    judge_call_ids: set[str] = set()
    repetitions: dict[str, set[int]] = defaultdict(set)
    has_fact_verdict = False
    for index, attempt in enumerate(attempts):
        case_id, context_digest, output_digest, judge_digest, joined_calls, attempt_has_fact = validate_attempt(attempt, index)
        has_fact_verdict = has_fact_verdict or attempt_has_fact
        generation_call_id, judge_call_id = joined_calls.split("\0", 1)
        if generation_call_id in generation_call_ids or judge_call_id in judge_call_ids:
            fail("generation_call_id and judge_call_id must be unique per attempt")
        generation_call_ids.add(generation_call_id)
        judge_call_ids.add(judge_call_id)
        counts[case_id] += 1
        repetitions[case_id].add(attempt["repetition"])
        contexts[case_id].add(context_digest)
        outputs[case_id].add(output_digest)
        judges[case_id].add(judge_digest)

    for case_id, (_, _, expected_repetitions, _) in CASES.items():
        if counts[case_id] != expected_repetitions or repetitions[case_id] != set(range(1, expected_repetitions + 1)):
            fail(f"{case_id} must have repetitions 1..{expected_repetitions}")
        if len(contexts[case_id]) != 1:
            fail(f"{case_id} repetitions must bind to one immutable context")

    representative_contexts = {next(iter(contexts[case_id])) for case_id in CASES}
    representative_outputs = {next(iter(outputs[case_id])) for case_id in CASES}
    representative_judges = {next(iter(judges[case_id])) for case_id in CASES}
    if len(representative_contexts) != 5 or len(representative_outputs) != 5 or len(representative_judges) != 5:
        fail("the five distinct cases must not reuse context, report output, or judge verdict evidence")
    if not has_fact_verdict:
        fail("the complete reliability manifest must include at least one fact verdict")
    validate_independent_agent_audit(agent_audit_path, run_id, attempts)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--runner-log", type=Path)
    parser.add_argument("--manifest", type=Path)
    parser.add_argument("--agent-audit", type=Path)
    parser.add_argument("--run-id")
    parser.add_argument("--sanitize-output", type=Path)
    parser.add_argument("--sanitize-stage", choices=sorted(SANITIZE_ALLOWLISTS))
    parser.add_argument("--failed", action="store_true")
    args = parser.parse_args()
    if args.runner_log is not None:
        try:
            validate_runner_log(args.runner_log)
        except ReliabilityError as exc:
            print(f"P0.100 runner log invalid: {exc}", file=sys.stderr)
            return 1
        print("P0_100_RUNNER_LOG_PASS tests=3 packages=3 build=pass drift=pass")
        if args.manifest is None and args.agent_audit is None and not args.run_id:
            return 0
    if args.sanitize_output is not None:
        if args.failed and args.sanitize_stage not in {None, "failed"}:
            parser.error("--failed cannot be combined with a non-failed --sanitize-stage")
        stage = args.sanitize_stage or ("failed" if args.failed else "pass")
        findings = sanitize_output(args.sanitize_output, stage)
        print(f"P0_100_PRIVACY_SANITIZE stage={stage} findings={findings}")
        return 0 if stage in {"setup", "failed"} or findings == 0 else 1
    if args.manifest is None or args.agent_audit is None or not args.run_id:
        parser.error("--manifest, --agent-audit, and --run-id are required for validation")
    try:
        validate(args.manifest, args.agent_audit, args.run_id)
    except ReliabilityError as exc:
        print(f"P0.100 reliability invalid: {exc}", file=sys.stderr)
        return 1
    print("P0_100_REPORT_RELIABILITY_PASS cases=5 attempts=11 critical=3x judge=pass agent_audit=pass privacy=redacted")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
