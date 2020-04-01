package main

import (
	"log"
	"nba-pick-and-play/config"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-playground/validator.v9"
)

var (
	validate *validator.Validate
)

func main() {
	config.LoadConfig("config/config_dev.toml")

	setupDatabase()

	if config.Config.Rapid.Enabled {
		c := cron.New()
		c.AddFunc("0 9 * * *", func() { // 9am daily
			evaluateMatches(time.Now())
		})
		c.Start()
	}

	// interface for the Rapid API requests
	rapidAPIClient = baseRapidAPIClient{}

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
	userRouter.HandleFunc("/matches", getMatchesByDate).Methods("GET")

	adminRouter := router.PathPrefix("/v1/admin").Subrouter()
	adminRouter.HandleFunc("/evaluate", forceEvaluation).Methods("POST")
}
