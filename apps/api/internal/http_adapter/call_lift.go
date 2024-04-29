package http_adapter

import (
	"encoding/json"
	"github.com/leow93/miffed-api/internal/lift"
	"net/http"
)

type callLiftReq struct {
	Floor int `json:"floor"`
}

type callLiftRes struct {
	Floor int `json:"floor"`
}

func callLiftHandler(lift *lift.Lift) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := callLiftReq{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			errResponse(w, http.StatusBadRequest, err)
			return
		}
		floor := body.Floor
		called := lift.Call(floor)
		var status int
		if called {
			status = http.StatusCreated
		} else {
			status = http.StatusOK
		}
		okResponse(w, status, callLiftRes{floor})
	})
}
