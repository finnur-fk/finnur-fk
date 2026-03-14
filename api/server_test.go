package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_HandleHealth(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestServer_HandleUpload_Success(t *testing.T) {
	server := NewServer()

	csvData := `Transaction ID,Date,Gross,Fee,Net
TX001,2024-01-15,100.00,-2.90,97.10
TX002,2024-01-16,50.00,-1.75,48.25`

	body, contentType := createMultipartForm("file", "test.csv", csvData)
	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "CSV uploaded and parsed successfully" {
		t.Errorf("Unexpected message: %v", response["message"])
	}

	if response["transaction_count"] != float64(2) {
		t.Errorf("Expected transaction_count 2, got %v", response["transaction_count"])
	}
}

func TestServer_HandleUpload_MethodNotAllowed(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/api/upload", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestServer_HandleUpload_NoFile(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServer_HandleUpload_InvalidCSV(t *testing.T) {
	server := NewServer()

	// CSV without required Transaction ID column
	csvData := `Date,Amount
2024-01-15,100.00`

	body, contentType := createMultipartForm("file", "test.csv", csvData)
	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(response["error"], "Transaction ID") {
		t.Errorf("Expected error about Transaction ID, got: %s", response["error"])
	}
}

func TestServer_HandleGetLiquidity_Success(t *testing.T) {
	server := NewServer()

	// First upload a CSV
	csvData := `Transaction ID,Date,Gross,Fee,Net,Balance,Status,Currency
TX001,2024-01-15,100.00,-2.90,97.10,500.00,Completed,USD
TX002,2024-01-16,50.00,-1.75,48.25,548.25,Completed,USD`

	body, contentType := createMultipartForm("file", "test.csv", csvData)
	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Then get liquidity
	req = httptest.NewRequest(http.MethodGet, "/api/liquidity", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["total_gross"] != 150.0 {
		t.Errorf("Expected total_gross 150.0, got %v", response["total_gross"])
	}

	if response["transaction_count"] != float64(2) {
		t.Errorf("Expected transaction_count 2, got %v", response["transaction_count"])
	}
}

func TestServer_HandleGetLiquidity_NoData(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/api/liquidity", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(response["error"], "No transactions") {
		t.Errorf("Expected error about no transactions, got: %s", response["error"])
	}
}

func TestServer_HandleGetLiquidity_MethodNotAllowed(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodPost, "/api/liquidity", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestServer_HandleUpload_EmptyCSV(t *testing.T) {
	server := NewServer()

	csvData := ``

	body, contentType := createMultipartForm("file", "test.csv", csvData)
	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestIsCSVFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.csv", true},
		{"data.CSV", false},
		{"file.txt", false},
		{"noextension", false},
		{"test.csv.bak", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isCSVFilename(tt.filename)
		if result != tt.expected {
			t.Errorf("isCSVFilename(%q) = %v, expected %v", tt.filename, result, tt.expected)
		}
	}
}

// Helper function to create multipart form data
func createMultipartForm(fieldName, filename, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, _ := writer.CreateFormFile(fieldName, filename)
	io.WriteString(part, content)
	
	contentType := writer.FormDataContentType()
	writer.Close()
	
	return body, contentType
}
