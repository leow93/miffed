package httpadapter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/leow93/miffed-api/internal/lift"
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
	Floor        int `json:"floor"`
	FloorDelayMs int `json:"floor_delay_ms"`
}

type createLiftRes struct {
	Id    lift.LiftId `json:"id"`
	Floor int         `json:"floor"`
}

func createLiftHandler(svc *lift.LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body createLiftReq
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&body)
		if err != nil && err != io.EOF {
			errResponse(w, 400, err)
			return
		}

		lift, err := svc.AddLift(lift.LiftConfig{Floor: body.Floor, FloorDelayMs: body.FloorDelayMs})
		if err != nil {
			errResponse(w, 500, err)
			return
		}

		okResponse(w, 201, createLiftRes{Id: lift.Id, Floor: lift.Floor})
	})
}

type callLiftReq struct {
	Floor int `json:"floor"`
}

func callLiftHandler(svc *lift.LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body createLiftReq
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&body)
		if err != nil && err != io.EOF {
			errResponse(w, 400, err)
			return
		}

		id, err := lift.ParseLiftId(r.PathValue("id"))
		if err != nil {
			errResponse(w, 404, lift.ErrLiftNotFound)
			return
		}

		if err = svc.CallLift(r.Context(), id, body.Floor); err != nil {
			if errors.Is(err, lift.ErrLiftNotFound) {
				errResponse(w, 404, lift.ErrLiftNotFound)
				return
			}
			errResponse(w, 500, err)
			return
		}
		okResponse(w, 201, struct{}{})
	})
}

type getLiftRes struct {
	Id    lift.LiftId `json:"id"`
	Floor int         `json:"floor"`
}

func getLiftsHandler(svc *lift.LiftService) http.Handler {
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

func getLiftHandler(svc *lift.LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := lift.ParseLiftId(r.PathValue("id"))
		if err != nil {
			errResponse(w, 404, lift.ErrLiftNotFound)
			return
		}
		if l, err := svc.GetLift(r.Context(), id); err != nil {
			statusCode := 500
			if errors.Is(err, lift.ErrLiftNotFound) {
				statusCode = 404
			}
			errResponse(w, statusCode, err)
		} else {
			okResponse(w, 200, getLiftRes{Id: l.Id, Floor: l.Floor})
		}
	})
}

func NewController(mux *http.ServeMux, svc *lift.LiftService) *http.ServeMux {
	mux.Handle("POST /lift", createLiftHandler(svc))
	mux.Handle("GET /lift", getLiftsHandler(svc))
	mux.Handle("GET /lift/{id}", getLiftHandler(svc))
	mux.Handle("POST /lift/{id}/call", callLiftHandler(svc))
	return mux
}
