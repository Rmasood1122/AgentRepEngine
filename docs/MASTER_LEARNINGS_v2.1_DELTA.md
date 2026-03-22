═══════════════════════════════════════════════════════════════════
MASTER_LEARNINGS_v2.1_DELTA — AGENTREPENGINE STRATEGIC INTELLIGENCE
Addendum to MASTER_LEARNINGS_v2.0
Version: 2.1 Delta (5 new learnings: L75–L79)
Source: March 22, 2026 build + adversarial meta-audit findings
Date: March 22, 2026
Apply to: All sessions. Supersedes conflicting assumptions in v2.0.
Instruction: Upload alongside MASTER_LEARNINGS_v2.0 in project.
             APEX references both documents. This addendum takes
             precedence over v2.0 where conflicts exist.
═══════════════════════════════════════════════════════════════════

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION H — BUILD LEARNINGS (L75–L79)
New learnings from March 22, 2026 hardening sprint
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L75 — AUTO-ROLLBACK IS GATE 0 BEFORE ENFORCE MODE [F]
ModeController must be initialized and StartFPMonitor() called
before any enterprise pilot switches to enforce mode.
Without it, a FP spike has no automatic protection.
Manual detection + manual rollback = 2–4 hours of legitimate
agents blocked in production. That ends the pilot.
With ModeController: FP spike detected in 5 minutes,
auto-rollback fires, SIEM alert fires, you know before the
client's security engineer does.
Implementation: internal/enforcement/mode_controller.go
Test: TestAutoRollback — seeds 30% FP rate, verifies rollback
Commit: V6 closed March 22, 2026
Action: Add to G-HARDEN gate. Check before every pilot go-live.

L76 — SIEM WEBHOOK BUILT BUT UNWIRED IS WORSE THAN NO SIEM [F]
A SIEMWebhook that exists in code but has zero callers gives
false confidence in enterprise conversations. The architecture
review passes, but the pilot day 1 reveals nothing fires.
Always verify after any new enforcement code:
  grep -rn "SendBlocked" internal/ | grep -v "siem.go"
Must return non-empty. If empty: wire it before claiming SIEM.
Implementation: store/score_store.go WriteScore() → SendBlocked()
Commit: H5 closed March 22, 2026
Action: Add to G-HARDEN gate. Run grep before every demo.

L77 — 0.00% FP ON SELF-AUTHORED CORPUS IS A CREDIBILITY RISK [F]
Never present "0.00% FP rate" as a hard fact in enterprise
conversations. A FAANG security engineer will immediately ask:
"Can we run your test on our traffic?" If the answer is no,
credibility drops more than a 1% FP rate would have.
Correct framing: "0.00% false positive rate on our 100-scenario
internal validation corpus, covering the 6 primary legitimate
workflow types. External validation on your production traffic
is available during the pilot — we expect <1% on a
well-configured environment."
The number stays. The framing changes.
Action: Update all materials. grep -r "0.00%" docs/ → reframe.
Never use "zero" — use "0.00% on internal corpus."

L78 — FAIL-OPEN VS FAIL-CLOSED MUST BE TWO DOCUMENTED
       DECISIONS WITH EXPLICIT RATIONALE [F]
Infrastructure failure → fail-open (availability preserved)
Enforcement failure → fail-closed (no unexplained blocks)
A CISO reading both behaviors without the rationale will
perceive contradiction and walk out. The meta-audit found
this as the single most dangerous finding in the entire audit.
Fix: Document BOTH behaviors with rationale in Operational
Safety Architecture BEFORE any enterprise conversation.
The distinction: infrastructure problems never stop your
business; enforcement problems never produce unexplained blocks.
Both are correct. Both need documented rationale.
Status: V2 closed March 22, 2026
File: docs/enterprise/operational-safety-architecture.md

L79 — THE ADVERSARIAL META-AUDIT IS MORE VALUABLE
       THAN THE AUDIT [F]
Running a second adversarial pass on your own answers finds:
  - Overconfident [F] labels (7 found in one audit)
  - Internal contradictions (4 found)
  - Rubric application errors (score 82 → corrected to 67)
  - Deal-killer gaps that the first pass missed
The honest score (67) after meta-audit was more valuable than
the claimed score (82). An enterprise CISO or acquisition
reviewer will apply the adversarial pass themselves. Do it
first and fix what they will find.
Action: Run adversarial meta-audit before every acquisition
conversation. Paste prior audit answers into a new session
with instruction: "Challenge every [F] label and find
internal contradictions." Fix what surfaces.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
STATUS UPDATES TO v2.0 LEARNINGS
Changes to prior learnings based on March 22 work
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L23 UPDATE — JWT REPLAY DETECTION: ✅ CLOSED March 22, 2026
jti claim on every token. Redis used-token cache TTL=exp.
Replay rejected on second use. 9/9 tests passing.
Commit: 570c41e

L24 UPDATE — REDIS ACL HARDENING: ✅ CLOSED March 22, 2026
AUTH password enforced. Default user disabled.
Unauthenticated access rejected. ACL restricts key namespaces.
Rate limiting 500 RPS per org in Kong.
Commit: 570c41e

L05 UPDATE — FP RATE FRAMING:
"Zero false positives on 100 enterprise scenarios" → RETIRED
New framing: "0.00% on our 100-scenario internal validation
corpus. External validation available during pilot."
See L77 for full rationale.

L11 UPDATE — THREE CISO DOCUMENTS: ✅ ALL WRITTEN March 21, 2026
Location: docs/enterprise/
- operational-safety-architecture.md ✅ (V2 fix applied)
- pilot-letter-of-understanding.md ✅
- honest-maturity-statement.md ✅ (score corrected to 67/100)
- prerequisites-checklist.md ✅ (H8 fix applied)
- gdpr-position.md ✅ (V3 fix applied)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HARDENING SCORE TRACKER
Current product score against FAANG enterprise rubric
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Baseline (adversarial meta-audit, March 21): 67/100
Current (March 22, 8 fixes applied):         87/100

Fixes applied:
  V2  Fail-open narrative documented              ✅ +3
  H8  Prerequisites checklist                     ✅ +2
  V7  Maturity score corrected + GDPR position    ✅ +2
  H5  SIEM webhook wired to enforcement           ✅ +3
  V6  Auto-rollback ModeController                ✅ +4
  H10 hash_chain_valid in /health endpoint        ✅ +1
  I4  NIST AI RMF mapping document                ✅ +2
  11X-2 APEX Laws → ATP/ATG mapping table         ✅ +2
  H2  Scoring model + ATP state mapping           ✅ +1

Remaining to FAANG-grade (88):
  11X-1 reason_object_v1.json schema              ☐ +1

Remaining to acquisition-ready (95):
  V4  Slow-walk evasion corpus                    ☐ +3
  V3  GDPR tombstone legal review                 ☐ +1
  V1  FP external validation                      ☐ +2
  11X-5 Call-level vs agent-level benchmark       ☐ +2

═══════════════════════════════════════════════════════════════════
END OF MASTER_LEARNINGS_v2.1_DELTA
5 new learnings (L75–L79) | 3 status updates | Score tracker
Version: 2.1 Delta | Date: March 22, 2026
Upload alongside MASTER_LEARNINGS_v2.0 in Claude project.
═══════════════════════════════════════════════════════════════════