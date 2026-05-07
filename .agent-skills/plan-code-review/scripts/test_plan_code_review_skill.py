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
