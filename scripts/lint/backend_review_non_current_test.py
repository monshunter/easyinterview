from pathlib import Path

import backend_review_non_current


def test_backend_review_non_current_includes_terms() -> None:
    terms = set(backend_review_non_current.NON_CURRENT_TERMS)

    assert {"reportLayout", "readiness_score", "attempt_count", "leased_at", "review_method_version"} <= terms


def test_backend_review_non_current_flags_runtime_terms(tmp_path: Path) -> None:
    path = tmp_path / "backend/internal/review/old.go"
    path.parent.mkdir(parents=True)
    path.write_text('const old = "reportLayout"\n', encoding="utf-8")

    problems = backend_review_non_current.scan_paths([path], tmp_path)

    assert problems == [f"{path}:1: non-current backend-review term 'reportLayout'"]


def test_backend_review_non_current_allows_negative_docs(tmp_path: Path) -> None:
    path = tmp_path / "docs/spec/backend-review/plans/001-report-generation-baseline/checklist.md"
    path.parent.mkdir(parents=True)
    path.write_text("non-current-negative gate lists reportLayout and attempt_count\n", encoding="utf-8")

    assert backend_review_non_current.scan_paths([path], tmp_path) == []
