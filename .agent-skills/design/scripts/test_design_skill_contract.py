"""Contract checks for the /design skill and its canonical BDD docs."""

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
SKILL_PATH = REPO_ROOT / ".agent-skills" / "design" / "SKILL.md"
SPEC_TEMPLATE_PATH = REPO_ROOT / "docs" / "spec" / "TEMPLATES.md"
INIT_SPEC_TEMPLATE_PATH = REPO_ROOT / ".agent-skills" / "init-docs" / "templates" / "spec-templates.md"


def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def test_design_skill_requires_layer_numbering_context_for_bdd():
    text = _read(SKILL_PATH)

    assert "read `test/scenarios/README.md` plus the relevant layer `README.md` / `INDEX.md`" in text
    assert "scenario IDs that follow those conventions" in text
    assert "Each scenario uses a behavior-oriented scenario ID such as `E2E.P0.001` or `E2E.P1.003`" in text


def test_design_skill_prohibits_ac_style_bdd_gate_ids_for_new_docs():
    text = _read(SKILL_PATH)

    assert "Generating BDD-Gate checklist items with `AC-*` references for new documents" in text


def test_design_skill_prohibits_hard_coverage_gates_in_test_plans():
    text = _read(SKILL_PATH)

    assert "Do not generate hard coverage-percentage gates in acceptance criteria or checklist items." in text
    assert "observational background rather than a completion, commit, or phase-exit condition" in text


def test_plan_docs_use_scenario_ids_for_bdd_gate_examples():
    plan_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    assert "BDD-Gate: 验证 E2E.P0.001 通过" in plan_template
    assert "BDD-Gate: 验证 E2E.P0.001 通过" in init_template
    assert "BDD-Gate: 验证 AC-1, AC-2 通过" not in plan_template
    assert "BDD-Gate: 验证 AC-1, AC-2 通过" not in init_template


def test_design_skill_generates_bdd_plan_and_checklist_together():
    text = _read(SKILL_PATH)
    spec_template = _read(SPEC_TEMPLATE_PATH)

    assert "generate bdd-plan.md and bdd-checklist.md" in text
    assert "add `bddPlan` and `bddChecklist` to context.yaml" in text
    assert "bdd-checklist.md" in spec_template


def test_design_skill_requires_tdd_for_code_and_bdd_for_user_behavior():
    text = _read(SKILL_PATH)

    assert "Code plan requires TDD" in text
    assert "Feature plan requires BDD" in text
    assert "user-visible UI, API behavior, business workflow, or end-to-end flow" in text
    assert "BDD is not a discretionary optional artifact" in text


def test_design_skill_requires_explicit_coverage_matrix_for_plan_and_tests():
    text = _read(SKILL_PATH)

    assert "Step 3.5: Build the Coverage Matrix" in text
    assert "Primary path" in text
    assert "Failure / recovery path" in text
    assert "Boundary condition" in text
    assert "Cross-layer contract" in text
    assert "Privacy / security / observability" in text
    assert "Regression / legacy-negative" in text
    assert "Every non-docs checklist item must name its verification source" in text
    assert "Test plans must include a coverage matrix" in text
    assert "BDD scenario selection must cover the primary user journey plus the highest-risk alternate or failure/recovery journey" in text


def test_spec_templates_include_quality_gate_classification():
    plan_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    for text in (plan_template, init_template):
        assert "## 3 质量门禁分类" in text
        assert "Plan 类型" in text
        assert "TDD 策略" in text
        assert "BDD 策略" in text
        assert "替代验证 gate" in text


def test_spec_template_marks_acceptance_ids_as_descriptive_only():
    text = _read(SPEC_TEMPLATE_PATH)

    assert "`ID` 列是文档内的说明性编号" in text
    assert "它不是 BDD 场景编号" in text


def test_spec_templates_include_user_decision_section():
    spec_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    assert "## 3 用户决策 / 待确认事项" in spec_template
    assert "## 3 用户决策 / 待确认事项" in init_template
    assert "无待确认事项时可省略" in spec_template
    assert "无待确认事项时可省略" in init_template


def test_plan_templates_prohibit_hard_coverage_threshold_items():
    plan_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    assert "coverage >= N%" not in plan_template
    assert "覆盖率 ≥ N%" not in init_template
    assert "本计划列出的实现 / 测试项全部通过" in plan_template
    assert "本计划列出的实现 / 测试项全部通过" in init_template
