package rest

import "net/http"

func (h *RestHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
