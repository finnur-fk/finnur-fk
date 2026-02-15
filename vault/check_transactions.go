package main

/*
vault/check_transactions.go

Secure tool to cross-reference PayPal CSV transaction logs with a Go Report Card JSON output.
Usage:
  go build -o vault/vault-check ./vault
  ./vault/vault-check -paypal /path/to/paypal.csv -goreport ./goreportcard-report.json -out ./vault/report-crossref.json

Security:
 - Do NOT hardcode PayPal credentials.
 - Use environment variables for any API keys and read them securely.
 - Validate and sanitize all inputs.
*/

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type PayPalTx struct {
	TxID      string `json:"tx_id"`
	Timestamp string `json:"timestamp"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Note      string `json:"note"`
	// add more fields as needed
}

type GoReportCard struct {
	Package string  `json:"package"`
	Grade   string  `json:"grade"`
	Score   float64 `json:"score"`
	// structure depends on the goreportcard JSON format; adjust as necessary
}

type CrossRefRecord struct {
	PayPalTx       PayPalTx       `json:"paypal_tx"`
	ReferencedSHA  string         `json:"referenced_sha,omitempty"`
	GoReportCard   *GoReportCard  `json:"goreportcard,omitempty"`
	Matched        bool           `json:"matched"`
	Notes          string         `json:"notes,omitempty"`
	CheckedAt      time.Time      `json:"checked_at"`
}

func readPayPalCSV(path string) ([]PayPalTx, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1 // tolerate variable columns
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("paypal csv empty")
	}

	// Heuristic: find header indices for common PayPal columns
	header := rows[0]
	indices := map[string]int{}
	for i, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		indices[h] = i
	}

	txs := make([]PayPalTx, 0, len(rows)-1)
	for _, row := range rows[1:] {
		tx := PayPalTx{}
		if idx, ok := indices["transaction id"]; ok && idx < len(row) {
			tx.TxID = row[idx]
		}
		if idx, ok := indices["date"]; ok && idx < len(row) {
			tx.Timestamp = row[idx]
		}
		if idx, ok := indices["gross"]; ok && idx < len(row) {
			tx.Amount = row[idx]
		}
		if idx, ok := indices["currency"]; ok && idx < len(row) {
			tx.Currency = row[idx]
		}
		// Note/message fields are not consistent across exports
		if idx, ok := indices["notes"]; ok && idx < len(row) {
			tx.Note = row[idx]
		} else if idx, ok := indices["item title"]; ok && idx < len(row) {
			tx.Note = row[idx]
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func readGoReportCardJSON(path string) ([]GoReportCard, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var data []GoReportCard
	if err := json.Unmarshal(b, &data); err != nil {
		// Try to unmarshal single-object reports
		var single GoReportCard
		if err2 := json.Unmarshal(b, &single); err2 == nil {
			return []GoReportCard{single}, nil
		}
		return nil, err
	}
	return data, nil
}

func shaForTx(tx PayPalTx) string {
	h := sha256.New()
	io.WriteString(h, tx.TxID)
	io.WriteString(h, "|")
	io.WriteString(h, tx.Timestamp)
	io.WriteString(h, "|")
	io.WriteString(h, tx.Amount)
	io.WriteString(h, "|")
	io.WriteString(h, tx.Note)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func main() {
	var paypalPath, goreportPath, outPath string
	var timeoutSec int
	flag.StringVar(&paypalPath, "paypal", "", "Path to PayPal CSV export (required)")
	flag.StringVar(&goreportPath, "goreport", "", "Path to goreportcard JSON output (required)")
	flag.StringVar(&outPath, "out", "vault/crossref-output.json", "Output path")
	flag.IntVar(&timeoutSec, "timeout", 30, "Timeout seconds for operations")
	flag.Parse()

	if paypalPath == "" || goreportPath == "" {
		fmt.Fprintln(os.Stderr, "paypal and goreport flags are required")
		flag.Usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	_ = ctx // kept for future expansion (HTTP calls, parallelism)

	txs, err := readPayPalCSV(paypalPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read paypal csv: %v\n", err)
		os.Exit(1)
	}

	grc, err := readGoReportCardJSON(goreportPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read goreportcard json: %v\n", err)
		os.Exit(1)
	}

	// Map goreport entries by package name for quick lookup
	grcMap := map[string]GoReportCard{}
	for _, g := range grc {
		grcMap[g.Package] = g
	}

	out := make([]CrossRefRecord, 0, len(txs))
	for _, tx := range txs {
		rec := CrossRefRecord{
			PayPalTx:  tx,
			Matched:   false,
			CheckedAt: time.Now().UTC(),
		}
		// Heuristic: try to extract a referenced package or commit SHA from tx.Note
		note := strings.TrimSpace(tx.Note)
		if note != "" {
			rec.ReferencedSHA = note
			// Attempt to find package in goreport by substring match
			for pkg, g := range grcMap {
				if strings.Contains(strings.ToLower(pkg), strings.ToLower(note)) || strings.Contains(strings.ToLower(note), strings.ToLower(pkg)) {
					tmp := g
					rec.GoReportCard = &tmp
					rec.Matched = true
					rec.Notes = "matched by package substring"
					break
				}
			}
		}
		// If not matched, attach best-effort mapping: tie via hashed tx as identifier
		if !rec.Matched {
			hash := shaForTx(tx)
			rec.ReferencedSHA = hash
			rec.Notes = "no package match; stored transaction hash"
		}
		out = append(out, rec)
	}

	bout, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal output: %v\n", err)
		os.Exit(1)
	}

	// Ensure vault directory exists
	if err := os.MkdirAll("vault", 0700); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create vault dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, bout, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cross-reference output written to %s\n", outPath)
}