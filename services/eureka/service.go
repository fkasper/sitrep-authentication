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
	exit     chan int
	instance fargo.Instance
	eureka   fargo.EurekaConnection
	Logger   *log.Logger
	ticker   *time.Ticker
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
	ipAddr := GetLocalIP()
	ticker := time.NewTicker(time.Second * 10)
	hname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	s := &Service{
		host:    hname,
		port:    port,
		ticker:  ticker,
		config:  c.ConfigPath,
		appName: c.AppName,
		vip:     c.VipAddress,
		err:     make(chan error),
		Logger:  log.New(os.Stderr, "[eureka] ", log.LstdFlags),
		instance: fargo.Instance{
			HostName:       hname,
			Port:           port,
			App:            c.AppName,
			IPAddr:         ipAddr,
			VipAddress:     c.VipAddress,
			HomePageUrl:    "http://" + ipAddr + ":" + strPort + "/",
			StatusPageUrl:  "http://" + ipAddr + ":" + strPort + "/status",
			HealthCheckUrl: "http://" + ipAddr + ":" + strPort + "/healthcheck",
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
		return nil
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

		for _ = range s.ticker.C {
			err := s.eureka.HeartBeatInstance(&s.instance)
			if err != nil {
				s.err <- err
			}
		}
	}()
	return nil
}

// Close closes the underlying listener.
func (s *Service) Close() error {
	s.ticker.Stop()
	err := s.eureka.DeregisterInstance(&s.instance)
	if err != nil {
		s.err <- err
	}
	return nil
}

func GetLocalIP() string {
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
