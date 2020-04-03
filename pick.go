package main

import (
	"fmt"
)

func evaluatePicks(report gameDayReport, date string) error {
	pickReports, err := findPickReportsByGameDayID(date)

	if err != nil {
		return err
	}

	for _, pickReport := range pickReports {
		updatedPicksReport := evaluateUserPicks(report, pickReport)
		upsertGameDayPicks(updatedPicksReport)
	}

	return nil
}

func evaluateUserPicks(report gameDayReport, picksReport gameDayPicks) gameDayPicks {
	if picksReport.Evaluated {
		return picksReport
	}

	var score int64
	for gameID, pick := range picksReport.Picks {
		if pick.SelectionID == report.Games[gameID].WinnerID {
			pick.Status = "CORRECT"
			score++
		} else {
			pick.Status = "INCORRECT"
		}

		picksReport.Picks[gameID] = pick
	}

	picksReport.Score = score
	picksReport.Evaluated = true
	return picksReport
}

func verifyPicks(gameDate string, userPicks map[int64]int64) (map[int64]pick, error) {
	// get the game day report
	report, err := findGameDayReportByID(gameDate)

	if err != nil {
		return nil, err
	}

	if report.Deadline.Before(clockClient.now()) {
		return nil, fmt.Errorf("missed deadline: %v", report.Deadline) // TODO: turn into error struct?
	}

	// create a map with all possible picks
	picks := make(map[int64]pick)
	for gameID := range report.Games {
		picks[gameID] = pick{}
	}

	for gameID, userPick := range userPicks {
		_, ok := picks[gameID]

		if !ok {
			return nil, fmt.Errorf("game with id %d is not being played on this game day", gameID)
		}

		game := report.Games[gameID]

		if game.HomeTeam.ID != userPick && game.AwayTeam.ID != userPick {
			return nil, fmt.Errorf("team %d is not playing in the game %d", userPick, gameID)
		}

		picks[gameID] = pick{
			SelectionID: userPick,
			Status:      "PENDING",
		}
	}

	return picks, nil
}
