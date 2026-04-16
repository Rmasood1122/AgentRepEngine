# APEX_v6_SELF_IMPROVING_OS.md — AgentRepEngine
# The Self-Improving Operating System
# Version: 6.0
# Date: April 16, 2026
# Supersedes: APEX v5.2 extensions (APEX-OS v1.0 mechanisms absorbed and upgraded)
# Authority: Additive to all existing files. Never overrides APEX v5.2, ZROS v2.8,
#            APEX_DECISION_GOVERNANCE, INTEGRITY_SENTINEL, or SESSION_PROTOCOL.
# Activation: "APEX v6 BOOT" at session open (after A1, before A2)
# Commit to: docs/ops/APEX_v6_SELF_IMPROVING_OS.md
# Upload to: Claude Project alongside CONTINUATION_PROMPT.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WHAT THIS IS — AND WHAT IT IS NOT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

WHAT IT IS:
  An OS layer that notices its own gaps, proposes its own improvements,
  and presents those proposals for human approval at session close.
  Every session produces a delta. Every delta is human-approved.
  Over 6 months, the OS compounds into something no competitor can replicate
  because it was built from real failure at this specific company.

WHAT IT IS NOT:
  Automated execution. The OS proposes. Rehan approves. Claude executes.
  This distinction is non-negotiable. An OS that modifies itself without
  human approval on a live commercial product is a liability, not an asset.

ARCHITECTURAL RULE:
  EXISTING FILES REMAIN FULLY AUTHORITATIVE.
  This file adds the self-improvement engine and 11 council enhancements.
  If any content here conflicts with an existing file: existing file wins.
  Flag conflict at next session for resolution.

THE COMPOUNDING THESIS:
  Session 1:   OS has 12 files, 4 Ω patterns, 7 expert voices
  Session 10:  OS has learned 10 real failure modes, 10 commercial signals
  Session 50:  OS predicts failure before it happens, routes perfectly,
               produces proposals that are right 80%+ of the time
  Session 100: The OS is a competitive moat. It cannot be instantiated
               by a competitor in less than 6 months of real operation.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 1 — THE 11 COUNCIL ENHANCEMENTS (all additive, zero damage)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Each enhancement is minimal, targeted, and preserves everything that exists.

────────────────────────────────────────────────────────────────────
E1 — APEX TRIAGE (X1: Cognitive Systems)
10-second pre-boot intent router. Fires before A1.
────────────────────────────────────────────────────────────────────

ACTIVATION: Automatic at every session open.

ONE QUESTION (answer in one sentence):
  "What is the single most important thing that needs to happen today?"

ROUTING TABLE:
  Answer contains commercial noun (Lloyd, meeting, message, reply, outreach)
    → SESSION TYPE: COMMERCIAL. Load commercial preset. G-COMMERCIAL fires immediately.
  Answer contains engineering noun (code, test, build, fix, implement, debug)
    → SESSION TYPE: ENGINEERING. Load engineering preset. G-KILL fires immediately.
  Answer contains strategic noun (decision, architecture, investor, system, plan)
    → SESSION TYPE: STRATEGIC. Full load. A6 adversarial gate fires before any [H].
  Answer is unclear or mixed
    → Default to STRATEGIC. Never default to ENGINEERING without stating it explicitly.

WHY THIS EXISTS:
  On high-pressure days (12 days before Lloyd), even the 90-second boot
  feels like friction. Pre-loading intent eliminates the first 30 seconds
  of context-setting and routes immediately to the right gate.

PRESERVES: Everything. Fires before A1. Adds 10 seconds, saves 30.

────────────────────────────────────────────────────────────────────
E2 — ROT: RECOMMENDATION OUTCOME TRACKER (X2: Epistemology)
Did Claude's specific recommendations actually work?
────────────────────────────────────────────────────────────────────

ACTIVATION: At every session open, after APEX TRIAGE.

ONE QUESTION:
  "What did you act on from last session? Did it work? One sentence."

