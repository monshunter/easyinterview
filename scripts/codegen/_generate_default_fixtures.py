#!/usr/bin/env python3
"""One-shot bootstrap to populate openapi/fixtures/*.json with rich default
content satisfying spec §3.1 / §4.6 / §4.7 invariants. After this runs once
the files are hand-maintained or refreshed via `make sync-fixtures-from-prototype`.

This script intentionally lives under scripts/codegen/ but is NOT wired into a
Make target. It exists as a reproducible record of the Phase 1.2 bootstrap.
"""

from __future__ import annotations

import json
from collections import OrderedDict
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
FIX = REPO_ROOT / "openapi" / "fixtures"

# Deterministic UUIDv7 emitter: 0x01918fa0 prefix is a v7-shaped (millis since
# 1970-01-01) value approximating 2026-04-28; bumping `seq` fans out unique ids.
def u7(seq: int) -> str:
    s = f"{seq:020x}"  # 20 hex chars to fill the suffix (20 + 12 prefix = 32)
    head = "01918fa0"  # 8
    rand_a = s[:4]     # 4 → bytes 9..12
    seg7 = "7" + s[4:7]   # 4: 7XXX → bytes 13..16
    seg8 = "8" + s[7:10]  # 4: 8XXX → bytes 17..20
    tail = s[10:22] if len(s) >= 22 else (s[10:] + "0" * (22 - len(s)))
    raw = head + rand_a + seg7 + seg8 + tail
    return f"{raw[:8]}-{raw[8:12]}-{raw[12:16]}-{raw[16:20]}-{raw[20:32]}"

# Stable id pool (bumping causes diff churn, so reuse seqs explicitly).
USER_ID                    = u7(0x01)
RESUME_ASSET_ID            = u7(0x10)
RESUME_FILE_OBJECT_ID      = u7(0x11)
RESUME_TAILOR_FILE_ID      = u7(0x12)
TARGET_JOB_ID_1            = u7(0x20)
TARGET_JOB_ID_2            = u7(0x21)
TARGET_JOB_ID_3            = u7(0x22)
EXPERIENCE_CARD_ID_1       = u7(0x30)
EXPERIENCE_CARD_ID_2       = u7(0x31)
PRACTICE_PLAN_ID_1         = u7(0x40)
PRACTICE_SESSION_ID_1      = u7(0x50)
PRACTICE_SESSION_ID_2      = u7(0x51)
PRACTICE_TURN_ID_1         = u7(0x60)
PRACTICE_TURN_ID_2         = u7(0x61)
REPORT_ID_1                = u7(0x70)
REPORT_ID_2                = u7(0x71)
RESUME_TAILOR_RUN_ID       = u7(0x90)
DEBRIEF_ID                 = u7(0xA0)
JOB_ID_TARGET_IMPORT       = u7(0xB0)
JOB_ID_RESUME_PARSE        = u7(0xB1)
JOB_ID_REPORT_GENERATE     = u7(0xB2)
JOB_ID_RESUME_TAILOR       = u7(0xB3)
JOB_ID_DEBRIEF_GENERATE    = u7(0xB4)
JOB_ID_PRIVACY_DELETE      = u7(0xB5)
JOB_ID_PRIVACY_ME_DELETE   = u7(0xB6)
PRIVACY_REQUEST_ID_DEL     = u7(0xC0)
PRIVACY_REQUEST_ID_GET     = u7(0xC1)
PRIVACY_REQUEST_ID_ME_DEL  = u7(0xC2)
TARGET_REQ_ID_1            = u7(0xD0)
TARGET_REQ_ID_2            = u7(0xD1)
TARGET_REQ_ID_3            = u7(0xD2)
CLIENT_EVENT_ID_1          = u7(0xE0)
REQUEST_ID                 = "req_2026-04-28T13-45-12-abcdef"
FIXTURE_MODEL_PROFILE_ID   = "model-profile:contract.default"

NOW       = "2026-04-28T13:45:12Z"
EARLIER   = "2026-04-28T12:00:00Z"
EARLIEST  = "2026-04-22T09:30:00Z"

