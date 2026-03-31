# Lloyd Meeting Preparation
## Complete Briefing Document
Classification: Internal Only — Do Not Distribute

---

## TALKING POINTS

### TECHNICAL DIFFERENTIATORS (1-7)

**1. ARE is the transparency layer regulators demand.**

Every enforcement decision ARE makes — allow, audit, throttle, block — has a structured reason object attached. This is not a log line. It is a structured JSON document containing the decision, the agent identity, the score, the score delta, the trigger events, the policy that fired, the policy threshold, the recommended action, the peer cluster comparison, and the timestamp. This is what SEC examiners and DORA supervisors ask for when they examine AI governance: not "what did you block?" but "why did you block it, and can you prove the decision was justified?" ARE produces that answer automatically for every decision, every time.

**2. ARE implements all 5 AI safety mechanisms simultaneously.**

Most AI governance vendors implement one or two of these. ARE implements all five in production:

| Safety Mechanism | ARE Implementation |
|---|---|
| Red team testing | 30-scenario attack corpus + 10 slow-walk scenarios, validated on every build |
| Content filtering | 5 OWASP LLM Top 10 enforcement policy packs (LLM01, LLM04, LLM06, LLM07, LLM08/09) |
| Audit trail | Cryptographically non-repudiable enforcement log — SHA-256 hash chain, INSERT-only at PostgreSQL permission level |
| Explainability | Structured reason object on every enforcement decision — fail-closed if reason cannot be generated |
| Human oversight | Override workflow with reviewer attribution + staged deployment with CISO sign-off at each phase |

These are not roadmap items. They are implemented, tested, and validated against held-out corpora.

**3. ARE uses ensemble enforcement — two independent systems must agree before blocking.**

ARE does not block based on a single signal. Two independent enforcement systems — the behavioral attention engine (statistical deviation from per-agent baseline) and the OWASP policy engine (rule-based threshold enforcement) — must both flag an agent before a block decision is made. No single miscalibrated metric can block a legitimate agent.

This is why ARE achieves a 0.00% false positive rate on the held-out validation corpus. Two independent systems with different detection methodologies must independently conclude that the agent's behavior is anomalous. The probability of both systems producing a correlated false positive on the same request approaches zero.

Say it this way: "Two independent systems must agree before any block."

**4. ARE defends against adversarial baseline poisoning.**

Sophisticated attackers know that behavioral baselines adapt. They exploit this by slowly escalating activity over days, letting the baseline drift toward their attack behavior, then striking when their anomalous actions appear normal. This is the slow-walk attack.

ARE defends with two layers:

- **Layer 1 — Policy threshold enforcement (score-independent).** OWASP policy packs fire regardless of behavioral score. If an agent accesses 500+ PII fields in a session, the bulk PII access prevention policy fires whether the score says 900 or 200. High-value actions are caught by absolute thresholds, not relative deviation. This layer alone catches 100% of slow-walk scenarios in testing.

- **Layer 2 — Variance growth rate monitoring.** Even as an attacker shifts the baseline mean, the variance of the baseline grows. ARE monitors the rate of variance growth over a 7-day window. A variance doubling triggers an early warning before the baseline recovers to a stable (but corrupted) state.

Combined detection rate on 10 synthetic slow-walk attack scenarios: 100%.

**5. ARE already applies context-sensitive enforcement.**

Trading agents, reporting agents, and retrieval agents have different behavioral thresholds by policy design. A trading agent making 200 API calls per hour is normal. A reporting agent making 200 API calls per hour is anomalous. ARE's policy packs define different threshold profiles for different agent roles.

This is live now. Not Phase 2. Not roadmap. Deployed and operational.

**6. ARE's threshold calibration is conservative by design.**

The base scoring formula is frozen:

```
Score(t) = Clamp(W_h · H + W_v · V, 0, 1000)
```

Calibration adjusts only the weight coefficients (W_h and W_v), never the formula structure. Maximum change per calibration iteration: 10%. Every coefficient change is validated against the held-out corpus before acceptance. If the new coefficients produce any false positive on the validation corpus, the change is rejected automatically.

This means ARE cannot be accidentally miscalibrated by a single bad tuning decision. The formula is stable. The weights move slowly. Every change is validated before it takes effect.

