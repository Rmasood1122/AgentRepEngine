# L9 — CLAUDE CODE AUTOMATION TARGETS
# APEX 11X Layer 9
# Date: April 7, 2026
# Version: 1.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## OBJECTIVE

Eliminate 4–6 hrs/week of manual Claude overhead through automation.
Target: the mechanical, repetitive work that currently burns session
time that should go to engineering or commercial activity.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## AUTOMATION TARGETS — PRIORITY ORDER

### TARGET 1 — CONTINUATION_PROMPT.md auto-update (2 hrs build, 30 min/week saved)

Current: Manually written at end of every session. 15–30 minutes.
          Risk of forgetting, inconsistent format, missing items.

Automation: Script reads `git log --oneline -20`, parses commit prefixes,
            extracts feature/fix counts for ZROS rework rate,
            updates the COMMITS table, HEAD hash, and test status automatically.

Script: scripts/update_continuation.sh
Input: git log + go test ./... output
Output: Updated CONTINUATION_PROMPT.md CURRENT STATE section

Build when: After Lloyd meeting (Sprint 1 start)
Estimate: 2 hrs Python/bash

---

### TARGET 2 — Competitive audit scraping (3 hrs build, 1 hr/week saved)

Current: Manual Google searches for Gen Digital ADR updates, Check Point
          Lakera integration news, new entrants. Ad hoc, easily skipped.

Automation: Python script runs weekly, scrapes:
  - Gen Digital ADR product page and changelog
  - Check Point press releases filtered for "AI" and "agent"
  - GitHub trending repos for "agent security" / "AI agent trust"
  - Zenodo new publications: agent reputation, AI governance, runtime enforcement

Output: competitive_digest_YYYY-MM-DD.md in docs/competitive/
Claude reads the digest at session start and flags material changes.

Build when: After Lloyd meeting (low urgency, high compound value)
Estimate: 3 hrs Python + cron

---

### TARGET 3 — [H] claim expiry scanner (1 hr build, 30 min/month saved)

Current: L7 CLAIM_CALIBRATION.md requires manual review each month.
          Easy to skip. Claims expire unverified.

Automation: Script reads APEX_11X_L7_CLAIM_CALIBRATION.md,
            finds all claims with expiry <= today,
            outputs a verification prompt pre-formatted for Claude session.

Script: scripts/check_claims.sh
Output: "APEX CALIBRATE — paste this into session" with expired claims pre-pulled

Build when: This week (30 min build, immediate compound value)
Estimate: 30 min bash

---

### TARGET 4 — Weekly drift detection prompt generator (30 min build)

Current: L4 requires pasting last 7 days of actions manually.
          Requires remembering to run it on Mondays.

Automation: Script reads CONTINUATION_PROMPT.md (last session action log)
            + git log --since="7 days ago" --oneline
            + DECISION_AUDIT.md (decisions made this week)
            Outputs pre-formatted APEX DRIFT prompt ready to paste.

Script: scripts/weekly_drift.sh
Run: Every Monday via cron or manual trigger
Estimate: 30 min bash

---

### TARGET 5 — Doc generation pipeline for buyer materials (4 hrs build)

Current: Vendor package (MSA, pilot scope, data brief) not built.
          Building it manually will take 3–4 hrs of session time.

Automation: Template + data file approach.
            YAML config: buyer_name, environment, regulatory_framework, pilot_scope
            Script generates: pilot_scope_letter.md, data_brief.md, msa_template.md
            Output is paste-ready or converts to PDF via existing pdf skill.

Build when: Before Lloyd meeting if time allows; otherwise Sprint 1
Estimate: 4 hrs (high value — blocks vendor package item from CONTINUATION_PROMPT)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## BUILD SEQUENCE

Week of April 7:  TARGET 3 (30 min — immediate L7 compound value)
                  TARGET 4 (30 min — immediate L4 compound value)
Sprint 1 start:   TARGET 1 (2 hrs — CONTINUATION_PROMPT automation)
Month 2:          TARGET 2 (competitive scraping)
                  TARGET 5 (doc generation pipeline)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## CONSTRAINT — SCOPE DISCIPLINE

All scripts are utilities, not product features.
None of these go in the ARE main repo under a product directory.
Location: scripts/tools/ (separate from scripts/demo.sh and product scripts)
They are founder infrastructure, not customer-facing.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L9 Automation Targets v1.0 | APEX 11X Layer 9 | April 7, 2026
