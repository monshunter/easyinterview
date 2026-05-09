from pathlib import Path

import backend_practice_preflight


def test_completed_header_passes(tmp_path: Path) -> None:
    doc = tmp_path / "plan.md"
    doc.write_text(
        "\n".join(
            [
                "# Plan",
                "",
                "> **版本**: 1.2",
                "> **状态**: completed",
                "> **更新日期**: 2026-05-09",
                "",
            ]
        ),
        encoding="utf-8",
    )

    assert backend_practice_preflight.header_status(doc) == "completed"
    assert backend_practice_preflight.validate_completed_headers([doc]) == []


def test_active_header_fails(tmp_path: Path) -> None:
    doc = tmp_path / "plan.md"
    doc.write_text(
        "\n".join(
            [
                "# Plan",
                "",
                "> **版本**: 1.2",
                "> **状态**: active",
                "> **更新日期**: 2026-05-09",
                "",
            ]
        ),
        encoding="utf-8",
    )

    problems = backend_practice_preflight.validate_completed_headers([doc])

    assert problems == [f"{doc}: status active != completed"]
