package httpd

//
// func (h *Handler) serveArcGISMap(w http.ResponseWriter, r *http.Request, domain *models.Domain, user *models.User) {
// 	w.Header().Add("content-type", "text/html")
//
// 	tpl, err := ace.Load("html/arcgis", "", nil)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	var settings *models.Setting
// 	if domain != nil {
// 		settings, err = domain.Settings(h.Mongo)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
//
// 	var usr *models.LimitedPrintOutUser
// 	if user != nil {
// 		usr = user.LimitedReadOut()
// 	}
//
// 	data := map[string]interface{}{
// 		"IsAdmin":    true,
// 		"AppVersion": h.Version,
// 		"Settings": map[string]interface{}{
// 			"IsAdmin":          true,
// 			"DomainData":       domain,
// 			"ServerName":       "",
// 			"UserData":         usr,
// 			"SiteSettingsData": settings,
// 			"AppVersion":       h.Version,
// 		},
// 		"Renderer": map[string]interface{}{
// 			"Feature": h.Feature,
// 		},
//
// 		//"authenticationgraphy": authenticationJson
// 	}
// 	if err := tpl.Execute(w, data); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
