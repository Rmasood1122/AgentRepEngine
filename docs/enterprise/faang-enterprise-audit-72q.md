═══════════════════════════════════════════════════════════════════
AGENTREPENGINE — FAANG & ENTERPRISE GRADE AUDIT
72-Question Meta System Prompt
Version: 1.0 | Date: March 22, 2026
Purpose: Determine if AgentRepEngine is production-grade,
         enterprise-safe, and will never fail in a regulated
         enterprise deployment.
Evaluators: 5 expert archetypes — all adversarial
Instructions: Paste into any Claude session with full project
              context loaded. Answer every question with
              evidence from code, not from claims.
Confidence labels required on every answer:
  [F]   = Fact — confirmed from code or test output
  [OBS] = Observed in spec or docs
  [H]   = High confidence inference
  [INF] = Inferred — not directly confirmed
  [ASS] = Assumption — unverified, must be tested
  [GAP] = Gap — not implemented, not documented
═══════════════════════════════════════════════════════════════════

INSTRUCTION TO CLAUDE:
You are a panel of 5 adversarial expert evaluators. Read every
file in this project. Answer each question with evidence from
the actual codebase — not from the spec, not from claims.
If the answer requires running a command, provide the exact
command. If the answer is unknown, label it [GAP] and explain
what needs to be built or verified. Never state assumptions
as facts. A [GAP] finding is not a failure — it is honest
intelligence that drives the next sprint. A hidden gap
discovered in a CISO meeting is a deal-killer.

For each question:
1. State the answer in one sentence
2. Provide the evidence (file path, line, test name, or command)
3. Apply confidence label
4. If [GAP] or [ASS]: state exact fix required and time estimate

Score each section 0–100. Final composite score:
  FAANG-GRADE:       ≥90
  ENTERPRISE-READY:  ≥80
  APPROACHING:       70–79
  NEEDS WORK:        <70

═══════════════════════════════════════════════════════════════════
SECTION 1 — IDENTITY & AUTHENTICATION (Q1–Q9)
Evaluator: FAANG Principal Security Architect
Failure mode: Any gap here terminates a financial services POC
              in the first 10 minutes of technical review.
═══════════════════════════════════════════════════════════════════

Q1. What signing algorithm is used for agent JWTs and is HS256
    categorically rejected?
    Evidence required: grep for SigningMethodRS256 and any
    HS256 rejection logic in identity package.

Q2. Is the jti (JWT ID) claim present on every token issued
    and is it a UUID? Show the NewAgentClaims function.
    Evidence required: internal/identity/jwt.go — RegisteredClaims.ID

Q3. Is JWT replay detection implemented? Can the same token
    be presented twice and succeed on the second use?
    Evidence required: TestJTIReplayDetection test output.
    Command: go test ./internal/identity/... -run TestJTIReplayDetection -v

Q4. What happens when Redis is unavailable during token
    verification — does the system fail-open or fail-closed?
    Evidence required: VerifyToken Redis error handling path in jwt.go.

Q5. Are all 7 required JWT claims enforced at compile time
    and validated before signing?
    Evidence required: validateClaims() function — every field check.

Q6. What is the behavior when an agent presents a JWT from
    an unknown or unregistered issuer?
    Evidence required: code path in identity verification or
    gateway plugin that handles unknown issuers.

Q7. Is the JWKS endpoint implemented and does it return
    the correct public key for token verification?
    Evidence required: GET /jwks endpoint in scoring service.
    Command: curl http://localhost:8080/jwks

Q8. What is the sub-agent score inheritance rule and is it
    enforced — can a sub-agent inherit a score above 700?
    Evidence required: SubAgentScore() function + test.

Q9. Is identity cycling detection implemented — does the
    system alert when >5 new agent DIDs appear per org per hour?
    Evidence required: CheckIdentityCycling() in identity.go
    + verification it is called in the event pipeline.

═══════════════════════════════════════════════════════════════════
SECTION 2 — GATEWAY ENFORCEMENT (Q10–Q18)
Evaluator: FAANG Principal Security Architect
Failure mode: A bypass at the gateway layer makes the entire
              product worthless. Every path must be covered.
═══════════════════════════════════════════════════════════════════

Q10. Does every agent request pass through the Kong plugin
     before reaching the upstream service — is there any
     path that bypasses enforcement?
     Evidence required: kong/declarative/kong.yml — plugin
     attached to all routes, not just some.

