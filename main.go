package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("csvfile")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if filepath.Ext(handler.Filename) != ".csv" {
		http.Error(w, "Only CSV files are allowed", http.StatusBadRequest)
		return
	}

	// Create destination file
	dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Uploaded: %s (%.2f KB)\n", handler.Filename, float64(handler.Size)/1024)
	
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true, "message": "File uploaded successfully", "filename": "%s"}`, handler.Filename)
}
