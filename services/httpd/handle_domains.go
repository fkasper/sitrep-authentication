package httpd

import (
	"net/http"

	"github.com/gocql/gocql"
	"github.com/fkasper/sitrep-biometrics/models"
	"github.com/fkasper/sitrep-biometrics/schema"
)

//
// import (
// 	"encoding/json"
// 	"net/http"
//
// 	"github.com/gocql/gocql"
// 	"github.com/fkasper/sitrep-biometrics/models"
// 	"github.com/fkasper/sitrep-biometrics/schema"
// )
//
// func (h *Handler) insertDomain(w http.ResponseWriter, r *http.Request, user *bio.Users) {
// 	data, err := h.unwrapDoc(r.Body)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotAcceptable)
// 		return
// 	}
// 	model := &bio.Domains{}
//
// 	if err := json.Unmarshal(data.Data.Attributes, &model); err != nil {
// 		httpError(w, err.Error(), false, http.StatusBadRequest)
// 		return
// 	}
// 	if err := models.InsertDomain(h.Cassandra, h.Elasticsearch, model); err != nil {
// 		httpError(w, err.Error(), false, http.StatusBadRequest)
// 		return
// 	}
// 	MarshalEmber(w, model.Id, model, "domain", false)
// }
//
// //TBD
// func (h *Handler) updateDomain(w http.ResponseWriter, r *http.Request, user *bio.Users) {
// 	model, err := h.findDomain(r)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotFound)
// 		return
// 	}
// 	data, err := h.unwrapDoc(r.Body)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusBadRequest)
// 		return
// 	}
// 	if err := json.Unmarshal(data.Data.Attributes, &model); err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotAcceptable)
// 		return
// 	}
// 	if err := models.UpdateDomain(h.Cassandra, h.Elasticsearch, model); err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotAcceptable)
// 		return
// 	}
// 	MarshalEmber(w, model.Id, model, "domain", false)
// }
//
// func (h *Handler) searchDomains(w http.ResponseWriter, r *http.Request, user *bio.Users) {
// 	q := r.URL.Query()
// 	query := q.Get("query")
// 	limit := q.Get("limit")
// 	if limit == "" {
// 		limit = "30"
// 	}
// 	org, err := models.SearchDomains(h.Elasticsearch, query, limit)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotFound)
// 		return
// 	}
// 	w.Header().Add("content-type", "application/json")
// 	w.Write(MarshalJSON(org, false))
// }
//
// func (h *Handler) getDomain(w http.ResponseWriter, r *http.Request, user *bio.Users) {
// 	domain, err := h.findDomain(r)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotFound)
// 		return
// 	}
// 	if err := models.GetDomain(h.Cassandra, domain); err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotFound)
// 		return
// 	}
// 	MarshalEmber(w, domain.Id, domain, "domain", false)
// }
//
// func (h *Handler) deleteDomain(w http.ResponseWriter, r *http.Request, user *bio.Users) {
// 	domain, err := h.findDomain(r)
// 	if err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotFound)
// 		return
// 	}
// 	if err := models.DeleteDomain(h.Cassandra, h.Elasticsearch, domain.Id); err != nil {
// 		httpError(w, err.Error(), false, http.StatusNotFound)
// 		return
// 	}
// 	MarshalEmber(w, domain.Id, []string{}, "domain", false)
// }
//
func (h *Handler) findDomain(r *http.Request) (*bio.Domains, error) {
	domainID := r.URL.Query().Get(":id")
	domain := &bio.Domains{}
	if domainID == "" {
		return domain, &models.ValidationError{Field: "id", Reason: "ID is missing"}
	}
	id, err := gocql.ParseUUID(domainID)
	if err != nil {
		return domain, err
	}
	domain.Id = id
	return domain, nil
}
