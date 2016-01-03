package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	settingsDbColumn = "settings"
)

// Setting defines some settings
type Setting struct {
	ID          bson.ObjectId          `bson:"_id,omitempty" json:"id,omitempty"`
	DomainID    bson.ObjectId          `bson:"domain_id"`
	AuthEnabled bool                   `bson:"auth_enabled"`
	AdminUsers  []bson.ObjectId        `bson:"admin_users"`
	SettingsObj map[string]interface{} `bson:"settings_obj"`
	Features    []bson.ObjectId        `bson:"features"`
}

// Settings define a list of Setting
type Settings []Setting

// InitDefaultSettingsForDomain creates new settings
func InitDefaultSettingsForDomain(mongo *mgo.Database, domain *Domain) (*Setting, error) {
	defaults := &Setting{
		ID:          bson.NewObjectId(),
		DomainID:    domain.ID,
		AuthEnabled: false,
		SettingsObj: map[string]interface{}{
			"logo": "/api/v2/img/logo",
		},
		Features: []bson.ObjectId{},
	}
	err := PrepareQuery(mongo, settingsDbColumn).Insert(defaults)
	if err != nil {
		return defaults, err
	}
	return defaults, nil
}
