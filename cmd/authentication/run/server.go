package run

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/fkasper/sitrep-authentication/meta"
	"github.com/fkasper/sitrep-authentication/services/httpd"
	"github.com/fkasper/sitrep-authentication/services/metrics"
	"github.com/gocql/gocql"
	elastigo "github.com/mattbaird/elastigo/lib"
	regmeta "github.com/xpandmmi/registrator/meta"
	"github.com/xpandmmi/registrator/services/registration"
)

// BuildInfo represents the build details for the server code.
type BuildInfo struct {
	Version string
	Commit  string
	Branch  string
	Time    string
}

// Server represents a container for the metadata and storage data and services.
// It is built using a Config and it manages the startup and shutdown of all
// services in the proper order.
type Server struct {
	buildInfo BuildInfo

	err     chan error
	closing chan struct{}

	Hostname    string
	BindAddress string
	Listener    net.Listener

	Services []Service

	// These references are required for the tcp muxer.

	//Monitor *monitor.Monitor

	// Server reporting and registration
	reportingDisabled bool

	// Profiling
	CPUProfile string
	MemProfile string
	up         chan bool

	// Database
	elasticsearch *elastigo.Conn
	cassandra     *gocql.ClusterConfig
}

// NewServer returns a new instance of Server built from a config.
func NewServer(c *Config, buildInfo *BuildInfo) (*Server, error) {

	elasticsearch := elastigo.NewConn()
	elasticsearch.SetFromUrl(c.Meta.ElasticSearchUrl)

	db := gocql.NewCluster(c.Database.CassandraNodes...)
	db.Keyspace = c.Database.CassandraKeyspace
	db.NumConns = c.Database.CassandraConns
	db.Discovery = gocql.DiscoveryConfig{
		DcFilter:   "",
		RackFilter: "",
		Sleep:      30 * time.Second,
	}

	s := &Server{
		buildInfo: *buildInfo,
		err:       make(chan error),
		closing:   make(chan struct{}),

		Hostname:      c.Meta.Hostname,
		BindAddress:   c.Meta.BindAddress,
		elasticsearch: elasticsearch,
		cassandra:     db,
	}

	// Append services.
	//s.appendMongoService(c.Mongo)

	s.appendMetricsReportingService(c.Meta)
	s.appendHTTPDService(c.HTTPD)
	s.appendRegistrationService(c.Registration, c.RegMeta)
	return s, nil
}

func (s *Server) appendRegistrationService(c registration.Config, meta *regmeta.Config) {
	srv := registration.NewService(c, meta, false)
	srv.Up = s.up
	s.Services = append(s.Services, srv)
}

func (s *Server) appendMetricsReportingService(c *meta.Config) {
	srv := metrics.NewService(c)
	srv.Config = c
	s.Services = append(s.Services, srv)
}

func (s *Server) appendHTTPDService(c httpd.Config) {
	if !c.Enabled {
		return
	}
	srv := httpd.NewService(c)
	srv.Handler.Version = s.buildInfo.Version
	srv.Handler.Elasticsearch = s.elasticsearch
	srv.Handler.Cassandra = s.cassandra
	s.Services = append(s.Services, srv)
}

// Err returns an error channel that multiplexes all out of band errors received from all services.
func (s *Server) Err() <-chan error { return s.err }

// Open opens the meta and data store and all services.
func (s *Server) Open() error {
	if err := func() error {
		// Start profiling, if set.
		startProfile(s.CPUProfile, s.MemProfile)

		for _, service := range s.Services {
			if err := service.Open(); err != nil {
				return fmt.Errorf("open service: %s", err)
			}
			go s.monitorErrorChan(service.Err())
		}
		s.up <- true

		return nil

	}(); err != nil {
		s.Close()
		return err
	}

	return nil
}

func (s *Server) monitorErrorChan(err <-chan error) {
	for n := range err {
		s.err <- n
	}
}

// Close shuts down the meta and data stores and all services.
func (s *Server) Close() error {
	stopProfile()

	// Close services to allow any inflight requests to complete
	// and prevent new requests from being accepted.
	for _, service := range s.Services {
		service.Close()
	}

	close(s.closing)
	return nil
}

// hostAddr returns the host and port that remote nodes will use to reach this
// node.
func (s *Server) hostAddr() (string, string, error) {
	// Resolve host to address.
	_, port, err := net.SplitHostPort(s.BindAddress)
	if err != nil {
		return "", "", fmt.Errorf("split bind address: %s", err)
	}

	host := s.Hostname

	// See if we might have a port that will override the BindAddress port
	if host != "" && host[len(host)-1] >= '0' && host[len(host)-1] <= '9' && strings.Contains(host, ":") {
		hostArg, portArg, err := net.SplitHostPort(s.Hostname)
		if err != nil {
			return "", "", err
		}

		if hostArg != "" {
			host = hostArg
		}

		if portArg != "" {
			port = portArg
		}
	}
	return host, port, nil
}

// Service represents a service attached to the server.
type Service interface {
	Open() error
	Close() error
	Err() <-chan error
}

// prof stores the file locations of active profiles.
var prof struct {
	cpu *os.File
	mem *os.File
}

// StartProfile initializes the cpu and memory profile, if specified.
func startProfile(cpuprofile, memprofile string) {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("cpuprofile: %v", err)
		}
		log.Printf("writing CPU profile to: %s\n", cpuprofile)
		prof.cpu = f
		pprof.StartCPUProfile(prof.cpu)
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("memprofile: %v", err)
		}
		log.Printf("writing mem profile to: %s\n", memprofile)
		prof.mem = f
		runtime.MemProfileRate = 4096
	}

}

// StopProfile closes the cpu and memory profiles if they are running.
func stopProfile() {
	if prof.cpu != nil {
		pprof.StopCPUProfile()
		prof.cpu.Close()
		log.Println("CPU profile stopped")
	}
	if prof.mem != nil {
		pprof.Lookup("heap").WriteTo(prof.mem, 0)
		prof.mem.Close()
		log.Println("mem profile stopped")
	}
}

type tcpaddr struct{ host string }

func (a *tcpaddr) Network() string { return "tcp" }
func (a *tcpaddr) String() string  { return a.host }