Q11. What are the 4 enforcement bands and their exact score
     thresholds? Are they config-driven or hardcoded?
     Evidence required: config/scoring_weights.yaml or
     policy pack YAML — threshold values.

Q12. When an agent is BLOCKED, does the system return a
     synthetic response instead of a 403? Why does this matter?
     Evidence required: Kong plugin handler.lua — BLOCKED
     response construction.

Q13. Is observe mode the default for new deployments and
     is it impossible to accidentally ship in enforce mode?
     Evidence required: ENFORCEMENT_MODE env var default
     in docker-compose.yml + Kong plugin enforcement_mode config.

Q14. What is the gateway enforcement latency p99 — has it
     been measured under load, not estimated?
     Evidence required: wrk benchmark output or load test
     results — actual numbers, not targets.
     Command: wrk -t4 -c100 -d30s http://localhost:8000/test

Q15. What happens to agent traffic if the scoring service
     goes down — does the gateway fail-open or block all traffic?
     Evidence required: circuit breaker logic in Kong plugin
     handler.lua — scoring service unavailable path.

Q16. Are X-Agent-Score and X-Agent-DID headers injected
     downstream on every request — including allowed requests?
     Evidence required: header injection in Kong plugin
     for all decision outcomes, not just blocks.

Q17. Is rate limiting implemented at the gateway per org_id
     and what is the limit?
     Evidence required: kong/declarative/kong.yml —
     rate-limiting plugin config.

Q18. Has the Kong plugin been tested with a malformed JWT —
     what happens when Authorization header is missing,
     expired, or tampered?
     Evidence required: test cases for malformed JWT handling
     in eval harness or Kong plugin tests.

═══════════════════════════════════════════════════════════════════
SECTION 3 — SCORING ENGINE (Q19–Q27)
Evaluator: Google SRE + AI/ML Systems Engineer
Failure mode: An unexplainable score change or a score that
              cannot be reproduced from inputs destroys
              enterprise trust permanently.
═══════════════════════════════════════════════════════════════════

Q19. What is the exact Phase 1 scoring formula with weights?
     Are weights config-driven and validated to sum to 1.0?
     Evidence required: config/scoring_weights.yaml +
     weight validation on load in scoring service.

Q20. Is the exponential decay model implemented and what
     is the half-life? Show the formula in code.
     Evidence required: scoring formula in internal/scoring/ —
     H(t) = H(t-1) * e^(-0.1 * days) implementation.

Q21. Is z-score calculated per-agent individual baseline
     or globally? How many events are required before
     z-score activates — what happens before that threshold?
     Evidence required: baseline computation in scoring service
     + cold-start handling logic.

Q22. Are all 8 behavioral feature vectors stored on every
     scored event — not just the score?
     Evidence required: L74 query result.
     Command: SELECT COUNT(*) - COUNT(feature_vector) AS missing
     FROM agent_events WHERE created_at > NOW() - INTERVAL '7 days';

Q23. Is score update idempotent — does processing the same
     event_id twice produce the same result?
     Evidence required: idempotency check in event consumer
     or test proving duplicate event_id handling.

Q24. Are score bounds clamped to 0–1000 always enforced
     — can a score ever go below 0 or above 1000?
     Evidence required: Clamp() call in scoring formula.

Q25. What is the peer cluster assignment logic — how are
     agents grouped and what is the fallback for N=1?
     Evidence required: peer cluster assignment in scoring
     service + fallback when no cluster exists.

Q26. Is the scoring service async — does it ever block
     the gateway critical path waiting for a score update?
     Evidence required: async event queue pattern in
     cmd/scoring-service/main.go.

Q27. What is score update latency p95 — has it been
     measured, not estimated?
     Evidence required: Prometheus metric for score update
     duration or benchmark output.
     Command: curl -s http://localhost:9090/api/v1/query?query=score_update_duration_seconds

═══════════════════════════════════════════════════════════════════
SECTION 4 — FALSE POSITIVE GOVERNANCE (Q28–Q33)
Evaluator: Financial Services CISO
Failure mode: One false positive that blocks a legitimate
              production workflow = product ripped out in 30 days.
              This is the single most important section.
═══════════════════════════════════════════════════════════════════

Q28. What is the measured false positive rate on the
     100-scenario legitimate corpus — exact number?
     Evidence required: make test-fp output or
     tests/fp_scenarios/ test results.
     Command: go test ./tests/fp_scenarios/... -v

