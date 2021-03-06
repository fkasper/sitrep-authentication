package httpd

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/bmizerany/pat"
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
	"github.com/mattbaird/elastigo/lib"
	"github.com/rcrowley/go-metrics"
	"gopkg.in/mgo.v2"
	// "github.com/gorilla/websocket"
)

const (
	// DefaultChunkSize specifies the amount of data mappers will read
	// up to, before sending results back to the engine. This is the
	// default size in the number of values returned in a raw query.
	//
	// Could be many more bytes depending on fields returned.
	DefaultChunkSize = 10000
)

// TODO: Standard response headers (see: HeaderHandler)
// TODO: Compression (see: CompressionHeaderHandler)

// TODO: Check HTTP response codes: 400, 401, 403, 409.

type route struct {
	name        string
	method      string
	pattern     string
	gzipped     bool
	log         bool
	handlerFunc interface{}
}

// Feature describes additional beta and rollback features in a component
// livecycle
type Feature struct {
	ID                      string
	Name                    string
	GlobPermissions         bool
	RequiredPermissionLevel string
	GlobAvailable           bool
}

// Handler represents an HTTP handler for the InfluxDB server.
type Handler struct {
	mux                   *pat.PatternServeMux
	requireAuthentication bool
	Version               string

	Logger         *log.Logger
	loggingEnabled bool // Log every HTTP access.
	WriteTrace     bool // Detailed logging of write path
	Mongo          *mgo.Database
	Elasticsearch  *elastigo.Conn
	Cassandra      *gocql.ClusterConfig
	statMap        metrics.Registry
	Feature        *Feature
	//statMap        *expvar.Map
}

// NewHandler returns a new instance of handler with routes.
func NewHandler(requireAuthentication, loggingEnabled, writeTrace bool) *Handler {
	// c := metrics.NewCounter()
	// metrics.Register(statRequest, c)
	// c1 := metrics.NewCounter()
	// metrics.Register(statAuthFail, c1)
	h := &Handler{
		mux: pat.New(),
		requireAuthentication: requireAuthentication,
		Logger:                log.New(os.Stderr, "[http] ", log.LstdFlags),
		loggingEnabled:        loggingEnabled,
		WriteTrace:            writeTrace,
		statMap:               metrics.DefaultRegistry,
		Feature: &Feature{
			ID:                      "nyi",
			Name:                    "demo-feature",
			GlobPermissions:         true,
			RequiredPermissionLevel: "user",
			GlobAvailable:           true,
		},
	}

	h.SetRoutes([]route{
		route{
			"authentication_login-route",
			"POST", "/apis/authentication/login", true, true, h.authenticationLoginService,
		},
		route{
			"profiles-self",
			"GET", "/apis/authentication/me", true, true, h.receiveOwnProfileService,
		},
		route{
			"exercises-self",
			"GET", "/apis/authentication/exercises", true, true, h.authenticationGetExercisesService,
		},
		route{
			"exercises-settings-receive",
			"GET", "/apis/authentication/current-exercise-settings", true, true, h.authenticationGetExercisesSettings,
		},
		route{
			"exercises-settings-update",
			"PUT", "/apis/authentication/current-exercise-settings", true, true, h.authenticationUpdateExercisesSettings,
		},
		route{
			"exercises-current-permissions",
			"GET", "/apis/authentication/exercise-permissions", true, true, h.authenticationGetCurrentExercisePermissions,
		},
		route{
			"change-my-password",
			"POST", "/apis/authentication/change-password", true, true, h.authenticationPasswordChangeService,
		},
		route{
			"exercises-users-list",
			"GET", "/apis/authentication/user-list", true, true, h.getUsersList,
		},
		route{
			"authentication_options-route",
			"OPTIONS", "/apis/authentication/:option", true, false, h.serveOptions,
		},
		route{
			"healthcheck",
			"GET", "/healthcheck", true, true, h.serveHealthcheck,
		},
		route{
			"status", // Query serving route.
			"GET", "/status", true, true, h.serveHealthcheck,
		},
	})

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//h.statMap.Add(statRequest, 1)

	counter := metrics.GetOrRegisterCounter(statRequest, h.statMap)
	counter.Inc(1)

	meter := metrics.GetOrRegisterMeter(statRequestNew, h.statMap)
	meter.Mark(1)

	// FIXME(benbjohnson): Add pprof enabled flag.
	if strings.HasPrefix(r.URL.Path, "/debug/pprof") {
		switch r.URL.Path {
		case "/debug/pprof/cmdline":
			pprof.Cmdline(w, r)
		case "/debug/pprof/profile":
			pprof.Profile(w, r)
		case "/debug/pprof/symbol":
			pprof.Symbol(w, r)
		default:
			pprof.Index(w, r)
		}
	} else {
		h.mux.ServeHTTP(w, r)
		return
	}

}

