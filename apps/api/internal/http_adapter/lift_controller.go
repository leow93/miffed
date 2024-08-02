package http_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/leow93/miffed-api/internal/lift"
)

type callLiftReq struct {
	Floor int `json:"floor"`
}

type callLiftRes struct {
	Floor int `json:"floor"`
}

func callLiftHandler(manager *lift.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := lift.ParseId(r.PathValue("id"))
		if err != nil {
			errResponse(w, http.StatusBadRequest, err)
			return
		}
		body := callLiftReq{}
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			errResponse(w, http.StatusBadRequest, err)
			return
		}
		floor := body.Floor
		called := manager.CallLift(id, floor)
		var status int
		if called {
			status = http.StatusCreated
		} else {
			status = http.StatusOK
		}
		okResponse(w, status, callLiftRes{floor})
	})
}

type addLiftReq struct{}

type addLiftRes struct{}

func addLiftHandler(manager *lift.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okResponse(w, 201, addLiftRes{})
	})
}
