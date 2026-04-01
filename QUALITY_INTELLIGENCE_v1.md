━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
QUALITY INTELLIGENCE v1.0 — AGENTREPENGINE
The complete forensic learning log: what went wrong, what caused rework,
what produced FAANG-grade output, and how to maintain it permanently.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Source    : 20 complete chat sessions, March 17–31, 2026
Commits   : 80+ across all phases
Built by  : Rehan Masood + APEX v5.2 + 5-expert panel
Purpose   : Never repeat these mistakes. Apply to every future product.

This document supersedes all prior retrospective notes.
It is a living standard, not a historical record.
Update it after every pilot, every rework incident, every FAANG-grade win.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 1 — THE ROOT CAUSES OF EVERY REWORK INCIDENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

These are the actual causes extracted from 20 sessions.
Not theories. Not best practices. Real incidents with real consequences.

──────────────────────────────────────────────────────────────────────────────
RC-01: BUILDING WITHOUT READING THE EXISTING CODE FIRST
──────────────────────────────────────────────────────────────────────────────

What happened:
  bootstrap.go was written to initialize the scoring harness.
  It overwrote harness_test.go — the G-FP gate test file.
  FP rate jumped to 32%. Session stopped. Git restore required.
  2 hours of recovery time lost.

Root cause:
  Code was written without first reading what files existed
  in the target directory. The collision was invisible until
  the overwrite happened.

FAANG standard:
  Before writing any file: ls -la the target directory.
  Before touching any function: read the function and its callers.
  G-READ gate: mandatory code review before writing.

Law: Read before write. Always. No exceptions.

──────────────────────────────────────────────────────────────────────────────
RC-02: CLAIMING A FEATURE EXISTS BEFORE VERIFYING IT HAS CALLERS
──────────────────────────────────────────────────────────────────────────────

What happened:
  SIEM webhook (siem.go) was written and claimed in enterprise conversations.
  Expert panel ran: grep -rn "SendBlocked" internal/ | grep -v "siem.go"
  Result: empty. SendBlocked() had zero callers.
  The feature did not exist in production. The claim was false.

Root cause:
  Code was written but never wired. The function existed.
  The integration did not. These are not the same thing.

FAANG standard:
  A feature exists when it has callers, not when the function is defined.
  Verify with: grep -rn "FunctionName" | grep -v "filename_where_defined"
  Must return non-empty before claiming the feature.

Law: A function with zero callers does not exist in production.

──────────────────────────────────────────────────────────────────────────────
RC-03: APPLICATION-LEVEL SECURITY THAT THE DB LAYER COULD OVERRIDE
──────────────────────────────────────────────────────────────────────────────

What happened:
  enforcement_decisions table was claimed as INSERT-only (tamper-evident).
  Expert panel ran privilege check.
  Result: DELETE and UPDATE were permitted at PostgreSQL level.
  "Tamper-evident" claim was false. Any SQL injection destroyed the audit.

Root cause:
  INSERT-only was enforced at application layer (app never called DELETE).
  PostgreSQL permissions were never explicitly revoked.
  The application layer can be compromised. The DB layer cannot lie.

FAANG standard:
  Security guarantees must be enforced at the lowest possible layer.
  Application code is not a security boundary.
  DB permissions are. OS permissions are. Network ACLs are.
  Always verify: SELECT privilege_type FROM role_table_grants
                 WHERE table_name=X AND grantee=Y

Law: Never trust the application layer to enforce a security guarantee.
     Enforce it at the database, OS, or network level.

──────────────────────────────────────────────────────────────────────────────
RC-04: ENVIRONMENT VARIABLE WITH EMPTY STRING DEFAULT = OPEN
──────────────────────────────────────────────────────────────────────────────

What happened:
  SCORING_API_KEY was required for /score endpoint access.
  Default value in docker-compose.yml was empty string.
  Any deployment where the env var was not explicitly set
  had a fully open score endpoint. Any agent could probe its own score.

Root cause:
  Empty string was treated as "not configured."
  It is actually "configured to nothing" — which means open.

FAANG standard:
  Every security-sensitive env var must fail closed on empty string.
  Either: set a non-empty default, or refuse to start if empty.
  Never allow an empty string to function as "not set."

Law: Empty string is not "not set." It is "open."

