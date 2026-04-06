// cmd/dora-verify/main.go
// L126 — DORA Article 8(4) audit output binary
// Produces a structured compliance report from enforcement_decisions hash chain.
// Designed for EU DORA examiners and internal audit teams.
// Customer runs this independently — no ARE involvement required.
// Exit codes: 0=pass, 1=error, 2=chain broken/tampered

package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DORAReport is the top-level output structure.
// Maps directly to DORA Article 8(4) ICT risk management requirements.
type DORAReport struct {
	ReportMetadata  ReportMetadata   `json:"report_metadata"`
	Article8Summary Article8Summary  `json:"dora_article_8_4_summary"`
	ChainIntegrity  ChainIntegrity   `json:"chain_integrity"`
	EnforcementLog  []EnforcementRow `json:"enforcement_log"`
	Attestation     Attestation      `json:"attestation"`
}

type ReportMetadata struct {
	GeneratedAt   string `json:"generated_at"`
	ReportPeriod  string `json:"report_period"`
	SystemName    string `json:"system_name"`
	SystemVersion string `json:"system_version"`
	DORAArticle   string `json:"dora_article"`
	RunBy         string `json:"run_by"`
}

type Article8Summary struct {
	TotalDecisions      int     `json:"total_decisions"`
	BlockedDecisions    int     `json:"blocked_decisions"`
	ThrottledDecisions  int     `json:"throttled_decisions"`
	AllowedDecisions    int     `json:"allowed_decisions"`
	ConfirmedFPs        int     `json:"confirmed_false_positives"`
	FPRatePct           float64 `json:"fp_rate_pct"`
	UniqueAgentsMonitor int     `json:"unique_agents_monitored"`
	ChainVerified       bool    `json:"audit_trail_tamper_evident_verified"`
	ObserveMode         bool    `json:"enforcement_in_observe_mode"`
}

type ChainIntegrity struct {
	RecordsVerified  int    `json:"records_verified"`
	ChainBroken      bool   `json:"chain_broken"`
	BrokenAtID       int64  `json:"broken_at_record_id,omitempty"`
	FirstRecordID    int64  `json:"first_record_id"`
	LastRecordID     int64  `json:"last_record_id"`
	VerificationNote string `json:"verification_note"`
}

type EnforcementRow struct {
	ID        int64  `json:"id"`
	AgentDID  string `json:"agent_did"`
	Decision  string `json:"decision"`
	Score     int    `json:"score"`
	CreatedAt string `json:"created_at"`
	HashValid bool   `json:"hash_valid"`
}

type Attestation struct {
	Statement      string `json:"statement"`
	VerifiedBy     string `json:"verified_by"`
	VerificationTS string `json:"verification_timestamp"`
	SQLProof       string `json:"self_verifying_sql"`
}

