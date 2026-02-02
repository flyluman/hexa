package rest

import (
	"encoding/json"
	"net/http"
	"portfolio/internal/core/domain"
)

func (h *RestHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var msg domain.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.logger.Error("failed to unmarshal msg params", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg.RequestID = getRequestID(r.Context())

	if err := h.svc.PostMessage(r.Context(), &msg); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "https://flyluman.github.io", http.StatusSeeOther)
}