Q29. Does the FP test suite cover all 6 required legitimate
     scenario types: high-frequency analyst, authorized bulk
     export, multi-step research, new agent in probation,
     recovering agent, off-hours scheduled job?
     Evidence required: tests/fp_scenarios/ — list all
     scenario files and their descriptions.

Q30. Is there a production FP monitoring query that alerts
     when daily FP rate exceeds 2%?
     Evidence required: daily_fp_metrics view in PostgreSQL
     schema or monitoring config.
     Command: SELECT fp_rate FROM daily_fp_metrics ORDER BY date DESC LIMIT 7;

Q31. Is the override rate tracked as a FP proxy — what
     percentage of enforcement decisions get overridden?
     Evidence required: override_rate query on
     enforcement_decisions table.

Q32. What is the HIGH_RISK operation override — which
     operation types require human review regardless of score?
     Evidence required: HIGH_RISK class in YAML policy packs
     + enforcement logic that handles it.

Q33. If FP rate exceeds 2% in production, is there an
     automatic rollback to observe mode or does it require
     manual intervention?
     Evidence required: auto-rollback logic in enforcement
     service or documented runbook.

═══════════════════════════════════════════════════════════════════
SECTION 5 — EXPLAINABILITY (Q34–Q39)
Evaluator: Financial Services CISO + SOC2 Auditor
Failure mode: Any enforcement decision without a complete
              reason object = product is a black box =
              security team will never trust it.
═══════════════════════════════════════════════════════════════════

Q34. Does every enforcement decision produce a complete
     reason object with all required fields — zero nulls?
     Evidence required: make test-explain output.
     Command: go test ./... -run TestExplain -v

Q35. What happens if reason object generation fails at
     runtime — does the system block without explanation
     or fall back to AUDIT mode?
     Evidence required: error handling in explainability
     engine — failsafe behavior.

Q36. Is the reason object stored in the tamper-evident
     audit log and linked to the enforcement decision?
     Evidence required: enforcement_decisions table schema
     — reason_object column + foreign key.

Q37. Is the reason object delivered to a SIEM webhook
     on every BLOCKED decision?
     Evidence required: webhook POST logic in enforcement
     service + SIEM_WEBHOOK_URL config.

Q38. Does the reason object include peer cluster context —
     peer_cluster_avg_score and deviation_from_cluster?
     Evidence required: reason object schema in
     internal/scoring/explainability.go.

Q39. Can a security engineer reproduce exactly why an
     agent was blocked 30 days ago from the audit log alone?
     Evidence required: replay/forensics endpoint —
     GET /audit/replay?agent_did=X&from=Y&to=Z

═══════════════════════════════════════════════════════════════════
SECTION 6 — AUDIT TRAIL & COMPLIANCE (Q40–Q46)
Evaluator: Financial Services CISO + SOC2 Auditor
Failure mode: A tamper-evident log that can be tampered
              with is worse than no log — it creates false
              compliance confidence.
═══════════════════════════════════════════════════════════════════

Q40. Is the enforcement_decisions table INSERT-only at
     the database permission level — can any UPDATE or
     DELETE be executed by the application user?
     Evidence required: PostgreSQL role permissions for
     the 'are' user on enforcement_decisions.
     Command: docker exec postgres psql -U are -c
     "\dp enforcement_decisions"

Q41. Is the hash chain implemented correctly —
     this_hash = sha256(prev_hash + id + timestamp + decision)?
     Show the hash chain verification function.
     Evidence required: verify_hash_chain() SQL function
     in migrations/001_initial.sql.

Q42. Does the hash chain verification currently pass on
     the live database?
     Evidence required: live verification query.
     Command: docker exec postgres psql -U are -d agentrepengine
     -c "SELECT verify_hash_chain('enforcement_decisions') AS valid;"

Q43. Is a SOC2-compatible export endpoint implemented and
     does it include hash chain verification in the output?
     Evidence required: GET /audit/export?format=soc2
     endpoint + response schema.
     Command: curl http://localhost:8080/audit/export?format=soc2

Q44. Is behavioral telemetry limited to metadata only
     by default — no prompt or completion content stored?
     Evidence required: privacy_tier column in agent_events
     + default Tier 1 enforcement in event consumer.

