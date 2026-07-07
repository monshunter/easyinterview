"""Contract checks for the /bug-report skill instructions."""

from pathlib import Path


SKILL_PATH = Path(__file__).resolve().parent.parent / "SKILL.md"


def _skill_text() -> str:
    return SKILL_PATH.read_text(encoding="utf-8")


# --- Phase 5.2: no inline module enumeration ---


def _step3_text() -> str:
    """Extract Step 3 section from SKILL.md."""
    text = _skill_text()
    start = text.find("### Step 3:")
    assert start != -1, "Step 3 heading not found"
    end = text.find("### Step 4:", start)
    assert end != -1, "Step 4 heading not found"
    return text[start:end]


def _step5_text() -> str:
    """Extract Step 5 section from SKILL.md."""
    text = _skill_text()
    start = text.find("### Step 5:")
    assert start != -1, "Step 5 heading not found"
    end = text.find("### Step 6:", start)
    assert end != -1, "Step 6 heading not found"
    return text[start:end]


MODULE_ENUM_VALUES = [
    "workspace", "practice", "review", "materials", "debrief",
    "eval", "frontend", "platform", "schema", "test",
]


def test_step3_no_inline_module_enum():
    """Step 3 must not hardcode the module enum list (D-10, C-11)."""
    step3 = _step3_text()
    # The full inline enum pattern: "must be enum: workspace / practice / ..."
    assert "must be enum:" not in step3, "Step 3 still contains inline enum declaration"
    # The validation gate pattern with all values listed
    enum_line = " / ".join(MODULE_ENUM_VALUES)
    assert enum_line not in step3, "Step 3 still lists all enum values inline"


def test_step3_references_bugs_readme():
    """Step 3 must reference docs/bugs/README.md for module enum (D-10)."""
    step3 = _step3_text()
    assert "docs/bugs/README.md" in step3, \
        "Step 3 must reference docs/bugs/README.md for module enum"


def test_step5_no_inline_module_mapping():
    """Step 5 must not hardcode module-to-table mapping (D-10, C-11)."""
    step5 = _step5_text()
    # Should not list individual mapping entries like "`workspace` -> Workspace"
    inline_count = sum(1 for v in MODULE_ENUM_VALUES if f"`{v}`" in step5)
    assert inline_count == 0, \
        f"Step 5 still hardcodes {inline_count} module mapping entries"


# --- Phase 5.3: no non-current skill references ---


def test_no_non_current_test_investigate_reference():
    """SKILL.md must not reference non-current /test-investigate (C-13, D-14)."""
    text = _skill_text()
    assert "/test-investigate" not in text, \
        "SKILL.md still references non-current /test-investigate"
