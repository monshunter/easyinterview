"""Contract checks for the /tdd skill instructions."""

from pathlib import Path


SKILL_PATH = Path(__file__).resolve().parent.parent / "SKILL.md"


def _skill_text() -> str:
    return SKILL_PATH.read_text(encoding="utf-8")


def test_skill_declares_phase_commit_argument():
    text = _skill_text()

    assert "/tdd --file <checklist> --phase-commit <plan-name> --references f1,f2,..." in text
    assert "`--phase-commit <plan-name>`" in text
    assert "enable phase-boundary auto-commit via `/work-journal --auto`" in text


def test_skill_defines_step_95_phase_commit_gate():
    text = _skill_text()

    assert "### Step 9.5: Phase Commit Gate" in text
    assert "call `/work-journal --auto --plan <name> --phase <heading>` on the feature branch" in text
    assert "keep the current feature branch checked out; do not checkout, merge, or ff-only merge the base branch automatically" in text
    assert "Base branch integration is a separate explicit operation owned by the user" in text
    assert "Only after Step 9.5 succeeds may `/tdd` continue to the next implementation phase." in text


def test_section_mode_can_still_trigger_single_phase_commit():
    text = _skill_text()

    assert "When `--section` and `--phase-commit` are both present, trigger Step 9.5 exactly once when that section's last item and mapped test items are complete." in text


def test_phase_commit_failures_stop_tdd_run():
    text = _skill_text()

    assert "If `/work-journal --auto` or drift repair fails, stop the current `/tdd` run immediately." in text
    assert "Preserve the current branch and working tree for retry or manual intervention." in text
    assert "Do not advance to the next phase or Step 10 while the phase-commit failure remains unresolved." in text


def test_resume_path_must_not_bypass_tdd_owner_chain():
    text = _skill_text()

    assert "When resuming an in-flight plan, do not continue implementation outside `/tdd`." in text
    assert "Re-enter through `/implement` or the current `/tdd` owner path so Step 9.5 remains active." in text


def test_bdd_gate_prefers_behavior_scenario_ids():
    text = _skill_text()

    assert "New canonical plans key BDD verification by behavior-oriented" in text
    assert "`E2E.P0.001` / `E2E.P0.004`" in text
    assert "Parse scenario references" in text
    assert "Extract scenario identifiers from the item text" in text


def test_bdd_gate_keeps_legacy_ac_mapping_as_compatibility_path():
    text = _skill_text()

    assert "older `AC-*` style gate" in text
    assert "legacy compatibility path" in text
    assert "historical AC mapping in `bdd-plan.md`, `bdd-test-plan.md`, or the spec §验收标准 table" in text


def test_bdd_gate_requires_bdd_checklist_completion_before_gate():
    text = _skill_text()

    assert "basename `bdd-checklist.md`" in text
    assert "Every asset and execution item for those scenarios must already be checked" in text
    assert "Marking a BDD-Gate item complete while related `bdd-checklist.md`" in text


def test_hard_coverage_gate_must_stop_tdd_and_route_back_to_plan_fix():
    text = _skill_text()

    assert "### Step 3C: Hard Coverage Gate detection" in text
    assert "do **not** try to satisfy it by inventing extra tests" in text
    assert "route back to `/plan-review --fix`" in text
    assert "Hard coverage-percentage checklist gates count as a plan/checklist mismatch" in text


def test_tdd_skill_declares_any_code_logic_requires_tdd():
    text = _skill_text()

    assert "any code logic implementation" in text
    assert "front-end, back-end, tooling, migration, codegen, or test helper logic" in text
    assert "Every checklist item has at least one corresponding test assertion." in text
