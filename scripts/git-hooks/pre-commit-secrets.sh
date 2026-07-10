#!/usr/bin/env bash
# Pre-commit secret scan (secrets-and-config spec D-6 / §6 C-8).
# Scans staged additions/changes for known credential prefixes:
#   - AWS access key:  AKIA[0-9A-Z]{16}
#   - OpenAI / Anthropic API key:  sk-[A-Za-z0-9]{20,}
#   - Slack token:     xox[baprs]-[A-Za-z0-9-]+
# Reports file + line number; does NOT print the matched secret literal
# itself (prevents the log from re-leaking the credential).
#
# This hook is registered through scripts/git-hooks/pre-commit; call it from
# there or from the committer's local pre-commit framework.

set -euo pipefail

# Patterns: name|regex|advisory text
PATTERNS=(
  "aws-access-key|AKIA[0-9A-Z]{16}|AWS access key id"
  "openai-api-key|sk-[A-Za-z0-9]{20,}|API key (OpenAI/Anthropic shape)"
  "slack-token|xox[baprs]-[A-Za-z0-9-]+|Slack token"
)

# Allow a project-level allowlist file for documented test/fixture data.
ALLOWLIST_FILE=".secret-scan-allowlist"

# Collect staged additions: only +-prefixed diff content, never context.
DIFF=$(git diff --cached --no-color --unified=0 --diff-filter=ACMR -- '*')
if [ -z "$DIFF" ]; then
  exit 0
fi

current_file=""
current_line_old=0
current_line_new=0
violations=0

while IFS= read -r line; do
  case "$line" in
    "diff --git "*)
      # New file header: rest extracts the b/<path> target.
      current_file=$(printf '%s\n' "$line" | sed -E 's|^diff --git a/.* b/(.*)$|\1|')
      ;;
    "@@ "*)
      # Hunk header: parse the +start[,count] field.
      current_line_new=$(printf '%s\n' "$line" | sed -E 's|^@@ -[0-9,]+ \+([0-9]+).*|\1|')
      current_line_new=$((current_line_new - 1))
      ;;
    +++*|---*|"")
      # Skip diff metadata + blank lines.
      ;;
    +*)
      current_line_new=$((current_line_new + 1))
      payload="${line:1}"
      # Skip allowlisted files entirely.
      if [ -f "$ALLOWLIST_FILE" ] && grep -Fxq "$current_file" "$ALLOWLIST_FILE"; then
        continue
      fi
      for entry in "${PATTERNS[@]}"; do
        regex=$(printf '%s' "$entry" | cut -d'|' -f2)
        label=$(printf '%s' "$entry" | cut -d'|' -f3)
        if [[ "$payload" =~ $regex ]]; then
          printf 'pre-commit-secrets: %s:%d matches %s pattern (%s)\n' \
            "$current_file" "$current_line_new" "$(printf '%s' "$entry" | cut -d'|' -f1)" "$label" >&2
          violations=$((violations + 1))
        fi
      done
      ;;
    -*|"\\ No newline at end of file")
      ;;
  esac
done <<<"$DIFF"

if [ "$violations" -gt 0 ]; then
  echo "" >&2
  echo "pre-commit-secrets: refusing commit — remove the secret, rotate it, and try again." >&2
  echo "If this is a fixture, add the file path to ${ALLOWLIST_FILE} (one path per line)." >&2
  exit 1
fi

exit 0
