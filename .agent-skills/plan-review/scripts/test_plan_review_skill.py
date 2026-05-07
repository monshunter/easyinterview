"""Tests for the /plan-review skill contract."""
from pathlib import Path


SKILL_PATH = Path(__file__).resolve().parent.parent / "SKILL.md"


def _skill_text() -> str:
    return SKILL_PATH.read_text(encoding="utf-8")


class TestPlanReviewSkill:
    """Plan review should own L1 review and document fix behavior."""

    def test_uses_implement_shared_candidate_and_validator(self):
        text = _skill_text()
        assert ".agent-skills/implement/shared/scripts/list_context_candidates.py" in text
        assert ".agent-skills/implement/shared/scripts/validate_context.py" in text

    def test_reads_validated_docs_directly_without_parser_checker(self):
        text = _skill_text()
        assert ".agent-skills/implement/scripts/review_plan.py" not in text
        assert "reads the validated markdown" in text
        assert "Do not add python markdown-format validators" in text

    def test_fix_requires_explicit_plan_name(self):
        text = _skill_text()
        assert "requires an explicit plan name" in text

    def test_report_contract_is_present(self):
        text = _skill_text()
        assert "`Findings`" in text
        assert "`Dimension Coverage`" in text
        assert "`Strengths`" in text
        assert "`Optimization Opportunities`" in text

    def test_flags_hard_coverage_thresholds_as_document_findings(self):
        text = _skill_text()
        assert "`S-005`: test completion gates are execution-based" in text
        assert "raw code coverage percentages as completion, commit, or phase-exit criteria" in text

    def test_flags_missing_tdd_bdd_quality_gate_classification(self):
        text = _skill_text()
        assert "`S-006`: TDD/BDD quality gate classification" in text
        assert "Code plan requires TDD" in text
        assert "Feature plan requires BDD" in text
        assert "substitute verification gate" in text

    def test_reviews_coverage_matrix_for_edges_and_negative_scope(self):
        text = _skill_text()
        assert "Reconstruct the expected coverage matrix" in text
        assert "Primary path" in text
        assert "Failure / recovery path" in text
        assert "Boundary condition" in text
        assert "Cross-layer contract" in text
        assert "Privacy / security / observability" in text
        assert "Regression / legacy-negative" in text
        assert "`S-007`: coverage matrix adequacy" in text
        assert "`X-L1-Coverage`: coverage matrix clarity" in text
