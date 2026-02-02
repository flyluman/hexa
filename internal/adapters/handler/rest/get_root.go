package rest

import "net/http"

func (h *RestHandler) GetRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
