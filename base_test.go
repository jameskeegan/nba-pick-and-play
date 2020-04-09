package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"nba-pick-and-play/config"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	mockRapidAPIClient struct {
		seventeenthPath string
		eighteenthPath  string
		nineteenthPath  string
	}
)

// loads the file associated with a given select date
func (c mockRapidAPIClient) getMatchesByDateRequest(date string) (*rapidResponse, error) {
	var path string
	switch date {
	case "2020-01-17":
		path = c.seventeenthPath
	case "2020-01-18":
		path = c.eighteenthPath
	case "2020-01-19":
		path = c.nineteenthPath
	default: // unexpected behaviour
		log.Fatalf("ERROR: no file found for date %s", date)
	}

	file, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatalf("Failed to load file %s - %s", path, err.Error())
	}

	var response rapidResponse
	err = json.Unmarshal([]byte(file), &response)

	if err != nil {
		log.Fatalf("Failed to decode file %s - %s", path, err.Error())
	}

	return &response, nil
}

func setDefaultMockRapidAPIClient() {
	rapidAPIClient = mockRapidAPIClient{
		seventeenthPath: "test/2020-01-17.json",
		eighteenthPath:  "test/2020-01-18.json",
		nineteenthPath:  "test/2020-01-19.json",
	}
}

type (
	mockClock struct {
		date time.Time
	}
)

func (c mockClock) now() time.Time {
	return c.date
}

func setDefaultMockClock() {
	clockClient = mockClock{
		date: time.Date(2020, time.January, 18, 12, 0, 0, 0, time.UTC),
	}
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
