# AgentRepEngine — Operational Safety Architecture
## How This System Protects Your Production Environment
Version 1.1 | For: Security Review | Classification: Confidential

---

### The Core Safety Guarantee

AgentRepEngine operates as an observation and enforcement layer
on top of your existing infrastructure. It does not replace your
gateway, modify your agents, or touch your data. It watches,
scores, and — when you choose — blocks.

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

Rationale: Blocking an agent without being able to explain
why is worse than not blocking it. An unexplained block
destroys security team trust permanently. AUDIT mode preserves
the audit trail while protecting against unexplained blocks.

**These two behaviors are consistent, not contradictory:**
- Infrastructure problems never stop your business
- Enforcement problems never produce unexplained blocks
- You are always in control of both behaviors

Operator toggle: Set ENFORCEMENT_MODE=strict in docker-compose.yml
to change infrastructure failure behavior to fail-closed.
Default is fail-open for maximum availability.

---

### The 48-Hour Observe Window

Every new deployment starts in OBSERVE mode.

In OBSERVE mode:
- Every agent request is scored
- Enforcement decisions are computed and logged
- Nothing is blocked
- Your team reviews the decision log before any blocking begins

You control when enforcement goes live. We recommend 48 hours
minimum. You can stay in observe mode as long as you need.
The switch from OBSERVE to ENFORCE requires your explicit action.
We do not flip it automatically. Ever.

---

### Auto-Rollback Protection

If false positive rate exceeds 2% in production:
- System automatically rolls back to OBSERVE mode
- Alert fires to your SIEM within 60 seconds
- No agent is blocked until your team re-enables enforcement
- Re-enabling enforcement requires explicit human action

This means: a misconfigured threshold cannot cause a false
positive storm. The system protects itself and protects you.

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
- Enforcement decision and reason object

**What we never collect by default:**
- Prompt or completion content
- User data or PII
- Request or response payloads

Your behavioral telemetry stays in your Kubernetes cluster.
Nothing leaves your environment without your explicit configuration.
There is no phone-home, no SaaS dependency, no cloud egress required.

---

### Audit Evidence This System Produces

Every enforcement decision generates:
- Structured reason object (what triggered, what threshold, what action)
- Tamper-evident hash chain (cryptographic proof of log integrity)
- SOC2-compatible export via: `GET /audit/export?format=soc2`
- Verifiable chain: `SELECT verify_hash_chain('enforcement_decisions')`
- SIEM webhook: POST on every BLOCKED decision within 500ms

Maps directly to:
- SOC2 CC7.2 (system monitoring)
- NIST AI RMF GOVERN function
- Your AI governance policy documentation

---

### Contact for Security Review

Technical questions: rehan@naseem-a2a.com
Architecture review requests: rehan@naseem-a2a.com
Pilot agreement: rehan@naseem-a2a.com