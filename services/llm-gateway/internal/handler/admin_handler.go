package handler

import (
	"encoding/json"
	"log"
	"net/http"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/service"
)

// TODO handler exist but do we serve http on llm-gateway ? later on for admin dashboard or separate for APIs
// AdminHandler handles HTTP admin requests for the LLM service
type AdminHandler struct {
	llmService *service.LLMServiceImpl
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(llmService *service.LLMServiceImpl) *AdminHandler {
	return &AdminHandler{
		llmService: llmService,
	}
}

// RegisterRoutes registers all admin HTTP routes
func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/collections", h.handleCollections)
	mux.HandleFunc("/admin/ingest-document", h.handleIngestDocument)
	mux.HandleFunc("/health", h.handleHealth)
}

// handleHealth is a simple health check endpoint
func (h *AdminHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleCollections handles GET/POST/DELETE requests for managing collections
func (h *AdminHandler) handleCollections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List collections
		resp, err := h.llmService.ListCollections(r.Context(), &pbllm.ListCollectionsRequest{})
		if err != nil {
			http.Error(w, "Failed to list collections: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	case http.MethodPost:
		// Create collection
		var req pbllm.CreateCollectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := h.llmService.CreateCollection(r.Context(), &req)
		if err != nil {
			http.Error(w, "Failed to create collection: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	case http.MethodDelete:
		// Delete collection
		var req pbllm.DeleteCollectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := h.llmService.DeleteCollection(r.Context(), &req)
		if err != nil {
			http.Error(w, "Failed to delete collection: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleIngestDocument handles POST requests for document ingestion
func (h *AdminHandler) handleIngestDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req pbllm.IngestDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.llmService.IngestDocument(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to ingest document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
