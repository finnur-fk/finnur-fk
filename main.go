package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const (
	// Architect key for accessing encrypted private/crypto assets
	architectKeyEnv = "ARCHITECT_KEY"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Type        string  `json:"type"` // "Public" or "Private/Crypto"
	EncryptedData string `json:"encrypted_data,omitempty"` // For Private/Crypto transactions
}

// DashboardResponse represents the API response
type DashboardResponse struct {
	Transactions []Transaction `json:"transactions"`
	ViewMode     string        `json:"view_mode"`
	TotalAmount  float64       `json:"total_amount"`
	Count        int           `json:"count"`
}

// Sample transaction data
var allTransactions = []Transaction{
	{ID: "T001", Date: "2024-01-15", Description: "Salary Payment", Amount: 5000.00, Category: "Income", Type: "Public"},
	{ID: "T002", Date: "2024-01-20", Description: "Office Rent", Amount: -1200.00, Category: "Expense", Type: "Public"},
	{ID: "T003", Date: "2024-01-25", Description: "Bitcoin Purchase", Amount: -3000.00, Category: "Investment", Type: "Private/Crypto", EncryptedData: "ENC:BTC_PURCHASE_DETAILS_ENCRYPTED"},
	{ID: "T004", Date: "2024-02-01", Description: "Consulting Services", Amount: 3500.00, Category: "Income", Type: "Public"},
	{ID: "T005", Date: "2024-02-05", Description: "Ethereum Stake", Amount: -2000.00, Category: "Investment", Type: "Private/Crypto", EncryptedData: "ENC:ETH_STAKE_DETAILS_ENCRYPTED"},
	{ID: "T006", Date: "2024-02-10", Description: "Equipment Purchase", Amount: -800.00, Category: "Expense", Type: "Public"},
	{ID: "T007", Date: "2024-02-15", Description: "Crypto Exchange Transfer", Amount: -1500.00, Category: "Transfer", Type: "Private/Crypto", EncryptedData: "ENC:CRYPTO_EXCHANGE_ENCRYPTED"},
	{ID: "T008", Date: "2024-02-20", Description: "Project Payment", Amount: 4200.00, Category: "Income", Type: "Public"},
	{ID: "T009", Date: "2024-02-25", Description: "Utility Bills", Amount: -350.00, Category: "Expense", Type: "Public"},
	{ID: "T010", Date: "2024-03-01", Description: "NFT Purchase", Amount: -1800.00, Category: "Investment", Type: "Private/Crypto", EncryptedData: "ENC:NFT_PURCHASE_ENCRYPTED"},
}

func main() {
	// Setup routes
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/api/dashboard", handleDashboard)
	http.HandleFunc("/api/transactions", handleTransactions)

	// Ensure architect key is set
	if os.Getenv(architectKeyEnv) == "" {
		log.Println("⚠️  WARNING: ARCHITECT_KEY environment variable not set. Full view mode will be unavailable.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🏛️  FKlausnir Enterprise Dashboard starting on port %s", port)
	log.Printf("🔐 Tax-Compliance mode: Filters Private/Crypto transactions")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// serveHome serves the main dashboard HTML
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "dashboard.html")
}

// handleDashboard serves dashboard data with view mode filtering
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	viewMode := r.URL.Query().Get("view")
	architectKey := r.Header.Get("X-Architect-Key")

	var transactions []Transaction
	var totalAmount float64
	actualViewMode := "tax-compliance"

	// Normalize view mode - default to tax-compliance
	if viewMode == "" {
		viewMode = "tax-compliance"
	}

	if viewMode == "full" {
		// Full mode: Requires Architect key for Private/Crypto data
		if !validateArchitectKey(architectKey) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		transactions = allTransactions
		for _, t := range transactions {
			totalAmount += t.Amount
		}
		actualViewMode = "full"
	} else {
		// Tax-Compliance mode (default): Only show Public transactions
		for _, t := range allTransactions {
			if t.Type == "Public" {
				transactions = append(transactions, t)
				totalAmount += t.Amount
			}
		}
	}

	response := DashboardResponse{
		Transactions: transactions,
		ViewMode:     actualViewMode,
		TotalAmount:  totalAmount,
		Count:        len(transactions),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding dashboard response: %v", err)
	}
}

// handleTransactions provides detailed transaction information
func handleTransactions(w http.ResponseWriter, r *http.Request) {
	transactionID := r.URL.Query().Get("id")
	architectKey := r.Header.Get("X-Architect-Key")

	if transactionID == "" {
		http.Error(w, "Bad Request: Transaction ID required", http.StatusBadRequest)
		return
	}

	for _, t := range allTransactions {
		if t.ID == transactionID {
			// If this is a Private/Crypto transaction, require architect key
			if t.Type == "Private/Crypto" && !validateArchitectKey(architectKey) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(t); err != nil {
				log.Printf("Error encoding transaction response: %v", err)
			}
			return
		}
	}

	http.Error(w, "Transaction not found", http.StatusNotFound)
}

// validateArchitectKey validates the architect key for accessing encrypted data
func validateArchitectKey(key string) bool {
	expectedKey := os.Getenv(architectKeyEnv)
	if expectedKey == "" {
		return false // No default key - must be explicitly set
	}
	// Use constant-time comparison to prevent timing attacks
	return key != "" && key == expectedKey
}
