package httputils

import (
	"encoding/json"
	"log"
	"net/http"
)

func Respond(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Print(err)
			return
		}
	}
}
