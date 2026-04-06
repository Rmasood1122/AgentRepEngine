# L5 — PARALLEL SESSION ARCHITECTURE
# APEX 11X Layer 5
# Date: April 7, 2026
# Version: 1.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## THE STRUCTURAL PROBLEM

One sequential Claude session handles engineering + commercial + strategy.
Each domain contaminates the others. Engineering sessions drift into strategy.
Strategy sessions drift into meta-work. Commercial sessions drift into
competitive research instead of contact.

The result: 3 domains competing for one context window.
No domain gets full depth. All three get diluted output.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## THE ARCHITECTURE

Three specialist sessions run in parallel (separate browser tabs or windows).
Each session has a hard scope rule. No cross-contamination.
A master session synthesizes at the end.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## SESSION SPECS

### SESSION E — ENGINEERING
Scope: Go code, Kong plugin, test suite, infrastructure, debugging
Allowed: Everything in the ARE repo that requires `go build ./...`
Hard stop if: Any commercial, GTM, investor, or positioning discussion appears
System context: APEX v5.2 + ZROS v2.7 + CLAUDE_MASTER v2.0 + CONTINUATION_PROMPT

Activation: "You are in ARE engineering session. Scope is strictly technical.
If I raise a commercial or strategic topic, refuse it and redirect."

### SESSION C — COMMERCIAL
Scope: Outreach messages, meeting prep, pilot scope, buyer conversations
Allowed: Named contacts, message drafts, Lloyd prep, pilot materials, vendor package
Hard stop if: Any engineering, architecture, or product spec discussion appears
System context: APEX v5.2 + CONTINUATION_PROMPT (commercial pipeline section only)

Activation: "You are in ARE commercial session. Scope is strictly outreach
and buyer conversation. If I raise a technical or architectural topic, refuse it."

### SESSION S — STRATEGY
Scope: Positioning, fundraising narrative, competitive response, investor targeting
Allowed: Everything that informs decisions but produces no code and sends no message
Hard stop if: Scope creep into engineering tasks or commercial message drafting
System context: APEX v5.2 + 24X_STRATEGY + MASTER_LEARNINGS

Activation: "You are in ARE strategy session. Scope is analysis and positioning.
No code. No message drafts. Pure decision support."

### SESSION M — MASTER (synthesis only)
Scope: Synthesizes outputs from E + C + S sessions
Runs at end of parallel session block
Input: Paste key outputs from each specialist session
Output: Updated CONTINUATION_PROMPT, next session priorities, DECISION_AUDIT entries

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## WHEN TO RUN PARALLEL SESSIONS

Run parallel architecture when:
  - Work block is 3+ hours
  - Both engineering and commercial work need to happen in the same block
  - A major strategic decision needs analysis before commitment
  - Lloyd meeting prep requires simultaneous demo hardening + message drafting

Run sequential (single session) when:
  - Work block is under 90 minutes
  - Work is clearly in one domain only
  - Context is fresh from previous session

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## CROSS-CONTAMINATION RULES

These topics are banned from Session E:
  - Pricing, ACV, pilot scope decisions
  - Investor names, fundraising narrative
  - Competitor positioning
  - Outreach message content

These topics are banned from Session C:
  - Go code, test output, architectural decisions
  - Infrastructure topology
  - Performance benchmarks beyond what's needed for a buyer claim

These topics are banned from Session S:
  - Code-level implementation decisions
  - Specific message drafts
  - Meeting logistics

Contamination = context dilution = lower quality output in both domains.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## THROUGHPUT CALCULATION

Sequential single session (current):
  90 min engineering + 60 min commercial + 30 min strategy = 180 min
  Effective depth per domain: ~40% (contamination + context switching)

Parallel three sessions:
  90 min engineering (full depth) || 60 min commercial (full depth) || 30 min strategy
  Elapsed time: 90 min (longest session)
  Effective depth per domain: ~85%
  Output quality multiplier: ~3x on complex multi-domain blocks

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## FIRST USE CASE — LLOYD MEETING PREP (this week)

Run L5 parallel architecture for Lloyd prep:

SESSION E: Demo hardening — run scripts/demo.sh, verify all 4 metrics, verify claims C1–C12
SESSION C: Draft Lloyd follow-up message + pilot scope one-pager
SESSION S: L6 adversarial on H-C01 (Lloyd meeting hypothesis) + pricing decision
SESSION M: Synthesize → updated Lloyd_meeting_prep.md + CONTINUATION_PROMPT

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L5 Architecture v1.0 | APEX 11X Layer 5 | April 7, 2026
