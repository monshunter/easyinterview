#!/usr/bin/env python3
"""B4 migration lint gate.

This script starts with the file-contract checks needed by migrate-check and is
extended by later B4 phases for enum/check source drift and privacy red flags.
"""

from __future__ import annotations

import argparse
import hashlib
import re
import sys
from pathlib import Path

import yaml


MIGRATION_RE = re.compile(r"^([0-9]{6})_[a-z0-9]+(?:_[a-z0-9]+)*\.(up|down)\.sql$")
CHECK_RE = re.compile(
    r"CHECK\s*\(\s*(?:([a-z_]+)\s+IS\s+NULL\s+OR\s+)?([a-z_]+)\s+IN\s*\((.*?)\)\s*\)",
    re.IGNORECASE | re.DOTALL,
)
CREATE_TABLE_RE = re.compile(
    r"CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-z_]+)\s*\((.*?)\);",
    re.IGNORECASE | re.DOTALL,
)
ALTER_TABLE_RE = re.compile(
    r"ALTER\s+TABLE\s+(?:ONLY\s+)?([a-z_]+)\s+(.*?);",
    re.IGNORECASE | re.DOTALL,
)
VALUE_RE = re.compile(r"'([^']+)'")
FORBIDDEN_SECRET_RE = re.compile(r"\b(raw_token|session_cookie|api_key|provider_token)\b", re.IGNORECASE)
VENDOR_MODEL_TOKEN_RE = re.compile(
    r"\b(openrouter|anthropic|claude|gpt-|m4\.7|primary-llm|model-test|gpt-test|gemini|mistral|cohere)\b",
    re.IGNORECASE,
)

EXPECTED_BASELINE_TABLES = [
    "schema_backfills",
    "users",
    "user_settings",
    "file_objects",
    "resume_assets",
    "resume_versions",
    "resume_version_suggestions",
    "target_jobs",
    "target_job_requirements",
    "practice_plans",
    "idempotency_records",
    "practice_sessions",
    "practice_session_events",
    "practice_messages",
    "feedback_reports",
    "resume_tailor_runs",
    "source_records",
    "prompt_versions",
    "rubric_versions",
    "ai_task_runs",
    "async_jobs",
    "outbox_events",
    "privacy_requests",
    "audit_events",
    "auth_challenges",
    "sessions",
    "external_identities",
    "jd_match_recommendations",
    "watchlist_items",
    "saved_searches",
    "agent_scans",
    "jd_match_search_runs",
]
REMOVED_PRODUCT_SCOPE_TOKENS = [
    "mistake_entries",
    "open_mistake_count",
    "written_to_mistake_book",
    "mistake_book",
    "mistake.extract",
    "'single_drill'",
    "'core_interview'",
    "'fix_mistake'",
    "'counter_questions'",
    "voice_sessions",
    "voice_practice",
    "growth_stage",
    "growth_plan",
    "drill_session",
]
PRODUCT_SCOPE_REQUIRED_FRAGMENTS = [
    ("target_jobs.open_question_issue_count", "open_question_issue_count integer not null default 0"),
    ("practice_plans.goal", "goal text not null check (goal in ('baseline', 'retry_current_round', 'next_round'))"),
    ("idempotency_records.unique_key", "unique (user_id, domain, operation, idempotency_key_hash)"),
    ("idempotency_records.expires_at_index", "create index idx_idempotency_records_expires_at on idempotency_records (expires_at)"),
    ("feedback_reports.session_unique", "create unique index idx_feedback_reports_session_unique on feedback_reports (session_id)"),
    ("feedback_reports.dimension_assessments", "dimension_assessments jsonb not null default '[]'::jsonb"),
    ("feedback_reports.legacy_retry_focus_origin", "retry_focus_competency_codes text[] not null default '{}'::text[]"),
    ("feedback_reports.summary_v18", "add column summary text"),
    ("feedback_reports.generation_context_v18", "add column generation_context jsonb not null default '{}'::jsonb"),
    ("feedback_reports.retry_focus_dimension_codes_v18", "rename column retry_focus_competency_codes to retry_focus_dimension_codes"),
    ("practice_plans.focus_dimension_codes_v18", "rename column focus_competency_codes to focus_dimension_codes"),
    ("outbox_events.publish_attempts", "publish_attempts integer not null default 0"),
    ("outbox_events.next_attempt_at", "next_attempt_at timestamptz not null default now()"),
    ("outbox_events.pending_due_index", "create index idx_outbox_events_pending_due on outbox_events (publish_status, next_attempt_at, created_at)"),
    ("ai_task_runs.model_family", "model_family text"),
    ("ai_task_runs.model_profile_name", "model_profile_name text"),
    ("ai_task_runs.model_profile_version", "model_profile_version text"),
    ("ai_task_runs.feature_key", "feature_key text not null"),
    ("ai_task_runs.feature_flag", "feature_flag text not null default 'none'"),
    ("ai_task_runs.data_source_version", "data_source_version text not null default 'not_applicable'"),
    ("ai_task_runs.fallback_chain", "fallback_chain jsonb not null default '[]'::jsonb"),
    ("ai_task_runs.route", "route text"),
    ("ai_task_runs.validation_status", "validation_status text"),
    ("ai_task_runs.output_schema_version", "output_schema_version text"),
    ("ai_task_runs.dashboard_index", "create index idx_ai_task_runs_dashboard on ai_task_runs (model_profile_name, validation_status, created_at desc)"),
]
PRODUCT_SCOPE_FORBIDDEN_COMPATIBILITY_FRAGMENTS = [
    ("feedback_reports.retry_focus_dimension_codes", "add column retry_focus_dimension_codes"),
    ("practice_plans.focus_dimension_codes", "add column focus_dimension_codes"),
]
PRODUCT_SCOPE_FORBIDDEN_FRAGMENTS = [
    ("feedback_reports.llm_attempt_count", "llm_attempt_count"),
]
PRODUCT_SCOPE_TABLE_REQUIRED_FRAGMENTS = [
    ("feedback_reports.session_id", "feedback_reports", "session_id uuid not null references practice_sessions(id) on delete cascade"),
]
# F3 prompt-rubric-registry/001-baseline phase 4.2 expanded the
# feature_key column scope: ai_task_runs now carries the F3 coordinate as a
# typed column (alongside feature_flag and data_source_version) so B4 cross-
# layer provenance reads do not have to crawl the metadata jsonb. The lint
# requires every active table in the allowlist to keep the typed column.
FEATURE_KEY_ALLOWED_TABLES = {"prompt_versions", "rubric_versions", "ai_task_runs"}
FEATURE_KEY_REQUIRED_FRAGMENTS = {
    "prompt_versions": (
        "feature_key text not null",
        "unique (feature_key, version, language)",
    ),
    "rubric_versions": (
        "feature_key text not null",
        "unique (feature_key, version, language)",
    ),
    "ai_task_runs": (
        "feature_key text not null",
        "feature_flag text not null default 'none'",
        "data_source_version text not null default 'not_applicable'",
    ),
}

