from pathlib import Path
import subprocess
import sys

import core_loop_pruning_surface as audit
import runtime_topology as runtime_topology_lint


SCRIPT = Path(__file__).resolve().with_name("core_loop_pruning_surface.py")
RUNTIME_TOPOLOGY_SCRIPT = SCRIPT.with_name("runtime_topology.py")


def strict_lifecycle_tokens() -> tuple[str, ...]:
    return (
        "ret" "ired",
        "Ret" "ired",
        "de" "precated",
        "De" "precated",
        "退" "役",
    )


def obsolete_zh_status_tokens() -> tuple[str, ...]:
    return (
        "废" "弃",
    )


def old_scope_context_tokens() -> tuple[str, ...]:
    return (
        "\u65e7",
        "\u5386" "\u53f2",
    )


def write(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body, encoding="utf-8")


def bucket_paths(report: audit.AuditReport, bucket: str) -> list[str]:
    return [finding.path for finding in report.buckets[bucket]]


def test_lint_rule_sources_assemble_strict_lifecycle_terms_without_direct_literals() -> None:
    for script in (SCRIPT, RUNTIME_TOPOLOGY_SCRIPT):
        text = script.read_text(encoding="utf-8")
        leaked_tokens = [
            token
            for token in strict_lifecycle_tokens() + obsolete_zh_status_tokens() + old_scope_context_tokens()
            if token in text
        ]

        assert leaked_tokens == [], f"{script.name} leaks direct strict lifecycle tokens: {leaked_tokens}"


def test_negative_context_regexes_match_strict_lifecycle_terms() -> None:
    for token in strict_lifecycle_tokens():
        assert audit.NEGATIVE_CONTEXT_RE.search(token)
        assert runtime_topology_lint.OWNER_NEGATIVE_CONTEXT_RE.search(token)
    for token in obsolete_zh_status_tokens():
        assert audit.NEGATIVE_CONTEXT_RE.search(token)
    for token in old_scope_context_tokens():
        assert audit.NEGATIVE_CONTEXT_RE.search(token)
        assert runtime_topology_lint.OWNER_NEGATIVE_CONTEXT_RE.search(token)


def test_buckets_out_of_scope_surface_hits_by_allowed_context(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "migrations" / "000009_jd_match_baseline.up.sql",
        "CREATE TABLE jd_match_recommendations (id uuid);\n",
    )
    write(
        repo / "frontend" / "src" / "app" / "normalizeRoute.ts",
        'const aliases = { debrief: "home", profile: "home", jd_match: "home" };\n',
    )
    write(
        repo / "frontend" / "src" / "app" / "routeUrl.test.ts",
        'expect(normalizeRoute({ name: "debrief" })).toEqual({ name: "home" });\n',
    )
    write(
        repo / "backend" / "internal" / "api" / "debriefs" / "handler.go",
        "package debriefs\n",
    )

    report = audit.scan_repo(repo)

    assert bucket_paths(report, "migration_records") == [
        "migrations/000009_jd_match_baseline.up.sql"
    ]
    assert bucket_paths(report, "out_of_scope_normalization") == [
        "frontend/src/app/normalizeRoute.ts"
    ]
    assert bucket_paths(report, "negative_tests") == [
        "frontend/src/app/routeUrl.test.ts"
    ]
    assert bucket_paths(report, "real_residuals") == [
        "backend/internal/api/debriefs/handler.go"
    ]


def test_negative_context_words_do_not_allow_production_runtime_residuals(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "backend" / "internal" / "api" / "debriefs" / "handler.go",
        "package debriefs // delete this debrief handler before release\n",
    )

    report = audit.scan_repo(repo)

    assert bucket_paths(report, "real_residuals") == [
        "backend/internal/api/debriefs/handler.go"
    ]
    assert bucket_paths(report, "negative_tests") == []


def test_cli_fails_when_real_residuals_exist(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "config" / "prompts" / "jd_match.search" / "v0.1.0.yaml",
        "feature_key: jd_match.search\n",
    )

    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    assert result.returncode == 1
    assert "real_residuals (1)" in result.stderr
    assert "config/prompts/jd_match.search/v0.1.0.yaml" in result.stderr


def test_cli_passes_with_only_allowed_buckets(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "migrations" / "000014_drop_jd_match_module.up.sql",
        "DROP TABLE IF EXISTS jd_match_recommendations;\n",
    )
    write(
        repo / "frontend" / "src" / "app" / "routeUrl.ts",
        'const outOfScope = { "/debrief": "home", "/profile": "home" };\n',
    )
    write(
        repo / "scripts" / "lint" / "core_loop_pruning_surface_test.py",
        "assert 'candidate_profiles' not in sql\n",
    )

    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    assert result.returncode == 0, result.stderr
    assert "migration_records (1)" in result.stdout
    assert "out_of_scope_normalization (1)" in result.stdout
    assert "negative_tests (1)" in result.stdout
    assert "real_residuals (0)" in result.stdout


def test_model_profile_paths_are_not_candidate_profile_surface(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "backend" / "internal" / "ai" / "aiclient" / "profile" / "loader.go",
        'package profile\n\nconst env = "AI_MODEL_PROFILE_PATH"\n',
    )
    write(
        repo / "backend" / "internal" / "ai" / "aiclient" / "README.md",
        "Model Profile schema lives in [`profile/`](./profile/loader.go).\n",
    )

    report = audit.scan_repo(repo)

    assert report.buckets["real_residuals"] == []
    assert all(not findings for findings in report.buckets.values())


def test_mjs_contract_tests_are_negative_test_bucket(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "ui-design" / "ui-design-contract.test.mjs",
        'assert.doesNotMatch(app, /nav\\("profile"\\)|nav\\("debrief"\\)/);\n',
    )

    report = audit.scan_repo(repo)

    assert bucket_paths(report, "negative_tests") == [
        "ui-design/ui-design-contract.test.mjs"
    ]
    assert report.buckets["real_residuals"] == []
