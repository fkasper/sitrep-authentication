package models

import (
	"regexp"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	biographyDbColumn = "biography"
)

// Biography defines a single Sub-Domain
type Biography struct {
	ID                  bson.ObjectId     `bson:"_id,omitempty" json:"id,omitempty"`
	MainImage           string            `bson:"main_image" json:"main_image"`
	LeftImage           string            `bson:"left_image" json:"left_image"`
	RightImage          string            `bson:"right_image" json:"right_image"`
	Name                string            `bson:"name" json:"name"`
	MapCenter           string            `bson:"map_center" json:"map_center"`
	BoundingBox         string            `bson:"bounding_box" json:"boundary_box"`
	Title               string            `bson:"title" json:"title"`
	Summary             string            `bson:"summary" json:"summary"`
	Nationality         string            `bson:"nationality" json:"nationality"`
	Religion            string            `bson:"religion" json:"religion"`
	Age                 string            `bson:"age" json:"age"`
	UDP                 []interface{}     `bson:"udps" json:"udp"`
	Gender              string            `bson:"gender" json:"gender"`
	Downloads           []interface{}     `bson:"downloads" json:"downloads"`
	Ethnicity           string            `bson:"ethnicity" json:"ethnicity"`
	FingerPrintsImages  map[string]string `bson:"fingerprints_images" json:"fingerprints_images"`
	FingerPrintsMinutia map[string]string `bson:"fingerprints_minutia" json:"fingerprints_minutia"`
	LeftIris            string            `bson:"left_iris" json:"left_iris"`
	RightIris           string            `bson:"right_iris" json:"right_iris"`
	DomainID            bson.ObjectId     `bson:"domain_id" json:"domain_id"`
}

// Biographies define a list of Domain
type Biographies []Biography

// Fetch returns a biogrphy!
func (b *Biography) Fetch(mongo *mgo.Database, domain *Domain, slug string) error {
	err := PrepareQuery(mongo, biographyDbColumn).Find(&bson.M{
		"_id":       bson.ObjectIdHex(slug),
		"domain_id": domain.ID,
	}).One(&b)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a biogrphy!
func (b *Biography) Delete(mongo *mgo.Database, domain *Domain, slug string) error {
	objID := bson.ObjectIdHex(slug)

	err := PrepareQuery(mongo, biographyDbColumn).Remove(&bson.M{
		"_id":       objID,
		"domain_id": domain.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

// Insert inserts a document into mongodb if it does not exist!
func (b *Biography) Insert(mongo *mgo.Database, domain *Domain) error {
	if len(b.ID) != 12 || !b.ID.Valid() {
		b.ID = bson.NewObjectId()
	}
	//return &doc, NewInvalidError(document.ID.String())
	// if b.Name == "" {
	// 	return NewInvalidError("Name is empty")
	// }

	if b.MainImage == "" {
		return NewInvalidError("Main Image")
	}
	b.DomainID = domain.ID

	err := PrepareQuery(mongo, biographyDbColumn).Insert(b)
	if err != nil {
		return err
	}
	return nil
}

// Update updates a document in mongodb
func (b *Biography) Update(mongo *mgo.Database, domain *Domain, updated *Biography, slug string) error {

	if err := b.Fetch(mongo, domain, slug); err != nil {
		return err
	}
	//return &doc, NewInvalidError(document.ID.String())
	// if b.Name == "" {
	// 	return NewInvalidError("Name is empty")
	// }

	if b.MainImage == "" {
		return NewInvalidError("Main Image")
	}
	updated.ID = b.ID
	_, err := PrepareQuery(mongo, biographyDbColumn).UpsertId(b.ID, updated)
	if err != nil {
		return err
	}
	return nil
}

// Slugify compiles a name into a valid route
func Slugify(runes string) string {
	var re = regexp.MustCompile("[^a-z0-9]+")
	return strings.Trim(re.ReplaceAllString(strings.ToLower(runes), "-"), "-")
}

// IndexBiographies returns all biographies for a given exercise
func IndexBiographies(mongo *mgo.Database, domain *Domain) (Biographies, error) {
	var bios Biographies
	err := PrepareQuery(mongo, biographyDbColumn).Find(&bson.M{"domain_id": domain.ID}).All(&bios)
	if err != nil {
		return bios, err
	}
	return bios, nil
}
