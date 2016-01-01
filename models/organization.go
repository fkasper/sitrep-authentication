package models

import (
	"github.com/gocql/gocql"
	"github.com/mattbaird/elastigo/lib"
	"github.com/vatcinc/bio/schema"
)

// ORGANIZATIONS is a reference to the organizations table in cassandra
var ORGANIZATIONS = bio.OrganizationsTableDef()

// ValidateOrganization validates a domain for plausibility
func ValidateOrganization(organization *bio.Organizations) error {
	if organization.FullName == "" {
		return &ValidationError{
			Field:  "FullName",
			Reason: "is empty",
		}
	}
	if organization.Name == "" {
		return &ValidationError{
			Field:  "Name",
			Reason: "is empty!",
		}
	}
	return nil
}

// UpdateOrganization updates an existing Organization
func UpdateOrganization(cassandra *gocql.ClusterConfig, elastic *elastigo.Conn, organization *bio.Organizations) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	if err := ctx.Upsert(ORGANIZATIONS).
		SetTimestamp(ORGANIZATIONS.SLA_RESET, organization.SlaReset).
		SetString(ORGANIZATIONS.SLA_NAME, organization.SlaName).
		SetString(ORGANIZATIONS.SERVICE_CONTRACT_NAME, organization.ServiceContractName).
		SetString(ORGANIZATIONS.NAME, organization.Name).
		SetString(ORGANIZATIONS.FULL_NAME, organization.FullName).
		SetString(ORGANIZATIONS.CONTACT_PRIMARY_PHONE, organization.ContactPrimaryPhone).
		SetString(ORGANIZATIONS.CONTACT_PRIMARY_MOBILE, organization.ContactPrimaryMobile).
		SetString(ORGANIZATIONS.CONTACT_PRIMARY_FAX, organization.ContactPrimaryFax).
		SetString(ORGANIZATIONS.CONTACT_PRIMARY_EMAIL, organization.ContactPrimaryEmail).
		SetString(ORGANIZATIONS.BILLING_ADDRESS_ZIP, organization.BillingAddressZip).
		SetString(ORGANIZATIONS.BILLING_ADDRESS_STREET, organization.BillingAddressStreet).
		SetString(ORGANIZATIONS.BILLING_ADDRESS_STATE, organization.BillingAddressState).
		SetString(ORGANIZATIONS.BILLING_ADDRESS_COUNTRY, organization.BillingAddressCountry).
		SetString(ORGANIZATIONS.BILLING_ADDRESS_CITY, organization.BillingAddressCity).
		SetString(ORGANIZATIONS.BANNED_REASON, organization.BannedReason).
		SetInt32(ORGANIZATIONS.SLA_MINUTES_USED, organization.SlaMinutesUsed).
		SetInt32(ORGANIZATIONS.SLA_MINUTES_MAX, organization.SlaMinutesMax).
		SetInt32(ORGANIZATIONS.FLOAT_LICENSES, organization.FloatLicenses).
		SetInt32(ORGANIZATIONS.PAGES_LIMIT, organization.PagesLimit).
		SetUUID(ORGANIZATIONS.SLA_CONTRACT, organization.SlaContract).
		SetUUID(ORGANIZATIONS.SERVICE_CONTRACT, organization.ServiceContract).
		SetBoolean(ORGANIZATIONS.BILLABLE, organization.Billable).
		SetBoolean(ORGANIZATIONS.IS_ARCHIVED, organization.IsArchived).
		SetBoolean(ORGANIZATIONS.IS_BANNED, organization.IsBanned).
		Where(
		ORGANIZATIONS.ID.Eq(organization.Id)).
		Exec(session); err != nil {
		return err
	}
	if _, err := elastic.Update("bio", "organizations", organization.Id.String(), nil, organization); err != nil {
		return err
	}
	return nil
}

// InsertOrganization creates an organization
func InsertOrganization(cassandra *gocql.ClusterConfig, elastic *elastigo.Conn, organization *bio.Organizations) error {
	uid, err := gocql.RandomUUID()
	if err != nil {
		return err
	}
	organization.Id = uid
	if err := ValidateOrganization(organization); err != nil {
		return err
	}
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	if err := ctx.Store(ORGANIZATIONS.Bind(*organization)).Exec(session); err != nil {
		return err
	}
	if _, err := elastic.Index("bio", "organizations", organization.Id.String(), nil, organization); err != nil {
		return err
	}
	return nil
}

// GetOrganization receives an organization object from cassandra
func GetOrganization(cassandra *gocql.ClusterConfig, target *bio.Organizations) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(ORGANIZATIONS).
		Where(
		ORGANIZATIONS.ID.Eq(target.Id)).
		Into(
		ORGANIZATIONS.To(target)).
		FetchOne(session)
	if err != nil {
		return err
	}
	return nil
}

// DeleteOrganization ARCHIVES an organization in Cassandra, w/o deleting it!
func DeleteOrganization(cassandra *gocql.ClusterConfig, elastic *elastigo.Conn, id gocql.UUID) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	err := ctx.Upsert(DOMAINS).
		SetBoolean(ORGANIZATIONS.IS_ARCHIVED, true).
		Where(
		ORGANIZATIONS.ID.Eq(id)).
		Exec(session)
	if err != nil {
		return err
	}
	_, err = elastic.Delete("bio", "organizations", id.String(), nil)
	if err != nil {
		return err
	}
	return nil
}

// SearchOrganizations searches elasticsearch for a document
func SearchOrganizations(elastic *elastigo.Conn, query string, limit string) (*EmberMultiData, error) {
	search := &Search{
		Index:         "bio",
		Type:          "organizations",
		Elasticsearch: elastic,
	}
	var result EmberMultiData
	searchJSON := `{
	  "from" : 0, "size" : ` + limit + `,
	  "query" : {
	    "multi_match" : {
	      "query": "` + query + `",
	      "fields": [ "FullName^3", "Name^3" ],
	      "minimum_should_match": "50%"
	    }
	  }
	}`
	if err := search.Query(searchJSON, &result); err != nil {
		return &result, err
	}
	return &result, nil
}
