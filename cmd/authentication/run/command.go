package run

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

const logo = `
 _____           _
|_   _|         | |                 ®
  | |_   _ _ __ | |__   ___  _ __
  | | | | | '_ \| '_ \ / _ \| '_ \
  | | |_| | |_) | | | | (_) | | | |
  \_/\__, | .__/|_| |_|\___/|_| |_|
      __/ | |
     |___/|_|
  by XPANDMMI

`

// Command represents the command executed by "authentication run".
type Command struct {
	Version   string
	Branch    string
	Commit    string
	BuildTime string

	closing chan struct{}
	Closed  chan struct{}

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Server *Server
}

// NewCommand return a new instance of Command.
func NewCommand() *Command {
	return &Command{
		closing: make(chan struct{}),
		Closed:  make(chan struct{}),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}

// Run parses the config from args and runs the server.
func (cmd *Command) Run(args ...string) error {
	// Parse the command line flags.
	options, err := cmd.ParseFlags(args...)
	if err != nil {
		return err
	}
	fmt.Print(logo)
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Mark start-up in log.
	log.Printf("authentication starting, version %s, branch %s, commit %s, built %s",
		cmd.Version, cmd.Branch, cmd.Commit, cmd.BuildTime)
	log.Printf("Go version %s, GOMAXPROCS set to %d", runtime.Version(), runtime.GOMAXPROCS(0))

	// Write the PID file.
	if err := cmd.writePIDFile(options.PIDFile); err != nil {
		return fmt.Errorf("write pid file: %s", err)
	}

	// Turn on block profiling to debug stuck databases
	runtime.SetBlockProfileRate(int(1 * time.Second))

	// Parse config
	config, err := cmd.ParseConfig(options.ConfigPath)
	if err != nil {
		return fmt.Errorf("parse config: %s", err)
	}

	// Apply any environment variables on top of the parsed config
	if err := config.ApplyEnvOverrides(); err != nil {
		return fmt.Errorf("apply env config: %v", err)
	}

	// Override config hostname if specified in the command line args.
	if options.Hostname != "" {
		config.Meta.Hostname = options.Hostname
	}

	// Validate the configuration.
	if err := config.Validate(); err != nil {
		return fmt.Errorf("%s. To generate a valid configuration file run `authentication config > authentication.generated.conf`", err)
	}

	// Create server from config and start it.
	buildInfo := &BuildInfo{
		Version: cmd.Version,
		Commit:  cmd.Commit,
		Branch:  cmd.Branch,
		Time:    cmd.BuildTime,
	}
	s, err := NewServer(config, buildInfo)
	if err != nil {
		return fmt.Errorf("create server: %s", err)
	}
	s.CPUProfile = options.CPUProfile
	s.MemProfile = options.MemProfile
	if err := s.Open(); err != nil {
		return fmt.Errorf("open server: %s", err)
	}
	cmd.Server = s

	// Begin monitoring the server's error channel.
	go cmd.monitorServerErrors()

	return nil
}

// Close shuts down the server.
func (cmd *Command) Close() error {
	defer close(cmd.Closed)
	close(cmd.closing)
	if cmd.Server != nil {
		return cmd.Server.Close()
	}
	return nil
}

func (cmd *Command) monitorServerErrors() {
	logger := log.New(cmd.Stderr, "", log.LstdFlags)
	for {
		select {
		case err := <-cmd.Server.Err():
			logger.Println(err)
		case <-cmd.closing:
			return
		}
	}
}

// ParseFlags parses the command line flags from args and returns an options set.
func (cmd *Command) ParseFlags(args ...string) (Options, error) {
	var options Options
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&options.ConfigPath, "config", "", "")
	fs.StringVar(&options.PIDFile, "pidfile", "", "")
	fs.StringVar(&options.Hostname, "hostname", "", "")
	fs.StringVar(&options.CPUProfile, "cpuprofile", "", "")
	fs.StringVar(&options.MemProfile, "memprofile", "", "")
	fs.Usage = func() { fmt.Fprintln(cmd.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	return options, nil
}

// writePIDFile writes the process ID to path.
func (cmd *Command) writePIDFile(path string) error {
	// Ignore if path is not set.
	if path == "" {
		return nil
	}

	// Ensure the required directory structure exists.
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return fmt.Errorf("mkdir: %s", err)
	}

	// Retrieve the PID and write it.
	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(path, []byte(pid), 0666); err != nil {
		return fmt.Errorf("write file: %s", err)
	}

	return nil
}

// ParseConfig parses the config at path.
// Returns a demo configuration if path is blank.
func (cmd *Command) ParseConfig(path string) (*Config, error) {
	// Use demo configuration if no config path is specified.
	if path == "" {
		log.Println("no configuration provided, using default settings")
		return NewDemoConfig()
	}

	log.Printf("Using configuration at: %s\n", path)

	config := NewConfig()
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return config, nil
}

var usage = `usage: run [flags]
run starts the broker and data node server.
        -config <path>
                          Set the path to the configuration file.
        -hostname <name>
                          Override the hostname, the 'hostname' configuration
                          option will be overridden.
        -pidfile <path>
                          Write process ID to a file.
`

// Options represents the command line options that can be parsed.
type Options struct {
	ConfigPath string
	PIDFile    string
	Hostname   string
	CPUProfile string
	MemProfile string
}
