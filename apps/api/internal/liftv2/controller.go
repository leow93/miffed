package liftv2

import (
	"context"
	"encoding/json"
	"net/http"
)

type errorResponse struct {
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
	json.NewEncoder(w).Encode(errorResponse{
		Code:  status,
		Error: &errMessage,
	})
}

func createLiftHandler(svc *LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := svc.AddLift(context.TODO(), AddLift{Id: NewLiftId(), Floor: 0})
		if err != nil {
			errResponse(w, 500, err)
			return
		}
		okResponse(w, 201, struct{ status string }{status: "OK"})
	})
}

func getLiftsHandler(rm *LiftReadModel) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lifts := rm.Query()
		okResponse(w, 200, lifts)
	})
}

func getLiftHandler(svc *LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := ParseLiftId(r.PathValue("id"))
		if err != nil {
			errResponse(w, 400, err)
			return
		}
		lift, err := svc.GetLift(context.TODO(), id)
		if err != nil {
			errResponse(w, 500, err)
			return
		}
		okResponse(w, 200, lift)
	})
}

func NewController(mux *http.ServeMux, svc *LiftService, readModel *LiftReadModel) *http.ServeMux {
	mux.Handle("POST /lift", createLiftHandler(svc))
	mux.Handle("GET /lift", getLiftsHandler(readModel))
	mux.Handle("GET /lift/{id}", getLiftHandler(svc))
	return mux
}
