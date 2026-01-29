package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uploadDir = "./uploads"

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Handle CSV uploads
	http.HandleFunc("/upload", handleUpload)

	port := "8080"
	fmt.Printf("🚀 FKlausnir Enterprise Engine Dashboard\n")
	fmt.Printf("📊 Server running on http://localhost:%s\n", port)
	fmt.Printf("📁 Upload PayPal CSV files at: http://localhost:%s/\n\n", port)
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleUpload processes HTTP POST requests for CSV file uploads.
// It validates the file type, sanitizes the filename, and stores it securely in the uploads directory.
func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, false, "Method not allowed", "")
		return
	}

	// Parse multipart form with 10 MB limit
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		respondJSON(w, http.StatusBadRequest, false, "Unable to parse form", "")
		return
	}

	file, handler, err := r.FormFile("csvfile")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		respondJSON(w, http.StatusBadRequest, false, "Error retrieving file", "")
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(handler.Filename), ".csv") {
		respondJSON(w, http.StatusBadRequest, false, "Only CSV files are allowed", "")
		return
	}

	// Sanitize filename to prevent path traversal
	safeFilename := filepath.Base(handler.Filename)
	safeFilename = strings.ReplaceAll(safeFilename, "..", "")
	
	// Add timestamp to prevent overwriting
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(safeFilename)
	nameWithoutExt := strings.TrimSuffix(safeFilename, ext)
	uniqueFilename := fmt.Sprintf("%s-%s%s", nameWithoutExt, timestamp, ext)

	// Create destination file
	destPath := filepath.Join(uploadDir, uniqueFilename)
	dst, err := os.Create(destPath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		respondJSON(w, http.StatusInternalServerError, false, "Unable to save file", "")
		return
	}
	defer dst.Close()

	// Copy file content with size limit check
	written, err := io.CopyN(dst, file, 10<<20+1) // Try to copy 1 byte more than limit
	if err != nil && err != io.EOF {
		log.Printf("Error copying file: %v", err)
		respondJSON(w, http.StatusInternalServerError, false, "Unable to save file", "")
		return
	}
	if written > 10<<20 {
		os.Remove(destPath)
		respondJSON(w, http.StatusBadRequest, false, "File size exceeds 10 MB limit", "")
		return
	}

	log.Printf("✅ Uploaded: %s (%.2f KB)\n", uniqueFilename, float64(written)/1024)
	respondJSON(w, http.StatusOK, true, "File uploaded successfully", uniqueFilename)
}

// respondJSON sends a consistent JSON response
func respondJSON(w http.ResponseWriter, status int, success bool, message string, filename string) {
	w.WriteHeader(status)
	response := map[string]interface{}{
		"success": success,
		"message": message,
	}
	if filename != "" {
		response["filename"] = filename
	}
	json.NewEncoder(w).Encode(response)
}
