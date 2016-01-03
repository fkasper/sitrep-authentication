package run

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/vatcinc/bio/meta"
	"github.com/vatcinc/bio/services/eureka"
	"github.com/vatcinc/bio/services/httpd"
	"github.com/vatcinc/bio/services/metrics"
	"gopkg.in/mgo.v2"
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

	// Database
	mongo         *mgo.Database
	elasticsearch *elastigo.Conn
	cassandra     *gocql.ClusterConfig
}

// NewServer returns a new instance of Server built from a config.
func NewServer(c *Config, buildInfo *BuildInfo) (*Server, error) {

	dbSession, err := mgo.Dial(c.Meta.MongoUrl)
	if err != nil {
		panic(err)
	}
	dbSession.SetMode(mgo.Monotonic, true)
	if c.Meta.MongoAuthEnabled {
		cred := &mgo.Credential{
			Username: c.Meta.MongoUser,
			Password: c.Meta.MongoPass,
		}
		if err := dbSession.Login(cred); err != nil {
			panic(err)
		}
	}
	mongo := dbSession.DB(c.Meta.MongoDbName)

	elasticsearch := elastigo.NewConn()
	elasticsearch.SetFromUrl(c.Meta.ElasticSearchUrl)

	db := gocql.NewCluster(c.Database.CassandraNodes...)
	db.Keyspace = c.Database.CassandraKeyspace
	db.NumConns = c.Database.CassandraConns
	db.DiscoverHosts = true
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
		mongo:         mongo,
		elasticsearch: elasticsearch,
		cassandra:     db,
	}

	// Append services.
	//s.appendMongoService(c.Mongo)

	s.appendeurekaService(c.Eureka)
	s.appendMetricsReportingService(c.Meta)

	s.appendHTTPDService(c.HTTPD)
	return s, nil
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
	srv.Handler.Mongo = s.mongo
	srv.Handler.Elasticsearch = s.elasticsearch
	srv.Handler.Cassandra = s.cassandra
	s.Services = append(s.Services, srv)
}

func (s *Server) appendeurekaService(c eureka.Config) {
	if !c.Enabled {
		return
	}

	host, port, err := s.hostAddr()
	if err != nil {
		s.err <- fmt.Errorf("Got an error while fetching host addr=%s", err)
	}

	regiport, err := strconv.Atoi(port)
	if err != nil {
		s.err <- fmt.Errorf(err.Error())
	}

	srv := eureka.NewService(host, regiport, c)
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
		}

		return nil

	}(); err != nil {
		s.Close()
		return err
	}

	return nil
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
