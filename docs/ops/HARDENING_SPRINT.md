# HARDENING_SPRINT.md — AgentRepEngine
# "Zero Embarrassment" — FAANG-Grade Hardening Plan
# Created: April 16, 2026
# Authority: This file is the SOLE task list for hardening work.
#            No session creates new hardening tasks outside this file.
#            Every session opens by reading this file and picking up
#            the next unchecked item. No replanning. No re-diagnosing.
#            Check the box. Fill in the commit hash. Move to the next.
#
# RULE 1: Every item has a verification gate. Item is not done until
#         the gate passes. Gate = go test + demo.sh.
# RULE 2: If an item breaks the gate, revert immediately. Do not
#         debug for more than 30 minutes. Mark BLOCKED with reason.
# RULE 3: No item modifies existing test assertions. Tests are the
#         ground truth. New tests are additive only.
# RULE 4: Session opens → read this file → find first unchecked item
#         → implement → verify gate → check box → next item.
# RULE 5: This file is committed to the repo after every session.
#         It is the hardening equivalent of CONTINUATION_PROMPT.md.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BASELINE — VERIFIED BEFORE ANY WORK BEGINS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Date verified: April 16, 2026
HEAD at start: 876f699
Demo passes: YES (54 seconds, all 10 steps)
Unit tests: 12/14 packages green (tests/ and tests/kong/ require Docker)
FP rate: 0.00% (0/150)
TP rate: 88.00% (44/50)
F1: 0.9362
Precision: 100%
Slow-walk: 100% (10/10)
Hardening score: 96/100

These numbers must hold after every single item below.
If any number regresses, the item that caused it is reverted.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE A — DEFENSIVE INFRASTRUCTURE
No scoring path changes. Zero risk to existing metrics.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] A1 — YAML policy schema validation on startup
    File: internal/scoring/config.go (add validation), policy.go (add struct tags)
    What: Define strict Go struct with `validate` tags for every policy field.
          On startup, load all YAMLs, validate, refuse to start if invalid.
          Add test for malformed YAML → startup failure.
    Why compounds: Pilot engineer deploys with typo → ARE crashes with no
          explanation → support call → trust destroyed. Schema validation
          catches it with a clear error message naming the exact field.
    Risk: ZERO — startup only, no runtime change.
    Trade-off: Slightly slower startup (~50ms). Acceptable.
    5-Why: Why validate? → Prevent misconfiguration crash. Why not before? →
          Phase 1 had single deployer. Why now? → Pilot means external deployer.
    Gate: go test ./internal/scoring/... + demo.sh
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%

[ ] A2 — Rate limit on /event — per-agent-DID 100 events/min
    File: cmd/scoring-service/main.go (add middleware)
    What: Per-agent-DID rate limit. Return 429 + Retry-After header.
          Document ceiling in security-attestation.md.
          Add test: 101st event in 60 seconds returns 429.
    Why compounds: Pen tester floods /event → queue grows unbounded → OOM →
          enforcement runs on stale data. 429 is the professional answer.
    Risk: ZERO — additive middleware before event handler.
    Trade-off: Legitimate high-frequency agents capped at 100/min. Configurable.
    5-Why: Why 100/min? → Highest legitimate rate in corpus is ~6/min (120/hr ÷ 60).
          100/min gives 16x headroom. Why per-DID? → Fleet-level limit would
          let one agent's flood block another agent's events.
    Gate: go test ./cmd/... + demo.sh
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%

[ ] A3 — Redis memory alert + write return code checks
    File: internal/scoring/consumer.go (check write returns),
          internal/metrics/metrics.go (add redis_memory_pct gauge)
    What: Prometheus gauge `redis_memory_pct`. Alert threshold at 80%.
          Every HSET/SET call checks return code. Log warning on failure.
          Add test: simulated write failure → warning logged, not silent.
    Why compounds: Redis fills up silently → scores stop updating →
          enforcement runs on stale data → undetected drift.
    Risk: ZERO — monitoring only.
    Trade-off: One INFO-level Redis call per scoring cycle to check memory.
    5-Why: Why not before? → Single-machine dev never hit memory limits.
          Why now? → Pilot runs 30 days, memory growth is real.
    Gate: go test ./internal/scoring/... + go test ./internal/metrics/... + demo.sh
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%

