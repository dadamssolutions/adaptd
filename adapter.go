// Package adaptd provides a simple adapter interface for adding middleware to http frameworks
package adaptd

import "net/http"

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