# Provenance template (spec §4.6: 6 required fields, all non-empty).
def prov(prompt: str, rubric: str, model: str, lang: str, flag: str, dsv: str) -> dict:
    return OrderedDict([
        ("promptVersion", prompt),
        ("rubricVersion", rubric),
        ("modelId", model),
        ("language", lang),
        ("featureFlag", flag),
        ("dataSourceVersion", dsv),
    ])

PROV_TARGET_SUMMARY = prov(
    "target_job_summary.v3", "not_applicable", FIXTURE_MODEL_PROFILE_ID,
    "zh-CN", "none", "target_job.v17",
)
PROV_TARGET_FIT = prov(
    "target_job_fit.v2", "not_applicable", FIXTURE_MODEL_PROFILE_ID,
    "zh-CN", "fit_summary_v2", "target_job.v17",
)
PROV_ASSISTANT_ACTION = prov(
    "practice_session_assistant.v5", "not_applicable",
    FIXTURE_MODEL_PROFILE_ID, "zh-CN", "follow_up_v3",
    "practice_session.v9",
)
PROV_FEEDBACK_REPORT = prov(
    "feedback_report.v3", "feedback_report.rubric.v2",
    FIXTURE_MODEL_PROFILE_ID, "zh-CN", "none",
    "practice_session.v9",
)
PROV_RESUME_TAILOR = prov(
    "resume_tailor.v2", "not_applicable",
    FIXTURE_MODEL_PROFILE_ID, "zh-CN", "none",
    "target_job.v17",
)
PROV_DEBRIEF = prov(
    "debrief_generate.v1", "not_applicable",
    FIXTURE_MODEL_PROFILE_ID, "zh-CN", "none",
    "debrief.v4",
)


# Generic fixture builder ------------------------------------------------------

def fixture(opid: str, response_body, *, status: int = 200,
            request_body=None, response_headers=None) -> dict:
    out = OrderedDict()
    out["operationId"] = opid
    default = OrderedDict()
    if request_body is not None:
        default["request"] = OrderedDict([
            ("headers", OrderedDict()),
            ("body", request_body),
        ])
    response = OrderedDict()
    response["status"] = status
    response["headers"] = response_headers or OrderedDict([
        ("X-Request-ID", REQUEST_ID),
    ])
    if status != 204:
        response["body"] = response_body
    default["response"] = response
    out["scenarios"] = OrderedDict([("default", default)])
    return out


# ---------- Auth tag ---------------------------------------------------------

FIXTURES = {}

FIXTURES[("Auth", "getMe")] = fixture("getMe", OrderedDict([
    ("id", USER_ID),
    ("emailMasked", "ali***@example.com"),
    ("displayName", "Alice Example"),
    ("uiLanguage", "zh-CN"),
    ("preferredPracticeLanguage", "zh-CN"),
]))

FIXTURES[("Auth", "startAuthEmailChallenge")] = fixture(
    "startAuthEmailChallenge",
    None,
    status=202,
    request_body=OrderedDict([
        ("email", "alice@example.com"),
        ("returnTo", "/workspace"),
    ]),
)

FIXTURES[("Auth", "verifyAuthEmailChallenge")] = fixture(
    "verifyAuthEmailChallenge",
    OrderedDict([
        ("userId", USER_ID),
        ("sessionExpiresAt", "2026-05-28T13:45:12Z"),
    ]),
    response_headers=OrderedDict([
        ("X-Request-ID", REQUEST_ID),
        ("Set-Cookie", "ei_session=opaque-session-token; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=2592000"),
    ]),
)

FIXTURES[("Auth", "logout")] = fixture("logout", None, status=204)

FIXTURES[("Auth", "deleteMe")] = OrderedDict([
    ("operationId", "deleteMe"),
    ("scenarios", OrderedDict([
        ("default", OrderedDict([
            ("request", OrderedDict([
                ("headers", OrderedDict([
                    ("Idempotency-Key", "idem-delete-me-2026-04-29"),
                ])),
            ])),
            ("response", OrderedDict([
                ("status", 202),
                ("headers", OrderedDict([
                    ("X-Request-ID", REQUEST_ID),
                ])),
                ("body", OrderedDict([
                    ("privacyRequestId", PRIVACY_REQUEST_ID_ME_DEL),
                    ("job", OrderedDict([
                        ("id", JOB_ID_PRIVACY_ME_DELETE),
                        ("jobType", "privacy_delete"),
                        ("status", "queued"),
                        ("resourceType", "privacy_request"),
                        ("resourceId", PRIVACY_REQUEST_ID_ME_DEL),
                        ("errorCode", None),
                        ("createdAt", NOW),
                        ("updatedAt", NOW),
                    ])),
                ])),
            ])),
        ])),
    ])),
])

