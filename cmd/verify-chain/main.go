package main

import (
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	from := flag.String("from", "", "Start date (YYYY-MM-DD), optional")
	to := flag.String("to", "", "End date (YYYY-MM-DD), optional")
	dsn := flag.String("dsn", "postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable", "PostgreSQL DSN")
	flag.Parse()

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: database unreachable: %v\n", err)
		os.Exit(1)
	}

	query := `
		SELECT id, agent_did, decision, score, prev_hash, this_hash, created_at
		FROM enforcement_decisions
		WHERE 1=1`
	args := []interface{}{}

	if *from != "" {
		t, err := time.Parse("2006-01-02", *from)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid --from date: %v\n", err)
			os.Exit(1)
		}
		args = append(args, t)
		query += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if *to != "" {
		t, err := time.Parse("2006-01-02", *to)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid --to date: %v\n", err)
			os.Exit(1)
		}
		t = t.Add(24 * time.Hour)
		args = append(args, t)
		query += fmt.Sprintf(" AND created_at < $%d", len(args))
	}
	query += " ORDER BY id ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var count int
	var prevHash string

	for rows.Next() {
		var id int64
		var agentDID, decision string
		var score int
		var storedPrev, storedThis sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &agentDID, &decision, &score, &storedPrev, &storedThis, &createdAt); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: scan failed at record %d: %v\n", id, err)
			os.Exit(1)
		}

		// Verify prev_hash linkage
		if count > 0 {
			if !storedPrev.Valid || storedPrev.String != prevHash {
				fmt.Fprintf(os.Stderr, "\nCHAIN BROKEN at record ID %d\n", id)
				fmt.Fprintf(os.Stderr, "  Expected prev_hash: %s\n", prevHash)
				fmt.Fprintf(os.Stderr, "  Stored  prev_hash: %s\n", storedPrev.String)
				os.Exit(2)
			}
		}

		// Recompute this_hash
		raw := fmt.Sprintf("%d|%s|%s|%d|%s", id, agentDID, decision, score, prevHash)
		computed := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))

		if storedThis.Valid && storedThis.String != computed {
			fmt.Fprintf(os.Stderr, "\nHASH MISMATCH at record ID %d\n", id)
			fmt.Fprintf(os.Stderr, "  Computed: %s\n", computed)
			fmt.Fprintf(os.Stderr, "  Stored:   %s\n", storedThis.String)
			os.Exit(2)
		}

		prevHash = computed
		count++
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: row iteration failed: %v\n", err)
		os.Exit(1)
	}

	if count == 0 {
		fmt.Println("Chain intact: 0 records in range (no enforcement decisions recorded yet)")
		os.Exit(0)
	}

	fmt.Printf("Chain intact: %d records verified ✓\n", count)
}
