package httpd

import (
	"encoding/json"
	"net/http"

	"github.com/vatcinc/bio/models"
	"github.com/vatcinc/bio/schema"
	// "github.com/gorilla/websocket"
)

func (h *Handler) upsertWidget(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var o = &models.Widget{}
	err := decoder.Decode(&o)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusInternalServerError)
		h.Logger.Println("Error while Unmarshalling Request Body err")
		return
	}
	w.Write(MarshalJSON(models.UpsertWidget(h.Mongo, o), false))
}
func (h *Handler) deleteWidget(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	widget := r.URL.Query().Get("id")
	if widget == "" {
		httpError(w, "Widget ID is empty", false, http.StatusInternalServerError)
		h.Logger.Println("Error while Getting Query parameter")
		return
	}
	w.Write(MarshalJSON(models.DeleteWidget(h.Mongo, widget), false))
}

func (h *Handler) serveWidgetsList(w http.ResponseWriter, r *http.Request, u *bio.Users) {
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(models.GetWidgets(h.Mongo), false))
}
