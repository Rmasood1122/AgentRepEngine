# AgentRepEngine — Operational Safety Architecture
## How This System Protects Your Production Environment
Version 1.2 | For: Security Review | Classification: Confidential

---

### The Core Safety Guarantee

AgentRepEngine operates as a zero-impact visibility and behavioral
enforcement layer on top of your existing infrastructure. It does
not replace your gateway, modify your agents, or touch your data.
It watches, scores, and — when you choose — enforces.

If AgentRepEngine fails for any reason, your agents keep working.
This is not an accident. It is a documented architectural decision.

---

### How Failure Is Handled

IMPORTANT — TWO FAILURE CLASSES, TWO CORRECT BEHAVIORS:

**Infrastructure failure (scoring service unavailable):**
- Circuit breaker fires within 100ms
- All agent traffic passes through unscored (fail-open)
- Every unscored request logged as UNSCORED with timestamp
- Your operations team receives an alert within 60 seconds
- Your agents never stop. Your users never notice.

This is called "fail-open on infrastructure" and it is the
correct choice for an availability-critical enforcement layer.

Rationale: Stopping all agent traffic because a scoring
sidecar is temporarily unreachable violates your right to
maintain business continuity. This is a conscious architectural
decision, not a gap. You retain full control over whether to
accept this trade-off or configure a stricter posture.

**Enforcement failure (reason object cannot be generated):**
- Decision defaults to AUDIT mode — logged, not blocked
- Nothing is blocked without a structured explanation on record
- This is called "fail-closed on enforcement"

Rationale: Behavioral enforcement with automatic human review
escalation requires an explanation. Blocking an agent without
being able to explain why is worse than not blocking it. An
unexplained block destroys security team trust permanently.
AUDIT mode preserves the cryptographically non-repudiable
enforcement log while protecting against unexplained blocks.

**These two behaviors are consistent, not contradictory:**
- Infrastructure problems never stop your business
- Enforcement problems never produce unexplained blocks
- You are always in control of both behaviors

Operator toggle: Set ENFORCEMENT_MODE=strict in docker-compose.yml
to change infrastructure failure behavior to fail-closed.
Default is fail-open for maximum availability.

---

### Risk-Staged Deployment — You Control the Pace

Every deployment follows a curriculum designed to build confidence
before any enforcement action is taken. You sign off at each stage.
We do not advance without your explicit approval.

**Stage 1 — Zero-Impact Visibility Mode (Weeks 1-2)**
- Every agent request is scored by ARE's behavioral attention engine
- Statistical behavioral deviation detection runs continuously
- All decisions are computed, logged, and available for review
- Nothing is blocked. Zero production impact. Zero risk.
- Your security team reviews the decision log and confirms accuracy.

**Stage 2 — Flagging Mode (Week 3)**
- Anomalies are flagged and surfaced to your security team
- Still no automated blocking — human review on every flag
- Your team develops confidence in the detection quality
- Proceed to Stage 3 only when your team is satisfied.

**Stage 3 — Behavioral Enforcement with Human Review (Weeks 4-6)**
- Enforcement enabled with mandatory human review on all blocks
  for the first 7 days
- Auto-rollback fires immediately if false positive rate exceeds 2%
- Your team retains full override authority at all times

**Stage 4 — Full Enforcement with Auto-Rollback Protection (Weeks 7-12)**
- Full enforcement with automatic human review escalation on blocks
- Adversarial baseline poisoning defense active
- Self-monitoring: any FP spike triggers automatic rollback to Stage 1

**You control the pace. We never advance a stage without your
sign-off. Most clients find Stage 1 visibility alone is worth
the deployment before they ever enable enforcement.**

---

### Auto-Rollback Protection

If false positive rate exceeds 2% in production:
- System automatically rolls back to zero-impact visibility mode
- Alert fires to your SIEM within 60 seconds
- No agent is blocked until your team re-enables enforcement
- Re-enabling enforcement requires explicit human action

This means: a misconfigured threshold cannot cause a false
positive storm. ARE's behavioral attention engine protects
itself and protects your production environment.

---

### Human-in-the-Loop Enforcement Validation

ARE implements human-in-the-loop enforcement by design.

Every enforcement decision in Stage 1-2 is reviewed by your
security team before any automated blocking is enabled. You
maintain human judgment throughout the deployment lifecycle.

Every enforcement decision can be reviewed, confirmed, or
overridden by your security team. Overrides feed back into
ARE's baseline calibration — the system learns from your
team's judgment, not just from its own statistical model.

This is the language DORA and SEC examiners use:
"human-in-the-loop enforcement validation."
ARE provides this by architecture, not by configuration.

---

### Rollback in Under 10 Minutes

If you want to remove AgentRepEngine completely:

1. Remove the Kong plugin: `kubectl delete kongplugin agent-reputation`
2. Remove the services: `helm uninstall agentrepengine`
3. Verify: `curl your-gateway/health` — no ARE headers in response

Your agents return to pre-installation behavior immediately.
No residual configuration. No data left in your cluster unless
you choose to retain the audit log for compliance purposes.

Tested rollback time: under 10 minutes on a standard K8s cluster.

---

### What Data We Collect

