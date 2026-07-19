"""Contract checks for the /change-intake skill instructions."""

from pathlib import Path


SKILL_PATH = Path(__file__).resolve().parent.parent / "SKILL.md"


def test_skill_routes_completed_plan_to_in_place_revision():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert "`completed`: revise the original spec / plan / checklist in place before coding" in text
    assert "Do not create sibling follow-up / bugfix docs for same-subject revisions by default." in text


def test_skill_mentions_matcher_script_and_post_pass_reconcile():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert ".agent-skills/change-intake/scripts/match_change_context.py" in text
    assert "invoke `/retrospective --this` before final close-out" in text


def test_skill_prioritizes_exact_owner_evidence_and_bounds_status_tiebreaks():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert (
        "Exact scenario README Owner evidence outranks generic API / route keyword overlap."
        in text
    )
    assert "Active / draft status is only a bounded tie-breaker" in text


def test_skill_treats_low_confidence_as_repo_search_trigger():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert "`low` confidence: treat candidates as hypotheses and verify them with live repo search" in text
    assert "Only ask the user after live evidence remains ambiguous" in text


def test_skill_supports_current_owner_without_context_manifest():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert "If `contextPath` is null" in text
    assert "read the matcher-provided `plan` and `spec` paths directly" in text
    assert "do not invent a checklist, target, or temporary context manifest" in text


def test_skill_requires_branch_guard_before_mutation():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert "## Branch Guard Before Mutation" in text
    assert "git status --short --branch" in text
    assert "Never revise spec / plan / checklist on the default parent branch." in text
    assert (
        "If dirty changes already came from the current session while still on the"
        in text
    )


def test_skill_rejects_tool_name_branch_prefixes():
    text = SKILL_PATH.read_text(encoding="utf-8")
    assert "The branch prefix must" in text
    assert "`fix/`, `docs/`, `design/`, or" in text
    assert "`spec-design/`" in text
    assert "`codex/`, `claude/`, `gemini/`, `agent/`" in text
    assert "rename it to the semantic repository prefix before" in text
