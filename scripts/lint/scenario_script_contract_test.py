import json
import re
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
PRACTICE_SERVICE_SCENARIOS = {
    "p0-070-practice-derived-plan-create-read-replay": (
        "TestCreateDerivedPracticePlanIdempotencyMismatchHasNoSecondInsertOrLeak",
        "TestPracticeChatV020CandidateUsesSemanticFocus",
        "TestE2EP0070PracticeDerivedPlanCreateReadReplay",
    ),
    "p0-072-practice-derived-source-isolation-privacy": (
        "TestCreateDerivedPracticePlanRejectsCopiedServerFields",
        "TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy",
    ),
}
FRONTEND_TEST_PATH_RE = re.compile(
    r"(?<![A-Za-z0-9_./-])((?:frontend/)?(?:src|tests)/[A-Za-z0-9_./-]+\.(?:test|spec)\.(?:tsx|ts))"
)


def test_explicit_frontend_test_paths_in_scenario_triggers_exist() -> None:
    missing: list[str] = []
    triggers = sorted((REPO_ROOT / "test/scenarios/e2e").glob("*/scripts/trigger.sh"))

    for trigger in triggers:
        text = trigger.read_text(encoding="utf-8")
        for reference in sorted(set(FRONTEND_TEST_PATH_RE.findall(text))):
            candidate = (
                REPO_ROOT / reference
                if reference.startswith("frontend/")
                else REPO_ROOT / "frontend" / reference
            )
            if not candidate.is_file():
                missing.append(f"{trigger.relative_to(REPO_ROOT)} -> {reference}")

    assert not missing, "scenario triggers reference missing frontend tests:\n" + "\n".join(
        missing
    )


def test_practice_service_triggers_run_current_named_tests() -> None:
    for scenario, test_names in PRACTICE_SERVICE_SCENARIOS.items():
        trigger = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/trigger.sh"
        text = trigger.read_text(encoding="utf-8")
        for test_name in test_names:
            assert test_name in text
        assert "set -euo pipefail" in text


def test_practice_service_verifiers_require_passing_named_tests() -> None:
    for scenario, test_names in PRACTICE_SERVICE_SCENARIOS.items():
        verify = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/verify.sh"
        text = verify.read_text(encoding="utf-8")

        for test_name in test_names:
            assert f"--- PASS: {test_name}" in text
        assert "--- FAIL:" in text
        assert "no tests to run" in text


def test_report_derived_practice_scenarios_consume_f3_and_legacy_negative_markers() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    p0070 = (
        scenarios
        / "p0-070-practice-derived-plan-create-read-replay/scripts/verify.sh"
    ).read_text(encoding="utf-8")
    p0072 = (
        scenarios
        / "p0-072-practice-derived-source-isolation-privacy/scripts/verify.sh"
    ).read_text(encoding="utf-8")

    assert "PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS" in p0070
    assert "F3_PRACTICE_SEMANTIC_FOCUS_MARKER_CONSUMED_PASS" in p0070
    assert "REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS" in p0072
    assert "PROTOTYPE_MAPPING.md" in p0072
    for legacy in (
        "focusCompetencyCodes",
        "focus_competency_codes",
        "retryFocusCompetencyCodes",
        "retry_focus_competency_codes",
        "retryFocusTurnIds",
        "retry_focus_turn_ids",
        "retry_round",
        "semantic_focus",
    ):
        assert legacy in p0072


