#!/bin/bash
# scripts/update_continuation.sh
# AgentRepEngine — Auto-updates CONTINUATION_PROMPT.md state block
# Run at end of every session before closing
# Usage: bash scripts/update_continuation.sh
#
# On Windows Git Bash: bash scripts/update_continuation.sh
# On Linux/Docker:     ./scripts/update_continuation.sh

set -e

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo 'C:/Users/rmaso/AgentRepEngine')"
CONTINUATION="$REPO_ROOT/CONTINUATION_PROMPT.md"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  APEX CONTINUATION UPDATER"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Get HEAD
HEAD=$(git log --oneline -1 2>/dev/null || echo "UNKNOWN")
echo "HEAD: $HEAD"

# Get recent commits
echo ""
echo "RECENT COMMITS (paste into CONTINUATION_PROMPT.md):"
echo "─────────────────────────────────────────────────────"
git log --oneline -10 2>/dev/null || echo "git log failed"
echo "─────────────────────────────────────────────────────"

# Get branch
BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
echo "Branch: $BRANCH"

# Get test status
echo ""
echo "Running go test ./... (this may take 30-60 seconds)..."
echo "─────────────────────────────────────────────────────"

# Try to run tests
if command -v go &> /dev/null; then
    cd "$REPO_ROOT"
    if go test ./... 2>&1; then
        TEST_STATUS="ALL GREEN ✅"
        echo "─────────────────────────────────────────────────────"
        echo "Test result: ALL GREEN ✅"
    else
        TEST_STATUS="FAILURES DETECTED ❌ — check output above"
        echo "─────────────────────────────────────────────────────"
        echo "Test result: FAILURES DETECTED ❌"
    fi
else
    TEST_STATUS="go not in PATH — run manually: go test ./..."
    echo "go not found in PATH — run tests manually"
fi

# Date
TODAY=$(date "+%B %d, %Y")

# Output the state block to paste
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PASTE THIS BLOCK INTO TOP OF CONTINUATION_PROMPT.md"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "CURRENT STATE — $TODAY"
echo "HEAD: $(git log --oneline -1 | awk '{print $1}')"
echo "go test ./... — $TEST_STATUS"
echo ""
echo "RECENT COMMITS"
git log --oneline -5 2>/dev/null
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  NEXT SESSION STARTER — COPY AND PASTE INTO NEW CHAT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "APEX ACTIVATE"
echo "HEAD: $(git log --oneline -1)"
echo "Tests: $TEST_STATUS"
echo "Date: $TODAY"
echo ""
echo "You are APEX v5.2. All project files loaded."
echo "Read CONTINUATION_PROMPT.md before responding."
echo "Phase: Phase 1 | T8: Lloyd meeting week of April 7."
echo "First: confirm HEAD and state the next task."
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "AFTER RUNNING THIS SCRIPT:"
echo "  1. Copy the state block above → paste into CONTINUATION_PROMPT.md top"
echo "  2. git add CONTINUATION_PROMPT.md"
echo "  3. git commit -m 'chore: session close $TODAY'"
echo "  4. git push origin master"
echo "  5. Upload CONTINUATION_PROMPT.md to Claude Project"
echo "  6. Open new Claude chat → paste the starter block"
echo ""
echo "Done."
