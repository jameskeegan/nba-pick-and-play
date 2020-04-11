package main

import (
	"fmt"
	"nba-pick-and-play/pkg/rapid"
	"strconv"
	"time"
)

const (
	statusFinished = "Finished"
)

/*
	The public Rapid API does not allow for searching over a range of dates, instead just for a single, specified date.

	Because of this, to ensure you get all of the games for a given game night you have to do two calls as the date time
	is in UTC, meaning some games are before midnight (so on the correct game day) and some are after midnight (the day after the correct game day)
*/
func pollGames(dates ...string) error {
	log.Printf("Polling games for game date(s) %v...", dates)

	for _, date := range dates {
		err := pollGameDay(date)

		if err != nil {
			log.Error(err.Error())
			return err
		}
	}

	log.Printf("Successful poll for game date(s) %v...", dates)
	return nil
}

func pollGameDay(date string) error {
	res, err := rapidAPIClient.GetMatchesByDateRequest(date)

	if err != nil {
		return fmt.Errorf("could not evaluate matches for date %s: %s", date, err.Error())
	}

	for _, rapidGame := range res.ResponseWrapper.Games {
		game, err := rapidGameToGame(rapidGame) // convert to our mongo schema

		if err != nil {
			return fmt.Errorf("could not convert rapid game to game %s: %s", rapidGame.GameID, err.Error())
		}

		err = upsertMatch(*game)

		if err != nil {
			return fmt.Errorf("could not save game %d: %s", game.ID, err.Error())
		}
	}

	return nil
}

func rapidGameToGame(rapidGame rapid.NBAGame) (*game, error) {
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
		// game date id is for the previous day (e.g. game took place at 3am UTC = 8pm PST)
		game.GameDayID = rapidGame.StartTimeUTC.Add(-24 * time.Hour).Format(basicDateFormat)
	} else {
		// game took place on the date specified
		game.GameDayID = rapidGame.StartTimeUTC.Format(basicDateFormat)
	}

	return game, nil
}

func rapidTeamToTeam(rapidTeam rapid.NBATeam, status string) (*team, error) {
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
	Rapid dates are returned as UTC, which means some games are listed as being on the wrong "game day"
	e.g. if a game starts at 8pm in LA (PST) then it'll be listed as the following day at 3am (UTC)
*/
func isPreviousDayGame(date time.Time) bool {
	return date.Hour() < 12 // before/after noon can determine the day
}
