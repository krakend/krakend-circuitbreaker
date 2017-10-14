/*
Package gobreaker provides a circuit breaker proxy middleware using the sony/gobreaker lib.

Adding the middleware to your proxy stack

	import (
		"github.com/devopsfaith/krakend/proxy"
		gobreaker "github.com/devopsfaith/krakend-circuitbreaker/gobreaker/proxy"
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

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"

	gcb "github.com/devopsfaith/krakend-circuitbreaker/gobreaker"
)

// BackendFactory adds a cb middleware wrapping the internal factory
func BackendFactory(next proxy.BackendFactory) proxy.BackendFactory {
	return func(cfg *config.Backend) proxy.Proxy {
		return NewMiddleware(cfg)(next(cfg))
	}
}

// NewMiddleware builds a middleware based on the extra config params or fallbacks to the next proxy
func NewMiddleware(remote *config.Backend) proxy.Middleware {
	data := gcb.ConfigGetter(remote.ExtraConfig).(gcb.Config)
	if data == gcb.ZeroCfg {
		return proxy.EmptyMiddleware
	}
	cb := gcb.NewCircuitBreaker(data)

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
