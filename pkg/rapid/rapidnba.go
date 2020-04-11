package rapid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type (
	//NBAResponse response format for the rapid NBA API response
	NBAResponse struct {
		ResponseWrapper ResponseWrapper `json:"API"`
	}

	ResponseWrapper struct {
		Status  int       `json:"status"`
		Message string    `json:"message"`
		Results int       `json:"results"`
		Filters []string  `json:"filters"`
		Games   []NBAGame `json:"games"`
	}

	NBAGame struct {
		SeasonYear      string    `json:"seasonYear"`
		League          string    `json:"league"`
		GameID          string    `json:"gameId"`
		StartTimeUTC    time.Time `json:"startTimeUTC"`
		EndTimeUTC      string    `json:"endTimeUTC"`
		Arena           string    `json:"arena"`
		City            string    `json:"city"`
		Country         string    `json:"country"`
		Clock           string    `json:"clock"`
		GameDuration    string    `json:"gameDuration"`
		CurrentPeriod   string    `json:"currentPeriod"`
		Halftime        string    `json:"halftime"`
		EndOfPeriod     string    `json:"EndOfPeriod"`
		SeasonStage     string    `json:"seasonStage"`
		StatusShortGame string    `json:"statusShortGame"`
		StatusGame      string    `json:"statusGame"`
		VTeam           NBATeam   `json:"vTeam"`
		HTeam           NBATeam   `json:"hTeam"`
	}

	NBATeam struct {
		TeamID    string `json:"teamId"`
		ShortName string `json:"shortName"`
		FullName  string `json:"fullName"`
		NickName  string `json:"nickName"`
		Logo      string `json:"logo"`
		Score     Score  `json:"score"`
	}

	Score struct {
		Points string `json:"points"`
	}

	//Client interface for connecting to the Rapid API (or can be mocked for testing)
	Client interface {
		GetMatchesByDateRequest(date string) (*NBAResponse, error)
	}

	apiClient struct {
		baseURL string
		apiKey  string
	}
)

func (c apiClient) GetMatchesByDateRequest(date string) (*NBAResponse, error) {
	apiClient := http.Client{}
	req, err := http.NewRequest("GET", c.baseURL+date, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("x-rapidapi-key", c.apiKey)

	resp, err := apiClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response NBAResponse
	json.NewDecoder(resp.Body).Decode(&response)

	return &response, nil
}

//NewRapidAPIClient returns an implemented Client for use in the application
func NewRapidAPIClient(url string, apiKey string) *apiClient {
	return &apiClient{
		baseURL: url,
		apiKey:  apiKey,
	}
}

type (
	mockRapidAPIClient struct {
		matchesData map[string]string // e.g. "2020-01-18" -> "test/2020-01-18.json"
	}
)

// loads the file associated with a given select date
func (c mockRapidAPIClient) GetMatchesByDateRequest(date string) (*NBAResponse, error) {
	path, ok := c.matchesData[date]

	if !ok {
		return nil, fmt.Errorf("no file found for date %s", date)
	}

	file, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var response NBAResponse
	err = json.Unmarshal([]byte(file), &response)

	return &response, err
}

//NewMockRapidClient returns an implemented Client for testing use in the application
func NewMockRapidClient(matches map[string]string) *mockRapidAPIClient {
	return &mockRapidAPIClient{
		matchesData: matches,
	}
}
