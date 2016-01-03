package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	domainsDbColumn = "domains"
)

// Domain defines a single Sub-Domain
type Domain struct {
	ID              bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Hostname        string        `bson:"hostname"`
	NeverExpire     bool          `bson:"never_expire"`
	MaintenanceMode bool          `bson:"maintenance_mode"`
	StartDate       time.Time     `bson:"start_date"`
	EndDate         time.Time     `bson:"end_date"`
	ProjectLead     string        `bson:"project_lead"`
	ContactEmail    string        `bson:"contact_email"`
	MaxUsers        int           `bson:"max_users"`
	LicenseKey      string        `bson:"license_key"`
}

// Domains define a list of Domain
type Domains []Domain

// VirtualDomainCheck returns a domain!
func VirtualDomainCheck(mongo *mgo.Database, domain string, port string, materialized *Domain) error {
	err := PrepareQuery(mongo, domainsDbColumn).Find(&bson.M{"hostname": domain}).One(&materialized)
	if err != nil {
		return err
	}
	return nil
}

// Settings retreives settings for a domain
func (d *Domain) Settings(mongo *mgo.Database) (*Setting, error) {
	var settings Setting
	err := PrepareQuery(mongo, settingsDbColumn).Find(&bson.M{"domain_id": d.ID}).One(&settings)
	if err != nil {
		return InitDefaultSettingsForDomain(mongo, d)
	}
	return &settings, nil
}
