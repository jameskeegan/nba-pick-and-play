# nba-pick-and-play
WIP Go backend project for web-based, competitive pick & play competition with work colleagues. 

## Current Functionality
As of the 9th April '20, project has the capablility to
* Poll a [public sports API for NBA matches](https://rapid.com/api-sports/api/api-nba) daily to get recent results and upcoming games
* Allow users to make picks for the upcoming night of NBA games
* Evaluate these picks daily when the data poll is done, with the user's score calculated from this
* Provide PvP leaderboards, both for daily results and overall season results

## To-do
* Proper user logic
* Proper admin logic
* Fall back methods if the daily poll fails
* Expand tests further (currently up to 60.8% line coverage) - Due to the lack of live data, having tests and stub interfaces has become quite important

## Assumptions
Due to this being developed during the Corona virus, there is no "live" data to test this with as all games have been cancelled. Thus, a few assumptions about the public API have been made.
* Upcoming games will include the time upon which they're expected to start (I'm expecting that this isn't currently being included in the data as there is no official start time for cancelled games)
* The data from the endpoint is in a consistent format
