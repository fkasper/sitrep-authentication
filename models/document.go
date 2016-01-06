package models

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fkasper/sitrep-biometrics/schema"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	documentsDbColumn = "documents"
)

// Element defines a single Sub-Entity
type Element struct {
	Type          string                 `bson:"type" json:"type"`
	StyleID       string                 `bson:"styleId" json:"styleId"`
	ID            string                 `bson:"id" json:"id"`
	DataQuery     string                 `bson:"data_query" json:"data_query"`
	Skin          string                 `bson:"skin" json:"skin"`
	Layout        map[string]interface{} `bson:"layout" json:"layout"`
	Components    Elements               `bson:"components" json:"components,omitempty"`
	PropertyQuery string                 `bson:"propertyQuery" json:"propertyQuery"`
	ComponentType string                 `bson:"componentType" json:"componentType"`
}

// Elements define a list of Elements
type Elements []Element

// Document structure. See document-spec.md
type Document struct {
	ID         bson.ObjectId          `bson:"_id,omitempty" json:"id,omitempty"`
	Type       string                 `bson:"type" json:"type"`
	StyleID    string                 `bson:"styleId" json:"styleId"`
	Children   Elements               `json:"children" bson:",omitempty"`
	Properties map[string]interface{} `json:"properties" bson:"properties"`
	Data       map[string]interface{} `json:"data" bson:"data"`
	Styles     map[string]interface{} `json:"styles" bson:"styles"`
}

// CollapsedDocument is the public form of a document
type CollapsedDocument struct {
	ID       bson.ObjectId `json:"id"`
	Type     string        `json:"type"`
	StyleID  string        `json:"styleId"`
	Children Elements      `json:"children"`
}

// AggregatedData represents the data form of a public document
type AggregatedData struct {
	Properties map[string]interface{} `json:"component_properties"`
	Data       map[string]interface{} `json:"document_data"`
	Styles     map[string]interface{} `json:"theme_data"`
}

// ResponseDocument represents the collection of CollapsedDocument
// and AggregatedData
type ResponseDocument struct {
	Structure *CollapsedDocument `json:"structure"`
	Data      *AggregatedData    `json:"data"`
}

// NewResponseDocument creates a public Document
func NewResponseDocument(doc *Document) *ResponseDocument {
	return &ResponseDocument{
		Structure: &CollapsedDocument{
			ID:       doc.ID,
			Type:     doc.Type,
			StyleID:  doc.StyleID,
			Children: doc.Children,
		},
		Data: &AggregatedData{
			Properties: doc.Properties,
			Data:       doc.Data,
			Styles:     doc.Styles,
		},
	}
}

// InvalidDocumentError holds a specific error
type InvalidDocumentError struct {
	Message string
}

func (u *InvalidDocumentError) Error() string {
	return fmt.Sprintf("%v is not a valid HEX identifier", u.Message)
}

// NewInvalidDocumentError describes a new article
func NewInvalidDocumentError(objID string) *InvalidDocumentError {
	return &InvalidDocumentError{Message: objID}
}

// InvalidError holds a specific error
type InvalidError struct {
	Message string
}

func (u *InvalidError) Error() string {
	return fmt.Sprintf("%v is not valid", u.Message)
}

// NewInvalidError describes a new article
func NewInvalidError(mesg string) *InvalidError {
	return &InvalidError{Message: mesg}
}

// Documents define a list of Elements
type Documents []Document

// GetDocument receives a document from MONGODB.
func GetDocument(mongo *mgo.Database, id string) (*ResponseDocument, error) {
	var document Document
	var respDoc ResponseDocument
	if !bson.IsObjectIdHex(id) {
		return &respDoc, NewInvalidDocumentError(id)
	}
	objID := bson.ObjectIdHex(id)
	if !objID.Valid() {
		return &respDoc, NewInvalidDocumentError(id)
	}
	err := PrepareQuery(mongo, documentsDbColumn).FindId(objID).One(&document)
	if err != nil {
		return &respDoc, err
	}
	return NewResponseDocument(&document), nil
}

// GetDocuments lets find all documents inside a collection
func GetDocuments(mongo *mgo.Database) interface{} {

	var docs Documents
	iter := PrepareQuery(mongo, documentsDbColumn).Find(nil).Iter()

	if err := iter.All(&docs); err != nil {
		return err.Error()
	}

	if err := iter.Close(); err != nil {
		return err.Error()
	}
	return docs
}

// UpsertDocument lets you create new document or update an existing one
func UpsertDocument(mongo *mgo.Database, document *Document) (*ResponseDocument, error) {
	var repDoc ResponseDocument
	if len(document.ID) != 12 || !document.ID.Valid() {
		document.ID = bson.NewObjectId()
	}
	//return &doc, NewInvalidError(document.ID.String())
	if document.Type == "" {
		return &repDoc, NewInvalidError("Document type empty")
	}

	_, err := PrepareQuery(mongo, documentsDbColumn).UpsertId(document.ID, document)
	if err != nil {
		return &repDoc, err
	}
	return NewResponseDocument(document), nil
}

// WixFormatImport Imports a WIX Document
func WixFormatImport(mongo *mgo.Database, doc io.Reader) (*Document, error) {
	var document Document
	var respDoc ResponseDocument

	decoder := json.NewDecoder(doc)

	err := decoder.Decode(&respDoc)
	if err != nil {
		return &document, err
	}
	document.ID = respDoc.Structure.ID
	document.Type = respDoc.Structure.Type
	document.StyleID = respDoc.Structure.StyleID
	document.Children = respDoc.Structure.Children
	document.Properties = respDoc.Data.Properties
	document.Data = respDoc.Data.Data
	document.Styles = respDoc.Data.Styles
	newDoc, err := UpsertDocument(mongo, &document)
	if err != nil {
		return &document, err
	}
	document.ID = newDoc.Structure.ID
	return &document, nil
}

// Authorize authorizes a user against a document
func (d *Document) Authorize(mongo *mgo.Database, user *bio.Users) error {
	return nil
}
