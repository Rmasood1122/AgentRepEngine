# UNIVERSAL INTELLIGENCE FILE v1.0
# Distilled from AgentRepEngine build — March 2026
# Applicable to ANY project, ANY domain
# 12 Expert Panel | Meta System Prompt Style
# 
# HOW TO USE:
# Paste this file into any Claude Project as context.
# It will make every session start with maximum intelligence.
# Update it after every major project learning.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 1 — HOW TO GET MAXIMUM VALUE FROM CLAUDE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## THE SINGLE MOST IMPORTANT INSIGHT

Claude is not a chatbot. Claude is a system that compounds.
Every session you invest in context = every future session
returns higher value. The compounding rate is exponential.

A session without context: 20% efficiency.
A session with CONTINUATION_PROMPT: 80% efficiency.
A session with CONTINUATION_PROMPT + Intelligence files: 95% efficiency.

The investment is 5 minutes per session.
The return is hours saved per week.

---

## U1 — CONTINUATION_PROMPT DISCIPLINE
[Source: AgentRepEngine — 6 weeks of sessions]

The single highest-ROI habit in any Claude project.

RULE: At the end of EVERY session, update a file called
CONTINUATION_PROMPT.md with:
- What was done
- What is the current state
- What is next
- Any blockers or decisions pending

Then upload it to Claude Project.

WHY IT MATTERS:
- Without it: every session starts cold — 15-20 min rebuilding context
- With it: every session starts in 30 seconds
- Over 20 sessions: saves 5-7 hours of rebuilding

TEMPLATE:
```
## CURRENT STATE
Phase/Stage: ___
Last action: ___
Next action: ___
Key metrics: ___
Blockers: ___
First command next session: ___
```

NEVER skip this. It is the difference between compounding and
starting over every time.

---

## U2 — ACTIVATION COMMANDS SYSTEM
[Source: APEX v5.2 methodology]

Create named activation commands for your project.
Instead of explaining context every time, one command loads everything.

PATTERN:
```
PROJECT ACTIVATE — [current task]
→ Claude reads all project files and picks up exactly where you left off
```

Examples that work:
- "APEX ACTIVATE — Task 3"
- "PROJECT ACTIVATE — Sprint 2"
- "SYSTEM ACTIVATE — design phase"

WHY: Reduces session startup from 10 minutes to 30 seconds.
Every project should have 5-10 named commands that route to modes.

---

## U3 — EXPERT PANEL TECHNIQUE
[Source: APEX v5.2 expert voices]

Never ask Claude for one opinion. Ask for 5-12 expert voices.

PATTERN:
"Analyze this as top 5 experts in [domain]. Each expert should
give their honest verdict. No flattery. Hard truth only."

WHY IT WORKS:
- Single answer = one perspective, often generic
- Expert panel = multiple perspectives, often reveals blind spots
- The disagreement between experts is where the real insight lives

BEST USE CASES:
- Evaluating a decision before committing
- Auditing work before presenting to stakeholders
- Identifying gaps before they become problems
- Choosing between competing approaches

UPGRADE: Add "meta system prompt" — tell Claude to embody each
expert's actual thinking style, not just their domain knowledge.

---

## U4 — CONFIDENCE LABELING SYSTEM
[Source: APEX v5.2 confidence labels]

Always ask Claude to label confidence on every statement.

LABELS:
- [F]   — Fact, directly verifiable
- [H]   — High confidence inference
- [BP]  — Best practice, widely accepted
- [ASS] — Assumption, not verified
- [INF] — Inference from available data
- [OBS] — Observed in context

WHY: Without labels, Claude sounds equally confident about
facts and guesses. With labels, you know what to verify.

ADD TO ANY PROMPT: "Label every claim with confidence:
[F] fact, [H] high confidence, [ASS] assumption."

---

## U5 — ANTI-SCOPE DISCIPLINE
[Source: APEX v5.2 G-KILL gate]

Before any work session, run a scope check.

THE QUESTION: "Does this task belong to this phase/stage?"

If yes → proceed.
If no → park it. Build a parking lot.

WHY: The single biggest time waster in any project is building
the right thing at the wrong time. Phase 3 work in Phase 1
kills momentum and creates technical debt.

IMPLEMENTATION:
- Maintain a PARKING LOT section in your project docs
- Any good idea that belongs to a later phase goes there
- Review parking lot at phase transitions

THE GATE QUESTION: "If I do this now and it turns out wrong,
how much do I lose?" If the answer is more than 1 day — defer.

