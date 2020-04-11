package main

import (
	"context"
	"nba-pick-and-play/config"
	clockPkg "nba-pick-and-play/pkg/clock"
	"nba-pick-and-play/pkg/rapid"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func setDefaultMockRapidAPIClient() {
	matchesToFiles := map[string]string{
		"2020-01-17": "test/2020-01-17.json",
		"2020-01-18": "test/2020-01-18.json",
		"2020-01-19": "test/2020-01-19.json",
	}

	rapidAPIClient = rapid.NewMockRapidClient(matchesToFiles)
}

func setDefaultMockClock() {
	clock = clockPkg.NewMockClock(time.Date(2020, time.January, 18, 12, 0, 0, 0, time.UTC))
}

func init() {
	log = logrus.New()

	config.LoadConfig("config/config_test.toml")

	setupDatabase()

	// mock API to return the json test files data as responses
	setDefaultMockRapidAPIClient()

	// change time to be 18th Jan 2020 noon instead of the actual time.Now()
	setDefaultMockClock()
}
func cleanDatabase(t *testing.T) {
	db := getDatabase()

	_, err := db.Collection(gamesCollection).DeleteMany(
		context.Background(),
		bson.M{},
	)

	if err != nil {
		log.Fatalf("Failed to drop collection: %s", err.Error())
	}

	_, err = db.Collection(gameDaysCollection).DeleteMany(
		context.Background(),
		bson.M{},
	)

	if err != nil {
		log.Fatalf("Failed to drop collection: %s", err.Error())
	}

	_, err = db.Collection(picksCollection).DeleteMany(
		context.Background(),
		bson.M{},
	)

	if err != nil {
		log.Fatalf("Failed to drop collection: %s", err.Error())
	}

	_, err = db.Collection(gameDayResultsCollection).DeleteMany(
		context.Background(),
		bson.M{},
	)

	if err != nil {
		log.Fatalf("Failed to drop collection: %s", err.Error())
	}

	_, err = db.Collection(leaderboardCollection).DeleteMany(
		context.Background(),
		bson.M{},
	)

	if err != nil {
		log.Fatalf("Failed to drop collection: %s", err.Error())
	}
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
