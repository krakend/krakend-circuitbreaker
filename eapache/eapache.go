/*
Package eapache provides a circuit breaker adapter using the github.com/eapache/go-resiliency/breaker lib.

Sample backend extra config

	...
	"extra_config": {
		...
		"github.com/devopsfaith/krakend-circuitbreaker/eapache": {
			"Interval": 60,
			"Timeout": 10,
			"MaxErrors": 5
		},
		...
	},
	...

The eapache package provides an efficient circuit breaker implementation. See https://github.com/eapache/go-resiliency/breaker
and https://martinfowler.com/bliki/CircuitBreaker.html for more details.
*/
package eapache

import (
	"time"

	"github.com/devopsfaith/krakend/config"
)

// Namespace is the key to use to store and access the custom config data
const Namespace = "github.com/devopsfaith/krakend-circuitbreaker/eapache"

// Config is the custom config struct containing the params for the eapache/go-resiliency/breaker package
type Config struct {
	Error   int
	Success int
	Timeout time.Duration
}

// ZeroCfg is the zero value for the Config struct
var ZeroCfg = Config{}

// ConfigGetter implements the config.ConfigGetter interface. It parses the extra config for the
// eapache adapter and returns a ZeroCfg if something goes wrong.
func ConfigGetter(e config.ExtraConfig) interface{} {
	if v, ok := e[Namespace]; ok {
		if cfg, ok := v.(Config); ok {
			return cfg
		}
	}
	return ZeroCfg
}
