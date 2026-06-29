#!/usr/bin/env python3
"""Sync `ui-design/src/data.jsx` into the `scenarios.prototype-baseline`
section of every OpenAPI fixture under `openapi/fixtures/<tag>/<operationId>.json`.

Phase 2 owner per `002-fixtures-and-mock-source` plan §3 / spec §4.7. Reads
the prototype data via Node, applies the per-operation mapping table from
`openapi/fixtures/PROTOTYPE_MAPPING.md` (encoded below), normalizes ids and
times into UUIDv7 / RFC3339 UTC, then writes back deterministically. The tool
is idempotent — running twice in a row leaves `git diff --exit-code` clean.

The tool **fails fast** if data.jsx is missing a section the mapping table
declares as required (`Mapping gap: ...`), and re-runs Phase 1 schema
validation as a final gate.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import subprocess
import sys
from collections import OrderedDict
from pathlib import Path
from typing import Any, Callable, Iterable, List

REPO_ROOT_DEFAULT = Path(__file__).resolve().parents[2]


# ---------- deterministic UUIDv7 / time normalization ------------------------

def uuidv7_for(token: str) -> str:
    """Hash a stable prototype identifier into a UUIDv7-shaped string."""
    digest = hashlib.sha256(token.encode("utf-8")).hexdigest()
    raw = list("01918fa0" + digest[8:12] + "7" + digest[13:16] + "8" + digest[17:32])
    out = "".join(raw)
    return f"{out[:8]}-{out[8:12]}-{out[12:16]}-{out[16:20]}-{out[20:32]}"


NOW = "2026-04-28T13:45:12Z"
EARLIER = "2026-04-28T12:00:00Z"
EARLIEST = "2026-04-22T09:30:00Z"
REQUEST_ID = "req_2026-04-28T13-45-12-prototype"
PROTOTYPE_MODEL_PROFILE_ID = "model-profile:prototype-baseline.default"


# Provenance template reused across AI schemas.
def _prov(prompt: str, rubric: str = "not_applicable") -> OrderedDict:
    return OrderedDict([
        ("promptVersion", prompt),
        ("rubricVersion", rubric),
        ("modelId", PROTOTYPE_MODEL_PROFILE_ID),
        ("language", "zh-CN"),
        ("featureFlag", "none"),
        ("dataSourceVersion", "prototype-baseline.v1"),
    ])


# ---------- normalization tables --------------------------------------------

COMPANY_TRANSLATION = {
    "星环科技": "Acme",
    "Acme": "Acme",
    "Lumen Labs": "Lumen Labs",
    "Lumen": "Lumen",
    "云栖集团": "Helios Group",
}

TARGET_STATUS_TRANSLATION = {
    "面试中": "interviewing",
    "准备中": "preparing",
    "草稿": "draft",
    "已完成": "offer",
    "已结束": "archived",
}

LANGUAGE_TRANSLATION = {
    "中文": "zh-CN",
    "英文": "en",
    "zh-CN": "zh-CN",
    "en": "en",
}

SOURCE_TYPE_TRANSLATION = {
    "粘贴 JD": "manual_text",
    "岗位链接": "url",
    "招聘方邮件": "manual_text",
    "url": "url",
    "manual_text": "manual_text",
    "file": "file",
    "manual_form": "manual_form",
}

READINESS_TIER = {
    0: "not_ready",
    1: "needs_practice",
    2: "basically_ready",
    3: "well_prepared",
}

def _translate_company(value: str | None) -> str:
    if value is None:
        return "Acme"
    if value in COMPANY_TRANSLATION:
        return COMPANY_TRANSLATION[value]
    return "Acme"


def _mask_email(email: str) -> str:
    local, _, domain = email.partition("@")
    if not local or not domain:
        return "ali***@example.com"
    visible = local[:3] if len(local) >= 3 else local
    return f"{visible}***@example.com"


def _slug(value: str) -> str:
    safe = "".join(c if c.isalnum() else "_" for c in value).strip("_").lower()
    return safe or "item"


# ---------- data loader ------------------------------------------------------

def load_prototype_data(data_file: Path) -> dict:
    """Run Node to evaluate data.jsx (which assigns `window.EI_DATA = {...}`)
    and emit JSON. Node is the only realistic way to parse JS object literals
    that include trailing commas / Chinese strings without bringing in heavy
    Python deps."""
    if not data_file.is_file():
        raise FileNotFoundError(f"data.jsx not found at {data_file}")
    runner = (
        "const fs = require('fs');"
        "const vm = require('vm');"
        "const src = fs.readFileSync(process.argv[1], 'utf8');"
        "const ctx = { window: {} };"
        "vm.createContext(ctx);"
        "vm.runInContext(src, ctx);"
        "process.stdout.write(JSON.stringify(ctx.window.EI_DATA));"
    )
    out = subprocess.run(
        ["node", "-e", runner, str(data_file)],
        capture_output=True, text=True, check=True,
    )
    return json.loads(out.stdout)


# ---------- per-operation mappers --------------------------------------------

REQUIRED_SECTIONS: dict[str, tuple[str, ...]] = {
    "getMe": ("user",),
    "listTargetJobs": ("targetJobs",),
    "getTargetJob": ("targetJobs", "jdSample"),
    "getPracticeSession": ("targetJobs", "questions", "sessionTranscript"),
    "getFeedbackReport": ("report",),
}

OP_TAGS = {
    "getMe": "Auth",
    "listTargetJobs": "TargetJobs",
    "getTargetJob": "TargetJobs",
    "getPracticeSession": "PracticeSessions",
    "getFeedbackReport": "Reports",
}


def map_get_me(data: dict) -> OrderedDict:
    user = data["user"]
    body = OrderedDict([
        ("id", uuidv7_for(f"user:{user.get('email', 'fallback')}")),
        ("emailMasked", _mask_email(user.get("email", "alice@example.com"))),
        ("displayName", "Alice Example"),  # spec §4.7: avoid real names; use placeholder
        ("uiLanguage", LANGUAGE_TRANSLATION.get(user.get("locale", "zh-CN"), "zh-CN")),
        ("preferredPracticeLanguage", LANGUAGE_TRANSLATION.get(user.get("locale", "zh-CN"), "zh-CN")),
        ("profileCompletionRequired", False),
    ])
    return _wrap_response(200, body)


def _build_target_job(raw: dict, jd: dict | None = None) -> OrderedDict:
    target_id = uuidv7_for(f"targetJob:{raw['id']}")
    base = OrderedDict([
        ("id", target_id),
        ("status", TARGET_STATUS_TRANSLATION.get(raw.get("status"), "preparing")),
        ("analysisStatus", "ready"),
        ("title", raw.get("title", "Senior Engineer")),
        ("companyName", _translate_company(raw.get("company"))),
        ("locationText", raw.get("location")),
        ("targetLanguage", LANGUAGE_TRANSLATION.get(raw.get("language", "zh-CN"), "zh-CN")),
        ("sourceType", SOURCE_TYPE_TRANSLATION.get(raw.get("source"), "manual_text")),
        ("sourceUrl", None),
    ])
    if jd is not None:
        themes = list(jd.get("hidden", [])) or ["Cross-team alignment"]
        hypotheses = []
        for r in jd.get("rounds", []):
            label = f"{r.get('name', 'Round')}: {r.get('focus', '')}"
            hypotheses.append(label)
        base["summary"] = OrderedDict([
            ("coreThemes", themes),
            ("interviewHypotheses", hypotheses),
            ("provenance", _prov("target_job_summary.v3")),
        ])
        requirements = []
        for i, label in enumerate(jd.get("mustHave", [])):
            requirements.append(OrderedDict([
                ("id", uuidv7_for(f"req:must:{raw['id']}:{i}")),
                ("kind", "must_have"),
                ("label", label),
                ("evidenceLevel", "explicit"),
            ]))
        for i, label in enumerate(jd.get("nice", [])):
            requirements.append(OrderedDict([
                ("id", uuidv7_for(f"req:nice:{raw['id']}:{i}")),
                ("kind", "nice_to_have"),
                ("label", label),
                ("evidenceLevel", "implicit"),
            ]))
        base["requirements"] = requirements
        hits = list(raw.get("hits", []))
        gaps = list(raw.get("gaps", []))
        base["fitSummary"] = OrderedDict([
            ("strengths", hits or ["Cross-team coordination"]),
            ("gaps", gaps or ["Quantified impact stories"]),
            ("riskSignals", list(jd.get("hidden", []))[-1:] or ["Hiring manager pushes for measurable outcomes"]),
            ("provenance", _prov("target_job_fit.v2")),
        ])
        base["latestReportId"] = uuidv7_for(f"report:tj:{raw['id']}")
    else:
        base["summary"] = None
        base["requirements"] = []
        base["fitSummary"] = None
        base["latestReportId"] = None
    base["openQuestionIssueCount"] = int(raw.get("mistakes") or 0)
    base["createdAt"] = EARLIEST
    base["updatedAt"] = EARLIER
    return base


def map_list_target_jobs(data: dict) -> OrderedDict:
    items = [_build_target_job(raw) for raw in data["targetJobs"]]
    body = OrderedDict([
        ("items", items),
        ("pageInfo", OrderedDict([
            ("nextCursor", None),
            ("pageSize", 20),
            ("hasMore", False),
        ])),
    ])
    return _wrap_response(200, body)


def map_get_target_job(data: dict) -> OrderedDict:
    raw = data["targetJobs"][0]
    body = _build_target_job(raw, jd=data.get("jdSample"))
    return _wrap_response(200, body)


def map_get_practice_session(data: dict) -> OrderedDict:
    target = data["targetJobs"][0]
    questions = data.get("questions", [])
    transcript = data.get("sessionTranscript", [])
    asked_count = sum(1 for t in transcript if t.get("role") == "ai" and t.get("qId"))
    current_q = next((q for q in questions if q.get("id") == "q1"), questions[0] if questions else None)
    plan_id = uuidv7_for("plan:prototype:tj-1")
    target_id = uuidv7_for(f"targetJob:{target['id']}")
    session_id = uuidv7_for("session:prototype:tj-1")
    current_turn = None
    if current_q is not None:
        current_turn = OrderedDict([
            ("id", uuidv7_for(f"turn:prototype:{current_q['id']}")),
            ("turnIndex", max(asked_count, 1)),
            ("questionText", current_q.get("prompt", "")),
            ("questionIntent", "behavioral.self_intro"),
            ("status", "asked"),
            ("askedAt", NOW),
        ])
    body = OrderedDict([
        ("id", session_id),
        ("planId", plan_id),
        ("targetJobId", target_id),
        ("status", "running"),
        ("language", LANGUAGE_TRANSLATION.get(target.get("language", "zh-CN"), "zh-CN")),
        ("hintsEnabled", True),
        ("turnCount", max(asked_count, 1)),
        ("currentTurn", current_turn),
        ("createdAt", EARLIER),
        ("updatedAt", NOW),
    ])
    return _wrap_response(200, body)


def map_get_feedback_report(data: dict) -> OrderedDict:
    report = data["report"]
    target = data["targetJobs"][0] if data.get("targetJobs") else {"id": "tj-1"}
    target_id = uuidv7_for(f"targetJob:{target['id']}")
    session_id = uuidv7_for("session:prototype:tj-1")
    report_id = uuidv7_for(f"report:tj:{target['id']}")

    highlights = []
    for h in report.get("highlights", []):
        highlights.append(OrderedDict([
            ("dimension", h.get("title", "ownership")),
            ("evidence", h.get("body", "")),
            ("confidence", "high"),
        ]))
    issues = []
    for i in report.get("issues", []):
        confidence = "high" if i.get("severity") == "high" else "medium" if i.get("severity") == "medium" else "low"
        issues.append(OrderedDict([
            ("dimension", i.get("title", "communication")),
            ("evidence", i.get("body", "")),
            ("confidence", confidence),
        ]))
    next_actions = []
    for action in report.get("nextPractice", []):
        next_actions.append(OrderedDict([
            ("type", "retry_current_round"),
            ("label", action),
        ]))
    question_assessments = []
    dims = report.get("dimensions", [])
    dim_results = OrderedDict()
    for d in dims:
        status_map = {"达标": "meets_bar", "强项": "strong", "待加强": "needs_work"}
        dim_results[_slug(d.get("name", "dim"))] = OrderedDict([
            ("status", status_map.get(d.get("state"), "meets_bar")),
            ("confidence", "high" if d.get("confidence") == "高" else "medium"),
        ])
    for pq in report.get("perQuestion", []):
        included = pq.get("state") == "待加强"
        question_assessments.append(OrderedDict([
            ("turnId", uuidv7_for(f"turn:prototype:{pq.get('qId','q?')}")),
            ("questionIntent", _slug(pq.get("topic", "question"))),
            ("dimensionResults", dim_results),
            ("reviewStatus", "queued_for_retry" if included else "resolved"),
            ("includedInRetryPlan", included),
        ]))
    retry_focus_turn_ids = [
        qa["turnId"] for qa in question_assessments if qa.get("includedInRetryPlan")
    ][:3]
    preparedness = READINESS_TIER.get(report.get("readiness", 2), "basically_ready")

    body = OrderedDict([
        ("id", report_id),
        ("sessionId", session_id),
        ("targetJobId", target_id),
        ("status", "ready"),
        ("preparednessLevel", preparedness),
        ("highlights", highlights),
        ("issues", issues),
        ("nextActions", next_actions),
        ("questionAssessments", question_assessments),
        ("retryFocusTurnIds", retry_focus_turn_ids),
        ("provenance", _prov("feedback_report.v3", rubric="feedback_report.rubric.v2")),
        ("createdAt", EARLIER),
        ("updatedAt", NOW),
    ])
    return _wrap_response(200, body)

MAPPERS: dict[str, Callable[[dict], OrderedDict]] = {
    "getMe": map_get_me,
    "listTargetJobs": map_list_target_jobs,
    "getTargetJob": map_get_target_job,
    "getPracticeSession": map_get_practice_session,
    "getFeedbackReport": map_get_feedback_report,
}


def _wrap_response(status: int, body: Any) -> OrderedDict:
    return OrderedDict([
        ("response", OrderedDict([
            ("status", status),
            ("headers", OrderedDict([("X-Request-ID", REQUEST_ID)])),
            ("body", body),
        ])),
    ])


# ---------- write back -------------------------------------------------------

def write_prototype_baseline(repo_root: Path, opid: str, scenario: OrderedDict) -> None:
    tag = OP_TAGS[opid]
    fixture_path = repo_root / "openapi" / "fixtures" / tag / f"{opid}.json"
    with fixture_path.open("r", encoding="utf-8") as f:
        data = json.load(f, object_pairs_hook=OrderedDict)
    scenarios = data.get("scenarios") or OrderedDict()
    # Preserve `default` first; then `prototype-baseline`; drop and re-add to
    # keep deterministic ordering across re-runs.
    new_scenarios = OrderedDict()
    if "default" in scenarios:
        new_scenarios["default"] = scenarios["default"]
    new_scenarios["prototype-baseline"] = scenario
    for k, v in scenarios.items():
        if k not in new_scenarios:
            new_scenarios[k] = v
    data["scenarios"] = new_scenarios
    with fixture_path.open("w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        f.write("\n")


# ---------- CLI --------------------------------------------------------------

def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=REPO_ROOT_DEFAULT)
    parser.add_argument(
        "--data-file",
        type=Path,
        default=None,
        help="Override path to data.jsx (defaults to <repo>/ui-design/src/data.jsx).",
    )
    args = parser.parse_args(list(argv))
    repo_root = args.repo_root.resolve()
    data_file = args.data_file or repo_root / "ui-design" / "src" / "data.jsx"

    raw = load_prototype_data(data_file)

    # Mapping-gap fail-fast.
    gaps: List[str] = []
    for opid, sections in REQUIRED_SECTIONS.items():
        for section in sections:
            if section not in raw:
                gaps.append(
                    f"Mapping gap: {opid} requires data.{section} (declared in PROTOTYPE_MAPPING.md)"
                )
    if gaps:
        for g in gaps:
            print(g, file=sys.stderr)
        return 1

    for opid, mapper in MAPPERS.items():
        scenario = mapper(raw)
        write_prototype_baseline(repo_root, opid, scenario)

    # Phase 1 schema gate runs as the final wall.
    sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "lint"))
    import validate_fixtures  # noqa: E402  (path manipulated above)

    errors = validate_fixtures.validate(repo_root)
    if errors:
        for err in errors:
            print(f"sync-fixtures-from-prototype: {err}", file=sys.stderr)
        return 1
    print(
        f"sync-fixtures-from-prototype: OK — {len(MAPPERS)} fixtures populated under "
        f"{repo_root / 'openapi' / 'fixtures'}"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