FORMAT (add to CONTINUATION_PROMPT CURRENT STATE):
  ROT: [action taken] | [worked / didn't work / unknown] | [what it revealed]

EXAMPLES:
  ROT: Sent Lloyd prep framing C4 | worked — he confirmed April 28 | price anchoring lands
  ROT: Sent Gyamfi message | no reply yet | unknown — too early
  ROT: Added hash chain to demo | worked — passes DS-6 | keep in all demos

CALIBRATION LOOP:
  After 10 ROT entries: Claude audits its own recommendation accuracy.
  "Of the last 10 recommendations I made that were acted on:
   N worked, N didn't, N unknown. Accuracy: N/10."
  If accuracy < 60%: flag the expert voice that produced the failing recommendations.
  This applies the ARE FP rate concept to Claude's own outputs.

PRESERVES: DECISION_AUDIT intact — that tracks claims. ROT tracks actions.
           Two different loops. Both needed. Neither replaces the other.

────────────────────────────────────────────────────────────────────
E3 — APEX MVW: MINIMUM VIABLE WIN (X3: Military Planning)
What is the single binary outcome that defines success in 30 days?
────────────────────────────────────────────────────────────────────

ACTIVATION: Checked at every session open. Updated when milestone changes.

FORMAT (add to CONTINUATION_PROMPT CURRENT STATE):
  MVW: [one outcome — specific, binary, 30-day horizon]
  MVW FALLBACK: [if primary fails — what is the pivot trigger and new MVW?]

CURRENT MVW (April 16, 2026):
  MVW: Lloyd signs LoU at April 28 meeting OR agrees to specific next step
       with named date.
  MVW FALLBACK: If Lloyd outcome is neutral/negative by May 5 →
                MVW shifts to: one other warm contact (Gyamfi/Watkin-Child/Vann)
                agrees to pilot scope call by May 15.

WHY THIS EXISTS:
  If Lloyd cancels April 28, there is currently no clear pivot trigger.
  The system has pipeline contacts but no defined "abort Lloyd, activate X"
  decision rule. The MVW makes the pivot automatic and removes the decision
  load from a high-pressure moment.

PRESERVES: All pipeline tracking unchanged. One line added to CURRENT STATE.

────────────────────────────────────────────────────────────────────
E4 — APEX MOMENTUM SCORE (X4: Behavioral Economics)
Visible progress score that makes the right behaviors rewarding.
────────────────────────────────────────────────────────────────────

ACTIVATION: Calculated at session close. Displayed at session open.

SCORING (0–10 per session):
  +2  Commercial action confirmed executed (irreversible, not drafted)
  +2  Tests green at session close
  +1  DECISION_AUDIT updated (claim logged or verdict recorded)
  +1  Ω SCAN clean (all four = 0)
  +1  BUILD_INTELLIGENCE updated (engineering sessions only)
  +1  Pipeline health improved from last session
  +1  Irreversible action was commercial, not engineering
  +1  ROT entry recorded

DISPLAY (add to CONTINUATION_PROMPT CURRENT STATE):
  MOMENTUM: Last session = [N]/10 | 7-day avg = [N]/10 | Trend = [↑/↓/→]

INTERPRETATION:
  8–10: Compounding. Proceed at full speed.
  5–7:  Adequate. Identify lowest-scoring dimension. Fix it this session.
  3–4:  Warning. First task is the lowest-scoring dimension. Nothing else.
  0–2:  Stop. This session is recovery only. Name what broke.

WHY THIS EXISTS:
  Gates prevent wrong moves. Momentum Score makes right moves visible
  and rewarding. The behavioral economics principle: constraint alone
  produces compliance. Constraint + visible progress produces momentum.

PRESERVES: All gates unchanged. Score is information, not a gate.

────────────────────────────────────────────────────────────────────
E5 — APEX PROPAGATE (X7: Complexity Science)
When one file changes, what else must change?
────────────────────────────────────────────────────────────────────

ACTIVATION: Fires after any OS file update, before git commit.

PROPAGATION CHECK (30 seconds, fires at Phase 3 of APEX DECIDE):
  "Does this update change anything in any other active file?"

PROPAGATION MAP (which files affect which):
  ZROS update      → check CONTINUATION_PROMPT (version ref), APEX_REASONING_ENGINE
  CONTINUATION     → check nothing (it is the receiver, not the source)
  BUILD_INTEL      → check CLAUDE_MASTER (product claims must match code reality)
  APEX v5.2        → check all files (master OS change propagates everywhere)
  DECISION_AUDIT   → check CONTINUATION_PROMPT (APEX VERSION calibration status)
  APEX_DECISION_G  → check ZROS (new gates may need law entries)
  INTEGRITY_SENT   → check SESSION_PROTOCOL (new session steps)

RULE: If propagation is needed and not done in this session →
      log as deferred in CONTINUATION_PROMPT with explicit file pair:
      "PROPAGATE DEFERRED: [source file] → [target file] | reason: [one line]"

WHY THIS EXISTS:
  ZROS v2.7 persisted in CONTINUATION_PROMPT for 4 days after ZROS was
  updated to v2.8. File drift is the most common silent failure in this OS.
  Propagation check adds 30 seconds and eliminates the entire failure class.

PRESERVES: All file content unchanged. One check added to commit protocol.

────────────────────────────────────────────────────────────────────
E6 — APEX SIGNAL: POST-MEETING DECODER (X6: Sales Engineering)
What did a buyer actually reveal? Decoded within 24 hours.
────────────────────────────────────────────────────────────────────

ACTIVATION: Within 24 hours of any commercial conversation. Mandatory.

FOUR QUESTIONS (add as structured entry to COMMERCIAL_INTELLIGENCE.md):
  1. VERBATIM: What exact words did they use? (quote directly)
  2. SILENCE: What did they NOT say that you expected them to say?
  3. REAL CONCERN: What is the single most important thing they revealed
                   about their underlying concern — not their stated concern?
  4. MISSED QUESTION: What is the one question you wish you had asked?

LLOYD MEETING PRE-FILL (April 28 — fill within 2 hours of meeting end):
  1. VERBATIM: [fill after meeting]
  2. SILENCE: Did he not ask about FP rate? Did he not ask about pricing?
              Did he not ask about competitors? Silence is signal.
  3. REAL CONCERN: [fill after meeting — this is the most important field]
  4. MISSED QUESTION: [fill after meeting]

WHY THIS EXISTS:
  Every commercial conversation contains 10x more intelligence than
  what gets recorded in a pipeline status update. This decoder extracts
  it in a structured format that feeds the next conversation.
  Without it, every buyer conversation starts from zero.

PRESERVES: COMMERCIAL_INTELLIGENCE.md format enhanced, not replaced.
           Existing entries unchanged. New entries use enhanced format.

────────────────────────────────────────────────────────────────────
E7 — PRODUCT SIGNAL EXTRACTION (X8: Product Management)
What did a buyer reveal about what the product is missing?
────────────────────────────────────────────────────────────────────

ACTIVATION: Appended to every APEX SIGNAL entry (E6 above).

ONE ADDITIONAL FIELD in every COMMERCIAL_INTELLIGENCE entry:
  PRODUCT SIGNAL: [Does anything in this conversation imply a missing
                   feature, a wrong assumption about buyer workflow,
                   or a positioning gap? YES/NO. If YES: one sentence.]

EXAMPLES:
  PRODUCT SIGNAL: YES — Lloyd asked "how do I see trends over time?"
                  implies dashboard needs 30-day trend view. Phase 2 backlog.
  PRODUCT SIGNAL: NO — conversation confirmed current positioning is correct.
  PRODUCT SIGNAL: YES — Gyamfi asked "does it work with Azure AD?"
                  implies SAML/OIDC integration needed before enterprise close.

WHY THIS EXISTS:
  Phase 2 product direction should come from buyer signals, not from
  engineering imagination. Every commercial conversation is a free
  product requirements session. Currently that signal is lost.

PRESERVES: Nothing changed. One field per entry. Takes 30 seconds.

────────────────────────────────────────────────────────────────────
E8 — APEX PEAK: PRE-MEETING PROTOCOL (X9: Performance)
The 16 hours before Lloyd are not ordinary hours.
────────────────────────────────────────────────────────────────────

ACTIVATION: April 27, after BULLETPROOF RUN completes. Non-negotiable.

FIVE NON-NEGOTIABLES (in order):
  1. C1–C15 said aloud, timed, recorded — not read silently.
     Target: each claim delivered in under 20 seconds, no hesitation.
     If any claim requires >20 seconds: that claim needs one more pass.

  2. Three hardest questions answered out loud:
     "What's your FP rate in production?" → answer: Visa standard framing
     "How is this different from Wiz?" → answer: behavioral history vs snapshot
     "What happens if your system goes down?" → answer: fail-open, agents continue

  3. One full sleep cycle. No engineering after 10pm April 27.
     The Lloyd meeting is won or lost in the first 5 minutes.
     Cognitive sharpness matters more than one more commit.

  4. Demo run once April 28 morning — verify it still works.
     Not a full demo. The single most important step (score drop + reason object).
     Confirm: docker compose up, JWT, score endpoint, policy_fired field visible.

  5. One sentence written on paper before leaving:
     "The single thing I need Lloyd to agree to today is: ___"
     This is the session commitment for the Lloyd meeting itself.

WHY THIS EXISTS:
  The system has BULLETPROOF RUN for technical preparation.
  It has no protocol for human preparation.
  The meeting outcome depends on both.

PRESERVES: BULLETPROOF RUN unchanged. APEX PEAK fires after it, not instead.

────────────────────────────────────────────────────────────────────
E9 — APEX SCAN: MONTHLY MARKET INTELLIGENCE (X10: Intelligence Analysis)
What changed in the market this month?
────────────────────────────────────────────────────────────────────

ACTIVATION: Monthly, fires with APEX-OS CALIBRATE on the 1st.
            Takes 30 minutes. Not optional. Logged in COMMERCIAL_INTELLIGENCE.md.

FIVE QUESTIONS (answered with web search, one paragraph each):
  1. What new AI agent security products or features launched this month?
  2. What regulatory guidance on AI agents was published?
  3. What did Check Point / Lakera / Wiz / CrowdStrike announce?
  4. What did enterprise AI buyers publicly say about their concerns?
     (LinkedIn posts, conference talks, earnings calls, press)
  5. What acquisition happened in adjacent security spaces?

OUTPUT FORMAT (append to COMMERCIAL_INTELLIGENCE.md):
  MARKET SCAN: [month]
  Q1 — NEW ENTRANTS: [findings]
  Q2 — REGULATORY: [findings]
  Q3 — COMPETITORS: [findings]
  Q4 — BUYER SIGNAL: [findings]
  Q5 — M&A: [findings]
  IMPLICATION FOR ARE: [one sentence — does anything here change positioning,
                        urgency, or product direction?]

WHY THIS EXISTS:
  Wiz AI-SPM was discovered late. It had been building for months before
  ARE noticed. Proactive monthly scanning converts Unknown Threats to
  Known Threats before they become competitive events.

PRESERVES: Everything. One search session per month. 30 minutes.

────────────────────────────────────────────────────────────────────
E10 — APEX OS-HEALTH GATE (X11: Meta-System Architecture)
The self-improvement engine must fire. Non-optional.
────────────────────────────────────────────────────────────────────

ACTIVATION: G-COMMERCIAL checklist — fires on 1st of every month.

ADD TO G-COMMERCIAL (fires on 1st of month only):
  [ ] APEX-OS CALIBRATE run this month?
      If NO: this session IS the calibration. No other work opens first.
      APEX-OS CALIBRATE takes 45 minutes. It is not optional.
      An OS that doesn't improve is a snapshot, not a system.

APEX-OS CALIBRATE SEQUENCE (45 minutes, runs monthly):
  Step 1 — Ω PATTERN AUDIT (10 min)
    Pull Ω SCAN entries from last 4 weeks.
    For each Ω1–Ω4: count signal frequency.
    3+ signals in 4 weeks → ZROS law candidate. Write it. Submit for approval.

  Step 2 — ROT ACCURACY AUDIT (10 min) [NEW — from E2]
    Pull all ROT entries from last 4 weeks.
    Calculate: recommendations acted on / recommendations that worked.
    If accuracy < 60%: identify which expert voice produced failing outputs.
    Flag that voice for recalibration.

  Step 3 — KNOWLEDGE CONTRADICTION CHECK (10 min)
    Pull LEARNING_INTELLIGENCE entries with CONTRADICTS tags.
    Resolve: which learning is current? Label stale one [F-STALE].
    Cluster check: any cluster with 5+ COMPOUND learnings → meta-pattern.

  Step 4 — MOMENTUM SCORE TREND (5 min) [NEW — from E4]
    Pull last 4 weeks of momentum scores.
    Calculate: average, trend direction, lowest-scoring dimension.
    If avg < 5: structural fix required. Name it.

  Step 5 — MARKET SCAN (30 min) [NEW — from E9, runs same day]
    Run APEX SCAN five questions.
    Log in COMMERCIAL_INTELLIGENCE.md.

  Step 6 — PROPOSAL GENERATION (10 min) [NEW — self-improvement engine]
    Based on all above: what does the OS need to improve?
    Generate APEX v6 PROPOSALS (see Part 2).

OUTPUT: Updated CONTINUATION_PROMPT + COMMERCIAL_INTELLIGENCE + proposals file.

PRESERVES: APEX-OS CALIBRATE steps 1, 3, 4, 5 (original) unchanged.
           Steps 2, 4 (momentum), 5 (market scan), 6 (proposals) are additions.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 2 — THE SELF-IMPROVEMENT ENGINE (APEX v6 CORE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

This is the engine that makes APEX v6 different from every previous version.
Every session, Claude observes the work and generates improvement proposals.
Every month, those proposals are reviewed and approved or rejected.
Over time, the OS improves from real operation — not from design.

────────────────────────────────────────────────────────────────────
THE PROPOSAL ENGINE
────────────────────────────────────────────────────────────────────

ACTIVATION: At every session close, after C2 (irreversible action confirmed).

Claude runs a 5-minute internal audit and generates proposals.
Proposals are presented as structured items for human approval.
Nothing executes without explicit approval. Ever.

PROPOSAL FORMAT:
  P[N] — [one line description]
  TYPE: [LAW / GATE / ENHANCEMENT / DEPRECATION / PROPAGATION]
  TRIGGER: [what happened this session that suggests this proposal]
  CHANGE: [exactly what would be added/modified/removed]
  FILE: [which file receives this change]
  EFFORT: [minutes to implement]
  VERDICT REQUEST: APPROVE / DEFER / REJECT

PROPOSAL TYPES:
  LAW         → new ZROS law from a real incident this session
  GATE        → new gate from a pattern that should be stopped
  ENHANCEMENT → improvement to existing mechanism (from 11 council)
  DEPRECATION → gate or law that hasn't fired in 30 days (review for removal)
  PROPAGATION → cross-file update needed from a change made this session

────────────────────────────────────────────────────────────────────
PROPOSAL GENERATION TRIGGER (fires at session close, 5 minutes)
────────────────────────────────────────────────────────────────────

Claude asks itself 8 questions:

  Q1: Did any friction occur this session that a new law would prevent?
      If yes → generate LAW proposal.

  Q2: Did any avoidance pattern appear that wasn't caught by existing gates?
      If yes → generate GATE proposal.

  Q3: Did any existing gate fire unnecessarily or cause confusion?
      If yes → generate DEPRECATION or ENHANCEMENT proposal.

  Q4: Did any file drift from another file this session?
      If yes → generate PROPAGATION proposal.

  Q5: Did any ROT entry reveal that a recommendation was wrong?
      If yes → generate ENHANCEMENT proposal for the relevant expert voice.

  Q6: Did any commercial conversation produce a product signal?
      If yes → log in COMMERCIAL_INTELLIGENCE. No proposal needed.

  Q7: Did any [H] claim fail adversarial gate this session?
      If yes → log in DECISION_AUDIT. Generate ENHANCEMENT if pattern.

  Q8: Is any active file now stale relative to reality?
      If yes → generate PROPAGATION proposal.

OUTPUT: 0–3 proposals per session. More than 3 means the session was
        too unfocused. Name the root cause instead of generating proposals.

────────────────────────────────────────────────────────────────────
PROPOSAL FILE: APEX_v6_PROPOSALS.md
────────────────────────────────────────────────────────────────────

A running log of all proposals. Never deleted. Verdicts recorded.

FORMAT:
  ═══════════════════════════════════════════════════
  SESSION: [date] | MOMENTUM: [N]/10
  ═══════════════════════════════════════════════════

  P001 — [description]
  TYPE: LAW
  TRIGGER: [what happened]
  CHANGE: [exact text to add]
  FILE: ZROS v2.8 → L13
  EFFORT: 10 min
  VERDICT: [APPROVED / DEFERRED / REJECTED] by Rehan on [date]

  P002 — [description]
  ...

  SESSION PROPOSALS: [N] | Approved: [N] | Deferred: [N] | Rejected: [N]

MONTHLY REVIEW (fires with APEX-OS CALIBRATE):
  Pull all DEFERRED proposals from last 30 days.
  For each: is the trigger still valid? Re-present or close.
  Pattern: if same proposal deferred 3x → either APPROVE or REJECT permanently.
  A proposal deferred 3 times is Ω2 applied to the OS itself.

────────────────────────────────────────────────────────────────────
THE COMPOUNDING MECHANISM
────────────────────────────────────────────────────────────────────

Session N produces:
  → 0–3 proposals
  → 1 ROT entry
  → 1 momentum score
  → 1 APEX SIGNAL entry (if commercial conversation happened)

Month 1 produces:
  → ~15–30 proposals (10 sessions × avg 2 proposals)
  → ~10 ROT entries
  → ROT accuracy score
  → 1 APEX SCAN entry
  → APEX-OS CALIBRATE output

Month 6 produces:
  → ~90–180 proposals reviewed
  → ~60 ROT entries → recommendation accuracy score: X%
  → OS laws: L1–L25+ (all earned from real failure)
  → Buyer intelligence: N entries, patterns identified
  → Momentum average: N/10 (visible trend over 6 months)

The OS after 6 months is unrecognizable from the OS at day 1.
Not because it was redesigned. Because it improved from real operation.
That is the moat. Competitors can copy the design. They cannot copy
6 months of real earned failure converted into laws.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 3 — UPDATED SESSION SEQUENCE (full A1→C6 with v6 additions)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

SESSION OPEN (additions marked ★ NEW):

  ★ APEX TRIAGE     — 10 seconds. "What is most important today?" → routes session type.
  A1 LOAD STATE     — 90-second boot. All 6 lines stated. (unchanged)
  ★ ROT CHECK       — "What did you act on last session? Did it work?" (30 seconds)
  ★ MVW CHECK       — "Is the MVW still current?" (10 seconds)
  ★ MOMENTUM DISPLAY — "Last session score: [N]/10" (5 seconds)
  A2 G-COMMERCIAL   — TIER 0/1/2 checklist. (unchanged)
  A3 SENTINEL RUN   — Engineering sessions only. (unchanged)
  A4 DRIFT CHECK    — Mondays only. (unchanged)
  A5 PHASE CONFIRM  — State phase, mode, session type. (unchanged)
  A6 ADVERSARIAL    — Before any [H] claim. (unchanged)

SESSION CLOSE (additions marked ★ NEW):

  C1 TESTS GREEN    — go build && go test. (unchanged)
  C2 IRREVERSIBLE   — One real-world action confirmed. (unchanged)
  C3 DECISIONS      — Strategic/[H] claims logged. (unchanged)
  ★ PROPAGATION CHECK — "Does this session's changes require other file updates?" (30 sec)
  ★ PROPOSAL ENGINE — 5 minutes. 0–3 proposals generated for approval. (NEW)
  ★ MOMENTUM SCORE  — Calculate this session's score. Record in CONTINUATION_PROMPT. (2 min)
  C4 STATE UPDATE   — scripts/update_continuation.sh (unchanged)
  C5 COMMIT + PUSH  — git add -A && git commit && git push (unchanged)
  C6 UPLOAD         — Updated CONTINUATION_PROMPT to Claude Project (unchanged)
  ★ PROPOSALS FILE  — If proposals generated: update APEX_v6_PROPOSALS.md and upload. (2 min)

TOTAL ADDED TIME PER SESSION: ~8 minutes (open: 50 seconds | close: 7 minutes)
TOTAL VALUE PER SESSION: ROT entry + momentum score + 0–3 improvement proposals
                         + propagation verification + commercial signal capture

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 4 — CONTINUATION_PROMPT ADDITIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Add these fields to CONTINUATION_PROMPT CURRENT STATE block:

  ─────────────────────────────────────────────────
  APEX v6 STATE
  ─────────────────────────────────────────────────
  MVW: Lloyd signs LoU or agrees specific next step April 28
  MVW FALLBACK: If Lloyd neutral by May 5 → Gyamfi/Watkin-Child/Vann pilot call by May 15
  MOMENTUM: Last session = [N]/10 | 7-day avg = [N]/10 | Trend = [↑/↓/→]
  ROT ACCURACY: [N/N recommendations verified] | Accuracy = [N]%
  PROPOSALS PENDING: [N] awaiting verdict | Last proposal: [date]
  NEXT APEX-OS CALIBRATE: May 1, 2026 (mandatory — blocks all other work)
  APEX PEAK: Scheduled April 27 (after BULLETPROOF RUN)
  ─────────────────────────────────────────────────

Add to G-COMMERCIAL (fires 1st of every month):
  [ ] APEX-OS CALIBRATE run this month? NO → this session IS the calibration.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 5 — ACTIVATION COMMANDS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

APEX v6 BOOT          → Run full v6 open sequence (TRIAGE + ROT + MVW + MOMENTUM)
APEX TRIAGE           → 10-second session type router
APEX ROT              → Log recommendation outcome. "Did it work?"
APEX MVW              → Check or update Minimum Viable Win
APEX MOMENTUM         → Calculate and display session momentum score
APEX SIGNAL           → Post-meeting decoder. Fires within 24hr of buyer conversation.
APEX PROPAGATE        → Check cross-file propagation after any OS file change
APEX PROPOSE          → Generate session-close improvement proposals
APEX PEAK             → Pre-Lloyd performance protocol (April 27)
APEX SCAN             → Monthly market intelligence (fires with CALIBRATE)
APEX v6 CALIBRATE     → Full monthly meta-pass including all v6 additions
APEX v6 STATUS        → Full system health: momentum + ROT accuracy + proposals + MVW

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 6 — WHAT DOES NOT CHANGE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

These are explicitly NOT touched by APEX v6:

  ★ All ZROS laws L1–L12 — unchanged. New laws proposed, not imposed.
  ★ All APEX v5.2 gates — G-FP, G-EXPLAIN, G-IDENTITY, G-DEPLOY unchanged.
  ★ APEX_DECISION_GOVERNANCE Five-Why Tribunal — unchanged.
  ★ INTEGRITY_SENTINEL 15 invariants — unchanged.
  ★ ARE_BULLETPROOF stress test — unchanged. APEX PEAK fires after it.
  ★ SESSION_PROTOCOL A1→C6 — unchanged. v6 additions slot into existing sequence.
  ★ DECISION_AUDIT schema — unchanged. ROT is a separate tracker.
  ★ CONTINUATION_PROMPT structure — additions only, nothing removed.
  ★ Prime Directive — unchanged. Ship to one enterprise. Blocked incident. 6 months.

The OS that earned its laws does not discard them.
Every law in ZROS represents a real failure that cost real time.
APEX v6 adds new mechanisms on top. It does not replace the foundation.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 7 — DEPLOYMENT SEQUENCE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Step 1 — This session:
  [ ] Download APEX_v6_SELF_IMPROVING_OS.md
  [ ] Copy to docs/ops/ in repo
  [ ] Add APEX v6 STATE block to CONTINUATION_PROMPT (Part 4 above)
  [ ] Add G-COMMERCIAL monthly calibrate gate (Part 4 above)
  [ ] Create empty APEX_v6_PROPOSALS.md in docs/ops/
  [ ] git commit -m "ops: APEX v6 self-improving OS — 11 council enhancements + proposal engine"
  [ ] Upload both files to Claude Project

Step 2 — April 27 (day before Lloyd):
  [ ] Run BULLETPROOF RUN
  [ ] Then run APEX PEAK (5 items above)

Step 3 — April 28 (after Lloyd meeting):
  [ ] Run APEX SIGNAL (post-meeting decoder — within 2 hours)
  [ ] Record ROT entry
  [ ] Generate first APEX PROPOSE session proposals

Step 4 — May 1 (first monthly calibration):
  [ ] Run APEX v6 CALIBRATE (full 45-minute meta-pass)
  [ ] Run APEX SCAN (market intelligence)
  [ ] Review all pending proposals in APEX_v6_PROPOSALS.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VERSION HISTORY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

v6.0 — April 16, 2026
  Built from: Council of 11 experts (X1–X11) + APEX-OS v1.0 meta-layer
  Key additions: Self-improvement engine, proposal system, ROT tracker,
                 MVW, momentum score, post-meeting decoder, market scan,
                 APEX PEAK, propagation check, OS-health gate.
  Next update trigger: First APEX v6 CALIBRATE (May 1, 2026).
                       Version becomes v6.1 after first real proposal cycle.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
APEX v6.0 SELF-IMPROVING OS | April 16, 2026
11 council enhancements | Self-improvement engine | Proposal system
Preserves: APEX v5.2 + ZROS v2.8 + all 12 active files — zero damage
Next milestone: APEX PEAK (April 27) → APEX SIGNAL (April 28) → CALIBRATE (May 1)
File: docs/ops/APEX_v6_SELF_IMPROVING_OS.md
