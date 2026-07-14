import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import test from "node:test";

const readUiFile = (path) => readFileSync(new URL(path, import.meta.url), "utf8");
const readUiSources = () => readdirSync(new URL("./src/", import.meta.url))
  .filter((name) => name.endsWith(".jsx"))
  .map((name) => [name, readUiFile(`./src/${name}`)]);
const readUiJsxTree = (directory = new URL("./", import.meta.url), prefix = "") =>
  readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const relative = `${prefix}${entry.name}`;
    const entryUrl = new URL(`./${relative}${entry.isDirectory() ? "/" : ""}`, import.meta.url);
    if (entry.isDirectory()) {
      return readUiJsxTree(entryUrl, `${relative}/`);
    }
    return entry.name.endsWith(".jsx") ? [[relative, readFileSync(entryUrl, "utf8")]] : [];
  });

test("prototype JSX sources do not duplicate whole files", () => {
  const sourceByHash = new Map();
  for (const [name, source] of readUiJsxTree()) {
    const hash = createHash("sha256").update(source).digest("hex");
    assert.equal(sourceByHash.get(hash), undefined, `${name} duplicates ${sourceByHash.get(hash)}`);
    sourceByHash.set(hash, name);
  }
});

test("prototype runner uses the repository Python 3 toolchain only", () => {
  const runner = readUiFile("./run.sh");
  const toolchainCheck = runner.indexOf("if ! command -v python3");
  const urlEncoding = runner.indexOf("url_encode() {");

  assert.ok(toolchainCheck >= 0 && toolchainCheck < urlEncoding);
  assert.match(runner, /SCRIPT_PATH="\$SCRIPT_DIR\/\$\(basename "\$0"\)"/);
  assert.match(runner, /sed -n '2,7p' "\$SCRIPT_PATH"/);
  assert.match(runner, /exec python3 -m http\.server/);
  assert.doesNotMatch(runner, /SimpleHTTPServer|npx --yes serve|elif command -v python/);
});

test("prototype primitive globals expose only current consumers", () => {
  const primitives = readUiFile("./src/primitives.jsx");

  assert.doesNotMatch(primitives, /const (?:Sparkline|KV) =/);
  assert.doesNotMatch(primitives, /\b(?:Sparkline|KV)\b/);
  assert.match(
    primitives,
    /Object\.assign\(window, \{ Icon, Tag, Btn, Card, SectionHeader, ReadinessDial \}\);/,
  );
});

test("TopBar mobile truth source wraps without horizontal viewport overflow", () => {
  const app = readUiFile("./src/app.jsx");
  const primitives = readUiFile("./src/primitives.jsx");

  for (const className of [
    "ei-shell-topbar",
    "ei-topbar-brand",
    "ei-topbar-brand-copy",
    "ei-topbar-nav",
    "ei-topbar-spacer",
    "ei-topbar-controls",
    "ei-topbar-lang-current",
    "ei-topbar-user",
  ]) {
    assert.match(app, new RegExp(`className=\\"${className}\\"`));
  }
  assert.match(primitives, /@media \(max-width: 720px\)[\s\S]*\.ei-shell-topbar\s*\{[\s\S]*flex-wrap:\s*wrap/);
  assert.match(primitives, /\.ei-shell-topbar\s*\{[\s\S]*overflow-x:\s*clip/);
  assert.match(primitives, /\.ei-topbar-nav\s*\{[\s\S]*order:\s*10[\s\S]*width:\s*100%[\s\S]*overflow-x:\s*auto/);
  assert.match(primitives, /@media \(max-width: 460px\)[\s\S]*\.ei-topbar-brand-copy\s*\{[\s\S]*display:\s*none/);
  assert.match(app, /className="ei-topbar-theme-menu"/);
  assert.match(
    primitives,
    /@media \(max-width: 720px\)[\s\S]*\.ei-topbar-theme-menu\s*\{[\s\S]*left:\s*0\s*!important;[\s\S]*right:\s*auto\s*!important;/,
  );
  assert.match(app, /params\.get\("signedIn"\) === "1"/);
});

test("workspace static prototype is a pure plan list with no unreachable detail branch", () => {
  const workspace = readUiFile("./src/screen-workspace.jsx");

  assert.match(workspace, /const WorkspaceScreen = \(\{ T, lang, nav \}\) => \{/);
  assert.match(workspace, /return <WorkspacePlanList T=\{T\} lang=\{lang\} nav=\{nav\} jobs=\{jobs\} \/>;/);
  assert.match(workspace, /Object\.assign\(window, \{ getWorkspaceResumeOptions \}\);/);
  assert.doesNotMatch(workspace, /^\s+updated:/m);

  for (const symbol of [
    "hasPlanContext",
    "WorkspaceEmptyState",
    "WorkspaceMissingResumeState",
    "createWorkspaceInterviewContext",
    "getWorkspaceRoundId",
    "getWorkspaceSessionHistory",
    "getWorkspaceTargetLabel",
    "getWorkspaceRoundLabel",
    "getWorkspacePlanOptions",
    "getWorkspaceJDSample",
    "ResumePickerModal",
    "PlanSwitcherModal",
    "WorkspaceInsightCard",
    "InterviewRoundRail",
    "BindingPill",
    "ReqBlock",
  ]) {
    assert.doesNotMatch(workspace, new RegExp(`\\b${symbol}\\b`), `${symbol} must stay absent from the pure list prototype`);
  }
});

test("D-22 keeps debrief and user profile outside current static screens", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.ok(!existsSync(new URL("./src/screens-p1-depth.jsx", import.meta.url)), "screens-p1-depth.jsx must stay absent");
  assert.ok(!existsSync(new URL("./src/screen-profile.jsx", import.meta.url)), "screen-profile.jsx must stay absent");
  assert.match(app, /debrief:\s*"home"/);
  assert.match(app, /debrief_full:\s*"home"/);
  assert.match(app, /profile:\s*"home"/);
  assert.doesNotMatch(app, /debrief:\s*</);
  assert.doesNotMatch(app, /debrief_full:\s*</);
  assert.doesNotMatch(app, /profile:\s*</);
  assert.doesNotMatch(canvas, /<DCSection id="profile"|<DCArtboard id="profile-[^"]+"|route="profile"/);
  assert.doesNotMatch(canvas, /<DCSection id="p1-depth"|<DCArtboard id="debrief"|<DCSection id="p1-voice-debrief"|route="debrief"/);
});

test("current UI source does not expose out-of-scope inbox wording", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /Inbox|收件箱/, `${name} contains out-of-scope inbox wording`);
  }
});

test("resume workshop is a flat list without version-tree concepts", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const ResumeListView = /);
  assert.match(resume, /const openResume = \(r\) => nav\("resume_versions", \{ resumeId: r\.id \}\);/);
  assert.doesNotMatch(resume, /ResumeTreeView|ResumeBranchFlow|ResumeBranchMap/);
  assert.doesNotMatch(resume, /"MASTER"|"TARGETED"|主版本|岗位定制|选为底稿|分叉/);
  assert.doesNotMatch(resume, /createMode === "guided"|轻量问答|guideSteps|guideAnswers/);
  assert.doesNotMatch(resume, /versionType|parentVersionId|originalId/);
});

