package requests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Request interface {
	Get(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Post(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Put(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Patch(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Delete(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Head(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Connect(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Options(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Trace(ctx context.Context, url string, opts ...DialOption) (*response, error)
	Do(ctx context.Context, client *http.Client, request *http.Request) (*response, error)
}

var _ Request = New()

var defaultReq = New(WithClient(http.DefaultClient))

func Get(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodGet, url, opts...)
}

func Post(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodPost, url, opts...)
}

func Put(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodPut, url, opts...)
}

func Patch(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodPatch, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodDelete, url, opts...)
}

func Head(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodHead, url, opts...)
}

func Connect(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodConnect, url, opts...)
}

func Options(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodOptions, url, opts...)
}

func Trace(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return defaultReq.do(ctx, http.MethodTrace, url, opts...)
}

type requests struct {
	opts dialOptions
}

/*
init request client
eg: New(WithClient(http.DefaultClient))
*/
func New(opts ...DialOption) *requests {
	req := &requests{}
	for _, opt := range opts {
		opt(&req.opts)
	}
	return req
}

func (req *requests) Get(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodGet, url, opts...)
}

func (req *requests) Post(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodPost, url, opts...)
}

func (req *requests) Put(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodPut, url, opts...)
}

func (req *requests) Patch(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodPatch, url, opts...)
}

func (req *requests) Delete(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodDelete, url, opts...)
}

func (req *requests) Head(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodHead, url, opts...)
}

func (req *requests) Connect(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodConnect, url, opts...)
}

func (req *requests) Options(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodOptions, url, opts...)
}

func (req *requests) Trace(ctx context.Context, url string, opts ...DialOption) (*response, error) {
	return req.do(ctx, http.MethodTrace, url, opts...)
}

// 用于单独执行一个请求的goroutine，便于自定义请求过程
func (req *requests) Do(ctx context.Context, client *http.Client, request *http.Request) (*response, error) {
	return exec(ctx, client, request)
}

func (req requests) do(ctx context.Context, method string, url string, opts ...DialOption) (*response, error) {
	for _, opt := range opts {
		opt(&req.opts)
	}
	fmt.Printf("opts: %+v, addr=%p\n", req.opts, &req.opts)

	// 参数设置异常
	if req.opts.err != nil {
		return nil, req.opts.err
	}

	if req.opts.query != "" {
		url += "?" + req.opts.query
	}
	request, err := http.NewRequest(method, url, req.opts.body)
	if err != nil {
		return nil, err
	}

	// 设置请求headers
	for k, v := range req.opts.headers {
		request.Header.Set(k, v)
	}
	return exec(ctx, req.opts.client, request)
}

func exec(ctx context.Context, client *http.Client, req *http.Request) (*response, error) {
	c := make(chan error)
	req = req.WithContext(ctx)
	var resp *response
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
				switch x := r.(type) {
				case string:
					c <- errors.New(x)
				case error:
					c <- x
				default:
					c <- errors.New("unknown panic")
				}
			}
		}()
		r, err := client.Do(req)
		resp = &response{
			resp: r,
		}
		c <- err
	}()

	select {
	case <-ctx.Done():
		<-c
		return nil, ctx.Err()
	case err := <-c:
		return resp, err
	}
}