[ ] A4 — SIEM webhook delivery confirmation
    File: internal/audit/siem.go (add retry + tracking)
    What: Add pending_siem_events table (migration). Deliver with exponential
          backoff (3 retries, 30/60/120s). Mark delivered on 2xx.
          Prometheus counter: siem_delivery_failures_total.
          Add test: simulated 5xx → 3 retries → failure metric incremented.
    Why compounds: SIEM alert fires but never arrives → SOC misses incident →
          ARE blamed. Delivery tracking = SOC 2 CC7.2 evidence.
    Risk: LOW — additive table + retry logic. Does not change scoring path.
    Trade-off: Slightly higher DB write volume (one row per SIEM event).
    5-Why: Why retry? → Network failures are normal in enterprise. Why 3? →
          Industry standard. Why exponential? → Prevents thundering herd.
    Gate: go test ./internal/audit/... + demo.sh
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%

[ ] A5 — jti cache Redis restart recovery
    File: internal/identity/jwt.go (add Postgres recovery),
          migrations/ (add jti_cache table)
    What: Persist all issued jtis to Postgres jti_cache table (TTL = JWT expiry).
          Redis = fast lookup. Postgres = recovery source.
          On Redis restart, rebuild jti cache from Postgres.
          Add test: simulate Redis restart → jti cache rebuilt → replay blocked.
    Why compounds: Redis restarts during pilot → jti cache lost → replay
          window opens until TTLs rebuild. Postgres closes the window.
    Risk: LOW — additive table + recovery logic. Does not change JWT flow.
    Trade-off: Dual-write (Redis + Postgres) on every JWT verification.
    5-Why: Why dual-write? → Redis is volatile. Postgres is durable.
          Why not Postgres only? → Latency. Redis is the hot path.
    Gate: go test ./internal/identity/... + demo.sh
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE B — SCORING PATH HARDENING
These touch the scoring pipeline. Extra verification required.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] B1 — Welford MULTI/EXEC atomic baseline updates
    File: internal/scoring/baseline.go (wrap in Redis Lua script)
    What: All baseline read-modify-write operations wrapped in Redis
          MULTI/EXEC or Lua atomic script. Prevents race condition where
          two concurrent events for same agent corrupt the baseline.
          Add test: 50 concurrent goroutines updating same agent → no race.
          MUST PASS: go test -race ./internal/scoring/...
    Why compounds: Two concurrent events → race condition → corrupted
          baseline → cascading FP or FN. This is the single most dangerous
          open bug in the codebase.
    Risk: MODERATE — changes baseline write path. Could introduce new bugs.
    Trade-off: Slightly higher Redis latency per baseline update (~1ms).
    5-Why: Why atomic? → Welford's algorithm is read-modify-write. Two
          concurrent reads of the same mean/variance → one write overwrites
          the other → baseline drifts from reality. Why Lua? → Redis MULTI
          doesn't help for read-then-write; Lua script is truly atomic.
    Gate: go test -race ./internal/scoring/... + go test ./tests/eval_harness/...
          + demo.sh. FP and TP must not change.
    Commit: _______________
    Verified: [ ] tests green [ ] race clean [ ] demo passes [ ] FP 0.00% [ ] TP 88%

[ ] B2 — Decay constant to config — per-agent-type profiles
    File: config/scoring_weights.yaml (add decay_profiles section),
          internal/scoring/scorer.go (read profile), consumer.go (apply profile)
    What: Move decay_rate from single value (0.1) to per-agent-type:
          batch: 0.02 (slow decay — weekly agents keep reputation)
          realtime: 0.1 (current default)
          monitoring: 0.05 (moderate — health check agents)
          Agent type inferred from event patterns or explicit config.
    Why compounds: Batch agent runs once per week → H decays to near-zero
          in 7 days → next week's batch run scores RESTRICTED → FP.
          Configurable decay eliminates an entire class of false positives.
    Risk: MODERATE — changes scoring behavior. Must verify FP harness.
    Trade-off: Requires agent-type classification (can default to "realtime").
    5-Why: Why per-type? → Different agents have fundamentally different
          activity cadences. One decay rate cannot serve all. Why not before? →
          Phase 1 had no production agents to calibrate against.
    Gate: go test ./tests/eval_harness/... -v (FP and TP must not change)
          + go test ./tests/fp_scenarios/... + demo.sh
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%
    Note: If FP regresses, revert immediately. Document which scenarios failed.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE C — ENGINEERING EXCELLENCE
