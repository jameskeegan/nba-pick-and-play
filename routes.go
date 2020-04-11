package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"nba-pick-and-play/config"
	"nba-pick-and-play/pkg/response"
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

const genericError = "Something went wrong, speak to Keegan."

func getGameDayReport(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")

	if date == "" { // get date as the current date
		date = getCurrentGameDay(clock.Now())
	}

	gameDayReport, err := findGameDayReportByID(date)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.ReturnError(w, http.StatusNotFound, fmt.Sprintf("could not find game day for date %s", date))
			return
		}

		log.Error(err.Error())
		response.ReturnError(w, http.StatusInternalServerError, genericError)
		return
	}

	response.ReturnSuccess(w, http.StatusOK, gameDayReport)
}

func getGameDayResultsReport(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")

	if date == "" { // defaults to yesterday's results
		date = getCurrentGameDay(clock.Now().Add(-24 * time.Hour))
	}

	resultsReport, err := findGameDayResultsReportByID(date)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.ReturnError(w, http.StatusNotFound, fmt.Sprintf("could not find game day for date %s", date))
			return
		}

		log.Error(err.Error())
		response.ReturnError(w, http.StatusInternalServerError, genericError)
		return
	}

	response.ReturnSuccess(w, http.StatusOK, resultsReport)
}

func getLeaderboard(w http.ResponseWriter, r *http.Request) {
	season := r.URL.Query().Get("season")

	if season == "" { // defaults to the current season
		season = config.Config.Rapid.Season
	}

	leaderboard, err := findLeaderboardByID(season)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.ReturnError(w, http.StatusNotFound, fmt.Sprintf("could not find leaderboard for season %s", season))
			return
		}

		log.Error(err.Error())
		response.ReturnError(w, http.StatusInternalServerError, genericError)
		return
	}

	response.ReturnSuccess(w, http.StatusOK, leaderboard)
}

func makePicks(w http.ResponseWriter, r *http.Request) {
	var payload picksPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		response.ReturnError(w, http.StatusBadRequest, "could not decode json payload")
		return
	}

	picks, err := verifyPicks(payload.GameDayID, payload.Picks)

	if err != nil {
		response.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	// verified and legit so save them
	gameDayPicks := gameDayPicks{
		UserID:    12345, // TODO: user logic
		GameDayID: payload.GameDayID,
		SeasonID:  config.Config.Rapid.Season,
		Picks:     picks,
		Evaluated: false,
		Date:      clock.Now(),
	}

	err = upsertGameDayPicks(gameDayPicks)

	if err != nil {
		log.Error(err.Error())
		response.ReturnError(w, http.StatusInternalServerError, genericError)
		return
	}

	response.ReturnSuccess(w, http.StatusCreated, nil)
}