test("resume workshop keeps rewrites and second-step save surfaces absent", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.doesNotMatch(resume, /const acceptBullet = \(id\) => \{/);
  assert.doesNotMatch(resume, /const RewriteSaveConfirmModal = /);
  assert.doesNotMatch(resume, /覆盖原简历|保存为新简历/);
  assert.doesNotMatch(resume, /onConfirm\(mode\)|saveRewriteResult/);
  assert.doesNotMatch(resume, /ResumePreviewConfirm|ResumeParseFlow/);
  assert.doesNotMatch(resume, /"rejected"|已拒绝|拒绝/);
  assert.doesNotMatch(resume, /保存人工改写|Save manual edit/);
});

test("resume workshop detail is read-only original content without source preview/export controls", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const ResumeDetailView = /);
  assert.match(resume, /const lines = Array\.isArray\(resume\.text\) && resume\.text\.length > 0/);
  assert.match(resume, /background: "#f6f3ee"/);
  assert.match(resume, /minHeight: 664, margin: "0 auto", background: "#ffffff"/);
  assert.doesNotMatch(resume, /fontSize: 28, fontWeight: 600, letterSpacing: "-0\.02em" \}\}>\{resume\.name\}/);
  assert.doesNotMatch(resume, /fontSize: 14, color: "#666", marginTop: 4 \}\}>\{resume\.summary\}/);
  assert.doesNotMatch(resume, /sourcePreviewOpen|OriginalResumePreviewModal|onPreviewOriginal/);
  assert.doesNotMatch(resume, /onExport|exportPdf|onCopy|copyText/);
  assert.doesNotMatch(resume, /navigator\.clipboard\.writeText/);
  assert.doesNotMatch(resume, /icon="download"/);
});

test("resume workshop create flow keeps upload/paste only and opens detail directly", () => {
  const app = readUiFile("./src/app.jsx");
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const \[createdResumes, setCreatedResumes\] = React\.useState\(\[\]\);/);
  assert.match(resume, /setCreatedResumes\(\(prev\) => \[\.\.\.prev, resume\]\);/);
  assert.match(resume, /setCreatedResumes\(\(prev\) => \[\.\.\.prev, resume\]\);\s+setFlow\("list"\);\s+nav\("resume_versions", \{ resumeId: resume\.id \}\);/);
  assert.match(resume, /nav\("resume_versions", \{ resumeId: resume\.id \}\);/);
  assert.match(resume, /setFlow\(params\.flow === "create" \? "create" : "list"\);/);
  assert.match(app, /<div key=\{route\.name \+ \(route\.params\.targetJobId \|\| route\.params\.jobId \|\| ""\)\}>/);
  assert.match(resume, /onCreateResume\(sourceLabel, createMode === "paste" \? resumeText : ""\)/);
  assert.match(resume, /onCreateResume\(f\.name, ""\)/);
  assert.match(resume, /\{ k: "upload", icon: "upload"/);
  assert.match(resume, /\{ k: "paste", icon: "file"/);
  assert.doesNotMatch(resume, /onConfirm=|resume-preview-confirm|resume-parse-flow/);
});

test("P0 context routes use InterviewContext instead of fixed tj-1 nav payloads", () => {
  const app = readUiFile("./src/app.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");

  assert.match(app, /const DEFAULT_INTERVIEW_CONTEXT = /);
  assert.match(app, /const createInterviewContext = /);
  assert.match(app, /const shouldCarryInterviewContext = /);
  assert.doesNotMatch(app, /params:\s*\{\s*jobId:\s*"tj-1"\s*\}/);

  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(
      source,
      /nav\("[a-z_]+",\s*\{[^}]*jobId:\s*"tj-1"/,
      `${name} still hard-codes tj-1 in a nav payload`,
    );
  }

  assert.match(workspace, /const openPlan = \(job\) => nav\("workspace", \{/);
  assert.match(workspace, /targetJobId: job\.id/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /resumeVersionId/, `${name} still uses the out-of-scope resumeVersionId context key`);
  }
});

test("P0 report is reportId-only and renders the direct semantic contract", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /const ReportMissingState = /);
  assert.match(report, /const ReportFailureState = /);
  assert.match(report, /if \(!params\.reportId/);
  assert.match(report, /const context = report\.context/);
  assert.match(report, /const dimensions = report\.dimensionAssessments/);
  assert.match(report, /const highlights = report\.highlights \|\| \[\]/);
  assert.match(report, /const actions = report\.nextActions/);
  assert.match(report, /report\.summary/);
  assert.match(report, /item\.label/);
  assert.match(report, /localizeDimensionStatus/);
  assert.match(report, /localizeConfidence/);
  assert.match(report, /localizeReadiness/);
  assert.doesNotMatch(report, /params\.(?:reportStatus|sessionId|targetJobId|resumeId|roundId|roundName)/);
  assert.doesNotMatch(report, /perQuestion|replayItems|activeQuestion|qId|\.score\b/);
  assert.doesNotMatch(report, /nav\("workspace",\s*\{\s*jobId:\s*"tj-1"\s*\}\)/);
});

test("P0 report keeps three metrics, four always-visible sections, and server-owned actions", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.equal((report.match(/<ReportMetric\b/g) || []).length, 3);
  for (const testId of ["report-dimensions", "report-highlights", "report-issues", "report-actions"]) {
    assert.match(report, new RegExp(`data-testid="${testId}"`));
  }
  assert.match(report, /const firstAction = actions\[0\]/);
  assert.match(report, /firstAction\?\.type === "retry_current_round"/);
  assert.match(report, /firstAction\?\.type === "next_round"/);
  assert.match(report, /report\.context\.hasNextRound/);
  assert.match(report, /sourceReportId: report\.id/);
  assert.doesNotMatch(report, /focusCompetencyCodes|evidenceGaps|retryFocusTurnIds/);
  assert.doesNotMatch(report, /ReportDetailSurface|QuestionsTab|题目回顾|role="tab"/);
});

test("report prototype keeps the two-action and 24-word / 64-code-point quality boundary", () => {
  const data = readUiFile("./src/data.jsx");
  const report = readUiFile("./src/screen-report.jsx");
  const start = data.indexOf("\n  report: {");
  const end = data.indexOf("\n  }\n};", start);
  const reportData = data.slice(start, end);
  const actionBlock = reportData.match(/nextActions:\s*\[([\s\S]*?)\],\s*retryFocusDimensionCodes/);

  assert.ok(start >= 0 && end > start && actionBlock, "report action fixture must be present");
  const labels = [...actionBlock[1].matchAll(/label:\s*"([^"]+)"/g)].map((match) => match[1]);
  assert.equal(labels.length, 2);
  for (const label of labels) assert.ok([...label].length <= 64, `over-limit zh-CN action: ${label}`);
  assert.match(report, /const ACTION_LABEL_WIRE_MAX_CODE_POINTS = 200/);
  assert.match(report, /const codePoints = \[\.\.\.value\]\.length/);
  assert.match(report, /codePoints > ACTION_LABEL_WIRE_MAX_CODE_POINTS/);
  assert.match(report, /split\(\/\\s\+\/u\)\.length <= 24/);
  assert.match(report, /codePoints <= 64/);
  assert.match(report, /actions\.length <= 2/);
});