──────────────────────────────────────────────────────────────────────────────
RC-05: SCORING SERVICE NOT REBUILT AFTER CODE CHANGE
──────────────────────────────────────────────────────────────────────────────

What happened:
  Code was changed. docker compose restart was run.
  The old binary was still running. Changes had no effect.
  Diagnosis took 45 minutes across multiple sessions.

Root cause:
  docker compose restart does not rebuild containers.
  docker compose build && docker compose up -d does.

FAANG standard:
  After any code change: build first, then restart.
  docker compose build scoring-service && docker compose up -d scoring-service
  Never assume restart = rebuild.

Law: restart ≠ rebuild. Always build after code changes.

──────────────────────────────────────────────────────────────────────────────
RC-06: ASYNC PIPELINE SCORED SYNCHRONOUSLY IN TESTS
──────────────────────────────────────────────────────────────────────────────

What happened:
  Kong sends events asynchronously to /event.
  Tests checked scores immediately after sending events.
  Scores had not updated yet — consumer hadn't processed the queue.
  Tests failed intermittently. Root cause took hours to identify.

Root cause:
  The pipeline is: Kong → /event → PostgreSQL queue → consumer → Redis.
  There is intentional latency at every step.
  Tests assumed synchronous scoring.

FAANG standard:
  Async pipelines require async tests.
  Either: add explicit sleep + retry in tests, or
  test the consumer in isolation from the API layer.
  Never test async output with synchronous expectations.

Law: Async systems require async tests. Never assume immediate consistency.

──────────────────────────────────────────────────────────────────────────────
RC-07: DOCUMENT COMMITTED IN SESSION, NEVER GIT ADD'd
──────────────────────────────────────────────────────────────────────────────

What happened:
  pilot-letter-of-understanding.md was written in a session.
  It was never git add'd.
  Next session: the file did not exist in the repo.
  Expert panel found this on ls docs/enterprise/.
  Document had to be recreated from scratch.

Root cause:
  Assumption that writing a file = committing a file.
  git add is a separate step that is easy to forget.

FAANG standard:
  End of every session checklist must include:
  git status → verify all intended files are staged.
  git add -A is the safe default. Then review git diff --staged.

Law: Written ≠ committed. git status before every session close.

──────────────────────────────────────────────────────────────────────────────
RC-08: HARDCODED BASELINES IN PRODUCTION CODE
──────────────────────────────────────────────────────────────────────────────

What happened:
  consumer.go used a hardcoded map for z-score baselines.
  agent_baselines table was never written to.
  Welford's online algorithm never ran in production.
  Discovered March 31 — after weeks of running.
  The data moat was not accumulating.

Root cause:
  Baseline logic was designed correctly.
  Implementation used a local map instead of calling BaselineStore.
  The test passed because the test used the same hardcoded map.
  The production database was empty.

FAANG standard:
  Test the integration path, not the unit path.
  A unit test passing does not prove the database is being written.
  Verify data accumulation: SELECT COUNT(*) FROM agent_baselines;
  This must be non-zero after any scoring activity.

Law: Test the database, not the code. Code can lie. The database cannot.

──────────────────────────────────────────────────────────────────────────────
RC-09: CROSS-TENANT BASELINE CONTAMINATION
──────────────────────────────────────────────────────────────────────────────

What happened:
  baseline.go keyed baselines on (agent_did, feature_name) only.
  No org_id in the key.
  Agent from org A could contaminate baseline of same-named agent in org B.
  Discovered March 31 during TW-PRE-2 verification.
  Required: new UNIQUE constraint, new BaselineStore signatures, migration.

Root cause:
  Multi-tenancy was assumed at the application layer (auth checked org).
  But the data layer had no tenant isolation.
  Application auth and data isolation are separate concerns.

FAANG standard:
  Every data entity in a multi-tenant system must carry its tenant ID.
  Not just at auth time. At storage time. At query time. At index time.
  Verify: UNIQUE constraint includes org_id.

Law: Multi-tenancy must be enforced at the data layer, not just at auth.

──────────────────────────────────────────────────────────────────────────────
RC-10: DEMO RUNTIME CLAIM WAS WRONG
──────────────────────────────────────────────────────────────────────────────

What happened:
  Lloyd_meeting_prep.md said "12 seconds demo runtime."
  Demo was rewritten to streaming output — actual runtime: ~30s Windows Docker.
  The claim in the prep doc was never updated.
  Contradiction discovered by audit — March 31.

