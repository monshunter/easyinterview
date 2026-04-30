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


def write_repo(tmp_path: Path, *, sql: str, enum_sources: str) -> Path:
    repo = tmp_path / "repo"
    migrations = repo / "migrations"
    migrations.mkdir(parents=True)
    (migrations / "000001_test.up.sql").write_text(sql)
    (migrations / "000001_test.down.sql").write_text("-- down\n")
    (migrations / "enum-sources.yaml").write_text(enum_sources)
    return repo
