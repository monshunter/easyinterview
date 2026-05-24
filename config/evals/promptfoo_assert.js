// Promptfoo javascript assertion for the F3 offline eval harness (plan 004 §4.3).
// Grading is delegated to the single Go registry.LLMJudge via `evalkit grade`:
// it validates the candidate output against the registry output schema and
// scores every rubric dimension. The assertion passes when the judge returns a
// valid per-dimension verdict (offline: from the recorded fixture; EVAL_LIVE: a
// real judge call). No grading logic is reimplemented in JS (single source).
const { execFileSync } = require('node:child_process');
const path = require('node:path');

const repoRoot = path.resolve(__dirname, '..', '..');
const evalkitBin = process.env.EVALKIT_BIN || path.join(repoRoot, 'backend', 'bin', 'evalkit');

module.exports = (output, context) => {
  const caseId = context && context.vars && context.vars.caseId;
  if (!caseId) {
    return { pass: false, score: 0, reason: 'evalkit assert: missing caseId var' };
  }
  try {
    const raw = execFileSync(evalkitBin, ['grade', '--case', caseId, '--output', output], {
      cwd: repoRoot,
      encoding: 'utf8',
      env: process.env,
    });
    const verdict = JSON.parse(raw.trim().split('\n').pop());
    return {
      pass: verdict.pass === true,
      score: verdict.pass === true ? 1 : 0,
      reason: verdict.reason || (verdict.pass ? 'judge verdict valid' : 'judge fail-closed'),
    };
  } catch (err) {
    return { pass: false, score: 0, reason: `evalkit grade failed for ${caseId}: ${err.message}` };
  }
};
