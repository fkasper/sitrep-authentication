package httpd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/mattbaird/elastigo/lib"
	"github.com/vatcinc/bio/models"
	"github.com/vatcinc/bio/services/httpd"
)

const (
	dbURL = "mongodb://127.0.0.1:27017/bio_test"
	db    = "bio_test"
)

// func TestAll() {
//
// }

// GET /document
func TestHandler_Doc_Empty(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/document", nil))
	if w.Code != http.StatusNotAcceptable {
		t.Fatalf("unexpected status: %d", w.Code)
	} else if w.Body.String() != `{"error":"id parameter is not defined"}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandler_Doc_NotFound(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/document?id=SomeRandomIDthatcannotexistandisnotinhexformat", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", w.Code)
	} else if w.Body.String() != `{"error":"SomeRandomIDthatcannotexistandisnotinhexformat is not a valid HEX identifier"}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

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
func TestHandler_Doc_IndexEmpty(t *testing.T) {
	h := NewHandler(false)
	models.PrepareQuery(h.Mongo, "documents").RemoveAll(nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, MustNewJSONRequest("GET", "/api/v1.1/documents", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	} else if w.Body.String() != `null` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

// Upsert Document
func TestHandler_Doc_Create(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()
	doc := &models.Document{
		Type: "Test",
	}
	json, _ := json.Marshal(doc)
	b := bytes.NewReader(json)

	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

func TestHandler_Doc_Create_WithObjId(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()
	doc := &models.Document{
		ID:   bson.NewObjectId(),
		Type: "Test",
	}
	json, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Error while marshalling json")
	}
	b := bytes.NewReader(json)

	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

func TestHandler_Doc_Create_WithInvalidJSON(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()

	b := bytes.NewReader([]byte(`{"id":"abracadabra"-"type":"test"}`))

	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

func TestHandler_Doc_Create_WithEmptyType(t *testing.T) {
	h := NewHandler(false)

	w := httptest.NewRecorder()
	doc := &models.Document{
		ID: bson.NewObjectId(),
	}
	json, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Error while marshalling json")
	}
	b := bytes.NewReader(json)

	h.ServeHTTP(w, MustNewJSONRequest("PUT", "/api/v1.1/documents", b))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

//PRIVATE

type Handler struct {
	*httpd.Handler
}

// NewHandler returns a new instance of Handler.
func NewHandler(requireAuthentication bool) *Handler {
	h := &Handler{
		Handler: httpd.NewHandler(requireAuthentication, true, false),
	}
	dbSession, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	}
	elasticsearch := elastigo.NewConn()
	elasticsearch.SetFromUrl("http://127.0.0.1:9200")
	mongo := dbSession.DB(db)
	h.Handler.Mongo = mongo
	h.Handler.Elasticsearch = elasticsearch
	h.Handler.Version = "0.0.0"
	return h
}

func MustNewRequest(method, urlStr string, body io.Reader) *http.Request {
	r, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(err.Error())
	}
	return r
}

func MustNewJSONRequest(method, urlStr string, body io.Reader) *http.Request {
	r := MustNewRequest(method, urlStr, body)
	r.Header.Set("Content-Type", "application/json")
	return r
}