test("report action rows preserve full labels without clipping or horizontal overflow", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /className="ei-report-action-row"/);
  assert.match(report, /className="ei-report-action-label"/);
  assert.match(report, /minWidth:\s*0[\s\S]*overflowWrap:\s*"anywhere"[\s\S]*wordBreak:\s*"normal"/);
  assert.doesNotMatch(report, /ei-report-action-label[^\n]*(?:textOverflow:\s*"ellipsis"|whiteSpace:\s*"nowrap"|overflow:\s*"hidden")/);
});

test("report capability rows keep long mobile labels readable", () => {
  const report = readUiFile("./src/screen-report.jsx");

  for (const className of [
    "ei-report-dimension-row",
    "ei-report-dimension-label",
    "ei-report-dimension-status",
  ]) {
    assert.match(report, new RegExp(`className=\\"${className}\\"`));
  }
  assert.match(report, /flexWrap:\s*"wrap"/);
  assert.match(report, /flex:\s*"1 1 160px"/);
  assert.match(report, /overflowWrap:\s*"break-word",\s*wordBreak:\s*"normal"/);
  assert.match(report, /flex:\s*"0 1 auto",\s*maxWidth:\s*"100%"/);
});

test("report CTA pair lives only in the header (D-19)", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.equal((report.match(/goal: "retry_current_round"/g) || []).length, 1);
  assert.equal((report.match(/goal: "next_round"/g) || []).length, 1);
  assert.doesNotMatch(report, /ReportDetailSurface|QuestionsTab|题目回顾/);
});

test("report generating prototype exposes only honest server-projected states and actions", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");
  const start = p0.indexOf("const ReportGeneratingScreen = ");
  const end = p0.indexOf("// #8 SETTINGS", start);
  const generating = p0.slice(start, end);

  assert.ok(start >= 0 && end > start);
  assert.match(generating, /const report = window\.EI_DATA\.reportGeneration/);
  assert.match(generating, /const status = report\.status/);
  assert.match(generating, /data-testid="generating-screen"/);
  assert.match(generating, /REPORT_CONTEXT_TOO_LARGE/);
  assert.match(generating, /continueCheck/);
  assert.match(generating, /data-testid="generating-back-button"/);
  assert.match(generating, /nav\("reports", \{ targetJobId: report\.targetJobId \}\)/);
  assert.match(generating, /nav\("workspace"\)/);
  assert.doesNotMatch(generating, /nav\("parse"|section:\s*"reports"/);
  assert.doesNotMatch(generating, /setTimeout|setInterval|\bpct\b|\bphase\b|liveSnippets|LIVE OBSERVATIONS|实时观察|Notify me|好了通知我|session records|会话记录/);
});

test("current UI source does not expose out-of-scope mistakes/growth/drill product surfaces", () => {
  const report = readUiFile("./src/screen-report.jsx");
  const settings = readUiFile("./src/screens-p0-complete.jsx");
  const data = readUiFile("./src/data.jsx");

  assert.doesNotMatch(report, /错题|openDrill|addToMistakes|addedToMistakes/);
  assert.doesNotMatch(settings, /Mistakes derived|派生的错题/);
  assert.doesNotMatch(data, /\bmistakes\s*:|mistakesTotal|mistakesResolved|\bgrowth\s*:/);
});

test("prototype experience copy describes direct current actions", () => {
  const data = readUiFile("./src/data.jsx");

  assert.doesNotMatch(data, /兼容层/);
});

test("P0 auth success resumes the pending action instead of always returning home", () => {
  const app = readUiFile("./src/app.jsx");
  const auth = readUiFile("./src/screen-auth.jsx");

  assert.match(app, /const requestAuth = /);
  assert.match(app, /pendingAction\?\.route/);
  assert.match(app, /pendingAction\.params/);
  assert.doesNotMatch(app, /setRoute\(\{\s*name:\s*"home",\s*params:\s*\{\}\s*\}\);/);

  assert.match(auth, /pendingAction/);
  assert.match(auth, /nav\("auth_verify", \{ email, pendingAction \}\)/);
  assert.match(auth, /AuthProfileSetupScreen/);
  assert.doesNotMatch(auth, /nav\("auth_register"/);
});

test("job picks module is outside current scope and jd_match aliases back home (D-17)", () => {
  const app = readUiFile("./src/app.jsx");

  assert.ok(!existsSync(new URL("./src/screen-jd-match.jsx", import.meta.url)), "screen-jd-match.jsx must stay absent");
  assert.match(app, /jd_match:\s*"home"/);
  assert.doesNotMatch(app, /jd_match:\s*</);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /JDMatchScreen|岗位推荐|Job Picks|Job picks|job recommendations/, `${name} still references the out-of-scope job picks module`);
    if (name !== "app.jsx") {
      assert.doesNotMatch(source, /jd_match/, `${name} still references the out-of-scope jd_match route`);
    }
  }
});

