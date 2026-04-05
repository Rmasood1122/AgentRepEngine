#!/usr/bin/env bash
# ARE Sprint 0 — Execute in Git Bash on Windows
# Run from repo root: cd /path/to/AgentRepEngine && bash scripts/sprint0_setup.sh

set -e

echo "=== ARE Sprint 0 Setup ==="
echo "Creating directory structure..."

# Create directories
mkdir -p docs/competitive
mkdir -p docs/regulatory
mkdir -p docs/standards
mkdir -p docs/enterprise
mkdir -p docs/legal
mkdir -p docs/security
mkdir -p docs/demos
mkdir -p docs/research
mkdir -p docs/investor

echo "Directories created."

# Verify existing competitive positioning file
if [ -f "docs/competitive/competitive-positioning.md" ]; then
  echo "competitive-positioning.md exists — will update Microsoft section"
else
  echo "WARNING: competitive-positioning.md not found — create it"
fi

echo ""
echo "=== Sprint 0 File Checklist ==="
echo "Create these files (content in ARE_PRODUCT_BUILD_ROADMAP.md):"
echo ""
echo "[ ] docs/competitive/microsoft-response.md          (3 hours)"
echo "[ ] docs/regulatory/dora-examiner-protocol.md       (4 hours) → Zenodo"
echo "[ ] docs/standards/agentic-behavioral-certification-v0.1.md  (4 hours) → Zenodo"
echo "[ ] docs/competitive/competitive-positioning.md     (1 hour, update existing)"
echo "[ ] docs/enterprise/trust-staircase.md              (3 hours)"
echo "[ ] docs/legal/data-processing-agreement-template.md (1 day)"
echo "[ ] docs/security/threat-model.md                   (1 day)"
echo "[ ] docs/enterprise/fp-impact-analysis.md           (3 hours)"
echo ""
echo "EXIT GATE: Lloyd LoU signed → run sprint1_setup.sh"