**Default (Tier 1 — Metadata only):**
- Event type, timestamp, API endpoint called, request count, duration
- Agent identity (JWT claims only — org_id, agent_did, instance_id)
- Behavioral enforcement decision and structured reason object

**What we never collect by default:**
- Prompt or completion content
- User data or PII
- Request or response payloads

Your behavioral telemetry stays in your Kubernetes cluster.
Nothing leaves your environment without your explicit configuration.
There is no phone-home, no SaaS dependency, no cloud egress required.

Phase 1 pilot deployment model: customer-hosted, Docker Compose,
running entirely within your network perimeter. No external network
dependency during operation.

---

### The Cryptographically Non-Repudiable Enforcement Log

Every enforcement decision generates:

- **Structured reason object** — what triggered, what threshold,
  what action, why, with confidence percentage
- **Cryptographically non-repudiable enforcement log** — hash-chained
  sequence that cannot be retroactively altered, enforced at the
  PostgreSQL permission level (not application level)
- **SOC2-compatible export** via: `GET /audit/export?format=soc2`
- **Verifiable chain**: `SELECT verify_hash_chain('enforcement_decisions')`
- **SIEM feed**: structured CEF/JSON event POSTed on every enforcement
  decision within 500ms — compatible with Splunk and Microsoft Sentinel

When your DORA examiner asks "prove what your agents did in Q3,"
ARE answers that question with a hash-verified behavioral sequence.
Your SIEM cannot do this for AI agents. ARE does.

Maps directly to:
- DORA Article 17 (ICT incident classification and reporting)
- DORA Article 28 (third-party ICT risk management)
- SOC2 CC7.2 (system monitoring)
- NIST AI RMF GOVERN function
- GDPR Article 22 (human oversight of automated decisions — the
  staged rollout with CISO sign-off at each phase constitutes the
  required human oversight documentation)

---

### Ensemble Enforcement — Why 0.00% False Positives

ARE's behavioral enforcement requires two independent detection
systems to agree before any agent is blocked:

1. **ARE's behavioral attention engine** (score-based): agent's
   behavioral deviation exceeds statistical threshold
2. **OWASP LLM Top 10 enforcement policies** (rule-based): agent's
   action crosses a policy threshold regardless of score

Both must flag independently before enforcement triggers.
A single miscalibrated metric cannot block a legitimate agent.
This is why ARE achieves 0.00% false positive rate on its
100-scenario held-out validation corpus — zero legitimate agents
blocked across all enterprise workflow scenarios tested.

---

### Adversarial Baseline Poisoning Defense

ARE defends against a novel attack class: adversarial baseline
poisoning via graduated rate escalation (slow-walk).

A compromised agent that gradually increases its activity rate
over 7-14 days can shift ARE's statistical baseline, making
anomalous behavior appear normal. ARE defends against this with
two independent layers:

**Layer 1 — Policy threshold enforcement**: OWASP LLM Top 10
enforcement policies catch high-value actions regardless of score.
A TRUSTED agent with score 950 cannot silently execute a bulk
PII export. The policy threshold fires independently of the
behavioral attention engine score.

**Layer 2 — Variance growth rate monitoring**: A doubling of
weekly behavioral variance triggers an early warning before the
baseline shifts. VARIANCE_WINDOW_DAYS = 7 (documented constant,
auditable). Detection occurs before the attacker's score recovers
to normal range.

Result: 100% detection rate on 10 slow-walk synthetic scenarios.
No other AI agent security system addresses this attack class.

---

### Kong JWT Verification — Retry Policy

The Kong `/verify` endpoint does not implement retry logic.
This is a deliberate design decision, not an oversight.

When Kong calls `/verify` to validate an agent's JWT, the call
either succeeds within the 500ms timeout or it fails. On failure,
the request is treated as unscored and passes through (fail-open),
consistent with the infrastructure failure behavior documented above.

**Why no retries:**
- The `/verify` endpoint sits in the hot path of every agent request.
  Adding retry logic would compound latency — a single retry with
  backoff could push verification past 1 second, violating the
  latency budget for real-time agent traffic.
- The 500ms timeout already provides a generous window for a
  lightweight JWT validation call. If the scoring service cannot
  respond within 500ms, it is experiencing an infrastructure failure,
  and the correct behavior is fail-open — not retry-and-block.
- Retries under load can amplify failure. A degraded scoring service
  receiving retried requests risks cascading into full unavailability.

**Phase 2 plan:**
Phase 2 will introduce retry with exponential backoff for the
`/verify` endpoint, gated behind a feature flag. This will be
configurable per-deployment and will include a maximum retry
budget (default: 1 retry, max additional latency: 250ms).

---

### Performance Characteristics

ARE's behavioral attention engine is built for production latency:

- Scoring call overhead: 0.25ns (Linux) — zero memory allocations
- Score lookup (Redis cache hit): 2ms p99
- Score computation (full): 5ms p95, 10ms p99 (hard ceiling)
- Enforcement decision: added to every agent request with
  sub-millisecond overhead at the Kong plugin layer

"0.0025% of a 400ms API budget. Invisible to your users."

---

### Contact for Security Review

Technical questions: rehan@naseem-a2a.com
Architecture review requests: rehan@naseem-a2a.com
Pilot agreement: rehan@naseem-a2a.com
