# SPRINT_ROADMAP.md
# AgentRepEngine — Execution Roadmap
# Source: SUPREMACY v1.0 output + NWN reframe (April 6, 2026)
# Rule: This file is the task list. CONTINUATION_PROMPT.md is the state.
# Update this file when tasks complete. Commit on every change.
# Upload to Claude Project after every update.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SPRINT 0 — PRE-LoU (April 6–7) | GATE: Lloyd LoU
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

COMMERCIAL (DO FIRST — before any build)
[ ] Andrew Gyamfi — Boardy message for direct email            SITTING BUYER
[ ] Unmukt Raizada — LinkedIn direct                           PRIORITY
[ ] Andy Watkin-Child — LinkedIn direct                        PRIORITY
[ ] Lloyd — contact Monday April 7 to confirm meeting time
[ ] TW-REHEARSAL — say C1–C12 aloud before Lloyd meeting

SUPREMACY BUILD — Day 1 (6.5 hrs)
[ ] A1: internal/audit/regulatory_evidence.go                  2.5 hrs
    Framework: HIPAA §164.312 + SOX CC7.2 + FFIEC + NIST_RMF
    Test: internal/audit/regulatory_evidence_test.go
    Docs: docs/regulatory/hipaa-evidence-package-spec.md
          docs/regulatory/sox-evidence-package-spec.md
    Commit: "feat: US regulatory evidence package — HIPAA SOX FFIEC"
    G-FP: go test ./tests/regression/... GREEN before + after

[ ] A2: internal/identity/continuity.go                        2.5 hrs
    Migration: ALTER TABLE agent_baselines ADD COLUMN identity_seal
    Test: internal/identity/continuity_test.go
    Doc: docs/enterprise/agent-identity-continuity-hipaa.md
    Commit: "feat: agent identity continuity seal — HIPAA breach detection"
    G-FP: go test ./tests/regression/... GREEN before + after

[ ] A4: internal/audit/causal_chain.go                        2.0 hrs
    Test: internal/audit/causal_chain_test.go
    Doc: docs/regulatory/multi-agent-hipaa-attribution.md
    Commit: "feat: multi-agent causal chain — HIPAA breach attribution"
    G-FP: go test ./tests/regression/... GREEN before + after

SUPREMACY BUILD — Day 2 (5 hrs)
[ ] A3: internal/scoring/two_speed_baseline.go                 3.0 hrs
    Migration: ALTER TABLE agent_baselines ADD COLUMN fast_ewma
    Test: internal/scoring/two_speed_baseline_test.go
    Doc: docs/enterprise/override-fatigue-prevention.md
    Commit: "feat: two-speed baseline — ATTACK vs LEGITIMATE_CHANGE"
    G-FP: go test ./tests/regression/... GREEN before + after ← CRITICAL

[ ] A5: internal/api/dashboard_handler.go + static/dashboard.html  2.0 hrs
    Panels: HIPAA/SOX evidence | Identity continuity | Override fatigue | Causal chain
    Commit: "feat: NWN demo dashboard — HIPAA SOX panels for Lloyd meeting"
    G-FP: go test ./... GREEN

AUTOMATION (build alongside Day 1)
[ ] scripts/tools/session_monitor.py                           30 min
    Already built — copy from: scripts/tools/session_monitor.py
    Test: python session_monitor.py start → tick → status → close
[ ] scripts/tools/update_continuation.sh                       30 min
[ ] docs/ops/SPRINT_ROADMAP.md (this file)                     ✅ DONE

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SPRINT 1 — POST-LoU (after Lloyd LoU signed) | GATE: First pilot live
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

BLOCKED UNTIL: Lloyd LoU signed

[ ] cmd/dora-verify/main.go — DORA-formatted verify wrapper     2 hrs
[ ] certification/report.go Generate()                          per L9
[ ] scripts/update_continuation.sh full automation              2 hrs
[ ] Vendor package: MSA + pilot scope + data brief              4 hrs
[ ] FP corpus expansion: 100 → 300 scenarios                    4 hrs
[ ] Helm chart (only if NWN confirms Kubernetes)               conditional

COMMERCIAL
[ ] Andrew Gyamfi — schedule pilot conversation
[ ] Unmukt Raizada — finserv validation call
[ ] Andy Watkin-Child — DORA buyer introductions
[ ] BNY Mellon outreach (Kuntal Dutta — AIAI June 4)
[ ] Character Capital Labs G6 — follow up (~April 13)
[ ] Leaders in AI Summit NYC — April 21–22 prep
[ ] Momentum AI NYC — April 27–28 prep
[ ] Turner Novak — verify domain then email
[ ] Cassandra Mack (TensorWave) — schedule meeting
[ ] Secondary investor list — 10 names

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SPRINT 2 — POST-PILOT-GO-LIVE | GATE: First incident documented
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

BLOCKED UNTIL: First pilot enterprise live in observe mode

[ ] Two-speed baseline tuning against real traffic data
[ ] Competitive response update (monitor Gen Digital ADR monthly)
[ ] DORA evidence package (EU pipeline — Watkin-Child, BNY Mellon)
[ ] Agent supply chain risk scoring (requires 3+ deployments)
[ ] Industry norm benchmarking (requires 3+ deployments)
[ ] Series A data room preparation
[ ] NYC professor visits (Cornell Tech, NYU Tandon, Columbia)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ANTI-SCOPE — never appear in Sprint 0 or Sprint 1
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ Kafka
❌ OPA
❌ Isolation Forest
❌ Federation / multi-tenant before single-tenant validated
❌ DID/ledger
❌ CI/CD before runtime enforcement proven
❌ Phase 2–4 work while Phase 1 commercial gate open

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SPRINT STATUS LOG
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

| Date       | Sprint | Completed                    | Gate Status        |
|------------|--------|------------------------------|--------------------|
| Apr 6 2026 | 0      | Roadmap created              | Lloyd meeting TBD  |
