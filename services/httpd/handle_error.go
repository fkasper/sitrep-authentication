package httpd

import (
	"encoding/json"
	"net/http"
)

// httpError writes an error to the client in a standard format.
func httpError(w http.ResponseWriter, err string, pretty bool, code int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	var o struct {
		Err string `json:"error,omitempty"`
	}
	o.Err = err
	var b []byte
	var erro error
	if pretty {
		b, erro = json.MarshalIndent(o, "", "    ")
	} else {
		b, erro = json.Marshal(o)
	}
	if erro != nil {
		b = []byte("Json Encoding error!")
	}
	w.Write(b)
}

func resultError(w http.ResponseWriter, result Result, code int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(&result)
}
