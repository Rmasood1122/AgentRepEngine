# APEX_DECISION_GOVERNANCE_v1_0.md — AgentRepEngine
# The Five-Why Tribunal — Pre-Decision Architecture Protocol
# Version: 1.0
# Date: April 16, 2026
# Paired with: APEX v5.2 + ZROS v2.8 + INTEGRITY_SENTINEL v1.0
# Authority: Supersedes ad-hoc build decisions. Additive — never overrides APEX or ZROS.
# Activation command: APEX DECIDE — [component/change/deletion]
# Commit to: docs/ops/APEX_DECISION_GOVERNANCE_v1_0.md
# Upload to: Claude Project alongside CONTINUATION_PROMPT.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WHY THIS EXISTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ARE has ~5,587 lines of Go, a nuclear build stack, cryptographic audit
infrastructure, and a published DOI. Every component was built for a reason
rooted in regulatory positioning, enterprise trust, or moat architecture.

Deleting or changing anything without understanding that reason risks silently
destroying a capability that took weeks to build and cannot be recovered from
memory alone.

EXISTING GAP: ZROS L3 ("clarify before build") and the ZROS File Rule govern
individual files. Neither is sufficient for architectural-level decisions that
touch multiple components simultaneously, or that delete something whose
purpose may not be obvious from the code.

THIS PROTOCOL fills that gap. It is a formal pre-decision tribunal — not a
checklist, a structured multi-expert adversarial process — that fires before
any change to existing architecture.

TRIGGER: Any of these words in a session:
  "remove" / "replace" / "simplify" / "refactor" / "delete" / "change" /
  "add" / "swap out" / "upgrade" / "migrate"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 0 — ZROS FILE AUDIT (non-negotiable first step)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Before any deliberation begins:

  1. READ the file / function / component completely.
  2. Run: grep -r "[component name]" . — map every caller, every dependency.
  3. Run: git log --oneline [file] — trace the commit history.
  4. Locate the corresponding BUILD_INTELLIGENCE entry (Section 5).
  5. Locate the corresponding ZROS law (if one exists).
  6. State: "I have read this. I understand what it does. I understand who calls it."

RULE: No deliberation opens without Phase 0 complete.
A component you haven't fully read is a component you cannot evaluate.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 1 — THE FIVE WHYS TRIBUNAL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Five structured questions. Each answered by the relevant expert voice.
Minimum 3 expert voices must weigh in per question.

Expert panel:
  E1 — Security Architecture
  E2 — Distributed Systems / Scale
  E3 — AI Agent Systems
  E4 — Regulatory Compliance (HIPAA, SOC 2, FFIEC, NIST RMF)
  E5 — Adversarial / Solo Founder Coach
  E6 — Enterprise GTM / Category Definition
  E7 — Adversarial Pattern Detection (Ω1–Ω4)

────────────────────────────────────────────────────────────────────
WHY-1: WHY DOES THIS EXIST?
────────────────────────────────────────────────────────────────────

Question: What was the original problem this component was built to solve?
          Not what it does — what problem it answers.

E1 (Security): What threat does this component defend against?
               If removed, what attack surface opens?

E4 (Compliance): Which regulatory requirement does this satisfy?
                 HIPAA §164.312(b)? SOC 2 CC7.2? DORA Article 8(4)?
                 Which compliance CLI would break?

E6 (GTM): What does a CISO see when they look at this?
          Is this a differentiator in the Lloyd conversation?
          Would removing it create an objection?

VERDICT GATE:
  No expert can articulate the problem → label [ORPHAN] → candidate for
    removal after Phase 2 audit.
  Experts disagree → escalate to WHY-2.
  Clear answer found → label [PROBLEM-IDENTIFIED] → proceed to WHY-2.

────────────────────────────────────────────────────────────────────
WHY-2: WHY WAS IT BUILT THIS WAY?
────────────────────────────────────────────────────────────────────

Question: Why this implementation approach and not the obvious alternative?

PRIMARY SOURCE: BUILD_INTELLIGENCE.md Section 5 ("Architecture Decisions and Why").
Every major decision has a documented rationale. Read it before this question fires.

Reference decisions already documented:
  PostgreSQL queue → Kafka adds 4hr install, procurement reality
  YAML policy → OPA requires new language, readable in 5 min
  Synthetic 200 → 403 reveals deployment to adversary
  Fail-open → pilot survival, circuit breaker
  INSERT-only at DB → compromised app cannot lie to auditor

E3 (Distributed Systems): Is the original engineering constraint still true?
                          Has the environment changed in a way that invalidates it?

E7 (Solo Founder): Was this a deliberate Phase 1 choice with a Phase 2 upgrade path?
                   Is that path documented?

E5 (Adversarial): Name the strongest case the original rationale was wrong from day one.
                  Force-test whether it was a real constraint or a convenience assumption.

VERDICT GATE:
  Original rationale holds → label [RATIONALE-INTACT]
  Constraint has changed → label [RATIONALE-STALE] → proceed to WHY-3

