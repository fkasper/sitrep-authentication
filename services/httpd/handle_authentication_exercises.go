package httpd

import (
	"net/http"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/fkasper/sitrep-authentication/schema"
)

func (h *Handler) authenticationGetExercisesService(w http.ResponseWriter, r *http.Request, u *sitrep.UsersByEmail) {
	exercises, err := models.FindExercisesForUser(h.Cassandra, u)
	if err != nil {
		httpError(w, "Failed to fetch exercises", false, http.StatusInternalServerError)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(exercises, false))
}

func (h *Handler) authenticationGetCurrentExercisePermissions(w http.ResponseWriter, r *http.Request, u *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier) {
	exercises, err := models.FindExercisePermissionsForUser(h.Cassandra, u, exercise)
	if err != nil {
		httpError(w, "User is not authorized in this exercise at all!", false, http.StatusUnauthorized)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(exercises, false))
}
