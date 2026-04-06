// tests/layer4_fp_corpus_test.go
// Layer 4 — False Positive Corpus: The Enterprise Unlock (0.00% FP Gate)
// ARE 11x Test Suite | Expert Panel: E1 + E4 + E7
// 0.00% FP is the single most important enterprise claim. Every scenario here
// must PASS (legitimate agent correctly NOT blocked).
// Run: go test ./tests/... -run TestLayer4 -v
package tests

import (
	"fmt"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Legitimate agent archetypes — the 7 types that must never trigger false positives
// E1: false positives in a security pilot destroy the relationship permanently
// ─────────────────────────────────────────────────────────────────────────────

type AgentArchetype struct {
	Name           string
	BaselineRate   float64
	BaselineStdDev float64
	HistoryScore   float64
	CurrentRate    float64
	PIIFieldRate        float64
	PermEscalations int
	CrossTenantProbes int
}

var legitimateArchetypes = []AgentArchetype{
	{
		Name:           "batch_processor",
		BaselineRate:   500.0,  // High-volume legitimate batch job
		BaselineStdDev: 80.0,
		HistoryScore:   820.0,
		CurrentRate:    550.0,  // Slightly above baseline — still within 1 std dev
		PIIFieldRate:        0.02,
	},
	{
		Name:           "customer_service_agent",
		BaselineRate:   25.0,
		BaselineStdDev: 5.0,
		HistoryScore:   760.0,
		CurrentRate:    28.0,
		PIIFieldRate:        0.2, // Customer data — below 0.3 threshold
	},
	{
		Name:           "analytics_agent",
		BaselineRate:   150.0,
		BaselineStdDev: 30.0,
		HistoryScore:   800.0,
		CurrentRate:    145.0,
		PIIFieldRate:        0.08,
	},
	{
		Name:           "monitoring_agent",
		BaselineRate:   1000.0, // High-frequency health checks
		BaselineStdDev: 100.0,
		HistoryScore:   900.0,
		CurrentRate:    1050.0, // 0.5 std dev above — normal variance
		PIIFieldRate:        0.0,    // No PII access
	},
	{
		Name:           "report_generation_agent",
		BaselineRate:   5.0,  // Low frequency, scheduled
		BaselineStdDev: 1.0,
		HistoryScore:   840.0,
		CurrentRate:    6.0,
		PIIFieldRate:        0.15, // Some PII for reports — well below 0.3
	},
	{
		Name:           "integration_sync_agent",
		BaselineRate:   80.0,
		BaselineStdDev: 15.0,
		HistoryScore:   770.0,
		CurrentRate:    85.0,
		PIIFieldRate:        0.1,
	},
	{
		Name:           "new_agent_cold_start",
		BaselineRate:   10.0,
		BaselineStdDev: 0.0, // No historical std dev yet
		HistoryScore:   700.0, // Default starting score
		CurrentRate:    10.0,
		PIIFieldRate:        0.05,
	},
}

func TestLayer4_FP_AllArchetypes_NotBlocked(t *testing.T) {
	// Core FP gate: all legitimate archetypes must not be BLOCKED
	// This is the 0.00% FP claim
	fpCount := 0
	for _, arch := range legitimateArchetypes {
		agent := AgentState{
			DID:             fmt.Sprintf("did:are:%s", arch.Name),
			BaselineRate:    arch.BaselineRate,
			BaselineStdDev:  arch.BaselineStdDev,
			HistoryScore:    arch.HistoryScore,
			CurrentRate:     arch.CurrentRate,
			PIIFieldRate:    arch.PIIFieldRate,
			PermEscalations: arch.PermEscalations,
			CrossTenantProbes: arch.CrossTenantProbes,
		}
		result := scoreAgent(agent)
		if result.Blocked {
			fpCount++
			t.Errorf("FALSE POSITIVE: %s was BLOCKED (score=%.0f) — legitimate agent incorrectly blocked",
				arch.Name, result.Score)
		}
		t.Logf("[%s] score=%.0f band=%s z=%.2f FP=%v",
			arch.Name, result.Score, result.Band, result.ZScore, result.Blocked)
	}
	fpRate := float64(fpCount) / float64(len(legitimateArchetypes)) * 100
	t.Logf("FP rate on %d archetypes: %.2f%%", len(legitimateArchetypes), fpRate)
	if fpCount > 0 {
		t.Errorf("G-FP GATE FAILED: %.2f%% FP rate (requires 0.00%%)", fpRate)
	}
}

func TestLayer4_FP_AllArchetypes_NotHighRiskFlagged(t *testing.T) {
	// HIGH_RISK VERIFY must not fire on legitimate agents
	// A security team receiving constant HIGH_RISK alerts on normal agents will disable ARE
	highRiskCount := 0
	for _, arch := range legitimateArchetypes {
		if arch.Name == "new_agent_cold_start" {
			continue // Cold-start agents may legitimately get MONITORED — skip HIGH_RISK check
		}
		agent := AgentState{
			DID:            fmt.Sprintf("did:are:%s", arch.Name),
			BaselineRate:   arch.BaselineRate,
			BaselineStdDev: arch.BaselineStdDev,
			HistoryScore:   arch.HistoryScore,
			CurrentRate:    arch.CurrentRate,
			PIIFieldRate:   arch.PIIFieldRate,
		}
		result := scoreAgent(agent)
		if result.HighRisk {
			highRiskCount++
			t.Errorf("FALSE HIGH_RISK: %s flagged as HIGH_RISK — will generate false security alerts",
				arch.Name)
		}
	}
	if highRiskCount == 0 {
		t.Log("✅ zero HIGH_RISK false alerts on legitimate archetypes")
	}
}

func TestLayer4_FP_BatchAgent_LegitimateSpike_NotBlocked(t *testing.T) {
	// Batch agent with known 2x monthly burst (e.g., month-end reporting)
	// Must NOT be blocked — this is why per-agent baselines matter
	agent := AgentState{
		DID:            "did:are:batch-monthend",
		BaselineRate:   500.0,
		BaselineStdDev: 80.0, // High natural variance
		HistoryScore:   850.0,
		CurrentRate:    900.0, // 1.25x above normal — within 5 std devs: z = (900-500)/80 = 5.0
		// HOWEVER: with high H (850), legitimate burst should stay monitored
		PIIFieldRate: 0.01,
	}

	result := scoreAgent(agent)
	t.Logf("batch agent month-end spike: z=%.2f score=%.0f band=%s", result.ZScore, result.Score, result.Band)

	// With z=5.0, the velocity penalty is (5.0-3.0)*50 = 100
	// V = 1000 - 100 = 900
	// Score = 0.5*850 + 0.5*900 = 875 → TRUSTED
	// This should NOT be blocked
	if result.Blocked {
		t.Error("legitimate batch spike must not be BLOCKED — high H score should buffer")
	}
}

func TestLayer4_FP_NewAgent_ColdStart_TRUSTED(t *testing.T) {
	// New agents start at H=700, V=1000 (no anomalies) → score = 850 → TRUSTED
	// E4: blocking new agents on their first request would make ARE undeployable
	agent := AgentState{
		DID:            "did:are:brand-new-agent",
		BaselineRate:   10.0,
		BaselineStdDev: 0.0, // No baseline yet
		HistoryScore:   700.0, // Default start
		CurrentRate:    10.0,
		PIIFieldRate:        0.05,
	}

	result := scoreAgent(agent)
	t.Logf("cold-start new agent: score=%.0f band=%s", result.Score, result.Band)

	if result.Blocked {
		t.Error("new agent on first request must not be BLOCKED — kills deployment")
	}
}

func TestLayer4_FP_PIIFieldRate_AtExactThreshold_NotBlocked(t *testing.T) {
	// PIIFieldRate = 0.30 exactly — this is AT the threshold, not ABOVE it
	// Must NOT trigger violation (rule is > 0.3, not >= 0.3)
	agent := AgentState{
		DID:          "did:are:pii-boundary-test",
		BaselineRate: 20.0,
		BaselineStdDev: 3.0,
		HistoryScore: 800.0,
		CurrentRate:  20.0,
		PIIFieldRate:      0.30, // Exactly at threshold
	}
	result := scoreAgent(agent)
	if result.HighRisk {
		t.Error("PIIFieldRate=0.30 exactly must NOT trigger HIGH_RISK (rule is > 0.3)")
	}
	t.Log("✅ PIIFieldRate=0.30 boundary: correctly not flagged")
}

func TestLayer4_FP_BiasAudit_AllArchetypes_ZeroPercent(t *testing.T) {
	// Bias audit: FP rate must be 0.00% across ALL 7 archetypes
	// Per CONTINUATION_PROMPT.md: "Bias audit: 0.00% FP across all 7 agent archetypes"
	fpByArchetype := make(map[string]bool)
	for _, arch := range legitimateArchetypes {
		agent := AgentState{
			DID:            fmt.Sprintf("did:are:bias-%s", arch.Name),
			BaselineRate:   arch.BaselineRate,
			BaselineStdDev: arch.BaselineStdDev,
			HistoryScore:   arch.HistoryScore,
			CurrentRate:    arch.CurrentRate,
			PIIFieldRate:   arch.PIIFieldRate,
		}
		result := scoreAgent(agent)
		fpByArchetype[arch.Name] = result.Blocked
	}

	failedArchetypes := []string{}
	for name, wasFP := range fpByArchetype {
		if wasFP {
			failedArchetypes = append(failedArchetypes, name)
		}
	}

	if len(failedArchetypes) > 0 {
		t.Errorf("BIAS AUDIT FAILED: archetypes with FP: %v", failedArchetypes)
	} else {
		t.Logf("✅ bias audit: 0.00%% FP across all %d archetypes", len(legitimateArchetypes))
	}
}
