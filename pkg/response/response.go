package response

import (
	"encoding/json"
	"net/http"
	"time"
)

type (
	res struct {
		Code      int64       `json:"code"`
		Data      interface{} `json:"data,omitempty"`
		Error     *string     `json:"error,omitempty"`
		CreatedAt time.Time   `json:"createdAt"`
	}
)

//ReturnError for when you want to return a non 2xx response
func ReturnError(w http.ResponseWriter, statusCode int, errorMessage string) {
	send(w, statusCode, nil, &errorMessage)
}

//ReturnSuccess for when you want to return a 2xx response
func ReturnSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	send(w, statusCode, data, nil)
}

func send(w http.ResponseWriter, statusCode int, data interface{}, errorMessage *string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res := res{
		Code:      int64(statusCode),
		Data:      data,
		Error:     errorMessage,
		CreatedAt: time.Now(),
	}

	json.NewEncoder(w).Encode(res)
}
