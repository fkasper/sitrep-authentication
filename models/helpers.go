package models

import (
	"github.com/gocql/gocql"
	"github.com/relops/cqlc/cqlc"
)

// WithSession adds a cassandra Session to the mix ;)
func WithSession(cassandra *gocql.ClusterConfig) (*gocql.Session, *cqlc.Context, error) {
	ctx := cqlc.NewContext()
	session, err := cassandra.CreateSession()
	return session, ctx, err
}
