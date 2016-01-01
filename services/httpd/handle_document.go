package httpd

import (
	"encoding/json"
	"net/http"

	"github.com/vatcinc/bio/models"
	"github.com/vatcinc/bio/schema"
)

func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id := q.Get("id")
	if id == "" {
		httpError(w, "id parameter is not defined", false, http.StatusNotAcceptable)
		return
	}
	doc, err := models.GetDocument(h.Mongo, id)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	answer := MarshalJSON(doc, false)
	w.Header().Add("content-type", "application/json")
	w.Write(answer)
}

func (h *Handler) wixImport(w http.ResponseWriter, r *http.Request) {
	doc, err := models.WixFormatImport(h.Mongo, r.Body)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(doc, false))
}

func (h *Handler) upsertDocument(w http.ResponseWriter, r *http.Request, user *bio.Users) {

	decoder := json.NewDecoder(r.Body)
	var doc = &models.Document{}
	err := decoder.Decode(&doc)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusInternalServerError)
		return
	}
	if err := doc.Authorize(h.Mongo, user); err != nil {
		httpError(w, err.Error(), false, http.StatusUnauthorized)
	}
	responseDocument, err := models.UpsertDocument(h.Mongo, doc)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(responseDocument, false))
}

// func (h *Handler) insertDocument(w http.ResponseWriter, r *http.Request, user *bio.Users) {
// 	w.Header().Add("content-type", "application/json")
// 	decoder := json.NewDecoder(r.Body)
// 	var o = &models.Widget{}
// 	err := decoder.Decode(&o)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusInternalServerError)
// 		h.Logger.Println("Error while Unmarshalling Request Body err")
// 		return
// 	}
// 	w.Write(MarshalJSON(models.UpsertWidget(h.Mongo, o), false))
// }
//
// func (h *Handler) deleteWidget(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Add("content-type", "application/json")
// 	widget := r.URL.Query().Get("id")
// 	if widget == "" {
// 		httpError(w, "Widget ID is empty", false, http.StatusInternalServerError)
// 		h.Logger.Println("Error while Getting Query parameter")
// 		return
// 	}
// 	w.Write(MarshalJSON(models.DeleteWidget(h.Mongo, widget), false))
// }

// This handler should only be callable by an admin
func (h *Handler) getDocuments(w http.ResponseWriter, r *http.Request, u *bio.Users) {
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(models.GetDocuments(h.Mongo), false))
}
