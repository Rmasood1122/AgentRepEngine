# APEX_v6_PROPOSALS.md — AgentRepEngine
# Self-Improvement Proposal Log
# Version: 1.0 | Started: April 16, 2026
# Authority: Proposals only. Nothing executes without explicit human approval.
# Commit to: docs/ops/APEX_v6_PROPOSALS.md
# Upload to: Claude Project when proposals exist

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HOW THIS WORKS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

At every session close, Claude runs APEX PROPOSE (5 minutes).
0–3 proposals generated per session based on what actually happened.
Each proposal is presented here for Rehan's verdict.
Nothing in the OS changes without APPROVED verdict.

PROPOSAL TYPES:
  LAW         → new ZROS law from a real incident
  GATE        → new gate to stop a recurring pattern
  ENHANCEMENT → improvement to existing mechanism
  DEPRECATION → gate/law that hasn't fired in 30 days
  PROPAGATION → cross-file update needed

VERDICT OPTIONS:
  APPROVED  → Claude implements in current or next session
  DEFERRED  → revisit next month (3 deferrals = force APPROVE or REJECT)
  REJECTED  → closed permanently with reason logged

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PROPOSAL LOG
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

═══════════════════════════════════════════════════════════════════
SESSION: April 16, 2026 | MOMENTUM: [first session — baseline TBD]
═══════════════════════════════════════════════════════════════════

P001 — Add APEX v6 STATE block to CONTINUATION_PROMPT
TYPE: PROPAGATION
TRIGGER: APEX v6 built this session. CONTINUATION_PROMPT needs MVW,
         momentum, ROT accuracy, proposals pending fields.
CHANGE: Add Part 4 block from APEX_v6_SELF_IMPROVING_OS.md to
        CONTINUATION_PROMPT CURRENT STATE section.
FILE: CONTINUATION_PROMPT.md
EFFORT: 10 min
VERDICT: ________ by Rehan on ________

P002 — Create APEX PEAK checklist entry in ARE_BULLETPROOF
TYPE: ENHANCEMENT
TRIGGER: APEX PEAK defined in v6. Lloyd meeting April 28.
         BULLETPROOF RUN fires April 27. PEAK should fire after it.
CHANGE: Add 5-item APEX PEAK checklist to end of ARE_BULLETPROOF_v1_0.md
        under section "AFTER BULLETPROOF RUN — APEX PEAK"
FILE: ARE_BULLETPROOF_v1_0.md
EFFORT: 15 min
VERDICT: ________ by Rehan on ________

P003 — Add monthly APEX-OS CALIBRATE gate to G-COMMERCIAL
TYPE: GATE
TRIGGER: Self-improvement engine requires monthly calibration.
         Without a gate, it will be skipped. Gates are the only
         mechanism that has proven reliable in this OS.
CHANGE: Add one line to G-COMMERCIAL in CONTINUATION_PROMPT:
        "[ ] APEX-OS CALIBRATE run this month? (fires 1st of month only)"
FILE: CONTINUATION_PROMPT.md
EFFORT: 2 min
VERDICT: ________ by Rehan on ________

SESSION PROPOSALS: 3 | Approved: 0 | Deferred: 0 | Rejected: 0
[First session — verdicts pending]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CUMULATIVE STATS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total proposals: 3
Approved: 0 | Deferred: 0 | Rejected: 0 | Pending: 3
Approval rate: N/A (first session)
Most common type: PROPAGATION (1) / GATE (1) / ENHANCEMENT (1)
Next review: May 1, 2026 (APEX v6 CALIBRATE)
