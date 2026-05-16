from pathlib import Path

import backend_debrief_legacy


def test_backend_debrief_legacy_includes_terms() -> None:
    terms = set(backend_debrief_legacy.RETIRED_TERMS)

    assert {"mistakes_count", "generatedMistakeCount", "experience_library", "debrief_voice"} <= terms


def test_backend_debrief_legacy_flags_runtime_terms(tmp_path: Path) -> None:
    path = tmp_path / "backend/internal/debrief/old.go"
    path.parent.mkdir(parents=True)
    path.write_text('const old = "mistakes_count"\n', encoding="utf-8")

    problems = backend_debrief_legacy.scan_paths([path], tmp_path)

    assert problems == [f"{path}:1: retired backend-debrief term 'mistakes_count'"]


def test_backend_debrief_legacy_allows_negative_docs(tmp_path: Path) -> None:
    path = tmp_path / "docs/spec/backend-debrief/plans/001/checklist.md"
    path.parent.mkdir(parents=True)
    path.write_text("legacy-negative gate lists mistakes_count and debrief_voice\n", encoding="utf-8")

    assert backend_debrief_legacy.scan_paths([path], tmp_path) == []