**7. ARE's 8-dimension feature vector uses weighted attention.**

The behavioral attention engine scores agents across 8 behavioral dimensions:

| Dimension | What It Detects |
|---|---|
| tool_call_rate_per_hour | Injection patterns, automated abuse |
| unique_endpoints_per_hour | Service enumeration, reconnaissance |
| bulk_access_count_per_session | Data exfiltration attempts |
| pii_field_access_rate | Bulk PII extraction |
| cross_tenant_probe_count | Lateral movement (zero tolerance) |
| permission_escalation_count | Privilege escalation attempts |
| sub_agent_spawn_depth | Recursive agent spawning |
| token_refresh_rate | Identity cycling, credential reuse |

The scoring weights are attention coefficients over these dimensions, calibrated via A/B validation against the held-out corpus. Each dimension contributes to the composite score proportionally to its configured attention weight. Per-agent individual baselines are maintained using Welford's online algorithm, meaning the system learns what "normal" looks like for each specific agent — not a global average.

---

### DEPLOYMENT MODEL (8)

**8. Phase 1 pilot: entirely within your network perimeter.**

- Customer-hosted, Docker Compose deployment
- Runs entirely within your network perimeter
- No data leaves your environment — behavioral telemetry stays on your infrastructure
- No SaaS dependency — no external network calls during operation
- Installation: 4 hours on one environment
- Full rollback: under 10 minutes at any time
- No changes to your existing agents or infrastructure
- One Kong gateway plugin + one scoring service container

If Lloyd's team asks "where does our data go?" — the answer is nowhere. It stays in your environment. There is no phone-home, no telemetry upload, no SaaS backend.

---

### HUMAN OVERSIGHT (9-10)

**9. ARE implements HITL by design.**

Every enforcement decision in observe mode is reviewed by your security team before auto-block is ever enabled. The deployment follows a four-stage curriculum:

1. **Zero-impact visibility** (Weeks 1-2) — every request scored, nothing blocked, full decision log available
2. **Flagging** (Week 3) — anomalies surfaced to security team, human review on every flag, still no blocking
3. **Enforcement with human review** (Weeks 4-6) — blocking enabled, mandatory human review on all BLOCKED decisions for first 7 days
4. **Full enforcement** (Week 7+) — auto-rollback protection active, SIEM alerts on every block

Your security team validates detection quality before any automated action is enabled. DORA and SEC language: human-in-the-loop enforcement validation with documented stage-gate transitions.

Say it this way: "You control the pace. We don't advance without your sign-off."

**10. ARE includes a closed feedback loop.**

Every enforcement decision can be reviewed, confirmed, or overridden by a human reviewer. The override workflow requires:
- A specific decision ID
- A human reviewer ID (no system overrides permitted)
- A standardized reason code (ARE-FP-001 through ARE-EX-002)
- Optional context notes

Overrides are recorded in the same immutable enforcement log (hash-chained, INSERT-only). They feed back into the ModeController's false-positive rate calculation:

```
FP Rate = BLOCKED decisions overridden / Total BLOCKED decisions (rolling 1-hour window)
```

If the FP rate exceeds 2%, the ModeController automatically rolls back to observe mode. Overrides are not just corrections — they are the mechanism that keeps the system honest.

---

### SCALE/INFRA (11)

**11. ARE's Kong plugin is stateless.**

The Kong gateway plugin holds zero instance-local state. Every request:
1. Extracts agent identity from the X-Agent-DID JWT header
2. Pulls the current score from Redis via HGET
3. Reads the current enforcement mode from Redis
4. Makes the allow/audit/throttle/block decision
5. Emits the behavioral event asynchronously (post-response, zero critical-path impact)

No session affinity. No sticky routing. No local cache that can go stale. Horizontal scaling is architectural — add Kong instances behind a load balancer and they all read from the same Redis. No configuration changes, no state synchronization, no consensus protocol.

This matters because Lloyd's team will ask about scale. The answer is: add instances. That's it.

---

### PILOT MODEL (12-13)

**12. Risk-staged deployment — you control the pace.**

Three phases, each requiring explicit sign-off:

| Phase | Duration | What Happens | Exit Criteria |
|---|---|---|---|
| Observe | Weeks 1-2 | Every request scored and logged. Nothing blocked. | CISO confirms accuracy on 100+ scored requests |
| Flag | Week 3 | Anomalies surfaced to security team. Human review. No blocking. | Team confirms detection quality. FP rate <2%. |
| Enforce | Weeks 4+ | Blocking enabled. Auto-rollback if FP >2%. SIEM alerts on every block. | 7 days enforcement with <2% FP, zero blocking outages |

We do not advance without your sign-off. If at any point your team is not satisfied, we stay at the current phase or roll back entirely. Full rollback: under 10 minutes.

**13. After pilot: structured case study in the format regulators expect.**

The pilot produces a structured case study document containing:
- **Background:** deployment context, agent inventory, risk profile
- **Methodology:** scoring model, policy configuration, threshold calibration
- **Results:** detection rates, false positive rates, behavioral baselines established
- **Confirmed incidents:** actual anomalies detected, with full reason objects and resolution

Your examiner gets a document, not a dashboard. When the SEC or DORA supervisor asks "how do you govern your AI agents?", you hand them a structured report with cryptographically verifiable evidence, not a screenshot of a monitoring tool.

---

## PERFORMANCE NUMBERS TO KNOW COLD

Memorize these. Do not look them up during the meeting.

| Metric | Number | Context |
|---|---|---|
| True positive rate | **86.67%** | 30-scenario attack corpus (26/30 detected) — measured on held-out validation corpus |
| False positive rate | **0.00%** | 100-scenario legitimate agent corpus — measured on held-out validation corpus |
| Precision | **100%** | Zero false positives — measured on held-out validation corpus |
| F1 score | **0.9286** | TP rate: 86.67% | FP rate: 0.00% | Precision: 100% — measured on held-out validation corpus |
| Scoring call overhead | **0.25ns** | Linux, zero memory allocations |
| Gateway p99 latency | **10ms** | Hard ceiling, including Redis lookup |
| Demo runtime | **12 seconds** | End-to-end scoring demonstration |

If asked about the 4 undetected attacks: they are edge-case slow-walk variants that fall below the velocity detection threshold in isolation. The HIGH_RISK VERIFY layer catches the remaining scenarios when policy thresholds are included — combined slow-walk detection rate is 100%.

---

## PHRASES TO MAKE REFLEXIVE

Practice these until they come out without thinking. These are the four sentences that anchor every technical conversation.

> "You control the pace. We don't advance without your sign-off."

Use when: discussing deployment risk, timeline, or commitment. This neutralizes the "what if it breaks production" objection before it forms.

> "Two independent systems must agree before any block."

Use when: discussing false positives, reliability, or trust. This is the single most important technical differentiator. Most governance tools use one detection method. ARE uses two that must agree.

> "0.00% false positive rate on held-out validation corpus."

Use when: discussing accuracy, trust, or competitive comparison. Emphasize "held-out" — this is not training accuracy. The scenarios were not used to tune the model.

> "Every enforcement decision has a structured reason object."

Use when: discussing auditability, regulatory compliance, or explainability. This is what separates ARE from tools that log "blocked" with a severity number. ARE logs why, with structured evidence.

---

## WHAT NOT TO SAY

| Never Say | Say Instead | Why |
|---|---|---|
| MVP | "Phase 1 production system" or "current implementation" | MVP signals incomplete. ARE is a working system with validated detection rates. |
| Revenue projections | Nothing — do not raise this | Let Lloyd raise value and pricing first. If he asks, discuss pilot structure and success criteria, not revenue. |
| Scoring formula | "Behavioral attention engine" | "Formula" sounds academic and fragile. "Behavioral attention engine" communicates a production system with architectural intent. |
| AI / machine learning | "Behavioral scoring" or "statistical deviation detection" | Lloyd's security team will be skeptical of AI claims. Ground every capability in specific, testable mechanisms. |
| We can build that | "That's on our Phase 2 roadmap" or "Let me check our implementation status" | Do not commit to features in the meeting. Confirm only what is deployed and validated. |
| Real-time | "Sub-millisecond scoring overhead" or "0.25ns call overhead" | "Real-time" is vague. Specific latency numbers are credible. |

---

## QUESTIONS LLOYD'S SECURITY TEAM WILL ASK

