/*
Package gobreaker provides a circuit breaker adapter using the sony/gobreaker lib.

Sample backend extra config

	...
	"extra_config": {
		...
		"qos/circuit-breaker": {
			"interval":        60,
			"timeout":         10,
			"max_errors":       5,
			"log_status_change": true,
		},
		...
	},
	...

The gobreaker package provides an efficient circuit breaker implementation. See https://github.com/sony/gobreaker
and https://martinfowler.com/bliki/CircuitBreaker.html for more details.
*/
package gobreaker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/luraproject/lura/v3/config"
	"github.com/luraproject/lura/v3/logging"
	"github.com/sony/gobreaker/v2"
)

// Namespace is the key to use to store and access the custom config data
const Namespace = "qos/circuit-breaker"

// Config is the custom config struct containing the params for the sony/gobreaker package
type Config struct {
	Name            string
	Interval        int
	Timeout         int
	MaxErrors       int
	LogStatusChange bool

	StateEvent  func(context.Context, Event)
	ErrorsEvent func(context.Context, Event)
}

type Event struct {
	Name        string
	ErrorCount  gobreaker.Counts
	StateChange int
}

var (
	defaultStateEvent  = func(_ context.Context, _ Event) {}
	defaultErrorsEvent = func(_ context.Context, _ Event) {}
)

// ConfigGetter implements the config.ConfigGetter interface. It parses the extra config for the
// gobreaker adapter and returns a ZeroCfg if something goes wrong.
func ConfigGetter(e config.ExtraConfig) interface{} {
	v, ok := e[Namespace]
	if !ok {
		return nil
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	cfg := Config{}
	if v, ok := tmp["name"]; ok {
		if name, ok := v.(string); ok {
			cfg.Name = name
		}
	}
	if v, ok := tmp["interval"]; ok {
		switch i := v.(type) {
		case int:
			cfg.Interval = i
		case float64:
			cfg.Interval = int(i)
		}
	}
	if v, ok := tmp["timeout"]; ok {
		switch i := v.(type) {
		case int:
			cfg.Timeout = i
		case float64:
			cfg.Timeout = int(i)
		}
	}
	if v, ok := tmp["max_errors"]; ok {
		switch i := v.(type) {
		case int:
			cfg.MaxErrors = i
		case float64:
			cfg.MaxErrors = int(i)
		}
	}
	value, ok := tmp["log_status_change"].(bool)
	cfg.LogStatusChange = ok && value

	cfg.ErrorsEvent = defaultErrorsEvent
	cfg.StateEvent = defaultStateEvent
	return cfg
}

type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker[interface{}]
}

func (c *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return c.cb.Execute(req)
}

// NewCircuitBreaker builds a gobreaker circuit breaker with the injected config
func NewCircuitBreaker(cfg Config, logger logging.Logger) CircuitBreaker {
	settings := gobreaker.Settings{
		Name:     cfg.Name,
		Interval: time.Duration(cfg.Interval) * time.Second,
		Timeout:  time.Duration(cfg.Timeout) * time.Second,
		IsExcluded: func(err error) bool {
			return errors.Is(err, context.Canceled)
		},
	}

	settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
		cfg.ErrorsEvent(context.Background(),
			Event{
				Name:       cfg.Name,
				ErrorCount: counts,
			})
		return counts.ConsecutiveFailures > uint32(cfg.MaxErrors)
	}

	settings.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		if t, err := transition(from, to); err == nil {
			cfg.StateEvent(context.Background(),
				Event{
					Name:        name,
					StateChange: t,
				})
		}
		if cfg.LogStatusChange {
			logger.Warning(fmt.Sprintf("[CB] Circuit breaker named '%s' went from '%s' to '%s'", name, from.String(), to.String()))
		}
	}

	return CircuitBreaker{cb: gobreaker.NewCircuitBreaker[interface{}](settings)}
}

const (
	CLOSETOOPEN int = iota
	OPENTOHALF
	HALFTOOPEN
	HALFTOCLOSE
)

func transition(from, to gobreaker.State) (int, error) {
	switch from {
	case 0:
		return CLOSETOOPEN, nil
	case 1:
		if to == 0 {
			return HALFTOCLOSE, nil
		}
		return HALFTOOPEN, nil
	case 2:
		return OPENTOHALF, nil
	default:
	}
	return 0, fmt.Errorf("invalid cb state %d", from)
}
