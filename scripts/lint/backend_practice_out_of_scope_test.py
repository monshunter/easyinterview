from pathlib import Path

import backend_practice_out_of_scope


def seed_current_inventory(repo: Path) -> None:
    for parts in backend_practice_out_of_scope.REQUIRED_FILES:
        path = repo.joinpath(*parts)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("current conversation artifact\n", encoding="utf-8")


def test_scan_active_surfaces_flags_current_structured_practice_tokens(tmp_path: Path) -> None:
    seed_current_inventory(tmp_path)
    path = tmp_path / "backend/internal/practice/session.go"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        "type PracticeTurn struct{}\nconst first = \"practice.session.first_question\"\n",
        encoding="utf-8",
    )

    problems = backend_practice_out_of_scope.scan_active_surfaces(tmp_path)

    assert problems == [
        f"{path}:1: stale structured-practice contract PracticeTurn",
        f"{path}:2: stale structured-practice contract first-question feature key",
    ]


def test_scan_active_surfaces_ignores_tests_docs_and_excluded_directories(tmp_path: Path) -> None:
    seed_current_inventory(tmp_path)
    ignored = (
        tmp_path / "backend/internal/practice/session_test.go",
        tmp_path / "docs/reports/evidence.md",
        tmp_path / "node_modules/generated.go",
    )
    for path in ignored:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("type PracticeTurn struct{}\n", encoding="utf-8")

    assert backend_practice_out_of_scope.scan_active_surfaces(tmp_path) == []


def test_scan_file_inventory_requires_current_artifacts_and_rejects_retired_files(tmp_path: Path) -> None:
    missing = backend_practice_out_of_scope.scan_file_inventory(tmp_path)
    assert len(missing) == len(backend_practice_out_of_scope.REQUIRED_FILES)

    seed_current_inventory(tmp_path)
    assert backend_practice_out_of_scope.scan_file_inventory(tmp_path) == []

    forbidden = tmp_path.joinpath(*backend_practice_out_of_scope.FORBIDDEN_FILES[0])
    forbidden.parent.mkdir(parents=True, exist_ok=True)
    forbidden.write_text("retired\n", encoding="utf-8")
    assert backend_practice_out_of_scope.scan_file_inventory(tmp_path) == [
        f"{forbidden}: retired structured-practice artifact still exists"
    ]


def test_main_phase0_checks_inventory_and_phase3_adds_semantic_scan(tmp_path: Path) -> None:
    seed_current_inventory(tmp_path)
    stale = tmp_path / "backend/internal/store/practice/messages.go"
    stale.write_text("const practiceMode = \"strict\"\n", encoding="utf-8")

    assert backend_practice_out_of_scope.main(["--repo-root", str(tmp_path), "--phase", "phase0"]) == 0
    assert backend_practice_out_of_scope.main(["--repo-root", str(tmp_path), "--phase", "phase3"]) == 1


def test_current_repository_passes_phase3_scan() -> None:
    repo = Path(__file__).resolve().parents[2]
    assert backend_practice_out_of_scope.scan_file_inventory(repo) == []
    assert backend_practice_out_of_scope.scan_active_surfaces(repo) == []
