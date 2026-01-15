package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/browser"

	"qa-script/config"
	"qa-script/output"
	"qa-script/processor"
)

//go:embed frontend/*
var frontendFS embed.FS

// App holds the application state
type App struct {
	locationsFile  string
	prioritiesFile string
	rulesFile      string // Optional custom qa_loc_rules.yaml
	tempDir        string
	server         *http.Server
	shutdownChan   chan struct{}
}

func main() {
	// Create temp directory for uploaded files
	tempDir, err := os.MkdirTemp("", "qa-grouper-*")
	if err != nil {
		log.Fatal("Failed to create temp directory:", err)
	}
	defer os.RemoveAll(tempDir)

	app := &App{
		tempDir:      tempDir,
		shutdownChan: make(chan struct{}),
	}

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Failed to find available port:", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Set up HTTP routes
	mux := http.NewServeMux()

	// Serve static files from embedded filesystem
	mux.HandleFunc("/", app.serveStatic)

	// API endpoints
	mux.HandleFunc("/api/upload-locations", app.handleUploadLocations)
	mux.HandleFunc("/api/upload-priorities", app.handleUploadPriorities)
	mux.HandleFunc("/api/upload-rules", app.handleUploadRules)
	mux.HandleFunc("/api/process", app.handleProcess)
	mux.HandleFunc("/api/reset", app.handleReset)
	mux.HandleFunc("/api/status", app.handleStatus)
	mux.HandleFunc("/api/shutdown", app.handleShutdown)
	mux.HandleFunc("/api/download", app.handleDownload)

	app.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server at http://127.0.0.1:%d", port)
		if err := app.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Open browser
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Printf("Opening browser at %s", url)
	if err := browser.OpenURL(url); err != nil {
		log.Printf("Failed to open browser: %v (please open %s manually)", err, url)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
	case <-app.shutdownChan:
		log.Println("Shutdown requested via API")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// serveStatic serves the embedded frontend files
func (a *App) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Read from embedded filesystem
	content, err := frontendFS.ReadFile("frontend" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set content type
	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}

	w.Write(content)
}

// handleUploadLocations handles uploading the locations CSV file
func (a *App) handleUploadLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32 MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		http.Error(w, "Locations file must be a CSV file", http.StatusBadRequest)
		return
	}

	// Save to temp directory
	destPath := filepath.Join(a.tempDir, header.Filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	a.locationsFile = destPath

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"filename": header.Filename,
	})
}

// handleUploadPriorities handles uploading the priorities XLSX file
func (a *App) handleUploadPriorities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32 MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		http.Error(w, "Priorities file must be an XLSX file", http.StatusBadRequest)
		return
	}

	// Save to temp directory
	destPath := filepath.Join(a.tempDir, header.Filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	a.prioritiesFile = destPath

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"filename": header.Filename,
	})
}

// handleUploadRules handles uploading an optional custom qa_loc_rules.yaml file
func (a *App) handleUploadRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 1 MB for YAML)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".yaml") && !strings.HasSuffix(strings.ToLower(header.Filename), ".yml") {
		http.Error(w, "Rules file must be a YAML file (.yaml or .yml)", http.StatusBadRequest)
		return
	}

	// Save to temp directory as qa_loc_rules.yaml
	destPath := filepath.Join(a.tempDir, "qa_loc_rules.yaml")
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	a.rulesFile = destPath

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"filename": header.Filename,
	})
}

// handleProcess processes the files and generates output
func (a *App) handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.locationsFile == "" {
		http.Error(w, "No locations file selected", http.StatusBadRequest)
		return
	}
	if a.prioritiesFile == "" {
		http.Error(w, "No priorities file selected", http.StatusBadRequest)
		return
	}

	// Use custom qa_loc_rules.yaml if uploaded, otherwise create default
	var rulesPath string
	if a.rulesFile != "" {
		rulesPath = a.rulesFile
	} else {
		var err error
		rulesPath, err = config.EnsureRulesFile(a.tempDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to ensure rules file: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Process the data
	result, err := processor.Process(a.locationsFile, a.prioritiesFile, rulesPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Create output data
	outputData := output.NewOutputData(
		result.TitleOrder,
		result.TitleGroupedLocations,
		result.PriorityLocations,
		result.ColumnGap,
		result.MaxRows,
	)

	// Generate output filename in temp directory
	outputPath := filepath.Join(a.tempDir, "location_priorities.xlsx")

	// Write XLSX output
	if err := output.WriteXLSX(outputPath, outputData); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write output: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"downloadUrl":  "/api/download?file=location_priorities.xlsx",
		"filename":     "location_priorities.xlsx",
	})
}

// handleDownload serves the output file for download
func (a *App) handleDownload(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "No file specified", http.StatusBadRequest)
		return
	}

	// Security: only allow downloading from temp directory
	filePath := filepath.Join(a.tempDir, filepath.Base(filename))
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	
	http.ServeFile(w, r, filePath)
}

// handleReset clears the selected files
func (a *App) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Remove uploaded files
	if a.locationsFile != "" {
		os.Remove(a.locationsFile)
	}
	if a.prioritiesFile != "" {
		os.Remove(a.prioritiesFile)
	}
	if a.rulesFile != "" {
		os.Remove(a.rulesFile)
	}

	a.locationsFile = ""
	a.prioritiesFile = ""
	a.rulesFile = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleStatus returns current file selection status
func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	locationsName := ""
	prioritiesName := ""
	rulesName := ""
	
	if a.locationsFile != "" {
		locationsName = filepath.Base(a.locationsFile)
	}
	if a.prioritiesFile != "" {
		prioritiesName = filepath.Base(a.prioritiesFile)
	}
	if a.rulesFile != "" {
		rulesName = filepath.Base(a.rulesFile)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"locationsFile":  locationsName,
		"prioritiesFile": prioritiesName,
		"rulesFile":      rulesName,
	})
}

// handleShutdown triggers server shutdown
func (a *App) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	// Trigger shutdown after response is sent
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(a.shutdownChan)
	}()
}