B1_SOURCE_MAP = {
    "resume_assets.parse_status": "TargetJobParseStatus",
    "target_jobs.status": "TargetJobStatus",
    "target_jobs.analysis_status": "TargetJobParseStatus",
    "practice_plans.goal": "PracticeGoal",
    "practice_plans.interviewer_persona": "InterviewerRole",
    "practice_sessions.status": "SessionStatus",
    "feedback_reports.status": "ReportStatus",
    "feedback_reports.preparedness_level": "ReadinessTier",
    "async_jobs.status": "JobStatus",
    "privacy_requests.request_type": "PrivacyRequestType",
    "privacy_requests.status": "PrivacyRequestStatus",
}


def validate_file_contract(migrations_dir: Path) -> list[str]:
    problems: list[str] = []
    pairs: dict[int, set[str]] = {}

    if not migrations_dir.exists():
        return [f"migrations dir does not exist: {migrations_dir}"]

    for path in sorted(migrations_dir.iterdir()):
        if path.is_dir():
            if path.name == "backfill":
                continue
            problems.append(f"{path.name} is a directory; migrations must be flat")
            continue
        if path.suffix != ".sql":
            continue
        match = MIGRATION_RE.match(path.name)
        if not match:
            problems.append(f"invalid migration file name: {path.name}")
            continue
        version = int(match.group(1))
        direction = match.group(2)
        pairs.setdefault(version, set()).add(direction)

    versions = sorted(pairs)
    for offset, version in enumerate(versions, start=1):
        if version != offset:
            problems.append(f"expected version {offset:06d}, found {version:06d}")
        directions = pairs[version]
        if "up" not in directions:
            problems.append(f"missing up migration for {version:06d}")
        if "down" not in directions:
            problems.append(f"missing down migration for {version:06d}")

    return problems


