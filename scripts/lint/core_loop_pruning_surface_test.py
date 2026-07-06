from pathlib import Path
import subprocess
import sys

import core_loop_pruning_surface as audit


SCRIPT = Path(__file__).resolve().with_name("core_loop_pruning_surface.py")


def write(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body, encoding="utf-8")


def bucket_paths(report: audit.AuditReport, bucket: str) -> list[str]:
    return [finding.path for finding in report.buckets[bucket]]


def test_buckets_retired_surface_hits_by_allowed_context(tmp_path: Path) -> None:
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

    assert bucket_paths(report, "historical_migrations") == [
        "migrations/000009_jd_match_baseline.up.sql"
    ]
    assert bucket_paths(report, "legacy_normalization") == [
        "frontend/src/app/normalizeRoute.ts"
    ]
    assert bucket_paths(report, "negative_tests") == [
        "frontend/src/app/routeUrl.test.ts"
    ]
    assert bucket_paths(report, "real_residuals") == [
        "backend/internal/api/debriefs/handler.go"
    ]


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
        'const legacy = { "/debrief": "home", "/profile": "home" };\n',
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
    assert "historical_migrations (1)" in result.stdout
    assert "legacy_normalization (1)" in result.stdout
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
