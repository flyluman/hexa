package rest

import (
	"encoding/json"
	"net/http"
)

func (h *RestHandler) GetWhoAmI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://flyluman.github.io")

	metaIp, err := h.svc.IPFetcher(r.Context(), extractIP(r))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(metaIp); err != nil {
		h.logger.Error("failed to marshal meta ip", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
