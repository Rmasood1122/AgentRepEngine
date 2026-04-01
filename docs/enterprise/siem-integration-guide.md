# AgentRepEngine — SIEM Integration Guide
# Version: 1.0 | April 1, 2026
# Audience: Security Engineers, SOC Teams, Enterprise Architects

---

## OVERVIEW

AgentRepEngine produces a structured, hash-chained audit trail
for every agent enforcement decision. This guide documents how
to connect ARE's audit trail to your SIEM as a real-time event
feed — not as a log file, but as a structured security event
stream.

ARE's enforcement decisions are natively formatted as SIEM
events. No post-processing required. No custom parser needed.

---

## EVENT FORMAT

Every ARE enforcement decision produces one structured event:
```json
{
  "event_type": "agent_enforcement_decision",
  "timestamp": "2026-04-01T14:32:00Z",
  "agent_did": "did:jwt:finserv:trading-agent:001",
  "org_id": "org-finserv-001",
  "decision": "BLOCKED",
  "score": 187,
  "score_delta": -556,
  "confidence_pct": 94,
  "policy_fired": "bulk_pii_access_prevention_v1",
  "policy_threshold": "PIIFieldAccessRate > 500/hour",
  "trigger_events": [
    {
      "event_type": "pii_field_access_rate",
      "count": 847,
      "window_minutes": 90,
      "baseline_per_hour": 12,
      "deviation_sigma": 4.2
    }
  ],
  "recommended_action": "Isolate agent and review last 24h of activity",
  "prev_hash": "sha256:a1b2c3...",
  "this_hash": "sha256:d4e5f6...",
  "override": false
}
```

Every field is populated on every event. No nulls. No missing
fields. The hash chain (`prev_hash` → `this_hash`) provides
tamper-evidence — any modification breaks the chain.

---

## SIEM INTEGRATION OPTIONS

### Option 1 — Webhook (Recommended)
ARE posts enforcement events to your SIEM webhook endpoint
in real time. Zero polling. Zero lag.

**Configuration** (`config/policy/global.yaml`):
```yaml
siem:
  enabled: true
  webhook_url: "https://your-siem.company.com/api/events"
  webhook_secret: "${SIEM_WEBHOOK_SECRET}"
  events:
    - BLOCKED
    - AUDIT
    - THROTTLE
  # Set to [] to forward all decisions including ALLOW
```

**Supported SIEM webhook formats:**
- Splunk HTTP Event Collector (HEC)
- Microsoft Sentinel Data Connector
- IBM QRadar Universal Cloud REST API
- Elastic Security webhook
- Generic JSON webhook (any SIEM with HTTP ingestion)

### Option 2 — PostgreSQL Direct Query
Your SIEM queries ARE's `enforcement_decisions` table directly.
```sql
-- Pull all BLOCKED decisions in last 24 hours
SELECT
  agent_did,
  decision,
  score,
  confidence_pct,
  policy_fired,
  reason_object,
  this_hash,
  created_at
FROM enforcement_decisions
WHERE decision = 'BLOCKED'
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;
```

### Option 3 — SOC2 Export (Audit Mode)
On-demand export of complete audit trail for regulatory exams:
```bash
curl -X GET http://localhost:8080/audit/export \
  -H "X-API-Key: ${ARE_API_KEY}" \
  -o audit_export_$(date +%Y%m%d).json
```

Output: complete hash-chained log, all decisions, all scores,
all contributing features. Format: JSON array, one event per line.

---

## SPLUNK INTEGRATION

**Step 1 — Create HEC token in Splunk**
Settings → Data Inputs → HTTP Event Collector → New Token
Source type: `_json` | Index: `ai_agents` (create if needed)

**Step 2 — Configure ARE webhook**
```yaml
siem:
  enabled: true
  webhook_url: "https://splunk.company.com:8088/services/collector"
  webhook_secret: "Splunk YOUR_HEC_TOKEN"
  events: [BLOCKED, AUDIT, THROTTLE]
```

