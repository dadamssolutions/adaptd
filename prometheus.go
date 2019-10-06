package adaptd

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// CountHTTPResponses calls the handler and records the response as a prometheus counter
// with labels endpoint, code, and method.
// This should be applied once for an entire web server.
func CountHTTPResponses() Adapter {
	httpRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests processed, partitioned by endpoint, status code, and HTTP method.",
		},
		[]string{"endpoint", "code", "method"},
	)
	prometheus.MustRegister(httpRequests)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sr := &statusRecorder{w, 200}
			h.ServeHTTP(sr, r)
			httpRequests.WithLabelValues(r.URL.Path, strconv.Itoa(sr.status), r.Method).Inc()
		})
	}
}

// TrackHTTPResponseTimes calls the handler and records the response time
// as a prometheus summary with labels endpoint, code, and method.
// This should be applied once for an entire web server.
func TrackHTTPResponseTimes() Adapter {
	httpRequests := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_requests_secs",
			Help: "The response times to HTTP requests, partitioned by endpoint, status code, and HTTP method.",
		},
		[]string{"endpoint", "code", "method"},
	)
	prometheus.MustRegister(httpRequests)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sr := &statusRecorder{w, 200}
			start := time.Now().Unix()
			h.ServeHTTP(sr, r)
			httpRequests.WithLabelValues(r.URL.Path, strconv.Itoa(sr.status), r.Method).Observe(
				float64(time.Now().Unix() - start),
			)
		})
	}
}
