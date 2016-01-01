package httpd

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/bmizerany/pat"
	"github.com/gocql/gocql"
	"github.com/mattbaird/elastigo/lib"
	"github.com/rcrowley/go-metrics"
	"github.com/vatcinc/bio/models"
	"github.com/vatcinc/bio/schema"
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
	}

	h.SetRoutes([]route{
		route{
			"token_req",
			"POST", "/api/v1.1/token", true, true, h.authUser,
		},
		//DOCS
		route{
			"import_document",
			"PUT", "/api/v1.1/doc_import", true, true, h.wixImport,
		},
		route{
			"get_document",
			"GET", "/api/v1.1/document", true, true, h.getDocument,
		},
		route{
			"get_documents",
			"GET", "/api/v1.1/documents", true, true, h.getDocuments,
		},
		route{
			"upsert_document",
			"PUT", "/api/v1.1/documents", true, true, h.upsertDocument,
		},
		// Org
		route{
			"upsert_org_post",
			"POST", "/api/v1.1/organizations", true, true, h.insertCustomer,
		},
		route{
			"upsert_org_put",
			"PATCH", "/api/v1.1/organizations/:id", true, true, h.updateCustomer,
		},
		route{
			"get_customer",
			"GET", "/api/v1.1/organizations/:id", true, true, h.getCustomer,
		},
		route{
			"search_for",
			"GET", "/api/v1.1/organizations", true, true, h.searchCustomers,
		},
		route{
			"search_customer_domains",
			"GET", "/api/v1.1/domains", true, true, h.getCustomerDomains,
		},
		// Domain
		route{
			"insert_domain",
			"POST", "/api/v1.1/realdomains", true, true, h.insertDomain,
		},
		route{
			"update_domain",
			"PATCH", "/api/v1.1/realdomains/:id", true, true, h.updateDomain,
		},
		route{
			"get_domain",
			"GET", "/api/v1.1/realdomains/:id", true, true, h.getDomain,
		},
		route{
			"search_domains",
			"GET", "/api/v1.1/realdomains", true, true, h.searchDomains,
		},
		// Widgets
		route{
			"fetch_widgets",
			"GET", "/api/v1.1/widgets", true, true, h.serveWidgetsList,
		},
		route{
			"add_widgets_route",
			"PUT", "/api/v1.1/widgets", true, true, h.upsertWidget,
		},
		route{
			"remove_widget",
			"DELETE", "/api/v1.1/widgets", true, true, h.deleteWidget,
		},
		route{
			"query", // Query serving route.
			"GET", "/api/v1.1", true, true, h.serveRoot,
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
		if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *bio.Domains)); ok {
			handler = materializeDomain(hf, h)
		}

		// If it's a handler func that requires authorization, wrap it in authorization
		if hf, ok := r.handlerFunc.(func(http.ResponseWriter, *http.Request, *bio.Users)); ok {
			handler = authenticate(hf, h, h.requireAuthentication)
		}
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

func (h *Handler) serveRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	res := &RootAPIResult{
		AppName: "bio Web Services",
		Version: h.Version,
		AllowedPaths: []map[string]interface{}{
			{"path": "/", "info": "Global API Information", "methods": "GET,OPTIONS"},
			{"path": "/health", "info": "API Health Information", "methods": "GET,OPTIONS"},
		},
	}
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

// Helpers

// MarshalEmber wraps a document in an ember package
func MarshalEmber(w http.ResponseWriter, id gocql.UUID, data interface{}, typeString string, pretty bool) {
	w.Header().Add("content-type", "application/json")
	ember := &models.EmberData{
		Data: models.EmberDataObj{
			Type:       typeString,
			ID:         id.String(),
			Attributes: MarshalJSON(data, pretty),
		},
	}
	w.Write(MarshalJSON(ember, pretty))
}

// MarshalMultiEmber wraps a document in an ember package
func MarshalMultiEmber(w http.ResponseWriter, data *models.EmberMultiData, pretty bool) {
	w.Header().Add("content-type", "application/json")
	w.Write(MarshalJSON(data, pretty))
}

// EmberDataObj defines a inner-response object that follows embers standards
type EmberDataObj struct {
	Type       string          `json:"type"`
	ID         string          `json:"id,omitempty"`
	Attributes json.RawMessage `json:"attributes"`
}

// EmberData is the wrapper for EmberDataObj.
type EmberData struct {
	Data   EmberDataObj `json:"data"`
	Errors interface{}  `json:"errors,omitempty"`
	Meta   interface{}  `json:"meta,omitempty"`
}

// EmberMultiData defines multiple occurrencies of EmberData's Data
type EmberMultiData struct {
	Data   interface{} `json:"data"`
	Errors interface{} `json:"errors,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
}