Root cause:
  A metric was stated in a document.
  The underlying system changed.
  The document was not updated.
  Two sources of truth diverged.

FAANG standard:
  Metrics in documents must be verified by running the system.
  Never copy-paste a metric. Always run: time bash scripts/demo.sh
  Single source of truth: measure it, then write it.

Law: Measure first. Write second. Never write first and assume.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 2 — WHAT PRODUCED FAANG-GRADE OUTPUT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

These are the specific practices that produced 94/100 hardening score,
0.00% FP rate, and APEX TEST composite 84/100 enterprise-ready.
Not generic advice. Exact practices with exact outcomes.

──────────────────────────────────────────────────────────────────────────────
FG-01: ADVERSARIAL EXPERT PANEL BEFORE ENTERPRISE CONVERSATION
──────────────────────────────────────────────────────────────────────────────

What happened:
  5-expert panel ran 50+ adversarial commands against the live system
  before the first Lloyd conversation.
  Found: 8 real bugs that would have destroyed enterprise trust on day 1.
  Score corrected from self-assessed 82 → honest 67 → fixed to 94.

Why it worked:
  The panel simulated the exact questions a CISO security team asks.
  No assumptions. Actual commands. Actual output. Actual judgment.

FAANG standard:
  Before any enterprise conversation:
  Run the adversarial panel. Take every finding seriously.
  The panel is cheaper than a failed pilot.

Law: Self-assessment is always optimistic. Run the adversarial panel.

──────────────────────────────────────────────────────────────────────────────
FG-02: FAIL-OPEN VS FAIL-CLOSED DOCUMENTED WITH RATIONALE
──────────────────────────────────────────────────────────────────────────────

What happened:
  Infrastructure failure → fail-open (availability preserved)
  Enforcement failure → fail-closed (no unexplained blocks)
  Both documented explicitly with rationale in operational-safety-architecture.md

Why it worked:
  A CISO reading these behaviors without rationale perceives contradiction.
  With rationale: each behavior is correct for its failure class.
  The document pre-empts the question before it's asked.

FAANG standard:
  Every system has two failure classes: infrastructure and logic.
  Document both. Explain why each fails the way it does.
  Never let a buyer discover your failure behavior without explanation.

Law: Document failure behavior before it's discovered in production.

──────────────────────────────────────────────────────────────────────────────
FG-03: HASH CHAIN AT DB LEVEL, NOT APPLICATION LEVEL
──────────────────────────────────────────────────────────────────────────────

What happened:
  Hash chain links enforcement_decisions via SHA-256 prev_hash.
  verify_hash_chain() is a PostgreSQL function, not application code.
  INSERT-only enforced via REVOKE at migration level.
  Audit trail is cryptographically verifiable without trusting the app.

Why it worked:
  Enterprise buyers don't trust vendor attestation.
  They trust math. SHA-256 + INSERT-only + PostgreSQL function =
  verifiable by their own DBAs without involving ARE.

FAANG standard:
  Audit guarantees must be verifiable by the buyer's own team.
  Never ask a buyer to trust your word. Give them the SQL to verify.

Law: Audit claims must be independently verifiable. Give them the query.

──────────────────────────────────────────────────────────────────────────────
FG-04: AUTO-ROLLBACK BEFORE ENFORCE MODE
──────────────────────────────────────────────────────────────────────────────

What happened:
  ModeController + StartFPMonitor() implemented before enforce mode.
  FP spike → auto-rollback in 5 minutes.
  TestAutoRollback: seeds 30% FP, verifies rollback fires.
  This was Gate 0 before any pilot goes live.

Why it worked:
  The #1 pilot killer is a false positive that blocks a legitimate workflow.
  Without auto-rollback: FP spike = 2-4 hours of legitimate agents blocked
  = pilot ends permanently.
  With auto-rollback: FP spike detected in 5 minutes = pilot survives.

FAANG standard:
  Any enforcement system must have an automatic safety valve.
  The safety valve must be tested before enforcement is enabled.
  Never enable enforcement without auto-rollback tested and passing.

Law: Auto-rollback is not a feature. It is a prerequisite.

──────────────────────────────────────────────────────────────────────────────
FG-05: STREAMING DEMO THAT SHOWS DETECTION HAPPENING
──────────────────────────────────────────────────────────────────────────────