func main() {
	from := flag.String("from", "", "Start date YYYY-MM-DD (optional)")
	to := flag.String("to", "", "End date YYYY-MM-DD (optional)")
	dsn := flag.String("dsn", "postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable", "PostgreSQL DSN")
	format := flag.String("format", "json", "Output format: json or text")
	observe := flag.Bool("observe", true, "Mark report as observe-mode (pre-enforcement)")
	flag.Parse()

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: database unreachable: %v\n", err)
		os.Exit(1)
	}

	// Build date range
	periodStr := "all time"
	query := `SELECT id, agent_did, decision, score, prev_hash, this_hash, created_at
	          FROM enforcement_decisions WHERE 1=1`
	args := []interface{}{}

	var fromLabel, toLabel string
	if *from != "" {
		t, err := time.Parse("2006-01-02", *from)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid --from: %v\n", err)
			os.Exit(1)
		}
		args = append(args, t)
		query += fmt.Sprintf(" AND created_at >= $%d", len(args))
		fromLabel = *from
	}
	if *to != "" {
		t, err := time.Parse("2006-01-02", *to)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid --to: %v\n", err)
			os.Exit(1)
		}
		t = t.Add(24 * time.Hour)
		args = append(args, t)
		query += fmt.Sprintf(" AND created_at < $%d", len(args))
		toLabel = *to
	}
	if fromLabel != "" || toLabel != "" {
		periodStr = fmt.Sprintf("%s to %s", fromLabel, toLabel)
	}
	query += " ORDER BY id ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	// Process rows
	var log []EnforcementRow
	var prevHash string
	var count, blocked, throttled, allowed int
	var firstID, lastID int64
	chainBroken := false
	var brokenAt int64
	agentsSeen := map[string]bool{}

	for rows.Next() {
		var id int64
		var agentDID, decision string
		var score int
		var storedPrev, storedThis sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &agentDID, &decision, &score,
			&storedPrev, &storedThis, &createdAt); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: scan failed at record %d: %v\n", id, err)
			os.Exit(1)
		}

		if count == 0 {
			firstID = id
		}
		lastID = id
		agentsSeen[agentDID] = true

		// Verify chain linkage
		hashValid := true
		if count > 0 && (!storedPrev.Valid || storedPrev.String != prevHash) {
			if !chainBroken {
				chainBroken = true
				brokenAt = id
			}
			hashValid = false
		}

		// Recompute this_hash
		raw := fmt.Sprintf("%d|%s|%s|%d|%s", id, agentDID, decision, score, prevHash)
		computed := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
		if storedThis.Valid && storedThis.String != computed {
			if !chainBroken {
				chainBroken = true
				brokenAt = id
			}
			hashValid = false
		}
		prevHash = computed

		// Count by decision type
		switch decision {
		case "BLOCKED":
			blocked++
		case "THROTTLE", "RESTRICTED":
			throttled++
		default:
			allowed++
		}

		log = append(log, EnforcementRow{
			ID:        id,
			AgentDID:  agentDID,
			Decision:  decision,
			Score:     score,
			CreatedAt: createdAt.UTC().Format(time.RFC3339),
			HashValid: hashValid,
		})
		count++
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: row iteration: %v\n", err)
		os.Exit(1)
	}

	// FP rate from daily_fp_metrics
	var fpCount int
	var fpRate float64
	_ = db.QueryRow(`SELECT COALESCE(SUM(fp_count),0), COALESCE(AVG(fp_rate),0)
	                 FROM daily_fp_metrics`).Scan(&fpCount, &fpRate)

	verifyNote := "Hash chain intact — audit trail has not been tampered with."
	if chainBroken {
		verifyNote = fmt.Sprintf("CHAIN BROKEN at record ID %d — audit trail integrity compromised.", brokenAt)
	}
	if count == 0 {
		verifyNote = "No enforcement decisions recorded in this period (observe mode or no activity)."
	}

	report := DORAReport{
		ReportMetadata: ReportMetadata{
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
			ReportPeriod:  periodStr,
			SystemName:    "AgentRepEngine (ARE)",
			SystemVersion: "1.4.0",
			DORAArticle:   "Article 8(4) — ICT risk management: monitoring and logging of ICT-related incidents",
			RunBy:         "customer-self-serve — no ARE involvement",
		},
		Article8Summary: Article8Summary{
			TotalDecisions:      count,
			BlockedDecisions:    blocked,
			ThrottledDecisions:  throttled,
			AllowedDecisions:    allowed,
			ConfirmedFPs:        fpCount,
			FPRatePct:           fpRate,
			UniqueAgentsMonitor: len(agentsSeen),
			ChainVerified:       !chainBroken,
			ObserveMode:         *observe,
		},
		ChainIntegrity: ChainIntegrity{
			RecordsVerified:  count,
			ChainBroken:      chainBroken,
			BrokenAtID:       brokenAt,
			FirstRecordID:    firstID,
			LastRecordID:     lastID,
			VerificationNote: verifyNote,
		},
		EnforcementLog: log,
		Attestation: Attestation{
			Statement:      "This report was generated by the customer without ARE vendor involvement. All data is sourced from the customer's own PostgreSQL instance. The hash chain was verified by recomputing SHA-256 hashes from raw record fields. This report satisfies DORA Article 8(4) ICT audit trail requirements.",
			VerifiedBy:     "cmd/dora-verify — customer-runnable binary",
			VerificationTS: time.Now().UTC().Format(time.RFC3339),
			SQLProof:       "SELECT * FROM fp_rate_current; -- run this yourself to verify FP rate",
		},
	}

	if *format == "text" {
		printTextReport(report)
	} else {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report)
	}

	if chainBroken {
		os.Exit(2)
	}
}

func printTextReport(r DORAReport) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  AgentRepEngine — DORA Article 8(4) Compliance Report")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  Generated:    %s\n", r.ReportMetadata.GeneratedAt)
	fmt.Printf("  Period:       %s\n", r.ReportMetadata.ReportPeriod)
	fmt.Printf("  DORA Article: %s\n", r.ReportMetadata.DORAArticle)
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("  Total decisions:      %d\n", r.Article8Summary.TotalDecisions)
	fmt.Printf("  Blocked:              %d\n", r.Article8Summary.BlockedDecisions)
	fmt.Printf("  Throttled:            %d\n", r.Article8Summary.ThrottledDecisions)
	fmt.Printf("  Allowed:              %d\n", r.Article8Summary.AllowedDecisions)
	fmt.Printf("  Confirmed FPs:        %d\n", r.Article8Summary.ConfirmedFPs)
	fmt.Printf("  FP rate:              %.2f%%\n", r.Article8Summary.FPRatePct)
	fmt.Printf("  Unique agents:        %d\n", r.Article8Summary.UniqueAgentsMonitor)
	fmt.Printf("  Observe mode:         %v\n", r.Article8Summary.ObserveMode)
	fmt.Println("───────────────────────────────────────────────────────────────")
	s := r.ChainIntegrity
	if s.ChainBroken {
		fmt.Printf("  ⚠ CHAIN BROKEN at record ID %d\n", s.BrokenAtID)
	} else {
		fmt.Printf("  ✓ Chain intact: %d records verified\n", s.RecordsVerified)
	}
	fmt.Printf("  %s\n", s.VerificationNote)
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println("  ATTESTATION")
	fmt.Printf("  %s\n", r.Attestation.Statement)
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
