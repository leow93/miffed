package liftv3

import (
	"encoding/json"
	"errors"
	"io"
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

type createLiftReq struct {
	Floor int `json:"floor"`
}

type createLiftRes struct {
	Id    LiftId `json:"id"`
	Floor int    `json:"floor"`
}

func createLiftHandler(svc *LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body createLiftReq
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&body)
		if err != nil && err != io.EOF {
			errResponse(w, 400, err)
			return
		}

		lift, err := svc.AddLift(LiftConfig{Floor: body.Floor})
		if err != nil {
			errResponse(w, 500, err)
			return
		}

		okResponse(w, 201, createLiftRes{Id: lift.Id, Floor: lift.Floor})
	})
}

type getLiftRes struct {
	Id    LiftId `json:"id"`
	Floor int    `json:"floor"`
}

func getLiftsHandler(svc *LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if lifts, err := svc.GetLifts(r.Context()); err != nil {
			errResponse(w, 500, err)
		} else {
			body := make([]getLiftRes, len(lifts))
			for i, lift := range lifts {
				body[i] = getLiftRes{
					Id:    lift.Id,
					Floor: lift.Floor,
				}
			}
			okResponse(w, 200, body)
		}
	})
}

func getLiftHandler(svc *LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := ParseLiftId(r.PathValue("id"))
		if err != nil {
			errResponse(w, 404, liftNotFoundErr)
			return
		}
		if lift, err := svc.GetLift(r.Context(), id); err != nil {
			statusCode := 500
			if errors.Is(err, liftNotFoundErr) {
				statusCode = 404
			}
			errResponse(w, statusCode, err)
		} else {
			okResponse(w, 200, getLiftRes{Id: lift.Id, Floor: lift.Floor})
		}
	})
}

func NewController(mux *http.ServeMux, svc *LiftService) *http.ServeMux {
	mux.Handle("POST /lift", createLiftHandler(svc))
	mux.Handle("GET /lift", getLiftsHandler(svc))
	mux.Handle("GET /lift/{id}", getLiftHandler(svc))
	return mux
}
