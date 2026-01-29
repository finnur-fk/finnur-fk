package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaxComplianceView(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/dashboard?view=tax-compliance", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDashboard)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	// Tax-compliance view should only show Public transactions
	if response.Count != 6 {
		t.Errorf("Expected 6 public transactions, got %d", response.Count)
	}

	// Verify all transactions are Public
	for _, tx := range response.Transactions {
		if tx.Type != "Public" {
			t.Errorf("Found non-public transaction %s in tax-compliance view", tx.ID)
		}
	}

	// Verify total amount matches public transactions only
	expectedTotal := 10350.00 // Sum of public transactions
	if response.TotalAmount != expectedTotal {
		t.Errorf("Expected total amount %.2f, got %.2f", expectedTotal, response.TotalAmount)
	}
}

func TestFullViewWithoutKey(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/dashboard?view=full", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDashboard)
	handler.ServeHTTP(rr, req)

	// Should return 401 Unauthorized without architect key
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestFullViewWithValidKey(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/dashboard?view=full", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Architect-Key", defaultKey)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDashboard)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	// Full view should show all transactions
	if response.Count != 10 {
		t.Errorf("Expected 10 total transactions, got %d", response.Count)
	}

	// Verify we have both Public and Private/Crypto transactions
	publicCount := 0
	privateCount := 0
	for _, tx := range response.Transactions {
		if tx.Type == "Public" {
			publicCount++
		} else if tx.Type == "Private/Crypto" {
			privateCount++
		}
	}

	if publicCount != 6 {
		t.Errorf("Expected 6 public transactions, got %d", publicCount)
	}
	if privateCount != 4 {
		t.Errorf("Expected 4 private/crypto transactions, got %d", privateCount)
	}
}

func TestFullViewWithInvalidKey(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/dashboard?view=full", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Architect-Key", "wrong-key")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDashboard)
	handler.ServeHTTP(rr, req)

	// Should return 401 Unauthorized with wrong key
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestPrivateTransactionWithoutKey(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/transactions?id=T003", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTransactions)
	handler.ServeHTTP(rr, req)

	// T003 is a Private/Crypto transaction, should require key
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestPrivateTransactionWithKey(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/transactions?id=T003", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Architect-Key", defaultKey)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTransactions)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var transaction Transaction
	if err := json.NewDecoder(rr.Body).Decode(&transaction); err != nil {
		t.Fatal(err)
	}

	if transaction.ID != "T003" {
		t.Errorf("Expected transaction T003, got %s", transaction.ID)
	}
	if transaction.Type != "Private/Crypto" {
		t.Errorf("Expected Private/Crypto transaction, got %s", transaction.Type)
	}
	if transaction.EncryptedData == "" {
		t.Error("Expected encrypted data marker, got empty string")
	}
}

func TestPublicTransactionWithoutKey(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/transactions?id=T001", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTransactions)
	handler.ServeHTTP(rr, req)

	// T001 is a Public transaction, should not require key
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var transaction Transaction
	if err := json.NewDecoder(rr.Body).Decode(&transaction); err != nil {
		t.Fatal(err)
	}

	if transaction.ID != "T001" {
		t.Errorf("Expected transaction T001, got %s", transaction.ID)
	}
	if transaction.Type != "Public" {
		t.Errorf("Expected Public transaction, got %s", transaction.Type)
	}
}

func TestDefaultViewMode(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDashboard)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	// Default should behave like tax-compliance (only public transactions)
	if response.Count != 6 {
		t.Errorf("Expected 6 public transactions in default view, got %d", response.Count)
	}
}
