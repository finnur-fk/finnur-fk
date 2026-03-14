package liquidity

import (
	"testing"

	"github.com/finnur-fk/finnur-fk/internal/parser"
)

func TestCalculator_Calculate_Success(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Gross: 100.00, Fee: -2.90, Net: 97.10, Balance: 500.00, Status: "Completed", Currency: "USD"},
		{TransactionID: "TX2", Gross: 50.00, Fee: -1.75, Net: 48.25, Balance: 548.25, Status: "Completed", Currency: "USD"},
		{TransactionID: "TX3", Gross: 75.00, Fee: -2.50, Net: 72.50, Balance: 620.75, Status: "Completed", Currency: "USD"},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.TotalGross != 225.00 {
		t.Errorf("Expected TotalGross 225.00, got %.2f", report.TotalGross)
	}

	if report.TotalFees != -7.15 {
		t.Errorf("Expected TotalFees -7.15, got %.2f", report.TotalFees)
	}

	if report.TotalNet != 217.85 {
		t.Errorf("Expected TotalNet 217.85, got %.2f", report.TotalNet)
	}

	if report.FinalBalance != 620.75 {
		t.Errorf("Expected FinalBalance 620.75 (last transaction's balance), got %.2f", report.FinalBalance)
	}

	if report.TransactionCount != 3 {
		t.Errorf("Expected TransactionCount 3, got %d", report.TransactionCount)
	}

	if report.CompletedCount != 3 {
		t.Errorf("Expected CompletedCount 3, got %d", report.CompletedCount)
	}
}

func TestCalculator_Calculate_MultipleCurrencies(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Net: 100.00, Currency: "USD"},
		{TransactionID: "TX2", Net: 50.00, Currency: "EUR"},
		{TransactionID: "TX3", Net: 75.00, Currency: "USD"},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.ByCurrency["USD"] != 175.00 {
		t.Errorf("Expected USD total 175.00, got %.2f", report.ByCurrency["USD"])
	}

	if report.ByCurrency["EUR"] != 50.00 {
		t.Errorf("Expected EUR total 50.00, got %.2f", report.ByCurrency["EUR"])
	}
}

func TestCalculator_Calculate_EmptyTransactions(t *testing.T) {
	transactions := []parser.Transaction{}

	calculator := NewCalculator()
	_, err := calculator.Calculate(transactions)

	if err == nil {
		t.Fatal("Expected error for empty transactions, got nil")
	}

	if err.Error() != "no transactions to calculate" {
		t.Errorf("Expected 'no transactions to calculate' error, got: %v", err)
	}
}

func TestCalculator_Calculate_MixedStatuses(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Net: 100.00, Status: "Completed"},
		{TransactionID: "TX2", Net: 50.00, Status: "Pending"},
		{TransactionID: "TX3", Net: 75.00, Status: "Refunded"},
		{TransactionID: "TX4", Net: 25.00, Status: "Completed"},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.CompletedCount != 2 {
		t.Errorf("Expected CompletedCount 2, got %d", report.CompletedCount)
	}

	if report.PendingCount != 1 {
		t.Errorf("Expected PendingCount 1, got %d", report.PendingCount)
	}

	if report.RefundedCount != 1 {
		t.Errorf("Expected RefundedCount 1, got %d", report.RefundedCount)
	}

	// Total should include all transactions
	if report.TotalNet != 250.00 {
		t.Errorf("Expected TotalNet 250.00, got %.2f", report.TotalNet)
	}
}

func TestCalculator_Calculate_NoBalanceField(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Gross: 100.00, Fee: -2.90, Net: 97.10},
		{TransactionID: "TX2", Gross: 50.00, Fee: -1.75, Net: 48.25},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// When no balance field, should use TotalNet
	if report.FinalBalance != report.TotalNet {
		t.Errorf("Expected FinalBalance to equal TotalNet (%.2f), got %.2f", report.TotalNet, report.FinalBalance)
	}
}

func TestCalculator_Calculate_NegativeAmounts(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Gross: -50.00, Fee: -2.00, Net: -52.00, Status: "Refunded"},
		{TransactionID: "TX2", Gross: 100.00, Fee: -3.00, Net: 97.00, Status: "Completed"},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.TotalGross != 50.00 {
		t.Errorf("Expected TotalGross 50.00, got %.2f", report.TotalGross)
	}

	if report.TotalNet != 45.00 {
		t.Errorf("Expected TotalNet 45.00, got %.2f", report.TotalNet)
	}
}

func TestCalculator_CalculateForCompleted_OnlyCompleted(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Net: 100.00, Status: "Completed"},
		{TransactionID: "TX2", Net: 50.00, Status: "Pending"},
		{TransactionID: "TX3", Net: 75.00, Status: "Completed"},
		{TransactionID: "TX4", Net: 25.00, Status: "Refunded"},
	}

	calculator := NewCalculator()
	report, err := calculator.CalculateForCompleted(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.TotalNet != 175.00 {
		t.Errorf("Expected TotalNet 175.00 (only completed), got %.2f", report.TotalNet)
	}

	if report.TransactionCount != 2 {
		t.Errorf("Expected TransactionCount 2, got %d", report.TransactionCount)
	}

	if report.CompletedCount != 2 {
		t.Errorf("Expected CompletedCount 2, got %d", report.CompletedCount)
	}
}

func TestCalculator_CalculateForCompleted_NoCompleted(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Net: 100.00, Status: "Pending"},
		{TransactionID: "TX2", Net: 50.00, Status: "Pending"},
	}

	calculator := NewCalculator()
	report, err := calculator.CalculateForCompleted(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.TotalNet != 0.00 {
		t.Errorf("Expected TotalNet 0.00 (no completed), got %.2f", report.TotalNet)
	}

	if report.TransactionCount != 0 {
		t.Errorf("Expected TransactionCount 0, got %d", report.TransactionCount)
	}
}

func TestCalculator_Calculate_StatusNormalization(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Net: 100.00, Status: "completed"},
		{TransactionID: "TX2", Net: 50.00, Status: "Completed"},
		{TransactionID: "TX3", Net: 75.00, Status: "Success"},
		{TransactionID: "TX4", Net: 25.00, Status: "pending"},
		{TransactionID: "TX5", Net: 30.00, Status: "Pending"},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// All variations of completed should be counted
	if report.CompletedCount != 3 {
		t.Errorf("Expected CompletedCount 3, got %d", report.CompletedCount)
	}

	// All variations of pending should be counted
	if report.PendingCount != 2 {
		t.Errorf("Expected PendingCount 2, got %d", report.PendingCount)
	}
}

func TestCalculator_Calculate_ZeroAmounts(t *testing.T) {
	transactions := []parser.Transaction{
		{TransactionID: "TX1", Gross: 0.00, Fee: 0.00, Net: 0.00, Balance: 0.00},
	}

	calculator := NewCalculator()
	report, err := calculator.Calculate(transactions)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.TotalGross != 0.00 {
		t.Errorf("Expected TotalGross 0.00, got %.2f", report.TotalGross)
	}

	if report.TotalNet != 0.00 {
		t.Errorf("Expected TotalNet 0.00, got %.2f", report.TotalNet)
	}
}
