package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Transaction represents a parsed PayPal transaction
type Transaction struct {
	TransactionID string
	Date          string
	Name          string
	Type          string
	Status        string
	Currency      string
	Gross         float64
	Fee           float64
	Net           float64
	Balance       float64
	Note          string
}

// PayPalParser handles parsing of PayPal CSV files
type PayPalParser struct{}

// NewPayPalParser creates a new PayPal CSV parser
func NewPayPalParser() *PayPalParser {
	return &PayPalParser{}
}

// Parse reads and parses a PayPal CSV file
func (p *PayPalParser) Parse(reader io.Reader) ([]Transaction, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	csvReader.TrimLeadingSpace = true

	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Parse header to find column indices
	header := rows[0]
	indices, err := p.parseHeader(header)
	if err != nil {
		return nil, err
	}

	// Parse transactions
	transactions := make([]Transaction, 0, len(rows)-1)
	for i, row := range rows[1:] {
		if len(row) == 0 || isEmptyRow(row) {
			continue // Skip empty rows
		}

		tx, err := p.parseTransaction(row, indices)
		if err != nil {
			return nil, fmt.Errorf("error parsing row %d: %w", i+2, err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// parseHeader identifies column positions in the CSV header
func (p *PayPalParser) parseHeader(header []string) (map[string]int, error) {
	indices := make(map[string]int)
	
	for i, col := range header {
		col = strings.ToLower(strings.TrimSpace(col))
		
		// Map various PayPal header variations to standard names
		switch {
		case strings.Contains(col, "transaction") && strings.Contains(col, "id"):
			indices["transaction_id"] = i
		case col == "date" || col == "timestamp":
			indices["date"] = i
		case col == "name" || col == "from name" || col == "to name":
			indices["name"] = i
		case col == "type" || col == "transaction type":
			indices["type"] = i
		case col == "status" || col == "transaction status":
			indices["status"] = i
		case col == "currency" || col == "currency code":
			indices["currency"] = i
		case col == "gross" || col == "amount" || col == "gross amount":
			indices["gross"] = i
		case col == "fee" || col == "fee amount":
			indices["fee"] = i
		case col == "net" || col == "net amount":
			indices["net"] = i
		case col == "balance" || col == "account balance":
			indices["balance"] = i
		case strings.Contains(col, "note") || strings.Contains(col, "message") || col == "item title":
			indices["note"] = i
		}
	}

	// Validate required fields
	if _, ok := indices["transaction_id"]; !ok {
		return nil, fmt.Errorf("CSV missing required field: Transaction ID")
	}

	return indices, nil
}

// parseTransaction converts a CSV row to a Transaction struct
func (p *PayPalParser) parseTransaction(row []string, indices map[string]int) (Transaction, error) {
	tx := Transaction{}

	// Required field
	if idx, ok := indices["transaction_id"]; ok && idx < len(row) {
		tx.TransactionID = strings.TrimSpace(row[idx])
		if tx.TransactionID == "" {
			return tx, fmt.Errorf("transaction ID is empty")
		}
	} else {
		return tx, fmt.Errorf("transaction ID not found")
	}

	// Optional fields with safe access
	if idx, ok := indices["date"]; ok && idx < len(row) {
		tx.Date = strings.TrimSpace(row[idx])
	}
	if idx, ok := indices["name"]; ok && idx < len(row) {
		tx.Name = strings.TrimSpace(row[idx])
	}
	if idx, ok := indices["type"]; ok && idx < len(row) {
		tx.Type = strings.TrimSpace(row[idx])
	}
	if idx, ok := indices["status"]; ok && idx < len(row) {
		tx.Status = strings.TrimSpace(row[idx])
	}
	if idx, ok := indices["currency"]; ok && idx < len(row) {
		tx.Currency = strings.TrimSpace(row[idx])
	}
	if idx, ok := indices["note"]; ok && idx < len(row) {
		tx.Note = strings.TrimSpace(row[idx])
	}

	// Parse numeric fields
	var err error
	if idx, ok := indices["gross"]; ok && idx < len(row) {
		tx.Gross, err = parseAmount(row[idx])
		if err != nil {
			return tx, fmt.Errorf("invalid gross amount: %w", err)
		}
	}
	if idx, ok := indices["fee"]; ok && idx < len(row) {
		tx.Fee, err = parseAmount(row[idx])
		if err != nil {
			return tx, fmt.Errorf("invalid fee amount: %w", err)
		}
	}
	if idx, ok := indices["net"]; ok && idx < len(row) {
		tx.Net, err = parseAmount(row[idx])
		if err != nil {
			return tx, fmt.Errorf("invalid net amount: %w", err)
		}
	}
	if idx, ok := indices["balance"]; ok && idx < len(row) {
		tx.Balance, err = parseAmount(row[idx])
		if err != nil {
			return tx, fmt.Errorf("invalid balance amount: %w", err)
		}
	}

	return tx, nil
}

// parseAmount converts a string amount to float64, handling various formats
func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0.0, nil
	}

	// Remove common currency symbols and formatting
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "€", "")
	s = strings.ReplaceAll(s, "£", "")
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0.0, nil
	}

	amount, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, fmt.Errorf("cannot parse amount '%s': %w", s, err)
	}

	return amount, nil
}

// isEmptyRow checks if a CSV row is empty
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
