# DECISION_AUDIT.md — AgentRepEngine
# APEX 11X Layer 3 | Claim Calibration Foundation
# Date created: April 7, 2026
# Version: 1.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## PURPOSE

Track every significant decision with its confidence label and
predicted outcome. Verify at 30 days. Build epistemic accuracy score
over time.

This is the FP rate concept applied to judgment, not code.

Target: >70% of [H] claims verified correct within 90 days.
Baseline: unknown (this file establishes the baseline).

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## SCHEMA

| # | Date | Decision / Claim | Confidence | Predicted Outcome | Verify Date | Actual Outcome | Verdict |
|---|------|-----------------|------------|-------------------|-------------|----------------|---------|
| D001 | ... | ... | [H]/[BP]/[ASS]/[INF] | ... | +30 days | (fill at verify date) | CORRECT/WRONG/PARTIAL |

Confidence labels (APEX v5.2):
  [F]   = Fact — verified, measurable, reproducible
  [BP]  = Best practice — documented and widely accepted
  [H]   = Hypothesis — reasoned but not yet verified
  [A]   = Assumption — believed but not examined
  [OBS] = Observation — single data point, not generalized
  [INF] = Inference — conclusion from indirect evidence
  [ASS] = Assertion — stated without examination

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## DECISION LOG

| # | Date | Decision / Claim | Confidence | Predicted Outcome | Verify Date | Actual Outcome | Verdict |
|---|------|-----------------|------------|-------------------|-------------|----------------|---------|
| D001 | Apr 7, 2026 | Lloyd will engage meaningfully at first meeting (not just a courtesy call) | [H] | Meeting produces a next step: demo request, pilot scope discussion, or introduction to procurement | May 7, 2026 | | |
| D002 | Apr 7, 2026 | Andrew Gyamfi at Translucent AI is a sitting buyer (will engage on pilot) | [H] | Gyamfi responds to Boardy message with direct email and expresses interest in pilot | May 7, 2026 | | |
| D003 | Apr 7, 2026 | 30-day observe-mode pilot with no procurement gauntlet is the right GTM for regulated enterprise | [H] | At least one pilot converts to paid contract within 90 days of observe-mode start | Jul 7, 2026 | | |
| D004 | Apr 7, 2026 | Gen Digital ADR is a signal validator (proves market exists) not a fatal competitive threat to Phase 1 | [H] | ARE closes Phase 1 pilot despite ADR presence in market | Jul 7, 2026 | | |
| D005 | Apr 7, 2026 | Character Capital Labs G6 application will result in a positive decision | [H] | Acceptance or meaningful investor introduction from G6 process | Apr 13, 2026 | | |
| D006 | Apr 7, 2026 | $50K–$150K ACV is achievable from first pilot without a procurement gauntlet | [H] | First signed contract is within this range | Sep 7, 2026 | | |
| D007 | Apr 7, 2026 | Kong plugin as the deployment wedge (no SDK required) is a decisive GTM advantage vs competitors | [H] | Lloyd or first pilot buyer explicitly cites ease of deployment as a key decision factor | Jul 7, 2026 | | |
| D008 | Apr 7, 2026 | The behavioral grounding framing (RAG solved hallucination for outputs; ARE solves it for agent actions) resonates with CISOs | [H] | Lloyd or another CISO repeats or paraphrases this framing back unprompted | Jul 7, 2026 | | |
| D009 | Apr 7, 2026 | Publishing APEX v5.2 as a Zenodo paper creates a recruiting and credibility asset | [H] | Paper receives >100 views and generates at least one inbound contact within 60 days of publish | Jun 30, 2026 | | |
| D010 | Apr 7, 2026 | Andy Watkin-Child will convert from advisory intro to active DORA deal introducer | [H] | Watkin-Child makes at least one warm introduction to a DORA-regulated buyer within 60 days | Jun 7, 2026 | | |

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## CALIBRATION SCORE — RUNNING TOTAL

Last updated: April 7, 2026
Decisions logged: 10
Decisions verified: 0
Decisions due for verification: 0
Correct: — | Wrong: — | Partial: —
[H] accuracy rate: — (target: >70%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## MONTHLY VERIFICATION PROTOCOL

On the 7th of each month, Claude runs a verification pass:
  1. Pull all decisions with verify_date <= today
  2. For each: paste current state, Claude classifies CORRECT/WRONG/PARTIAL
  3. Update calibration score
  4. Identify systematic bias: are [H] claims skewing optimistic?
  5. Adjust confidence labeling calibration for next month

Command: APEX CALIBRATE → runs verification pass on this file

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## L6 ADVERSARIAL VALIDATION RECORD

Decisions run through L6 adversarial session (stronger signal):

| Decision # | L6 Date | Fatal Flaw Found | Label Upgrade |
|---|---|---|---|
| (none yet) | | | |

Decisions with [H-ADVERSARIAL] label are tracked separately.
Target: all D001–D008 run through L6 before Lloyd meeting.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DECISION_AUDIT v1.0 | APEX 11X Layer 3 | April 7, 2026
Update monthly. Commit to master after each verification pass.
