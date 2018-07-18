package adaptd

import (
	"log"
	"net/http"
)

// Notify adapter logs when the request is beginning to be processed and when it is finished.
func Notify(logger *log.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("Handling %v request at URL %v\n", r.Method, r.URL)
			defer logger.Printf("%v request at URL %v was handled\n", r.Method, r.URL)
			h.ServeHTTP(w, r)
		})
	}
}

// GetAndOtherRequest adapter uses two handlers to handle both get requests and another.
// All other requests are given a http.StatusMethodNotAllowed error.
// The other handler is provided to create the Adapter while the get handler should be provided to the Adapter
// e.g. `GetAndOtherRequest(other, http.MethodPost)(getHandler)`
func GetAndOtherRequest(other http.Handler, method string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case method:
				// Handler the 'other' request method
				other.ServeHTTP(w, r)
			case http.MethodGet:
				// Handle the GET request
				h.ServeHTTP(w, r)
			default:
				// We are not allowing this method so respond with an error
				http.Error(w, "Request method not allowed", http.StatusMethodNotAllowed)
			}
		})
	}
}

// AddHeader adapter adds the header before calling the handler
func AddHeader(name, value string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(name, value)
			h.ServeHTTP(w, r)
		})
	}
}

// AddHeaderWithFunc adds the header before calling the handler.
// This is useful for things like CSRF tokens.
func AddHeaderWithFunc(name string, tg func() string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(name, tg())
			h.ServeHTTP(w, r)
		})
	}
}

// DisallowLongerPaths adapter returns http.NotFound Error if the URL path is longer than the registered one
func DisallowLongerPaths(path string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path {
				log.Printf("Handler expects URL %v but received a request at %v\n", path, r.URL.Path)
				http.NotFound(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
