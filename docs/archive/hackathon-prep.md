# Hackathon Prep — AgentRepEngine
# Lightning AI + Validia | April 4, 2026
# Newlab Brooklyn | 9:30AM–6PM
# KEEP THIS OPEN DURING THE EVENT

---

## DEMO COMMAND
```bash
cd ~/AgentRepEngine
bash scripts/hackathon-demo.sh
```

Runtime: ~60 seconds streaming. All containers must be running first.

Check containers:
```bash
docker compose ps
```

If anything is down:
```bash
docker compose up -d
```

---

## 60-SECOND OPENER (say this before starting demo)

"AI agents are making decisions autonomously in production
right now. Most companies have no idea what their agents
are actually doing between API calls.

AgentRepEngine sits in your Kong gateway and scores every
agent action against that agent's own behavioral baseline —
in real time, before the action commits.

Not logs. Not alerts after the fact. Enforcement at the
point of action.

I'm going to show you two agents. One clean, one compromised.
Both have valid tokens. Watch what happens."

[Start demo]

---

## NARRATION TRACK (say this while demo streams)

[Phase 1 — JWT identity prints]
"Two agents. Same org. Agent A is legitimate.
Agent B is compromised. Both have valid JWT tokens —
you can't tell them apart from identity alone."

[Scoring starts]
"ARE scores every action against each agent's own
behavioral baseline — built from their history,
not from generic rules."

[Anomaly detected ~request 11]
"Agent B just crossed the threshold. PII access rate
23 standard deviations above its own baseline."

[BLOCKED appears]
"Blocked. Before it commits. Not after."

[Reason object prints]
"Every block comes with a structured reason object —
agent ID, score, confidence percentage, contributing
features, timestamp. Your compliance team reads this.
No data scientist required."

[Hash chain verifies]
"Hash-chained audit trail. Tamper-evident.
INSERT-only at the database level. Your auditors
verify this themselves — they don't trust our attestation."

[Clean agent shown untouched]
"Agent A — never touched. Zero false positives
on our 100-scenario validation corpus."

[Final screen]
"30-day observe-mode pilot. No enforcement until
you sign off. At day 30 you have the ROI case."

---

## ANSWERS TO QUESTIONS YOU WILL GET

**"How is this different from a WAF or API gateway security?"**
"A WAF checks requests against known signatures. ARE checks
behavior against that specific agent's own history. A WAF
would pass Agent B — the requests look valid. ARE blocks it
because the rate is 23 standard deviations above that agent's
baseline. It learns what normal looks like for each agent
and enforces against deviation."

**"Can I try it?"**
"Yes — repo is public: github.com/Rehanrana11/AgentRepEngine
Docker Compose up in under 10 minutes. Or I can walk you
through a 30-day observe-mode pilot in your environment —
no enforcement until you're ready."

**"What frameworks does it work with?"**
"Framework-agnostic. LangChain, LlamaIndex, custom agents.
If it goes through Kong, ARE sees it. Zero changes to your
agent code."

**"What's the false positive rate?"**
"0.00% on our 100-scenario internal validation corpus.
External validation on your production traffic is exactly
what the 30-day pilot produces."

**"How long to deploy?"**
"Under 4 hours. Docker Compose, Kong plugin, config.
Observe mode first — no enforcement until you sign off."

**"Who's using it?"**
"We're in pilot conversations with regulated finserv
and healthcare enterprises. The 30-day observe-mode
pilot is the entry point — no procurement gauntlet."

**"Is this compliant with DORA/GDPR?"**
"ARE runs in your infrastructure — data never leaves
your network. Every enforcement decision has a structured
reason object satisfying GDPR Article 22. Hash-chained
audit trail satisfies DORA Article 45. Regulators are
about to require AI audit trails — ARE is the implementation
of that requirement, already running."

**"What does it cost?"**
"Observe-mode pilot is free. Enforce-mode deployment
is $50K–$150K annually depending on scope. The ROI
case is quantifiable at day 30."

---

## METRICS — KNOW COLD
```
FP rate:             0.00% internal corpus
TP rate:             86.67%
F1:                  0.9286
Slow-walk detection: 100% (10/10 scenarios)
Demo runtime:        ~60 seconds streaming
Overhead:            0.25ns Linux
Hardening:           94/100
IP:                  Zenodo DOI 10.5281/zenodo.19169185
Repo:                github.com/Rehanrana11/AgentRepEngine
```

---

## ONE-SENTENCE PRODUCT DESCRIPTION

"ARE is a behavioral enforcement layer for AI agents —
it sits in your gateway, scores every agent action against
its own baseline, and blocks anomalous behavior before
it commits. Like a firewall, but for what AI agents do,
not what packets they carry."

---

## IF SOMEONE WANTS TO FOLLOW UP

LinkedIn: connect on the spot
Email: rehanrana@call2leads.com
Repo: github.com/Rehanrana11/AgentRepEngine
DOI: 10.5281/zenodo.19169185

Offer: "I can send you the pilot letter of understanding
today — 30-day observe mode, zero cost, zero enforcement
until you're ready."

---

## WHAT NOT TO SAY

- Never say "MVP"
- Never mention revenue projections first
- Never say "zero false positives" — say "0.00% on internal corpus"
- Never pitch investors unless they ask about funding directly
- Never mention Check Point by name

---

## MORNING OF APRIL 4 — CHECKLIST

- [ ] Docker Desktop running
- [ ] `docker compose ps` — all containers healthy
- [ ] `bash scripts/hackathon-demo.sh` — runs clean
- [ ] Laptop charged, charger in bag
- [ ] Repo URL memorized: github.com/Rehanrana11/AgentRepEngine
- [ ] Read opener aloud once before leaving
- [ ] CONTINUATION_PROMPT.md uploaded to Claude Project

---

*AgentRepEngine | Naseem A2A Research Lab*
*rehanrana@call2leads.com*