Q45. What NIST AI RMF functions does this product map to
     and is that mapping documented?
     Evidence required: docs/ — NIST mapping document
     or compliance matrix.

Q46. Is there a documented data retention policy for
     behavioral telemetry?
     Evidence required: docs/ — data retention policy
     document + any schema-level TTL enforcement.

═══════════════════════════════════════════════════════════════════
SECTION 7 — INFRASTRUCTURE SECURITY (Q47–Q54)
Evaluator: FAANG Principal Security Architect
Failure mode: A security product with exploitable
              infrastructure is worse than no security product.
═══════════════════════════════════════════════════════════════════

Q47. Is Redis password-protected and is unauthenticated
     access categorically rejected?
     Evidence required: redis-cli without auth returns
     NOAUTH error.
     Command: docker exec redis redis-cli ping

Q48. Is the Redis ACL restricting the scoring service
     to score:* and used_token:* key namespaces only?
     Evidence required: config/redis/users.acl — scoring_service
     user definition.

Q49. Are there any secrets, API keys, or private keys
     committed to the git repository?
     Evidence required: git log --all --full-history -- "*.pem"
     + secret scan.
     Command: git log --all --full-history -- "*.pem" "*.key" "*secret*"

Q50. Is the scoring service running as a non-root user
     inside the Docker container?
     Evidence required: Dockerfile — USER instruction.
     Command: docker exec agentrepengine-scoring-service-1 whoami

Q51. Are all inter-service communications internal to
     the Docker network — no unnecessary external exposure?
     Evidence required: docker-compose.yml — port mappings
     and network configuration.

Q52. Is there a documented rollback procedure that has
     been tested — not just written?
     Evidence required: docs/ — rollback procedure +
     evidence it has been executed successfully.

Q53. What is the DoS-to-fail-open attack surface — if
     an attacker floods the scoring service, do agents
     run freely?
     Evidence required: circuit breaker implementation
     in Kong plugin + rate limiting config.

Q54. Are application logs free of PII — no agent payload
     content, no user data in log output?
     Evidence required: log output scan for PII patterns.
     Command: docker logs agentrepengine-scoring-service-1
     2>&1 | grep -i "pii\|email\|ssn\|password\|payload"

═══════════════════════════════════════════════════════════════════
SECTION 8 — OPERATIONAL READINESS (Q55–Q63)
Evaluator: Google SRE
Failure mode: A product that cannot be operated, monitored,
              or debugged by a security engineer who did not
              build it will be abandoned within 60 days.
═══════════════════════════════════════════════════════════════════

Q55. Is there a reproducible blocked incident demo that
     runs on a clean machine in under 60 seconds?
     Evidence required: scripts/demo.sh — time it on a
     cold stack.
     Command: docker compose down -v && docker compose up -d
     && sleep 30 && time bash scripts/demo.sh

Q56. Is there a health endpoint that returns Redis,
     PostgreSQL, enforcement mode, and log format status?
     Evidence required: GET /health response.
     Command: curl -s http://localhost:8080/health | jq .

Q57. Are Prometheus metrics implemented and flowing to
     Grafana — not just registered, actually scraping?
     Evidence required: Prometheus targets page.
     Command: curl -s http://localhost:9090/api/v1/targets
     | jq '.data.activeTargets[].health'

Q58. Is there a structured JSON log format with
     request_id, agent_did, decision, and duration_ms
     on every enforcement decision?
     Evidence required: log output from a test request.
     Command: curl -H "Authorization: Bearer TEST"
     http://localhost:8000/test 2>/dev/null &&
     docker logs agentrepengine-scoring-service-1 2>&1 | tail -5

Q59. Is there an automated session-start health check
     script that runs all critical checks in under 15 seconds?
     Evidence required: scripts/gstate.sh or equivalent.
     Command: bash scripts/gstate.sh

Q60. What is the service startup time from cold
     docker compose up to all services healthy?
     Evidence required: time docker compose up -d output
     — measured, not estimated.

Q61. Is there a prerequisites checklist document for
     enterprise install — Kong version, K8s version,
     PostgreSQL version, Redis version, permissions?
     Evidence required: docs/ — prerequisites or
     install guide document.

Q62. Is memory usage per agent under 1KB in Redis
     — has this been measured at scale?
     Evidence required: Redis memory usage calculation
     or benchmark with 1000 agents.
     Command: docker exec redis redis-cli -u
     redis://are_admin:are_redis_dev@localhost:6379
     info memory | grep used_memory_human

