/*
Package gobreaker provides a circuit breaker adapter using the sony/gobreaker lib.

Sample backend extra config

	...
	"extra_config": {
		...
		"github.com/devopsfaith/krakend-circuitbreaker/gobreaker": {
			"interval": 60,
			"timeout": 10,
			"maxErrors": 5
		},
		...
	},
	...

The gobreaker package provides an efficient circuit breaker implementation. See https://github.com/sony/gobreaker
and https://martinfowler.com/bliki/CircuitBreaker.html for more details.
*/
package gobreaker

import (
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/sony/gobreaker"
)

// Namespace is the key to use to store and access the custom config data
const Namespace = "github.com/devopsfaith/krakend-circuitbreaker/gobreaker"

// Config is the custom config struct containing the params for the sony/gobreaker package
type Config struct {
	Interval  int
	Timeout   int
	MaxErrors int
}

// ZeroCfg is the zero value for the Config struct
var ZeroCfg = Config{}

// ConfigGetter implements the config.ConfigGetter interface. It parses the extra config for the
// gobreaker adapter and returns a ZeroCfg if something goes wrong.
func ConfigGetter(e config.ExtraConfig) interface{} {
	v, ok := e[Namespace]
	if !ok {
		return ZeroCfg
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return ZeroCfg
	}
	cfg := Config{}
	if v, ok := tmp["interval"]; ok {
		cfg.Interval = int(v.(float64))
	}
	if v, ok := tmp["timeout"]; ok {
		cfg.Timeout = int(v.(float64))
	}
	if v, ok := tmp["maxErrors"]; ok {
		cfg.MaxErrors = int(v.(float64))
	}
	return cfg
}

// NewCircuitBreaker builds a gobreaker circuit breaker with the injected config
func NewCircuitBreaker(cfg Config) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Interval: time.Duration(cfg.Interval) * time.Second,
		Timeout:  time.Duration(cfg.Timeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > uint32(cfg.MaxErrors)
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}
