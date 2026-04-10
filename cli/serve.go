package main

import (
	"net/http"
)

// serveHTTP starts a static file server for the UCLAW web UI.
func serveHTTP(addr, dir string) error {
	return http.ListenAndServe(addr, http.FileServer(http.Dir(dir)))
}