FIXTURES[("Auth", "getRuntimeConfig")] = fixture(
    "getRuntimeConfig",
    OrderedDict([
        ("defaultUiLanguage", "zh-CN"),
        ("featureFlags", OrderedDict([
            ("ai_assistant_actions", True),
            ("debrief_v2", True),
            ("resume_tailor_bullets", False),
        ])),
        ("appVersion", "1.0.0+dev.0428"),
        ("analyticsEnabled", True),
        ("postHogPublicKey", "phc_local_dev_placeholder"),
    ]),
)


# ---------- Uploads tag ------------------------------------------------------

FIXTURES[("Uploads", "createUploadPresign")] = fixture(
    "createUploadPresign",
    OrderedDict([
        ("fileObjectId", RESUME_FILE_OBJECT_ID),
        ("uploadUrl", "https://uploads.acme.example/presigned/upload?token=abc"),
        ("method", "PUT"),
        ("headers", OrderedDict([
            ("Content-Type", "application/pdf"),
            ("x-amz-server-side-encryption", "AES256"),
        ])),
        ("expiresAt", "2026-04-28T14:00:00Z"),
    ]),
    status=201,
    request_body=OrderedDict([
        ("purpose", "resume"),
        ("fileName", "alice-resume-2026.pdf"),
        ("contentType", "application/pdf"),
        ("byteSize", 248192),
    ]),
)


# ---------- Profile tag ------------------------------------------------------

CANDIDATE_PROFILE = OrderedDict([
    ("headline", "Senior frontend engineer focused on growth-stage SaaS"),
    ("yearsOfExperience", 5),
    ("currentRole", "Senior Frontend Engineer at Acme"),
    ("preferredPracticeLanguage", "zh-CN"),
    ("uiLanguage", "zh-CN"),
    ("region", "CN-SH"),
])

FIXTURES[("Profile", "getMyProfile")] = fixture("getMyProfile", CANDIDATE_PROFILE)

FIXTURES[("Profile", "updateMyProfile")] = fixture(
    "updateMyProfile",
    CANDIDATE_PROFILE,
    request_body=OrderedDict([
        ("headline", "Senior frontend engineer focused on growth-stage SaaS"),
        ("region", "CN-SH"),
    ]),
)

EXP_CARD_1 = OrderedDict([
    ("id", EXPERIENCE_CARD_ID_1),
    ("title", "Drove design-system migration across 12 product teams"),
    ("companyName", "Acme"),
    ("situation", "Acme had three competing design systems causing inconsistent UX across web."),
    ("task", "Lead a unified design-system migration without freezing product roadmaps."),
    ("action", "Authored RFC, established a 6-week parallel rollout, and paired with each team."),
    ("result", "Reduced UI defect rate by 38%; 12/12 teams shipped on the unified system."),
    ("skills", ["technical-leadership", "stakeholder-management", "frontend-architecture"]),
    ("language", "zh-CN"),
    ("createdAt", EARLIEST),
    ("updatedAt", EARLIER),
])

EXP_CARD_2 = OrderedDict([
    ("id", EXPERIENCE_CARD_ID_2),
    ("title", "Cut p95 first-paint latency from 3.2s to 1.1s"),
    ("companyName", "Acme"),
    ("situation", "Marketing pages had >3s p95 first-paint after a CMS migration."),
    ("task", "Restore first-paint within 1.5s without losing personalization."),
    ("action", "Migrated to streaming SSR + edge cache; rewrote the personalization fetch path."),
    ("result", "p95 first-paint = 1.1s; conversion lifted +9% on the landing funnel."),
    ("skills", ["performance", "ssr", "edge-computing"]),
    ("language", "zh-CN"),
    ("createdAt", EARLIEST),
    ("updatedAt", EARLIER),
])