────────────────────────────────────────────────────────────────────
WHY-3: WHAT IS THE LONG-TERM IMPLICATION?
────────────────────────────────────────────────────────────────────

Question: If we make this change, what does Phase 2 look like? Phase 3?
          What moat does it strengthen or weaken?

E2 (Scale): What happens to this component at 2,000 agents across 10 tenants?
            Does the proposed change hold at that scale or create a new problem?

E4 (Compliance): Does the change affect any compliance artifact?
                 Hash chain integrity? Reason object schema? OSCAL bundle structure?
                 Any change to these destroys customer-verifiable audit trails — the
                 core commercial moat.

E6 (Moat): Does this change make ARE harder or easier to replicate?
           The session-level behavioral history is architecturally impossible for
           per-call competitors to replicate. Any change that reduces this asymmetry
           requires explicit acknowledgment and justification.

E5 (Adversarial): What would a competitor do with this change if they learned about it?
                  Does it create an opening that didn't exist before?

VERDICT GATE (label required):
  [STRENGTHENS-MOAT]
  [NEUTRAL]
  [WEAKENS-MOAT]
  [DESTROYS-AUDIT-TRAIL]    ← immediate VERDICT D unless E4 + E1 both clear unanimously
  [REGULATORY-RISK]         ← immediate VERDICT D unless E4 clears explicitly

────────────────────────────────────────────────────────────────────
WHY-4: WHAT ARE THE REAL TRADE-OFFS?
────────────────────────────────────────────────────────────────────

Question: What do we gain? What do we lose?
          State both with equal precision.

RULE: The loss column must be as long as the gain column.
      If you can only articulate gains, you don't understand the trade-off yet.

Template (fill every line before proceeding):

  GAIN:                [specific, measurable, traceable to a goal]
  LOSS:                [specific, measurable, traceable to a risk]
  IRREVERSIBILITY:     [can this be undone in 24 hours if wrong? YES / NO]
  BLAST RADIUS:        [which other components are affected? list them]
  REGULATORY EXPOSURE: [does any compliance artifact change? YES / NO / which]
  DEMO INTEGRITY:      [does this touch any of the 10 demo steps? YES / NO]
  LLOYD-MEETING IMPACT:[does this change what we can claim on April 28? YES / NO]

