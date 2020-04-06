package main

import (
	"errors"
	"log"
	"sort"

	"go.mongodb.org/mongo-driver/mongo"
)

type result struct {
	UserID int64
	Score  int64
}

// get all the pick reports for that day, create a daily leaderboard
func createGameDayResults(date string) {
	pickReports, err := findPickReportsByGameDayIDAndEvaluated(date, true)

	if err != nil {
		log.Printf("ERROR creating game day results: %s", err.Error())
		return
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
		log.Printf("ERROR upserting game day results: %s", err.Error())
	}

	err = updateLeaderboard("2019", date, results)

	if err != nil {
		log.Printf("ERROR updating the leaderboard: %s", err.Error())
	}
}

// update the overall season leaderboard with the new results
func updateLeaderboard(season string, gameDay string, results []result) error {
	board, err := findLeaderboardByID(season)

	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}

		board = &leaderboard{
			ID: season,
		}
	}

	if gameDay == board.LastGameDayEvaluated {
		return nil
	}

	standingsMap := make(map[int64]int64) // userID -> score
	for _, user := range board.Standings {
		standingsMap[user.UserID] = user.Score
	}

	for _, result := range results {
		score, ok := standingsMap[result.UserID]

		if !ok {
			score = 0
		}

		standingsMap[result.UserID] = score + result.Score
	}

	// convert back to map
	var standings []leaderboardUser
	for userID, score := range standingsMap {
		standings = append(standings, leaderboardUser{
			UserID: userID,
			Score:  score,
		})
	}

	sort.Slice(standings, func(i, j int) bool {
		return standings[i].Score > standings[j].Score
	})

	board.Standings = standings
	board.LastGameDayEvaluated = gameDay

	err = upsertLeaderboard(*board)
	return err
}
