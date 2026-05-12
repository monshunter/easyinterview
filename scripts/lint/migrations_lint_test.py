from pathlib import Path

import pytest

import migrations_lint


def test_registered_check_passes(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="CREATE TABLE users (status text CHECK (status IN ('active', 'deleted')));",
        enum_sources="""
checks:
  - table: users
    column: status
    source: db-local
    values: [active, deleted]
    checksum: sha256:d5327efbfaf86d0c
""",
    )

    assert migrations_lint.run_checks(repo) == []


def test_unregistered_check_fails(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="CREATE TABLE users (status text CHECK (status IN ('active', 'deleted')));",
        enum_sources="checks: []\n",
    )

    problems = migrations_lint.run_checks(repo)

    assert any("users.status" in problem and "not registered" in problem for problem in problems)


def test_checksum_drift_fails(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="CREATE TABLE users (status text CHECK (status IN ('active', 'deleted')));",
        enum_sources="""
checks:
  - table: users
    column: status
    source: db-local
    values: [active, disabled]
    checksum: sha256:wrong
""",
    )

    problems = migrations_lint.run_checks(repo)

    assert any("users.status" in problem and "checksum drift" in problem for problem in problems)


def test_secret_red_lines_allow_hashes(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="""
CREATE TABLE auth_challenges (
  challenge_token_hash text NOT NULL,
  raw_token text
);
""",
        enum_sources="checks: []\n",
    )

    problems = migrations_lint.run_checks(repo)

    assert any("raw_token" in problem for problem in problems)
    assert not any("challenge_token_hash" in problem for problem in problems)


def test_shared_conventions_source_drift_fails(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="CREATE TABLE target_jobs (status text CHECK (status IN ('draft', 'archived')));",
        enum_sources="""
checks:
  - table: target_jobs
    column: status
    source: shared-conventions-codified
    values: [draft, archived]
    checksum: sha256:e6e609a18341d0a8
""",
    )
    write_shared_conventions(repo, target_job_status=["draft", "preparing", "archived"])

    problems = migrations_lint.run_checks(repo)

    assert any("target_jobs.status" in problem and "shared/conventions.yaml" in problem for problem in problems)


def test_event_jobs_and_openapi_subset_drift_fails(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="CREATE TABLE async_jobs (job_type text CHECK (job_type IN ('target_import', 'email_dispatch')));",
        enum_sources="""
checks:
  - table: async_jobs
    column: job_type
    source: event-and-outbox-contract
    values: [target_import, email_dispatch]
    checksum: sha256:9b918c6474ad59d8
""",
    )
    write_shared_jobs(
        repo,
        api_facing=["target_import", "email_dispatch"],
        jobs=[("target_import", True), ("email_dispatch", False)],
    )
    write_openapi_job_types(repo, ["target_import", "email_dispatch"])

    problems = migrations_lint.run_checks(repo)

    assert any("shared/jobs.yaml apiFacingSubset" in problem and "email_dispatch" in problem for problem in problems)
    assert any("OpenAPI JobType" in problem and "email_dispatch" in problem for problem in problems)


def test_missing_shared_source_file_fails(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="CREATE TABLE target_jobs (status text CHECK (status IN ('draft', 'archived')));",
        enum_sources="""
checks:
  - table: target_jobs
    column: status
    source: shared-conventions-codified
    values: [draft, archived]
    checksum: sha256:e6e609a18341d0a8
""",
    )

    problems = migrations_lint.run_checks(repo)

    assert any("shared/conventions.yaml" in problem and "does not exist" in problem for problem in problems)


def test_alter_table_check_requires_registration(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="""
CREATE TABLE users (status text);
ALTER TABLE users ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'deleted'));
""",
        enum_sources="checks: []\n",
    )

    problems = migrations_lint.run_checks(repo)

    assert any("users.status" in problem and "not registered" in problem for problem in problems)


def test_create_table_if_not_exists_nullable_check_passes(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="""
CREATE TABLE IF NOT EXISTS resume_versions (
  version_type text NOT NULL CHECK (version_type IN ('structured_master', 'targeted')),
  seed_strategy text CHECK (seed_strategy IS NULL OR seed_strategy IN ('copy_master', 'blank', 'ai_select'))
);
""",
        enum_sources="""
checks:
  - table: resume_versions
    column: version_type
    source: shared-conventions-codified.ResumeVersionType
    values: [structured_master, targeted]
    checksum: sha256:af6773b6df136007
  - table: resume_versions
    column: seed_strategy
    source: shared-conventions-codified.ResumeSeedStrategy
    values: [copy_master, blank, ai_select]
    checksum: sha256:2c9f6e4a35c9537e
""",
    )

    assert migrations_lint.run_checks(repo) == []


def test_alter_table_add_column_if_not_exists_check_requires_registration(tmp_path: Path) -> None:
    repo = write_repo(
        tmp_path,
        sql="""
CREATE TABLE resume_assets (id uuid);
ALTER TABLE resume_assets ADD COLUMN IF NOT EXISTS source_type text CHECK (source_type IS NULL OR source_type IN ('upload', 'paste', 'guided'));
""",
        enum_sources="checks: []\n",
    )

    problems = migrations_lint.run_checks(repo)

    assert any("resume_assets.source_type" in problem and "not registered" in problem for problem in problems)


