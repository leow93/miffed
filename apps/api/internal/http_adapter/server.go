package http_adapter

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/liftv2"
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

// V2
func createLiftHandler(svc *liftv2.LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := svc.AddLift(context.TODO(), liftv2.AddLift{Id: liftv2.NewLiftId(), Floor: 0})
		if err != nil {
			errResponse(w, 400, err)
			return
		}
		okResponse(w, 201, struct{ status string }{status: "OK"})
	})
}

func getLiftsHandler(rm *liftv2.LiftReadModel) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lifts := rm.Query()
		okResponse(w, 201, lifts)
	})
}

func NewServer(manager *lift.Manager, liftSvc *liftv2.LiftService, liftReadModel *liftv2.LiftReadModel) http.Handler {
	mux := http.NewServeMux()

	// v2
	mux = liftv2.NewController(mux, liftSvc, liftReadModel)
	return mux
}
