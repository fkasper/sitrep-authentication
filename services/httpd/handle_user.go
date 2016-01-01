package httpd

import (
	"net/http"

	"github.com/rcrowley/go-metrics"
	"github.com/vatcinc/bio/models"
	// "github.com/gorilla/websocket"
)

func (h *Handler) authUser(w http.ResponseWriter, r *http.Request) {
	counter := metrics.GetOrRegisterCounter(statAuthFail, h.statMap)

	grantType := r.PostFormValue("grant_type")
	if grantType != "password" {
		counter.Inc(1)
		httpError(w, "grant type must be password for now", false, http.StatusInternalServerError)
		return
	}
	email := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if email == "" || password == "" {
		counter.Inc(1)
		httpError(w, "username or password missing", false, http.StatusForbidden)
		return
	}
	user, err := models.SignInUser(h.Cassandra, email, password, grantType)
	if err != nil {
		counter.Inc(1)
		httpError(w, err.Error(), false, http.StatusForbidden)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(user, false))
}