What makes a FAANG engineer say "they know what they're doing."
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] C1 — Fix integration tests — skip when no Docker
    File: tests/layer2_integration_test.go, tests/kong/jwt_verification_test.go
    What: Add skip condition: if scoring service / Kong not reachable,
          t.Skip("requires running Docker stack"). This makes `go test ./...`
          show 14/14 green (with skips) instead of 12 pass + 2 FAIL.
    Why compounds: Anyone who clones the repo and runs `go test ./...` sees
          2 FAIL packages. First impression: broken. Skip + message = professional.
    Risk: ZERO — only changes test skip behavior.
    Gate: go test ./... shows 0 FAIL (some SKIP is fine)
    Commit: _______________
    Verified: [ ] 0 FAIL packages [ ] demo passes

[ ] C2 — go test -race clean on entire codebase
    File: any file with race conditions
    What: Run `go test -race ./...`. Fix every race detected.
          This is Google's pre-submit requirement for all Go code.
    Why compounds: Race conditions are the bugs that only appear in
          production under load. Fixing them now prevents pilot failures.
    Risk: LOW — fixes are typically mutex additions or channel fixes.
    Gate: go test -race ./... — zero races
    Commit: _______________
    Verified: [ ] zero races [ ] demo passes

[ ] C3 — go vet + staticcheck clean
    File: any file with warnings
    What: Run `go vet ./...` and `staticcheck ./...`. Fix all warnings.
          Add to README: "This codebase passes go vet and staticcheck
          with zero warnings."
    Why compounds: Static analysis catches bugs that tests miss. Zero
          warnings = confidence that no obvious defects remain.
    Risk: ZERO — fixes are typically unused variables, incorrect formats.
    Gate: go vet ./... + staticcheck ./... — zero warnings
    Commit: _______________
    Verified: [ ] zero warnings [ ] demo passes

[ ] C4 — GoDoc coverage — all exported functions documented
    File: all .go files with exported functions
    What: Every exported function, type, and constant has a doc comment.
          Run `go doc ./...` and verify no "missing comment" warnings.
    Why compounds: Open source credibility. When Kong's engineering team
          reads the code, documented APIs signal maturity.
    Risk: ZERO — comments only.
    Gate: go doc spot check on scoring, enforcement, identity, audit packages
    Commit: _______________
    Verified: [ ] all exported funcs documented

[ ] C5 — main.go refactor — extract handlers
    File: cmd/scoring-service/main.go → split into main.go + handlers.go
          + middleware.go + config.go
    What: Extract HTTP handlers, middleware (auth, rate limit), and config
          loading into separate files. main.go becomes ~50 lines: load config,
          wire dependencies, start server.
    Why compounds: New engineer onboarding. Kong's team reviewing code.
          Clean entry point = "this team writes production code."
    Risk: MODERATE — refactor of entry point. Must not change behavior.
    Gate: go test ./... + demo.sh — exact same behavior before and after.
    Commit: _______________
    Verified: [ ] tests green [ ] demo passes [ ] FP 0.00% [ ] TP 88%

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE D — OFFENSIVE VALIDATION
Prove ARE catches what it claims. Red team your own product.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] D1 — Adversarial evasion test suite — 10 scenarios
    File: tests/adversarial/ (new package)
    What: 10 scenarios testing known evasion techniques:
          1. Gradual escalation (stay below z=3.0 per-step)
          2. Feature vector rotation (spike different features each hour)
          3. Timing manipulation (burst during maintenance windows)
          4. Coordinated multi-agent (3 agents each below threshold, combined above)
          5. Baseline poisoning via slow drift (1% per day for 30 days)
          6. Score recovery gaming (behave well to rebuild score, then attack)
          7. JWT replay with valid jti (within TTL window)
          8. Event type spoofing (send tool_call labeled as http_request)
          9. Rate limit evasion (99 events/min sustained)
          10. Fail-open exploitation (flood scoring service to trigger circuit breaker)
    Why compounds: These are the tests a real red team would run. Having
          them in the repo means ARE has already been adversarially tested.
    Risk: ZERO — new test package only. Does not change production code.
    Gate: tests pass (some scenarios SHOULD be caught, some document known limits)
    Commit: _______________
    Verified: [ ] tests run [ ] results documented

