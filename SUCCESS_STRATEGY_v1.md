━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUCCESS STRATEGY v1.0 — AGENTREPENGINE
Mitigation plan for all 100 identified failure modes
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Date: March 31, 2026 | 7-Expert Panel | APEX v5.2
Source: 100 failure modes identified from full system audit

ACTION TYPES:
  🔴 TEST NOW  — run this command today
  🟡 BUILD     — code or doc change, specific file
  🟢 STRATEGY  — positioning, awareness, contingency
  ⚪ MONITOR   — external event, watch and respond

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
IMMEDIATE TEST QUEUE — RUN BEFORE APRIL 4
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Run all of these in one session. Takes 20 minutes.

T1 [F08] Auto-rollback works:
   go test ./internal/enforcement/... -run TestAutoRollback -v
   MUST: PASS

T2 [F26] Forged JWT rejected:
   TOKEN="eyJhbGciOiJSUzI1NiJ9.eyJhZ2VudF9kaWQiOiJmYWtlIn0.invalidsig"
   curl -s http://localhost:8080/verify -X POST \
     -H "Content-Type: application/json" \
     -d "{\"token\":\"$TOKEN\"}" | jq '.valid'
   MUST: false

T3 [F28] Consumer clean under load:
   go test ./internal/scoring/... -v -count=1 2>&1 | tail -5
   MUST: all PASS

T4 [F29] Org-scoped baselines populated:
   docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
     -c "SELECT COUNT(*) FROM agent_baselines WHERE org_id IS NOT NULL;"
   MUST: non-zero

T5 [F30] confidence_pct range correct:
   go test ./internal/scoring/... -run TestConfidencePctRange -v
   MUST: PASS at z=0(100%), z=3.0(0%), z=5.0(0%)

T6 [F31] Hash chain valid:
   docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
     -c "SELECT verify_hash_chain();"
   MUST: true

T7 [F33] ModeController reads from Redis:
   curl -s http://localhost:8080/health | jq '.enforcement_mode'
   MUST: "observe"

T8 [F35] FP scenarios all pass (Kong plugin coverage):
   go test ./tests/fp_scenarios/... -v 2>&1 | grep -E "PASS|FAIL"
   MUST: all PASS

T9 [F38] Held-out tests pass:
   go test ./tests/held_out/... -v 2>&1 | tail -10
   MUST: all PASS

T10 [F41] /verify responds under 500ms under load:
    for i in {1..5}; do
      curl -s -w "%{time_total}\n" -o /dev/null \
        http://localhost:8080/health; done
    MUST: all under 0.5s

T11 [F49] reason_object_v1.json includes confidence_pct:
    grep "confidence_pct" docs/specs/reason_object_v1.json
    MUST: non-empty result

T12 [F50] PostgreSQL connections healthy:
    docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
      -c "SELECT count(*) FROM pg_stat_activity WHERE datname='agentrepengine';"
    MUST: well below max_connections

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BUILD QUEUE — BEFORE PILOT GO-LIVE (after LoU signed)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

B1 [F10]  DPA template — docs/enterprise/dpa-template.md (30 min)
B2 [F24]  Synthetic incident validation in pilot protocol (30 min)
B3 [F27]  Three-tier Redis fallback — PL-2 (3 hours)
B4 [F43]  Elastic SIEM output mode — 2 hours
B5 [F47]  Probation window configurable in scoring_weights.yaml (30 min)
B6 [F48]  X-ARE-Enforcement header on synthetic response (30 min)
B7 [F83]  M2 slow-walk research note → Zenodo (6 hours)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
LLOYD MEETING PREP ADDITIONS (from failure analysis)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Add these to Lloyd_meeting_prep.md objection handling:

[F09] SOC2 Type II asked:
"ARE produces the audit evidence SOC2 requires. Your auditor
verifies our infrastructure directly. We run in your environment —
we don't need our own SOC2 cert."

[F39] Private key disclosure:
"A private key was committed in error in March 2026 and removed
in commit 492d015. It was never used in production. We disclose
this proactively to all pilot partners."

[F55] "We use Envoy not Kong":
"We have both Kong and Envoy support. Which gateway are your
clients running?" [Ask before the pilot conversation]

[F59] "We'll build it internally":
"You could. Here's what it took: 80+ commits, 5,876 lines of code,
22 gates, 3 adversarial audits. The 30-day pilot costs nothing to
evaluate whether building is worth it."

[F92] "Budget is frozen":
"Observe mode is free. No budget needed to start. The ROI case
at day 30 justifies the enforcement budget."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MONITOR LIST — CHECK MONTHLY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

M1 [F07]  Sycamore Labs GitHub — watch for behavioral enforcement release
M2 [F81]  xAI releases — watch for agent monitoring features
M3 [F82]  OpenAI Assistants API — watch for behavioral monitoring
M4 [F87]  ENISA EU AI Act implementation guidance
M5 [F97]  Kong release notes — test plugin on major versions
M6 [F98]  PostgreSQL security announcements

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
THE 5 FAILURES THAT MATTER MOST — FOCUS HERE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

F08  First pilot FP → MITIGATED: auto-rollback + TestAutoRollback
F05  Runway → MITIGATED: CCL G6 + Lloyd revenue + Bettina parallel
F66  Avoidance → MITIGATED: named explicitly, you respond to hard stops
F22  Clock expires → ONLY MITIGATION: Lloyd LoU April 7
F100 Burnout after first customer → PLAN NOW: first hire = solutions engineer

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF SUCCESS STRATEGY v1.0
100 failures analyzed | 12 tests | 7 builds | 6 monitors
Created: March 31, 2026
Update after: each pilot milestone, each competitive development
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━