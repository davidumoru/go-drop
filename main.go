package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileResponse struct {
	Status  string `json:"status"`
	File    string `json:"file"`
	Message string `json:"message,omitempty"`
	Path    string `json:"path,omitempty"`
}

type File struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	maxUploadSize = 100 << 20
	uploadDir     = "shared"
	serverPort    = 8080
)

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		sendJSONError(w, "File too large (maximum 100MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		sendJSONError(w, "Error reading file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := sanitizeFilename(header.Filename)
	if filename == "" {
		sendJSONError(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		sendJSONError(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		sendJSONError(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	host := r.Host
	fileURL := fmt.Sprintf("http://%s/shared/%s", host, filename)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := FileResponse{
		Status:  "success",
		File:    filename,
		Message: "File uploaded successfully",
		Path:    fileURL,
	}
	json.NewEncoder(w).Encode(response)
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		sendJSONError(w, "Error reading directory", http.StatusInternalServerError)
		return
	}

	fileList := []File{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fileInfo, err := f.Info()
		if err != nil {
			continue
		}

		fileURL := fmt.Sprintf("http://%s/shared/%s", r.Host, f.Name())
		fileList = append(fileList, File{
			Name:      f.Name(),
			Size:      fileInfo.Size(),
			URL:       fileURL,
			Timestamp: fileInfo.ModTime(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileList)
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	filename := filepath.Base(r.URL.Path)
	filePath := filepath.Join(uploadDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filePath)
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "public/index.html")
		return
	}

	http.ServeFile(w, r, filepath.Join("public", r.URL.Path))
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := FileResponse{
		Status:  "error",
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

func sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	if filename == "" || filename == "." {
		return ""
	}
	return filename
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

func main() {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.Mkdir(uploadDir, 0755)
		if err != nil {
			log.Fatal("Failed to create upload directory:", err)
		}
	}

	if _, err := os.Stat("public"); os.IsNotExist(err) {
		err = os.Mkdir("public", 0755)
		if err != nil {
			log.Fatal("Failed to create public directory:", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveUI)
	mux.HandleFunc("/upload", handleFileUpload)
	mux.HandleFunc("/shared/", serveFiles)
	mux.HandleFunc("/api/files", listFiles)

	localIP := getLocalIP()
	localURL := fmt.Sprintf("http://%s:%d", localIP, serverPort)
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", localIP, serverPort),
		Handler: mux,
	}

	fmt.Println("🚀 Go Drop started successfully")
	fmt.Printf("📡 Local access: http://localhost:%d\n", serverPort)
	fmt.Printf("🌍 Network access: %s\n", localURL)
	fmt.Printf("📂 Files are stored in the '%s' directory\n", uploadDir)
	fmt.Println("Press Ctrl+C to stop the server")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server error:", err)
	}
}
