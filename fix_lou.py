import re

with open('docs/enterprise/pilot-letter-of-understanding.md', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix 1: Remove ACV exposure from Cost section
old_cost = """### Cost
- Observe-mode pilot: no cost
- Enforce-mode deployment pricing: discussed at Day 30 review
- Target ACV: $50,000–$150,000 annually depending on scope"""

new_cost = """### Cost
- Observe-mode pilot: no cost
- Enforce-mode deployment pricing: discussed at Day 30 review
- Pricing is scoped to your environment at Day 30 based on
  agent count, enforcement scope, and compliance requirements

### Licensing
ARE uses a two-tier licensing model. Both tiers are included
in the pilot at no cost. The distinction matters for your
procurement and legal review.

**Open components (Apache 2.0 — no license key required):**
- EWMA behavioral scoring engine (`scorer/ewma.go`)
- Kong gateway plugin (`kong/plugins/agent-reputation/`)
- Evaluation harness and FP measurement tooling
- All source code available at: github.com/Rehanrana11/AgentRepEngine

These components are irrevocably open source. You may inspect,
fork, audit, or deploy them independently of any commercial
agreement with ARE.

**Commercial components (source-available, license key required
for enforcement mode beyond 30-day pilot):**
- Merkle hash-chain audit trail (`internal/audit/`)
- Auto-rollback enforcement controller (`internal/enforcement/`)
- OSCAL evidence bundle generator (`cmd/oscal-generate`)
- DORA Article 8(4) compliance CLI (`cmd/dora-verify`)

During the pilot, a time-limited license key is provided at
no cost. Commercial license terms are negotiated at Day 30
if your team elects to proceed to enforce mode.

**Source escrow:** Available on request for enterprise contracts.
The commercial components can be placed in escrow with a
mutually agreed third party as a condition of contract close.

**What this means for your security team:**
The enforcement logic that protects your production agents
is auditable by you — either directly via Apache 2.0 source
(open components) or via escrow (commercial components).
You never depend on a vendor black box for runtime enforcement."""

if old_cost in content:
    content = content.replace(old_cost, new_cost)
    print("SUCCESS: Cost section updated, licensing section added")
else:
    print("ERROR: Cost section not found — check exact whitespace in file")
    # Show what we're looking for context
    idx = content.find("### Cost")
    if idx >= 0:
        print("Found '### Cost' at position", idx)
        print("Content around it:")
        print(repr(content[idx:idx+300]))
    else:
        print("'### Cost' not found in file at all")

with open('docs/enterprise/pilot-letter-of-understanding.md', 'w', encoding='utf-8') as f:
    f.write(content)
