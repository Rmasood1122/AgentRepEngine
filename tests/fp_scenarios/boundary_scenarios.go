package fp_scenarios

// BoundaryScenarioSpec documents the behavioral specification for each
// boundary scenario. Written BEFORE running against the model.
// Spec commit: this file is committed before boundary_scenarios.go is run.
// Each scenario targets z-score 4.0–5.5 on at least one feature —
// the zone where score approaches 500 and FP risk is highest.
// All scenarios represent legitimate enterprise agent behaviors.
//
// Score boundary math (from fp_test.go computeScore):
//   Score = round(0.5*700 + 0.5*V)
//   V = max(0, 1000 - penalty*3)
//   penalty = min(100*(worstZ-3.0), 300)  [only when worstZ > 3.0]
//   Score=500 requires V=300, penalty=233, worstZ=5.33
//   Score=550 requires V=450, penalty=183, worstZ=4.83
//   Score=600 requires V=600, penalty=133, worstZ=4.33
//   Score=650 requires V=650, penalty=117, worstZ=4.17
//
// Baselines (from fp_test.go):
//   tool_call_rate_per_hour:       mean=80,  std=60
//   unique_endpoints_per_hour:     mean=25,  std=30
//   bulk_access_count_per_session: mean=400, std=400
//   pii_field_access_rate:         mean=20,  std=25
//
// z=4.0 values: tool=320, endpoints=145, bulk=2000, pii=120
// z=4.5 values: tool=350, endpoints=160, bulk=2200, pii=132
// z=5.0 values: tool=380, endpoints=175, bulk=2400, pii=145
// z=5.3 values: tool=398, endpoints=184, bulk=2520, pii=152

