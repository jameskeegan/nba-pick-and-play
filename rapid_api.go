package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type (
	// response format for the rapid NBA API response
	rapidResponse struct {
		ResponseWrapper rapidResponseWrapper `json:"API"`
	}

	rapidResponseWrapper struct {
		Status  int         `json:"status"`
		Message string      `json:"message"`
		Results int         `json:"results"`
		Filters []string    `json:"filters"`
		Games   []rapidGame `json:"games"`
	}

	rapidGame struct {
		SeasonYear      string    `json:"seasonYear"`
		League          string    `json:"league"`
		GameID          string    `json:"gameId"`
		StartTimeUTC    time.Time `json:"startTimeUTC"`
		EndTimeUTC      string    `json:"endTimeUTC"`
		Arena           string    `json:"arena"`
		City            string    `json:"city"`
		Country         string    `json:"country"`
		Clock           string    `json:"clock"`
		GameDuration    string    `json:"gameDuration"`
		CurrentPeriod   string    `json:"currentPeriod"`
		Halftime        string    `json:"halftime"`
		EndOfPeriod     string    `json:"EndOfPeriod"`
		SeasonStage     string    `json:"seasonStage"`
		StatusShortGame string    `json:"statusShortGame"`
		StatusGame      string    `json:"statusGame"`
		VTeam           rapidTeam `json:"vTeam"`
		HTeam           rapidTeam `json:"hTeam"`
	}

	rapidTeam struct {
		TeamID    string     `json:"teamId"`
		ShortName string     `json:"shortName"`
		FullName  string     `json:"fullName"`
		NickName  string     `json:"nickName"`
		Logo      string     `json:"logo"`
		Score     rapidScore `json:"score"`
	}

	rapidScore struct {
		Points string `json:"points"`
	}

	// interface so that we can substitute in a mock for testing
	rapidAPIInterface interface {
		getMatchesByDateRequest(date string) (*rapidResponse, error)
	}

	baseRapidAPIClient struct {
		baseURL string
		apiKey  string
	}
)

func (c baseRapidAPIClient) getMatchesByDateRequest(date string) (*rapidResponse, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", c.baseURL+date, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("x-rapidapi-key", c.apiKey)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response rapidResponse
	json.NewDecoder(resp.Body).Decode(&response)

	return &response, nil
}

const (
	statusFinished = "Finished"
)

var (
	rapidAPIClient rapidAPIInterface
)

func pollGames(dates ...string) error {
	log.Printf("Polling games for date(s) %v...", dates)

	for _, date := range dates {
		err := saveMatchDay(date)

		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	log.Printf("Successfully polled for games")
	return nil
}

func saveMatchDay(date string) error {
	res, err := rapidAPIClient.getMatchesByDateRequest(date)

	if err != nil {
		return fmt.Errorf("ERROR: Could not evaluate matches for date %s: %s", date, err.Error())
	}

	for _, rapidGame := range res.ResponseWrapper.Games {
		game, err := rapidGameToGame(rapidGame) // convert to our mongo schema

		if err != nil {
			return fmt.Errorf("ERROR: could not conver from rapid game to game %s: %s", rapidGame.GameID, err.Error())
		}

		err = upsertMatch(*game)

		if err != nil {
			return fmt.Errorf("ERROR: could not save game %d: %s", game.ID, err.Error())
		}
	}

	return nil
}

func rapidGameToGame(rapidGame rapidGame) (*game, error) {
	matchID, err := strconv.ParseInt(rapidGame.GameID, 10, 64)

	if err != nil {
		return nil, err
	}

	homeTeam, err := rapidTeamToTeam(rapidGame.HTeam, rapidGame.StatusGame)

	if err != nil {
		return nil, err
	}

	awayTeam, err := rapidTeamToTeam(rapidGame.VTeam, rapidGame.StatusGame)

	if err != nil {
		return nil, err
	}

	// TODO: parse date irregularities
	game := &game{
		ID:          matchID,
		SeasonID:    rapidGame.SeasonYear,
		Status:      rapidGame.StatusGame,
		SeasonStage: rapidGame.SeasonStage,
		StartDate:   rapidGame.StartTimeUTC,
		HomeTeam:    *homeTeam,
		AwayTeam:    *awayTeam,
		Venue: venue{
			Name:    rapidGame.Arena,
			City:    rapidGame.City,
			Country: rapidGame.Country,
		},
	}

	if game.Status == statusFinished {
		game.WinnerID = determineWinner(game.HomeTeam, game.AwayTeam)
	}

	// work out the game date id
	if isPreviousDayGame(rapidGame.StartTimeUTC) {
		// game date id is for the previous day (e.g. game took place at 3am UTC - 8pm PST)
		game.GameDayID = rapidGame.StartTimeUTC.Add(-24 * time.Hour).Format(basicDateFormat)
	} else {
		// game took place on the date specified
		game.GameDayID = rapidGame.StartTimeUTC.Format(basicDateFormat)
	}

	return game, nil
}

func rapidTeamToTeam(rapidTeam rapidTeam, status string) (*team, error) {
	id, err := strconv.ParseInt(rapidTeam.TeamID, 10, 64)

	if err != nil {
		return nil, err
	}

	team := team{
		ID:       id,
		Name:     rapidTeam.FullName,
		Nickname: rapidTeam.NickName,
		Logo:     rapidTeam.Logo,
	}

	if status == statusFinished {
		points, err := strconv.ParseInt(rapidTeam.Score.Points, 10, 64)

		if err != nil {
			return nil, err
		}

		team.Score = points
	}

	return &team, nil
}

func determineWinner(home team, away team) int64 {
	if home.Score > away.Score {
		return home.ID
	}

	return away.ID
}

/*
	Rapid dates are returned as UTC, which means some games are listed as being on the wrong "game" day
	e.g. if a game starts at 8pm in LA (PST) then it'll be listed as the following day at 3am (UTC)
*/
func isPreviousDayGame(date time.Time) bool {
	return date.Hour() < 12 // before/after noon can determine the day
}
