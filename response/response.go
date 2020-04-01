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

//ReturnStatusOK returns a 200 and the data in json format
func ReturnStatusOK(w http.ResponseWriter, data interface{}) {
	send(w, http.StatusOK, data, nil)
}

//ReturnBadRequest returns a 400 and the data in json format
func ReturnBadRequest(w http.ResponseWriter, data interface{}, errorMessage string) {
	send(w, http.StatusBadRequest, data, &errorMessage)
}

//ReturnNotFound returns a 404 and the data in json format
func ReturnNotFound(w http.ResponseWriter, data interface{}, errorMessage string) {
	send(w, http.StatusNotFound, data, &errorMessage)
}

//ReturnInternalServerError returns a 500 and the data in json format
func ReturnInternalServerError(w http.ResponseWriter, data interface{}, errorMessage string) {
	send(w, http.StatusInternalServerError, data, &errorMessage)
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
