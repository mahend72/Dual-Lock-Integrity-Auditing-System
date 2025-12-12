package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type UploadResponse struct {
	FileID string `json:"fileId"`
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "uid required", http.StatusBadRequest)
		return
	}

	// In real code: read file into memory, encrypt, chunk, compute tags, post to storage-node and Fabric
	tmp, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	// Example: compute a fake fileId as base64 of first 16 bytes
	fid := base64.StdEncoding.EncodeToString(tmp[:min(16, len(tmp))])

	// (TODO: real SAM-BCADA logic + chaincode calls here)

	log.Printf("Received file for UID=%s, bytes=%d, FID=%s\n", uid, len(tmp), fid)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(UploadResponse{FileID: fid})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	http.HandleFunc("/upload", handleUpload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Coordinator listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
