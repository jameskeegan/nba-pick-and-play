package main

import (
	"nba-pick-and-play/config"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type (
	clock interface { // so we can mock this
		now() time.Time
	}

	realClock struct{}
)

func (realClock) now() time.Time {
	return time.Now()
}

var (
	validate    *validator.Validate
	clockClient clock

	log *logrus.Logger
)

func init() {
	log = logrus.New()
	config.LoadConfig("config/config_dev.toml")
}

func main() {
	setupDatabase()

	if config.Config.Rapid.Enabled {
		c := cron.New()
		c.AddFunc("0 9 * * *", dailyCron)
		c.Start()
	}

	// interface for the Rapid API requests
	rapidAPIClient = baseRapidAPIClient{
		baseURL: config.Config.Rapid.BaseURL,
		apiKey:  config.Config.Rapid.APIKey,
	}

	clockClient = realClock{}

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
	dateNow := clockClient.now()

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
