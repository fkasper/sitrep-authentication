package httpd

import (
	"encoding/json"
	"net/http"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/fkasper/sitrep-authentication/schema"
)

func (h *Handler) authenticationPasswordChangeService(w http.ResponseWriter, r *http.Request, u *sitrep.UsersByEmail) {

	req, err := unmarshalPasswordChangeRequest(r)
	if err != nil {
		httpError(w, "Password could not be changed", false, http.StatusInternalServerError)
		return
	}
	if req.NewPassword != req.NewPasswordConfirmation {
		httpError(w, "Passwords do not match", false, http.StatusExpectationFailed)
		return
	}

	if req.OldPassword == "" {
		httpError(w, "Old password is empty", false, http.StatusExpectationFailed)
		return
	}
	pwChange, err := models.UserChangePassword(h.Cassandra, u, req.OldPassword, req.NewPassword)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusForbidden)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(pwChange, false))
}

func unmarshalPasswordChangeRequest(r *http.Request) (PasswordChangeRequest, error) {
	decoder := json.NewDecoder(r.Body)
	var req PasswordChangeRequest
	err := decoder.Decode(&req)
	if err != nil {
		return req, err
	}
	return req, nil
}

// PasswordChangeRequest defines an inbound authentication req
type PasswordChangeRequest struct {
	NewPassword             string `json:"new_password"`
	NewPasswordConfirmation string `json:"new_password_confirmation"`
	OldPassword             string `json:"old_password"`
}
