package httpd

import (
	"encoding/json"
	"net/http"

	"github.com/fkasper/sitrep-biometrics/models"
	"github.com/yosssi/ace"
)

func (h *Handler) serveArcGISMap(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	w.Header().Add("content-type", "text/html")

	tpl, err := ace.Load("html/arcgis", "", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var settings *models.Setting
	if domain != nil {
		settings, err = domain.Settings(h.Mongo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var usr *models.LimitedPrintOutUser
	if user != nil {
		usr = user.LimitedReadOut()
	}

	data := map[string]interface{}{
		"IsAdmin":    true,
		"AppVersion": h.Version,
		"Settings": map[string]interface{}{
			"IsAdmin":          true,
			"DomainData":       domain,
			"ServerName":       "",
			"UserData":         usr,
			"SiteSettingsData": settings,
			"AppVersion":       h.Version,
		},
		"Renderer": map[string]interface{}{
			"Feature": h.Feature,
		},

		//"Biography": bioJson
	}
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) serveBiographyResult(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	w.Header().Add("content-type", "text/html")

	tpl, err := ace.Load("html/biography", "", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var settings *models.Setting
	if domain != nil {
		settings, err = domain.Settings(h.Mongo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var usr *models.LimitedPrintOutUser
	if user != nil {
		usr = user.LimitedReadOut()
	}

	data := map[string]interface{}{
		"Settings": map[string]interface{}{
			"DomainData":       domain,
			"ServerName":       "",
			"UserData":         usr,
			"SiteSettingsData": settings,
			"AppVersion":       h.Version,
		},
		"Renderer": map[string]interface{}{
			"Feature": h.Feature,
		},

		//"Biography": bioJson
	}
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) showBiography(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	var t models.Biography
	query := r.URL.Query()
	slug := query.Get("id")
	if slug == "" {
		httpError(w, "No id given", false, http.StatusNotFound)
		return
	}
	if err := t.Fetch(h.Mongo, domain, slug); err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(t, false))
}

func (h *Handler) deleteBiography(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	if err := user.CheckPermission("delete_biography"); err != nil {
		httpError(w, "You don't have permission to perform this action", false, http.StatusUnauthorized)
		return
	}
	var t models.Biography
	query := r.URL.Query()
	slug := query.Get("id")
	if slug == "" {
		httpError(w, "No id given", false, http.StatusNotFound)
		return
	}
	if err := t.Delete(h.Mongo, domain, slug); err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(t, false))
}

func (h *Handler) createBiography(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	if err := user.CheckPermission("create_biography"); err != nil {
		httpError(w, "You don't have permission to perform this action", false, http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var t models.Biography
	err := decoder.Decode(&t)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	if err := t.Insert(h.Mongo, domain); err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(t, false))
}
func (h *Handler) updateBiography(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	if err := user.CheckPermission("update_biography"); err != nil {
		httpError(w, "You don't have permission to perform this action", false, http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var target models.Biography
	var t models.Biography
	err := decoder.Decode(&t)
	query := r.URL.Query()
	slug := query.Get("id")
	if err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	if err := target.Update(h.Mongo, domain, &t, slug); err != nil {
		httpError(w, err.Error(), false, http.StatusBadRequest)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(t, false))
}

func (h *Handler) indexBiographies(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
	if err := user.CheckPermission("index_biography"); err != nil {
		httpError(w, "You don't have permission to perform this action", false, http.StatusUnauthorized)
		return
	}
	bios, err := models.IndexBiographies(h.Mongo, domain)
	if err != nil {
		httpError(w, err.Error(), false, http.StatusNotFound)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(bios, false))
}
