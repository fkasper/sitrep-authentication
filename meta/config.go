package meta

import (
	"net"
	"time"

	//"github.com/fkasper/sitrep-authentication/toml"
)

const (
	// DefaultHostname is the default hostname if one is not provided.
	DefaultHostname = "localhost"

	// DefaultBindAddress is the default address to bind to.
	DefaultBindAddress = ":7717"

	// DefaultHeartbeatTimeout is the default heartbeat timeout for the store.
	DefaultHeartbeatTimeout = 1000 * time.Millisecond

	// DefaultElectionTimeout is the default election timeout for the store.
	DefaultElectionTimeout = 1000 * time.Millisecond

	// DefaultLeaderLeaseTimeout is the default leader lease for the store.
	DefaultLeaderLeaseTimeout = 500 * time.Millisecond

	// DefaultCommitTimeout is the default commit timeout for the store.
	DefaultCommitTimeout = 50 * time.Millisecond

	// DefaultRaftPromotionEnabled is the default for auto promoting a node to a raft node when needed
	DefaultRaftPromotionEnabled = true

	// DefaultLoggingEnabled determines if log messages are printed for the meta service
	DefaultLoggingEnabled = true

	// DefaultElasticSearchUrl sets the default elasticsearch contact point
	DefaultElasticSearchUrl = "http://localhost:9200"

	// DefaultInfluxDB defines the DB to use with influxDB
	DefaultInfluxDB = "authentication"

	// DefaultInfluxHost defines the host to use with influxDB
	DefaultInfluxHost = "http://127.0.0.1:8086"

	// DefaultInfluxUser defines the user to use with influxDB
	DefaultInfluxUser = "test"

	// DefaultInfluxPass defines the password to use with influxDB
	DefaultInfluxPass = "test"
)

// Config represents the meta configuration.
type Config struct {
	Dir              string `toml:"dir"`
	Hostname         string `toml:"hostname"`
	BindAddress      string `toml:"bind-address"`
	LoggingEnabled   bool   `toml:"logging-enabled"`
	ElasticSearchUrl string `toml:"elastic-search-url"`
	InfluxDB         string `toml:"influx-database"`
	InfluxHost       string `toml:"influx-hostname"`
	InfluxUser       string `toml:"influx-username"`
	InfluxPass       string `toml:"influx-password"`
	MongoAuthEnabled bool   `toml:"mongo-enable-auth"`
}

// NewConfig builds a new configuration with default values.
func NewConfig() *Config {
	return &Config{
		Hostname:         DefaultHostname,
		BindAddress:      getLocalIP(),
		LoggingEnabled:   DefaultLoggingEnabled,
		ElasticSearchUrl: DefaultElasticSearchUrl,
		InfluxDB:         DefaultInfluxDB,
		InfluxHost:       DefaultInfluxHost,
		InfluxUser:       DefaultInfluxUser,
		InfluxPass:       DefaultInfluxPass,
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
