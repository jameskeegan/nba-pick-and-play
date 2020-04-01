package main

import (
	"fmt"
	"nba-pick-and-play/response"
	"net/http"
	"time"
)

func getMatchesByDate(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")

	if date == "" { // get date as the current date
		date = time.Now().Format(basicDateFormat)
	}

	matches, err := findMatchesByGameDateID(date)

	if err != nil {
		response.ReturnInternalServerError(w, nil, err.Error())
		return
	}

	if len(matches) == 0 {
		response.ReturnNotFound(w, nil, fmt.Sprintf("no matches found for game date '%s'", date))
		return
	}

	response.ReturnStatusOK(w, matches)
}
