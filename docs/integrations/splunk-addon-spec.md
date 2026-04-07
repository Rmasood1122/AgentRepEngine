# AgentRepEngine — Splunk Add-on Specification
# Version: 1.0 — April 7, 2026

## OVERVIEW
ARE sends enforcement decisions to Splunk via HTTP Event Collector (HEC).
Every BLOCKED or FLAGGED decision fires a structured event in real time.

Integration type: Push (ARE to Splunk HEC)
Protocol: HTTPS POST / Format: Splunk HEC JSON
Trigger: Every BLOCKED or HIGH_RISK enforcement decision

## CONFIGURATION
Step 1 — Create HEC Token: Settings > Data Inputs > HTTP Event Collector > New Token
  Token name: agentrepengine / Source type: _json / Index: security

Step 2 — Configure ARE in docker-compose.yml:
  SIEM_WEBHOOK_URL: https://your-splunk:8088/services/collector/event
  SIEM_TOKEN: your-hec-token-here / SIEM_ENABLED: "true"

Step 3 — Verify: curl -k https://your-splunk:8088/services/collector/event
  -H "Authorization: Splunk your-token" -H "Content-Type: application/json"
  -d '{"event":{"test":"are_connectivity_check"}}'
  Expected: {"text":"Success","code":0}

## EVENT SCHEMA
{ "sourcetype":"are:decision", "event": { "agent_did":"did:jwt:agent-001",
  "decision":"BLOCKED", "score":187, "confidence_pct":94,
  "policy_fired":"pii_field_access_rate", "worst_z_score":4.2,
  "enforcement_mode":"enforce", "hash":"sha256:abc123..." } }

## SEARCH QUERIES
All blocked: index=security sourcetype="are:decision" decision=BLOCKED | table _time agent_did score confidence_pct policy_fired | sort -_time
High confidence: index=security sourcetype="are:decision" decision=BLOCKED confidence_pct>90 | stats count by agent_did policy_fired
Slow-walk: index=security sourcetype="are:decision" policy_fired=variance_growth_rate_exceeded | table _time agent_did score worst_z_score
FP review: index=security sourcetype="are:decision" band=BLOCKED OR band=RESTRICTED | eval review_needed=if(confidence_pct<70,"YES","NO") | table _time agent_did score confidence_pct policy_fired review_needed

## DASHBOARD PANELS
Panel 1 Decisions today: sourcetype="are:decision" | timechart count by decision
Panel 2 Top blocked agents: decision=BLOCKED | stats count by agent_did | head 10
Panel 3 FP rate trend: | eval fp=if(band="BLOCKED" AND confidence_pct<50,1,0) | timechart avg(fp)
Panel 4 Policy violations: | stats count by policy_fired | sort -count
Panel 5 Slow-walk alerts: policy_fired=variance_growth_rate_exceeded

## ALERTS
Alert 1 High-confidence block: decision=BLOCKED confidence_pct>90 earliest=-5m -> Email security team
Alert 2 Auto-rollback: event_type=auto_rollback earliest=-10m -> PagerDuty
Alert 3 Hash chain breach: hash_chain_valid=false earliest=-5m -> PagerDuty CRITICAL

## COMPLIANCE
DORA Article 17 / SOC 2 CC7.2 / NIST AI RMF MG-4.1
Splunk Add-on Spec v1.0 | AgentRepEngine | April 7, 2026
