from pathlib import Path

import backend_practice_legacy


def test_flags_removed_practice_mode_context(tmp_path: Path) -> None:
    removed = "debrief" + "_replay"
    path = tmp_path / "active.md"
    path.write_text(f"PracticeMode still lists `{removed}`\n", encoding="utf-8")

    problems = backend_practice_legacy.scan_paths([path], tmp_path)

    assert problems == [f"{path}:1: removed practice mode literal in active context"]


def test_ignores_historical_reports(tmp_path: Path) -> None:
    removed = "debrief" + "_replay"
    path = tmp_path / "docs/reports/report.md"
    path.parent.mkdir(parents=True)
    path.write_text(f"PracticeMode used to list `{removed}`\n", encoding="utf-8")

    assert backend_practice_legacy.scan_paths([path], tmp_path) == []


def test_phase3_flags_retired_terms_on_practice_surfaces(tmp_path: Path) -> None:
    path = tmp_path / "backend/internal/practice/session.go"
    path.parent.mkdir(parents=True)
    path.write_text('const old = "single_drill"\n', encoding="utf-8")

    problems = backend_practice_legacy.scan_phase3_paths([path], tmp_path)

    assert problems == [f"{path}:1: retired backend-practice term 'single_drill'"]


def test_phase3_ignores_owner_plan_gate_wording(tmp_path: Path) -> None:
    path = tmp_path / "docs/spec/backend-practice/plans/001-plan-and-session-orchestration/checklist.md"
    path.parent.mkdir(parents=True)
    path.write_text("legacy-negative grep checks warmup and practiceModeCard\n", encoding="utf-8")

    assert backend_practice_legacy.scan_phase3_paths([path], tmp_path) == []


def test_phase3_flags_standalone_voice_route_but_allows_voice_mvp_placeholder(tmp_path: Path) -> None:
    bad = tmp_path / "backend/cmd/api/routes.go"
    bad.parent.mkdir(parents=True)
    bad.write_text('router.Handle("/voice", h)\n', encoding="utf-8")
    allowed = tmp_path / "backend/internal/practice/comment.go"
    allowed.parent.mkdir(parents=True)
    allowed.write_text("// practice-voice-mvp may keep voice operation placeholder\n", encoding="utf-8")

    problems = backend_practice_legacy.scan_phase3_paths([bad, allowed], tmp_path)

    assert problems == [f"{bad}:1: retired standalone voice route"]
