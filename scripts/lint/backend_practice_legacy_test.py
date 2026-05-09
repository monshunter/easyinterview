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
