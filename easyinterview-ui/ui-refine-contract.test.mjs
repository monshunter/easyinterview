import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import test from "node:test";

const readUiFile = (path) => readFileSync(new URL(path, import.meta.url), "utf8");
const readUiSources = () => readdirSync(new URL("./src/", import.meta.url))
  .filter((name) => name.endsWith(".jsx"))
  .map((name) => [name, readUiFile(`./src/${name}`)]);

test("workspace mock history is scoped to the active mock plan", () => {
  const workspace = readUiFile("./src/screen-workspace.jsx");
  assert.match(
    workspace,
    /const sessionHistory = getWorkspaceSessionHistory\(lang, job, currentRound\?\.name\);/,
  );

  const historyStart = workspace.indexOf("const getWorkspaceSessionHistory");
  const historyEnd = workspace.indexOf("const getWorkspacePlanOptions");
  const historySource = workspace.slice(historyStart, historyEnd);

  assert.notEqual(historyStart, -1);
  assert.notEqual(historyEnd, -1);
  assert.match(historySource, /getWorkspaceTargetLabel/);
  assert.doesNotMatch(historySource, /Lumen Labs · Frontend Platform Engineer/);
});

test("debrief context cards open local pickers instead of cross-page navigation", () => {
  const debrief = readUiFile("./src/screens-p1-depth.jsx");
  assert.doesNotMatch(debrief, /Inbox|收件箱/);
  assert.match(debrief, /DebriefContextPickerModal/);

  const stripStart = debrief.indexOf("const DebriefContextStrip");
  const stripEnd = debrief.indexOf("const GuidedDebriefRecord");
  const stripSource = debrief.slice(stripStart, stripEnd);

  assert.notEqual(stripStart, -1);
  assert.notEqual(stripEnd, -1);
  assert.match(stripSource, /onOpenPicker/);
  assert.doesNotMatch(stripSource, /nav\("workspace"|nav\("report"|nav\("resume_versions"/);
});

test("current UI source does not expose removed inbox wording", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /Inbox|收件箱/, `${name} contains removed inbox wording`);
  }
});