[ ] D2 — Benchmark suite with published numbers
    File: internal/scoring/benchmark_test.go (already exists — expand)
    What: Benchmark score computation, event processing, enforcement decision.
          Publish p50/p95/p99 latency in README under "Performance" section.
          Target: score computation < 1ms p99, event processing < 5ms p99.
    Why compounds: Published benchmarks are what separates serious infra
          from hobby projects. Kong's team will look for this.
    Risk: ZERO — benchmarks only.
    Gate: go test -bench ./internal/scoring/... -benchmem
    Commit: _______________
    Verified: [ ] benchmarks documented in README

[ ] D3 — End-to-end integration test with Docker Compose
    File: tests/e2e/ (new), Makefile or scripts/test-e2e.sh
    What: Single command that: starts Docker stack, waits for healthy,
          runs all 10 demo steps programmatically (not bash — Go test),
          verifies all assertions, tears down.
          `make test-e2e` or `bash scripts/test-e2e.sh`
    Why compounds: Proves the demo isn't fragile. CI-ready. Any engineer
          can verify the entire product works with one command.
    Risk: LOW — new test only. Does not change production code.
    Gate: test-e2e passes on clean Docker start
    Commit: _______________
    Verified: [ ] e2e passes

[ ] D4 — Security scan — gosec + dependency audit
    File: README.md (add security section), .github/ (optional CI)
    What: Run `gosec ./...`. Fix any HIGH/MEDIUM findings.
          Run `go mod verify` — passes.
          Run `govulncheck ./...` — no known CVEs.
          Document in security-attestation.md.
    Why compounds: Enterprise security review will run these tools.
          Clean results = pre-approved. Findings = delayed pilot.
    Risk: ZERO — scan + fix only.
    Gate: gosec clean, govulncheck clean
    Commit: _______________
    Verified: [ ] gosec clean [ ] govulncheck clean

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE E — DOCUMENTATION HARDENING
What the repo looks like when someone opens it for the first time.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] E1 — README.md rewrite — first-impression quality
    File: README.md
    What: Architecture diagram (Mermaid), quick start (3 commands),
          performance benchmarks (from D2), test coverage summary,
          regulatory mapping table, published DOI link.
          Someone reading this README should understand ARE in 2 minutes.
    Risk: ZERO — documentation only.
    Commit: _______________

[ ] E2 — Kong AI Gateway 3.14 technical brief
    File: docs/partnerships/kong-ai-gateway-integration.md
    What: 2-page technical brief: ARE as behavioral enforcement layer
          for Kong AI Gateway 3.14. Maps ARE capabilities to Kong's
          stated A2A governance gaps. Gravity Action 3.
    Risk: ZERO — documentation only.
    Commit: _______________

[ ] E3 — Updated security-attestation.md
    File: docs/enterprise/security-attestation.md
    What: Incorporate all hardening items from this sprint as evidence.
          gosec clean, go vet clean, race-free, rate-limited, schema-validated.
    Risk: ZERO — documentation only.
    Commit: _______________

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EXECUTION LOG — FILL IN AS ITEMS COMPLETE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

| Date | Item | Commit | Tests | Demo | FP | TP | Notes |
|------|------|--------|-------|------|----|----|-------|
| | | | | | | | |

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SESSION PROTOCOL FOR HARDENING SESSIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Session opens → "HARDENING SPRINT — read HARDENING_SPRINT.md"
2. Claude reads this file, finds first unchecked item
3. Claude reads the relevant source file(s)
4. Claude implements the item
5. Human runs gate commands, pastes output
6. If gate passes → check box, fill commit hash, move to next
7. If gate fails → revert, mark BLOCKED with reason
8. Session closes → commit this file with updated checkboxes
9. Next session opens → step 1 again

NO REPLANNING. NO RE-DIAGNOSING. NO NEW ITEMS WITHOUT APPROVAL.
The plan is set. Execute it.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL EFFORT ESTIMATE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase A (Defensive Infrastructure):  12 hrs
Phase B (Scoring Path Hardening):     6 hrs
Phase C (Engineering Excellence):    11 hrs
Phase D (Offensive Validation):      13 hrs
Phase E (Documentation):              7 hrs
─────────────────────────────────────────
Total:                               ~49 hrs

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF HARDENING SPRINT PLAN
Created: April 16, 2026 | HEAD: 876f699
Next item: A1 — YAML policy schema validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
