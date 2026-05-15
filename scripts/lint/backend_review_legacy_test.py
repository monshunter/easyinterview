from pathlib import Path

import backend_review_legacy


def test_backend_review_legacy_includes_terms() -> None:
    terms = set(backend_review_legacy.RETIRED_TERMS)

    assert {"reportLayout", "readiness_score", "attempt_count", "leased_at", "review_method_version"} <= terms


def test_backend_review_legacy_flags_runtime_terms(tmp_path: Path) -> None:
    path = tmp_path / "backend/internal/review/old.go"
    path.parent.mkdir(parents=True)
    path.write_text('const old = "reportLayout"\n', encoding="utf-8")

    problems = backend_review_legacy.scan_paths([path], tmp_path)

    assert problems == [f"{path}:1: retired backend-review term 'reportLayout'"]


def test_backend_review_legacy_allows_negative_docs(tmp_path: Path) -> None:
    path = tmp_path / "docs/spec/backend-review/plans/001-report-generation-baseline/checklist.md"
    path.parent.mkdir(parents=True)
    path.write_text("legacy-negative gate lists reportLayout and attempt_count\n", encoding="utf-8")

    assert backend_review_legacy.scan_paths([path], tmp_path) == []
