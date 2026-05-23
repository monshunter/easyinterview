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

    def test_reviews_scenario_wrappers_as_process_success_evidence(self):
        text = _skill_text()
        assert "Treat scenario wrapper scripts as evidence artifacts" in text
        assert "Do not stop after reading the Go test body" in text
        assert "preserves the real test process exit status" in text
        assert "`go test | tee`" in text
        assert "`--- PASS`" in text
        assert "package-level `ok`" in text
        assert "reject `--- FAIL`, package `FAIL`, and `no tests to run`" in text
        assert "merely grepping a test name or package path" in text

    def test_review_reconstructs_coverage_matrix_against_current_artifacts(self):
        text = _skill_text()
        assert "Reconstruct the expected coverage matrix" in text
        assert "Coverage rows to verify" in text
        assert "Primary path" in text
        assert "Failure / recovery path" in text
        assert "Boundary condition" in text
        assert "Cross-layer contract" in text
        assert "Privacy / security / observability" in text
        assert "Regression / legacy-negative" in text
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
