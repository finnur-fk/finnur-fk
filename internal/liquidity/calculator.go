package liquidity

import (
	"fmt"
	"github.com/finnur-fk/finnur-fk/internal/parser"
)

// Calculator handles liquidity calculations
type Calculator struct{}

// NewCalculator creates a new liquidity calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// LiquidityReport contains the calculated liquidity information
type LiquidityReport struct {
	TotalGross      float64            `json:"total_gross"`
	TotalFees       float64            `json:"total_fees"`
	TotalNet        float64            `json:"total_net"`
	FinalBalance    float64            `json:"final_balance"`
	TransactionCount int               `json:"transaction_count"`
	ByCurrency      map[string]float64 `json:"by_currency"`
	CompletedCount  int                `json:"completed_count"`
	PendingCount    int                `json:"pending_count"`
	RefundedCount   int                `json:"refunded_count"`
}

// Calculate computes total liquidity from a list of transactions
func (c *Calculator) Calculate(transactions []parser.Transaction) (*LiquidityReport, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions to calculate")
	}

	report := &LiquidityReport{
		ByCurrency: make(map[string]float64),
	}

	var maxBalance float64
	maxBalanceSet := false

	for _, tx := range transactions {
		// Sum up amounts
		report.TotalGross += tx.Gross
		report.TotalFees += tx.Fee
		report.TotalNet += tx.Net
		report.TransactionCount++

		// Track by currency
		if tx.Currency != "" {
			report.ByCurrency[tx.Currency] += tx.Net
		}

		// Track the latest balance (highest balance value found)
		if tx.Balance != 0 {
			if !maxBalanceSet || tx.Balance > maxBalance {
				maxBalance = tx.Balance
				maxBalanceSet = true
			}
		}

		// Count transaction statuses
		status := normalizeStatus(tx.Status)
		switch status {
		case "completed":
			report.CompletedCount++
		case "pending":
			report.PendingCount++
		case "refunded":
			report.RefundedCount++
		}
	}

	// Set final balance
	if maxBalanceSet {
		report.FinalBalance = maxBalance
	} else {
		// If no balance field, use total net as approximation
		report.FinalBalance = report.TotalNet
	}

	return report, nil
}

// CalculateForCompleted calculates liquidity only for completed transactions
func (c *Calculator) CalculateForCompleted(transactions []parser.Transaction) (*LiquidityReport, error) {
	completed := make([]parser.Transaction, 0)
	for _, tx := range transactions {
		if normalizeStatus(tx.Status) == "completed" {
			completed = append(completed, tx)
		}
	}

	if len(completed) == 0 {
		return &LiquidityReport{
			ByCurrency: make(map[string]float64),
		}, nil
	}

	return c.Calculate(completed)
}

// normalizeStatus converts various status strings to a standard format
func normalizeStatus(status string) string {
	switch status {
	case "Completed", "completed", "Complete", "Success", "success":
		return "completed"
	case "Pending", "pending", "In Progress", "Processing":
		return "pending"
	case "Refunded", "refunded", "Reversed", "Cancelled", "cancelled":
		return "refunded"
	default:
		return "other"
	}
}
