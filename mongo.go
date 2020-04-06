package main

import (
	"context"
	"log"
	"nba-pick-and-play/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type (
	game struct {
		ID          int64     `bson:"_id" json:"id"`
		SeasonID    string    `bson:"seasonId" json:"seasonId"`
		Status      string    `bson:"status" json:"status"`
		GameDayID   string    `bson:"gameDayId" json:"gameDayId"` // simple "YYYY-MM-DD" to determine the game's actual date (UTC != PST)
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

	gameDayReport struct {
		ID        string               `bson:"_id" json:"id"`
		Games     map[int64]gameReport `bson:"games" json:"games"`
		Deadline  time.Time            `bson:"deadline" json:"deadline"`
		Evaluated bool                 `bson:"evaluated" json:"evaluated"`
	}

	gameReport struct {
		HomeTeam team      `bson:"homeTeam" json:"homeTeam"`
		AwayTeam team      `bson:"awayTeam" json:"awayTeam"`
		Venue    venue     `bson:"venue" json:"venue"`
		Date     time.Time `bson:"date" json:"date"`
		WinnerID int64     `bson:"winnerId" json:"winnerId,omitempty"`
	}

	gameDayPicks struct {
		ID        primitive.ObjectID `bson:"_id" json:"id"`
		UserID    int64              `bson:"userId" json:"userId"`
		GameDayID string             `bson:"gameDayId" json:"gameDayId"`
		Picks     map[int64]pick     `bson:"picks" json:"picks"`
		Evaluated bool               `bson:"evaluated" json:"evaluated"`
		Score     int64              `bson:"score" json:"score"`
		Date      time.Time          `bson:"date" json:"date"`
	}

	pick struct {
		SelectionID int64  `bson:"selectionId" json:"selectionId"`
		Status      string `bson:"status" json:"status"`
	}

	gameDayResults struct {
		ID         string   `bson:"_id" json:"id"`
		UserScores []result `bson:"scores" json:"scores"`
	}

	leaderboard struct {
		ID                   string            `bson:"_id" json:"id"` // the specific season
		Standings            []leaderboardUser `bson:"standings" json:"standings"`
		LastGameDayEvaluated string            `bson:"lastGameDay" json:"lastGameDay"`
	}

	leaderboardUser struct {
		UserID int64 `bson:"userId" json:"userId"`
		Score  int64 `bson:"score" json:"score"`
	}
)

const (
	gameDaysCollection       = "gameDays"
	gameDayResultsCollection = "gameDayResults"
	gamesCollection          = "games"
	leaderboardCollection    = "leaderboards"
	picksCollection          = "picks"
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

	_, err := db.Collection(gamesCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bsonx.Doc{
				{"gameDayId", bsonx.Int32(1)},
			},
			Options: options.Index().SetName("gameDayIdIndex").SetBackground(true),
		},
	)

	return err
}

func findMatchesByGameDateID(gameDateID string) ([]game, error) {
	db := getDatabase()

	filter := bson.D{
		{"gameDayId", gameDateID},
	}

	options := options.FindOptions{}
	options.SetSort(bson.D{{"startDateUTC", 1}})

	cur, err := db.Collection(gamesCollection).Find(
		context.Background(),
		filter,
		&options,
	)

	if err != nil {
		return nil, err
	}

	var games []game
	err = cur.All(context.Background(), &games)

	return games, err
}

func findGameDayReportByID(id string) (*gameDayReport, error) {
	db := getDatabase()

	var report gameDayReport
	err := db.Collection(gameDaysCollection).FindOne(
		context.Background(),
		bson.D{
			{"_id", id},
		},
	).Decode(&report)

	return &report, err
}

func findGameDayResultsReportByID(id string) (*gameDayResults, error) {
	db := getDatabase()

	var results gameDayResults
	err := db.Collection(gameDayResultsCollection).FindOne(
		context.Background(),
		bson.D{
			{"_id", id},
		},
	).Decode(&results)

	return &results, err
}

func findLeaderboardByID(id string) (*leaderboard, error) {
	db := getDatabase()

	var leaderboard leaderboard
	err := db.Collection(leaderboardCollection).FindOne(
		context.Background(),
		bson.D{
			{"_id", id},
		},
	).Decode(&leaderboard)

	return &leaderboard, err
}

func findPickReportsByGameDayID(date string) ([]gameDayPicks, error) {
	db := getDatabase()

	cur, err := db.Collection(picksCollection).Find(
		context.Background(),
		bson.D{
			{"gameDayId", date},
		},
	)

	if err != nil {
		return nil, err
	}

	var picks []gameDayPicks
	err = cur.All(context.Background(), &picks)

	return picks, err
}

func findPickReportsByGameDayIDAndEvaluated(date string, evaluated bool) ([]gameDayPicks, error) {
	db := getDatabase()

	cur, err := db.Collection(picksCollection).Find(
		context.Background(),
		bson.D{
			{"gameDayId", date},
			{"evaluated", evaluated},
		},
	)

	if err != nil {
		return nil, err
	}

	var picks []gameDayPicks
	err = cur.All(context.Background(), &picks)

	return picks, err
}

func upsertMatch(game game) error {
	db := getDatabase()

	options := options.ReplaceOptions{}
	options.SetUpsert(true)

	_, err := db.Collection(gamesCollection).ReplaceOne(
		context.Background(),
		bson.D{
			{"_id", game.ID},
		},
		game,
		&options,
	)

	return err
}

func upsertGameDayPicks(picks gameDayPicks) error {
	db := getDatabase()

	options := options.UpdateOptions{}
	options.SetUpsert(true)

	_, err := db.Collection(picksCollection).UpdateOne(
		context.Background(),
		bson.D{
			{"userId", picks.UserID},
			{"gameDayId", picks.GameDayID},
		},
		bson.D{
			{"$set", bson.D{
				{"picks", picks.Picks},
				{"evaluated", picks.Evaluated},
				{"score", picks.Score},
				{"date", picks.Date},
			}},
		},
		&options,
	)

	return err
}

func upsertGameDayReport(report gameDayReport) error {
	db := getDatabase()

	options := options.ReplaceOptions{}
	options.SetUpsert(true)

	_, err := db.Collection(gameDaysCollection).ReplaceOne(
		context.Background(),
		bson.D{
			{"_id", report.ID},
		},
		report,
		&options,
	)

	return err
}

func upsertGameDayResults(date string, results []result) error {
	db := getDatabase()

	options := options.UpdateOptions{}
	options.SetUpsert(true)

	_, err := db.Collection(gameDayResultsCollection).UpdateOne(
		context.Background(),
		bson.D{
			{"_id", date},
		},
		bson.D{
			{"$set", bson.D{
				{"scores", results},
			}},
		},
		&options,
	)

	return err
}

func upsertLeaderboard(leaderboard leaderboard) error {
	db := getDatabase()

	options := options.ReplaceOptions{}
	options.SetUpsert(true)

	_, err := db.Collection(leaderboardCollection).ReplaceOne(
		context.Background(),
		bson.D{
			{"_id", leaderboard.ID},
		},
		leaderboard,
		&options,
	)

	return err
}

func getDatabase() *mongo.Database {
	return mongoClient.Database(config.Config.Mongo.Name)
}
