package models_test

import (
	"testing"

	"github.com/fkasper/sitrep-biometrics/models"
)

func Test_Document_Structure(t *testing.T) {
	doc := &models.Document{}
	if doc == nil {
		t.Fatalf("no doc")
	}
}
