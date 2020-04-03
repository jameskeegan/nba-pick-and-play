package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPollGamesSuccess(t *testing.T) {
	defer cleanDatabase(t)

	// polls for games that took place on this date (UTC)
	err := pollGames("2020-01-18")
	assert.Nil(t, err)

	// check if parsing was genuinely successful
	matches, err := findMatchesByGameDateID("2020-01-18")
	assert.Nil(t, err)

	assert.NotNil(t, matches)
	assert.Equal(t, 2, len(matches))

	// Pelicans @ Clippers
	gameOne := matches[0]
	assert.Equal(t, int64(7015), gameOne.ID)
	assert.Equal(t, "Scheduled", gameOne.Status)

	assert.Equal(t, int64(16), gameOne.AwayTeam.ID)
	assert.Equal(t, "Clippers", gameOne.AwayTeam.Nickname)
	assert.Zero(t, gameOne.AwayTeam.Score)

	assert.Equal(t, int64(23), gameOne.HomeTeam.ID)
	assert.Equal(t, "Pelicans", gameOne.HomeTeam.Nickname)
	assert.Zero(t, gameOne.HomeTeam.Score)

	// Bucks @ Nets
	assert.Equal(t, int64(7016), matches[1].ID)

	// check that parsing for games that took place on the previous day worked too
	matches, err = findMatchesByGameDateID("2020-01-17")
	assert.Nil(t, err)

	assert.NotNil(t, matches)
	assert.Equal(t, 7, len(matches))
}

func TestCreateGameDayReportSuccess(t *testing.T) {
	defer cleanDatabase(t)

	// polls for games that took place on this date (UTC)
	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	report, err := findGameDayReportByID("2020-01-18")
	assert.Nil(t, err)

	assert.NotNil(t, report)
	assert.Equal(t, "2020-01-18", report.ID)
	assert.Equal(t, 11, len(report.Games))
	assert.Equal(t, false, report.Evaluated)

	// 8:30pm Jan 18th 2020 (UTC) - time of the start of the first game
	assert.Equal(t, time.Date(2020, time.January, 18, 20, 30, 0, 0, time.UTC), report.Deadline)
}

func TestEvaluateGameDayReportSuccess(t *testing.T) {
	defer cleanDatabase(t)

	err := pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	err = createGameDayReport("2020-01-18")
	assert.Nil(t, err)

	// create and insert some user picks
	gameDayPicks := gameDayPicks{
		UserID:    12345,
		GameDayID: "2020-01-18",
		Picks:     createPicks(),
		Evaluated: false,
		Score:     0,
		Date:      clockClient.now(),
	}

	err = upsertGameDayPicks(gameDayPicks)
	assert.Nil(t, err)

	// substitute the client again, this time for one with the results data
	rapidAPIClient = mockRapidAPIClient{
		eighteenthPath: "test/2020-01-18_nextday.json",
		nineteenthPath: "test/2020-01-19_nextday.json",
	}

	// undo this change for other tests
	defer setDefaultMockRapidAPIClient()

	// force a poll as if it was the following day
	err = pollGames("2020-01-18", "2020-01-19")
	assert.Nil(t, err)

	// check the games have been updated
	matches, err := findMatchesByGameDateID("2020-01-18")
	assert.Nil(t, err)

	for _, m := range matches {
		assert.Equal(t, "Finished", m.Status)

		// (naively) check that the scores have been updated too
		assert.NotZero(t, m.HomeTeam.Score)
		assert.NotZero(t, m.AwayTeam.Score)
		assert.NotZero(t, m.WinnerID)
	}

	// evaluate the game day report
	err = evaluateGameDayReport("2020-01-18")
	assert.Nil(t, err)

	report, err := findGameDayReportByID("2020-01-18")
	assert.Nil(t, err)
	assert.True(t, report.Evaluated)

	for _, m := range report.Games {
		assert.NotZero(t, m.HomeTeam.Score)
		assert.NotZero(t, m.AwayTeam.Score)
		assert.NotZero(t, m.WinnerID)
	}

	// check the user picks
	pickReports, err := findPickReportsByGameDayID("2020-01-18")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pickReports))

	rep := pickReports[0]
	assert.True(t, rep.Evaluated)
	assert.Equal(t, "2020-01-18", rep.GameDayID)
	assert.Equal(t, int64(7), rep.Score)
}

func createPicks() map[int64]pick {
	return map[int64]pick{
		7015: pick{
			SelectionID: 23,
			Status:      "PENDING",
		},
		7016: pick{
			SelectionID: 21,
			Status:      "PENDING",
		},
		7017: pick{
			SelectionID: 2,
			Status:      "PENDING",
		},
		7018: pick{
			SelectionID: 1,
			Status:      "PENDING",
		},
		7019: pick{
			SelectionID: 27,
			Status:      "PENDING",
		},
		7020: pick{
			SelectionID: 7,
			Status:      "PENDING",
		},
		7021: pick{
			SelectionID: 38,
			Status:      "PENDING",
		},
		7022: pick{
			SelectionID: 17,
			Status:      "PENDING",
		},
		7023: pick{
			SelectionID: 11,
			Status:      "PENDING",
		},
		7024: pick{
			SelectionID: 25,
			Status:      "PENDING",
		},
		7025: pick{
			SelectionID: 40,
			Status:      "PENDING",
		},
	}

}
