package tarpit

import (
	"net/http"
	"time"
)

// ListenAndServe - same as http.ListenAndServe(addr string, handler http.Handler) error
// unless it adds writeTimeout parameter and ensure tcp keep alive to prevent the client from timing out.
// Typically writeTimeout takes high value (i.e time.Hour) to ensure the tarpit is effective.
func ListenAndServe(addr string, handler http.Handler, writeTimeout time.Duration) error {
	server := http.Server{
		Addr:         addr,
		Handler:      handler,
		WriteTimeout: writeTimeout,
	}
	server.SetKeepAlivesEnabled(true)
	return server.ListenAndServe()
}