def test_practice_failure_and_completion_scenarios_require_regression_markers() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    required = {
        "p0-046-practice-text-loop-failure-and-recovery": (
            "TestSendPracticeMessageProviderFailureKeepsReservationUncommitted",
            "TestSendPracticeMessageExactReplayReturnsOriginalResultWithoutAICall",
            "TestSendPracticeMessageMapsClientMismatchAndCrossUserAccess",
            "TestSQLRepositoryReservePracticeMessageRetriesOnlyRetryableFailure",
            "TestSQLRepositoryReservePracticeMessageRejectsNewMessageWhileReplyPending",
        ),
        "p0-047-practice-text-loop-complete-and-generating-handoff": (
            "TestE2EP0047RejectsZeroAnswerCompletion",
            "TestE2EP0047FreezesReportContext",
            "TestE2EP0047CompletionReplayPreservesReportContext",
        ),
    }

    for scenario, markers in required.items():
        trigger = (scenarios / scenario / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (scenarios / scenario / "scripts/verify.sh").read_text(encoding="utf-8")
        for marker in markers:
            assert marker in trigger
            assert f"--- PASS: {marker}" in verify
        assert "--- FAIL:" in verify
        assert "no tests to run" in verify


def test_practice_failure_verifier_does_not_treat_test_titles_as_runner_failures() -> None:
    verify = (
        REPO_ROOT
        / "test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/verify.sh"
    ).read_text(encoding="utf-8")

    assert "frontend-real-backend-verify.sh" in verify
    assert "[[:space:]][0-9]+ failed([[:space:]]|$)" not in verify
    assert "^[[:space:]]*[0-9]+ failed([[:space:]]|$)" in verify
    assert "^[[:space:]]*(Test Files|Tests)[[:space:]].*failed" not in verify


def test_practice_screenshot_results_are_bound_to_the_current_source_tree() -> None:
    helper = REPO_ROOT / "test/scenarios/_shared/scripts/capture-source-fingerprint.py"
    assert helper.is_file()
    helper_text = helper.read_text(encoding="utf-8")
    for field in (
        "git_head",
        "git_dirty",
        "git_status_sha256",
        "source_sha256",
        "source_paths",
    ):
        assert field in helper_text

    scenarios = REPO_ROOT / "test/scenarios/e2e"
    for scenario in (
        "p0-044-practice-text-loop-assisted-happy-path",
        "p0-046-practice-text-loop-failure-and-recovery",
    ):
        trigger = (scenarios / scenario / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (scenarios / scenario / "scripts/verify.sh").read_text(encoding="utf-8")
        assert "capture-source-fingerprint.py" in trigger
        assert "source-fingerprint.json" in trigger
        assert "source-fingerprint.verify.json" in verify
        assert "source fingerprint changed after screenshot capture" in verify
        assert '"source_fingerprint"' in verify


def test_practice_phase10_uses_one_shared_source_manifest() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    manifest_path = scenarios / "practice-source-fingerprint-paths.json"
    manifest = manifest_path.read_text(encoding="utf-8")
    manifest_payload = json.loads(manifest)
    assert manifest_payload["schema_version"] == "practice-source-paths.v1"
    source_paths = manifest_payload["source_paths"]
    assert source_paths == sorted(set(source_paths))
    assert all((REPO_ROOT / path).exists() for path in source_paths)
    required_paths = (
        "test/scenarios/e2e/practice-source-fingerprint-paths.json",
        "docs/ui-design/module-practice-review.md",
        "ui-design/ui-design-contract.test.mjs",
        "ui-design/src/screen-practice.jsx",
        "ui-design/src/primitives.jsx",
        "ui-design/src/data.jsx",
        "ui-design/src/app.jsx",
        "package.json",
        "pnpm-lock.yaml",
        "frontend/package.json",
        "frontend/playwright.config.ts",
        "frontend/vite.config.ts",
        "frontend/scripts/serve-pixel-parity.mjs",
        "frontend/scripts/spaFallback.mjs",
        "frontend/src/app/screens/practice",
        "frontend/src/app/i18n",
        "frontend/src/app/routeStore.ts",
        "frontend/src/app/routeUrl.ts",
        "frontend/src/app/routes.ts",
        "frontend/src/api/generated/client.ts",
        "frontend/src/api/generated/types.ts",
        "frontend/tests/pixel-parity/report-parity-helpers.ts",
        "openapi/fixtures/PracticeSessions",
        "openapi/openapi.yaml",
        "openapi/templates/ts/client.tmpl",
        "backend/internal/practice",
        "backend/internal/api/practice",
        "backend/internal/store/practice",
        "migrations/000001_create_baseline.up.sql",
        "migrations/000001_create_baseline.down.sql",
        "frontend/tests/pixel-parity/practice.spec.ts",
        "test/scenarios/_shared/scripts/capture-source-fingerprint.py",
        "test/scenarios/_shared/scripts/frontend-real-backend-gate.sh",
        "test/scenarios/_shared/scripts/frontend-real-backend-verify.sh",
        "test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path",
        "test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery",
    )
    for path in required_paths:
        assert f'"{path}"' in manifest

    for scenario in (
        "p0-044-practice-text-loop-assisted-happy-path",
        "p0-046-practice-text-loop-failure-and-recovery",
    ):
        root = scenarios / scenario
        trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")
        assert 'SOURCE_MANIFEST="$ROOT/test/scenarios/e2e/practice-source-fingerprint-paths.json"' in trigger
        assert '--source-paths-from "$SOURCE_MANIFEST"' in trigger
        assert "--path " not in trigger
        assert '--source-paths-from "$SOURCE_FINGERPRINT"' in verify
        assert 'cmp -s "$SOURCE_FINGERPRINT" "$VERIFY_FINGERPRINT"' in verify
        assert "source fingerprint changed after screenshot capture" in verify


def test_practice_phase10_playwright_uses_current_dist_and_four_state_prototype_parity() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    spec = (
        REPO_ROOT / "frontend/tests/pixel-parity/practice.spec.ts"
    ).read_text(encoding="utf-8")

    assert 'from "./report-parity-helpers"' in spec
    assert "expectSurfaceParity" in spec
    assert "expectPixelParity" in spec
    assert "normalizedText" in spec
    assert spec.count("expectSurfaceParity(") >= 4
    assert spec.count("expectPixelParity(") >= 4
    assert spec.count("expectPracticeStateCopyParity(") >= 5
    assert 'page.goto(`/ui-design/' in spec or 'page.goto("/ui-design/' in spec
    for state in (
        "immediate-pending",
        "persisted-pending",
        "retryable-failed",
        "terminal-failed",
    ):
        assert state in spec

    for scenario in (
        "p0-044-practice-text-loop-assisted-happy-path",
        "p0-046-practice-text-loop-failure-and-recovery",
    ):
        trigger = (scenarios / scenario / "scripts/trigger.sh").read_text(
            encoding="utf-8"
        )
        assert "CI=1 pnpm exec playwright test" in trigger


def test_practice_four_state_parity_covers_complete_dom_a11y_geometry_and_pixels() -> None:
    spec = (
        REPO_ROOT / "frontend/tests/pixel-parity/practice.spec.ts"
    ).read_text(encoding="utf-8")

    assert "expectFullPagePixelParity" in spec
    assert spec.count("expectFullPagePixelParity(") >= 4
    assert spec.count("expectPracticeDomA11yParity(") >= 5
    assert spec.count("expectPracticeCoreSurfaceParity(") >= 5
    for snapshot_field in (
        "role",
        "accessibleName",
        "ariaLive",
        "ariaDescribedBy",
        "disabled",
        "placeholder",
        "text",
        "dataState",
    ):
        assert snapshot_field in spec
    for selector in (
        "practice-screen",
        "practice-transcript",
        "practice-input",
        "practice-input-textarea",
        "practice-input-send",
        "practice-finish-cta",
        "practice-finish-disabled-reason",
        "practice-interviewer-thinking",
        "practice-message-retry",
        "practice-terminal-recovery",
    ):
        assert selector in spec
    assert "absoluteSurfaceSnapshot" in spec
    absolute_snapshot = spec[
        spec.index("async function absoluteSurfaceSnapshot"):
        spec.index("async function expectPracticeCoreSurfaceParity")
    ]
    assert "relativeTo" not in absolute_snapshot
    assert "PRACTICE_ROOT" not in absolute_snapshot
    assert "maxChangedRatio" not in spec


def test_practice_parity_keeps_raf_live_and_snapshots_the_same_retry_draft() -> None:
    spec = (
        REPO_ROOT / "frontend/tests/pixel-parity/practice.spec.ts"
    ).read_text(encoding="utf-8")

    assert "settleVisualSurface" in spec
    assert "pauseDeterministicClock" not in spec
    assert ".clock.pauseAt" not in spec
    assert '"value" in element' in spec
    prototype_fill = (
        'await prototype.getByTestId("practice-input-textarea").fill(nextDraft)'
    )
    retry_snapshot = (
        'expectPracticeDomA11yParity(page, prototype, "retryable-failed")'
    )
    assert prototype_fill in spec
    assert spec.index(prototype_fill) < spec.index(retry_snapshot)


def test_practice_persisted_pending_evidence_is_reload_bound_and_zero_post() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    spec = (
        REPO_ROOT / "frontend/tests/pixel-parity/practice.spec.ts"
    ).read_text(encoding="utf-8")
    trigger = (
        scenarios
        / "p0-044-practice-text-loop-assisted-happy-path/scripts/trigger.sh"
    ).read_text(encoding="utf-8")
    verify = (
        scenarios
        / "p0-044-practice-text-loop-assisted-happy-path/scripts/verify.sh"
    ).read_text(encoding="utf-8")
    exact_unit_title = (
        "rehydrates a server pending row, shows thinking, polls server truth, "
        "and never resends it"
    )
    exact_browser_title = (
        "reloads a persisted pending reply, keeps all actions locked, and sends zero POSTs"
    )

    assert exact_unit_title in verify
    assert exact_browser_title in spec
    assert exact_browser_title in trigger
    assert exact_browser_title in verify
    assert "await page.reload()" in spec
    assert "messagePostCount" in spec
    assert "expect(messagePostCount).toBe(0)" in spec
    marker = (
        "PRACTICE_PERSISTED_PENDING_PASS reload=true message_posts=0 "
        "lease_before_expiry=true lease_seconds=90"
    )
    assert marker in trigger
    assert marker in verify


def test_practice_p0046_cleanup_is_retryable_residual_checked_and_not_silenced() -> None:
    root = (
        REPO_ROOT
        / "test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery"
    )
    setup = (root / "scripts/setup.sh").read_text(encoding="utf-8")
    trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
    cleanup = (root / "scripts/cleanup.sh").read_text(encoding="utf-8")
    helper = (root / "scripts/isolated-postgres-cleanup.sh").read_text(
        encoding="utf-8"
    )
    verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")

    assert "isolated-database.env" in setup
    assert "run scripts/cleanup.sh first" in setup
    for source in (trigger, cleanup):
        assert "isolated-database.env" in source
        assert "practice_database_name_for_run_id" in source
        assert "practice_cleanup_isolated_database" in source
    assert "practice_write_database_state" in trigger
    assert "run_id=%s" in helper
    assert "isolated_database_name=%s" in helper
    assert "residual=%s" in helper
    assert "for attempt in 1 2 3" in helper
    assert "dropdb --if-exists --force" in helper
    assert "SELECT count(*) FROM pg_database" in helper
    assert "residual=0" in helper
    assert "residual=1" in helper
    assert "cleanup_isolated_database || true" not in trigger
    assert "cleanup_on_exit" in trigger
    assert 'exit "$status"' in trigger
    assert "PRACTICE_ISOLATED_POSTGRES_CLEANUP_RETRY" in helper
    assert 'grep -Fqx \'residual=0\' "$DATABASE_STATE"' in verify
    assert (
        "PRACTICE_ISOLATED_POSTGRES_CLEANUP_PASS residual=0" in verify
    )


def test_practice_phase10_scenarios_require_exact_runtime_and_fresh_evidence() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    p0044 = scenarios / "p0-044-practice-text-loop-assisted-happy-path"
    p0046 = scenarios / "p0-046-practice-text-loop-failure-and-recovery"
    p0044_trigger = (p0044 / "scripts/trigger.sh").read_text(encoding="utf-8")
    p0044_verify = (p0044 / "scripts/verify.sh").read_text(encoding="utf-8")
    p0046_trigger = (p0046 / "scripts/trigger.sh").read_text(encoding="utf-8")
    p0046_verify = (p0046 / "scripts/verify.sh").read_text(encoding="utf-8")

    assert "usePracticeSessionLoader.test.tsx" in p0044_trigger
    assert "TestSQLRepositoryGetSessionKeepsPendingBeforeLeaseBoundary" in p0044_trigger
    for marker in (
        "PRACTICE_IMMEDIATE_PENDING_PASS",
        "PRACTICE_PERSISTED_PENDING_PASS",
        "PRACTICE_EVIDENCE_FINGERPRINT_PASS",
    ):
        assert marker in p0044_trigger or marker in p0044_verify
        assert marker in p0044_verify

    integration_tests = (
        "TestIntegrationPracticeReplyStateRecovery",
        "TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce",
        "TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce",
        "TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration",
        "TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery",
    )
    for test_name in integration_tests:
        assert test_name in p0046_trigger
        assert f"--- PASS: {test_name}" in p0046_verify

    assert "usePracticeSessionLoader.test.tsx" in p0046_trigger
    for test_title in (
        "aborts the POST exactly at 95,000 ms",
        "lets a later-started timeout reconcile win when an older loader refresh resolves first",
        "lets a later-started loader refresh win when the older timeout reconcile resolves first",
        "timeout reconciliation ends in missing-id",
        "timeout reconciliation ends in read-failure",
        "terminal state with one safe exact current-plan CTA",
    ):
        assert test_title in p0046_verify

    for marker in (
        "PRACTICE_PENDING_LEASE_RECOVERY_PASS",
        "PRACTICE_STALE_GENERATION_FENCED_PASS",
        "PRACTICE_CONCURRENT_RESERVATION_PASS",
        "PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS",
        "PRACTICE_TERMINAL_PLAN_RECOVERY_PASS",
        "PRACTICE_EVIDENCE_FINGERPRINT_PASS",
    ):
        assert marker in p0046_trigger or marker in p0046_verify
        assert marker in p0046_verify

    for setup in (p0044 / "scripts/setup.sh", p0046 / "scripts/setup.sh"):
        assert "setup_epoch=" in setup.read_text(encoding="utf-8")
    for verify in (p0044_verify, p0046_verify):
        for evidence_field in (
            '"sha256"',
            '"css_viewport"',
            '"device_scale_factor"',
            '"png_size"',
        ):
            assert evidence_field in verify
        assert "hashlib.sha256(data).hexdigest()" in verify
        assert "st_mtime" in verify
        assert "stale evidence artifact" in verify
        assert 'grep -Fq "RUN_ID=$run_id"' in verify
        assert "no tests to run" in verify
        assert "--- FAIL:" in verify
        assert "metadata_path" in verify
        assert 'metadata["css_viewport"]' in verify
        assert 'metadata["device_scale_factor"]' in verify


def test_practice_terminal_playwright_case_closes_parity_and_exact_route() -> None:
    spec = (
        REPO_ROOT / "frontend/tests/pixel-parity/practice.spec.ts"
    ).read_text(encoding="utf-8")

    for required in (
        'getByTestId("practice-terminal-recovery-cta")',
        'getByTestId("practice-error-state")',
        'getByTestId("practice-message-retry")',
        'getByTestId("practice-input-textarea")',
        "backgroundColor",
        "borderColor",
        "letterSpacing",
        "transitionDuration",
        "getBoundingClientRect",
        "document.documentElement.scrollWidth",
        "window.devicePixelRatio",
        ".metadata.json",
        'url.pathname).toBe("/parse")',
        'url.searchParams.get("targetJobId")',
        'url.searchParams.has("planId")',
        'url.pathname).not.toBe("/workspace")',
    ):
        assert required in spec
    assert spec.index('attachStateScreenshot(page, testInfo, "practice-terminal-failed")') < spec.index("await cta.click()")


def test_home_and_parse_scenarios_persist_source_bound_viewport_screenshots() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    required = {
        "p0-014-home-default-render": (
            "home-formal-viewport-desktop.png",
            "home-formal-viewport-mobile.png",
        ),
        "p0-015-jd-import-and-parse": (
            "home-formal-viewport-desktop.png",
            "home-formal-viewport-mobile.png",
            "parse-loading-formal-viewport-desktop.png",
            "parse-loading-formal-viewport-mobile.png",
        ),
    }
    for scenario, screenshot_names in required.items():
        root = scenarios / scenario
        setup = (root / "scripts/setup.sh").read_text(encoding="utf-8")
        trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")
        assert "run_id=" in setup
        assert "capture-source-fingerprint.py" in trigger
        assert "source-fingerprint.json" in trigger
        assert "--output=" in trigger
        assert "source-fingerprint.verify.json" in verify
        assert "source fingerprint changed after screenshot capture" in verify
        assert '"source_fingerprint"' in verify
        assert '"screenshots"' in verify
        assert "struct.unpack" in verify
        for screenshot_name in screenshot_names:
            assert screenshot_name in trigger
            assert screenshot_name in verify


def test_p0047_verifier_owns_redacted_completion_evidence_artifact() -> None:
    scenario = (
        REPO_ROOT
        / "test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
    )
    setup = (scenario / "scripts/setup.sh").read_text(encoding="utf-8")
    trigger = (scenario / "scripts/trigger.sh").read_text(encoding="utf-8")
    verify = (scenario / "scripts/verify.sh").read_text(encoding="utf-8")
    cleanup = (scenario / "scripts/cleanup.sh").read_text(encoding="utf-8")

    assert "run_correlation_id" in setup
    assert "TestIntegrationE2EP0047RejectsZeroAnswerCompletion" in trigger
    assert "practice-completion-evidence.v1" in verify
    assert "completion-backend-evidence.json" in verify
    for key in (
        "schemaVersion",
        "scenarioId",
        "command",
        "tests",
        "markers",
        "database",
        "result",
    ):
        assert key in verify
    for marker in (
        "ZERO_ANSWER_COMPLETION_REJECTED_PASS",
        "REPORT_CONTEXT_SNAPSHOT_PASS",
        "REPORT_CONTEXT_REPLAY_PASS",
        "concurrent_mutation_blocked=true",
        "snapshot_replay_equal=true",
        "mismatch_side_effect_count=0",
    ):
        assert marker in verify
    assert "PracticeScreen.test.tsx" in trigger
    assert "TestE2EP0047RejectsZeroAnswerCompletion" in trigger
    assert "*.log" in cleanup
    assert "completion-backend-evidence.json" not in cleanup


def test_report_scenarios_require_current_backend_evidence_contract() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    required = {
        "p0-056-generating-to-report-happy-path": {
            "test": "TestE2EP0056ReportBackendEvidence",
            "scenario_id": "E2E.P0.056",
            "evidence_schema": "report-backend-evidence.v1",
            "markers": (
                "REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS",
                "REPORT_DIRECT_READY_PASS",
                "REPORT_FROZEN_CONTEXT_READ_PASS",
                "REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS",
            ),
        },
        "p0-058-report-failure-and-missing-session": {
            "test": "TestE2EP0058ReportFailureBackendEvidence",
            "scenario_id": "E2E.P0.058",
            "evidence_schema": "report-backend-evidence.v3",
            "markers": (
                "REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS",
                "REPORT_CONTEXT_TOO_LARGE_PASS",
                "REPORT_OUTPUT_RETRY_PASS",
                "REPORT_FOUR_INVALID_FAIL_CLOSED_PASS",
                "REPORT_ACTION_RETRY_RESET_PASS",
                "REPORT_RETRY_LAYER_SEPARATION_PASS",
            ),
        },
    }

    for scenario, contract in required.items():
        root = scenarios / scenario
        setup = (root / "scripts/setup.sh").read_text(encoding="utf-8")
        trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")
        cleanup = (root / "scripts/cleanup.sh").read_text(encoding="utf-8")

        assert "completion-backend-evidence.json" in setup
        assert "practice-completion-evidence.v1" in setup
        for owner_marker in (
            "ZERO_ANSWER_COMPLETION_REJECTED_PASS",
            "REPORT_CONTEXT_SNAPSHOT_PASS",
            "REPORT_CONTEXT_REPLAY_PASS",
        ):
            assert owner_marker in setup

        assert "PRACTICE_COMPLETION_EVIDENCE_PATH" in trigger
        assert "./internal/review ./internal/store/review ./internal/api/reports" in trigger
        assert f"-run '^{contract['test']}$'" in trigger
        assert "-count=1 -v" in trigger

        assert f"=== RUN   {contract['test']}" in verify
        assert f"--- PASS: {contract['test']}" in verify
        for marker in contract["markers"]:
            assert marker in verify
        for field in (
            "schemaVersion",
            "scenarioId",
            "command",
            "tests",
            "consumedOwnerEvidence",
            "markers",
            "database",
            "result",
        ):
            assert field in verify
        assert contract["evidence_schema"] in verify
        assert contract["scenario_id"] in verify
        assert "backend-evidence.json" in verify
        assert "keys ==" in verify
        assert "--- FAIL:" in verify
        assert "no tests to run" in verify
        assert "raw_(cookie|jd|resume|transcript|prompt|output)" in verify

        if scenario == "p0-058-report-failure-and-missing-session":
            for runtime_field in (
                "runtime",
                "firstActionCallCount",
                "secondActionInitialAttempt",
                "retryStateDestroyedAfterAction",
                "actionRetryScheduleSeconds",
                "asyncAttemptsAffectProductAttempt",
            ):
                assert runtime_field in verify

        assert 'ARTIFACT="$OUT/backend-evidence.json"' not in trigger
        assert 'ARTIFACT="$OUT/backend-evidence.json"' not in cleanup
        assert "*.log" in cleanup


def test_p0059_requires_semantic_geometry_and_pixel_difference_evidence() -> None:
    root = (
        REPO_ROOT
        / "test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative"
    )
    readme = (root / "README.md").read_text(encoding="utf-8")
    seed = (root / "data/seed-input.md").read_text(encoding="utf-8")
    expected = (root / "data/expected-outcome.md").read_text(encoding="utf-8")
    trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
    verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")
    contract = "\n".join((readme, seed, expected, trigger, verify))

    for required in (
        "DOM/style/bbox",
        "pixelmatch",
        "changed-pixel ratio",
        "0.5%",
        "1440",
        "390",
        "formal/prototype/diff",
        "threshold 0.1",
        "tests/pixel-parity/reports.spec.ts",
        "tests/pixel-parity/generating.spec.ts",
        "tests/pixel-parity/report.spec.ts",
    ):
        assert required in contract

    stale = contract.lower()
    for forbidden in (
        "non-empty in-memory screenshot",
        "non-empty screenshot",
        "missing-session",
        "mobile-overflow",
    ):
        assert forbidden not in stale

    assert (
        "E2E.P0.059: running ReportsScreen, Report, and Generating Playwright pixel parity"
        in trigger
    )
    assert "E2E.P0.059: Playwright pixel parity complete" in trigger
    assert "Playwright pass marker missing" in verify
    assert "--- FAIL:" in verify
    assert "no tests to run" in verify


def test_resume_runtime_negative_gate_is_shared_and_ignores_tests() -> None:
    mode_pattern = "(tailor|mode).*(inline|rewrite|mirror)|(inline|rewrite|mirror).*(tailor|mode)"
    module_pattern = "mistakes|growth|drill|inline-debrief-record"
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    invocation = '"$ROOT/test/scenarios/_shared/scripts/resume-runtime-negative-gate.sh"'
    consumers = []
    for scenario in (
        "p0-075-resume-update-flat-fields-and-ik",
        "p0-076-resume-duplicate-save-as-new",
        "p0-077-resume-tailor-async-dispatch-and-ready",
        "p0-078-resume-tailor-failure-and-retry",
        "p0-079-resume-rewrites-accept-only-save",
        "p0-080-resume-tailor-privacy-negative",
    ):
        consumers.append(scenarios / scenario / "scripts/verify.sh")
    consumers.append(
        scenarios
        / "p0-080-resume-tailor-privacy-negative"
        / "scripts/trigger.sh"
    )

    for consumer in consumers:
        text = consumer.read_text(encoding="utf-8")
        assert invocation in text
        assert mode_pattern not in text
        assert module_pattern not in text

    gate = (
        REPO_ROOT
        / "test/scenarios/_shared/scripts/resume-runtime-negative-gate.sh"
    ).read_text(encoding="utf-8")
    assert mode_pattern in gate
    assert module_pattern in gate
    assert "--glob '!**/*_test.go'" in gate
    assert not (
        REPO_ROOT
        / "test/scenarios/_shared/scripts/resume-mode-negative-gate.sh"
    ).exists()
