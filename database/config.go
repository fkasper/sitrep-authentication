package database

//"github.com/vatcinc/bio/toml"

const (

	// DefaultCassandraKeyspace sets the keyspace for cassandra
	DefaultCassandraKeyspace = "bio"

	// DefaultCassandraConns sets the default connections for cassandra pool
	DefaultCassandraConns = 5
)

// Config represents the meta configuration.
type Config struct {
	CassandraKeyspace string   `toml:"cassandra-keyspace"`
	CassandraConns    int      `toml:"cassandra-num-connections"`
	CassandraNodes    []string `toml:"cassandra-peers"`
}

// NewConfig builds a new configuration with default values.
func NewConfig() *Config {
	return &Config{
		CassandraKeyspace: DefaultCassandraKeyspace,
		CassandraConns:    DefaultCassandraConns,
		CassandraNodes:    []string{"127.0.0.1"},
	}
}