def test_product_scope_contract_accepts_current_schema() -> None:
    sql = current_migration_up_sql()
    enum_sources = current_enum_sources()

    problems = migrations_lint.validate_product_scope_sql(sql, enum_sources)

    assert problems == []


def test_product_scope_contract_requires_binary_practice_mode() -> None:
    normalized_contract = migrations_lint.normalize_sql(current_migration_up_sql() + "\n" + current_enum_sources())
    removed_mode = "debrief" + "_replay"

    assert "mode text not null check (mode in ('assisted', 'strict'))" in normalized_contract
    assert removed_mode not in normalized_contract


def test_product_scope_contract_requires_shared_idempotency_records() -> None:
    normalized_sql = migrations_lint.normalize_sql(current_migration_up_sql())

    assert "create table idempotency_records" in normalized_sql
    assert "unique (user_id, domain, operation, idempotency_key_hash)" in normalized_sql
    assert "create index idx_idempotency_records_expires_at on idempotency_records (expires_at)" in normalized_sql


def test_product_scope_contract_rejects_removed_mistake_schema() -> None:
    sql = current_baseline_sql() + "\nCREATE TABLE mistake_entries (id uuid PRIMARY KEY);\n"

    problems = migrations_lint.validate_product_scope_sql(sql, current_enum_sources())

    assert any("mistake_entries" in problem for problem in problems)


def test_product_scope_contract_rejects_non_session_scoped_report() -> None:
    sql = current_baseline_sql()
    start = sql.index("CREATE TABLE feedback_reports (")
    end = sql.index("CREATE UNIQUE INDEX idx_feedback_reports_session_unique")
    report_table = sql[start:end].replace(
        "  session_id uuid NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,\n",
        "",
        1,
    )
    sql = sql[:start] + report_table + sql[end:]

    problems = migrations_lint.validate_product_scope_sql(sql, current_enum_sources())

    assert any("feedback_reports.session_id" in problem for problem in problems)


def test_product_scope_contract_rejects_feature_key_outside_f3_tables() -> None:
    # F3 prompt-rubric-registry/001-baseline phase 4.2 expanded the
    # feature_key allowlist to include ai_task_runs. The lint must still
    # reject feature_key in any other table; users is a safe canary that
    # has nothing to do with F3 provenance.
    sql = current_baseline_sql().replace(
        "CREATE TABLE users (\n",
        "CREATE TABLE users (\n  feature_key text,\n",
        1,
    )

    problems = migrations_lint.validate_product_scope_sql(sql, current_enum_sources())

    assert any(
        "feature_key" in problem and "ai_task_runs" in problem and "prompt_versions" in problem
        for problem in problems
    ), problems


def test_product_scope_contract_rejects_vendor_model_tokens() -> None:
    sql = current_baseline_sql() + "\n-- fixture leak: openrouter:anthropic/claude-sonnet-4.6\n"

    problems = migrations_lint.validate_product_scope_sql(sql, current_enum_sources())

    assert any("vendor/model" in problem and "openrouter" in problem for problem in problems)


def write_repo(tmp_path: Path, *, sql: str, enum_sources: str) -> Path:
    repo = tmp_path / "repo"
    migrations = repo / "migrations"
    migrations.mkdir(parents=True)
    (migrations / "000001_test.up.sql").write_text(sql)
    (migrations / "000001_test.down.sql").write_text("-- down\n")
    (migrations / "enum-sources.yaml").write_text(enum_sources)
    return repo


def write_shared_conventions(repo: Path, *, target_job_status: list[str]) -> None:
    shared = repo / "shared"
    shared.mkdir(parents=True)
    values = "\n".join(f"      - {value}" for value in target_job_status)
    (shared / "conventions.yaml").write_text(
        f"""
enums:
  - name: TargetJobStatus
    values:
{values}
"""
    )


def write_shared_jobs(repo: Path, *, api_facing: list[str], jobs: list[tuple[str, bool]]) -> None:
    shared = repo / "shared"
    shared.mkdir(parents=True)
    api_values = "\n".join(f"  - {value}" for value in api_facing)
    job_values = "\n".join(
        f"  - canonical: {canonical}\n    apiFacing: {str(api_facing_flag).lower()}"
        for canonical, api_facing_flag in jobs
    )
    (shared / "jobs.yaml").write_text(
        f"""
apiFacingSubset:
{api_values}
jobs:
{job_values}
"""
    )


def write_openapi_job_types(repo: Path, values: list[str]) -> None:
    openapi = repo / "openapi"
    openapi.mkdir(parents=True)
    enum_values = "\n".join(f"          - {value}" for value in values)
    (openapi / "openapi.yaml").write_text(
        f"""
components:
  schemas:
    JobType:
      type: string
      enum:
{enum_values}
"""
    )


def repo_root() -> Path:
    return Path(__file__).resolve().parents[2]


def current_baseline_sql() -> str:
    return (repo_root() / "migrations" / "000001_create_baseline.up.sql").read_text()


def current_migration_up_sql() -> str:
    return "\n".join(path.read_text() for path in sorted((repo_root() / "migrations").glob("*.up.sql")))


def current_enum_sources() -> str:
    return (repo_root() / "migrations" / "enum-sources.yaml").read_text()
