package main

import (
	"encoding/json"
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
		response.ReturnBadRequest(w, nil, "Could not decode JSON payload")
		return
	}

	r.Body.Close()

	err = validate.Struct(payload)

	if err != nil {
		response.ReturnBadRequest(w, nil, err.Error())
		return
	}

	date := payload.Date.Format(basicDateFormat)
	err = evaluateGameDayReport(date)

	if err != nil {
		response.ReturnBadRequest(w, nil, err.Error())
		return
	}

	response.ReturnStatusOK(w, payload.Date)
}

func doAPoll(w http.ResponseWriter, r *http.Request) {
	var payload forceEvaluationPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		response.ReturnBadRequest(w, nil, "Could not decode JSON payload")
		return
	}

	r.Body.Close()

	err = validate.Struct(payload)

	if err != nil {
		response.ReturnBadRequest(w, nil, err.Error())
		return
	}

	dateToday := payload.Date.Format(basicDateFormat)
	err = pollGames(dateToday, payload.Date.Add(24*time.Hour).Format(basicDateFormat))

	if err != nil { // TODO: handle errors
		response.ReturnInternalServerError(w, nil, err.Error())
		return
	}

	createReport := r.URL.Query().Get("createReport")
	if createReport == "true" {
		err = createGameDayReport(dateToday)

		if err != nil {
			response.ReturnInternalServerError(w, nil, err.Error())
			return
		}
	}

	response.ReturnStatusOK(w, payload.Date)
}
