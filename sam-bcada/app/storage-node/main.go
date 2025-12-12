package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type StoredBlock struct {
	UID        string
	FileID     string
	BlockIndex int
	Data       []byte
}

var (
	storeMu sync.Mutex
	store   []StoredBlock
)

type StoreRequest struct {
	UID    string `json:"uid"`
	FileID string `json:"fileId"`
	Blocks []struct {
		BlockIndex int    `json:"blockIndex"`
		Data       string `json:"data"` // base64
	} `json:"blocks"`
}

type GetBlocksRequest struct {
	FileID  string `json:"fileId"`
	Indices []int  `json:"indices"`
}

type GetBlocksResponse struct {
	Blocks []struct {
		BlockIndex int    `json:"blockIndex"`
		Data       string `json:"data"`
	} `json:"blocks"`
}

func handleStore(w http.ResponseWriter, r *http.Request) {
	var req StoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	for _, b := range req.Blocks {
		data, err := base64.StdEncoding.DecodeString(b.Data)
		if err != nil {
			http.Error(w, "bad base64", http.StatusBadRequest)
			return
		}
		store = append(store, StoredBlock{
			UID:        req.UID,
			FileID:     req.FileID,
			BlockIndex: b.BlockIndex,
			Data:       data,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	var req GetBlocksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	resp := GetBlocksResponse{}
	for _, sb := range store {
		if sb.FileID != req.FileID {
			continue
		}
		for _, idx := range req.Indices {
			if sb.BlockIndex == idx {
				resp.Blocks = append(resp.Blocks, struct {
					BlockIndex int    `json:"blockIndex"`
					Data       string `json:"data"`
				}{
					BlockIndex: idx,
					Data:       base64.StdEncoding.EncodeToString(sb.Data),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/store", handleStore)
	http.HandleFunc("/getBlocks", handleGetBlocks)

	log.Println("Storage node listening on :4000")
	log.Fatal(http.ListenAndServe(":4000", nil))
}