def run_checks(repo_root: Path) -> list[str]:
    migrations_dir = repo_root / "migrations"
    problems = validate_file_contract(migrations_dir)
    enum_sources = load_enum_sources(migrations_dir / "enum-sources.yaml")
    checks = extract_sql_checks(migrations_dir)
    problems.extend(validate_enum_sources(enum_sources, checks))
    problems.extend(validate_declared_sources(repo_root, enum_sources))
    problems.extend(validate_secret_red_lines(migrations_dir))
    problems.extend(validate_product_scope_files(migrations_dir))
    return problems


def validate_enum_sources(enum_sources: dict[str, dict], checks: dict[str, list[str]]) -> list[str]:
    problems: list[str] = []

    for key, values in sorted(checks.items()):
        source = enum_sources.get(key)
        if source is None:
            problems.append(f"{key} check list is not registered in migrations/enum-sources.yaml")
            continue
        declared_values = source.get("values") or []
        if declared_values != values:
            problems.append(f"{key} checksum drift: SQL values {values} != enum-sources values {declared_values}")
            continue
        want_checksum = checksum_values(values)
        if source.get("checksum") != want_checksum:
            problems.append(f"{key} checksum drift: expected {want_checksum}, got {source.get('checksum')}")

    for key in sorted(set(enum_sources) - set(checks)):
        problems.append(f"{key} registered in enum-sources.yaml but not present in SQL checks")

    return problems


def load_enum_sources(path: Path) -> dict[str, dict]:
    if not path.exists():
        return {}
    data = yaml.safe_load(path.read_text()) or {}
    out: dict[str, dict] = {}
    for item in data.get("checks", []):
        key = f"{item.get('table')}.{item.get('column')}"
        out[key] = item
    return out


def extract_sql_checks(migrations_dir: Path) -> dict[str, list[str]]:
    checks: dict[str, list[str]] = {}
    for path in sorted(migrations_dir.glob("*.up.sql")):
        sql = path.read_text()
        for table_match in CREATE_TABLE_RE.finditer(sql):
            table = table_match.group(1)
            body = table_match.group(2)
            checks.update(extract_fragment_checks(table, body))
        for alter_match in ALTER_TABLE_RE.finditer(sql):
            table = alter_match.group(1)
            body = alter_match.group(2)
            checks.update(extract_fragment_checks(table, body))
    return checks


def extract_fragment_checks(table: str, sql_fragment: str) -> dict[str, list[str]]:
    checks: dict[str, list[str]] = {}
    for check_match in CHECK_RE.finditer(sql_fragment):
        nullable_column = check_match.group(1)
        column = check_match.group(2)
        if nullable_column and nullable_column.lower() != column.lower():
            continue
        values = VALUE_RE.findall(check_match.group(3))
        checks[f"{table}.{column}"] = values
    return checks


def validate_declared_sources(repo_root: Path, enum_sources: dict[str, dict]) -> list[str]:
    problems: list[str] = []
    if any(item.get("source") == "shared-conventions-codified" for item in enum_sources.values()):
        problems.extend(validate_shared_conventions_source(repo_root, enum_sources))
    if any(item.get("source") == "event-and-outbox-contract" for item in enum_sources.values()):
        problems.extend(validate_event_job_source(repo_root, enum_sources))
    return problems


def validate_shared_conventions_source(repo_root: Path, enum_sources: dict[str, dict]) -> list[str]:
    path = repo_root / "shared" / "conventions.yaml"
    data, error = load_yaml_file(path)
    if error:
        return [error]

    values_by_name = {item.get("name"): item.get("values") or [] for item in data.get("enums", [])}
    values_by_name["JobStatus"] = data.get("jobStatuses") or []
    problems: list[str] = []
    for key, source in sorted(enum_sources.items()):
        if source.get("source") != "shared-conventions-codified":
            continue
        enum_name = B1_SOURCE_MAP.get(key)
        if enum_name is None:
            problems.append(f"{key} source shared-conventions-codified has no B4 source mapping")
            continue
        source_values = values_by_name.get(enum_name)
        if source_values is None:
            problems.append(f"{key} source drift: shared/conventions.yaml missing {enum_name}")
            continue
        declared_values = source.get("values") or []
        if declared_values != source_values:
            problems.append(
                f"{key} source drift: migrations/enum-sources.yaml values {declared_values} != "
                f"shared/conventions.yaml {enum_name} values {source_values}"
            )
    return problems


