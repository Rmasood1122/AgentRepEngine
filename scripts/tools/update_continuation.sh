#!/usr/bin/env bash
# scripts/tools/update_continuation.sh
# Auto-updates CONTINUATION_PROMPT.md on session close
#
# Usage:
#   ./scripts/tools/update_continuation.sh
#
# What it updates:
#   - HEAD commit hash
#   - Recent commits (last 10)
#   - Test status (runs go test ./...)
#   - Session date
#
# Does NOT update:
#   - Pipeline status (manual — requires human judgment)
#   - Active tasks (manual — requires human judgment)
#   - Commercial status (manual — requires human judgment)

set -euo pipefail

CONTINUATION_FILE="CONTINUATION_PROMPT.md"
SCORING_KEY="${SCORING_API_KEY:-are-internal-key-change-in-production}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ARE CONTINUATION_PROMPT Auto-Updater"
echo "  File: ${CONTINUATION_FILE}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ── Verify we are in the repo root ──────────────────────────────────────────
if [[ ! -f "$CONTINUATION_FILE" ]]; then
  echo "ERROR: ${CONTINUATION_FILE} not found. Run from repo root." >&2
  exit 1
fi

# ── Get current HEAD ─────────────────────────────────────────────────────────
HEAD=$(git rev-parse --short HEAD 2>/dev/null)
if [[ -z "$HEAD" ]]; then
  echo "ERROR: not a git repository" >&2
  exit 1
fi
echo "  HEAD: ${HEAD}"

# ── Get current date ─────────────────────────────────────────────────────────
SESSION_DATE=$(date +"%B %d, %Y")
echo "  Date: ${SESSION_DATE}"

# ── Run tests ────────────────────────────────────────────────────────────────
echo ""
echo "  Running go test ./... (this may take 30-60 seconds)..."
TEST_OUTPUT=$(SCORING_API_KEY="$SCORING_KEY" go test ./... 2>&1)
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^FAIL" || true)
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^ok" || true)

if [[ "$FAIL_COUNT" -eq 0 ]]; then
  TEST_STATUS="ALL GREEN ✅ (${PASS_COUNT} packages passing)"
  echo "  Tests: ${TEST_STATUS}"
else
  TEST_STATUS="❌ FAILING — ${FAIL_COUNT} packages failed"
  echo "  Tests: ${TEST_STATUS}"
  echo ""
  echo "  Failed packages:"
  echo "$TEST_OUTPUT" | grep "^FAIL" | sed 's/^/    /'
fi

# ── Get recent commits ───────────────────────────────────────────────────────
echo ""
echo "  Fetching recent commits..."
RECENT_COMMITS=$(git log --oneline -10 2>/dev/null)

# ── Update HEAD in CONTINUATION_PROMPT ──────────────────────────────────────
echo ""
echo "  Updating CONTINUATION_PROMPT.md..."

# Create backup
cp "$CONTINUATION_FILE" "${CONTINUATION_FILE}.bak"

# Use Python for reliable multi-line replacement on Windows Git Bash
python3 - <<PYTHON
import re

with open('${CONTINUATION_FILE}', 'r', encoding='utf-8') as f:
    content = f.read()

# Update HEAD commit hash
content = re.sub(
    r'HEAD: [0-9a-f]{7}',
    'HEAD: ${HEAD}',
    content
)

# Update test status line
content = re.sub(
    r'go test \./\.\.\. — .*',
    'go test ./... — ${TEST_STATUS}',
    content
)

# Update recent commits block
new_commits = """RECENT COMMITS
${RECENT_COMMITS}"""

content = re.sub(
    r'RECENT COMMITS\n(?:[0-9a-f]{7}.*\n){1,12}',
    new_commits + '\n',
    content
)

with open('${CONTINUATION_FILE}', 'w', encoding='utf-8') as f:
    f.write(content)

print('  ✅ CONTINUATION_PROMPT.md updated')
PYTHON

# ── Show diff ────────────────────────────────────────────────────────────────
echo ""
echo "  Changes made:"
diff "${CONTINUATION_FILE}.bak" "${CONTINUATION_FILE}" | grep "^[<>]" | head -20 || echo "  (no changes detected — file may already be current)"

# Clean up backup
rm -f "${CONTINUATION_FILE}.bak"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  DONE. Review changes then:"
echo "  git add CONTINUATION_PROMPT.md"
echo "  git commit -m \"docs: session close — HEAD ${HEAD}\""
echo "  git push origin master"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""