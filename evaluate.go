package main

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	basicDateFormat = "2006-01-02"
)

// for a given game day, get the correct picks and evaluate every pick
func evaluateGameDayReport(date string) error {
	report, err := findGameDayReportByID(date)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// one doesn't exist for whatever reason, so make one
			err = createGameDayReport(date)

			if err != nil {
				return err
			}
		}

		return err
	}

	games, err := findMatchesByGameDateID(date)

	if err != nil {
		return err
	}

	for _, game := range games {
		gameReport := report.Games[game.ID]
		gameReport.HomeTeam.Score = game.HomeTeam.Score
		gameReport.AwayTeam.Score = game.AwayTeam.Score
		gameReport.WinnerID = determineWinner(game.HomeTeam, game.AwayTeam)

		report.Games[game.ID] = gameReport
	}

	report.Evaluated = true

	if err := upsertGameDayReport(*report); err != nil {
		return err
	}

	return evaluatePicks(*report, date)
}

func createGameDayReport(date string) error {
	// get all matches for this game day, sorted with earliest first
	matches, err := findMatchesByGameDateID(date)

	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for date %s", date)
	}

	reportGames := make(map[int64]gameReport)
	for _, game := range matches {
		gameReport := gameReport{
			HomeTeam: game.HomeTeam,
			AwayTeam: game.AwayTeam,
			Venue:    game.Venue,
			Date:     game.StartDate,
		}

		reportGames[game.ID] = gameReport
	}

	report := gameDayReport{
		ID:        date,
		Games:     reportGames,
		Deadline:  matches[0].StartDate,
		Evaluated: false,
	}

	return upsertGameDayReport(report)
}