E7 (Adversarial Coach): Name the Ω-pattern this decision might be hiding.
  Ω1 — engineering substituting for commercial action?
  Ω2 — premature completeness ("let me clean this up before the meeting")?
  Ω3 — three-roadmap complexity (solving a problem that doesn't exist yet)?
  Ω4 — language-action gap (planning the change instead of taking the simpler action)?
  If none apply: state "clean — no Ω pattern detected" explicitly.

VERDICT GATE:
  IRREVERSIBILITY = NO and BLAST RADIUS = HIGH → mandatory 48-hour hold.
    Write the decision in DECISION_AUDIT.md first. Sleep on it.
  Ω-pattern confirmed → VERDICT D until pattern is resolved.

────────────────────────────────────────────────────────────────────
WHY-5: WHAT IS THE SMALLEST SAFE VERSION OF THIS CHANGE?
────────────────────────────────────────────────────────────────────

Question: If we proceed, what is the minimum intervention that achieves the goal
          without touching the components that don't need to change?

ZROS L4: "The system gets lighter over time, not heavier."
Every change should either eliminate something or make something smaller.
If the proposed change adds complexity, it must justify that addition against
Phase 1 completion pressure.

E7 (Solo Founder): Can this be done in one function without touching the schema?
                   Can this be feature-flagged so it can be reverted without a rollback?
                   Can this be a new file rather than a modification to a critical path file?

E1 (Security): Does the minimal version still satisfy the security invariant
               that justified the original component?

E4 (Compliance): Does the minimal version still produce a valid compliance artifact?

VERDICT GATE:
  [MINIMAL-SAFE-PATH-EXISTS] → proceed with minimal path only
  [FULL-CHANGE-REQUIRED]     → DS-6 Excellence Gate fires before implementation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 2 — VERDICT PROTOCOL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

After all five whys, one of four verdicts is issued. No partial verdicts.

VERDICT A — PROCEED-AS-PROPOSED
  Condition: All 5 whys answered, no fatal flaw found, trade-offs documented.
  Action: Log in DECISION_AUDIT.md. Execute.

VERDICT B — PROCEED-MINIMAL
  Condition: Original goal valid, but proposed scope exceeds minimum safe path.
  Action: Define minimal safe path explicitly. Log in DECISION_AUDIT.md. Execute minimal only.
          If scope creep reappears during execution: stop and re-run WHY-5.

VERDICT C — DEFER
  Condition: Rationale intact, but this is Phase 2 work, or Lloyd meeting is imminent,
             or the change has no bearing on the immediate commercial gate.
  Action: Log in CONTINUATION_PROMPT.md backlog with reason.
          Do not touch. Close the discussion. No revisiting until Phase 2 opens.

VERDICT D — REJECT
  Condition: Destroys audit trail, weakens moat, no clear gain, or active Ω-pattern.
  Action: Log in DECISION_AUDIT.md as [REJECTED] with rationale.
          This is not failure. This is the system working.
          A rejected change that would have destroyed the hash chain moat is an S4 success.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 3 — COMMIT PROTOCOL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

If VERDICT A or B:

  1. git add -A && git status
     Verify blast radius matches prediction from WHY-4.
     If new files appear that weren't in the blast radius: STOP. Re-run WHY-4.

  2. Commit message format:
     ops: [decision summary — one line]
     Verdict: [A/B] | Blast: [components touched] | Reversible: [YES/NO]

  3. Run: go build ./... && go test ./...
     Tests must be GREEN before session closes. No exceptions.

  4. Update DECISION_AUDIT.md:
     Record actual outcome vs predicted.
     This is the calibration loop. Skip it and future decisions are less accurate.

  5. If any G-* gate was touched by this change:
     Re-verify the gate explicitly before closing.
     Gate verification is not optional even for small changes.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
STANDING RULES — NON-NEGOTIABLE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

RULE 1 — NO DELETE WITHOUT WHY-1 COMPLETE.
  If you cannot articulate what problem a component solves,
  you are not ready to delete it. Full stop.

RULE 2 — HASH CHAIN IS SACRED.
  Any change to internal/audit/, migrations/001_initial.sql,
  or the INSERT-only enforcement at PostgreSQL level is VERDICT D
  unless E4 and E1 both clear it unanimously and explicitly.
  The hash chain is the customer-verifiable audit trail.
  It is the single most defensible compliance artifact in the product.

RULE 3 — REASON OBJECT SCHEMA IS SACRED.
  docs/specs/reason_object_v1.json and internal/scoring/explainability.go
  are the SOC 2 artifact generators. Schema changes break existing customer
  audit trails. Any change requires E4 sign-off and a written migration plan
  before a single line is touched.

RULE 4 — DEMO INTEGRITY IS INVIOLABLE WITHIN 12 DAYS OF LLOYD.
  No change to any of the 10 demo steps, the pre-seeded baseline, or the
  enforcement decision flow before April 28, 2026. Period.
  If a "necessary" change is proposed: VERDICT C. It waits.

RULE 5 — Ω-PATTERN NAMING IS MANDATORY.
  Every decision session must end with an explicit statement:
  "This is [Ω1 / Ω2 / Ω3 / Ω4] — detected and blocked"
  OR
  "No Ω pattern — confirmed clean."
  Silence on this question is not acceptable.

RULE 6 — DECISION_AUDIT.md IS THE LIVING RECORD.
  Every verdict is logged. No exceptions. The log produces calibration.
  Calibration is what makes future decisions faster and more accurate.
  Skip the log and you pay compound interest in rework.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
INTEGRATION MAP — WHERE THIS FITS IN THE OS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  APEX v5.2 G-KILL          → fires before any NEW build work begins
  THIS PROTOCOL             → fires before any CHANGE to existing build
  ZROS L3                   → "clarify before build" is the spirit; this is the mechanism
  ZROS FILE RULE            → Phase 0 of this protocol (read before deliberate)
  DECISION_AUDIT.md         → receives every verdict from this protocol
  DS-6 EXCELLENCE GATE      → fires if WHY-5 returns [FULL-CHANGE-REQUIRED]
  SESSION_PROTOCOL A6       → adversarial gate; this protocol is A6 made operational
  INTEGRITY_SENTINEL        → verifies live system state before any change executes
  BUILD_INTELLIGENCE Sec 5  → primary source for WHY-2 (documented rationale)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ACTIVATION COMMAND
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  APEX DECIDE — [component / change description / deletion target]

Routes immediately to:
  Phase 0 (ZROS File Audit)
  → Phase 1 (Five Whys Tribunal: WHY-1 through WHY-5)
  → Phase 2 (Verdict: A / B / C / D)
  → Phase 3 (Commit Protocol, if A or B)

No change to existing architecture executes without this sequence.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VERSION HISTORY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

v1.0 — April 16, 2026
  Initial version. Established during STRATEGIC SESSION after product
  complexity milestone. Motivated by: nuclear build complete, Lloyd meeting
  April 28, need for decision-making process at same sophistication level
  as the product itself.
  Next update trigger: after first real APEX DECIDE tribunal fires and
  produces a lesson. Version 1.1 incorporates that lesson.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
APEX_DECISION_GOVERNANCE v1.0 | April 16, 2026
Source: APEX v5.2 (E1–E7), ZROS v2.8 (L1–L4, L12), APEX-OS v1.0 (DS-6),
        SESSION_PROTOCOL v1.0 (A6), BUILD_INTELLIGENCE v1 (Section 5)
File: docs/ops/APEX_DECISION_GOVERNANCE_v1_0.md
