package proxy

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/eapache/go-resiliency/breaker"

	"github.com/devopsfaith/krakend-circuitbreaker/eapache"
)

func TestNewMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r != proxy.ErrTooManyProxies {
			t.Errorf("The code did not panic\n")
		}
	}()
	NewMiddleware(&config.Backend{})(proxy.NoopProxy, proxy.NoopProxy)
}

func TestNewMiddleware_zeroConfig(t *testing.T) {
	for _, cfg := range []*config.Backend{
		&config.Backend{},
		&config.Backend{ExtraConfig: map[string]interface{}{eapache.Namespace: 42}},
	} {
		resp := proxy.Response{}
		mdw := NewMiddleware(cfg)
		p := mdw(dummyProxy(&resp, nil))

		request := proxy.Request{
			Path: "/tupu",
		}

		for i := 0; i < 100; i++ {
			r, err := p(context.Background(), &request)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if &resp != r {
				t.Fail()
			}
		}
	}
}

func TestNewMiddleware_ok(t *testing.T) {
	resp := proxy.Response{}
	mdw := NewMiddleware(&config.Backend{
		ExtraConfig: map[string]interface{}{
			eapache.Namespace: map[string]interface{}{
				"error":   1.0,
				"success": 1.0,
				"timeout": "1s",
			},
		},
	})
	p := mdw(dummyProxy(&resp, nil))

	request := proxy.Request{
		Path: "/tupu",
	}

	for i := 0; i < 100; i++ {
		r, err := p(context.Background(), &request)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if &resp != r {
			t.Fail()
		}
	}
}

func TestNewMiddleware_ko(t *testing.T) {
	expectedErr := fmt.Errorf("Some error")
	calls := uint64(0)
	mdw := NewMiddleware(&config.Backend{
		ExtraConfig: map[string]interface{}{
			eapache.Namespace: map[string]interface{}{
				"error":   1.0,
				"success": 0.0,
				"timeout": "1s",
			},
		},
	})
	p := mdw(func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		total := atomic.AddUint64(&calls, 1)
		if total > 2 {
			t.Error("This proxy shouldn't been executed!")
		}
		return nil, expectedErr
	})

	request := proxy.Request{
		Path: "/tupu",
	}

	r, err := p(context.Background(), &request)
	if err != expectedErr {
		t.Error("error expected")
	}
	if nil != r {
		t.Error("not nil response")
	}

	for i := 0; i < 100; i++ {
		r, err := p(context.Background(), &request)
		if err != breaker.ErrBreakerOpen {
			t.Error("error expected")
		}
		if nil != r {
			t.Error("not nil response")
		}
	}
}

func dummyProxy(r *proxy.Response, err error) proxy.Proxy {
	return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return r, err
	}
}
