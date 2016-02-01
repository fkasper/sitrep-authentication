package httpd

import (
	"net/http"

	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/fkasper/sitrep-authentication/utils"
)

func (h *Handler) receiveOwnProfileService(w http.ResponseWriter, r *http.Request, u *sitrep.UsersByEmail) {
	w.Header().Add("content-type", "application/json")

	w.Write(MarshalJSON(utils.MapUser(u), false))
}