def validate_event_job_source(repo_root: Path, enum_sources: dict[str, dict]) -> list[str]:
    problems: list[str] = []
    jobs_path = repo_root / "shared" / "jobs.yaml"
    jobs_data, error = load_yaml_file(jobs_path)
    if error:
        return [error]

    jobs = jobs_data.get("jobs") or []
    canonical = [job.get("canonical") for job in jobs if job.get("canonical")]
    expected_api_facing = [job.get("canonical") for job in jobs if job.get("canonical") and job.get("apiFacing") is True]
    declared_api_facing = jobs_data.get("apiFacingSubset") or []

    job_source = enum_sources.get("async_jobs.job_type")
    if job_source and job_source.get("source") == "event-and-outbox-contract":
        declared_values = job_source.get("values") or []
        if declared_values != canonical:
            problems.append(
                f"async_jobs.job_type source drift: migrations/enum-sources.yaml values {declared_values} != "
                f"shared/jobs.yaml canonical values {canonical}"
            )
        if declared_api_facing != expected_api_facing:
            problems.append(
                f"shared/jobs.yaml apiFacingSubset {declared_api_facing} != jobs marked apiFacing=true {expected_api_facing}"
            )
        openapi_values, openapi_error = load_openapi_job_types(repo_root / "openapi" / "openapi.yaml")
        if openapi_error:
            problems.append(openapi_error)
        elif openapi_values != expected_api_facing:
            problems.append(f"OpenAPI JobType enum {openapi_values} != jobs marked apiFacing=true {expected_api_facing}")
    return problems


def load_yaml_file(path: Path) -> tuple[dict, str | None]:
    if not path.exists():
        return {}, f"{path.relative_to(path.parents[1])} does not exist"
    return yaml.safe_load(path.read_text()) or {}, None


def load_openapi_job_types(path: Path) -> tuple[list[str], str | None]:
    if not path.exists():
        return [], f"{path.relative_to(path.parents[1])} does not exist"
    data = yaml.safe_load(path.read_text()) or {}
    try:
        values = data["components"]["schemas"]["JobType"]["enum"]
    except KeyError:
        return [], "OpenAPI JobType enum is missing"
    return values or [], None


def validate_secret_red_lines(migrations_dir: Path) -> list[str]:
    problems: list[str] = []
    for path in sorted(migrations_dir.glob("*.sql")):
        for lineno, line in enumerate(path.read_text().splitlines(), start=1):
            match = FORBIDDEN_SECRET_RE.search(line)
            if match:
                problems.append(f"{path.name}:{lineno}: forbidden plaintext secret field marker {match.group(1)}")
    return problems


def validate_product_scope_files(migrations_dir: Path) -> list[str]:
    if not (migrations_dir / "000001_create_baseline.up.sql").exists():
        return []
    sql = "\n".join(path.read_text() for path in sorted(migrations_dir.glob("*.up.sql")))
    enum_sources = (migrations_dir / "enum-sources.yaml").read_text() if (migrations_dir / "enum-sources.yaml").exists() else ""
    return validate_product_scope_sql(sql, enum_sources)


def validate_product_scope_sql(sql: str, enum_sources: str) -> list[str]:
    problems: list[str] = []
    normalized_sql = normalize_sql(sql)
    normalized_contract = normalize_sql(sql + "\n" + enum_sources)
    table_bodies = extract_create_table_bodies(sql)
    table_names = list(table_bodies)

    missing_tables = [table for table in EXPECTED_BASELINE_TABLES if table not in table_bodies]
    extra_tables = sorted(set(table_names) - set(EXPECTED_BASELINE_TABLES))
    if missing_tables:
        problems.append(f"product-scope baseline missing tables: {missing_tables}")
    if extra_tables:
        problems.append(f"product-scope baseline has unexpected tables: {extra_tables}")

    for token in REMOVED_PRODUCT_SCOPE_TOKENS:
        if token in normalized_contract:
            problems.append(f"product-scope removed token still present: {token}")

    for label, fragment in PRODUCT_SCOPE_REQUIRED_FRAGMENTS:
        if normalize_sql(fragment) not in normalized_sql:
            problems.append(f"product-scope required fragment missing: {label}")
    for label, fragment in PRODUCT_SCOPE_FORBIDDEN_COMPATIBILITY_FRAGMENTS:
        if normalize_sql(fragment) in normalized_sql:
            problems.append(
                f"product-scope compatibility mirror forbidden: {label} ({fragment})"
            )
    for label, fragment in PRODUCT_SCOPE_FORBIDDEN_FRAGMENTS:
        if normalize_sql(fragment) in normalized_sql:
            problems.append(f"product-scope forbidden fragment: {label} ({fragment})")
    for label, table, fragment in PRODUCT_SCOPE_TABLE_REQUIRED_FRAGMENTS:
        if normalize_sql(fragment) not in normalize_sql(table_bodies.get(table, "")):
            problems.append(f"product-scope required fragment missing: {label}")

    problems.extend(validate_feature_key_scope(sql, table_bodies))

    match = VENDOR_MODEL_TOKEN_RE.search(sql + "\n" + enum_sources)
    if match:
        problems.append(f"migration contract contains vendor/model token {match.group(1)!r}; use provider-neutral profile identifiers")

    return problems


