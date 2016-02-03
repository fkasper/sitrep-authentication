package httpd

import (
	"net/http"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/fkasper/sitrep-authentication/schema"
)

func (h *Handler) getUsersList(w http.ResponseWriter, r *http.Request, u *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier) {
	exercises, err := models.FindExercisePermissionsForUser(h.Cassandra, u, exercise)
	if err != nil {
		httpError(w, "User is not authorized in this exercise at all!", false, http.StatusUnauthorized)
		return
	}
	if !exercises.IsAdmin || !u.IsAdmin {
		httpError(w, "User is authorized to fetch a list of users", false, http.StatusUnauthorized)
		return
	}
	users, err := models.FetchAllUsers(h.Cassandra)
	if err != nil {
		httpError(w, "Error occured while fetching data", false, http.StatusInternalServerError)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(users, false))
}
