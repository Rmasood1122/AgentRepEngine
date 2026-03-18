package scoring

type FeatureVector struct {
	ToolCallRatePerHour       float64 `json:"tool_call_rate_per_hour"`
	UniqueEndpointsPerHour    float64 `json:"unique_endpoints_per_hour"`
	BulkAccessCountPerSession float64 `json:"bulk_access_count_per_session"`
	PIIFieldAccessRate        float64 `json:"pii_field_access_rate"`
	CrossTenantProbeCount     float64 `json:"cross_tenant_probe_count"`
	PermissionEscalationCount float64 `json:"permission_escalation_count"`
	SubAgentSpawnDepth        float64 `json:"sub_agent_spawn_depth"`
	TokenRefreshRate          float64 `json:"token_refresh_rate"`
}
