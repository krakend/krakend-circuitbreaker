/*
Package eapache provides a circuit breaker proxy middleware using the github.com/eapache/go-resiliency/breaker lib.

Adding the middleware to your proxy stack

	import eapache "github.com/devopsfaith/krakend-circuitbreaker/eapache/proxy"

	...

	var p proxy.Proxy
	var backend *config.Backend

	...

	p = eapache.NewMiddleware(backend)(p)

	...

*/
package proxy

import (
	"context"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"

	"github.com/devopsfaith/krakend-circuitbreaker/eapache"
)

// BackendFactory adds a cb middleware wrapping the internal factory
func BackendFactory(next proxy.BackendFactory) proxy.BackendFactory {
	return func(cfg *config.Backend) proxy.Proxy {
		return NewMiddleware(cfg)(next(cfg))
	}
}

// NewMiddleware builds a middleware based on the extra config params or fallbacks to the next proxy
func NewMiddleware(remote *config.Backend) proxy.Middleware {
	data := eapache.ConfigGetter(remote.ExtraConfig).(eapache.Config)
	if data == eapache.ZeroCfg {
		return proxy.EmptyMiddleware
	}
	cb := eapache.NewCircuitBreaker(data)

	return func(next ...proxy.Proxy) proxy.Proxy {
		if len(next) > 1 {
			panic(proxy.ErrTooManyProxies)
		}
		return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
			var res *proxy.Response
			if err := cb.Run(func() error {
				var err1 error
				res, err1 = next[0](ctx, request)
				return err1
			}); err != nil {
				return nil, err
			}
			return res, nil
		}
	}
}
