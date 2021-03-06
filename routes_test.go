package main

import (
	"bytes"
	"encoding/json"
	clockPkg "nba-pick-and-play/pkg/clock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	matchesResponse struct {
		Code      int           `json:"code"`
		Report    gameDayReport `json:"data,omitempty"`
		Error     string        `json:"error,omitempty"`
		CreatedAt string        `json:"createdAt"`
	}

	resultsResponse struct {
		Code      int            `json:"code"`
		Results   gameDayResults `json:"data,omitempty"`
		Error     string         `json:"error,omitempty"`
		CreatedAt string         `json:"createdAt"`
	}

	picksResponse struct {
		Code      int         `json:"code"`
		Data      interface{} `json:"data,omitempty"`
		Error     string      `json:"error,omitempty"`
		CreatedAt string      `json:"createdAt"`
	}
)

func TestGetGameDayReportSuccess(t *testing.T) {
	defer cleanDatabase(t)

	// poll matches, create a report for the day
	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	_, err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/games", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getGameDayReport)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response matchesResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.Nil(t, err)

	assert.Equal(t, "2020-01-18", response.Report.ID)
	assert.Equal(t, 11, len(response.Report.Games))
}

// current game day is still technically the previous night if the service is called pre 9am
func TestGetGameDayReportSuccessPreRollover(t *testing.T) {
	defer cleanDatabase(t)

	// poll matches, create a report for the day
	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	_, err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	clock = clockPkg.NewMockClock(time.Date(2020, time.January, 19, 8, 30, 0, 0, time.UTC))

	defer setDefaultMockClock()

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/games", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getGameDayReport)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response matchesResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.Nil(t, err)

	assert.Equal(t, "2020-01-18", response.Report.ID)
	assert.Equal(t, 11, len(response.Report.Games))
}

func TestGetGameDayResultsReportSuccess(t *testing.T) {
	defer cleanDatabase(t)

	// noon on the 19th Jan
	clock = clockPkg.NewMockClock(time.Date(2020, time.January, 19, 12, 0, 0, 0, time.UTC))

	defer setDefaultMockClock()

	scores := []result{
		result{
			UserID: 12345,
			Score:  9,
		},
		result{
			UserID: 67890,
			Score:  7,
		},
		result{
			UserID: 13579,
			Score:  2,
		},
	}

	err := upsertGameDayResults("2020-01-18", scores)
	assert.Nil(t, err)

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/results", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getGameDayResultsReport)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response resultsResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.Nil(t, err)

	assert.NotNil(t, response)
	assert.Equal(t, "2020-01-18", response.Results.ID)

	assert.Equal(t, 3, len(response.Results.UserScores))
}

func TestGetLeaderboard(t *testing.T) {
	defer cleanDatabase(t)

	lboard := leaderboard{
		ID: "2019",
		Standings: []leaderboardUser{
			{
				UserID: 12345,
				Score:  9,
			},
			{
				UserID: 67890,
				Score:  7,
			},
			{
				UserID: 13579,
				Score:  2,
			},
		},
		LastGameDayEvaluated: "2020-01-18",
	}

	err := upsertLeaderboard(lboard)
	assert.Nil(t, err)

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/leaderboards", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getLeaderboard)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMakePicksPastDeadline(t *testing.T) {
	defer cleanDatabase(t)

	// poll matches, create a report for the day
	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	_, err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	// missed deadline by half an hour (8:30pm is tip-off for first game)
	clock = clockPkg.NewMockClock(time.Date(2020, time.January, 18, 21, 0, 0, 0, time.UTC))

	defer setDefaultMockClock()

	payload := picksPayload{
		GameDayID: "2020-01-18",
		Picks: map[int64]int64{
			7015: 23,
		},
	}

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(payload)

	// call the endpoint
	req, err := http.NewRequest("POST", "/v1/user/picks", body)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(makePicks)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var response picksResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.Nil(t, err)

	assert.Equal(t, "missed deadline: 2020-01-18 20:30:00 +0000 UTC", response.Error)
}

func TestMakePicksWrongGame(t *testing.T) {
	defer cleanDatabase(t)

	// poll matches, create a report for the day
	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	_, err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	payload := picksPayload{
		GameDayID: "2020-01-18",
		Picks: map[int64]int64{
			12345: 23, // not a game being played on this date
		},
	}

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(payload)

	// call the endpoint
	req, err := http.NewRequest("POST", "/v1/user/picks", body)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(makePicks)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var response picksResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.Nil(t, err)

	assert.Equal(t, "game with id 12345 is not being played on this game day", response.Error)
}

func TestMakePicksSuccess(t *testing.T) {
	defer cleanDatabase(t)

	// poll matches, create a report for the day
	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	_, err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	payload := picksPayload{
		GameDayID: "2020-01-18",
		Picks: map[int64]int64{
			7015: 23,
			7016: 21,
			7017: 2,
			7018: 1,
			7019: 27,
			7020: 7,
			7021: 38,
			7022: 17,
			7023: 11,
			7024: 25,
			7025: 40,
		},
	}

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(payload)

	// call the endpoint
	req, err := http.NewRequest("POST", "/v1/user/picks", body)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(makePicks)
	handler.ServeHTTP(w, req)

	res := w.Result()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	picks, err := findPickReportsByGameDayID("2020-01-18")
	assert.Nil(t, err)

	assert.NotNil(t, picks)
	assert.Equal(t, 1, len(picks))

	pickReport := picks[0]
	assert.Equal(t, "2020-01-18", pickReport.GameDayID)
	assert.Equal(t, 11, len(pickReport.Picks))
	assert.False(t, pickReport.Evaluated)
	assert.Zero(t, pickReport.Score)

	for _, p := range pickReport.Picks {
		assert.NotZero(t, p.SelectionID)
		assert.Equal(t, "PENDING", p.Status)
	}
}