What happened:
  Original demo: batch output, 12 seconds, result appeared at end.
  Rewritten: streaming output, ~60 seconds, every step visible.
  Audience sees JWT verification → scoring → anomaly → BLOCK → reason object.
  Nothing is asserted. Everything is demonstrated.

Why it worked:
  A demo that prints nothing for 12 seconds feels broken.
  A demo that streams results for 60 seconds feels alive.
  The viewer understands the product by watching it, not by reading slides.

FAANG standard:
  Demos must show the system thinking, not just the result.
  Stream output. Add pacing. Let the viewer see every step.
  Never batch results and display at the end.

Law: Show the work. Never just show the answer.

──────────────────────────────────────────────────────────────────────────────
FG-06: MEASURE BEFORE CLAIMING — FOUR REPRODUCIBLE NUMBERS
──────────────────────────────────────────────────────────────────────────────

What happened:
  Four metrics established with exact reproduction commands:
  FP: go test ./tests/fp_scenarios/...
  TP: go test ./tests/eval_harness/...
  Slow-walk: go test ./tests/attack_corpus/... -run TestSlowWalk
  Demo: time bash scripts/demo.sh
  These can be run by anyone, anytime, against the live system.

Why it worked:
  Enterprise buyers ask for evidence, not claims.
  A number with a reproduction command is evidence.
  A number without one is a claim.

FAANG standard:
  Every metric in a sales conversation must have a one-line
  reproduction command that anyone can run.
  If you cannot reproduce it on demand, you cannot claim it.

Law: A metric without a reproduction command is a claim, not evidence.

──────────────────────────────────────────────────────────────────────────────
FG-07: IP ANCHORED BEFORE ANY COMMERCIAL CONVERSATION
──────────────────────────────────────────────────────────────────────────────

What happened:
  Zenodo DOI published (10.5281/zenodo.19169185) before Lloyd outreach.
  IP chain: ATP (Dec 2025) → ATG (Dec 2025) → AgentRepEngine (Mar 2026).
  Timestamp is permanent, public, and citable.
  Check Point/Lakera cannot claim prior art after this date.

Why it worked:
  Enterprise buyers and investors check IP position.
  A DOI is a credible timestamp that a solo founder can create.
  It anchors the invention date permanently.

FAANG standard:
  Publish IP anchor before any commercial conversation.
  Zenodo DOI costs nothing. Takes 30 minutes. Lasts forever.
  Do this on Day 1 of any new product with novel claims.

Law: Anchor IP before first commercial conversation. Always.

──────────────────────────────────────────────────────────────────────────────
FG-08: REASON OBJECT ON EVERY ENFORCEMENT DECISION
──────────────────────────────────────────────────────────────────────────────

What happened:
  Every BLOCK/ALLOW/RESTRICT produces a structured JSON reason object.
  Fields: decision, agent_did, score, confidence_pct, policy_fired,
          trigger_events, recommended_action, computed_at.
  G-EXPLAIN gate: fails if any field is missing.
  FM2 prevention: returns error → AUDIT not BLOCK if generation fails.

Why it worked:
  An unexplained block in an enterprise pilot creates a support ticket.
  A block with a reason object creates a conversation.
  The security engineer reads the reason and explains it themselves.
  The product becomes self-defending.

FAANG standard:
  Every automated decision that affects a user must have
  a human-readable explanation attached.
  Not optional. Not Phase 2. Day 1.

Law: Never block without explanation. Unexplained blocks destroy trust.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 3 — WHAT DILUTED PREVIOUS WORK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

These are specific incidents where new work weakened or contradicted
something that was already working correctly.

──────────────────────────────────────────────────────────────────────────────
DIL-01: DEMO RUNTIME STATED IN DOCS, SYSTEM CHANGED, DOCS NOT UPDATED
──────────────────────────────────────────────────────────────────────────────

Effect: Lloyd_meeting_prep.md said 12 seconds. Demo ran 30 seconds.
        Two documents contradicted each other. Discovered by audit.
Prevention: Single source of truth for metrics. Measure → write.
            Never write a metric in a doc without noting the command
            that reproduces it.

──────────────────────────────────────────────────────────────────────────────
DIL-02: HARDCODED BASELINES MASKED WELFORD'S ALGORITHM
──────────────────────────────────────────────────────────────────────────────

