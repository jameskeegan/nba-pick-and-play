package main

import (
	"encoding/json"
	"fmt"
	"nba-pick-and-play/response"
	"net/http"
	"time"
)

type (
	forceEvaluationPayload struct {
		Date time.Time `json:"date" validate:"required"`
	}
)

func forceEvaluation(w http.ResponseWriter, r *http.Request) {
	var payload forceEvaluationPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		response.ReturnBadRequest(w, fmt.Errorf("Could not decode JSON payload"))
		return
	}

	r.Body.Close()

	err = validate.Struct(payload)

	if err != nil {
		response.ReturnBadRequest(w, err.Error())
		return
	}

	err = evaluateMatches(payload.Date)

	if err != nil {
		response.ReturnBadRequest(w, err.Error())
		return
	}

	response.ReturnStatusOK(w, payload.Date)
}
