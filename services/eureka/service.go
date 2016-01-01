package eureka

import (
	//"crypto/tls"
	//"expvar"

	"log"
	"net"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hudl/fargo"
	//"net/http"
	"os"
	//"strings"
)

// Service manages the listener and handler for an HTTP endpoint.
type Service struct {
	host     string
	config   string
	vip      string
	appName  string
	port     int
	err      chan error
	instance fargo.Instance
	eureka   fargo.EurekaConnection
	Logger   *log.Logger
	//statMap *expvar.Map
}

// NewService returns a new instance of Service.
func NewService(host string, port int, c Config) *Service {
	// Configure expvar monitoring. It's OK to do this even if the service fails to open and
	// should be done before any data could arrive for the service.
	//key := strings.Join([]string{"httpd", c.BindAddress}, ":")
	//tags := map[string]string{"bind": c.BindAddress}
	//statMap := influxdb.NewStatistics(key, "httpd", tags)
	strPort := strconv.Itoa(port)

	s := &Service{
		host:    host,
		port:    port,
		config:  c.ConfigPath,
		appName: c.AppName,
		vip:     c.VipAddress,
		err:     make(chan error),
		Logger:  log.New(os.Stderr, "[eureka] ", log.LstdFlags),
		instance: fargo.Instance{
			HostName:       host,
			Port:           port,
			App:            c.AppName,
			IPAddr:         host,
			VipAddress:     c.VipAddress,
			HomePageUrl:    "http://" + host + ":" + strPort + "/",
			StatusPageUrl:  "http://" + host + ":" + strPort + "/status",
			HealthCheckUrl: "http://" + host + ":" + strPort + "/healthcheck",
			DataCenterInfo: fargo.DataCenterInfo{Name: fargo.MyOwn},
			Status:         fargo.UP,
		},
	}
	// s.Handler.Logger = s.Logger
	return s
}

// Open starts the service
func (s *Service) Open() error {
	e, err := fargo.NewConnFromConfigFile(s.config)
	if err != nil {
		s.err <- err
		return err
	}
	s.eureka = e
	s.Logger.Println("Starting EUREKA discovery")

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, os.Interrupt)
	signal.Notify(exitChannel, syscall.SIGTERM)

	go func() {

		err = s.eureka.RegisterInstance(&s.instance)
		if err != nil {
			s.err <- err
		}

		for {
			err := s.eureka.HeartBeatInstance(&s.instance)
			if err != nil {
				s.err <- err
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

// Close closes the underlying listener.
func (s *Service) Close() error {

	err := s.eureka.DeregisterInstance(&s.instance)
	if err != nil {
		s.err <- err
	}
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
