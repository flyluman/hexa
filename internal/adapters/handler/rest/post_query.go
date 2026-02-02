package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"portfolio/internal/core/domain"
)

func (h *RestHandler) PostQuery(w http.ResponseWriter, r *http.Request) {
	var query domain.Query
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.logger.Error("failed to unmarshal query params", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logs, err := h.svc.PostQuery(r.Context(), &query)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrIncorrectPassword):
			h.logger.Error("unauthorized request to the Server")
			http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(logs); err != nil {
		h.logger.Error("failed to marshal logs", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