FIXTURES[("Profile", "listExperienceCards")] = fixture(
    "listExperienceCards",
    OrderedDict([
        ("items", [EXP_CARD_1, EXP_CARD_2]),
        ("pageInfo", OrderedDict([
            ("nextCursor", None),
            ("pageSize", 20),
            ("hasMore", False),
        ])),
    ]),
)

FIXTURES[("Profile", "createExperienceCard")] = fixture(
    "createExperienceCard",
    EXP_CARD_1,
    status=201,
    request_body=OrderedDict([
        ("title", EXP_CARD_1["title"]),
        ("companyName", EXP_CARD_1["companyName"]),
        ("situation", EXP_CARD_1["situation"]),
        ("task", EXP_CARD_1["task"]),
        ("action", EXP_CARD_1["action"]),
        ("result", EXP_CARD_1["result"]),
        ("skills", EXP_CARD_1["skills"]),
        ("language", EXP_CARD_1["language"]),
    ]),
)

FIXTURES[("Profile", "updateExperienceCard")] = fixture(
    "updateExperienceCard",
    EXP_CARD_1,
    request_body=OrderedDict([
        ("result", EXP_CARD_1["result"]),
    ]),
)


# ---------- Resumes tag ------------------------------------------------------

RESUME_ASSET = OrderedDict([
    ("id", RESUME_ASSET_ID),
    ("title", "Alice Example — Senior Frontend Engineer"),
    ("language", "zh-CN"),
    ("parseStatus", "ready"),
    ("fileObjectId", RESUME_FILE_OBJECT_ID),
    ("parsedSummary", OrderedDict([
        ("headline", "Senior frontend engineer focused on growth-stage SaaS"),
        ("yearsOfExperience", 5),
    ])),
    ("createdAt", EARLIEST),
    ("updatedAt", EARLIER),
])

FIXTURES[("Resumes", "registerResume")] = fixture(
    "registerResume",
    OrderedDict([
        ("resumeAssetId", RESUME_ASSET_ID),
        ("job", OrderedDict([
            ("id", JOB_ID_RESUME_PARSE),
            ("jobType", "resume_parse"),
            ("status", "queued"),
            ("resourceType", "resume_asset"),
            ("resourceId", RESUME_ASSET_ID),
            ("errorCode", None),
            ("createdAt", NOW),
            ("updatedAt", NOW),
        ])),
    ]),
    status=202,
    request_body=OrderedDict([
        ("fileObjectId", RESUME_FILE_OBJECT_ID),
        ("title", "Alice Example — Senior Frontend Engineer"),
        ("language", "zh-CN"),
    ]),
)

FIXTURES[("Resumes", "getResume")] = fixture("getResume", RESUME_ASSET)


# ---------- TargetJobs tag ---------------------------------------------------

TARGET_JOB_FULL = OrderedDict([
    ("id", TARGET_JOB_ID_1),
    ("status", "interviewing"),
    ("analysisStatus", "ready"),
    ("title", "Senior Frontend Engineer"),
    ("companyName", "Acme"),
    ("locationText", "Shanghai · Hybrid"),
    ("targetLanguage", "zh-CN"),
    ("sourceType", "url"),
    ("sourceUrl", "https://acme.example/careers/senior-frontend"),
    ("summary", OrderedDict([
        ("coreThemes", [
            "Design-system & component-library leadership",
            "Performance budgets and SSR pipelines",
            "Cross-team RFC ownership",
        ]),
        ("interviewHypotheses", [
            "Will probe scaling design systems across 10+ teams",
            "Likely STAR follow-up on observability for FE infra",
        ]),
        ("provenance", PROV_TARGET_SUMMARY),
    ])),
    ("requirements", [
        OrderedDict([
            ("id", TARGET_REQ_ID_1),
            ("kind", "must_have"),
            ("label", "5+ years building component libraries used by ≥5 teams"),
            ("evidenceLevel", "explicit"),
        ]),
        OrderedDict([
            ("id", TARGET_REQ_ID_2),
            ("kind", "must_have"),
            ("label", "Production experience with SSR / streaming rendering"),
            ("evidenceLevel", "explicit"),
        ]),
        OrderedDict([
            ("id", TARGET_REQ_ID_3),
            ("kind", "nice_to_have"),
            ("label", "Familiarity with edge runtime deployments"),
            ("evidenceLevel", "inferred"),
        ]),
    ]),
    ("fitSummary", OrderedDict([
        ("strengths", [
            "Drove design-system migration across 12 teams (matches must-have #1)",
            "p95 first-paint cut to 1.1s with SSR (matches must-have #2)",
        ]),
        ("gaps", [
            "Limited explicit edge-runtime experience",
        ]),
        ("riskSignals", [
            "Role mentions hiring manager prefers very deep observability stories",
        ]),
        ("provenance", PROV_TARGET_FIT),
    ])),
    ("latestReportId", REPORT_ID_1),
    ("openQuestionIssueCount", 2),
    ("createdAt", EARLIEST),
    ("updatedAt", EARLIER),
])

