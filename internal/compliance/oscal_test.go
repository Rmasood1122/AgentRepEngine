package compliance

import (
	"testing"
	"time"
)

func defaultInput() *AREEvidenceInput {
	now := time.Now().UTC()
	return &AREEvidenceInput{
		FPRate:           0.0,
		TPRate:           0.88,
		TotalDecisions:   1250,
		BlockedIncidents: 3,
		AuditTrailIntact: true,
		MerkleRootHash:   "abc123def456abc123def456abc123def456abc123def456abc123def456abcd",
		PQCEnabled:       true,
		TEEEnabled:       true,
		ZKProofEnabled:   true,
		RaftEnabled:      true,
		AuditPeriodStart: now.Add(-30 * 24 * time.Hour),
		AuditPeriodEnd:   now,
	}
}

func TestGenerateOSCALBundle(t *testing.T) {
	result, err := GenerateOSCALBundle(defaultInput())
	if err != nil {
		t.Fatalf("GenerateOSCALBundle error: %v", err)
	}
	if result == nil {
		t.Fatal("result must not be nil")
	}
	if result.UUID == "" {
		t.Fatal("UUID must not be empty")
	}
	if result.Metadata.Title == "" {
		t.Fatal("metadata title must not be empty")
	}
}

func TestGenerateOSCALBundleNilInput(t *testing.T) {
	_, err := GenerateOSCALBundle(nil)
	if err == nil {
		t.Fatal("nil input should return error")
	}
}

func TestControlCount(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if len(result.Controls) == 0 {
		t.Fatal("controls must not be empty")
	}
	// Must cover SOC2 + DORA
	hasDORA, hasSOC2 := false, false
	for _, c := range result.Controls {
		if c.Framework == "DORA" {
			hasDORA = true
		}
		if c.Framework == "SOC2" {
			hasSOC2 = true
		}
	}
	if !hasDORA {
		t.Fatal("must include DORA controls")
	}
	if !hasSOC2 {
		t.Fatal("must include SOC2 controls")
	}
}

func TestAllControlsSatisfied(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	for _, c := range result.Controls {
		if c.Status != "satisfied" {
			t.Errorf("control %s has status %s, expected satisfied", c.ControlID, c.Status)
		}
	}
}

func TestSummaryOverallSatisfied(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if result.Summary.OverallStatus != "satisfied" {
		t.Fatalf("expected satisfied, got %s", result.Summary.OverallStatus)
	}
	if result.Summary.Satisfied != result.Summary.TotalControls {
		t.Fatalf("all controls should be satisfied: %d/%d",
			result.Summary.Satisfied, result.Summary.TotalControls)
	}
}

func TestSummaryMetrics(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if result.Summary.TotalControls == 0 {
		t.Fatal("total controls must be > 0")
	}
	if result.Summary.RegulatoryNote == "" {
		t.Fatal("regulatory note must not be empty")
	}
}

func TestFindingsMatchControls(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if len(result.Findings) != len(result.Controls) {
		t.Fatalf("findings count %d != controls count %d",
			len(result.Findings), len(result.Controls))
	}
}

func TestAllFindingsPass(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	for _, f := range result.Findings {
		if f.Status != "pass" {
			t.Errorf("finding %s has status %s, expected pass", f.FindingID, f.Status)
		}
	}
}

func TestComponentsPresent(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if len(result.Components) == 0 {
		t.Fatal("components must not be empty")
	}
}

func TestEvidencePresent(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	for _, c := range result.Controls {
		if len(c.Evidence) == 0 {
			t.Errorf("control %s has no evidence", c.ControlID)
		}
	}
}

func TestExportOSCALJSON(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	out, err := ExportOSCALJSON(result)
	if err != nil {
		t.Fatalf("ExportOSCALJSON error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("JSON output must not be empty")
	}
}

func TestSystemComponent(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if result.System.Title != ARESystemName {
		t.Fatalf("expected system title %s, got %s", ARESystemName, result.System.Title)
	}
	if result.System.Status != "operational" {
		t.Fatal("system status must be operational")
	}
}

func TestMetadataOSCALVersion(t *testing.T) {
	result, _ := GenerateOSCALBundle(defaultInput())
	if result.Metadata.OSCALVersion != OSCALVersion {
		t.Fatalf("expected OSCAL version %s, got %s", OSCALVersion, result.Metadata.OSCALVersion)
	}
}
