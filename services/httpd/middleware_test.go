package httpd_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// func TestAll() {
//
// }

// GET /document
func TestHandler_WithAuthDisabled(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/documents", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
func TestHandler_WithAuthEnabledWOCredentials(t *testing.T) {
	h := NewHandler(true)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/documents", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

// func TestHandler_TokenAuthInvalidPassword(t *testing.T) {
// 	h := NewHandler(true)
//
// 	w := httptest.NewRecorder()
// 	form := url.Values{}
// 	form.Set("email", "florian@xpandmmi.com")
// 	form.Set("password", "testpassword")
// 	form.Set("grant_type", "password")
// 	t.Log("Form: %v", form)
// 	req, _ := http.NewRequest("POST", "/api/v1.1/token", strings.NewReader(form.Encode()))
// 	h.ServeHTTP(w, req)
// 	if w.Code != http.StatusForbidden {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	}
// }