TARGET_JOB_LIST_ITEM_2 = OrderedDict([
    ("id", TARGET_JOB_ID_2),
    ("status", "preparing"),
    ("analysisStatus", "ready"),
    ("title", "Frontend Architect"),
    ("companyName", "Acme"),
    ("locationText", "Remote (APAC)"),
    ("targetLanguage", "en"),
    ("sourceType", "manual_text"),
    ("sourceUrl", None),
    ("summary", None),
    ("requirements", []),
    ("fitSummary", None),
    ("latestReportId", None),
    ("openQuestionIssueCount", 0),
    ("createdAt", EARLIEST),
    ("updatedAt", EARLIER),
])

FIXTURES[("TargetJobs", "importTargetJob")] = fixture(
    "importTargetJob",
    OrderedDict([
        ("targetJobId", TARGET_JOB_ID_1),
        ("job", OrderedDict([
            ("id", JOB_ID_TARGET_IMPORT),
            ("jobType", "target_import"),
            ("status", "queued"),
            ("resourceType", "target_job"),
            ("resourceId", TARGET_JOB_ID_1),
            ("errorCode", None),
            ("createdAt", NOW),
            ("updatedAt", NOW),
        ])),
    ]),
    status=202,
    request_body=OrderedDict([
        ("source", OrderedDict([
            ("type", "url"),
            ("url", "https://acme.example/careers/senior-frontend"),
        ])),
        ("titleHint", "Senior Frontend Engineer"),
        ("companyNameHint", "Acme"),
        ("targetLanguage", "zh-CN"),
    ]),
)

FIXTURES[("TargetJobs", "listTargetJobs")] = fixture(
    "listTargetJobs",
    OrderedDict([
        ("items", [TARGET_JOB_FULL, TARGET_JOB_LIST_ITEM_2]),
        ("pageInfo", OrderedDict([
            ("nextCursor", None),
            ("pageSize", 20),
            ("hasMore", False),
        ])),
    ]),
)

FIXTURES[("TargetJobs", "getTargetJob")] = fixture("getTargetJob", TARGET_JOB_FULL)

FIXTURES[("TargetJobs", "updateTargetJob")] = fixture(
    "updateTargetJob",
    TARGET_JOB_FULL,
    request_body=OrderedDict([
        ("status", "interviewing"),
        ("notes", "Recruiter mentioned next-round prep is due 2026-05-02."),
    ]),
)


# ---------- PracticePlans tag ------------------------------------------------

PRACTICE_PLAN_1 = OrderedDict([
    ("id", PRACTICE_PLAN_ID_1),
    ("targetJobId", TARGET_JOB_ID_1),
    ("goal", "baseline"),
    ("mode", "assisted"),
    ("interviewerPersona", "hiring_manager"),
    ("difficulty", "standard"),
    ("language", "zh-CN"),
    ("timeBudgetMinutes", 30),
    ("questionBudget", 6),
    ("status", "ready"),
    ("createdAt", EARLIER),
])

FIXTURES[("PracticePlans", "createPracticePlan")] = fixture(
    "createPracticePlan",
    PRACTICE_PLAN_1,
    status=201,
    request_body=OrderedDict([
        ("targetJobId", TARGET_JOB_ID_1),
        ("goal", "baseline"),
        ("mode", "assisted"),
        ("interviewerPersona", "hiring_manager"),
        ("difficulty", "standard"),
        ("language", "zh-CN"),
        ("questionBudget", 6),
        ("timeBudgetMinutes", 30),
        ("resumeAssetId", RESUME_ASSET_ID),
        ("focusCompetencyCodes", ["communication", "design-systems"]),
    ]),
)

