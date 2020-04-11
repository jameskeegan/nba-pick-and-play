package main

import (
	"flag"
	"nba-pick-and-play/config"
	clockPkg "nba-pick-and-play/pkg/clock"
	"nba-pick-and-play/pkg/rapid"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

var (
	clock          clockPkg.Clock
	rapidAPIClient rapid.Client
	validate       *validator.Validate

	log *logrus.Logger
)

func main() {
	configPath := flag.String("config", "config/config_dev.toml", "location of the config to be used")
	flag.Parse()

	config.LoadConfig(*configPath)

	log = logrus.New()

	setupDatabase()

	if config.Config.Rapid.Enabled {
		c := cron.New()
		c.AddFunc("0 9 * * *", dailyCron) // 9am daily
		c.Start()
	}

	// interface for the Rapid API requests
	rapidAPIClient = rapid.NewRapidAPIClient(config.Config.Rapid.BaseURL, config.Config.Rapid.APIKey)

	clock = clockPkg.NewClock()

	validate = validator.New()

	router := mux.NewRouter()
	initRouter(router)

	log.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalln(err.Error())
	}
}

func initRouter(router *mux.Router) {
	userRouter := router.PathPrefix("/v1/user").Subrouter()

	userRouter.HandleFunc("/games", getGameDayReport).Methods("GET")
	userRouter.HandleFunc("/results", getGameDayResultsReport).Methods("GET")
	userRouter.HandleFunc("/leaderboards", getLeaderboard).Methods("GET")
	userRouter.HandleFunc("/picks", makePicks).Methods("POST")

	// TODO: admin router for handling resyncs
}

/*
	What to do on a daily basis

	- get results of last night's matches
		- requires two calls to get games either side of midnight
	- get the upcoming games tomorrow
		- requires one more call (to get games after midnight tonight)
	- evaluate last night's matches
		- get the scores, update the game day report, mark the users picks
	- create game day report for tonight's upcoming matches
*/
func dailyCron() {
	dateNow := clock.Now()

	dateToday := dateNow.Format(basicDateFormat)
	dateYesterday := dateNow.Add(-24 * time.Hour).Format(basicDateFormat)
	dateTomorrow := dateNow.Add(24 * time.Hour).Format(basicDateFormat)

	// get game information for the 24 hours eitherside of now
	err := pollGames(dateYesterday, dateToday, dateTomorrow)

	if err != nil {
		log.Error(err.Error())
		return
	}

	// evaluate yesterday's matches
	err = evaluateGameDayReport(dateYesterday)

	if err != nil { // don't return if err occurs, still create upcoming report
		log.Error(err.Error())
	} else {
		err = createGameDayResults(dateYesterday)

		if err != nil {
			log.Error(err.Error())
		}

		err = updateLeaderboard(config.Config.Rapid.Season)

		if err != nil {
			log.Error(err.Error())
		}
	}

	// create a report for the upcoming matches tonight
	_, err = createGameDayReport(dateToday)

	if err != nil {
		log.Error(err.Error())
	}
}