---

## U6 — PRE-MORTEM BEFORE EVERY BUILD
[Source: ZROS v2.6/2.7 Law L3]

Before writing any code or creating any artifact:
Name 3 ways it can fail.

TEMPLATE:
```
Before building [X]:
1. Most likely failure: ___
2. Second most likely failure: ___  
3. Silent failure (hard to detect): ___
Detection signal for each: ___
Recovery path for each: ___
```

WHY: If you cannot name 3 failure modes, you do not understand
the component well enough to build it. This is not pessimism —
it is precision. Pre-mortem takes 5 minutes. Debugging takes hours.

---

## U7 — STATE → EVIDENCE → HYPOTHESIS → TEST
[Source: ZROS v2.6 Law L5 — Never debug blind]

When something breaks, always follow this sequence:

1. STATE: What is the current state? (exact error, exact output)
2. EVIDENCE: What changed? (last working state, what was different)
3. HYPOTHESIS: What is the most likely cause? (one specific guess)
4. TEST: One change, measure result.

NEVER: Try multiple fixes at once. Never patch forward from broken.
ALWAYS: Establish last known-good state before any fix.

THE 3-PATCH RULE: If 3 patches on the same file do not fix it —
STOP. Switch to structured diagnosis. You are guessing.

---

## U8 — DEFINITION OF DONE (6 LEVELS)
[Source: ZROS v2.6 C9]

"Done" means L6. Not L1.

L1 — Code written
L2 — Build passes, zero errors
L3 — Unit tests pass (>90% coverage)
L4 — FP/edge case tests pass
L5 — Output validated (schema correct, format correct)
L6 — Performance verified (latency, memory, throughput measured)

Apply this to ANY deliverable — not just code:
L1 — Draft written
L2 — Logic checks out
L3 — Edge cases covered
L4 — Reviewed by second perspective
L5 — Output matches intended use
L6 — Tested in real context

---

## U9 — REWORK RATE TRACKING
[Source: ZROS v2.6/2.7 primary metric]

Track the ratio of fix commits to feature commits.

GREEN  < 10% fix rate — continue building
YELLOW 10-20% fix rate — audit root cause before next session
RED    > 20% fix rate — STOP. Fix the system before building.

WHY: Fix rate is a leading indicator of system health. High fix
rate means you are building on a shaky foundation. Every fix
commit means something was not understood before building.

APPLY TO ANY PROJECT:
- Track "rework" vs "new work" ratio weekly
- If rework > 20% — find the root cause, not just the symptom

---

## U10 — NEVER STATE ASSUMPTIONS AS FACTS
[Source: APEX v5.2 core principle]

Claude will sometimes present assumptions as facts if not
instructed otherwise. Always prompt:

"Distinguish between what you know for certain and what you
are inferring. Label assumptions explicitly."

AND: When Claude says something that matters — verify it.
Especially: numbers, timelines, technical specifications,
market data. Claude's training data has a cutoff. Markets move.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 2 — HIGHEST VALUE MOVES (E2 McKinsey + E10 VC)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U11 — THE HIGHEST VALUE MOVE FRAMEWORK

Before ANY action, ask: "What is the highest value move right now?"

RANKING CRITERIA:
1. Irreversibility — can this be undone if wrong?
2. Compounding — does this make future work easier?
3. Blocking — is something else waiting on this?
4. Time sensitivity — does this close a window?
5. Leverage — does this unlock multiple outcomes?

THE RULE: Always do the highest leverage action first.
Not the easiest. Not the most comfortable. The highest leverage.

EXAMPLES:
- IP filing before outreach (irreversible window)
- Evaluation harness before sales (blocking)
- CONTINUATION_PROMPT before session end (compounding)
- Design partner before feature build (validation)

---

## U12 — COMPETITIVE CLOCK DISCIPLINE
[Source: APEX v5.2 — 12-month clock]

Every project has a competitive window.
Name it explicitly at project start.

TEMPLATE:
```
Competitive window opens: [date]
Competitive window closes: [date/event]
Current position: [months remaining]
What closes the window: [specific event]
Our advantage before window closes: [specific action]
```

WHY: Without an explicit clock, urgency is abstract.
With a clock, every session has a clear cost of inaction.

CHECK EVERY SESSION: "Does what I am doing today
advance the goal before the window closes?"

---

## U13 — VALUE BEFORE VOLUME
[Source: AgentRepEngine GTM]

In any outreach, sales, or partnership effort:

WRONG: Send 100 messages → hope for responses
RIGHT: Identify 10 highest-value targets → send 10 perfect messages

The difference: 10 perfect messages to right people > 100 generic
messages to anyone.

HOW TO IDENTIFY HIGHEST VALUE:
1. Who has the most direct pain?
2. Who has budget authority?
3. Who has the shortest path from conversation to commitment?
4. Who becomes a reference that unlocks 10 more?

Do the analysis first. Then execute.

---

## U14 — PHASE LOCKING
[Source: APEX v5.2 phase gates]

Every project has phases. Lock each phase before advancing.

PHASE LOCK CRITERIA:
- Define 3 concrete exit criteria for each phase
- All 3 must be true before advancing
- No exceptions — phases exist to prevent premature scaling

COMMON MISTAKE: Doing Phase 2 work while Phase 1 is incomplete.
This is the most common startup failure mode. Phase 1 incomplete
+ Phase 2 started = neither phase works.

---

## U15 — ONE ENTERPRISE, ONE INCIDENT
[Source: APEX v5.2 Prime Directive]

In any B2B product: do not try to close 10 customers.
Close 1 customer. Get 1 documented success. Then scale.

WHY: 1 documented success is worth more than 10 prospects.
It answers every objection. It provides a reference.
It validates the entire business model.

THE TRAP: Building features for imaginary customers instead of
deploying to a real one. Features do not validate. Deployments do.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 3 — IP PROTECTION PROTOCOL (E4 IP Attorney)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U16 — ZENODO FIRST PRINCIPLE

For any original idea, framework, or methodology:
Publish on Zenodo BEFORE building, BEFORE pitching, BEFORE sharing.

WHY: Zenodo creates a timestamped, DOI-indexed public record.
This is your prior art. Cost: zero. Time: 30 minutes.

SEQUENCE:
1. Write 1-page abstract of your idea
2. Upload to zenodo.org as Preprint or Software
3. Select "All Rights Reserved" or appropriate license
4. Get DOI — this is your IP anchor

THEN: In every subsequent document, reference the DOI.
"This work implements the framework published at DOI: 10.5281/..."

---

## U17 — IP CHAIN BUILDING

Build an IP chain — each implementation links to its theoretical basis.

PATTERN:
Theory paper (DOI) → Protocol paper (DOI) → Implementation (DOI + GitHub)

WHY: A chain is harder to challenge than a single record.
Each link reinforces the others. A competitor cannot claim
priority on your implementation if your theory predates them.

---

## U18 — TRADEMARK ™ IN ZENODO

In Zenodo titles and descriptions, use ™ for any coined term.
Example: "Agentic Trust Protocol™"

This does not constitute legal trademark registration but
establishes public record of use — relevant in priority disputes.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 4 — PRODUCT BUILD DISCIPLINE (E1 FAANG + E8 Founder)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U19 — FAIL OPEN VS FAIL CLOSED

Know which failures should fail open and which should fail closed.

FAIL OPEN: Infrastructure failures — keep the system running
even without full functionality. Never block users because
a sidecar is down.

FAIL CLOSED: Security/enforcement failures — never allow
an unverified action through. A missing reason object means
AUDIT mode, never silent pass.

APPLY UNIVERSALLY:
- Payment system down? → Fail open (show product, collect payment later)
- Auth system down? → Fail closed (no access without verification)

---

## U20 — BENCHMARK ON PRODUCTION ENVIRONMENT

Never present performance numbers from development environment.
Always benchmark on Linux/production equivalent.

Windows dev → Linux production: typically 20-40% faster.
This matters in enterprise conversations.

RULE: Any number you present to a customer must come from
an environment that matches their production.

---

## U21 — EVERY COMMIT IS A FEATURE

[Source: ZROS Law L1]

Target: >80% of commits are "feat:" not "fix:"
If fix rate exceeds 20% — stop building, audit root cause.

FIX COMMITS MEAN: Something was not understood before building.
The root cause is almost always: built before fully understanding
the requirements, the interface, or the failure modes.

---

## U22 — READ BEFORE TOUCH

Before modifying any file, read it completely.
Before modifying any system, understand its current state.

THE TRAP: Claude (and humans) will often start writing
before fully reading. This creates conflicts, regressions,
and unnecessary rewrites.

RULE: G-READ before every edit. No exceptions.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 5 — CLAUDE OPTIMIZATION ADVANCED (E6 AI Researcher + E7 Cognitive)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U23 — PROJECT FILES ARE PERMANENT MEMORY

