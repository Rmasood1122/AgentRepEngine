# L6 — ADVERSARIAL SESSION PROTOCOL
# APEX 11X Layer 6
# Date: April 7, 2026
# Version: 1.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## WHAT THIS IS

A dedicated Claude session whose only job is to destroy a decision
before you commit to it.

Not balanced analysis. Not devil's advocate. Not "here are some
considerations on both sides."

Destruction. Every assumption named. Every flaw exposed. Every
alternative identified. No hedging. No help.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## ACTIVATION COMMAND

  APEX ADVERSARIAL — [decision or claim to destroy]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## SYSTEM PROMPT — PASTE AT SESSION START FOR FULL ADVERSARIAL MODE

Use this as the system prompt when opening a fresh Claude session
dedicated to adversarial destruction:

---

You are a hostile expert reviewer. Your only job is to destroy the
decision, claim, or plan submitted to you.

Rules:
1. Find every flaw. Name it specifically. No vague concerns.
2. Name every assumption that, if wrong, causes the plan to fail.
3. Identify every alternative the person has not considered.
4. Do not help. Do not hedge. Do not soften.
5. If the plan survives your destruction, say so explicitly and state why.
6. If it does not survive, state the single most fatal flaw first.

Output format:
FATAL FLAW (if exists): [one sentence]
NAMED ASSUMPTIONS (list): [each assumption that is not verified fact]
SPECIFIC FLAWS (list): [each structural problem]
ALTERNATIVES IGNORED (list): [paths not considered]
VERDICT: SURVIVES / DOES NOT SURVIVE
If SURVIVES: state what would need to change to make it fail

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## WHEN TO RUN L6

Mandatory before:
  - Any pricing decision (pilot ACV, contract structure)
  - Any pivot or positioning change
  - Any enterprise pilot scope agreement
  - Any fundraising narrative before first investor meeting
  - Any competitive response before publishing

Optional but high-value before:
  - Major outreach message to a named buyer
  - Any claim made in a public document (Zenodo, GitHub README)
  - Enforce-mode activation decision

Not needed for:
  - Routine engineering tasks
  - Tactical message drafts
  - Operational decisions with clear reversibility

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## WHAT TO DO WITH THE OUTPUT

Accept the destruction. Do not argue with it.

Three outcomes:
  A) Fatal flaw found → revise the decision before committing
  B) No fatal flaw, but assumptions exposed → commit with explicit
     awareness of which assumptions carry the most risk
  C) Plan survives fully → commit with higher confidence, not certainty

In all cases: log the result in DECISION_AUDIT.md (L3) with the
confidence label adjusted based on L6 output.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## INTEGRATION WITH CLAIM CALIBRATION (L7)

Any [H] claim that survives L6 gets labeled [H-ADVERSARIAL].
This is a stronger signal than [H] alone.

Any [H] claim that was NOT run through L6 gets an implicit
asterisk: "adversarial validation not run."

Over time: track whether [H-ADVERSARIAL] claims verify at higher
rates than [H] claims. This is the calibration data.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## ARE-SPECIFIC HIGH-VALUE L6 TARGETS (current)

1. "Lloyd will sign a pilot LOI at or after the first meeting" [H]
2. "30-day observe mode with no procurement gauntlet is the right GTM" [H]
3. "Gen Digital ADR is not a direct threat to the Phase 1 pilot" [ASS]
4. "$50K–$150K ACV is achievable from first pilot" [H]
5. "Andrew Gyamfi at Translucent AI is a sitting buyer" [H]

Run L6 on each of these before Lloyd meeting.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L6 Protocol v1.0 | APEX 11X Layer 6 | April 7, 2026
