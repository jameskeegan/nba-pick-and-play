package main

import (
	"sort"
)

type result struct {
	UserID int64
	Score  int64
}

// get all the pick reports for that day, create a daily leaderboard
func createGameDayResults(date string) error {
	filter := make(filter)
	filter["evaluated"] = true

	pickReports, err := findPickReportsByGameDayID(date, filter)

	if err != nil {
		log.Errorf("when creating game day results: %s", err.Error())
		return err
	}

	var results []result
	for _, rep := range pickReports {
		results = append(results, result{
			UserID: rep.UserID,
			Score:  rep.Score,
		})
	}

	// sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	err = upsertGameDayResults(date, results)

	if err != nil {
		log.Errorf("when upserting game day results: %s", err.Error())
	}

	return err
}

// do a full update of the season's results
func updateLeaderboard(season string) error {
	userScores, err := aggregateUserScoresForSeason(season)

	if err != nil {
		log.Errorf("when creating leaderboard: %s", err.Error())
		return err
	}

	var users []leaderboardUser
	for _, user := range userScores {
		users = append(users, leaderboardUser{
			UserID: user.ID,
			Score:  user.Score,
		})
	}

	board := leaderboard{
		ID:        season,
		Standings: users,
	}

	err = upsertLeaderboard(board)
	return err
}