FIXTURES[("PracticePlans", "getPracticePlan")] = fixture("getPracticePlan", PRACTICE_PLAN_1)


# ---------- PracticeSessions tag --------------------------------------------

PRACTICE_TURN_1 = OrderedDict([
    ("id", PRACTICE_TURN_ID_1),
    ("turnIndex", 1),
    ("questionText", "请用 STAR 描述你主导设计系统迁移的项目，重点说明跨 12 个团队的协调过程。"),
    ("questionIntent", "behavioral.leadership.design_system"),
    ("status", "asked"),
    ("askedAt", NOW),
])

PRACTICE_SESSION_1 = OrderedDict([
    ("id", PRACTICE_SESSION_ID_1),
    ("planId", PRACTICE_PLAN_ID_1),
    ("targetJobId", TARGET_JOB_ID_1),
    ("status", "running"),
    ("language", "zh-CN"),
    ("hintsEnabled", True),
    ("turnCount", 1),
    ("currentTurn", PRACTICE_TURN_1),
    ("createdAt", EARLIER),
    ("updatedAt", NOW),
])

FIXTURES[("PracticeSessions", "startPracticeSession")] = fixture(
    "startPracticeSession",
    PRACTICE_SESSION_1,
    status=201,
    request_body=OrderedDict([
        ("planId", PRACTICE_PLAN_ID_1),
        ("hintsEnabled", True),
    ]),
)

FIXTURES[("PracticeSessions", "getPracticeSession")] = fixture("getPracticeSession", PRACTICE_SESSION_1)

ASSISTANT_ACTION = OrderedDict([
    ("type", "ask_follow_up"),
    ("turnId", PRACTICE_TURN_ID_1),
    ("questionText", "在 12 个团队里推动迁移时，最大的反对意见是什么？你怎么处理的？"),
    ("hint", None),
    ("sessionStatus", "running"),
    ("provenance", PROV_ASSISTANT_ACTION),
])

FIXTURES[("PracticeSessions", "appendSessionEvent")] = fixture(
    "appendSessionEvent",
    OrderedDict([
        ("acknowledged", True),
        ("session", PRACTICE_SESSION_1),
        ("assistantAction", ASSISTANT_ACTION),
    ]),
    request_body=OrderedDict([
        ("clientEventId", CLIENT_EVENT_ID_1),
        ("kind", "answer_submitted"),
        ("occurredAt", NOW),
        ("payload", OrderedDict([
            ("turnId", PRACTICE_TURN_ID_1),
            ("answerText", "在 Acme 我主导了三个并行的设计系统合并..."),
        ])),
    ]),
)

FIXTURES[("PracticeSessions", "completePracticeSession")] = fixture(
    "completePracticeSession",
    OrderedDict([
        ("reportId", REPORT_ID_1),
        ("job", OrderedDict([
            ("id", JOB_ID_REPORT_GENERATE),
            ("jobType", "report_generate"),
            ("status", "queued"),
            ("resourceType", "feedback_report"),
            ("resourceId", REPORT_ID_1),
            ("errorCode", None),
            ("createdAt", NOW),
            ("updatedAt", NOW),
        ])),
    ]),
    status=202,
    request_body=OrderedDict([
        ("clientCompletedAt", NOW),
    ]),
)


# ---------- Reports / question review ---------------------------------------

QUESTION_ASSESSMENT_1 = OrderedDict([
    ("turnId", PRACTICE_TURN_ID_1),
    ("questionIntent", "behavioral.leadership.design_system"),
    ("dimensionResults", OrderedDict([
        ("communication", OrderedDict([("status", "meets_bar"), ("confidence", "high")])),
        ("technical_depth", OrderedDict([("status", "needs_work"), ("confidence", "medium")])),
        ("ownership", OrderedDict([("status", "strong"), ("confidence", "high")])),
    ])),
    ("reviewStatus", "queued_for_retry"),
    ("includedInRetryPlan", True),
])

