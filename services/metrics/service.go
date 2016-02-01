package metrics

import (
	//"crypto/tls"
	//"expvar"

	"log"
	"net"
	"os"
	"time"

	"github.com/fkasper/sitrep-authentication/meta"
	"github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

// Service manages the listener and handler for an HTTP endpoint.
type Service struct {
	Config *meta.Config
	err    chan error
	Logger *log.Logger
	//statMap *expvar.Map
}

// NewService returns a new instance of Service.
func NewService(c *meta.Config) *Service {

	s := &Service{
		err:    make(chan error),
		Config: c,
		Logger: log.New(os.Stderr, "[metrics] ", log.LstdFlags),
	}
	return s
}

// Open starts the service
func (s *Service) Open() error {
	s.Logger.Println("Starting Metrics reporting")

	go influxdb.InfluxDB(
		metrics.DefaultRegistry, // metrics registry
		time.Second*10,          // interval
		s.Config.InfluxHost,
		s.Config.InfluxDB,
		s.Config.InfluxUser,
		s.Config.InfluxPass,
	)
	return nil
}

// Close closes the underlying listener.
func (s *Service) Close() error {
	return nil
}

// SetLogger sets the internal logger to the logger passed in.
func (s *Service) SetLogger(l *log.Logger) {
	s.Logger = l
}

// Err returns a channel for fatal errors that occur on the listener.
func (s *Service) Err() <-chan error { return s.err }

// Addr returns the listener's address. Returns nil if listener is closed.
func (s *Service) Addr() net.Addr {
	return nil
}
