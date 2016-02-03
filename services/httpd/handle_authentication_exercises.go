package httpd

import (
	"encoding/json"
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

func (h *Handler) authenticationGetExercisesSettings(w http.ResponseWriter, r *http.Request, exercise *sitrep.ExerciseByIdentifier) {
	settings, err := models.FindOrInitSettingsForExercise(h.Cassandra, exercise.Id)
	if err != nil {
		httpError(w, "An unexpected error occured, while fetching your data!", false, http.StatusInternalServerError)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(settings, false))
}

func (h *Handler) authenticationUpdateExercisesSettings(w http.ResponseWriter, r *http.Request, u *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier) {
	exercises, err := models.FindExercisePermissionsForUser(h.Cassandra, u, exercise)
	if err != nil {
		httpError(w, "User is not authorized in this exercise at all!", false, http.StatusUnauthorized)
		return
	}
	if !exercises.IsAdmin || !u.IsAdmin {
		httpError(w, "User is authorized to update settings", false, http.StatusUnauthorized)
		return
	}
	req, err := unmarshalSettingsUpdateRequest(r)
	if err != nil {
		httpError(w, "Error occured while processing your settings!", false, http.StatusInternalServerError)
		return
	}
	updated, err := models.UpdateExerciseSetting(h.Cassandra, exercise.Id, req.Values)
	if err != nil {
		httpError(w, "Error occured while saving your settings!", false, http.StatusInternalServerError)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(updated, false))
}

func unmarshalSettingsUpdateRequest(r *http.Request) (SettingsUpdateRequest, error) {
	decoder := json.NewDecoder(r.Body)
	var req SettingsUpdateRequest
	err := decoder.Decode(&req)
	if err != nil {
		return req, err
	}
	return req, nil
}

// SettingsUpdateRequest defines an inbound settings update req
type SettingsUpdateRequest struct {
	Values map[string]string `json:"values"`
}