FEEDBACK_REPORT = OrderedDict([
    ("id", REPORT_ID_1),
    ("sessionId", PRACTICE_SESSION_ID_1),
    ("targetJobId", TARGET_JOB_ID_1),
    ("status", "ready"),
    ("preparednessLevel", "basically_ready"),
    ("highlights", [
        OrderedDict([
            ("dimension", "ownership"),
            ("evidence", "明确指出推动 12 个团队使用统一组件库；说明了反对意见的处理路径。"),
            ("confidence", "high"),
        ]),
    ]),
    ("issues", [
        OrderedDict([
            ("dimension", "technical_depth"),
            ("evidence", "回答中没有量化迁移过程中的回归率或灰度策略。"),
            ("confidence", "medium"),
        ]),
    ]),
    ("nextActions", [
        OrderedDict([("type", "retry_current_round"), ("label", "围绕灰度策略复练当前轮追问。")]),
        OrderedDict([("type", "review_evidence"), ("label", "回顾迁移期间的回归数据并准备一段量化叙事。")]),
    ]),
    ("questionAssessments", [QUESTION_ASSESSMENT_1]),
    ("retryFocusTurnIds", [PRACTICE_TURN_ID_1]),
    ("provenance", PROV_FEEDBACK_REPORT),
    ("createdAt", EARLIER),
    ("updatedAt", NOW),
])

FIXTURES[("Reports", "getFeedbackReport")] = fixture("getFeedbackReport", FEEDBACK_REPORT)

FIXTURES[("Reports", "listTargetJobReports")] = fixture(
    "listTargetJobReports",
    OrderedDict([
        ("items", [FEEDBACK_REPORT]),
        ("pageInfo", OrderedDict([
            ("nextCursor", None),
            ("pageSize", 20),
            ("hasMore", False),
        ])),
    ]),
)

# ---------- ResumeTailor -----------------------------------------------------

RESUME_TAILOR_RUN = OrderedDict([
    ("id", RESUME_TAILOR_RUN_ID),
    ("status", "ready"),
    ("targetJobId", TARGET_JOB_ID_1),
    ("resumeAssetId", RESUME_ASSET_ID),
    ("matchSummary", OrderedDict([
        ("strengths", [
            "Cross-team design-system leadership directly maps to must-have #1.",
        ]),
        ("gaps", [
            "Edge-runtime exposure is implicit at best; consider adding a quantified bullet.",
        ]),
    ])),
    ("suggestions", [
        OrderedDict([
            ("originalBullet", "Led design-system migration."),
            ("suggestedBullet", "Led design-system migration across 12 teams; reduced UI defect rate by 38% over 6 weeks."),
            ("reason", "Matches must-have #1 and adds quantified outcome."),
        ]),
    ]),
    ("provenance", PROV_RESUME_TAILOR),
    ("createdAt", EARLIER),
    ("updatedAt", NOW),
])

FIXTURES[("ResumeTailor", "requestResumeTailor")] = fixture(
    "requestResumeTailor",
    OrderedDict([
        ("tailorRunId", RESUME_TAILOR_RUN_ID),
        ("job", OrderedDict([
            ("id", JOB_ID_RESUME_TAILOR),
            ("jobType", "resume_tailor"),
            ("status", "queued"),
            ("resourceType", "resume_tailor_run"),
            ("resourceId", RESUME_TAILOR_RUN_ID),
            ("errorCode", None),
            ("createdAt", NOW),
            ("updatedAt", NOW),
        ])),
    ]),
    status=202,
    request_body=OrderedDict([
        ("targetJobId", TARGET_JOB_ID_1),
        ("resumeAssetId", RESUME_ASSET_ID),
        ("mode", "bullet_suggestions"),
    ]),
)

FIXTURES[("ResumeTailor", "getResumeTailorRun")] = fixture("getResumeTailorRun", RESUME_TAILOR_RUN)


# ---------- Debriefs ---------------------------------------------------------

DEBRIEF = OrderedDict([
    ("id", DEBRIEF_ID),
    ("targetJobId", TARGET_JOB_ID_1),
    ("status", "completed"),
    ("roundType", "hiring_manager"),
    ("interviewerRole", "hiring_manager"),
    ("questions", [
        OrderedDict([
            ("questionText", "你过去推动跨团队 alignment 的最难一次是什么？"),
            ("myAnswerSummary", "用 STAR 讲了设计系统迁移，但漏掉了财务影响的量化。"),
            ("interviewerReaction", "追问了财务影响 2 次。"),
            ("aiAnalysis", "下一轮提前准备 1-2 个量化的财务/运营指标作为后置 punchline。"),
        ]),
    ]),
    ("riskItems", [
        OrderedDict([("label", "缺少财务量化叙事"), ("severity", "medium")]),
    ]),
    ("provenance", PROV_DEBRIEF),
    ("createdAt", EARLIER),
    ("updatedAt", NOW),
])

