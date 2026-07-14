"""Tests for the /plan-code-review skill contract."""
from pathlib import Path


SKILL_PATH = Path(__file__).resolve().parent.parent / "SKILL.md"


def _skill_text() -> str:
    return SKILL_PATH.read_text(encoding="utf-8")


class TestPlanCodeReviewSkill:
    """Plan code review should own L2 review and /tdd remediation routing."""

    def test_requires_explicit_plan_name(self):
        text = _skill_text()
        assert "Plan name is mandatory" in text

    def test_uses_shared_validator(self):
        text = _skill_text()
        assert ".agent-skills/implement/shared/scripts/validate_context.py" in text

    def test_reads_docs_directly_without_precheck_script(self):
        text = _skill_text()
        assert ".agent-skills/implement/scripts/review_code_precheck.py" not in text
        assert "Do not add parser-only gates" in text
        assert "After validation, read the returned markdown files directly." in text

    def test_fix_is_routed_through_tdd_section(self):
        text = _skill_text()
        assert "/tdd --file {checklist-path} --section {phase-prefix}" in text
        assert "does not own plan lifecycle sync or retrospective" in text

    def test_target_level_findings_stay_preview_only_without_section_mapping(self):
        text = _skill_text()
        assert "target-level-only findings stay preview-only" in text
        assert "no concrete checklist-section mapping" in text
        assert "preview-only" in text

    def test_review_checks_tdd_and_bdd_evidence_for_completed_code_phases(self):
        text = _skill_text()
        assert "For completed code phases, verify actual test evidence" in text
        assert "For completed feature phases, verify BDD evidence" in text

    def test_rejects_no_op_go_test_run_gates(self):
        text = _skill_text()
        assert "`go test" in text
        assert "[no tests to run]" in text
        assert "go test -list" in text
        assert "executed at least one intended test" in text

    def test_review_rejects_code_test_wrappers_and_backend_mocks_as_e2e(self):
        text = _skill_text()
        assert "A domain Behavior ID may map directly to a code-level owner behavior test" in text
        assert "An `E2E.*` ID is valid only when" in text
        assert "already running frontend/backend through real HTTP API calls or browser UI" in text
        assert "E2E script runs `go test`, Vitest/npm test, pytest, lint" in text
        assert "route interception, or request mocking replaces the business backend" in text
        assert "do not require or create an E2E directory" in text

    def test_review_requires_root_make_test_for_frontend_backend_regression(self):
        text = _skill_text()
        assert "require a current repository-root `make test` result" in text
        assert "代码测试不得进入 `test/scenarios/e2e/`" in text
        assert "单 package/file PASS 也不是全量单元回归证据" in text
        assert "state the repository-root `make test`" in text
        assert "run repository-root `make test` as the aggregate unit-regression gate" in text

    def test_review_reconstructs_coverage_matrix_against_current_artifacts(self):
        text = _skill_text()
        assert "Reconstruct the expected coverage matrix" in text
        assert "Coverage rows to verify" in text
        assert "Primary path" in text
        assert "Failure / recovery path" in text
        assert "Boundary condition" in text
        assert "Cross-layer contract" in text
        assert "Privacy / security / observability" in text
        assert "Regression / non-current-negative" in text
        assert "`C-series`: coverage matrix proof" in text
        assert "`Coverage Matrix Evidence` section" in text

    def test_dev_infra_reviews_image_volume_runtime_contracts(self):
        text = _skill_text()
        assert "Docker Compose / dev-infra targets" in text
        assert "dependency image major version" in text
        assert "named volume" in text
        assert "entrypoint" in text
        assert "default UID" in text
        assert "persistent data layout" in text
        assert "clean volume" in text
        assert "stale-volume path" in text
        assert "never count automatic volume deletion" in text

    def test_local_integration_reviews_current_runner_boundaries(self):
        text = _skill_text()
        assert "Docker Compose external dependencies" in text
        assert "host-run app commands" in text
        assert "repo-tracked real API/UI scenario boundaries" in text
        assert "local scenario runner entrypoint" in text
        assert "Kind scenario target" not in text
        assert "Docker Compose vs Kind" not in text

    def test_runtime_lifecycle_reviews_reverse_audit_production_entrypoint(self):
        text = _skill_text()
        assert "reverse-audit the production" in text
        assert "worker, dispatcher, outbox" in text
        assert "`cmd/api`, `main`" in text
        assert "production-wiring test" in text
        assert "constructs, attaches/registers, starts, and shuts down" in text
        assert "internal package tests alone as insufficient" in text
        assert "state the production entrypoint audited" in text

    def test_scheduler_reviews_require_starvation_and_scan_cadence_evidence(self):
        text = _skill_text()
        assert "scheduler" in text
        assert "queue weights" in text
        assert "scan-cycle SLAs" in text
        assert "long-running handler" in text
        assert "lower-priority but user-visible job" in text
        assert "bucket ordering" in text
        assert "starvation closure" in text
        assert "long-running handler/starvation gate" in text

    def test_runtime_reviews_audit_migration_escape_hatches(self):
        text = _skill_text()
        assert "`AsyncJobFinalized`" in text
        assert "`AlreadyHandled`" in text
        assert "`SkipFinalize`" in text
        assert "enumerate every handler/service/store path" in text
        assert "owner\n     kernel/shared retry/backoff/finalize policy" in text
        assert "deterministic clock evidence" in text
        assert "audited setters/observers" in text

    def test_frontend_first_handoff_reviews_require_real_backend_sweep(self):
        text = _skill_text()
        assert "frontend-first plans" in text
        assert "reverse-audit the adjacent backend owner" in text
        assert "`VITE_EI_API_MODE=real` generated-client gate" in text
        assert "`credentials: \"include\"`" in text
        assert "absence of fixture `Prefer` headers" in text
        assert "side-effect\n     `Idempotency-Key`" in text
        assert "drive the running frontend/backend\n     through real HTTP/UI" in text
        assert "must not wrap the focused frontend test" in text
        assert "sweep sibling/completed\n     plans in the same subspec" in text
        assert "adjacent backend owner evidence" in text