Effect: Welford's online algorithm was claimed as the scoring mechanism.
        It was implemented correctly. But consumer.go used a hardcoded map
        that bypassed it entirely. The claim was true in code. False in production.
Prevention: Verify the data layer after any algorithm implementation.
            SELECT COUNT(*) FROM agent_baselines must be non-zero.

──────────────────────────────────────────────────────────────────────────────
DIL-03: SIEM STORY CLAIMED BEFORE SIEM WAS WIRED
──────────────────────────────────────────────────────────────────────────────

Effect: SIEM was a selling point. siem.go existed.
        SendBlocked() had zero callers. Nothing fired.
        Enterprise conversation would have failed on pilot day 1.
Prevention: grep for callers before any claim.
            grep -rn "FunctionName" | grep -v "file_where_defined"
            Must return non-empty.

──────────────────────────────────────────────────────────────────────────────
DIL-04: COMPETITIVE POSITIONING DOC WRITTEN BEFORE STRONGEST CLAIMS EXISTED
──────────────────────────────────────────────────────────────────────────────

Effect: competitive-positioning.md created March 24.
        M6 (confidence_pct), M9 (audit-native), C12 (regulatory inevitability)
        all emerged after. The doc doesn't include the strongest claims.
        If Lloyd shares it with a client's security team — weakest version sent.
Prevention: After any new multiplier or major claim — update positioning doc.
            Positioning docs are never done. They trail the product.

──────────────────────────────────────────────────────────────────────────────
DIL-05: TASKS PLANNED BUT NEVER EXECUTED — PLANS ACCUMULATED OVER PLANS
──────────────────────────────────────────────────────────────────────────────

Effect: M2, M3, M5, M7 decided 3 weeks ago. Still unexecuted.
        PL-2, PL-3, PL-5 — in queue for weeks. Never touched.
        Each new planning session added new tasks to an existing backlog.
        The backlog grew. Execution rate stayed flat.
Prevention: Before any new planning session — close or explicitly defer
            every open task from prior sessions.
            A task that has been open for 2+ weeks without execution
            should be closed (not deferred) unless it's explicitly
            gated on an external event.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 4 — THE CORE PSYCHOLOGICAL PATTERN AND HOW TO BREAK IT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

This pattern was identified in session 13 (March 27) and confirmed
across 8 subsequent sessions. It is the single most important finding
in this entire document.

THE PATTERN:
  High-quality internal production (specs, frameworks, versioned documents,
  meta-documents, learning systems) substitutes for external customer contact.
  
  The production feels like progress. It is not.
  It is a sophisticated permission structure for delaying commercial action.

HOW TO DETECT IT:
  Signal 1: A new framework, file, or system is being built in response
            to a commercial gap.
  Signal 2: Feed scrolling or finding new engagement targets instead of
            executing existing outreach.
  Signal 3: A task that has been in the queue for 2+ weeks is being
            re-planned instead of executed or closed.

THE HARD STOP:
  When the pattern is named explicitly, it breaks.
  The antidote is a direct question: "Does this move T8 forward?"
  If no → stop. Do the T8 work instead.

THE COMPOUNDING TRUTH:
  One signed LoU compounds into everything.
  All the frameworks, learnings, multipliers, claims — none of them
  compound until there is a customer.
  The customer is the only thing that makes everything else real.

Law: Production without a customer is inventory, not value.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 5 — THE FAANG AND BEYOND-FAANG GRADE STANDARD
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

What FAANG grade actually means in this context:
  A FAANG security engineer, seeing the product for the first time,
  cannot find a hole that destroys the core claim.
  
  Not: the code is perfect.
  Not: the architecture is complete.
  Means: the core claims survive adversarial inspection.

What beyond-FAANG means:
  The product is verifiable by the buyer without trusting the vendor.
  DB permissions, hash chains, reproduction commands, open repo.
  The buyer's DBAs, security engineers, and auditors can all verify
  independently. The vendor's attestation is not required.

The five things that separate beyond-FAANG from FAANG:

  1. Every claim has a reproduction command anyone can run
  2. Every security guarantee is enforced at DB/OS/network level
  3. Every enforcement decision is independently verifiable
  4. Every failure mode is documented with rationale before production
  5. Auto-rollback is tested before enforcement is enabled

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 6 — THE LAWS FOR FUTURE PRODUCTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Apply these to every future product. No exceptions.

