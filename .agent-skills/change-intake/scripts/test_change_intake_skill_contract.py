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