Q63. Is there a dead letter queue for failed events
     — what happens to an event that fails processing
     3 times?
     Evidence required: agent_event_dlq table in schema
     + DLQ logic in event consumer.

═══════════════════════════════════════════════════════════════════
SECTION 9 — ADVERSARIAL RESILIENCE (Q64–Q69)
Evaluator: Red Team Lead
Failure mode: A product that cannot withstand its own
              threat model is not a security product.
═══════════════════════════════════════════════════════════════════

Q64. Can an agent build a TRUSTED score (800+) over 30 days
     and then execute a HIGH_RISK action to bypass enforcement?
     Evidence required: HIGH_RISK operation class in policy
     packs + VERIFY state regardless of score logic.

Q65. Can an attacker reset their score by creating a new
     agent identity — what is the starting score for a
     newly registered agent vs an orphaned agent?
     Evidence required: InitialScore=700 with probation
     vs OrphanScore=500 — probation enforcement.

Q66. Can an agent determine its own score by probing
     the API — is the score lookup rate-limited and
     is the threshold hidden from agents?
     Evidence required: score lookup endpoint access
     controls + rate limiting.

Q67. Has slow-walk evasion been tested — an attacker
     distributing malicious behavior over 7+ days
     staying below per-session velocity thresholds?
     Evidence required: multi-day evasion test scenarios
     in attack corpus or red team results.

Q68. What is the attack corpus coverage — how many
     distinct OWASP LLM Top 10 attack patterns are
     in the evaluation harness?
     Evidence required: tests/eval_harness/ or
     tests/attack_corpus/ — scenario count by OWASP category.

Q69. Has the product been tested against its own
     infrastructure — can an attacker with network
     access to port 6379 compromise the scoring state?
     Evidence required: Redis unauthenticated access
     rejection + ACL enforcement verification.

═══════════════════════════════════════════════════════════════════
SECTION 10 — ENTERPRISE SALES READINESS (Q70–Q72)
Evaluator: B2B Security GTM Specialist
Failure mode: A technically perfect product that cannot
              be sold is worthless. These 3 questions are
              what Lloyd's contact asks in the first meeting.
═══════════════════════════════════════════════════════════════════

Q70. Can you demonstrate a live blocked incident with
     full reason object in under 60 seconds from a
     cold stack in a client's meeting room?
     Evidence required: demo.sh timed run on cold stack.
     This must be verified on a clean machine, not your
     dev environment.

Q71. What are the three CISO documents and do they
     exist in the repository — operational safety
     architecture, pilot letter, maturity statement?
     Evidence required: docs/enterprise/ directory listing.
     Command: ls -la ~/AgentRepEngine/docs/enterprise/

Q72. What is the install time on a clean machine with
     Docker and Kong available — has this been measured
     or only claimed?
     Evidence required: documented install procedure +
     timed run on a machine where the repo has never
     been cloned.

═══════════════════════════════════════════════════════════════════
SCORING RUBRIC
═══════════════════════════════════════════════════════════════════

Score each section:
  All [F] answers:              100
  Mostly [F], some [OBS]:        85
  Mix of [F] and [H]:            70
  Multiple [INF] or [ASS]:       50
  Any [GAP] in critical section: MAX 40 for that section

Critical sections (any [GAP] = deal-killer in enterprise POC):
  Section 1 — Identity:         Q1, Q2, Q3, Q5
  Section 2 — Gateway:          Q10, Q12, Q13, Q15
  Section 4 — FP Governance:    Q28, Q29
  Section 5 — Explainability:   Q34, Q35, Q36
  Section 6 — Audit Trail:      Q40, Q41, Q42

Final composite score:
  FAANG-GRADE:       ≥90  — ready for Google, Meta, Stripe review
  ENTERPRISE-READY:  ≥80  — ready for financial services POC
  APPROACHING:       70–79 — 1–2 week sprint closes the gap
  NEEDS WORK:        <70  — do not approach enterprise yet

═══════════════════════════════════════════════════════════════════
END OF 72-QUESTION AUDIT
AgentRepEngine FAANG & Enterprise Grade Evaluation
Version 1.0 | March 22, 2026
Paste into any session with full project context loaded.
Answer every question with evidence from code, not claims.
═══════════════════════════════════════════════════════════════════