// SetRoutes sets the provided routes on the handler.
func (h *Handler) SetRoutes(routes []route) {
	for _, r := range routes {
		var handler http.Handler
		// If it's a handler func that requires a domain, wrap it in a domain :lol:
		// if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *models.Domain)); ok {
		// 	handler = materializeDomain(hf, h)
		// }

		// If it's a handler func that requires authorization, wrap it in authorization
		if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *sitrep.UsersByEmail)); ok {
			handler = authenticate(hf, h, h.requireAuthentication)
		}

		if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *sitrep.UsersByEmail, *sitrep.ExerciseByIdentifier)); ok {
			handler = exercisify(hf, h, h.requireAuthentication)
		}

		if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *sitrep.ExerciseByIdentifier)); ok {
			handler = exercisifyOnly(hf, h)
		}
		// If it's a handler func that requires authorization, wrap it in authorization
		// if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *models.User)); ok {
		// 	handler = authenticate(hf, h, h.requireAuthentication)
		// }
		// This is a normal handler signature and does not require authorization
		if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request)); ok {
			handler = http.HandlerFunc(hf)
		}

		if r.gzipped {
			handler = gzipFilter(handler)
		}
		handler = versionHeader(handler, h)
		handler = cors(handler)
		handler = requestID(handler)

		if h.loggingEnabled && r.log {
			handler = logging(handler, r.name, h.Logger)
		}

		handler = recovery(handler, r.name, h.Logger) // make sure recovery is always last

		h.mux.Add(r.method, r.pattern, handler)
	}
}

func (h *Handler) serveOptions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// RootAPIResult describes the API Result of the Root Document
type RootAPIResult struct {
	AppName      string                   `json:"app"`
	Version      string                   `json:"version"`
	AllowedPaths []map[string]interface{} `json:"paths"`
}

func (h *Handler) serveHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	res := map[string]string{"status": "ok"}
	w.Write(MarshalJSON(res, false))
}

// MarshalJSON will marshal v to JSON. Pretty prints if pretty is true.
func MarshalJSON(v interface{}, pretty bool) []byte {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(v, "", "    ")
	} else {
		b, err = json.Marshal(v)
	}

	if err != nil {
		return []byte(err.Error())
	}
	return b
}

// Filters and filter helpers

// Response represents a list of statement results.
type Response struct {
	Results []interface{} `json:"results"`
	Err     string        `json:"error"`
}

// MarshalJSON encodes a Response struct into JSON.
func (r Response) MarshalJSON() ([]byte, error) {
	// Define a struct that outputs "error" as a string.
	var o struct {
		Results []interface{} `json:"results,omitempty"`
		Err     string        `json:"error,omitempty"`
	}

	// Copy fields to output struct.
	o.Results = r.Results
	// if r.Err != nil {
	// 	o.Err = r.Err.Error()
	// }

	return json.Marshal(&o)
}

// UnmarshalJSON decodes the data into the Response struct
func (r *Response) UnmarshalJSON(b []byte) error {
	var o struct {
		Results []interface{} `json:"results,omitempty"`
		Err     string        `json:"error,omitempty"`
	}

	err := json.Unmarshal(b, &o)
	if err != nil {
		return err
	}
	r.Results = o.Results
	if o.Err != "" {
		r.Err = o.Err
	}
	return nil
}

// Error returns the first error from any statement.
// Returns nil if no errors occurred on any statements.
func (r *Response) Error() error {
	// if r.Err != nil {
	// 	return r.Err
	// }
	// for _, rr := range r.Results {
	// 	// if rr.Err != nil {
	// 	// 	return rr.Err
	// 	// }
	// }
	return nil
}
