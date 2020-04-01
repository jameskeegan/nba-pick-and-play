package main

import (
	"context"
	"log"
	"nba-pick-and-play/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type (
	match struct {
		ID          int64     `bson:"_id" json:"id"`
		SeasonID    string    `bson:"seasonId" json:"seasonId"`
		Status      string    `bson:"status" json:"status"`
		GameDateID  string    `bson:"gameDateId" json:"gameDateId"` // simple "YYYY-MM-DD" to determine the game's actual date (UTC != PST)
		SeasonStage string    `bson:"seasonStage" json:"seasonStage"`
		StartDate   time.Time `bson:"startDate" json:"startDate"` // UTC
		WinnerID    int64     `bson:"winnerId" json:"winnerId"`   // id of the winning team
		HomeTeam    team      `bson:"homeTeam" json:"homeTeam"`
		AwayTeam    team      `bson:"awayTeam" json:"awayTeam"`
		Venue       venue     `bson:"venue" json:"venue"`
	}

	team struct {
		ID       int64  `bson:"id" json:"id"`
		Name     string `bson:"name" json:"name"`
		Nickname string `bson:"nickname" json:"nickname"`
		Logo     string `bson:"logo" json:"logo"`
		Score    int64  `bson:"score" json:"score"`
	}

	venue struct {
		Name    string `bson:"name" json:"name"`
		City    string `bson:"city" json:"city"`
		Country string `bson:"country" json:"country"`
	}
)

const (
	matchesCollection = "matches"
	picksCollection   = "picks"
)

var (
	mongoClient *mongo.Client
)

func setupDatabase() {
	clientOptions := options.Client().ApplyURI(config.Config.Mongo.HostURI)
	client, err := mongo.NewClient(clientOptions)

	if err != nil {
		log.Fatalf("couldn't connect to mongo: %s", err.Error())
	}

	err = client.Connect(context.Background())

	if err != nil {
		log.Fatalf("couldn't connect with client: %s", err.Error())
	}

	mongoClient = client

	err = createIndexes()

	if err != nil {
		log.Fatalf("couldn't create indexes: %s", err.Error())
	}

	log.Println("connected to mongodb")
}

func createIndexes() error {
	db := getDatabase()

	_, err := db.Collection(matchesCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bsonx.Doc{
				{"gameDateId", bsonx.Int32(1)},
			},
			Options: options.Index().SetName("gameDateIdIndex").SetBackground(true),
		},
	)

	return err
}

func findMatchesByGameDateID(gameDateID string) ([]match, error) {
	db := getDatabase()

	filter := bson.D{
		{"gameDateId", gameDateID},
	}

	options := options.FindOptions{}
	options.SetSort(bson.D{{"startDateUTC", 1}})

	cur, err := db.Collection(matchesCollection).Find(
		context.Background(),
		filter,
		&options,
	)

	if err != nil {
		return nil, err
	}

	var matches []match
	err = cur.All(context.Background(), &matches)

	return matches, err
}

func insertMatch(match match) error {
	db := getDatabase()

	options := options.ReplaceOptions{}
	options.SetUpsert(true)

	_, err := db.Collection(matchesCollection).ReplaceOne(
		context.Background(),
		bson.D{
			{"_id", match.ID},
		},
		match,
		&options,
	)

	return err
}

func getDatabase() *mongo.Database {
	return mongoClient.Database(config.Config.Mongo.Name)
}
