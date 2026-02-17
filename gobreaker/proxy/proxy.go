/*
Package gobreaker provides a circuit breaker proxy middleware using the sony/gobreaker lib.

Adding the middleware to your proxy stack

	import (
		"github.com/devopsfaith/lura/v2/proxy"
		gobreaker "github.com/krakend/krakend-circuitbreaker/gobreaker/proxy"
	)

	...

	var p proxy.Proxy
	var backend *config.Backend

	...

	p = gobreaker.NewMiddleware(backend)(p)

	...
*/
package proxy

import (
	"context"
	"fmt"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"

	gcb "github.com/krakend/krakend-circuitbreaker/v3/gobreaker"
)

// BackendFactory adds a cb middleware wrapping the internal factory
func BackendFactory(next proxy.BackendFactory, logger logging.Logger) proxy.BackendFactory {
	return func(cfg *config.Backend) proxy.Proxy {
		return NewMiddleware(cfg, logger)(next(cfg))
	}
}

// NewMiddleware builds a middleware based on the extra config params or fallbacks to the next proxy
func NewMiddleware(remote *config.Backend, logger logging.Logger) proxy.Middleware {
	data := gcb.ConfigGetter(remote.ExtraConfig).(gcb.Config)
	if data == gcb.ZeroCfg {
		return proxy.EmptyMiddleware
	}
	cb := gcb.NewCircuitBreaker(data, logger)

	logger.Debug(fmt.Sprintf("[BACKEND: %s][CB] Creating the circuit breaker named '%s'", remote.URLPattern, data.Name))

	return func(next ...proxy.Proxy) proxy.Proxy {
		if len(next) > 1 {
			panic(proxy.ErrTooManyProxies)
		}
		return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
			result, err := cb.Execute(func() (interface{}, error) { return next[0](ctx, request) })
			if err != nil {
				return nil, err
			}
			return result.(*proxy.Response), err
		}
	}
}
