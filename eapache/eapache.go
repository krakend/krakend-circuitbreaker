/*
Package eapache provides a circuit breaker adapter using the github.com/eapache/go-resiliency/breaker lib.

Sample backend extra config

	...
	"extra_config": {
		...
		"github.com/devopsfaith/krakend-circuitbreaker/eapache": {
			"success": 60,
			"timeout": "10s",
			"error": 5
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
	"github.com/eapache/go-resiliency/breaker"
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
	v, ok := e[Namespace]
	if !ok {
		return ZeroCfg
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return ZeroCfg
	}
	cfg := Config{}
	if v, ok := tmp["error"]; ok {
		cfg.Error = int(v.(float64))
	}
	if v, ok := tmp["success"]; ok {
		cfg.Success = int(v.(float64))
	}
	if v, ok := tmp["timeout"]; ok {
		if d, err := time.ParseDuration(v.(string)); err == nil {
			cfg.Timeout = d
		}
	}
	return cfg
}

// NewCircuitBreaker builds a eapache circuit breaker with the injected config
func NewCircuitBreaker(cfg Config) *breaker.Breaker {
	return breaker.New(cfg.Error, cfg.Success, cfg.Timeout)
}
