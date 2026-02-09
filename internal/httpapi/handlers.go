package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/restream/reindexer"

	"reindexer-service/internal/dto"
	"reindexer-service/internal/model"
	"reindexer-service/internal/service"
)

type Handler struct {
	svc *service.DocumentService
}

func NewHandler(s *service.DocumentService) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) Register(mux *http.ServeMux) {

	mux.HandleFunc("/documents", h.handleDocuments)

	mux.HandleFunc("/documents/", h.handleDocumentByID)
}

func (h *Handler) handleDocuments(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/documents/" {
		http.Redirect(w, r, "/documents", http.StatusTemporaryRedirect)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		methodNotAllowed(w, []string{http.MethodGet, http.MethodPost})
	}
}

func (h *Handler) handleDocumentByID(w http.ResponseWriter, r *http.Request) {
	idPart := strings.TrimPrefix(r.URL.Path, "/documents/")
	if idPart == "" {
		http.Redirect(w, r, "/documents", http.StatusPermanentRedirect)
		return
	}

	if strings.Contains(idPart, "/") {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		methodNotAllowed(w, []string{http.MethodGet, http.MethodPut, http.MethodDelete})
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var doc model.Document
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	doc.ID = 0

	created, err := h.svc.Create(r.Context(), &doc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, id int64) {
	dtoDoc, hit, err := h.svc.GetDTO(r.Context(), id)
	if err != nil {
		if errors.Is(err, reindexer.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	if hit {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}

	writeJSON(w, http.StatusOK, dtoDoc)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, id int64) {
	defer r.Body.Close()

	var doc model.Document
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&doc); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	doc.ID = id

	if err := h.svc.Update(r.Context(), &doc); err != nil {
		if errors.Is(err, reindexer.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, reindexer.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "error limit"})
			return
		}

		limit = n

	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "error offset"})
			return
		}
		offset = n

	}

	res, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	items := make([]dto.Document, len(res.Items))

	var wg sync.WaitGroup
	wg.Add(len(res.Items))

	for i := range res.Items {
		i := i
		document := res.Items[i]

		go func() {
			defer wg.Done()
			items[i] = service.ToDocumentDTO(document)
		}()
	}

	wg.Wait()

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"Total":  res.Total,
		"Limit":  res.Limit,
		"Offset": res.Offset,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func methodNotAllowed(w http.ResponseWriter, allow []string) {
	w.Header().Set("Allow", strings.Join(allow, ", "))
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
}
