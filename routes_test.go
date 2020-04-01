package main

import (
	"encoding/json"
	"fmt"
	"nba-pick-and-play/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type (
	matchesResponse struct {
		Code      int     `json:"code"`
		Matches   []match `json:"data,omitempty"`
		Error     string  `json:"error,omitempty"`
		CreatedAt string  `json:"createdAt"`
	}
)

func init() {
	config.LoadConfig("config/config_test.toml")

	setupDatabase()

	// mock API to return the json test files data as responses
	rapidAPIClient = mockRapidAPIClient{}
}

func TestGetMatchesByDateNotFound(t *testing.T) {
	defer cleanDatabase(t)

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/matches", nil)

	if err != nil {
		t.Errorf(err.Error())
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getMatchesByDate)
	handler.ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", res.StatusCode)
	}

	var response matchesResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		t.Errorf("Unexpected error when decoding response: %s", err.Error())
	}

	dateToday := time.Now().Format(basicDateFormat)
	expectedMessage := fmt.Sprintf("no matches found for game date '%s'", dateToday)

	if response.Error != expectedMessage {
		t.Errorf("Expected %q, got %q", expectedMessage, response.Error)
	}
}

func TestGetMatchesByDateSpecificDateNotFound(t *testing.T) {
	defer cleanDatabase(t)

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/matches?date=2020-01-18", nil)

	if err != nil {
		t.Errorf(err.Error())
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getMatchesByDate)
	handler.ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", res.StatusCode)
	}

	var response matchesResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		t.Errorf("Unexpected error when decoding response: %s", err.Error())
	}

	if response.Error != "no matches found for game date '2020-01-18'" {
		t.Errorf("Expected \"no matches found for game date '2020-01-18'\", got %q", response.Error)
	}
}

func TestGetMatchesByDateSuccess(t *testing.T) {
	defer cleanDatabase(t)

	// load the mock data into the db
	date := time.Date(2020, time.January, 18, 9, 0, 0, 0, time.UTC) // 9am, Jan 18th 2020
	err := evaluateMatches(date)

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	// call the endpoint
	req, err := http.NewRequest("GET", "/v1/user/matches?date=2020-01-18", nil)

	if err != nil {
		t.Errorf(err.Error())
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(getMatchesByDate)
	handler.ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", res.StatusCode)
	}

	var response matchesResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		t.Errorf("Unexpected error when decoding response: %s", err.Error())
	}

	if len(response.Matches) != 11 {
		t.Errorf("Expected 11 matches, got %d", len(response.Matches))
	}

	// first match in sorted list of matches
	pelsVsClippers := response.Matches[0]

	if pelsVsClippers.GameDateID != "2020-01-18" {
		t.Errorf("Expected game date \"2020-01-18\", got %q", pelsVsClippers.GameDateID)
	}

	if pelsVsClippers.ID != 7015 {
		t.Errorf("Expected id '7015', got '%d'", pelsVsClippers.ID)
	}
}
