package httputils

import (
	"encoding/json"
	"log"
	"net/http"
)

func Respond(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		//_, _, err := easyjson.MarshalToHTTPResponseWriter(data, w)
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Print(err, data)
			return
		}
	}
}


