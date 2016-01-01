package httpd_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// const (
// 	dbURL = "mongodb://127.0.0.1:27017/bio_test"
// 	db    = "bio_test"
// )
//
// // func TestAll() {
// //
// // }
//
// // GET /document
// func TestHandler_Doc_Empty(t *testing.T) {
// 	h := NewHandler(false)
//
// 	w := httptest.NewRecorder()
// 	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/document", nil))
// 	if w.Code != http.StatusNotAcceptable {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	} else if w.Body.String() != `{"error":"id parameter is not defined"}` {
// 		t.Fatalf("unexpected body: %s", w.Body.String())
// 	}
// }
//
// func TestHandler_Doc_NotFound(t *testing.T) {
// 	h := NewHandler(false)
//
// 	w := httptest.NewRecorder()
// 	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/document?id=SomeRandomIDthatcannotexistandisnotinhexformat", nil))
// 	if w.Code != http.StatusNotFound {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	} else if w.Body.String() != `{"error":"SomeRandomIDthatcannotexistandisnotinhexformat is not a valid HEX identifier"}` {
// 		t.Fatalf("unexpected body: %s", w.Body.String())
// 	}
// }

// func TestHandler_Doc_Found(t *testing.T) {
// 	h := NewHandler(false)
// 	id := bson.NewObjectId()
// 	element := models.Element{
// 		Type:          "Component",
// 		StyleID:       "1234",
// 		ID:            "1234",
// 		DataQuery:     "",
// 		Skin:          "",
// 		Layout:        map[string]interface{}{"test": "test"},
// 		PropertyQuery: "test",
// 		ComponentType: "1234",
// 	}
// 	doc := models.Document{
// 		ID:         id,
// 		Type:       "Document",
// 		StyleID:    "Test_1",
// 		Properties: map[string]interface{}{"Test_1": "test"},
// 		Data:       map[string]interface{}{"Test_1": "test"},
// 		Styles:     map[string]interface{}{"Test_1": "test"},
// 		Children:   []models.Element{element},
// 	}
//
// 	_, err := h.mongo.C("documents").UpsertId(id, &doc)
// 	t.Fatalf("Found Doc: %v", doc)
// 	return
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// 	w := httptest.NewRecorder()
// 	h.ServeHTTP(w, MustNewJSONRequest("GET", fmt.Sprintf("/api/v1.1/document?id=%s", id), nil))
// 	if w.Code != http.StatusOK {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	} else if w.Body.String() != `{"error":"SomeRandomIDthatcannotexistandisnotinhexformat is not a valid HEX identifier"}` {
// 		t.Fatalf("unexpected body: %s", w.Body.String())
// 	}
// }

// Index Documents
func TestHandler_Org_SearchIndexQueryNonExisting(t *testing.T) {
	h := NewHandler(false)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/organizations?query=empty_for_sure", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	} else if w.Body.String() != `{"data":null}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandler_Org_SearchIndexQueryEmpty(t *testing.T) {
	h := NewHandler(false)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/organizations", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	} else if w.Body.String() != `{"data":null}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandler_Org_SearchIndexLimitSet(t *testing.T) {
	h := NewHandler(false)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/organizations?limit=30", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	} else if w.Body.String() != `{"data":null}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

//
// // Upsert Document
// func TestHandler_Doc_Create(t *testing.T) {
// 	h := NewHandler(false)
//
// 	w := httptest.NewRecorder()
// 	doc := &models.Document{
// 		Type: "Test",
// 	}
// 	json, _ := json.Marshal(doc)
// 	b := bytes.NewReader(json)
//
// 	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
// 	if w.Code != http.StatusOK {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	}
// }
//
// func TestHandler_Doc_Create_WithObjId(t *testing.T) {
// 	h := NewHandler(false)
//
// 	w := httptest.NewRecorder()
// 	doc := &models.Document{
// 		ID:   bson.NewObjectId(),
// 		Type: "Test",
// 	}
// 	json, err := json.Marshal(doc)
// 	if err != nil {
// 		t.Fatalf("Error while marshalling json")
// 	}
// 	b := bytes.NewReader(json)
//
// 	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
// 	if w.Code != http.StatusOK {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	}
// }
//
// func TestHandler_Doc_Create_WithInvalidJSON(t *testing.T) {
// 	h := NewHandler(false)
//
// 	w := httptest.NewRecorder()
//
// 	b := bytes.NewReader([]byte(`{"id":"abracadabra"-"type":"test"}`))
//
// 	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
// 	if w.Code != http.StatusInternalServerError {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	}
// }
//
// func TestHandler_Doc_Create_WithEmptyType(t *testing.T) {
// 	h := NewHandler(false)
//
// 	w := httptest.NewRecorder()
// 	doc := &models.Document{
// 		ID: bson.NewObjectId(),
// 	}
// 	json, err := json.Marshal(doc)
// 	if err != nil {
// 		t.Fatalf("Error while marshalling json")
// 	}
// 	b := bytes.NewReader(json)
//
// 	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
// 	if w.Code != http.StatusBadRequest {
// 		t.Fatalf("unexpected status: %d", w.Code)
// 	}
// }