**Step 3 — Splunk search query**
```
index=ai_agents source="agentrepengine"
| where decision="BLOCKED"
| stats count by agent_did, policy_fired
| sort -count
```

**Step 4 — Alert rule**
```
index=ai_agents decision=BLOCKED confidence_pct>80
| alert action=email to="soc@company.com"
```

---

## MICROSOFT SENTINEL INTEGRATION

**Step 1 — Create Data Collection Rule in Sentinel**
Azure Portal → Microsoft Sentinel → Data Connectors →
Custom Logs via AMA → Create DCR

**Step 2 — Configure ARE webhook**
```yaml
siem:
  enabled: true
  webhook_url: "https://YOUR_DCE.ingest.monitor.azure.com/dataCollectionRules/YOUR_DCR_ID/streams/Custom-ARE_Events_CL"
  webhook_secret: "Bearer ${AZURE_TOKEN}"
  events: [BLOCKED, AUDIT, THROTTLE]
```

**Step 3 — KQL detection rule**
```kql
ARE_Events_CL
| where decision_s == "BLOCKED"
| where confidence_pct_d > 80
| project TimeGenerated, agent_did_s, policy_fired_s,
          score_d, confidence_pct_d, recommended_action_s
| order by TimeGenerated desc
```

---

## FIELD MAPPING REFERENCE

| ARE Field | Splunk Field | Sentinel Field | CEF Field |
|-----------|-------------|----------------|-----------|
| agent_did | agent_did | agent_did_s | src |
| decision | decision | decision_s | act |
| score | score | score_d | cn1 |
| confidence_pct | confidence_pct | confidence_pct_d | cn2 |
| policy_fired | policy_fired | policy_fired_s | cs1 |
| deviation_sigma | deviation_sigma | deviation_sigma_d | cn3 |
| this_hash | this_hash | this_hash_s | cs2 |
| org_id | org_id | org_id_s | dvc |

---

## TAMPER EVIDENCE VERIFICATION

ARE's hash chain can be verified at any time:
```bash
# Verify chain integrity — any break indicates tampering
curl -X GET http://localhost:8080/audit/verify-chain \
  -H "X-API-Key: ${ARE_API_KEY}"

# Response:
{
  "chain_valid": true,
  "decisions_verified": 1247,
  "earliest_decision": "2026-03-01T00:00:00Z",
  "latest_decision": "2026-04-01T14:32:00Z",
  "broken_links": 0
}
```

Present this output to regulators. The chain either holds or it
doesn't. No attestation required. Auditors verify it themselves.

---

## REGULATORY MAPPING

| SIEM Capability | DORA Requirement | SEC Requirement | HIPAA Requirement |
|----------------|-----------------|-----------------|-------------------|
| Real-time enforcement events | Art. 17 incident detection | Risk management documentation | § 164.312 audit controls |
| Hash-chained audit trail | Art. 12 ICT incident log | Books and records | § 164.312 integrity |
| Tamper-evidence verification | Art. 12 log integrity | Audit trail authenticity | § 164.312 transmission security |
| On-demand SOC2 export | Art. 30 reporting | Examiner production | § 164.308 evaluation |

---

## DEPLOYMENT CHECKLIST

- [ ] SIEM webhook endpoint confirmed reachable from ARE host
- [ ] Webhook secret configured in `config/policy/global.yaml`
- [ ] Test event sent and received: `curl -X POST http://localhost:8080/audit/test-webhook`
- [ ] SIEM index/stream created for ARE events
- [ ] Alert rule configured for BLOCKED decisions with confidence_pct > 80
- [ ] SOC team notified of new event source
- [ ] Hash chain verification scheduled weekly

---

## SUPPORT

Engineering support available during pilot deployment.
Contact: rehanrana@call2leads.com
GitHub: github.com/Rehanrana11/AgentRepEngine
DOI: 10.5281/zenodo.19169185

---

*AgentRepEngine v1.0 | Naseem A2A Research Lab*
*IP: Zenodo DOI 10.5281/zenodo.19169185*
