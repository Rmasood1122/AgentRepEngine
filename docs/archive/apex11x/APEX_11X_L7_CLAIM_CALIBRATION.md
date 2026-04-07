# L7 — CLAIM CALIBRATION LOOP
# APEX 11X Layer 7
# Date: April 7, 2026
# Version: 1.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## WHAT THIS IS

Every [H] claim in any ARE document gets a 30-day expiry tag.
APEX runs a monthly verification pass.
Track prediction accuracy rate across [H] claims.
Target: >70% of [H] claims verified correct within 90 days.

This is the FP rate concept applied to reasoning.
The [H] label is the claim. The expiry date is the measurement gate.
The verification is the ground truth.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## ACTIVATION COMMAND

  APEX CALIBRATE → runs verification pass on all expired [H] claims
  APEX H-AUDIT — [document name] → tags all [H] claims in that document

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## CORE [H] CLAIMS REGISTER — AGENTREPENGINE

The highest-stakes [H] claims currently active across ARE documents.
These are the claims that, if wrong, change the commercial or technical direction.

### COMMERCIAL [H] CLAIMS

| ID | Claim | Source Doc | Tagged | Expiry | Status |
|----|-------|-----------|--------|--------|--------|
| H-C01 | Lloyd meeting produces a next step (not a courtesy call) | CONTINUATION_PROMPT | Apr 7 | May 7 | OPEN |
| H-C02 | Gyamfi is a sitting buyer who will engage on a pilot | CONTINUATION_PROMPT | Apr 7 | May 7 | OPEN |
| H-C03 | 30-day observe mode with no procurement gauntlet closes regulated enterprise pilots | APEX v5.2 GTM | Apr 7 | Jul 7 | OPEN |
| H-C04 | Andy Watkin-Child converts to active DORA deal introducer | CONTINUATION_PROMPT | Apr 7 | Jun 7 | OPEN |
| H-C05 | $50K–$150K ACV is achievable without full procurement cycle | 24X_STRATEGY | Apr 7 | Sep 7 | OPEN |
| H-C06 | Behavioral grounding framing resonates with CISOs as a category-defining analogy | CLAUDE_MASTER | Apr 7 | Jul 7 | OPEN |
| H-C07 | Kong plugin (no SDK required) is a decisive GTM advantage vs ADR and Check Point | competitive docs | Apr 7 | Jul 7 | OPEN |
| H-C08 | Gen Digital ADR validates market existence, does not block Phase 1 close | gen-digital-adr-response.md | Apr 7 | Jul 7 | OPEN |

### TECHNICAL [H] CLAIMS

| ID | Claim | Source Doc | Tagged | Expiry | Status |
|----|-------|-----------|--------|--------|--------|
| H-T01 | 0.00% FP rate holds on enterprise production traffic (not just internal corpus) | CLAUDE_MASTER | Apr 7 | verify at pilot start | OPEN |
| H-T02 | 88% TP rate is sufficient to detect the attacks that matter to a CISO | CLAUDE_MASTER | Apr 7 | verify at pilot | OPEN |
| H-T03 | Slow-walk detection at 100% holds on real multi-agent environments (not just single-agent simulations) | ZROS-v2.7 | Apr 7 | verify at pilot | OPEN |
| H-T04 | <1ms Kong plugin overhead on enterprise production traffic | CLAUDE_MASTER | Apr 7 | verify at pilot | OPEN |
| H-T05 | Auto-rollback at 5-minute FP monitoring interval is fast enough to prevent pilot failure | ZROS-v2.7 | Apr 7 | verify at pilot | OPEN |

### INVESTOR / FUNDRAISING [H] CLAIMS

| ID | Claim | Source Doc | Tagged | Expiry | Status |
|----|-------|-----------|--------|--------|--------|
| H-I01 | Character Capital Labs G6 results in acceptance or meaningful investor introduction | CONTINUATION_PROMPT | Apr 7 | Apr 13 | OPEN |
| H-I02 | Publishing APEX v5.2 as Zenodo paper generates inbound credibility signal | APEX 11X L11 | Apr 7 | Jun 30 | OPEN |
| H-I03 | Pre-seed raise of $500K–$1M is achievable in parallel with Phase 1 pilot close | 24X_STRATEGY | Apr 7 | Sep 7 | OPEN |

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## CALIBRATION SCORE — RUNNING TOTAL

Last updated: April 7, 2026
[H] claims tagged: 16
[H] claims expired: 0
[H] claims verified: 0
Correct: — | Wrong: — | Partial: —
[H] accuracy rate: — (target: >70%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## MONTHLY VERIFICATION PROTOCOL

On the 7th of each month, Claude pulls all claims with expiry <= today.
For each expired claim:
  1. State what actually happened (paste context from CONTINUATION_PROMPT)
  2. Claude classifies: CORRECT / WRONG / PARTIAL
  3. If WRONG: what was the incorrect assumption?
  4. Update calibration score

After 3 monthly passes: identify systematic bias direction.
Are [H] claims skewing optimistic? Pessimistic? Domain-specific?
Adjust labeling behavior to close the gap.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## [H-ADVERSARIAL] UPGRADE PROTOCOL

Any [H] claim run through L6 adversarial session and surviving → upgrades to [H-ADVERSARIAL].
Track separately. Target: [H-ADVERSARIAL] accuracy rate should be >80% (vs >70% for [H]).

Claims currently at [H-ADVERSARIAL]: none yet.
Recommended for immediate L6 run: H-C01, H-C02, H-C03, H-T01.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L7 Protocol v1.0 | APEX 11X Layer 7 | April 7, 2026
Update monthly. This file is the epistemic accuracy instrument.
