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

type callLiftReq struct {
	Floor int `json:"floor"`
}

func callLiftHandler(svc *LiftService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body createLiftReq
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&body)
		if err != nil && err != io.EOF {
			errResponse(w, 400, err)
			return
		}

		id, err := ParseLiftId(r.PathValue("id"))
		if err != nil {
			errResponse(w, 404, errLiftNotFound)
			return
		}

		if err = svc.CallLift(r.Context(), id, body.Floor); err != nil {
			if errors.Is(err, errLiftNotFound) {
				errResponse(w, 404, errLiftNotFound)
				return
			}
			errResponse(w, 500, err)
			return
		}
		okResponse(w, 201, struct{}{})
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
			errResponse(w, 404, errLiftNotFound)
			return
		}
		if lift, err := svc.GetLift(r.Context(), id); err != nil {
			statusCode := 500
			if errors.Is(err, errLiftNotFound) {
				statusCode = 404
			}
			errResponse(w, statusCode, err)
		} else {
			okResponse(w, 200, getLiftRes{Id: lift.Id, Floor: lift.Floor})
		}
	})
}

func openCorsPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obviously don't do this in prod
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func NewController(mux *http.ServeMux, svc *LiftService) *http.ServeMux {
	mux.Handle("POST /lift", openCorsPolicy(createLiftHandler(svc)))
	mux.Handle("GET /lift", openCorsPolicy(getLiftsHandler(svc)))
	mux.Handle("GET /lift/{id}", openCorsPolicy(getLiftHandler(svc)))
	mux.Handle("POST /lift/{id}/call", openCorsPolicy(callLiftHandler(svc)))
	return mux
}
