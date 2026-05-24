// Promptfoo custom provider for the F3 offline eval harness (plan 004 §4.3).
// It is a thin bridge: all real work (registry single-source resolution and,
// in EVAL_LIVE mode, the live model call) happens in the Go `evalkit` binary.
// The candidate output for each case comes from `evalkit complete`, which is a
// recorded golden transcript by default and a live provider call when EVAL_LIVE=1.
const { execFileSync } = require('node:child_process');
const path = require('node:path');

const repoRoot = path.resolve(__dirname, '..', '..');
const evalkitBin = process.env.EVALKIT_BIN || path.join(repoRoot, 'backend', 'bin', 'evalkit');

class EvalkitProvider {
  id() {
    return 'evalkit-registry-provider';
  }

  async callApi(_prompt, context) {
    const caseId = context && context.vars && context.vars.caseId;
    if (!caseId) {
      return { error: 'evalkit provider: missing caseId var' };
    }
    const live = process.env.EVAL_LIVE === '1' ? ['--live'] : [];
    try {
      const out = execFileSync(evalkitBin, ['complete', '--case', caseId, ...live], {
        cwd: repoRoot,
        encoding: 'utf8',
        env: process.env,
      });
      return { output: out.trim() };
    } catch (err) {
      return { error: `evalkit complete failed for ${caseId}: ${err.message}` };
    }
  }
}

module.exports = EvalkitProvider;
