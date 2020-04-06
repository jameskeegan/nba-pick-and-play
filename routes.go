package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nba-pick-and-play/config"
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

const (
	genericError = "Something went wrong, speak to Keegan."
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

		log.Printf("ERROR: %s", err.Error())
		response.ReturnInternalServerError(w, nil, genericError)
		return
	}

	response.ReturnStatusOK(w, gameDayReport)
}

func getGameDayResultsReport(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")

	if date == "" { // defaults to yesterday's results
		date = getCurrentGameDay(clockClient.now().Add(-24 * time.Hour))
	}

	resultsReport, err := findGameDayResultsReportByID(date)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.ReturnNotFound(w, nil, fmt.Sprintf("could not find game day for date %s", date))
			return
		}

		log.Printf("ERROR: %s", err.Error())
		response.ReturnInternalServerError(w, nil, genericError)
		return
	}

	response.ReturnStatusOK(w, resultsReport)
}

func getLeaderboard(w http.ResponseWriter, r *http.Request) {
	season := r.URL.Query().Get("season")

	if season == "" { // defaults to the current season
		season = config.Config.Rapid.Season
	}

	leaderboard, err := findLeaderboardByID(season)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.ReturnNotFound(w, nil, fmt.Sprintf("could not find leaderboard for season %s", season))
			return
		}

		log.Printf("ERROR: %s", err.Error())
		response.ReturnInternalServerError(w, nil, genericError)
		return
	}

	response.ReturnStatusOK(w, leaderboard)
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
		log.Printf("ERROR: %s", err.Error())
		response.ReturnInternalServerError(w, nil, genericError)
		return
	}

	response.ReturnStatusOK(w, nil)
}
