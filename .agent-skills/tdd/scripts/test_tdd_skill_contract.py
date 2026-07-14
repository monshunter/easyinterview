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


def test_bdd_gate_supports_domain_behavior_and_real_e2e_ids():
    text = _skill_text()

    assert "domain\n   Behavior ID such as `BDD.AUTH.001`" in text
    assert "real E2E ID such as `E2E.P0.001` only when" in text
    assert "Parse behavior references" in text
    assert "Extract domain Behavior IDs or real E2E IDs" in text
    assert "Do not mix a domain Behavior ID and a real E2E ID in one `BDD-Gate:`" in text
    assert "record the latter as a separate `E2E-HANDOFF:`" in text
    mixed_layer_example = "BDD-Gate: 验证 BDD.AUTH.001, "
    mixed_layer_example += "E2E.P0.101 通过"
    assert mixed_layer_example not in text


def test_e2e_handoff_is_not_an_execution_gate():
    text = _skill_text()

    assert "An `E2E-HANDOFF:` item is a static ownership/reference check" in text
    assert "must not run the scenario, change `Ready` to PASS, or activate Step 5B" in text
    assert "Only an explicit real-E2E `BDD-Gate:` in the E2E suite owner" in text


def test_bdd_gate_keeps_ac_style_mapping_as_compatibility_input():
    text = _skill_text()

    assert "AC-style compatibility gate" in text
    assert "compatibility input" in text
    assert "AC mapping in `bdd-plan.md`, `bdd-test-plan.md`, or the spec §验收标准 table" in text


def test_bdd_gate_requires_prerequisites_then_records_execution_evidence():
    text = _skill_text()

    assert "basename `bdd-checklist.md`" in text
    assert "behavior definition, evidence-layer choice, and required test/scenario assets" in text
    assert "do not require them to be pre-checked before running their own verification" in text
    assert "Complete the matching BDD checklist execution/evidence items" in text


def test_domain_behavior_tests_are_not_wrapped_as_e2e():
    text = _skill_text()

    assert "Focused Go/Vitest execution is development evidence" in text
    assert "must not be wrapped in `test/scenarios/e2e/`" in text
    assert "scripts do not run `go test`, Vitest/npm test, pytest, lint" in text
    assert "browser does not intercept/mock the business backend" in text


def test_frontend_backend_final_regression_uses_root_make_test():
    text = _skill_text()

    assert "run `make test` from the repository root before closing the phase" in text
    assert "authoritative whole frontend/backend unit-regression gate" in text
    assert "confirm a current repository-root `make test` PASS" in text
    assert "do not run `make test` from an E2E scenario" in text


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