Claude has no memory between conversations by default.
Project files are the exception — they persist.

WHAT TO PUT IN PROJECT FILES:
- CONTINUATION_PROMPT.md (session state)
- Intelligence files (this document)
- Spec files (what you are building)
- Decision log (why you made key choices)

WHAT NOT TO PUT:
- Raw data (too large, wastes context)
- Temporary files (clutters memory)
- Duplicate information (confuses priority)

RULE: Every project file should answer one question that
Claude would otherwise have to ask you.

---

## U24 — CONTEXT WINDOW MANAGEMENT

Claude's context window is finite. Manage it actively.

HIGH VALUE context (always include):
- Current state / what was decided
- Current task / what we are doing now
- Key constraints / what we cannot do
- Success criteria / what done looks like

LOW VALUE context (exclude):
- Historical decisions already made
- Verbose explanations of obvious things
- Duplicate information

RULE: If a piece of context does not change Claude's next
response — remove it from the prompt.

---

## U25 — SURGICAL EDITS NOT REWRITES

Always ask Claude for surgical edits — not full rewrites.

WRONG: "Rewrite this file with these changes"
RIGHT: "Change only line 47 — replace X with Y"

WHY: Full rewrites introduce unintended changes.
Surgical edits are verifiable — you can check exactly what changed.

RULE: If Claude offers to rewrite a file — ask for the
minimum change instead. Verify the diff.

---

## U26 — ONE QUESTION AT A TIME

Claude performs better with one clear question than with
five questions bundled together.

WRONG: "What should I do about X, Y, Z, and also how does W work?"
RIGHT: "What is the single highest priority action for X?"

THEN: Ask about Y separately after X is resolved.

WHY: Bundled questions get averaged answers.
Focused questions get precise answers.

---

## U27 — TELL CLAUDE WHAT EXPERT TO BE

Generic Claude = generic answer.
Expert-instructed Claude = domain-specific precision.

PATTERN:
"You are a [specific expert] with [specific experience].
Analyze this from that perspective only."

BEST EXPERTS TO INVOKE:
- "FAANG principal engineer reviewing production code"
- "Enterprise CISO evaluating a security product"
- "YC partner doing a 10-minute company evaluation"
- "Adversarial red team trying to break this system"
- "Enterprise buyer who has seen 50 similar products"

---

## U28 — ASK FOR THE COUNTER-ARGUMENT

After Claude gives you an answer — ask:
"What is the strongest argument against this?"

WHY: Claude optimizes for helpfulness. It will often tell you
what you want to hear. Explicitly asking for counter-arguments
activates a different mode — more critical, more honest.

PATTERN:
1. Get Claude's recommendation
2. Ask: "What is the strongest reason this is wrong?"
3. Evaluate both sides
4. Make decision

---

## U29 — COMPOUNDING INTELLIGENCE PATTERN

Every project session should produce at least one reusable insight.

AT END OF EVERY SESSION:
1. What was the most important thing learned?
2. Does it apply to other projects? If yes — add to Universal Intelligence file
3. What would have saved the most time? → Add as a new rule

COMPOUNDING EFFECT:
Session 1: 1 insight
Session 10: 10 insights, each session builds on all previous
Session 50: 50 insights, exponentially faster execution

---

## U30 — STRUCTURED DIAGNOSTIC BEFORE ASKING FOR HELP

Before asking Claude to debug anything, provide:
1. Current state (exact error, exact output)
2. Expected state (what should happen)
3. Last working state (when did it last work)
4. What changed (diff or description)
5. What was tried (each attempt and result)

WHY: Claude with full diagnostic context = 1 attempt to fix.
Claude without context = 3-5 attempts. Time multiplier: 3-5x.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 6 — TIME ECONOMY (E12 Time Economist)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U31 — THE 10-MINUTE RULE

Any action that takes less than 10 minutes and has permanent
value should be done immediately — not deferred.

Examples:
- Updating CONTINUATION_PROMPT: 5 minutes → saves hours
- Publishing to Zenodo: 15 minutes → permanent IP protection
- Writing a one-liner for an outreach message: 5 minutes → opens doors

TRAP: Deferring small high-value actions because they feel
small. Small + permanent value = do it now.

---

## U32 — PARALLEL TRACK DISCIPLINE

While waiting for one thing — always advance another.

