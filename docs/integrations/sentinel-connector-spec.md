# AgentRepEngine — Microsoft Sentinel Connector Specification
# Version: 1.0 — April 7, 2026

## OVERVIEW
ARE sends enforcement decisions to Microsoft Sentinel via Log Analytics HTTP Data Collector API.
Events appear as custom log type ARE_Enforcement_CL.

Integration type: Push / Protocol: HTTPS POST / Format: CEF + JSON
Trigger: Every BLOCKED or HIGH_RISK enforcement decision

## CONFIGURATION
Step 1 — Get Workspace Credentials: Azure Portal > Log Analytics Workspaces > your-workspace
  > Agents > Copy: Workspace ID + Primary Key

Step 2 — Configure ARE in docker-compose.yml:
  SIEM_FORMAT: sentinel / SIEM_WORKSPACE_ID: your-workspace-id
  SIEM_WORKSPACE_KEY: your-primary-key / SIEM_ENABLED: "true"

Step 3 — Verify (events appear within 5 minutes): ARE_Enforcement_CL | take 10

## CEF EVENT FORMAT
CEF:0|AgentRepEngine|ARE|1.0|BLOCKED|Agent behavioral enforcement|7|
  src=did:jwt:agent-001 act=BLOCKED cs1=187 cs1Label=score
  cs2=94 cs2Label=confidence_pct cs3=pii_field_access_rate cs3Label=policy_fired
  cs4=4.2 cs4Label=worst_z_score cs5=sha256:abc123... cs5Label=chain_hash
  msg=Agent blocked: PII access rate 4.2 sigma above 30-day baseline

## JSON EVENT SCHEMA
{ "TimeGenerated":"2026-04-07T14:32:00Z", "AgentDID_s":"did:jwt:agent-001",
  "Decision_s":"BLOCKED", "Score_d":187, "ConfidencePct_d":94,
  "PolicyFired_s":"pii_field_access_rate", "WorstZScore_d":4.2,
  "ChainHash_s":"sha256:abc123...", "EnforcementMode_s":"enforce" }

## KQL QUERIES
All blocked (24h): ARE_Enforcement_CL | where TimeGenerated > ago(24h) | where Decision_s == "BLOCKED" | project TimeGenerated, AgentDID_s, Score_d, ConfidencePct_d, PolicyFired_s | sort by TimeGenerated desc
High confidence: ARE_Enforcement_CL | where Decision_s == "BLOCKED" and ConfidencePct_d > 90 | summarize count() by AgentDID_s, PolicyFired_s | sort by count_ desc
FP candidates: ARE_Enforcement_CL | where Decision_s in ("BLOCKED","RESTRICTED") | where ConfidencePct_d < 70 | extend ReviewRecommended = "YES" | project TimeGenerated, AgentDID_s, Score_d, ConfidencePct_d, PolicyFired_s, ReviewRecommended
Slow-walk: ARE_Enforcement_CL | where PolicyFired_s == "variance_growth_rate_exceeded" | project TimeGenerated, AgentDID_s, Score_d, WorstZScore_d | sort by TimeGenerated desc
Daily summary: ARE_Enforcement_CL | where TimeGenerated > ago(7d) | summarize count() by bin(TimeGenerated, 1d), Decision_s | render timechart

## ANALYTICS RULES
Rule 1 High-confidence block: Decision_s=="BLOCKED" and ConfidencePct_d>90 | Every 5min | Severity: Medium | Tactics: Exfiltration
Rule 2 Auto-rollback: Decision_s=="AUTO_ROLLBACK" | Every 10min | Severity: High | Tactics: DefenseEvasion
Rule 3 Audit trail breach: Decision_s=="HASH_CHAIN_BREACH" | Every 5min | Severity: Critical | Tactics: Tampering

## DEFENDER XDR INTEGRATION
ARE events appear in Defender XDR incidents when analytics rules fire.
Correlation: ARE blocked agent + Defender endpoint alert on same device
= combined signal of compromised identity + anomalous agent behavior.

## COMPLIANCE
DORA Article 17 / SOC 2 CC7.2 / NIST AI RMF MG-4.1 / Microsoft Security baseline
Sentinel Connector Spec v1.0 | AgentRepEngine | April 7, 2026
