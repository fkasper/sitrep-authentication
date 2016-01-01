package mongo

// Config represents a configuration for a HTTP service.
type Config struct {
	Enabled bool     `toml:"enabled"`
	Peers   []string `toml:"peers"`
}

// NewConfig returns a new Config with default settings.
func NewConfig() Config {
	return Config{
		Peers: []string{"127.0.0.1:27017"},
	}
}
