package http_adapter

import (
	"encoding/json"
	"errors"
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

type addLiftReq struct {
	LowestFloor     *int `json:"lowest_floor"`
	HighestFloor    *int `json:"highest_floor"`
	CurrentFloor    *int `json:"current_floor"`
	FloorsPerSecond *int `json:"floors_per_second"`
	DoorCloseWaitMs *int `json:"door_close_wait_ms"`
}

type addLiftRes struct {
	Id lift.Id `json:"id"`
}

func validateReq(r *http.Request) (lift.NewLiftOpts, error) {
	var opts addLiftReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&opts)
	if err != nil {
		return lift.NewLiftOpts{}, err
	}
	if opts.DoorCloseWaitMs == nil || opts.FloorsPerSecond == nil || opts.CurrentFloor == nil || opts.HighestFloor == nil || opts.LowestFloor == nil {
		return lift.NewLiftOpts{}, errors.New("invalid request")
	}

	result := lift.NewLiftOpts{
		LowestFloor:     *opts.LowestFloor,
		HighestFloor:    *opts.HighestFloor,
		CurrentFloor:    *opts.CurrentFloor,
		FloorsPerSecond: *opts.FloorsPerSecond,
		DoorCloseWaitMs: *opts.DoorCloseWaitMs,
	}

	return result, nil
}

func addLiftHandler(manager *lift.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts, err := validateReq(r)
		if err != nil {
			errResponse(w, 400, err)
			return
		}

		l := manager.AddLift(opts)
		okResponse(w, 201, addLiftRes{Id: l.Id})
	})
}
