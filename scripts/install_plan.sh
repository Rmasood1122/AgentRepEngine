#!/usr/bin/env bash
# ARE MASTER PLAN INSTALL SCRIPT
# Run from repo root in Git Bash on Windows:
# cd /c/path/to/AgentRepEngine
# bash /path/to/ARE_PLAN/install_plan.sh
#
# What this does:
# 1. Creates all directory structure
# 2. Creates all Go file stubs
# 3. Creates documentation file stubs
# 4. Runs go build ./... to verify nothing is broken
# 5. Commits everything as a single "plan" commit
#
# DOES NOT: implement any logic. Stubs only. All TODOs intact.
# Run go test ./... after — all new tests will skip (not fail).

set -e

REPO_ROOT="$(pwd)"
echo "=== ARE MASTER PLAN INSTALL ==="
echo "Repo root: $REPO_ROOT"
echo ""

# Verify we are in the right repo
if [ ! -f "go.mod" ]; then
  echo "ERROR: go.mod not found. Run this script from the AgentRepEngine repo root."
  exit 1
fi

echo "go.mod found. Proceeding."
echo ""

# ============================================================
# DIRECTORY STRUCTURE
# ============================================================
echo "Creating directory structure..."

mkdir -p docs/competitive
mkdir -p docs/regulatory
mkdir -p docs/standards
mkdir -p docs/enterprise
mkdir -p docs/legal
mkdir -p docs/security
mkdir -p docs/demos
mkdir -p docs/research
mkdir -p docs/investor
mkdir -p internal/certification
mkdir -p internal/intelligence
mkdir -p internal/trust
mkdir -p internal/api
mkdir -p cmd/certify
mkdir -p cmd/certification-portal
mkdir -p web/certification
mkdir -p eval/attacks
mkdir -p data/norms
mkdir -p data/model_fingerprints
mkdir -p scripts

echo "Directories created."
echo ""

# ============================================================
# COPY SCRIPTS INTO REPO
# ============================================================
PLAN_DIR="$(dirname "$0")"

cp "$PLAN_DIR/sprint0_setup.sh" scripts/sprint0_setup.sh
cp "$PLAN_DIR/sprint1_setup.sh" scripts/sprint1_setup.sh
cp "$PLAN_DIR/sprint2_setup.sh" scripts/sprint2_setup.sh
cp "$PLAN_DIR/sprint3_setup.sh" scripts/sprint3_setup.sh
cp "$PLAN_DIR/ARE_PRODUCT_BUILD_ROADMAP.md" docs/ARE_PRODUCT_BUILD_ROADMAP.md

echo "Plan files copied into repo."
echo ""

# ============================================================
# GO FILE STUBS — run sprint scripts to create them
# ============================================================
echo "Creating Go file stubs via sprint scripts..."

bash scripts/sprint1_setup.sh
bash scripts/sprint2_setup.sh
bash scripts/sprint3_setup.sh

echo ""

# ============================================================
# DOCUMENTATION STUBS
# ============================================================
echo "Creating documentation stubs..."

# S0 docs — content to be written by Rehan
for f in \
  "docs/competitive/microsoft-response.md" \
  "docs/regulatory/dora-examiner-protocol.md" \
  "docs/standards/agentic-behavioral-certification-v0.1.md" \
  "docs/enterprise/trust-staircase.md" \
  "docs/legal/data-processing-agreement-template.md" \
  "docs/security/threat-model.md" \
  "docs/enterprise/fp-impact-analysis.md" \
  "docs/demos/microsoft-comparison-demo.md" \
  "docs/research/cross_framework_behavioral_consistency_v1.md" \
  "docs/investor/technical-due-diligence-package.md"
do
  if [ ! -f "$f" ]; then
    echo "# $(basename $f .md | tr '_-' '  ' | tr '[:lower:]' '[:upper:]')" > "$f"
    echo "" >> "$f"
    echo "TODO: See ARE_PRODUCT_BUILD_ROADMAP.md for content specification." >> "$f"
    echo "Created stub: $f"
  else
    echo "Exists (skipped): $f"
  fi
done

echo ""

# ============================================================
# BUILD VERIFICATION
# ============================================================
echo "=== Running go build ./... ==="
go build ./...
echo "Build: PASS"
echo ""

echo "=== Running go test ./... ==="
go test ./... 2>&1 | tail -20
echo ""

# ============================================================
# GIT COMMIT
# ============================================================
echo "=== Git status ==="
git status --short

echo ""
echo "=== Staging all new files ==="
git add .

echo ""
echo "Ready to commit. Run:"
echo "  git commit -m 'plan: ARE product build roadmap — sprints 0-4, Go stubs, doc stubs'"
echo "  git push origin master"
echo ""
echo "=== ARE MASTER PLAN INSTALL COMPLETE ==="
echo ""
echo "NEXT ACTION:"
echo "  Sprint 0 starts now — write docs in order:"
echo "  1. docs/competitive/microsoft-response.md      (3 hours)"
echo "  2. docs/regulatory/dora-examiner-protocol.md   (4 hours) → Zenodo"
echo "  3. docs/standards/agentic-behavioral-certification-v0.1.md (4 hours) → Zenodo"
echo "  4. Update docs/competitive/competitive-positioning.md"
echo ""
echo "  Then: Lloyd meeting April 7. LoU signed. Sprint 1 begins."
