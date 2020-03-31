package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"nba-pick-and-play/config"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	rapidAPIMock struct{}
)

func (rapidAPIMock) getMatchesByDateRequest(date string) (*rapidResponse, error) {
	path := "test/" + date + ".json"
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

func init() {
	config.LoadConfig("config/config_test.toml")

	setupDatabase()

	// mock API to return the json test files data as responses
	rapidAPIObject = rapidAPIMock{}
}

func TestEvaluate(t *testing.T) {
	defer cleanDatabase(t)

	date := time.Date(2020, time.January, 18, 9, 0, 0, 0, time.UTC) // 9am, Jan 17th 2020
	err := evaluateMatches(date)

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}
}

func cleanDatabase(t *testing.T) {
	db := getDatabase()

	_, err := db.Collection(matchesCollection).DeleteMany(
		context.Background(),
		bson.M{},
	)

	if err != nil {
		log.Fatalf("Failed to drop collection: %s", err.Error())
	}
}
