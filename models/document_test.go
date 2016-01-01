package models_test

import (
	"testing"

	"github.com/vatcinc/bio/models"
)

func Test_Document_Structure(t *testing.T) {
	doc := &models.Document{}
	if doc == nil {
		t.Fatalf("no doc")
	}
}
