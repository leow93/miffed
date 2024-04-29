package http_adapter

import (
	"encoding/json"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"net/http"
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

func NewServer(lift *lift.Lift, ps pubsub.PubSub) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /call", callLiftHandler(lift))
	mux.Handle("/socket", socketHandler(lift, ps))
	return mux
}