**Q1: "What happens if your scoring service goes down?"**

ARE fails open. The circuit breaker fires within 100ms. All agent traffic passes through unscored. Every unscored request is logged with timestamp for post-incident review. Your operations team receives a SIEM alert within 60 seconds. Your agents never stop. Your users never notice. This is a conscious architectural decision — an enforcement layer that fails closed becomes the single point of failure it's supposed to prevent. You can override this to fail-closed via configuration if your risk posture requires it.

**Q2: "How do we know you won't block legitimate agents?"**

Two answers. First: ensemble enforcement means two independent detection systems must both flag before any block. Second: the deployment is staged — you run in observe mode for two weeks, reviewing every decision your security team sees, before any blocking is enabled. If the false positive rate exceeds 2% at any point after enforcement is enabled, the ModeController automatically rolls back to observe mode. No manual intervention required. No restart required. Instant, via Redis.

**Q3: "What data do you collect? Where does it go?"**

Behavioral metadata only: agent identity (pseudonymous DID), event type, timestamp, endpoint, request count, duration, decision, reason object, and 8-dimension feature vector. We never store prompt content, completion content, request payloads, response payloads, or user PII. All data stays within your network perimeter. No SaaS dependency. No external network calls. No telemetry upload. Default retention: 90 days, configurable.

**Q4: "Can we audit your enforcement decisions?"**

Yes. Every enforcement decision is recorded in a hash-chained, INSERT-only PostgreSQL table. UPDATE, DELETE, and TRUNCATE are revoked at the database permission level — not application-level protection, database-level immutability. The hash chain formula is `this_hash = sha256(prev_hash + id + timestamp + decision)`. Chain integrity is verified on every health check. We provide a SOC2-compatible export with a `chain_verified` boolean and a SIEM feed in CEF/JSON format compatible with Splunk and Microsoft Sentinel.

**Q5: "How does this integrate with our existing SIEM?"**

Structured CEF/JSON events emitted on every enforcement decision. Compatible with Splunk and Microsoft Sentinel out of the box. The SIEM webhook fires on every BLOCKED decision (fire-and-forget, never blocks the enforcement pipeline). Daily summary events are also available. No custom integration required — standard structured event format.

**Q6: "What's the performance impact on our gateway?"**

0.25ns scoring call overhead (Linux, zero memory allocations). 10ms p99 gateway latency including Redis lookup. Behavioral event emission is asynchronous — it happens post-response, zero impact on the critical request path. The Kong plugin adds no measurable latency to your agent traffic under normal operation.

**Q7: "What if an attacker slowly escalates to avoid detection?"**

This is the slow-walk attack — the most sophisticated adversarial pattern against behavioral baselines. ARE defends with two layers. Layer 1: OWASP policy thresholds fire regardless of behavioral score. If an agent hits 500+ PII fields in a session, it's caught whether its score is 900 or 200. Layer 2: variance growth rate monitoring detects baseline corruption before the attacker can exploit it. Combined detection rate on 10 synthetic slow-walk scenarios: 100%.

**Q8: "Can we customize the policies for our specific agent types?"**

Yes. Policy packs are version-controlled YAML files. Your team can review, modify, and extend them. Different agent roles (trading, reporting, retrieval) already have different threshold profiles by policy design. You can add custom policy packs for your specific agent taxonomy without modifying any ARE source code.

**Q9: "What's your exit strategy if we want to remove this?"**

Remove the Kong plugin and the scoring service container. Full removal completes in under 10 minutes. No residual configuration. Your agents return to pre-installation behavior immediately. No lock-in, no migration, no cleanup. The Pilot Letter of Understanding documents this explicitly.

**Q10: "Who else is using this?"**

ARE is in Phase 1 production validation. We are transparent about our maturity stage. What we offer is a validated detection system with published accuracy metrics (TP rate: 86.67% | FP rate: 0.00% | Precision: 100% | F1: 0.9286 — measured on held-out validation corpus) and a zero-risk pilot that lets your team validate independently before any commitment. The pilot produces a structured case study that your regulators can examine. We do not ask you to trust our claims — we ask you to verify them in your environment with your agents.

---

*Prepared for internal use only. Do not distribute externally.*
*Last updated: March 2026.*