// Phase1BoundaryCorpus contains 50 legitimate agent scenarios operating
// near the score=500 enforcement boundary. These are the scenarios most
// likely to expose co-design bias — if 0.00% FP holds here, the claim
// is materially stronger than the base corpus alone.
var Phase1BoundaryCorpus = []LegitScenario{

	// ── BATCH JOB AGENTS (10 scenarios) ─────────────────────────────
	// Spec: Scheduled batch agents burst at scheduled times.
	// Their tool_call_rate spikes 4–5x normal for 60–90 minute windows.
	// Legitimate. Predictable. Often run at off-peak hours.

	{ID: "BND-001", Name: "Nightly ETL burst — peak 10min window",
		Vector: map[string]float64{"tool_call_rate_per_hour": 320, "bulk_access_count_per_session": 1800}, MinScore: 510},
	{ID: "BND-002", Name: "Monthly billing close — 90min burst",
		Vector: map[string]float64{"tool_call_rate_per_hour": 350, "bulk_access_count_per_session": 2000, "pii_field_access_rate": 35}, MinScore: 505},
	{ID: "BND-003", Name: "Quarterly audit extract — full scan",
		Vector: map[string]float64{"tool_call_rate_per_hour": 340, "bulk_access_count_per_session": 2200, "pii_field_access_rate": 55}, MinScore: 500},
	{ID: "BND-004", Name: "Year-end data archive — sustained burst",
		Vector: map[string]float64{"tool_call_rate_per_hour": 360, "bulk_access_count_per_session": 2400}, MinScore: 500},
	{ID: "BND-005", Name: "Weekly data warehouse full reload",
		Vector: map[string]float64{"tool_call_rate_per_hour": 330, "bulk_access_count_per_session": 1900}, MinScore: 505},
	{ID: "BND-006", Name: "Nightly ML feature pipeline — high throughput",
		Vector: map[string]float64{"tool_call_rate_per_hour": 345, "bulk_access_count_per_session": 2100}, MinScore: 500},
	{ID: "BND-007", Name: "Daily compliance snapshot — full table",
		Vector: map[string]float64{"tool_call_rate_per_hour": 310, "bulk_access_count_per_session": 1700, "pii_field_access_rate": 42}, MinScore: 510},
	{ID: "BND-008", Name: "Batch invoice processing — month-end",
		Vector: map[string]float64{"tool_call_rate_per_hour": 355, "bulk_access_count_per_session": 1600, "pii_field_access_rate": 60}, MinScore: 500},
	{ID: "BND-009", Name: "Batch KYC refresh — annual cycle",
		Vector: map[string]float64{"tool_call_rate_per_hour": 325, "bulk_access_count_per_session": 1500, "pii_field_access_rate": 75}, MinScore: 500},
	{ID: "BND-010", Name: "Batch transaction reconciliation — daily",
		Vector: map[string]float64{"tool_call_rate_per_hour": 338, "bulk_access_count_per_session": 2000}, MinScore: 505},

	// ── HIGH-ENDPOINT MONITORING AGENTS (10 scenarios) ───────────────
	// Spec: Watchdog and monitoring agents touch many endpoints by design.
	// unique_endpoints_per_hour reaches z=4.0–5.0 (145–175 endpoints/hr).
	// Legitimate: they're supposed to probe the entire service mesh.

	{ID: "BND-011", Name: "Service mesh health monitor — all nodes",
		Vector: map[string]float64{"unique_endpoints_per_hour": 145, "tool_call_rate_per_hour": 60}, MinScore: 530},
	{ID: "BND-012", Name: "SLA compliance monitor — all contracts",
		Vector: map[string]float64{"unique_endpoints_per_hour": 155, "tool_call_rate_per_hour": 55}, MinScore: 520},
	{ID: "BND-013", Name: "Security posture scanner — full inventory",
		Vector: map[string]float64{"unique_endpoints_per_hour": 165, "tool_call_rate_per_hour": 70}, MinScore: 510},
	{ID: "BND-014", Name: "API gateway latency monitor — all routes",
		Vector: map[string]float64{"unique_endpoints_per_hour": 150, "tool_call_rate_per_hour": 65}, MinScore: 525},
	{ID: "BND-015", Name: "Infrastructure drift detector — full scan",
		Vector: map[string]float64{"unique_endpoints_per_hour": 160, "tool_call_rate_per_hour": 58}, MinScore: 515},
	{ID: "BND-016", Name: "Certificate expiry monitor — all services",
		Vector: map[string]float64{"unique_endpoints_per_hour": 170, "tool_call_rate_per_hour": 45}, MinScore: 510},
	{ID: "BND-017", Name: "Database connection pool monitor",
		Vector: map[string]float64{"unique_endpoints_per_hour": 148, "tool_call_rate_per_hour": 72}, MinScore: 525},
	{ID: "BND-018", Name: "Microservice dependency mapper — weekly",
		Vector: map[string]float64{"unique_endpoints_per_hour": 175, "tool_call_rate_per_hour": 50}, MinScore: 505},
	{ID: "BND-019", Name: "Chaos engineering monitor — fault injection run",
		Vector: map[string]float64{"unique_endpoints_per_hour": 158, "tool_call_rate_per_hour": 80}, MinScore: 515},
	{ID: "BND-020", Name: "Observability agent — full telemetry scrape",
		Vector: map[string]float64{"unique_endpoints_per_hour": 162, "tool_call_rate_per_hour": 68}, MinScore: 510},

	// ── HIGH-PII AUTHORIZED AGENTS (10 scenarios) ────────────────────
	// Spec: Medical records, financial compliance, HR agents — their job
	// IS high pii_field_access_rate. Authorized. Documented. Audited.
	// pii_field_access_rate reaches z=4.0–5.3 (120–152 range).

	{ID: "BND-021", Name: "Medical records agent — patient chart review",
		Vector: map[string]float64{"pii_field_access_rate": 120, "tool_call_rate_per_hour": 30}, MinScore: 530},
	{ID: "BND-022", Name: "Clinical trial data agent — subject records",
		Vector: map[string]float64{"pii_field_access_rate": 130, "tool_call_rate_per_hour": 25}, MinScore: 520},
	{ID: "BND-023", Name: "Insurance underwriting agent — full profile",
		Vector: map[string]float64{"pii_field_access_rate": 125, "tool_call_rate_per_hour": 28, "bulk_access_count_per_session": 300}, MinScore: 520},
	{ID: "BND-024", Name: "GDPR subject access request agent — bulk",
		Vector: map[string]float64{"pii_field_access_rate": 140, "tool_call_rate_per_hour": 20}, MinScore: 510},
	{ID: "BND-025", Name: "HR compensation analysis agent",
		Vector: map[string]float64{"pii_field_access_rate": 135, "tool_call_rate_per_hour": 22, "bulk_access_count_per_session": 250}, MinScore: 515},
	{ID: "BND-026", Name: "AML transaction monitoring agent",
		Vector: map[string]float64{"pii_field_access_rate": 122, "tool_call_rate_per_hour": 45, "bulk_access_count_per_session": 800}, MinScore: 510},
	{ID: "BND-027", Name: "Fraud investigation agent — case review",
		Vector: map[string]float64{"pii_field_access_rate": 128, "tool_call_rate_per_hour": 35}, MinScore: 520},
	{ID: "BND-028", Name: "Benefits enrollment agent — open enrollment",
		Vector: map[string]float64{"pii_field_access_rate": 118, "tool_call_rate_per_hour": 32, "bulk_access_count_per_session": 400}, MinScore: 525},
	{ID: "BND-029", Name: "Credit scoring agent — portfolio review",
		Vector: map[string]float64{"pii_field_access_rate": 145, "tool_call_rate_per_hour": 18}, MinScore: 505},
	{ID: "BND-030", Name: "Patient discharge summary agent",
		Vector: map[string]float64{"pii_field_access_rate": 132, "tool_call_rate_per_hour": 26, "bulk_access_count_per_session": 200}, MinScore: 515},

	// ── COMBINED HIGH-LOAD AGENTS (10 scenarios) ─────────────────────
	// Spec: Agents with elevated rates on multiple features simultaneously.
	// No single feature at z>5.3, but two features both at z=3.5–4.5.
	// These test whether multi-feature elevation causes unexpected FPs.

	{ID: "BND-031", Name: "Data pipeline agent — high throughput multi-source",
		Vector: map[string]float64{"tool_call_rate_per_hour": 290, "unique_endpoints_per_hour": 130, "bulk_access_count_per_session": 1600}, MinScore: 500},
	{ID: "BND-032", Name: "Compliance reporting agent — cross-system audit",
		Vector: map[string]float64{"tool_call_rate_per_hour": 270, "unique_endpoints_per_hour": 120, "pii_field_access_rate": 80}, MinScore: 505},
	{ID: "BND-033", Name: "Financial consolidation agent — group reporting",
		Vector: map[string]float64{"tool_call_rate_per_hour": 300, "bulk_access_count_per_session": 1800, "pii_field_access_rate": 45}, MinScore: 500},
	{ID: "BND-034", Name: "Risk aggregation agent — portfolio-wide",
		Vector: map[string]float64{"tool_call_rate_per_hour": 280, "unique_endpoints_per_hour": 110, "bulk_access_count_per_session": 1500}, MinScore: 505},
	{ID: "BND-035", Name: "Regulatory submission agent — full data pull",
		Vector: map[string]float64{"tool_call_rate_per_hour": 260, "bulk_access_count_per_session": 1700, "pii_field_access_rate": 65}, MinScore: 505},
	{ID: "BND-036", Name: "Trade settlement agent — end-of-day batch",
		Vector: map[string]float64{"tool_call_rate_per_hour": 315, "bulk_access_count_per_session": 1400, "pii_field_access_rate": 30}, MinScore: 510},
	{ID: "BND-037", Name: "Payroll processing agent — monthly run",
		Vector: map[string]float64{"tool_call_rate_per_hour": 285, "bulk_access_count_per_session": 1200, "pii_field_access_rate": 90}, MinScore: 500},
	{ID: "BND-038", Name: "Contract lifecycle agent — bulk renewal",
		Vector: map[string]float64{"tool_call_rate_per_hour": 265, "unique_endpoints_per_hour": 105, "bulk_access_count_per_session": 1300}, MinScore: 505},
	{ID: "BND-039", Name: "Customer onboarding agent — high volume day",
		Vector: map[string]float64{"tool_call_rate_per_hour": 295, "pii_field_access_rate": 100, "bulk_access_count_per_session": 600}, MinScore: 500},
	{ID: "BND-040", Name: "Document processing agent — OCR pipeline burst",
		Vector: map[string]float64{"tool_call_rate_per_hour": 305, "bulk_access_count_per_session": 1100, "unique_endpoints_per_hour": 95}, MinScore: 505},

	// ── REACTIVATION AGENTS (10 scenarios) ───────────────────────────
	// Spec: Agents dormant 14–30+ days. H has decayed (7-day half-life).
	// When reactivated, H is low and V must compensate. Score near 500.
	// These are the hardest FP scenarios — low H + normal V = near-boundary.

	{ID: "BND-041", Name: "Reactivated agent — 14 days dormant, normal load",
		Vector: map[string]float64{"tool_call_rate_per_hour": 80, "unique_endpoints_per_hour": 25, "bulk_access_count_per_session": 400}, MinScore: 500},
	{ID: "BND-042", Name: "Reactivated agent — 21 days dormant, normal load",
		Vector: map[string]float64{"tool_call_rate_per_hour": 70, "unique_endpoints_per_hour": 20, "bulk_access_count_per_session": 350}, MinScore: 500},
	{ID: "BND-043", Name: "Reactivated agent — 30 days dormant, light load",
		Vector: map[string]float64{"tool_call_rate_per_hour": 50, "unique_endpoints_per_hour": 15, "bulk_access_count_per_session": 200}, MinScore: 500},
	{ID: "BND-044", Name: "Reactivated agent — seasonal, quarterly return",
		Vector: map[string]float64{"tool_call_rate_per_hour": 90, "bulk_access_count_per_session": 500}, MinScore: 500},
	{ID: "BND-045", Name: "Reactivated agent — post-maintenance window",
		Vector: map[string]float64{"tool_call_rate_per_hour": 75, "unique_endpoints_per_hour": 22}, MinScore: 500},
	{ID: "BND-046", Name: "Reactivated agent — after system migration",
		Vector: map[string]float64{"tool_call_rate_per_hour": 85, "unique_endpoints_per_hour": 28, "bulk_access_count_per_session": 450}, MinScore: 500},
	{ID: "BND-047", Name: "Reactivated agent — fiscal year restart",
		Vector: map[string]float64{"tool_call_rate_per_hour": 95, "bulk_access_count_per_session": 550, "pii_field_access_rate": 18}, MinScore: 500},
	{ID: "BND-048", Name: "Reactivated agent — post-audit freeze lifted",
		Vector: map[string]float64{"tool_call_rate_per_hour": 65, "unique_endpoints_per_hour": 18, "bulk_access_count_per_session": 300}, MinScore: 500},
	{ID: "BND-049", Name: "Reactivated agent — disaster recovery test",
		Vector: map[string]float64{"tool_call_rate_per_hour": 100, "unique_endpoints_per_hour": 30, "bulk_access_count_per_session": 600}, MinScore: 500},
	{ID: "BND-050", Name: "Reactivated agent — 45 days dormant, minimal load",
		Vector: map[string]float64{"tool_call_rate_per_hour": 40, "unique_endpoints_per_hour": 12, "bulk_access_count_per_session": 150}, MinScore: 500},
}
