package response

import (
	"encoding/json"
	"net/http"
)

//ReturnBadRequest returns a 400 and the data in json format
func ReturnBadRequest(w http.ResponseWriter, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(data)
}

//ReturnStatusOK returns a 200 and the data in json format
func ReturnStatusOK(w http.ResponseWriter, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
