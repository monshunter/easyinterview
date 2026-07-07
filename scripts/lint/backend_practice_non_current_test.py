from pathlib import Path

import backend_practice_non_current


def test_flags_non_current_practice_mode_context(tmp_path: Path) -> None:
    non_current = "debrief" + "_replay"
    path = tmp_path / "active.md"
    path.write_text(f"PracticeMode still lists `{non_current}`\n", encoding="utf-8")

    problems = backend_practice_non_current.scan_paths([path], tmp_path)

    assert problems == [f"{path}:1: non-current practice mode literal in active context"]


def test_ignores_evidence_reports(tmp_path: Path) -> None:
    non_current = "debrief" + "_replay"
    path = tmp_path / "docs/reports/report.md"
    path.parent.mkdir(parents=True)
    path.write_text(f"PracticeMode used to list `{non_current}`\n", encoding="utf-8")

    assert backend_practice_non_current.scan_paths([path], tmp_path) == []


def test_iter_repo_files_skips_symlinked_files_outside_repo(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    repo.mkdir()
    active = repo / "active.md"
    active.write_text("active file\n", encoding="utf-8")
    outside = tmp_path / "python3.13"
    outside.write_text("external interpreter shim\n", encoding="utf-8")
    symlink = repo / "node_modules" / ".bin" / "python3.13"
    symlink.parent.mkdir(parents=True)
    symlink.symlink_to(outside)

    assert backend_practice_non_current.iter_repo_files(repo) == [active]


def test_phase3_flags_non_current_terms_on_practice_surfaces(tmp_path: Path) -> None:
    path = tmp_path / "backend/internal/practice/session.go"
    path.parent.mkdir(parents=True)
    path.write_text('const old = "single_drill"\n', encoding="utf-8")

    problems = backend_practice_non_current.scan_phase3_paths([path], tmp_path)

    assert problems == [f"{path}:1: non-current backend-practice term 'single_drill'"]


def test_phase3_scans_backend_practice_003_scenario_assets(tmp_path: Path) -> None:
    path = tmp_path / "test/scenarios/e2e/p0-048-practice-hint-assisted-across-goals/data/expected-outcome.md"
    path.parent.mkdir(parents=True)
    non_current_term = "leg" "acy_hint_policy"
    path.write_text(f"{non_current_term}\n", encoding="utf-8")

    problems = backend_practice_non_current.scan_phase3_paths([path], tmp_path)

    assert problems == [f"{path}:1: non-current backend-practice term '{non_current_term}'"]


def test_phase3_ignores_owner_plan_gate_wording(tmp_path: Path) -> None:
    path = tmp_path / "docs/spec/backend-practice/plans/001-plan-and-session-orchestration/checklist.md"
    path.parent.mkdir(parents=True)
    path.write_text("non-current-negative grep checks warmup and practiceModeCard\n", encoding="utf-8")

    assert backend_practice_non_current.scan_phase3_paths([path], tmp_path) == []


def test_phase3_flags_standalone_voice_route_but_allows_voice_mvp_placeholder(tmp_path: Path) -> None:
    bad = tmp_path / "backend/cmd/api/routes.go"
    bad.parent.mkdir(parents=True)
    bad.write_text('router.Handle("/voice", h)\n', encoding="utf-8")
    nested_bad = tmp_path / "backend/cmd/api/nested_routes.go"
    nested_bad.write_text('router.Handle("/api/v1/voice/sessions", h)\n', encoding="utf-8")
    allowed = tmp_path / "backend/internal/practice/comment.go"
    allowed.parent.mkdir(parents=True)
    allowed.write_text("// practice-voice-mvp may keep voice operation placeholder\n", encoding="utf-8")
    session_scoped = tmp_path / "backend/cmd/api/session_voice.go"
    session_scoped.write_text(
        'mux.Handle("POST /api/v1/practice/sessions/{sessionId}/voice-turns", h)\n',
        encoding="utf-8",
    )
    feature_key = tmp_path / "backend/internal/practice/voice_feature.go"
    feature_key.write_text('const key = "practice.voice.stt"\n', encoding="utf-8")

    problems = backend_practice_non_current.scan_phase3_paths(
        [bad, nested_bad, allowed, session_scoped, feature_key],
        tmp_path,
    )

    assert problems == [
        f"{bad}:1: non-current standalone voice route",
        f"{nested_bad}:1: non-current standalone voice route",
    ]


def test_phase3_allows_voice_mvp_operation_profiles_and_refs(tmp_path: Path) -> None:
    service = tmp_path / "backend/internal/practice/voice_turn_service.go"
    service.parent.mkdir(parents=True)
    service.write_text(
        "\n".join(
            [
                'const voiceSTTFeatureKey = "practice.voice.stt"',
                'const voiceTTSFeatureKey = "practice.voice.tts"',
                'const persistedRef = "voice-turn://voice-turn-1/chunks/chunk-1"',
            ]
        ),
        encoding="utf-8",
    )
    route = tmp_path / "backend/cmd/api/main.go"
    route.parent.mkdir(parents=True)
    route.write_text(
        'mux.Handle("POST /api/v1/practice/sessions/{sessionId}/voice-turns", createPracticeVoiceTurn)\n',
        encoding="utf-8",
    )
    fixture = tmp_path / "openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json"
    fixture.parent.mkdir(parents=True)
    fixture.write_text(
        '{"sttProfile":"practice.voice.stt.default","ttsProfile":"practice.voice.tts.default"}\n',
        encoding="utf-8",
    )

    assert backend_practice_non_current.scan_phase3_paths([service, route, fixture], tmp_path) == []


def write_backend_practice_002_bdd_inputs(repo: Path, assigned_ids: list[str], test_ids: list[str]) -> None:
    bdd = repo / "docs/spec/backend-practice/plans/002-event-loop-and-completion/bdd-plan.md"
    bdd.parent.mkdir(parents=True)
    bdd.write_text("- 编号分配: " + " / ".join(f"`{scenario_id}`" for scenario_id in assigned_ids) + "\n", encoding="utf-8")
    test_file = repo / "backend/cmd/api/practice_http_scenario_test.go"
    test_file.parent.mkdir(parents=True)
    test_file.write_text("\n".join(f"func TestE2EP0{scenario_id.rsplit('.', maxsplit=1)[1]}Practice(t *testing.T) {{}}" for scenario_id in test_ids), encoding="utf-8")
    index = repo / "test/scenarios/e2e/INDEX.md"
    index.parent.mkdir(parents=True)
    index.write_text("| E2E.P0.034 | backend-resume register/list |\n| E2E.P0.035 | backend-resume parse lifecycle |\n", encoding="utf-8")


def test_backend_practice_002_bdd_ids_do_not_collide_with_indexed_resume_ids(tmp_path: Path) -> None:
    assigned = [f"E2E.P0.{number:03d}" for number in range(38, 44)]
    write_backend_practice_002_bdd_inputs(tmp_path, assigned, assigned)

    assert backend_practice_non_current.scan_backend_practice_002_bdd_ids(tmp_path) == []


def test_backend_practice_002_bdd_ids_flag_non_practice_index_collision(tmp_path: Path) -> None:
    assigned = ["E2E.P0.034", "E2E.P0.039", "E2E.P0.040", "E2E.P0.041", "E2E.P0.042", "E2E.P0.043"]
    write_backend_practice_002_bdd_inputs(tmp_path, assigned, assigned)

    problems = backend_practice_non_current.scan_backend_practice_002_bdd_ids(tmp_path)

    assert problems == [
        f"{tmp_path / 'test/scenarios/e2e/INDEX.md'}: backend-practice 002 id E2E.P0.034 collides with indexed scenario: | E2E.P0.034 | backend-resume register/list |"
    ]


def test_backend_practice_002_bdd_ids_require_matching_http_scenario_tests(tmp_path: Path) -> None:
    assigned = [f"E2E.P0.{number:03d}" for number in range(38, 44)]
    write_backend_practice_002_bdd_inputs(tmp_path, assigned, assigned[:-1])

    problems = backend_practice_non_current.scan_backend_practice_002_bdd_ids(tmp_path)

    assert problems == [
        f"{tmp_path / 'backend/cmd/api/practice_http_scenario_test.go'}: missing Go HTTP scenario test for E2E.P0.043 (TestE2EP0043*)"
    ]
