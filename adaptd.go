// Package adaptd provides a simple adapter interface for adding middleware to http frameworks
package adaptd

import (
	"log"
	"net/http"
)

// Adapter is a type that helps with http middleware.
type Adapter func(http.Handler) http.Handler

// Adapt is a helper to add all the adapters required for a given Handler
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	// Attach adapters in reverse order because that is what should be implied by the ordering of the caller.
	// They way the middleware will work is the first adapter applied will be the last one to get called.
	for i := len(adapters) - 1; i >= 0; i-- {
		h = adapters[i](h)
	}
	return h
}

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

// RequestMethod adapter allow allows the given request method.
// All other requests are given a http.StatusMethodNotAllowed error.
func RequestMethod(method string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == method {
				h.ServeHTTP(w, r)
			} else {
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

// AddCookieWithFunc adds the header before calling the handler.
// This is useful for things like CSRF tokens.
func AddCookieWithFunc(name string, tg func(http.ResponseWriter) error) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tg(w)
			h.ServeHTTP(w, r)
		})
	}
}

// DisallowLongerPaths adapter returns http.NotFound Error if the URL path is longer than the registered one
func DisallowLongerPaths(path string, notFoundHandler http.Handler) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path {
				log.Printf("Handler expects URL %v but received a request at %v\n", path, r.URL.Path)
				notFoundHandler.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// HTTPSRedirect adapter redirects all HTTP requests to HTTPS requests.
// Most users should simply call this as `http.ListenAndServer(":80", HTTPSRedirect())`
func HTTPSRedirect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := "https://" + r.Host + r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}
		log.Printf("redirect to: %s", target)
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	})
}

// EnsureHTTPS adapter redirects an HTTP request to an HTTPS request.
// Some hosts forward requests and use 'X-Forward-Proto == "https"'
// to indicate that he request was made with https protocol.
// If you would like to allow this as a valid check, then the parameter should be true.
func EnsureHTTPS(allowXForwardedProto bool) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isHTTPS(r, allowXForwardedProto) {
				target := "https://" + r.Host + r.URL.Path
				if len(r.URL.RawQuery) > 0 {
					target += "?" + r.URL.RawQuery
				}
				log.Printf("redirect to: %s", target)
				http.Redirect(w, r, target, http.StatusTemporaryRedirect)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// OnCheck adapter checks the return of the function. On false, it calls the handler.
// On true, it will call the handler passed to the Adapter.
func OnCheck(f func(http.ResponseWriter, *http.Request) bool, falseHandler http.Handler, logOnFalse string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !f(w, r) {
				log.Println(logOnFalse)
				falseHandler.ServeHTTP(w, r)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

// CheckAndRedirect adapter checks the return of the function. On false, it redirects to the given URL.
// On true, it will call the handler passed to the Adapater.
func CheckAndRedirect(f func(http.ResponseWriter, *http.Request) bool, redirect http.Handler, logOnRedirect string) Adapter {
	return OnCheck(f, redirect, logOnRedirect+" redirecting")
}

func isHTTPS(r *http.Request, allowXForwardedProto bool) bool {
	return (r.TLS != nil && r.TLS.HandshakeComplete) || (allowXForwardedProto && r.Header.Get("X-Forwarded-Proto") == "https")
}
