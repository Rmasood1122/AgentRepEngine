package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type PortfolioResponse struct {
	AccountID   string     `json:"account_id"`
	AgentDID    string     `json:"agent_did"`
	TotalValue  float64    `json:"total_value_usd"`
	Positions   []Position `json:"positions"`
	LastUpdated string     `json:"last_updated"`
	RequestedAt string     `json:"requested_at"`
}

type Position struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	Value    float64 `json:"value_usd"`
}

type TransactionResponse struct {
	AccountID    string        `json:"account_id"`
	AgentDID     string        `json:"agent_did"`
	Transactions []Transaction `json:"transactions"`
	Count        int           `json:"count"`
	RequestedAt  string        `json:"requested_at"`
}

type Transaction struct {
	TxID      string  `json:"tx_id"`
	Type      string  `json:"type"`
	Symbol    string  `json:"symbol"`
	Amount    float64 `json:"amount_usd"`
	Timestamp string  `json:"timestamp"`
}

func portfolioHandler(w http.ResponseWriter, r *http.Request) {
	agentDID := r.Header.Get("X-Agent-DID-Verified")
	if agentDID == "" {
		agentDID = "did:jwt:finserv-demo:trading-agent:001"
	}
	resp := PortfolioResponse{
		AccountID:  "ACC-7821-DEMO",
		AgentDID:   agentDID,
		TotalValue: 4287500.00,
		Positions: []Position{
			{Symbol: "AAPL", Quantity: 1500, Value: 285000.00},
			{Symbol: "MSFT", Quantity: 2000, Value: 742000.00},
			{Symbol: "BRK.B", Quantity: 800, Value: 320000.00},
			{Symbol: "SPY", Quantity: 5000, Value: 2440500.00},
			{Symbol: "TLT", Quantity: 1200, Value: 500000.00},
		},
		LastUpdated: "2026-04-16T09:30:00Z",
		RequestedAt: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Finserv-Demo", "true")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	agentDID := r.Header.Get("X-Agent-DID-Verified")
	if agentDID == "" {
		agentDID = "did:jwt:finserv-demo:trading-agent:001"
	}
	resp := TransactionResponse{
		AccountID: "ACC-7821-DEMO",
		AgentDID:  agentDID,
		Count:     5,
		Transactions: []Transaction{
			{TxID: "TX-001", Type: "BUY", Symbol: "AAPL", Amount: 47500.00, Timestamp: "2026-04-15T14:22:00Z"},
			{TxID: "TX-002", Type: "SELL", Symbol: "MSFT", Amount: 18550.00, Timestamp: "2026-04-15T11:05:00Z"},
			{TxID: "TX-003", Type: "BUY", Symbol: "SPY", Amount: 122025.00, Timestamp: "2026-04-14T15:48:00Z"},
			{TxID: "TX-004", Type: "BUY", Symbol: "TLT", Amount: 41667.00, Timestamp: "2026-04-14T09:31:00Z"},
			{TxID: "TX-005", Type: "SELL", Symbol: "BRK.B", Amount: 40000.00, Timestamp: "2026-04-13T13:17:00Z"},
		},
		RequestedAt: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Finserv-Demo", "true")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"finserv-demo-api","version":"1.0.0"}`))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/portfolio", portfolioHandler)
	mux.HandleFunc("/api/v1/transactions", transactionsHandler)
	mux.HandleFunc("/", portfolioHandler)
	log.Printf("finserv-demo-api listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
