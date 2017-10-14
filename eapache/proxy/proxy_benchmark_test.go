package proxy

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"

	"github.com/devopsfaith/krakend-circuitbreaker/eapache"
)

func BenchmarkNewCircuitBreakerMiddleware_ok(b *testing.B) {
	p := NewMiddleware(&cfg)(dummyProxy(&proxy.Response{}, nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(context.Background(), &proxy.Request{
			Path: "/tupu",
		})
	}
}

func BenchmarkNewCircuitBreakerMiddleware_ko(b *testing.B) {
	p := NewMiddleware(&cfg)(dummyProxy(nil, errors.New("sample error")))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(context.Background(), &proxy.Request{
			Path: "/tupu",
		})
	}
}

func BenchmarkNewCircuitBreakerMiddleware_burst(b *testing.B) {
	err := errors.New("sample error")
	p := NewMiddleware(&cfg)(burstProxy(&proxy.Response{}, err, 100, 6))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(context.Background(), &proxy.Request{
			Path: "/tupu",
		})
	}
}

var cfg = config.Backend{
	ExtraConfig: map[string]interface{}{
		eapache.Namespace: map[string]interface{}{
			"error":   10.0,
			"success": 10.0,
			"timeout": "1s",
		},
	},
}

func burstProxy(r *proxy.Response, err error, ok, ko int) proxy.Proxy {
	tmp := make([]bool, ok+ko)
	for i := 0; i < ok+ko; i++ {
		tmp[i] = i < ok
	}
	calls := uint64(0)
	return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		total := atomic.AddUint64(&calls, 1) - 1
		if tmp[total%uint64(len(tmp))] {
			return r, nil
		}
		return nil, err
	}
}