test("home uses a resume dropdown and caps recent mocks at three", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(home, /data-testid="home-jd-input-card"/);
  assert.doesNotMatch(home, /data-testid="home-jd-source-controls"/);
  assert.doesNotMatch(home, /data-testid="home-upload-trigger"/);
  assert.doesNotMatch(home, /data-testid="home-url-trigger"/);
  assert.doesNotMatch(home, /data-testid="home-source-layout"/);
  assert.doesNotMatch(home, /data-testid="home-upload-source-panel"/);
  assert.match(home, /data-testid="home-resume-row"/);
  assert.match(home, /data-testid="home-submit-row"/);
  assert.match(home, /<select[\s\S]*data-testid="home-resume-select"[\s\S]*value=\{selectedResumeId\}/);
  assert.match(home, /style=\{\{ width: 360, maxWidth: "100%"/);
  assert.match(home, /recentJobs\.slice\(0, 3\)/);
  assert.match(home, /hasMoreRecentJobs/);
  assert.match(home, /nav\("workspace", \{\}\)/);
  assert.doesNotMatch(home, /home-resume-option/);
});

test("TargetJob round assumptions use structured interview rounds across parse and recent mocks", () => {
  const home = readUiFile("./src/screen-home.jsx");
  const parse = readUiFile("./src/screens-p0-complete.jsx");
  const data = readUiFile("./src/data.jsx");

  assert.match(data, /const jdSampleInterviewRounds = \[/);
  assert.match(data, /interviewRounds: jdSampleInterviewRounds/);
  assert.match(data, /durationMinutes: 45/);
  assert.match(home, /round\.durationMinutes/);
  assert.match(parse, /roundId: currentRound \? `round-\$\{currentRound\.sequence\}-\$\{currentRound\.type\}` : ""/);
  assert.match(parse, /gridTemplateColumns: `repeat\(\$\{Math\.min\(parsed\.rounds\.length \|\| 1, 4\)\}, 1fr\)`/);
  assert.doesNotMatch(data, /interviewHypotheses/);
  assert.doesNotMatch(parse, /interviewHypotheses/);
  assert.doesNotMatch(data, /focus: "动机、求职节奏、薪资期望"/);
  assert.doesNotMatch(parse, /focus: lang === "en" \? "Motivation, timing, comp"/);
});

test("Parse exposes a page-level reports handoff and owns no report list", () => {
  const parse = readUiFile("./src/screens-p0-complete.jsx");
  const app = readUiFile("./src/app.jsx");
  const start = parse.indexOf("const ParseScreen = ");
  const end = parse.indexOf("const ReportsScreen = ", start);
  const parseScreen = parse.slice(start, end);
  const topBarStart = app.indexOf("const TopBar = ");
  const topBar = app.slice(topBarStart);

  assert.ok(start >= 0 && end > start);
  assert.ok(topBarStart >= 0);
  assert.match(parseScreen, /data-testid="parse-reports-entry"/);
  assert.match(parseScreen, /nav\("reports", \{ targetJobId: targetJob\.id \}\)/);
  assert.doesNotMatch(parseScreen, /params\.section|reportsSectionRef|reportOverviewFixtures|reportDemoState/);
  assert.doesNotMatch(parseScreen, /data-testid="parse-report-(?:section|loading|error|round|current|generating|failed)/);
  assert.doesNotMatch(parseScreen, /section:\s*"reports"|listTargetJobReports|Retry report|重试报告|parse-report-retry/);
  assert.doesNotMatch(parseScreen, /\.items\b|\.pageInfo\b|latestReportId|provenance|modelId|rubricVersion/);
  assert.doesNotMatch(topBar, /\{\s*k:\s*"reports"/);
});

test("independent ReportsScreen renders only the current plan canonical current/latest projection", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");
  const start = p0.indexOf("const ReportsScreen = ");
  const end = p0.indexOf("// #3 REPORT GENERATING", start);
  const reports = p0.slice(start, end);

  assert.ok(start >= 0 && end > start);
  assert.match(reports, /window\.EI_DATA\.targetJobs\.find/);
  assert.match(reports, /window\.EI_DATA\.jdSample\.interviewRounds/);
  assert.match(reports, /window\.EI_DATA\.reportOverviewFixtures/);
  assert.match(reports, /reportDemoStates\.has\(demoState\) \? demoState : "ready"/);
  assert.doesNotMatch(reports, /params\.reportState/);
  for (const state of ["ready", "loading", "empty", "error", "latest-ready", "mismatch"]) {
    assert.match(reports, new RegExp(`"${state}"`));
  }
  for (const testId of [
    "reports-screen",
    "reports-back-button",
    "reports-loading",
    "reports-empty",
    "reports-error",
    "reports-list",
    "reports-current",
    "reports-generating",
    "reports-failed",
    "reports-latest-ready",
  ]) {
    assert.match(reports, new RegExp(`data-testid=["\\{][^\\n]*${testId}`));
  }
  assert.match(reports, /overview\.targetJobId !== targetJob\.id/);
  assert.match(reports, /!item\?\.round/);
  assert.match(reports, /item\.round\.roundSequence/);
  assert.match(reports, /item\.round\.roundId/);
  assert.match(reports, /latestAttempt\.id !== item\.currentReport\?\.id/);
  assert.match(reports, /nav\("report", \{ reportId: item\.currentReport\.id \}\)/);
  assert.match(reports, /nav\("generating", \{ reportId: item\.latestAttempt\.id \}\)/);
  assert.match(reports, /nav\("workspace", \{ targetJobId: targetJob\.id \}\)/);
  assert.match(reports, /nav\("workspace"\)/);
  assert.doesNotMatch(reports, /Retry report|重试报告|reports-retry|timeline|时间线|fullHistory|reportHistory/);
  assert.doesNotMatch(reports, /section:\s*"reports"|provenance|modelId|rubricVersion/);
});

test("Report detail and pending states return to trusted ReportsScreen or safe workspace", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /data-testid="report-back-button"[\s\S]*nav\("reports", \{ targetJobId: report\.targetJobId \}\)/);
  assert.match(report, /const ReportPendingState = \(\{ T, lang, nav, reportId, targetJobId \}\) =>/);
  assert.match(report, /targetJobId \? nav\("reports", \{ targetJobId \}\) : nav\("workspace"\)/);
  assert.match(report, /const ReportMissingState = [\s\S]*nav\("workspace"\)/);
  assert.doesNotMatch(report, /nav\("parse"|section:\s*"reports"/);
});

test("prototype round progress is backend-projected and never inferred from lifecycle text", () => {
  const app = readUiFile("./src/app.jsx");
  const data = readUiFile("./src/data.jsx");
  const home = readUiFile("./src/screen-home.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");
  const parse = readUiFile("./src/screens-p0-complete.jsx");
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(app, /const eiResolvePracticeProgress = /);
  assert.match(app, /completedRounds/);
  assert.match(app, /currentRound/);
  assert.doesNotMatch(app, /round\.sequence !== index \+ 1/);
  assert.match(app, /round\.sequence <= 0/);
  assert.match(app, /round\.sequence <= rounds\[index - 1\]\.sequence/);
  assert.match(data, /practiceProgress:/);
  assert.doesNotMatch(data, /nextRound:/);
  assert.match(home, /eiResolvePracticeProgress\(rounds, job\.practiceProgress\)/);
  assert.match(workspace, /eiResolvePracticeProgress\(rounds, job\.practiceProgress\)/);
  assert.match(parse, /eiResolvePracticeProgress\(parsed\.rounds, targetJob\.practiceProgress\)/);
  assert.match(report, /report\.context\.hasNextRound/);
  assert.doesNotMatch(report, /eiResolvePracticeProgress|eiResolveInterviewRoundContext/);

  for (const [name, source] of [["home", home], ["workspace", workspace]]) {
    assert.doesNotMatch(source, /job\?\.nextRound|job\.nextRound/, `${name} still reads free-text nextRound`);
    assert.doesNotMatch(source, /job\?\.status ===|job\.status ===/, `${name} still derives a round from lifecycle status`);
  }
});

test("prototype Workspace detail exposes persisted round states with three semantic treatments", () => {
  const app = readUiFile("./src/app.jsx");
  const parse = readUiFile("./src/screens-p0-complete.jsx");

  assert.match(app, /TARGET_JOB_LOCATOR_ROUTES = new Set\(\["parse", "reports", "workspace"\]\)/);
  assert.match(app, /workspace: route\.params\.targetJobId\s*\?\s*<ParseScreen/);
  assert.match(parse, /const roundState = !progress\.valid/);
  assert.match(parse, /data-round-state=\{roundState \|\| undefined\}/);
  assert.match(parse, /roundState === "done" \? T\.okSoft/);
  assert.match(parse, /roundState === "current" \? T\.accentSoft/);
  assert.match(parse, /roundState === "done" \? T\.ok/);
  assert.match(parse, /roundState === "current" \? T\.accent/);
  assert.match(parse, /"已进行"/);
  assert.match(parse, /"即将进行"/);
  assert.match(parse, /"未进行"/);
  assert.doesNotMatch(parse, /targetJob\?*\.status[^\n]*(?:roundState|completedCount|currentIndex)/);
});

test("prototype does not persist interview business progress in browser state", () => {
  const app = readUiFile("./src/app.jsx");
  assert.doesNotMatch(app, /ei-route|ei-signed-in|ei-profile-complete/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(
      source,
      /(?:localStorage|sessionStorage)\.(?:setItem|getItem)\(["'`](?:practiceProgress|completedRounds|currentRound|roundProgress|practicePlan)/,
      `${name} persists interview business progress in browser storage`,
    );
    assert.doesNotMatch(source, /indexedDB\.(?:open|deleteDatabase)\(/, `${name} uses IndexedDB as a business-state source`);
  }
});

test("workspace insight source stays absent from the pure plan-list prototype", () => {
  const app = readUiFile("./src/app.jsx");
  const index = readUiFile("./index.html");
  const workspace = readUiFile("./src/screen-workspace.jsx");
  const outOfScopeInsightTerms = new RegExp([
    "company_" + "intel",
    "Company" + "Intel",
    "getCompany" + "Intel",
  ].join("|"));

  assert.ok(!existsSync(new URL("./src/screen-company-" + "intel.jsx", import.meta.url)), "standalone insight source must stay absent");
  assert.ok(!existsSync(new URL("./src/screen-workspace-insight.jsx", import.meta.url)), "workspace insight source must stay absent");
  assert.doesNotMatch(app, outOfScopeInsightTerms);
  assert.doesNotMatch(index, /screen-workspace-insight\.jsx/);
  assert.doesNotMatch(workspace, /WorkspaceInsightCard/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, outOfScopeInsightTerms, `${name} still references standalone company insight naming`);
  }
});

test("settings keeps only profile and privacy tabs (D-21)", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");

  assert.match(p0, /\{ k: "profile", t: "个人资料" \}, \{ k: "privacy", t: "隐私与数据" \}/);
  assert.doesNotMatch(p0, /SettingsNotif|SettingsBilling|"notifications"|"billing"|订阅/);
});

test("theme defaults to ocean and keeps the custom accent picker (D-21 v2.1)", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.match(app, /"theme": "ocean"/);
  assert.match(app, /const CUSTOM_ACCENT_SEEDS = /);
  assert.match(app, /const AccentPicker = /);
  assert.match(app, /customAccent/);
  assert.match(app, /\? tweaks\.theme : "ocean"/);
  assert.match(canvas, /Ocean · 深海（默认）/);
  assert.match(canvas, /customAccent/);
});

test("custom accent picker keeps only hue and saturation controls", () => {
  const app = readUiFile("./src/app.jsx");
  const call = app.match(/<AccentPicker[\s\S]*?\/>/)?.[0] ?? "";
  const pickerStart = app.indexOf("const AccentPicker = ");
  const pickerEnd = app.indexOf("\nwindow.App", pickerStart);
  const picker = app.slice(pickerStart, pickerEnd);

  assert.ok(pickerStart >= 0 && pickerEnd > pickerStart, "AccentPicker source must be present");
  assert.match(
    picker,
    /const AccentPicker = \(\{ T, lang, dark, value, onChange \}\) =>/,
  );
  assert.match(picker, /lang === "en" \? "Hue" : "色相"/);
  assert.match(picker, /lang === "en" \? "Chroma" : "饱和度"/);
  assert.match(picker, /type="range" min=\{0\} max=\{360\}/);
  assert.match(picker, /type="range" min=\{0\} max=\{0\.25\}/);

  assert.doesNotMatch(call, /\b(?:active|onClear)=/);
  assert.doesNotMatch(picker, /\b(?:active|onClear|previewAccent)\b/);
  assert.doesNotMatch(
    picker,
    /Reset to theme accent|恢复主题默认色|Drag to apply|拖动应用/,
  );
});

test("home keeps the debrief auxiliary entry outside current scope (D-22)", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.doesNotMatch(home, /POST-INTERVIEW|post-interview|nav\("debrief"\)|Open debrief|打开复盘/);
  assert.doesNotMatch(home, /JOB PICKS|jobPicks/);
});

test("user menu no longer exposes the user profile route (D-22)", () => {
  const app = readUiFile("./src/app.jsx");

  assert.doesNotMatch(app, /label:\s*"用户画像"|label:\s*"Profile"|nav\("profile"\)/);
  assert.match(app, /labelZh:\s*"设置与隐私"/);
  assert.match(app, /nav\("settings"\)/);
  assert.match(app, /nav\("auth_logout"\)/);
});

test("P0 empty and failure states avoid showing fake data", () => {
  const home = readUiFile("./src/screen-home.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");
  const report = readUiFile("./src/screen-report.jsx");
  const practice = readUiFile("./src/screen-practice.jsx");

  assert.match(home, /const recentJobs = /);
  assert.match(home, /HomeEmptyState/);
  assert.match(workspace, /visibleJobs\.length === 0/);
  assert.match(workspace, /data-testid="workspace-plan-list-empty"/);
  assert.doesNotMatch(workspace, /WorkspaceEmptyState|WorkspaceMissingResumeState/);
  assert.match(report, /ReportMissingState/);
  assert.match(report, /ReportFailureState/);
  assert.doesNotMatch(practice, /VoiceTranscriptionFailure|transcriptFailed/);
});

test("Home recent mock interviews are signed-in only", () => {
  const app = readUiFile("./src/app.jsx");
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(app, /home:\s*<HomeScreen[^>]*signedIn=\{signedIn\}/);
  assert.match(home, /const HomeScreen = \(\{ T, lang, nav, signedIn/);
  assert.match(home, /\{signedIn && \(/);
  assert.match(home, /Recent mock interviews/);
  assert.match(home, /选择已有简历/);
  assert.match(home, /还没有简历？1 分钟创建 →/);
  assert.match(home, /立即面试/);
  assert.doesNotMatch(home, /粘贴 JD，或继续最近一次模拟面试。每一次练习都绑定具体岗位，而不是泛用题库。/);
  assert.doesNotMatch(home, /解析并确认面试/);
});

test("Home JD intake exposes one paste-only path", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(home, /data-testid="home-jd-textarea"/);
  assert.match(home, /data-testid="home-resume-select"/);
  assert.match(home, /data-testid="home-submit-row"/);
  assert.doesNotMatch(home, /assistOpen|JDAssistModal|home-jd-source-controls/);
  assert.doesNotMatch(home, /home-upload-trigger|home-url-trigger|uploadSource|orUpload/);
  assert.doesNotMatch(home, /Paste or upload|粘贴或上传|source:\s*assistOpen|source:\s*"pasted"/);
});

test("Home and workspace share action card behavior", () => {
  const home = readUiFile("./src/screen-home.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");

  assert.match(home, /onClick=\{\(\) => nav\("workspace", \{ targetJobId: j\.id \}\)\}/);
  assert.match(home, /nav\("parse", \{ targetJobId: "tj-import-pending" \}\)/);
  assert.equal((home.match(/nav\("parse"/g) || []).length, 1);
  assert.doesNotMatch(workspace, /nav\("parse"/);
  assert.match(home, /onStart=\{\(round\) => nav\("practice", \{ targetJobId: j\.id, roundId: round\.id/);
  assert.match(home, /Start interview now/);
  assert.doesNotMatch(home, /showDelete=\{true\}/);
  assert.match(workspace, /onClick=\{\(\) => openPlan\(job\)\}/);
  assert.match(workspace, /startInterview\(job\)/);
  assert.match(workspace, /Icon name="trash"/);
  assert.match(workspace, /position:\s*"absolute", top:\s*20, right:\s*20/);
  assert.doesNotMatch(workspace, /open:\s*"Open plan"|open:\s*"进入规划"|L\.open/);
});

test("practice is one continuous text conversation with phone disabled", () => {
  const practice = readUiFile("./src/screen-practice.jsx");
  const primitives = readUiFile("./src/primitives.jsx");
  assert.match(practice, /data-testid="practice-topbar-phone-toggle"/);
  assert.equal((practice.match(/data-testid="practice-topbar-phone-toggle"/g) || []).length, 1);
  assert.match(practice, /data-testid="practice-topbar-phone-toggle"[\s\S]*disabled[\s\S]*aria-disabled="true"/);
  assert.match(practice, /电话模式暂未开放|Phone mode is temporarily unavailable/);
  assert.match(primitives, /phone:\s*<>/);
  assert.match(practice, /data-testid="practice-conversation"/);
  assert.match(practice, /width: "100%"/);
  assert.doesNotMatch(practice, /SESSION MAP|本轮题目|QuestionHeader|QuestionCard|qIdx|currentQ|questions\.map/);
  assert.doesNotMatch(practice, /Question[^a-z]|题\s*\{/);
  assert.doesNotMatch(practice, /PhoneSessionSurface|WaveformBars|practice-phone-surface|practice-phone-captions/);
  assert.match(practice, /const \{ currentRound \} = window\.eiResolveInterviewRoundContext/);
  assert.match(practice, /currentRound \? formatElapsed\(currentRound\.durationMinutes \* 60\) : "--:--"/);
  assert.match(practice, /const hasCommittedCandidateMessage = messages\.some/);
  assert.match(practice, /const canFinishInterview = /);
  assert.match(practice, /const finishReasonId = "practice-finish-disabled-reason"/);
  assert.match(practice, /data-testid="practice-finish-cta"/);
  assert.match(practice, /disabled=\{!canFinishInterview\}/);
  assert.match(practice, /aria-describedby=\{!canFinishInterview \? finishReasonId : undefined\}/);
  assert.match(practice, /data-testid="practice-finish-disabled-reason"/);
  assert.match(practice, /请先完成至少一次回答|Complete at least one answer first/);
  assert.match(practice, /nav\("generating", \{ reportId: D\.report\.id \}\)/);
  assert.doesNotMatch(practice, /nav\("generating", \{ \.\.\.context \}\)/);
  assert.doesNotMatch(practice, /25:00/);
});

test("practice shows immediate user rows, interviewer thinking, and failed-row retry only", () => {
  const practice = readUiFile("./src/screen-practice.jsx");
  const primitives = readUiFile("./src/primitives.jsx");

  assert.match(practice, /status: message\.role === "user" \? \(message\.status \|\| "complete"\) : undefined/);
  assert.match(practice, /const \[pendingMessageId, setPendingMessageId\] = React\.useState\(null\)/);
  assert.match(practice, /const requestReply = /);
  assert.match(practice, /data-testid="practice-interviewer-thinking"/);
  assert.match(practice, /aria-live="polite"/);
  assert.match(practice, /data-testid="practice-message-retry"/);
  assert.match(practice, /msg\.status === "retryable_failed"/);
  assert.match(practice, /retryFailedMessage\(msg\)/);
  assert.match(practice, /disabled=\{paused \|\| isThinking \|\| hasTerminalFailedMessage\}/);
  assert.match(practice, /disabled=\{paused \|\| isThinking \|\| hasFailedCandidateMessage \|\| !input\.trim\(\)\}/);
  assert.match(practice, /data-testid="practice-terminal-recovery" role="alert"/);
  assert.match(practice, /nav\("workspace", \{ targetJobId: job\.id \}\)/);
  assert.doesNotMatch(practice, /nav\("parse", \{ targetJobId: job\.id \}\)/);
  assert.match(practice, /返回当前面试规划|Return to this interview plan/);
  assert.doesNotMatch(practice, /msg\.status === "terminal_failed"[\s\S]{0,200}data-testid="practice-message-retry"/);
  assert.match(primitives, /refresh:\s*</);
  assert.doesNotMatch(practice, /setTimeout\(\(\) => setMessages\(\(current\) => \[\.\.\.current, \{\s*role: "ai"/);
});

test("practice exposes deterministic reply-state demos for four-state source parity", () => {
  const practice = readUiFile("./src/screen-practice.jsx");

  assert.match(practice, /params\.replyState/);
  for (const state of [
    "immediate-pending",
    "persisted-pending",
    "retryable-failed",
    "terminal-failed",
  ]) {
    assert.match(practice, new RegExp(`["]${state}["]`));
  }
  for (const testId of [
    "practice-screen",
    "practice-topbar",
    "practice-conversation",
    "practice-transcript",
    "practice-input",
    "practice-input-textarea",
    "practice-input-send",
    "practice-terminal-recovery-cta",
  ]) {
    assert.match(practice, new RegExp(`data-testid=["]${testId}["]`));
  }
  assert.match(practice, /const isThinking = /);
  assert.match(practice, /const finishDisabledReason = /);
  assert.match(practice, /isThinking \? \(lang === "en" \? "Wait for the interviewer reply\." : "请等待面试官回复。"\)/);
  assert.match(practice, /hasFailedCandidateMessage \? \(lang === "en" \? "Resolve the unfinished reply first\." : "请先恢复这条未完成回复的消息。"\)/);
});

test("Parse loading keeps progress but hides internal model and rubric metadata", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");
  const start = p0.indexOf("const ParseScreen = ");
  const end = p0.indexOf("// #3 REPORT GENERATING", start);
  const parse = p0.slice(start, end);

  assert.ok(start >= 0 && end > start);
  assert.match(parse, /steps\.map/);
  assert.doesNotMatch(parse, /claude-haiku|rubric ·|prompt@|typical ·|provenance/i);
});

test("Report context strip hides internal session and report locators", () => {
  const report = readUiFile("./src/screen-report.jsx");
  const start = report.indexOf("const ReportContextStrip = ");
  const end = report.indexOf("const ReportMetric = ", start);
  const strip = report.slice(start, end);

  assert.ok(start >= 0 && end > start);
  assert.match(strip, /targetJobCompany/);
  assert.match(strip, /roundName/);
  assert.match(strip, /resumeDisplayName/);
  assert.doesNotMatch(strip, /report\.(?:sessionId|id)|report-context-session|SESSION|会话/);
});

test("prototype report routes retain only their stable resource locators", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.match(app, /const REPORT_LOCATOR_ROUTES = new Set\(\["generating", "report"\]\)/);
  assert.match(app, /const TARGET_JOB_LOCATOR_ROUTES = new Set\(\["parse", "reports", "workspace"\]\)/);
  assert.match(app, /return stripUndefined\(\{ reportId: params\.reportId \}\)/);
  assert.match(app, /return stripUndefined\(\{ targetJobId: params\.targetJobId \}\)/);
  assert.match(app, /const prototypeReportDemoState = activeRouteName === "reports"/);
  assert.match(app, /reports:\s*<ReportsScreen[^>]*params=\{route\.params \|\| \{\}\}[^>]*demoState=\{prototypeReportDemoState\}/);
  assert.match(canvas, /\["generating", "report"\]\.includes\(route\).*params\.set\("reportId", "report-24"\)/);
  assert.doesNotMatch(canvas, /\["practice", "generating", "report"\]\.includes\(route\).*sessionId/);
});

test("P0 report has no voice or modality-specific report branch", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.doesNotMatch(report, /params\.modality|practiceMode|hintUsed|hintCount/);
  assert.doesNotMatch(report, /Phone|电话模式|Voice|语音/);
});

test("Parse is command progress and Workspace owns the readonly saved-plan receipt", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");
  const app = readUiFile("./src/app.jsx");

  assert.match(p0, /const PlanBindingPill = /);
  assert.match(p0, /<PlanBindingPill T=\{T\}/);
  assert.doesNotMatch(p0, /window\.BindingPill/);
  assert.match(p0, /const ParseScreen = \(\{ T, lang, nav, requestAuth, params = \{\}, readyDetail = false \}\) =>/);
  assert.match(p0, /React\.useState\(readyDetail \? "preview" : "loading"\)/);
  assert.match(p0, /nav\("workspace", \{ targetJobId: params\.targetJobId \}\)/);
  assert.match(p0, /window\.getWorkspaceResumeOptions/);
  assert.match(p0, /立即面试/);
  assert.match(p0, /type: "create_session"/);
  assert.match(p0, /nav\("practice", startContext\)/);
  assert.doesNotMatch(p0, /ResumePickerModal/);
  assert.doesNotMatch(p0, /仅保存规划|Save plan only/);
  assert.doesNotMatch(p0, /nav\("workspace", buildParseInterviewContext\(\)\)/);
  assert.doesNotMatch(p0, /确认并进入面试前确认|Confirm & open interview setup/);
  assert.match(app, /parse:\s*<ParseScreen[^>]*requestAuth=\{requestAuth\}/);
  assert.match(app, /workspace: route\.params\.targetJobId\s*\?\s*<ParseScreen[^>]*readyDetail/);
});

test("D-22 does not leave debrief source or navigation hooks", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /DebriefFullScreen|DebriefContextPickerModal|GuidedDebriefRecord|VoiceDebriefRecord|DebriefReplayPlan/, `${name} still contains debrief components`);
    if (name !== "app.jsx") {
      assert.doesNotMatch(source, /nav\("debrief"|nav\("debrief_full"/, `${name} still navigates to debrief`);
    }
  }

  const app = readUiFile("./src/app.jsx");
  assert.doesNotMatch(app, /INTERVIEW_CONTEXT_ROUTES = new Set\(\[[^\]]*"debrief"/);
});

test("auth has no standalone reset page and aliases auth_reset back to login", () => {
  const app = readUiFile("./src/app.jsx");
  const auth = readUiFile("./src/screen-auth.jsx");

  assert.match(app, /auth_reset:\s*"auth_login"/);
  assert.doesNotMatch(app, /AuthResetScreen/);
  assert.doesNotMatch(app, /auth_reset:\s*</);
  assert.doesNotMatch(auth, /AuthResetScreen|PASSWORD RESET|找回密码|发送重置说明|auth_reset/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /忘记密码|两步验证|Two-step verification/, `${name} still references password-era auth wording`);
    if (name !== "app.jsx") {
      assert.doesNotMatch(source, /auth_reset/, `${name} still references the out-of-scope reset route`);
    }
  }
});

test("practice owns one deterministic semantic GFM message-body and mobile overflow contract", () => {
  const practice = readUiFile("./src/screen-practice.jsx");

  assert.match(practice, /"markdown-gfm"/);
  assert.match(practice, /const PracticeMessageBody = /);
  assert.match(practice, /renderPracticeMarkdownBlocks\(text\)/);
  assert.match(practice, /PRACTICE_MARKDOWN_SAFE_SCHEMES/);
  assert.match(practice, /noopener noreferrer/);
  assert.match(practice, /data-testid="practice-message-body"/);
  assert.match(practice, /className="ei-practice-message-body"/);
  assert.match(practice, /PRACTICE_MESSAGE_BODY_CSS/);
  assert.match(practice, /<blockquote key=/);
  assert.match(practice, /<table key=/);
  assert.match(practice, /<pre key=/);
  assert.match(practice, /\.ei-practice-message-body pre \{/);
  assert.match(practice, /\.ei-practice-message-body table \{/);
  assert.match(practice, /max-width: 100%/);
  assert.match(practice, /overflow-x: auto/);
  assert.match(practice, /overscroll-behavior-x: contain/);
  assert.doesNotMatch(practice, /markdownDemo/);
  assert.doesNotMatch(practice, /dangerouslySetInnerHTML/);
  assert.doesNotMatch(practice, /<img\b/);
});

test("debrief thank-you letter draft stays absent", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /ThankYouLetter|感谢信/, `${name} still contains the out-of-scope thank-you letter draft`);
  }
});

test("out-of-scope resume versions screen stays absent", () => {
  assert.ok(!existsSync(new URL("./src/screens-p1-depth.jsx", import.meta.url)), "screens-p1-depth.jsx must stay absent");
  for (const [name, source] of readUiSources()) {
    const outOfScopeResumeSourcePattern = new RegExp(`_${"Leg" + "acy"}ResumeVersionsScreen|ResumeSourceMap`);
    assert.doesNotMatch(source, outOfScopeResumeSourcePattern, `${name} still contains out-of-scope resume version source`);
  }
});

test("phone interview has no positive canvas or route entry while disabled", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.doesNotMatch(app, /voice:\s*"practice"/);
  assert.doesNotMatch(app, /rawRoute === "voice"/);
  assert.doesNotMatch(app, /voice:\s*<VoicePracticeScreen/);
  assert.doesNotMatch(app, /route\.name === "voice"/);
  assert.doesNotMatch(canvas, /mode="phone"|practice-phone|phone-light|phone-dark/);
  assert.match(canvas, /电话入口暂不可用/);
});

test("design canvas component surface matches its only tracked consumer", () => {
  const wrapper = readUiFile("./design-canvas.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.match(wrapper, /function DesignCanvas\(\{ children \}\)/);
  assert.match(wrapper, /function DCViewport\(\{ children \}\)/);
  assert.match(wrapper, /function DCSection\(\{ id, title, subtitle, children \}\)/);
  assert.match(wrapper, /function DCPostIt\(\{ children, top, left, right, rotate = -2, width = 180 \}\)/);
  assert.doesNotMatch(wrapper, /\bminScale\b|\bmaxScale\b/);
  assert.doesNotMatch(wrapper, /const \{[^}]*\bstyle = \{\}[^}]*\} = artboard\.props/);

  assert.match(canvas, /const Screen = \(\{ route, mode, dark, theme, customAccent, fontPreset \}\) =>/);
  assert.doesNotMatch(canvas, /const PracticeScreen\b/);
  assert.doesNotMatch(canvas, /params\.set\("(?:nochrome|jobId|targetJobId|planId|jdId|resumeId|roundId|roundName)"/);
  assert.match(canvas, /params\.set\("lang", "zh"\)/);
  assert.match(canvas, /route === "practice".*params\.set\("sessionId", "session-24"\)/);
  assert.match(canvas, /\["generating", "report"\]\.includes\(route\).*params\.set\("reportId", "report-24"\)/);
});

test("design canvas keeps edits in memory without an unavailable sidecar bridge", () => {
  const wrapper = readUiFile("./design-canvas.jsx");

  assert.doesNotMatch(wrapper, /\.design-canvas\.state\.json|DC_STATE_FILE|window\.omelette/);
  assert.doesNotMatch(wrapper, /\b(?:didRead|skipNextWrite|setReady)\b/);
  assert.doesNotMatch(wrapper, /fetch\(['"]\.\//);
  assert.match(wrapper, /const \[state, setState\] = React\.useState\(\{ sections: \{\}, focus: null \}\)/);
  assert.match(wrapper, /patchSection: \(id, p\) => setState/);
  assert.match(wrapper, /<DCViewport>\{children\}<\/DCViewport>/);
});

test("prototype display controls do not depend on an unavailable edit-mode host", () => {
  const app = readUiFile("./src/app.jsx");
  const home = readUiFile("./src/screen-home.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.doesNotMatch(app, /EDITMODE|__activate_edit_mode|__deactivate_edit_mode|__edit_mode_|window\.parent\.postMessage/);
  assert.doesNotMatch(app, /TweaksPanel|TweakRow|selectStyle|tweaksOpen|tweaksAvailable/);
  assert.doesNotMatch(app, /["']role["']\s*:|\["dark","role"|role=\{tweaks\.role\}|setRole=/);
  assert.doesNotMatch(home, /const HomeScreen = \(\{ T, lang, nav, role,/);
  assert.doesNotMatch(canvas, /\bTweaks\b|edit mode bridge/);

  assert.match(app, /<TopBar[^>]*setDark=\{\(v\) => updateTweak\("dark", v\)\}/);
  assert.match(app, /setCustomAccent=\{\(v\) => updateTweak\("customAccent", v\)\}/);
  assert.match(app, /settings:\s*<SettingsScreen[^>]*setFontPreset=\{setFontPreset\}/);
});

test("prototype app shell has no zero-read canvas mode binding", () => {
  const app = readUiFile("./src/app.jsx");

  assert.doesNotMatch(app, /\bconst isCanvasIframe\b/);
  assert.match(app, /const hideTopBar = [^;]*data-nochrome[^;]*;/);
});

test("auth prototype screens expose only consumed navigation and completion callbacks", () => {
  const app = readUiFile("./src/app.jsx");
  const auth = readUiFile("./src/screen-auth.jsx");

  assert.match(auth, /const AuthLoginScreen = \(\{ T, lang, nav, pendingAction \}\) =>/);
  assert.match(auth, /const AuthVerifyScreen = \(\{ T, lang, nav, email, onSignIn, pendingAction \}\) =>/);
  assert.match(auth, /const AuthProfileSetupScreen = \(\{ T, lang, onCompleteProfile, pendingAction \}\) =>/);
  assert.doesNotMatch(app, /auth_login:\s*<AuthLoginScreen[^>]*\bonSignIn=/);
  assert.match(app, /auth_verify:\s*<AuthVerifyScreen[^>]*\bonSignIn=\{completeSignIn\}/);
  assert.doesNotMatch(app, /auth_profile_setup:\s*<AuthProfileSetupScreen[^>]*\bnav=/);
  assert.match(app, /auth_profile_setup:\s*<AuthProfileSetupScreen[^>]*\bonCompleteProfile=\{completeProfile\}/);
});

test("settings prototype screen exposes only consumed display dependencies", () => {
  const app = readUiFile("./src/app.jsx");
  const screens = readUiFile("./src/screens-p0-complete.jsx");

  assert.match(screens, /const SettingsScreen = \(\{ T, lang, fontPreset, setFontPreset \}\) =>/);
  assert.doesNotMatch(app, /settings:\s*<SettingsScreen[^>]*\bnav=/);
  assert.match(app, /settings:\s*<SettingsScreen[^>]*\bfontPreset=\{tweaks\.fontPreset\}/);
  assert.match(app, /settings:\s*<SettingsScreen[^>]*\bsetFontPreset=\{setFontPreset\}/);
});

test("home mini round rail exposes only rendered structured-round dependencies", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(home, /const MiniRoundRail = \(\{ T, rounds, currentIndex \}\) =>/);
  assert.match(home, /<MiniRoundRail T=\{T\} rounds=\{rounds\} currentIndex=\{currentRoundIndex\} \/>/);
  assert.match(home, /gridTemplateColumns: `repeat\(\$\{rounds\.length\}, 1fr\)`/);
  assert.match(home, /\{round\.name\}\{round\.durationMinutes \? ` · \$\{round\.durationMinutes\}m` : ""\}/);
  assert.match(home, /const current = i === currentIndex/);
});

test("same-name round rails key nodes by canonical round identity", () => {
  for (const screen of ["./src/screen-home.jsx", "./src/screen-workspace.jsx"]) {
    const source = readUiFile(screen);
    assert.match(source, /key=\{`round-\$\{round\.sequence\}-\$\{round\.type\}`\}/);
    assert.doesNotMatch(source, /key=\{round\.name\}/);
  }
});

test("report detail surface exposes dimensions, evidence, and actions only", () => {
  const report = readUiFile("./src/screen-report.jsx");
  assert.match(report, /CAPABILITY ASSESSMENT/);
  assert.match(report, /STRENGTH EVIDENCE/);
  assert.match(report, /NEXT ACTIONS/);
  assert.doesNotMatch(report, /question|Question|题目|qId|turnId/);
});

test("resume create prototype exposes only input-state owner callbacks", () => {
  const workshop = readUiFile("./src/screen-resume-workshop.jsx");
  const createCall = workshop.match(/<ResumeCreateFlow[\s\S]*?\/>/)?.[0];

  assert.ok(createCall, "ResumeCreateFlow call must remain present");
  assert.match(workshop, /const ResumeCreateFlow = \(\{ T, lang, onBack, onCreateResume \}\) =>/);
  assert.doesNotMatch(createCall, /\bnav=/);
  assert.match(createCall, /onBack=\{\(\) => setFlow\("list"\)\}/);
  assert.match(createCall, /onCreateResume=\{addCreatedResume\}/);
  assert.match(workshop, /\{ k: "upload"[\s\S]*\{ k: "paste"/);
  assert.match(workshop, /onCreateResume\(sourceLabel, createMode === "paste" \? resumeText : ""\)/);
});
