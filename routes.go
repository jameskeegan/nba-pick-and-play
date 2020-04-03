package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"nba-pick-and-play/response"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type (
	picksPayload struct {
		GameDayID string          `json:"gameDayId"`
		Picks     map[int64]int64 `json:"picks"` // game id -> winner
	}
)

func getGameDayReport(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")

	if date == "" { // get date as the current date
		date = getCurrentGameDay(clockClient.now())
	}

	gameDayReport, err := findGameDayReportByID(date)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.ReturnNotFound(w, nil, fmt.Sprintf("could not find game day for date %s", date))
			return
		}

		response.ReturnInternalServerError(w, nil, err.Error())
		return
	}

	response.ReturnStatusOK(w, gameDayReport)
}

func makePicks(w http.ResponseWriter, r *http.Request) {
	var payload picksPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		response.ReturnBadRequest(w, nil, "could not decode json payload")
		return
	}

	picks, err := verifyPicks(payload.GameDayID, payload.Picks)

	if err != nil {
		response.ReturnBadRequest(w, nil, err.Error())
		return
	}

	// verified and legit so save them
	gameDayPicks := gameDayPicks{
		UserID:    12345, // TODO: user logic
		GameDayID: payload.GameDayID,
		Picks:     picks,
		Evaluated: false,
		Date:      clockClient.now(),
	}

	err = upsertGameDayPicks(gameDayPicks)

	if err != nil {
		response.ReturnInternalServerError(w, nil, err.Error())
		return
	}

	response.ReturnStatusOK(w, nil)
}

func getCurrentGameDay(date time.Time) string {
	timeNow := clockClient.now()

	// game day rolls over at 9am
	if clockClient.now().Hour() < 9 {
		return timeNow.Add(-24 * time.Hour).Format(basicDateFormat)
	}

	return timeNow.Format(basicDateFormat)
}
