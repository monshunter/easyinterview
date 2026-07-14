#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import re
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[4]
APP_ROOT = ROOT / "frontend/src/app"
SCREENS = APP_ROOT / "screens"


def production_sources(root: pathlib.Path, suffixes: tuple[str, ...]) -> list[pathlib.Path]:
    return [
        path
        for path in root.rglob("*")
        if path.is_file()
        and path.suffix in suffixes
        and ".test." not in path.name
        and ".spec." not in path.name
        and "__tests__" not in path.parts
    ]


class ParseReportsEntrySourceContractTest(unittest.TestCase):
    def test_reports_screen_is_the_only_production_screen_list_consumer(self) -> None:
        consumers = [
            path.relative_to(ROOT).as_posix()
            for path in production_sources(SCREENS, (".ts", ".tsx"))
            if "listTargetJobReports" in path.read_text(encoding="utf-8")
        ]
        self.assertEqual(
            consumers,
            ["frontend/src/app/screens/reports/ReportsScreen.tsx"],
        )

    def test_parse_has_one_page_entry_and_no_embedded_reports_contract(self) -> None:
        formal = read(ROOT / "frontend/src/app/screens/parse/ParseScreen.tsx")
        prototype = read(ROOT / "ui-design/src/screens-p0-complete.jsx")
        for source in (formal, prototype):
            self.assertEqual(source.count('data-testid="parse-reports-entry"'), 1)
            self.assertIn('nav("reports"', source.replace("navigate({ name: ", "nav("))
            self.assertNotIn('data-testid="parse-reports"', source)
            self.assertNotIn("parse-report-section", source)
            self.assertNotIn("parse-report-round-", source)
        self.assertNotIn("listTargetJobReports", formal)
        self.assertNotIn("section=reports", formal)

    def test_parse_route_drops_retired_section_and_topbar_has_no_reports_entry(self) -> None:
        route_url = read(APP_ROOT / "routeUrl.ts")
        parse_safe = re.search(
            r"const PARSE_SAFE = new Set\(\[(.*?)\]\);",
            route_url,
            flags=re.DOTALL,
        )
        self.assertIsNotNone(parse_safe)
        self.assertNotIn("section", parse_safe.group(1))

        topbar = read(APP_ROOT / "topbar/TopBar.tsx")
        self.assertNotIn("topbar-nav-reports", topbar)
        self.assertNotRegex(topbar, r"\breports\s*:\s*[\"']nav\.")

        production_text = "\n".join(
            path.read_text(encoding="utf-8")
            for path in production_sources(APP_ROOT, (".ts", ".tsx"))
        )
        self.assertNotIn("section=reports", production_text)


def read(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


if __name__ == "__main__":
    unittest.main()
