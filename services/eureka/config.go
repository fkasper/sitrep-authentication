package eureka

// Config represents a configuration for a HTTP service.
type Config struct {
	Enabled    bool   `toml:"enabled"`
	AppName    string `toml:"app-name"`
	ConfigPath string `toml:"config-path"`
	VipAddress string `toml:"vip-address"`
}

// NewConfig returns a new Config with default settings.
func NewConfig() Config {
	return Config{
		Enabled:    true,
		AppName:    "bios-profilemgmt",
		ConfigPath: "/etc/eureka.gcfg",
		VipAddress: "profiles-management.cluster.sitrep-vatcinc.com",
	}
}