QI-L1  Read before write. Always read the existing code before adding to it.
       Read the function. Read its callers. Read the test. Then write.

QI-L2  A function with zero callers does not exist in production.
       Verify callers before claiming a feature.

QI-L3  Security guarantees belong at the DB/OS/network layer.
       Application code is not a security boundary.

QI-L4  Empty string is not "not set." It is "open."
       All security env vars must fail closed on empty string.

QI-L5  restart ≠ rebuild. Always build after code changes.

QI-L6  Async systems require async tests.
       Never test async output with synchronous expectations.

QI-L7  git status before every session close.
       Written ≠ committed. Committed ≠ pushed.

QI-L8  Multi-tenancy must be enforced at the data layer.
       Not just at auth. At storage, query, and index level.
       Every table row must carry its tenant ID.

QI-L9  Measure first. Write second.
       Every metric in a document must have a reproduction command.
       Run it before writing it.

QI-L10 Anchor IP before first commercial conversation.
       Zenodo DOI. 30 minutes. Permanent.

QI-L11 Auto-rollback is a prerequisite, not a feature.
       Test it before enabling enforcement. Always.

QI-L12 Never block without explanation.
       Every automated decision must have a human-readable reason.

QI-L13 Audit claims must be independently verifiable.
       Give the buyer the SQL to verify. Never ask them to trust you.

QI-L14 Document failure behavior before it's discovered in production.
       Both failure classes: infrastructure and logic.
       Both documented with rationale.

QI-L15 Self-assessment is always optimistic.
       Run the adversarial panel before any enterprise conversation.
       The panel finds what you cannot see yourself.

QI-L16 Show the work. Never just show the answer.
       Streaming demo. Every step visible.
       Let the viewer understand the product by watching it.

QI-L17 Production without a customer is inventory, not value.
       Every planning session must answer: does this move T8 forward?
       If no — do the T8 work instead.

QI-L18 After any new major claim — update the positioning docs.
       Positioning docs trail the product. Close the gap within 24 hours.

QI-L19 Close tasks before opening new ones.
       A task open for 2+ weeks without execution is either closed
       or explicitly gated on a named external event.
       Never plan over an unexecuted backlog.

QI-L20 The customer is the only thing that makes everything else real.
       One signed LoU compounds into everything.
       Nothing compounds without it.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 7 — SESSION-START QUALITY GATE (run before every session)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Run these before any build work. Takes 5 minutes.
Catches RC-05, RC-02, RC-03, RC-08, RC-10 before they compound.

# Stack health
docker compose ps | grep -E "healthy|unhealthy"

# FP rate (must be 0.00%)
go test ./tests/fp_scenarios/... 2>&1 | tail -3

# Callers check on any new function before claiming it
grep -rn "FunctionName" internal/ | grep -v "file_where_defined"

# INSERT-only enforced
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
"SELECT privilege_type FROM information_schema.role_table_grants
 WHERE table_name='enforcement_decisions' AND grantee='are';"
# Must show INSERT, SELECT only

# Baseline accumulation (non-zero means Welford's is running)
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
"SELECT COUNT(*) FROM agent_baselines;"
# Must be non-zero after any scoring activity

# Demo runtime (measure, don't assume)
time bash scripts/demo.sh 2>&1 | tail -3

# Hash chain valid
curl -s http://localhost:8080/health | jq '.hash_chain_valid'
# Must return true

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 8 — HOW TO USE THIS DOCUMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Before starting any new product:
  Read Part 1 (root causes) and Part 6 (laws).
  Apply every law from Day 1.

Before any enterprise conversation:
  Run Part 7 (session-start quality gate).
  Run adversarial panel (FG-01).

After any rework incident:
  Add a new entry to Part 1.
  Add a new law to Part 6 if not already covered.

After any FAANG-grade win:
  Add to Part 2.
  Identify which law it confirms.

After any pilot or customer interaction:
  Add learnings to Part 3 and Part 4.
  Update laws if the customer revealed a gap the laws didn't cover.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF QUALITY_INTELLIGENCE v1.0
10 root causes | 8 FAANG-grade practices | 5 dilution incidents
20 laws | 1 core psychological pattern | 1 session-start gate
Source: 20 sessions | 80+ commits | March 17–31, 2026
Next version: after first pilot closes
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━