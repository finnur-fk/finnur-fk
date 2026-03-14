package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/finnur-fk/finnur-fk/internal/liquidity"
	"github.com/finnur-fk/finnur-fk/internal/parser"
)

// Server represents the API server
type Server struct {
	parser     *parser.PayPalParser
	calculator *liquidity.Calculator
	storage    *Storage
	mux        *http.ServeMux
}

// Storage holds uploaded CSV data in memory
type Storage struct {
	mu           sync.RWMutex
	transactions []parser.Transaction
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		transactions: make([]parser.Transaction, 0),
	}
}

// SetTransactions stores transactions
func (s *Storage) SetTransactions(txs []parser.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactions = txs
}

// GetTransactions retrieves stored transactions (returns a copy to prevent race conditions)
func (s *Storage) GetTransactions() []parser.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy of the slice to prevent race conditions
	txsCopy := make([]parser.Transaction, len(s.transactions))
	copy(txsCopy, s.transactions)
	return txsCopy
}

// NewServer creates a new API server
func NewServer() *Server {
	s := &Server{
		parser:     parser.NewPayPalParser(),
		calculator: liquidity.NewCalculator(),
		storage:    NewStorage(),
		mux:        http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up the API routes
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/upload", s.handleUpload)
	s.mux.HandleFunc("/api/liquidity", s.handleGetLiquidity)
	s.mux.HandleFunc("/health", s.handleHealth)
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleUpload processes CSV file uploads
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		sendError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		sendError(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type (filename must end with .csv)
	if !isCSVFilename(header.Filename) {
		sendError(w, "File must have .csv extension", http.StatusBadRequest)
		return
	}

	// Parse the CSV (this provides content validation)
	// The parser will fail if the file doesn't contain valid CSV data
	transactions, err := s.parser.Parse(file)
	if err != nil {
		sendError(w, "Failed to parse CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Store transactions
	s.storage.SetTransactions(transactions)

	// Return success response
	response := map[string]interface{}{
		"message":           "CSV uploaded and parsed successfully",
		"transaction_count": len(transactions),
	}

	sendJSON(w, response, http.StatusOK)
	log.Printf("Uploaded CSV with %d transactions", len(transactions))
}

// handleGetLiquidity calculates and returns liquidity information
func (s *Server) handleGetLiquidity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get stored transactions
	transactions := s.storage.GetTransactions()
	if len(transactions) == 0 {
		sendError(w, "No transactions available. Please upload a CSV first.", http.StatusNotFound)
		return
	}

	// Calculate liquidity
	report, err := s.calculator.Calculate(transactions)
	if err != nil {
		sendError(w, "Failed to calculate liquidity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, report, http.StatusOK)
	log.Printf("Calculated liquidity for %d transactions", len(transactions))
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
	}
	sendJSON(w, response, http.StatusOK)
}

// isCSVFilename checks if filename has .csv extension
func isCSVFilename(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".csv"
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{
		"error": message,
	}
	sendJSON(w, response, statusCode)
}

// Start starts the HTTP server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s)
}
