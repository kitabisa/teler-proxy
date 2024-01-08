package common

import "github.com/charmbracelet/log"

type Options struct {
	*Port
	// NOTE(dwisiswant0): I think it would be fine if we just added
	// our own metrics route (w/ --metrics flag) for the Prometheus
	// handler instead of running a new server (--metrics-port)
	Destination string
	*Config
	*TLS
	*log.Logger
}
