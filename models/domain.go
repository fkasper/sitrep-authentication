package models

import (
	"encoding/json"

	"github.com/gocql/gocql"
	"github.com/mattbaird/elastigo/lib"
	"github.com/vatcinc/bio/schema"
)

// VLINK defines a virtual link to a real domain.
var VLINK = bio.VlinkTableDef()

// DOMAINS represent a link between documents and organizational pages
var DOMAINS = bio.DomainsTableDef()

// OLD

// VirtualDomainCheck checks the database for an existing domain.
// TODO: Add caching here. This has to be a high-performance function
// MUSTN't slow down the request significally
func VirtualDomainCheck(cassandra *gocql.ClusterConfig, domain string, port string, m *bio.Domains) error {
	var vLink bio.Vlink
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(VLINK).
		Where(
		VLINK.DOMAIN_NAME.Eq(domain),
		VLINK.DOMAIN_PORT.Eq(port)).
		Into(
		VLINK.To(&vLink)).
		FetchOne(session)
	if err != nil {
		return err
	}
	_, err = ctx.Select().
		From(DOMAINS).
		Where(
		DOMAINS.ID.Eq(vLink.DomainId)).
		Into(
		DOMAINS.To(m)).
		FetchOne(session)
	if err != nil {
		return err
	}
	return nil
}

// Materialized

// ValidateDomain validates a domain for plausibility
func ValidateDomain(domain *bio.Domains) error {
	if domain.DomainName == "" {
		return &ValidationError{
			Field:  "DomainName",
			Reason: "is empty",
		}
	}
	if domain.Port == "" {
		return &ValidationError{
			Field:  "Port",
			Reason: "is empty!",
		}
	}
	if domain.Type == "" {
		return &ValidationError{
			Field:  "Type",
			Reason: "is empty!",
		}
	}
	return nil
}

// UpdateDomain updates an existing Domain
func UpdateDomain(cassandra *gocql.ClusterConfig, elastic *elastigo.Conn, domain *bio.Domains) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	if err := ctx.Upsert(DOMAINS).
		SetString(DOMAINS.DOMAIN_NAME, domain.DomainName).
		SetString(DOMAINS.ORGANIZATION_ID, domain.OrganizationId).
		SetString(DOMAINS.ORGANIZATION_NAME, domain.OrganizationName).
		SetString(DOMAINS.PORT, domain.Port).
		SetStringStringMap(DOMAINS.SETTINGS, domain.Settings).
		SetString(DOMAINS.TRANSPARENT_TARGET, domain.TransparentTarget).
		SetString(DOMAINS.TYPE, domain.Type).
		Where(
		DOMAINS.ID.Eq(domain.Id)).
		Exec(session); err != nil {
		return err
	}
	return nil
}

// InsertDomain creates a domain
func InsertDomain(cassandra *gocql.ClusterConfig, elastic *elastigo.Conn, domain *bio.Domains) error {
	if err := ValidateDomain(domain); err != nil {
		return err
	}
	var vLink bio.Vlink
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(VLINK).
		Where(
		VLINK.DOMAIN_NAME.Eq(domain.DomainName),
		VLINK.DOMAIN_PORT.Eq(domain.Port)).
		Into(
		VLINK.To(&vLink)).
		FetchOne(session)
	if err == nil {
		return &ValidationError{
			Field:  "domain/port",
			Reason: "combination already exists",
		}
	}

	err = ctx.Store(DOMAINS.Bind(*domain)).Exec(session)
	if err != nil {
		return err
	}

	id, err := gocql.ParseUUID(domain.OrganizationId)
	if err != nil {
		return err
	}
	vlink := bio.Vlink{
		DomainId:         domain.Id,
		DomainName:       domain.DomainName,
		DomainPort:       domain.Port,
		OrganizationId:   id,
		OrganizationName: domain.OrganizationName,
	}
	err = ctx.Store(VLINK.Bind(vlink)).Exec(session)
	if err != nil {
		return err
	}
	_, err = elastic.Index("bio", "domains", domain.Id.String(), nil, domain)
	if err != nil {
		return err
	}
	return nil
}

// GetDomain receives a domain object from cassandra
func GetDomain(cassandra *gocql.ClusterConfig, target *bio.Domains) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(DOMAINS).
		Where(
		DOMAINS.ID.Eq(target.Id)).
		Into(
		DOMAINS.To(target)).
		FetchOne(session)
	if err != nil {
		return err
	}
	return nil
}

// DeleteDomain removes a domain from Cassandra
func DeleteDomain(cassandra *gocql.ClusterConfig, elastic *elastigo.Conn, id gocql.UUID) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	err := ctx.Delete().
		From(DOMAINS).
		Where(
		DOMAINS.ID.Eq(id)).
		Exec(session)
	if err != nil {
		return err
	}
	_, err = elastic.Delete("bio", "domains", id.String(), nil)
	if err != nil {
		return err
	}
	return nil
}

// SearchDomains elasticsearch for a document
func SearchDomains(elastic *elastigo.Conn, query string, limit string) (*EmberMultiData, error) {
	search := &Search{
		Index:         "bio",
		Type:          "domains",
		Elasticsearch: elastic,
	}
	var result EmberMultiData
	searchJSON := `{
	  "from" : 0, "size" : ` + limit + `,
	  "query" : {
	    "multi_match" : {
	      "query":    "` + query + `",
	      "fields": [ "full-name" ],
	      "minimum_should_match": "50%"
	    }
	  }
	}`
	if err := search.Query(searchJSON, &result); err != nil {
		return &result, err
	}
	return &result, nil
}

// GetOrganizationDomains receives all domains for a specific organization From
// cassandra database
func GetOrganizationDomains(cassandra *gocql.ClusterConfig, domainID gocql.UUID, target *EmberMultiData) error {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	query, err := ctx.Select().
		From(DOMAINS).
		Where(
		DOMAINS.ORGANIZATION_ID.Eq(domainID.String())).Prepare(session)
	if err != nil {
		return err
	}
	mapper, err := MapDomainsEmber(query.Iter())
	if err != nil {
		return err
	}
	target.Data = mapper

	return nil
}

// MapDomainsEmber maps domains to an ember response format
func MapDomainsEmber(iter *gocql.Iter) ([]*EmberDataObj, error) {
	var array []*EmberDataObj
	err := bio.MapDomains(iter, func(t bio.Domains) (bool, error) {
		json, err := json.Marshal(t)
		if err != nil {
			return false, err
		}
		tmp := &EmberDataObj{
			Attributes: json,
			ID:         t.Id.String(),
			Type:       "domain",
		}
		array = append(array, tmp)
		return true, nil
	})
	return array, err
}