FIXTURES[("Debriefs", "createDebrief")] = fixture(
    "createDebrief",
    OrderedDict([
        ("debriefId", DEBRIEF_ID),
        ("job", OrderedDict([
            ("id", JOB_ID_DEBRIEF_GENERATE),
            ("jobType", "debrief_generate"),
            ("status", "queued"),
            ("resourceType", "debrief"),
            ("resourceId", DEBRIEF_ID),
            ("errorCode", None),
            ("createdAt", NOW),
            ("updatedAt", NOW),
        ])),
    ]),
    status=202,
    request_body=OrderedDict([
        ("targetJobId", TARGET_JOB_ID_1),
        ("roundType", "hiring_manager"),
        ("interviewerRole", "hiring_manager"),
        ("language", "zh-CN"),
        ("questions", [
            OrderedDict([
                ("questionText", "你过去推动跨团队 alignment 的最难一次是什么？"),
                ("myAnswerSummary", "用 STAR 讲了设计系统迁移，但漏掉了财务影响的量化。"),
                ("interviewerReaction", "追问了财务影响 2 次。"),
            ]),
        ]),
        ("notes", "面试官非常关注 ROI 与量化叙事。"),
    ]),
)

FIXTURES[("Debriefs", "getDebrief")] = fixture("getDebrief", DEBRIEF)


# ---------- Jobs -------------------------------------------------------------

FIXTURES[("Jobs", "getJob")] = fixture(
    "getJob",
    OrderedDict([
        ("id", JOB_ID_REPORT_GENERATE),
        ("jobType", "report_generate"),
        ("status", "running"),
        ("resourceType", "feedback_report"),
        ("resourceId", REPORT_ID_1),
        ("errorCode", None),
        ("createdAt", EARLIER),
        ("updatedAt", NOW),
    ]),
)


# ---------- Privacy ----------------------------------------------------------

FIXTURES[("Privacy", "requestPrivacyExport")] = fixture(
    "requestPrivacyExport",
    OrderedDict([
        ("error", OrderedDict([
            ("code", "PRIVACY_EXPORT_NOT_AVAILABLE"),
            ("message", "Privacy export is not available in P0 (spec D-12 / ADR-Q5)."),
            ("requestId", REQUEST_ID),
            ("retryable", False),
            ("details", OrderedDict([
                ("availableAt", "P1"),
                ("see", "docs/spec/openapi-v1-contract/spec.md#3-用户决策--待确认事项"),
            ])),
        ])),
    ]),
    status=501,
)

FIXTURES[("Privacy", "requestPrivacyDelete")] = fixture(
    "requestPrivacyDelete",
    OrderedDict([
        ("privacyRequestId", PRIVACY_REQUEST_ID_DEL),
        ("job", OrderedDict([
            ("id", JOB_ID_PRIVACY_DELETE),
            ("jobType", "privacy_delete"),
            ("status", "queued"),
            ("resourceType", "privacy_request"),
            ("resourceId", PRIVACY_REQUEST_ID_DEL),
            ("errorCode", None),
            ("createdAt", NOW),
            ("updatedAt", NOW),
        ])),
    ]),
    status=202,
)

FIXTURES[("Privacy", "getPrivacyRequest")] = fixture(
    "getPrivacyRequest",
    OrderedDict([
        ("id", PRIVACY_REQUEST_ID_GET),
        ("type", "delete"),
        ("status", "completed"),
        ("requestedAt", EARLIER),
        ("completedAt", NOW),
        ("artifactUrl", None),
    ]),
)


def main() -> None:
    for (tag, opid), data in FIXTURES.items():
        path = FIX / tag / f"{opid}.json"
        path.parent.mkdir(parents=True, exist_ok=True)
        with path.open("w", encoding="utf-8") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
            f.write("\n")
    print(f"wrote {len(FIXTURES)} fixtures under {FIX}")


if __name__ == "__main__":
    main()
