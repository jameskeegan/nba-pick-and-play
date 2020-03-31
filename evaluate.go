package main

import (
	"log"
	"time"
)

const (
	rapidBasicDataFormat = "2006-01-02"
)

func evaluateMatches(evaluationDate time.Time) error {
	log.Println("Evaluating...")

	// evaluate yesterday's matches
	date := evaluationDate.Add(-24 * time.Hour).Format(rapidBasicDataFormat)
	err := evaluateMatchDay(date)

	if err != nil {
		log.Printf("ERROR: Could not evaluate matches for %s: %s", date, err.Error())
		return err
	}

	// evaluate today's matches
	date = evaluationDate.Format(rapidBasicDataFormat)
	err = evaluateMatchDay(date)

	if err != nil {
		log.Printf("ERROR: Could not evaluate matches for %s: %s", date, err.Error())
		return err
	}

	// evaluate tomorrow's matches
	date = evaluationDate.Add(24 * time.Hour).Format(rapidBasicDataFormat)
	err = evaluateMatchDay(date)

	if err != nil {
		log.Printf("ERROR: Could not evaluate matches for %s: %s", date, err.Error())
		return err
	}

	log.Println("Evaluation complete")
	return nil
}

func evaluateMatchDay(date string) error {
	res, err := rapidAPIObject.getMatchesByDateRequest(date)

	if err != nil {
		log.Printf("ERROR: Could not evaluate matches: %s", err.Error())
		return err
	}

	for _, game := range res.ResponseWrapper.Games {
		match, err := gameToMatch(game) // convert to our mongo schema

		if err != nil {
			log.Printf("ERROR: could not save match %s: %s", game.GameID, err.Error())
			return err
		}

		err = insertMatch(*match)

		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	return nil
}
