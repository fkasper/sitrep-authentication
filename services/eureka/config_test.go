package eureka_test

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/vatcinc/bio/services/eureka"
)

func TestConfig_Parse(t *testing.T) {
	// Parse configuration.
	var c eureka.Config
	if _, err := toml.Decode(`
enabled = true
register-port = 8080
etcd-address = "http://127.0.0.1:4002"
register-by-host = false
registration-host = "google.com"
registration-name = "bioapp"
registration-path = "/"
registration-path-is-regex = false
`, &c); err != nil {
		t.Fatal(err)
	}

	// Validate configuration.
	if c.Enabled != true {
		t.Fatalf("unexpected enabled: %v", c.Enabled)
	} else if c.RegisterPort != 8080 {
		t.Fatalf("unexpected bind address: %s", c.RegisterPort)
	} else if c.ETCDAddress != "http://127.0.0.1:4002" {
		t.Fatalf("unexpected auth enabled: %v", c.ETCDAddress)
	} else if c.RegisterByHost != false {
		t.Fatalf("unexpected register field: %v", c.RegisterByHost)
	} else if c.RegistrationHost != "google.com" {
		t.Fatalf("unexpected registration host: %v", c.RegistrationHost)
	} else if c.RegistrationPath != "/" {
		t.Fatalf("unexpected registration host: %v", c.RegistrationPath)
	} else if c.RegistrationEndpointName != "bioapp" {
		t.Fatalf("unexpected registration host: %v", c.RegistrationEndpointName)
	} else if c.RegistrationPathIsRegex != false {
		t.Fatalf("unexpected registration host: %v", c.RegistrationPathIsRegex)
	}
}

// func TestConfig_WriteTracing(t *testing.T) {
// 	c := httpd.Config{WriteTracing: true}
// 	s := httpd.NewService(c)
// 	if !s.Handler.WriteTrace {
// 		t.Fatalf("write tracing was not set")
// 	}
// }
