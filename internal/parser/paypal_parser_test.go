package parser

import (
	"strings"
	"testing"
)

func TestPayPalParser_Parse_Success(t *testing.T) {
	csvData := `Transaction ID,Date,Name,Type,Status,Currency,Gross,Fee,Net,Balance
TX123,2024-01-15,John Doe,Payment,Completed,USD,100.00,-2.90,97.10,500.00
TX124,2024-01-16,Jane Smith,Payment,Completed,USD,50.00,-1.75,48.25,548.25`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}

	// Verify first transaction
	tx := transactions[0]
	if tx.TransactionID != "TX123" {
		t.Errorf("Expected TransactionID 'TX123', got '%s'", tx.TransactionID)
	}
	if tx.Date != "2024-01-15" {
		t.Errorf("Expected Date '2024-01-15', got '%s'", tx.Date)
	}
	if tx.Name != "John Doe" {
		t.Errorf("Expected Name 'John Doe', got '%s'", tx.Name)
	}
	if tx.Gross != 100.00 {
		t.Errorf("Expected Gross 100.00, got %.2f", tx.Gross)
	}
	if tx.Fee != -2.90 {
		t.Errorf("Expected Fee -2.90, got %.2f", tx.Fee)
	}
	if tx.Net != 97.10 {
		t.Errorf("Expected Net 97.10, got %.2f", tx.Net)
	}
	if tx.Balance != 500.00 {
		t.Errorf("Expected Balance 500.00, got %.2f", tx.Balance)
	}
}

func TestPayPalParser_Parse_WithCurrencySymbols(t *testing.T) {
	csvData := `Transaction ID,Date,Gross,Fee,Net
TX125,2024-01-17,"$1,000.50","-$29.90","$970.60"`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.Gross != 1000.50 {
		t.Errorf("Expected Gross 1000.50, got %.2f", tx.Gross)
	}
	if tx.Fee != -29.90 {
		t.Errorf("Expected Fee -29.90, got %.2f", tx.Fee)
	}
	if tx.Net != 970.60 {
		t.Errorf("Expected Net 970.60, got %.2f", tx.Net)
	}
}

func TestPayPalParser_Parse_EmptyFile(t *testing.T) {
	csvData := ``

	parser := NewPayPalParser()
	_, err := parser.Parse(strings.NewReader(csvData))

	if err == nil {
		t.Fatal("Expected error for empty CSV, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected 'empty' in error message, got: %v", err)
	}
}

func TestPayPalParser_Parse_MissingTransactionID(t *testing.T) {
	csvData := `Date,Amount,Fee
2024-01-15,100.00,2.90`

	parser := NewPayPalParser()
	_, err := parser.Parse(strings.NewReader(csvData))

	if err == nil {
		t.Fatal("Expected error for missing Transaction ID, got nil")
	}

	if !strings.Contains(err.Error(), "Transaction ID") {
		t.Errorf("Expected 'Transaction ID' in error message, got: %v", err)
	}
}

func TestPayPalParser_Parse_EmptyTransactionID(t *testing.T) {
	csvData := `Transaction ID,Date,Gross
,2024-01-15,100.00`

	parser := NewPayPalParser()
	_, err := parser.Parse(strings.NewReader(csvData))

	if err == nil {
		t.Fatal("Expected error for empty transaction ID, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected 'empty' in error message, got: %v", err)
	}
}

func TestPayPalParser_Parse_InvalidAmount(t *testing.T) {
	csvData := `Transaction ID,Gross
TX126,not-a-number`

	parser := NewPayPalParser()
	_, err := parser.Parse(strings.NewReader(csvData))

	if err == nil {
		t.Fatal("Expected error for invalid amount, got nil")
	}

	if !strings.Contains(err.Error(), "gross amount") {
		t.Errorf("Expected 'gross amount' in error message, got: %v", err)
	}
}

func TestPayPalParser_Parse_VariableColumns(t *testing.T) {
	csvData := `Transaction ID,Date,Gross
TX127,2024-01-18,100.00
TX128,2024-01-19,200.00,Extra,Columns`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error with variable columns, got: %v", err)
	}

	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}
}

func TestPayPalParser_Parse_EmptyRows(t *testing.T) {
	csvData := `Transaction ID,Date,Gross
TX129,2024-01-20,100.00
,,,
TX130,2024-01-21,200.00`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should skip empty row
	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}

	if transactions[0].TransactionID != "TX129" {
		t.Errorf("Expected first transaction ID 'TX129', got '%s'", transactions[0].TransactionID)
	}
	if transactions[1].TransactionID != "TX130" {
		t.Errorf("Expected second transaction ID 'TX130', got '%s'", transactions[1].TransactionID)
	}
}

func TestPayPalParser_Parse_AlternativeHeaders(t *testing.T) {
	// Test with different header variations that PayPal might use
	csvData := `transaction id,timestamp,gross amount,fee amount,net amount,account balance
TX131,2024-01-22,150.00,4.50,145.50,700.00`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error with alternative headers, got: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.TransactionID != "TX131" {
		t.Errorf("Expected TransactionID 'TX131', got '%s'", tx.TransactionID)
	}
	if tx.Gross != 150.00 {
		t.Errorf("Expected Gross 150.00, got %.2f", tx.Gross)
	}
	if tx.Balance != 700.00 {
		t.Errorf("Expected Balance 700.00, got %.2f", tx.Balance)
	}
}

func TestPayPalParser_Parse_ZeroAmounts(t *testing.T) {
	csvData := `Transaction ID,Gross,Fee,Net
TX132,0.00,0.00,0.00`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error with zero amounts, got: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.Gross != 0.00 {
		t.Errorf("Expected Gross 0.00, got %.2f", tx.Gross)
	}
}

func TestPayPalParser_Parse_NegativeAmounts(t *testing.T) {
	csvData := `Transaction ID,Gross,Fee,Net
TX133,-50.00,-2.50,-52.50`

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error with negative amounts, got: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.Gross != -50.00 {
		t.Errorf("Expected Gross -50.00, got %.2f", tx.Gross)
	}
	if tx.Fee != -2.50 {
		t.Errorf("Expected Fee -2.50, got %.2f", tx.Fee)
	}
}

func TestPayPalParser_Parse_WithWhitespace(t *testing.T) {
	csvData := `  Transaction ID  ,  Date  ,  Gross  
  TX134  ,  2024-01-23  ,  100.00  `

	parser := NewPayPalParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error with whitespace, got: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.TransactionID != "TX134" {
		t.Errorf("Expected TransactionID 'TX134', got '%s'", tx.TransactionID)
	}
	if tx.Date != "2024-01-23" {
		t.Errorf("Expected Date '2024-01-23', got '%s'", tx.Date)
	}
}
