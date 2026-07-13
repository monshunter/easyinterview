#!/usr/bin/env python3
"""Run five report cases through evalkit without persisting raw context/output."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import stat
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any

import yaml


CASES = (
    ("report.generate-complete-grounded", "complete_grounded", False, 1, "en"),
    ("report.generate-partial-evidence-limited", "partial_evidence_limited", False, 1, "zh-CN"),
    ("report.generate-short-conservative", "short_conservative", True, 3, "en"),
    ("report.generate-pending-followup", "pending_followup", True, 3, "en"),
    ("report.generate-injection-resistant", "injection_resistant", True, 3, "en"),
)
CASE_LANGUAGES = {case_id: language for case_id, _, _, _, language in CASES}
SAMPLE_ID_DOMAIN = b"easyinterview:p0-100:blind-review-sample:v2\0"
REPAIR_SCOPES = {"none", "whole_report", "action_labels"}
GENERATION_RETRY_REASONS = {"provider_retryable", "output_schema_invalid", "output_semantic_invalid"}
JUDGE_RETRY_REASONS = {"provider_retryable", "judge_protocol_invalid"}
BLIND_SAMPLE_KEYS = {
    "language",
    "context_digest",
    "output_digest",
    "context",
    "transcript",
    "output",
    "generation",
}
GENERATION_SUMMARY_KEYS = {
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
}
GENERATION_COORDINATE_KEYS = {
    "feature_key",
    "prompt_version",
    "rubric_version",
    "model_profile",
    "model_profile_version",
    "language",
    "feature_flag",
    "data_source_version",
    "provider_ref",
    "model_id",
}
GENERATION_USAGE_KEYS = {"input_tokens", "output_tokens", "total_tokens"}


class LiveRunError(RuntimeError):
    pass


def fail(message: str) -> None:
    raise LiveRunError(message)


def blind_review_sample_id(run_id: str, context_digest: str, output_digest: str) -> str:
    """Bind an opaque review handle to this run and one representative output."""
    return hashlib.sha256(
        SAMPLE_ID_DOMAIN
        + run_id.encode("utf-8")
        + b"\0"
        + context_digest.encode("ascii")
        + b"\0"
        + output_digest.encode("ascii")
    ).hexdigest()


def blind_generation_summary(value: Any, index: int) -> dict[str, Any]:
    if not isinstance(value, dict) or set(value) != GENERATION_SUMMARY_KEYS:
        fail(f"blind review source sample {index} has non-current generation metadata")
    coordinate_value = value["coordinate"]
    usage_value = value["usage"]
    if not isinstance(coordinate_value, dict) or set(coordinate_value) != GENERATION_COORDINATE_KEYS:
        fail(f"blind review source sample {index} has non-current generation metadata")
    if not isinstance(usage_value, dict) or set(usage_value) != GENERATION_USAGE_KEYS:
        fail(f"blind review source sample {index} has non-current generation metadata")
    if any(not isinstance(coordinate_value[key], str) or len(coordinate_value[key]) > 160 for key in coordinate_value):
        fail(f"blind review source sample {index} has invalid generation metadata")
    if (
        coordinate_value["feature_key"] != "report.generate"
        or coordinate_value["prompt_version"] != "v0.2.0"
        or coordinate_value["rubric_version"] != "v0.2.0"
        or coordinate_value["model_profile"] != "report.generate.default"
        or coordinate_value["language"] not in {"en", "zh-CN"}
        or coordinate_value["feature_flag"] not in {"", "none"}
        or coordinate_value["data_source_version"] != "report-context.v1"
        or any(not coordinate_value[key] for key in ("model_profile_version", "provider_ref", "model_id"))
        or any(
            re.search(r"(?i)(?:\bsk-[a-z0-9_-]{12,}\b|api[_-]?key|session[_-]?cookie)", coordinate_value[key])
            for key in coordinate_value
        )
    ):
        fail(f"blind review source sample {index} has invalid generation metadata")
    if any(
        not isinstance(usage_value[key], int) or isinstance(usage_value[key], bool) or usage_value[key] <= 0
        for key in GENERATION_USAGE_KEYS
    ) or usage_value["total_tokens"] != usage_value["input_tokens"] + usage_value["output_tokens"]:
        fail(f"blind review source sample {index} has invalid generation metadata")
    if (
        not isinstance(value["latency_ms"], int)
        or isinstance(value["latency_ms"], bool)
        or value["latency_ms"] <= 0
        or value["finish_reason"] != "stop"
        or value["validation_status"] != "ok"
        or not isinstance(value["repair_used"], bool)
        or not isinstance(value["repair_scope"], str)
        or value["repair_scope"] not in REPAIR_SCOPES
        or (value["repair_used"] is False and value["repair_scope"] != "none")
        or (value["repair_used"] is True and value["repair_scope"] == "none")
        or not isinstance(value["attempt_count"], int)
        or isinstance(value["attempt_count"], bool)
        or not 1 <= value["attempt_count"] <= 4
        or not isinstance(value["retry_count"], int)
        or isinstance(value["retry_count"], bool)
        or value["retry_count"] != value["attempt_count"] - 1
        or not isinstance(value["retry_reasons"], list)
        or len(value["retry_reasons"]) != value["retry_count"]
        or any(reason not in GENERATION_RETRY_REASONS for reason in value["retry_reasons"])
        or not isinstance(value["repair_scopes"], list)
        or len(value["repair_scopes"]) != value["retry_count"]
        or any(scope not in REPAIR_SCOPES for scope in value["repair_scopes"])
        or value["repair_used"] != any(scope != "none" for scope in value["repair_scopes"])
    ):
        fail(f"blind review source sample {index} has invalid generation metadata")
    return {
        "coordinate": {key: coordinate_value[key] for key in sorted(GENERATION_COORDINATE_KEYS)},
        "usage": {key: usage_value[key] for key in sorted(GENERATION_USAGE_KEYS)},
        "latency_ms": value["latency_ms"],
        "finish_reason": value["finish_reason"],
        "validation_status": value["validation_status"],
        "repair_used": value["repair_used"],
        "repair_scope": value["repair_scope"],
        "attempt_count": value["attempt_count"],
        "retry_count": value["retry_count"],
        "retry_reasons": list(value["retry_reasons"]),
        "repair_scopes": list(value["repair_scopes"]),
    }


def build_blind_review_packet(run_id: str, review_samples: list[dict[str, Any]]) -> dict[str, Any]:
    """Build the temporary packet without case intent, gold, or judge metadata."""
    if not run_id or len(review_samples) != 5:
        fail("blind review packet requires a run id and exactly five representative samples")
    samples: list[dict[str, Any]] = []
    for index, source in enumerate(review_samples):
        if not isinstance(source, dict) or set(source) != BLIND_SAMPLE_KEYS:
            fail(f"blind review source sample {index} contains non-current metadata")
        context_digest = source["context_digest"]
        output_digest = source["output_digest"]
        if (
            not isinstance(context_digest, str)
            or re.fullmatch(r"[0-9a-f]{64}", context_digest) is None
            or not isinstance(output_digest, str)
            or re.fullmatch(r"[0-9a-f]{64}", output_digest) is None
        ):
            fail(f"blind review source sample {index} has an invalid digest")
        if source["language"] not in {"en", "zh-CN"}:
            fail(f"blind review source sample {index} has an invalid language coordinate")
        generation = blind_generation_summary(source["generation"], index)
        if generation["coordinate"]["language"] != source["language"]:
            fail(f"blind review source sample {index} has inconsistent language metadata")
        samples.append(
            {
                "sample_id": blind_review_sample_id(run_id, context_digest, output_digest),
                "language": source["language"],
                "context_digest": context_digest,
                "output_digest": output_digest,
                "context": source["context"],
                "transcript": source["transcript"],
                "output": source["output"],
                "generation": generation,
            }
        )
    samples.sort(key=lambda sample: sample["sample_id"])
    if len({sample["sample_id"] for sample in samples}) != 5:
        fail("blind review sample ids must be unique")
    return {
        "schema_version": "p0-100-agent-review-packet.v3",
        "scenario_id": "E2E.P0.100",
        "run_id": run_id,
        "source": "blind_independent_agent_review_handoff",
        "samples": samples,
        "privacy": {
            "synthetic_redacted_inputs": True,
            "contains_secret": False,
            "selection_metadata_exposed": False,
            "evaluation_material_exposed": False,
        },
    }


def classify_evalkit_failure(stdout: bytes, stderr: bytes) -> str:
    reason = ""
    try:
        payload = json.loads(stdout)
        if isinstance(payload, dict) and isinstance(payload.get("reason"), str):
            reason = payload["reason"]
    except (UnicodeDecodeError, json.JSONDecodeError):
        pass
    combined = reason + "\n" + stderr.decode("utf-8", errors="replace")
    lowered = combined.lower()
    dimension = re.search(r"report dimension ([a-z0-9_]+) score ([0-9.]+) is below 0\.70", lowered)
    if dimension:
        return f"reason_code=dimension_below_threshold dimension={dimension.group(1)} score={dimension.group(2)}"
    weighted = re.search(r"report weighted score ([0-9.]+) is below 0\.80", lowered)
    if weighted:
        return f"reason_code=weighted_below_threshold score={weighted.group(1)}"
    unsupported = re.search(r'unsupported report item ["\']([^"\']+)["\']', combined, re.IGNORECASE)
    if unsupported:
        return f"reason_code=unsupported_item path={unsupported.group(1)}"
    partial = re.search(r'partial item ["\']([^"\']+)["\']', combined, re.IGNORECASE)
    if partial:
        return f"reason_code=invalid_partial path={partial.group(1)}"
    causal = re.search(r'causal check ["\']([^"\']+)["\'] failed', combined, re.IGNORECASE)
    if causal:
        return f"reason_code=causal_mismatch dimension={causal.group(1)}"
    if "zero-tolerance violation" in lowered:
        return "reason_code=zero_tolerance_violation"
    if "critical safety verdict did not pass" in lowered:
        return "reason_code=critical_safety_failed"
    if "rate_limited" in lowered or "rate limit" in lowered:
        return "reason_code=rate_limited"
    if "ai_provider_config_invalid" in lowered:
        return "reason_code=provider_config_invalid"
    if "ai_provider_secret_missing" in lowered:
        return "reason_code=provider_secret_missing"
    if "ai_unsupported_capability" in lowered:
        return "reason_code=unsupported_capability"
    if "fallback exhausted" in lowered:
        return "reason_code=fallback_exhausted"
    if "context deadline exceeded" in lowered or "timed out" in lowered or "timeout" in lowered:
        return "reason_code=timeout"
    if "empty judge response" in lowered:
        return "reason_code=empty_judge_response"
    if "evaluated output failed schema" in lowered:
        return "reason_code=evaluated_output_schema_invalid"
    if "strict grounded report output" in lowered or "grounded report output has trailing content" in lowered:
        return "reason_code=evaluated_output_contract_invalid"
    if "parse judge response" in lowered or "trailing judge response content" in lowered:
        return "reason_code=judge_response_parse_invalid"
    if "invalid or duplicate item verdict" in lowered:
        return "reason_code=judge_item_coordinate_invalid"
    if "item verdict count" in lowered and "does not cover" in lowered:
        return "reason_code=judge_item_coverage_invalid"
    if "invalid or duplicate causal check" in lowered or "causal checks cover" in lowered:
        return "reason_code=judge_causal_coverage_invalid"
    if "llm judge response is invalid" in lowered or ("judge response" in lowered and "invalid" in lowered):
        return "reason_code=judge_output_invalid"
    if "output invalid" in lowered or "output schema invalid" in lowered or "validation" in lowered or "strict json" in lowered:
        return "reason_code=output_invalid"
    if "status code" in lowered or "provider" in lowered or "read response" in lowered:
        return "reason_code=provider_error"
    return "reason_code=unknown"


def load_case_contexts(repo_root: Path) -> dict[str, dict[str, Any]]:
    path = repo_root / "config" / "evals" / "report.generate" / "cases.yaml"
    payload = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(payload, dict) or payload.get("feature_key") != "report.generate":
        fail("registered report.generate case file is invalid")
    contexts: dict[str, dict[str, Any]] = {}
    expected_languages = {case_id: language for case_id, _, _, _, language in CASES}
    for case in payload.get("cases", []):
        if not isinstance(case, dict) or not isinstance(case.get("id"), str):
            fail("registered report.generate case entry is invalid")
        if case.get("language") != expected_languages.get(case["id"]):
            fail(f"registered report.generate case {case['id']} language coordinate is invalid")
        canonical = json.dumps(
            {"context": case.get("context"), "transcript": case.get("transcript")},
            ensure_ascii=False,
            sort_keys=True,
            separators=(",", ":"),
        ).encode("utf-8")
        contexts[case["id"]] = {
            "digest": hashlib.sha256(canonical).hexdigest(),
            "context": case.get("context"),
            "transcript": case.get("transcript"),
        }
    expected = {case_id for case_id, _, _, _, _ in CASES}
    if set(contexts) != expected or len({entry["digest"] for entry in contexts.values()}) != 5:
        fail("registered report.generate suite must contain exactly five distinct contexts")
    return contexts


def run_evalkit(
    command: list[str],
    *,
    repo_root: Path,
    stage: str,
    output_digest: str = "",
    input_bytes: bytes | None = None,
    timeout: int,
) -> bytes:
    try:
        completed = subprocess.run(
            command,
            cwd=repo_root,
            input=input_bytes,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            timeout=timeout,
            env={**os.environ, "AI_DEBUG_PRINT_RAW_OUTPUT": "false"},
        )
    except subprocess.TimeoutExpired:
        fail(f"evalkit stage timed out after {timeout}s")
    if completed.returncode != 0:
        classification = classify_evalkit_failure(completed.stdout, completed.stderr)
        digest = f" output_digest={output_digest}" if output_digest else ""
        fail(f"evalkit {stage} failed {classification}{digest}; raw stderr/reason prose was discarded")
    return completed.stdout


def load_audit(path: Path, stage: str, case_id: str, critical: bool) -> dict[str, Any]:
    if not path.is_file() or stat.S_IMODE(path.stat().st_mode) != 0o600:
        fail(f"{stage} audit must exist with mode 0600")
    try:
        audit = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        fail(f"{stage} audit is not valid JSON")
    required = {
        "schemaVersion",
        "stage",
        "caseId",
        "critical",
        "pass",
        "featureKey",
        "promptVersion",
        "rubricVersion",
        "language",
        "provider",
        "modelId",
        "modelProfileName",
        "modelProfileVersion",
        "finishReason",
        "inputTokens",
        "outputTokens",
        "latencyMs",
        "validationStatus",
        "repairUsed",
        "repairScope",
        "attemptCount",
        "retryCount",
        "retryReasons",
        "repairScopes",
        "outputSha256",
        "outputBytes",
    }
    if not isinstance(audit, dict) or not required.issubset(audit):
        fail(f"{stage} audit is missing required redacted fields")
    allowed = required | {"errorClass", "featureFlag", "dataSourceVersion"}
    if set(audit) - allowed:
        fail(f"{stage} audit has unexpected audit fields")
    if (
        audit["schemaVersion"] != "evalkit-live-call-audit.v2"
        or audit["stage"] != stage
        or audit["caseId"] != case_id
        or audit["critical"] is not critical
        or audit["pass"] is not True
        or audit["featureKey"] != "report.generate"
        or audit["promptVersion"] != "v0.2.0"
        or audit["rubricVersion"] != "v0.2.0"
        or audit["language"] != ("multi" if stage == "judge" else CASE_LANGUAGES[case_id])
        or audit["finishReason"] != "stop"
        or audit["validationStatus"] != "ok"
        or not isinstance(audit["inputTokens"], int)
        or audit["inputTokens"] <= 0
        or not isinstance(audit["outputTokens"], int)
        or audit["outputTokens"] <= 0
        or not isinstance(audit["latencyMs"], int)
        or audit["latencyMs"] <= 0
        or not isinstance(audit["outputBytes"], int)
        or audit["outputBytes"] <= 0
        or not isinstance(audit["outputSha256"], str)
        or len(audit["outputSha256"]) != 64
    ):
        fail(f"{stage} audit failed its live-call provenance gate")
    repair_used = audit["repairUsed"]
    repair_scope = audit["repairScope"]
    if (
        not isinstance(repair_used, bool)
        or not isinstance(repair_scope, str)
        or repair_scope not in REPAIR_SCOPES
        or (stage == "completion" and repair_used is False and repair_scope != "none")
        or (stage == "completion" and repair_used is True and repair_scope == "none")
        or (stage == "judge" and (repair_used is not False or repair_scope != "none"))
    ):
        fail(f"{stage} audit has invalid repair provenance")
    for key in ("provider", "modelId", "modelProfileName", "modelProfileVersion"):
        if not isinstance(audit[key], str) or not audit[key]:
            fail(f"{stage} audit missing {key}")
    attempt_count = audit["attemptCount"]
    retry_count = audit["retryCount"]
    retry_reasons = audit["retryReasons"]
    repair_scopes = audit["repairScopes"]
    allowed_reasons = JUDGE_RETRY_REASONS if stage == "judge" else GENERATION_RETRY_REASONS
    if (
        not isinstance(attempt_count, int)
        or isinstance(attempt_count, bool)
        or not 1 <= attempt_count <= 4
        or retry_count != attempt_count - 1
        or not isinstance(retry_reasons, list)
        or len(retry_reasons) != retry_count
        or any(reason not in allowed_reasons for reason in retry_reasons)
        or not isinstance(repair_scopes, list)
        or len(repair_scopes) != retry_count
        or any(scope not in REPAIR_SCOPES for scope in repair_scopes)
        or (stage == "judge" and any(scope != "none" for scope in repair_scopes))
        or (stage == "completion" and repair_used != any(scope != "none" for scope in repair_scopes))
        or (
            stage == "completion"
            and repair_scope != next((scope for scope in reversed(repair_scopes) if scope != "none"), "none")
        )
    ):
        fail(f"{stage} audit has invalid bounded retry provenance")
    return audit


def coordinate(audit: dict[str, Any]) -> dict[str, Any]:
    return {
        "feature_key": audit["featureKey"],
        "prompt_version": audit["promptVersion"],
        "rubric_version": audit["rubricVersion"],
        "model_profile": audit["modelProfileName"],
        "model_profile_version": audit["modelProfileVersion"],
        "language": audit["language"],
        "feature_flag": audit.get("featureFlag", ""),
        "data_source_version": audit.get("dataSourceVersion", ""),
        "provider_ref": audit["provider"],
        "model_id": audit["modelId"],
    }


def call_summary(audit: dict[str, Any]) -> dict[str, Any]:
    input_tokens = audit["inputTokens"]
    output_tokens = audit["outputTokens"]
    result = {
        "coordinate": coordinate(audit),
        "usage": {
            "input_tokens": input_tokens,
            "output_tokens": output_tokens,
            "total_tokens": input_tokens + output_tokens,
        },
        "latency_ms": audit["latencyMs"],
        "finish_reason": audit["finishReason"],
        "validation_status": audit["validationStatus"],
        "attempt_count": audit["attemptCount"],
        "retry_count": audit["retryCount"],
        "retry_reasons": list(audit["retryReasons"]),
        "repair_scopes": list(audit["repairScopes"]),
    }
    if audit["stage"] == "completion":
        result["repair_used"] = audit["repairUsed"]
        result["repair_scope"] = audit["repairScope"]
    return result


def redacted_structural_meta(
    report: dict[str, Any],
    focus: dict[str, Any],
    completion_audit: dict[str, Any],
    judge_audit: dict[str, Any],
) -> dict[str, Any]:
    preparedness = report.get("preparednessLevel")
    if preparedness not in {"not_ready", "needs_practice", "basically_ready", "well_prepared"}:
        preparedness = "invalid"
    issues = report.get("issues")
    issue_count = len(issues) if isinstance(issues, list) else -1
    assessments = report.get("dimensionAssessments")
    needs_work_count = (
        sum(
            1
            for assessment in assessments
            if isinstance(assessment, dict) and assessment.get("status") == "needs_work"
        )
        if isinstance(assessments, list)
        else -1
    )
    action_types = []
    for action in report.get("nextActions", []):
        action_type = action.get("type") if isinstance(action, dict) else None
        if action_type in {"retry_current_round", "next_round", "review_evidence"}:
            action_types.append(action_type)
        else:
            action_types.append("invalid")
    digest = completion_audit.get("outputSha256")
    if not isinstance(digest, str) or re.fullmatch(r"[0-9a-f]{64}", digest) is None:
        digest = "invalid"

    def token_total(audit: dict[str, Any]) -> int:
        input_tokens = audit.get("inputTokens")
        output_tokens = audit.get("outputTokens")
        if not isinstance(input_tokens, int) or isinstance(input_tokens, bool):
            return 0
        if not isinstance(output_tokens, int) or isinstance(output_tokens, bool):
            return 0
        return max(input_tokens, 0) + max(output_tokens, 0)

    return {
        "preparedness": preparedness,
        "action_types": ",".join(action_types) if action_types else "none",
        "issue_count": issue_count,
        "needs_work_count": needs_work_count,
        "focus_count": focus.get("focus_count", -1),
        "mode": focus.get("mode", "invalid"),
        "generation_tokens": token_total(completion_audit),
        "judge_tokens": token_total(judge_audit),
        "output_digest": digest,
    }


def format_structural_meta(meta: dict[str, Any]) -> str:
    return (
        f"preparedness={meta['preparedness']} action_types={meta['action_types']} "
        f"issue_count={meta['issue_count']} needs_work_count={meta['needs_work_count']} "
        f"focus_count={meta['focus_count']} mode={meta['mode']} "
        f"generation_tokens={meta['generation_tokens']} judge_tokens={meta['judge_tokens']} "
        f"output_digest={meta['output_digest']}"
    )


def redacted_failure_usage(path: Path) -> dict[str, int]:
    empty = {"inputTokens": 0, "outputTokens": 0}
    if not path.is_file() or stat.S_IMODE(path.stat().st_mode) != 0o600:
        return empty
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return empty
    if not isinstance(payload, dict):
        return empty
    result: dict[str, int] = {}
    for key in ("inputTokens", "outputTokens"):
        value = payload.get(key)
        result[key] = value if isinstance(value, int) and not isinstance(value, bool) and value >= 0 else 0
    return result


def redacted_item_verdicts(verdict: dict[str, Any]) -> list[dict[str, Any]]:
    items = verdict.get("item_verdicts")
    if not isinstance(items, list) or not items:
        fail("judge verdict is missing item-level support evidence")
    redacted = []
    for item in items:
        if not isinstance(item, dict):
            fail("judge item verdict is invalid")
        support = item.get("support")
        kind = item.get("kind")
        if support not in {"supported", "partial"} or kind not in {"fact", "judgment", "advice"}:
            fail("judge item verdict failed support/kind gate")
        limited = item.get("evidence_limited_explicit")
        negative = item.get("used_for_negative_claim")
        if support == "partial" and (limited is not True or negative is not False):
            fail("judge partial verdict violates evidence-limit policy")
        redacted.append(
            {
                "path": item.get("path"),
                "kind": kind,
                "support": support,
                "evidence_limited_explicit": limited is True,
                "used_for_negative_claim": negative is True,
                "reason_code": f"judge_{support}_{kind}",
            }
        )
    if not {"judgment", "advice"}.issubset({item["kind"] for item in redacted}):
        fail("judge verdict does not include both report judgments and executable advice")
    return redacted


def redacted_causal_checks(verdict: dict[str, Any]) -> list[dict[str, Any]]:
    checks = verdict.get("causal_checks")
    if not isinstance(checks, list):
        fail("judge causal checks are invalid")
    redacted = []
    for check in checks:
        if not isinstance(check, dict):
            fail("judge causal check is invalid")
        if any(check.get(key) is not True for key in ("issue_supported", "focus_supported", "action_supported")):
            fail("judge found a fact-to-judgment-to-action causal mismatch")
        redacted.append(
            {
                "dimension_code": check.get("dimension_code"),
                "issue_supported": True,
                "focus_supported": True,
                "action_supported": True,
                "reason_code": "judge_closed_chain",
            }
        )
    return redacted


def score_summary(verdict: dict[str, Any]) -> tuple[dict[str, float], float]:
    scores = verdict.get("scores")
    if not isinstance(scores, list) or not scores:
        fail("judge verdict is missing rubric scores")
    result: dict[str, float] = {}
    for score in scores:
        if not isinstance(score, dict) or not isinstance(score.get("dimension"), str):
            fail("judge score entry is invalid")
        value = score.get("value")
        if not isinstance(value, (int, float)) or isinstance(value, bool) or value < 0.70:
            fail("judge dimension score is below 0.70")
        result[score["dimension"]] = float(value)
    weighted = verdict.get("weighted_score")
    if not isinstance(weighted, (int, float)) or isinstance(weighted, bool) or weighted < 0.80:
        fail("registered judge weighted score is below 0.80")
    return result, float(weighted)


def focus_shape(report: dict[str, Any]) -> dict[str, Any]:
    focus = report.get("retryFocusDimensionCodes")
    issues = report.get("issues")
    actions = report.get("nextActions")
    if not isinstance(focus, list) or not isinstance(issues, list) or not isinstance(actions, list):
        fail("live report focus/issues/actions contract is missing")
    retry_action_present = any(
        isinstance(action, dict) and action.get("type") == "retry_current_round" for action in actions
    )
    issue_codes = {issue.get("dimensionCode") for issue in issues if isinstance(issue, dict)}
    backed = all(isinstance(code, str) and code in issue_codes for code in focus)
    mode = "focused" if retry_action_present and focus else "generic" if retry_action_present else "none"
    return {
        "retry_action_present": retry_action_present,
        "focus_count": len(focus),
        "mode": mode,
        "nonempty_focus_issue_backed": backed if focus else True,
    }


def focus_failure(reason_code: str, output_digest: str, audit: dict[str, Any]) -> None:
    fail(
        f"report focus gate failed reason_code={reason_code} "
        f"output_digest={output_digest} "
        f"retry_action_present={str(audit['retry_action_present']).lower()} "
        f"focus_count={audit['focus_count']} mode={audit['mode']} "
        f"nonempty_focus_issue_backed={str(audit['nonempty_focus_issue_backed']).lower()}"
    )


def action_label_audit(report: dict[str, Any], language: str, output_digest: str) -> dict[str, Any]:
    actions = report.get("nextActions")
    if not isinstance(actions, list) or not actions:
        fail("live report actions contract is missing")
    if language == "en":
        unit, limit = "words", 24
    elif language == "zh-CN":
        unit, limit = "code_points", 64
    else:
        fail("live report action label language is unsupported")
    counts: list[int] = []
    for index, action in enumerate(actions):
        label = action.get("label") if isinstance(action, dict) else None
        if not isinstance(label, str) or not label.strip():
            fail(f"live report action label contract is missing at index={index}")
        count = len(re.findall(r"\S+", label.strip())) if language == "en" else len(label)
        if count > limit:
            reason_code = "action_label_word_limit" if language == "en" else "action_label_code_point_limit"
            fail(
                f"report action label gate failed reason_code={reason_code} "
                f"path=$.nextActions[{index}].label count={count} limit={limit} "
                f"language={language} output_digest={output_digest}"
            )
        counts.append(count)
    return {"language": language, "unit": unit, "limit": limit, "counts": counts}


def focus_audit(report: dict[str, Any], case_id: str, output_digest: str) -> dict[str, Any]:
    audit = focus_shape(report)
    dimensions = report.get("dimensionAssessments")
    if not isinstance(dimensions, list):
        fail("live report dimension contract is missing")
    action_types = [
        action.get("type")
        for action in report["nextActions"]
        if isinstance(action, dict)
    ]
    if len(action_types) > 2:
        focus_failure("too_many_actions", output_digest, audit)
    if len(action_types) != len(set(action_types)):
        focus_failure("duplicate_action_type", output_digest, audit)
    if audit["focus_count"] > 0 and not audit["nonempty_focus_issue_backed"]:
        focus_failure("focus_without_same_code_issue", output_digest, audit)
    focus = report["retryFocusDimensionCodes"]
    needs_work_codes = {
        dimension.get("code")
        for dimension in dimensions
        if isinstance(dimension, dict) and dimension.get("status") == "needs_work"
    }
    issue_codes = [
        issue.get("dimensionCode")
        for issue in report["issues"]
        if isinstance(issue, dict)
    ]
    expected_focus = sorted({code for code in issue_codes if code in needs_work_codes})
    exact_generic_exception = len(issue_codes) == 1 and issue_codes[0] in {
        "answer_depth",
        "answer_relevance",
    }
    if not audit["retry_action_present"] and focus:
        focus_failure("retry_action_required", output_digest, audit)
    if audit["retry_action_present"]:
        if exact_generic_exception and focus:
            focus_failure("generic_exception_requires_empty_focus", output_digest, audit)
        if not exact_generic_exception and (not expected_focus or focus != expected_focus):
            if len(report["issues"]) > 1 and audit["focus_count"] == 0:
                focus_failure("multi_issue_empty_focus", output_digest, audit)
            focus_failure("retry_focus_mismatch", output_digest, audit)
    if case_id == "report.generate-short-conservative" and audit != {
        "retry_action_present": True,
        "focus_count": 0,
        "mode": "generic",
        "nonempty_focus_issue_backed": True,
    }:
        focus_failure("short_generic_focus_mismatch", output_digest, audit)
    return audit


def run_attempt(
    repo_root: Path,
    evalkit: Path,
    temp_dir: Path,
    run_id: str,
    case_id: str,
    case_type: str,
    critical: bool,
    repetition: int,
    context_digest: str,
) -> tuple[dict[str, Any], dict[str, Any]]:
    prefix = f"{case_type}-r{repetition}"
    completion_audit_path = temp_dir / f"{prefix}-completion-audit.json"
    judge_audit_path = temp_dir / f"{prefix}-judge-audit.json"

    stdout = run_evalkit(
        [str(evalkit), "complete", "--case", case_id, "--live", "--audit-out", str(completion_audit_path)],
        repo_root=repo_root,
        stage="completion",
        timeout=180,
    )
    if not stdout.endswith(b"\n"):
        fail("evalkit completion stdout did not terminate cleanly")
    raw_output = stdout[:-1]
    completion_audit = load_audit(completion_audit_path, "completion", case_id, critical)
    if hashlib.sha256(raw_output).hexdigest() != completion_audit["outputSha256"]:
        fail("completion audit digest does not match the in-memory candidate output")
    try:
        report_output = json.loads(raw_output)
    except json.JSONDecodeError:
        fail("live report output is not strict JSON")
    if not isinstance(report_output, dict):
        fail("live report output is not an object")

    try:
        label_audit = action_label_audit(
            report_output,
            CASE_LANGUAGES[case_id],
            completion_audit["outputSha256"],
        )
        focus = focus_audit(report_output, case_id, completion_audit["outputSha256"])
    except LiveRunError as exc:
        shape = focus_shape(report_output)
        structural_meta = redacted_structural_meta(
            report_output,
            shape,
            completion_audit,
            {},
        )
        fail(f"{exc}; {format_structural_meta(structural_meta)}")

    try:
        grade_stdout = run_evalkit(
            [str(evalkit), "grade", "--case", case_id, "--live", "--audit-out", str(judge_audit_path)],
            repo_root=repo_root,
            stage="judge",
            output_digest=completion_audit["outputSha256"],
            input_bytes=raw_output,
            timeout=180,
        )
    except LiveRunError as exc:
        structural_meta = redacted_structural_meta(
            report_output,
            focus,
            completion_audit,
            redacted_failure_usage(judge_audit_path),
        )
        fail(f"{exc}; {format_structural_meta(structural_meta)}")
    try:
        verdict = json.loads(grade_stdout)
    except json.JSONDecodeError:
        fail("evalkit grade stdout is not valid JSON")
    if not isinstance(verdict, dict) or verdict.get("pass") is not True or verdict.get("caseId") != case_id:
        fail("registered context-aware judge rejected the live report")
    judge_audit = load_audit(judge_audit_path, "judge", case_id, critical)

    scores, weighted = score_summary(verdict)
    zero_tolerance = verdict.get("zero_tolerance_violations")
    if zero_tolerance != []:
        fail("registered judge reported a zero-tolerance violation")
    if verdict.get("critical_safety_pass") is not True:
        fail("registered judge reported a critical safety failure")

    generation_call_id = hashlib.sha256(
        f"{run_id}:{case_id}:{repetition}:completion:{completion_audit['outputSha256']}".encode()
    ).hexdigest()
    judge_call_id = hashlib.sha256(
        f"{run_id}:{case_id}:{repetition}:judge:{judge_audit['outputSha256']}".encode()
    ).hexdigest()
    attempt = {
        "case_id": case_id,
        "case_type": case_type,
        "critical": critical,
        "repetition": repetition,
        "context_digest": context_digest,
        "output_digest": completion_audit["outputSha256"],
        "judge_digest": judge_audit["outputSha256"],
        "generation_call_id": generation_call_id,
        "judge_call_id": judge_call_id,
        "generation": call_summary(completion_audit),
        "judge": {
            **call_summary(judge_audit),
            "scores": scores,
            "weighted_score": weighted,
            "item_verdicts": redacted_item_verdicts(verdict),
            "causal_checks": redacted_causal_checks(verdict),
            "zero_tolerance_violations": [],
            "critical_safety_pass": True,
        },
        "action_label_audit": label_audit,
        "focus_audit": focus,
        "raw_persisted": False,
    }
    return attempt, report_output


def write_manifest(output_dir: Path, manifest: dict[str, Any]) -> Path:
    output_dir.mkdir(parents=True, exist_ok=True, mode=0o700)
    target = output_dir / "reliability-manifest.json"
    temporary = output_dir / ".reliability-manifest.json.tmp"
    data = json.dumps(manifest, ensure_ascii=False, indent=2, sort_keys=True).encode("utf-8") + b"\n"
    fd = os.open(temporary, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    try:
        os.write(fd, data)
    finally:
        os.close(fd)
    os.chmod(temporary, 0o600)
    os.replace(temporary, target)
    os.chmod(target, 0o600)
    return target


def write_secure_json(path: Path, payload: dict[str, Any]) -> None:
    data = json.dumps(payload, ensure_ascii=False, indent=2, sort_keys=True).encode("utf-8") + b"\n"
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    try:
        os.write(fd, data)
    finally:
        os.close(fd)
    os.chmod(path, 0o600)


def await_independent_agent_audit(path: Path, timeout_seconds: int) -> None:
    deadline = time.monotonic() + timeout_seconds
    print("P0_100_AGENT_REVIEW_REQUIRED cases=5 source=independent_agent_review", flush=True)
    while time.monotonic() < deadline:
        if path.is_file() and path.stat().st_size > 0:
            if path.stat().st_mode & 0o777 != 0o600:
                fail("independent Agent audit must be written with mode 0600")
            print("P0_100_AGENT_REVIEW_RECEIVED", flush=True)
            return
        time.sleep(1)
    fail("independent Agent review timed out without a redacted audit")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", type=Path, required=True)
    parser.add_argument("--evalkit", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--agent-audit", type=Path, required=True)
    parser.add_argument("--agent-review-timeout", type=int, default=1800)
    parser.add_argument("--run-id", required=True)
    args = parser.parse_args()
    try:
        repo_root = args.repo_root.resolve()
        evalkit = args.evalkit.resolve()
        if not evalkit.is_file() or not os.access(evalkit, os.X_OK):
            fail("evalkit executable is missing")
        contexts = load_case_contexts(repo_root)
        attempts: list[dict[str, Any]] = []
        review_cases: list[dict[str, Any]] = []
        if args.agent_audit.exists():
            fail("independent Agent audit already exists before the current live run")
        if args.agent_review_timeout < 60 or args.agent_review_timeout > 3600:
            fail("agent review timeout must be between 60 and 3600 seconds")
        with tempfile.TemporaryDirectory(prefix=f"easyinterview-p0-100-review-{args.run_id}-") as temporary:
            temp_dir = Path(temporary)
            os.chmod(temp_dir, 0o700)
            for case_id, case_type, critical, repetitions, _language in CASES:
                for repetition in range(1, repetitions + 1):
                    attempt, report_output = run_attempt(
                        repo_root,
                        evalkit,
                        temp_dir,
                        args.run_id,
                        case_id,
                        case_type,
                        critical,
                        repetition,
                        contexts[case_id]["digest"],
                    )
                    attempts.append(attempt)
                    if repetition == 1:
                        review_cases.append(
                            {
                                "language": _language,
                                "context_digest": contexts[case_id]["digest"],
                                "output_digest": attempt["output_digest"],
                                "context": contexts[case_id]["context"],
                                "transcript": contexts[case_id]["transcript"],
                                "output": report_output,
                                "generation": attempt["generation"],
                            }
                        )
                    generation_tokens = attempt["generation"]["usage"]["total_tokens"]
                    judge_tokens = attempt["judge"]["usage"]["total_tokens"]
                    print(
                        f"P0_100_LIVE_ATTEMPT case={case_type} repetition={repetition}/{repetitions} "
                        f"generation_tokens={generation_tokens} judge_tokens={judge_tokens}",
                        flush=True,
                    )

            review_packet = temp_dir / "agent-review-packet.json"
            write_secure_json(
                review_packet,
                build_blind_review_packet(args.run_id, review_cases),
            )
            await_independent_agent_audit(args.agent_audit, args.agent_review_timeout)
            review_packet.unlink(missing_ok=True)

        manifest = {
            "schema_version": "p0-100-reliability-manifest.v2",
            "scenario_id": "E2E.P0.100",
            "run_id": args.run_id,
            "trust_boundary": "review.BuildReportPromptMessages",
            "provider_mode": "real",
            "thresholds": {"minimum_dimension": 0.70, "minimum_weighted": 0.80, "critical_repetitions": 3},
            "attempts": attempts,
            "privacy": {
                "redacted": True,
                "raw_context_written": False,
                "raw_output_written": False,
                "cookie_written": False,
                "secret_written": False,
            },
        }
        target = write_manifest(args.output_dir, manifest)
        print(f"P0_100_LIVE_MANIFEST_READY attempts={len(attempts)} path={target}")
        return 0
    except LiveRunError as exc:
        print(f"P0.100 live reliability failed: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