PATTERN (from AgentRepEngine):
- Waiting for Lloyd to respond → message 5 more LinkedIn targets
- Waiting for pilot feedback → build SDK
- Waiting for legal review → build technical docs

RULE: There should always be a productive action available
that does not depend on the blocked item.
If you are truly blocked on everything → something is wrong
with the project structure.

---

## U33 — BEFORE REPLACING — COMPARE VALUE

Before replacing any existing solution:
1. What does current solution do well?
2. What does new solution do better?
3. What does new solution do worse?
4. What is the migration cost?
5. Is net value positive?

This applies to: tools, processes, frameworks, people, code.

FROM AGENTREPENGINE: Never replace PostgreSQL with Kafka in Phase 1.
Kafka is better at scale. PostgreSQL costs zero to migrate from.
Net value at Phase 1 scale: negative. Decision: keep PostgreSQL.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 7 — SESSION MANAGEMENT SYSTEM
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U34 — SESSION START PROTOCOL

Every session, in this order:
1. Read CONTINUATION_PROMPT — understand current state
2. Check for blockers (responses, decisions, external deps)
3. Run G-KILL equivalent — should we do today's planned task?
4. Identify highest value action for this session
5. Execute

Never: open editor before steps 1-4 are complete.

---

## U35 — SESSION END PROTOCOL

Every session, in this order:
1. Update CONTINUATION_PROMPT with current state
2. Commit all work to version control
3. Upload CONTINUATION_PROMPT to Claude Project
4. Note one key learning from the session
5. Identify first action for next session

Time required: 5-10 minutes.
Value: makes next session 15-20 minutes faster.
Over 20 sessions: 5 hours saved.

---

## U36 — SINGLE SOURCE OF TRUTH

Every project needs one file that is always current:
- Current state
- Current task
- Key decisions made
- Next action

If this file is not current — the project is at risk.
Every team member (or Claude session) working from stale
context creates divergence. Divergence creates rework.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 8 — THE META PRINCIPLES (All 12 Experts)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## U37 — CLARITY BEFORE SPEED

The most common cause of wasted time is starting before
fully understanding the problem.

BEFORE ANY TASK: Can you write one sentence stating:
- What success looks like
- What failure looks like
- What the single most important constraint is

If you cannot write this sentence — you are not ready to start.

---

## U38 — GATES NOT GUIDELINES

Guidelines are suggestions. Gates are requirements.

For any high-stakes decision or deliverable — create a gate:
A binary pass/fail check that must pass before proceeding.

Examples:
- FP rate ≤ 2% before enforcement goes live (gate)
- Not "try to keep FP rate low" (guideline)

Gates eliminate negotiation with yourself.
When you are tired, rushed, or excited — gates protect you.

---

## U39 — DATA MOAT FIRST

In any product that learns from usage:
Instrument data collection from day one.

You cannot recover historical data that was never collected.
You can always decide later not to use data you collected.

RULE: Before first user, data collection must be working.
This is not optional. This is the moat.

---

## U40 — THE PARKING LOT IS NOT A GRAVEYARD

Good ideas that belong to later phases:
- Write them down
- Date them
- Review at each phase transition

Many "later" ideas become critical when the time is right.
The parking lot is where Phase 3 strategy lives while
you are executing Phase 1.

---

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
QUICK REFERENCE — TOP 10 HIGHEST VALUE ACTIONS IN ANY PROJECT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. CONTINUATION_PROMPT updated and uploaded — every session
2. Zenodo DOI published — before any public sharing
3. Expert panel review — before any major decision
4. Pre-mortem — before any major build
5. Gate defined — before any high-stakes deliverable
6. Parking lot maintained — when good ideas arrive at wrong time
7. Rework rate tracked — weekly
8. One design partner — before building for imaginary customers
9. Competitive clock named — at project start
10. Data collection instrumented — before first user

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HOW TO USE THIS FILE IN A NEW PROJECT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Upload this file to your Claude Project
2. At session start: "Read UNIVERSAL_INTELLIGENCE_v1.md and
   apply these principles to everything we do today"
3. When making a decision: "Which principle in
   UNIVERSAL_INTELLIGENCE_v1.md applies here?"
4. At session end: "Did we follow the principles?
   What should be added based on today?"

UPDATE THIS FILE: After every major project — add new learnings.
Version it: v1.0, v1.1, v2.0
The file compounds. Each project makes it smarter.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
UNIVERSAL INTELLIGENCE FILE v1.0
40 principles | 12 expert voices | Distilled March 2026
Update after every major project.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━