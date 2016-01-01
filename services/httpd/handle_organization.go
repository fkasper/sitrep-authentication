package httpd

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/vatcinc/bio/models"
	"github.com/vatcinc/bio/schema"
)

func (h *Handler) insertCustomer(w http.ResponseWriter, r *http.Request, user *bio.Users) {
	data, err := h.unwrapDoc(r.Body)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotAcceptable)
		return
	}
	model := &bio.Organizations{}

	if err := json.Unmarshal(data.Data.Attributes, &model); err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	if err := models.InsertOrganization(h.Cassandra, h.Elasticsearch, model); err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	MarshalEmber(w, model.Id, model, "organization", false)
}

//
// //TBD
func (h *Handler) updateCustomer(w http.ResponseWriter, r *http.Request, user *bio.Users) {
	model, err := h.findCustomer(r, ":id")
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	data, err := h.unwrapDoc(r.Body)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data.Data.Attributes, &model); err != nil {
		httpError(w, err.Error(), false, http.StatusNotAcceptable)
		return
	}
	if err := models.UpdateOrganization(h.Cassandra, h.Elasticsearch, model); err != nil {
		httpError(w, err.Error(), false, http.StatusNotAcceptable)
		return
	}
	MarshalEmber(w, model.Id, model, "organization", false)
}

//
func (h *Handler) searchCustomers(w http.ResponseWriter, r *http.Request, user *bio.Users) {
	q := r.URL.Query()
	query := q.Get("query")
	limit := q.Get("limit")
	if limit == "" {
		limit = "30"
	}
	org, err := models.SearchOrganizations(h.Elasticsearch, query, limit)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(org, false))
}

//
func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request, user *bio.Users) {
	organization, err := h.findCustomer(r, ":id")
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	if err := models.GetOrganization(h.Cassandra, organization); err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	MarshalEmber(w, organization.Id, organization, "organization", false)
}

func (h *Handler) getCustomerDomains(w http.ResponseWriter, r *http.Request, user *bio.Users) {
	organization, err := h.findCustomer(r, "id")
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	var domains models.EmberMultiData
	if err := models.GetOrganizationDomains(h.Cassandra, organization.Id, &domains); err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	MarshalMultiEmber(w, &domains, false)
}

//
func (h *Handler) deleteCustomer(w http.ResponseWriter, r *http.Request, user *bio.Users) {
	organization, err := h.findCustomer(r, ":id")
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	if err := models.DeleteOrganization(h.Cassandra, h.Elasticsearch, organization.Id); err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	MarshalEmber(w, organization.Id, []string{}, "organization", false)
}

func (h *Handler) findCustomer(r *http.Request, field string) (*bio.Organizations, error) {
	customerID := r.URL.Query().Get(field)
	organization := &bio.Organizations{}
	if customerID == "" {
		return organization, &models.ValidationError{Field: "id", Reason: "ID is missing"}
	}
	uid, err := gocql.ParseUUID(customerID)
	if err != nil {
		return organization, err
	}
	organization.Id = uid
	return organization, nil
}

func (h *Handler) unwrapDoc(from io.Reader) (*EmberData, error) {
	decoder := json.NewDecoder(from)
	var doc = &EmberData{}
	err := decoder.Decode(&doc)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
