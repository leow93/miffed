package http_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/leow93/miffed-api/internal/lift"
)

type ErrResponse struct {
	Code  int     `json:"code"`
	Error *string `json:"error"`
}

func okResponse[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	errMessage := err.Error()
	json.NewEncoder(w).Encode(ErrResponse{
		Code:  status,
		Error: &errMessage,
	})
}

func NewServer(manager *lift.Manager) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /lift/{id}/call", callLiftHandler(manager))
	mux.Handle("/socket", socketHandler(manager))
	return mux
}
