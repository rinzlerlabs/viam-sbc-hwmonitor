//go:build debug
// +build debug

package main

import (
	"net/http"
	_ "net/http/pprof" // Import pprof for profiling
)

func init() {
	go func() {
		http.ListenAndServe(":6060", nil) // Start the pprof server
	}()
}
