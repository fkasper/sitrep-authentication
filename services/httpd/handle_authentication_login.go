package httpd

import (
	"encoding/json"
	"net/http"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/rcrowley/go-metrics"
)

func (h *Handler) authenticationLoginService(w http.ResponseWriter, r *http.Request) {
	counter := metrics.GetOrRegisterCounter(statAuthFail, h.statMap)
	req, err := unmarshalRequest(r)
	if err != nil {
		httpError(w, "Login failed", false, http.StatusInternalServerError)
		return
	}
	if req.GrantType != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
		counter.Inc(1)
		httpError(w, "grant type must be urn:ietf:params:oauth:grant-type:jwt-bearer to request a password", false, http.StatusInternalServerError)
		return
	}

	if req.Username == "" || req.Password == "" {
		counter.Inc(1)
		httpError(w, "username or password missing", false, http.StatusForbidden)
		return
	}
	jwtResponse, err := models.UserSignIn(h.Cassandra, req.Username, req.Password, req.GrantType)
	if err != nil {
		counter.Inc(1)
		httpError(w, err.Error(), false, http.StatusForbidden)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(jwtResponse, false))
}

func unmarshalRequest(r *http.Request) (AuthenticationRequest, error) {
	decoder := json.NewDecoder(r.Body)
	var req AuthenticationRequest
	err := decoder.Decode(&req)
	if err != nil {
		return req, err
	}
	return req, nil
}

// AuthenticationRequest defines an inbound authentication req
type AuthenticationRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}
