package httpd

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
	"github.com/rcrowley/go-metrics"
	//"github.com/fkasper/sitrep-authentication/models"
)

// determines if the client can accept compressed responses, and encodes accordingly
func gzipFilter(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			inner.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		inner.ServeHTTP(gzw, r)
	})
}

// versionHeader takes a HTTP handler and returns a HTTP handler
// and adds the X-authentication-VERSION header to outgoing responses.
func versionHeader(inner http.Handler, h *Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-authentication-Version", h.Version)
		inner.ServeHTTP(w, r)
	})
}

// cors responds to incoming requests and adds the appropriate cors headers
// TODO: corylanou: add the ability to configure this in our config
func cors(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set(`Access-Control-Allow-Origin`, origin)
			w.Header().Set(`Access-Control-Allow-Methods`, strings.Join([]string{
				`DELETE`,
				`GET`,
				`OPTIONS`,
				`POST`,
				`PUT`,
			}, ", "))

			w.Header().Set(`Access-Control-Allow-Headers`, strings.Join([]string{
				`Accept`,
				`Accept-Encoding`,
				`Authorization`,
				`Content-Length`,
				`Content-Type`,
				`X-EXERCISE-ID`,
				`X-CSRF-Token`,
				`X-HTTP-Method-Override`,
			}, ", "))

			w.Header().Set(`Access-Control-Expose-Headers`, strings.Join([]string{
				`Date`,
				`X-authentication-Version`,
			}, ", "))
		}

		if r.Method == "OPTIONS" {
			return
		}

		inner.ServeHTTP(w, r)
	})
}

func requestID(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//uid := uuid.TimeUUID()
		uuid, err := gocql.RandomUUID()
		if err == nil {
			r.Header.Set("Request-Id", uuid.String())
		}
		w.Header().Set("Request-Id", r.Header.Get("Request-Id"))

		inner.ServeHTTP(w, r)
	})
}

func logging(inner http.Handler, name string, weblog *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		l := &responseLogger{w: w}
		inner.ServeHTTP(l, r)
		logLine := buildLogLine(l, r, start)
		weblog.Println(logLine)
	})
}

func recovery(inner http.Handler, name string, weblog *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		l := &responseLogger{w: w}

		defer func() {
			if err := recover(); err != nil {
				logLine := buildLogLine(l, r, start)
				logLine = fmt.Sprintf(`%s [panic:%s]`, logLine, err)
				weblog.Println(logLine)
			}
		}()

		inner.ServeHTTP(l, r)
	})
}

func makeForbidden(w http.ResponseWriter, err error) {
	httpError(w, fmt.Sprintf("You are not allowed to access this resource: %s", err.Error()), false, http.StatusForbidden)
}

func authenticate(inner func(http.ResponseWriter, *http.Request, *sitrep.UsersByEmail), h *Handler, requireAuthentication bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireAuthentication {
			inner(w, r, nil)
			return
		}
		counter := metrics.GetOrRegisterCounter(statAuthFail, h.statMap)
		accessToken, err := parseCredentials(r)
		if err != nil {
			counter.Inc(1)
			makeForbidden(w, err)
			return
		}

		user, err := models.VerifyUserRequest(h.Cassandra, accessToken)
		if err != nil {
			counter.Inc(1)
			makeForbidden(w, err)
			return
		}
		inner(w, r, user)
	})
}

func exercisifyOnly(inner func(http.ResponseWriter, *http.Request, *sitrep.ExerciseByIdentifier), h *Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exerciseIDRaw, err := parseExerciseID(r)
		if err != nil {
			makeForbidden(w, err)
			return
		}

		exerciseID, err := gocql.ParseUUID(exerciseIDRaw)
		if err != nil {
			makeForbidden(w, err)
			return
		}

		exercise, err := models.FindExerciseByID(h.Cassandra, exerciseID)
		if err != nil {
			makeForbidden(w, err)
			return
		}
		inner(w, r, exercise)
	})
}

func exercisify(inner func(http.ResponseWriter, *http.Request, *sitrep.UsersByEmail, *sitrep.ExerciseByIdentifier), h *Handler, requireAuthentication bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exerciseIDRaw, err := parseExerciseID(r)
		if err != nil {
			makeForbidden(w, err)
			return
		}

		exerciseID, err := gocql.ParseUUID(exerciseIDRaw)
		if err != nil {
			makeForbidden(w, err)
			return
		}

		exercise, err := models.FindExerciseByID(h.Cassandra, exerciseID)
		if err != nil {
			makeForbidden(w, err)
			return
		}
		if !requireAuthentication {
			inner(w, r, nil, exercise)
			return
		}
		counter := metrics.GetOrRegisterCounter(statAuthFail, h.statMap)
		accessToken, err := parseCredentials(r)
		if err != nil {
			counter.Inc(1)
			makeForbidden(w, err)
			return
		}

		user, err := models.VerifyUserRequest(h.Cassandra, accessToken)
		if err != nil {
			counter.Inc(1)
			makeForbidden(w, err)
			return
		}
		inner(w, r, user, exercise)
	})
}

func parseDomain(r *http.Request) (string, error) {
	q := r.URL.Query()

	if u := q.Get("exercise_ident"); u != "" {
		return u, nil
	}
	if len(r.Header["X-EXERCISE-ID"]) > 0 {
		return r.Header["Authorization"][0], nil
	}
	cookie, err := r.Cookie("ex_id")
	if err == nil {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("unable to parse Domain")
}

// parseCredentials returns the acccess token encoded in
// a request. The credentials may be present as URL query params, or as
// a Authorization header.
// as params: http://127.0.0.1/query?access_token=<token>
// as basic auth: http://127.0.0.1/query (Header: Authorization: Bearer <token>)
func parseCredentials(r *http.Request) (string, error) {
	q := r.URL.Query()

	if u := q.Get("access_token"); u != "" {
		return u, nil
	}
	if len(r.Header["Authorization"]) > 0 {
		u := strings.SplitN(r.Header["Authorization"][0], " ", 2)

		if len(u) == 2 && u[0] == "Bearer" {
			return u[1], nil
		}
	}
	cookie, err := r.Cookie("sid")
	if err == nil {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("unable to parse Bearer Auth credentials")
}

// parseExerciseID returns the exercise currently supplied, can be:
// COOKIE: ex_id (string)
// GET: ex_id (string)
// HEADER: X-EXERCISE-ID (string)
func parseExerciseID(r *http.Request) (string, error) {
	q := r.URL.Query()

	if u := q.Get("ex_id"); u != "" {
		return u, nil
	}
	if len(r.Header["X-Exercise-Id"]) > 0 {
		return r.Header["X-Exercise-Id"][0], nil
	}
	cookie, err := r.Cookie("ex_id")
	if err == nil {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("unable to parse Exercise ID")
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w gzipResponseWriter) Flush() {
	w.Writer.(*gzip.Writer).Flush()
}
