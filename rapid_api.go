package main

import (
	"encoding/json"
	"nba-pick-and-play/config"
	"net/http"
	"strconv"
	"time"
)

type (
	// interface so that we can substitute in a mock for testing
	rapidAPIInterface interface {
		getMatchesByDateRequest(date string) (*rapidResponse, error)
	}

	baseRapidAPIClient struct{}
)

var (
	rapidAPIClient rapidAPIInterface
)

func (baseRapidAPIClient) getMatchesByDateRequest(date string) (*rapidResponse, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", config.Config.Rapid.BaseURL+date, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("x-rapidapi-key", config.Config.Rapid.APIKey)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response rapidResponse
	json.NewDecoder(resp.Body).Decode(&response)

	return &response, nil
}

type (
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
		SeasonYear      string     `json:"seasonYear"`
		League          string     `json:"league"`
		GameID          string     `json:"gameId"`
		StartTimeUTC    time.Time  `json:"startTimeUTC"`
		EndTimeUTC      *time.Time `json:"endTimeUTC"`
		Arena           string     `json:"arena"`
		City            string     `json:"city"`
		Country         string     `json:"country"`
		Clock           string     `json:"clock"`
		GameDuration    string     `json:"gameDuration"`
		CurrentPeriod   string     `json:"currentPeriod"`
		Halftime        string     `json:"halftime"`
		EndOfPeriod     string     `json:"EndOfPeriod"`
		SeasonStage     string     `json:"seasonStage"`
		StatusShortGame string     `json:"statusShortGame"`
		StatusGame      string     `json:"statusGame"`
		VTeam           rapidTeam  `json:"vTeam"`
		HTeam           rapidTeam  `json:"hTeam"`
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
)

const (
	statusFinished = "Finished"
)

func gameToMatch(game rapidGame) (*match, error) {
	matchID, err := strconv.ParseInt(game.GameID, 10, 64)

	if err != nil {
		return nil, err
	}

	homeTeam, err := rapidTeamToTeam(game.HTeam, game.StatusGame)

	if err != nil {
		return nil, err
	}

	awayTeam, err := rapidTeamToTeam(game.VTeam, game.StatusGame)

	if err != nil {
		return nil, err
	}

	// TODO: parse date irregularities
	match := &match{
		ID:          matchID,
		SeasonID:    game.SeasonYear,
		Status:      game.StatusGame,
		SeasonStage: game.SeasonStage,
		StartDate:   game.StartTimeUTC,
		HomeTeam:    *homeTeam,
		AwayTeam:    *awayTeam,
		Venue: venue{
			Name:    game.Arena,
			City:    game.City,
			Country: game.Country,
		},
	}

	if match.Status == statusFinished {
		match.WinnerID = determineWinner(match.HomeTeam, match.AwayTeam)
	}

	// work out the game date id
	if isPreviousDayGame(game.StartTimeUTC) {
		// game date id is for the previous day (e.g. game took place at 3am UTC - 8pm PST)
		match.GameDateID = game.StartTimeUTC.Add(-24 * time.Hour).Format(basicDateFormat)
	} else {
		// game took place on the date specified
		match.GameDateID = game.StartTimeUTC.Format(basicDateFormat)
	}

	return match, nil
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
	Rapid dates are returned as UTC, which means some games are listed as being on the wrong day
	e.g. if a game starts at 8pm in LA (PST) then it'll be listed as the following day at 3am (UTC)
*/
func isPreviousDayGame(date time.Time) bool {
	return date.Hour() < 12
}