def extract_create_table_bodies(sql: str) -> dict[str, str]:
    bodies: dict[str, str] = {}
    for match in CREATE_TABLE_RE.finditer(sql):
        bodies[match.group(1).lower()] = match.group(2)
    return bodies


def validate_feature_key_scope(sql: str, table_bodies: dict[str, str]) -> list[str]:
    problems: list[str] = []
    allowed_label = ", ".join(sorted(FEATURE_KEY_ALLOWED_TABLES))
    for table, body in table_bodies.items():
        normalized_body = normalize_sql(body)
        if "feature_key" in normalized_body and table not in FEATURE_KEY_ALLOWED_TABLES:
            problems.append(f"feature_key is only allowed in {allowed_label}, found in {table}")
        if table in FEATURE_KEY_ALLOWED_TABLES:
            required = FEATURE_KEY_REQUIRED_FRAGMENTS.get(table, ())
            for fragment in required:
                if fragment not in normalized_body:
                    problems.append(f"{table} must keep F3 feature_key coordinate fragment: {fragment}")

    # Strip dollar-quoted body literals AND single-line SQL comments
    # before scanning so multi-line prompt bodies in seed migrations
    # cannot derail the feature_key boundary check (a literal ';' inside
    # a $body$...$body$ string would otherwise truncate the DML span
    # detection below) and so harmless `-- ...` comments mentioning
    # feature_key do not trigger false positives.
    stripped_sql = re.sub(r"\$([a-zA-Z_]*)\$.*?\$\1\$", "", sql, flags=re.DOTALL)
    stripped_sql = re.sub(r"--[^\n]*", "", stripped_sql)
    allowed_spans = [
        match.span()
        for match in CREATE_TABLE_RE.finditer(stripped_sql)
        if match.group(1).lower() in FEATURE_KEY_ALLOWED_TABLES
    ]
    # Also exempt DML statements (INSERT/UPDATE/DELETE/SELECT) targeting
    # the allowlisted tables. Seed migrations under
    # migrations/*seed_baseline_prompt_rubric*.up.sql legitimately
    # reference feature_key in INSERT VALUES against prompt_versions /
    # rubric_versions; flagging those would break the F3 baseline seed
    # contract introduced in plan prompt-rubric-registry/001-baseline §4.4.
    dml_re = re.compile(
        r"(?:INSERT\s+INTO|UPDATE|DELETE\s+FROM|SELECT[^;]*?FROM)\s+("
        + "|".join(re.escape(t) for t in FEATURE_KEY_ALLOWED_TABLES)
        + r")\b[^;]*;",
        re.IGNORECASE | re.DOTALL,
    )
    for match in dml_re.finditer(stripped_sql):
        allowed_spans.append(match.span())
    outside_allowed = remove_spans(stripped_sql.lower(), allowed_spans)
    if "feature_key" in outside_allowed:
        problems.append(f"feature_key is only allowed in {allowed_label}")
    return problems


def remove_spans(text: str, spans: list[tuple[int, int]]) -> str:
    if not spans:
        return text
    chunks: list[str] = []
    cursor = 0
    for start, end in sorted(spans):
        chunks.append(text[cursor:start])
        cursor = end
    chunks.append(text[cursor:])
    return "".join(chunks)


def normalize_sql(value: str) -> str:
    return re.sub(r"\s+", " ", value.lower()).strip()


def checksum_values(values: list[str]) -> str:
    digest = hashlib.sha256("|".join(values).encode()).hexdigest()[:16]
    return f"sha256:{digest}"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    args = parser.parse_args(argv)

    repo_root = Path(args.repo_root).resolve()
    problems = run_checks(repo_root)
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1
    print("migration lint: ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
