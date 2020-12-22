package requests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type Http interface {
	Get(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Post(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Put(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Patch(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Delete(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Head(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Connect(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Options(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Trace(ctx context.Context, url string, opts ...DialOption) (*Response, error)
	Do(request *http.Request) (*Response, error)
}

var _ Http = New()

var defaultReq = New(WithClient(http.DefaultClient))

func Get(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodGet, url, opts...)
}

func Post(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodPost, url, opts...)
}

func Put(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodPut, url, opts...)
}

func Patch(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodPatch, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodDelete, url, opts...)
}

func Head(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodHead, url, opts...)
}

func Connect(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodConnect, url, opts...)
}

func Options(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodOptions, url, opts...)
}

func Trace(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return defaultReq.do(ctx, http.MethodTrace, url, opts...)
}

type Request struct {
	opts    dialOptions
	Request *http.Request
}

/*
init request client
eg: New(WithClient(http.DefaultClient))
*/
func New(opts ...DialOption) *Request {
	req := &Request{}
	for _, opt := range opts {
		opt(&req.opts)
	}
	return req
}

func (req *Request) Get(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodGet, url, opts...)
}

func (req *Request) Post(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodPost, url, opts...)
}

func (req *Request) Put(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodPut, url, opts...)
}

func (req *Request) Patch(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodPatch, url, opts...)
}

func (req *Request) Delete(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodDelete, url, opts...)
}

func (req *Request) Head(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodHead, url, opts...)
}

func (req *Request) Connect(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodConnect, url, opts...)
}

func (req *Request) Options(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodOptions, url, opts...)
}

func (req *Request) Trace(ctx context.Context, url string, opts ...DialOption) (*Response, error) {
	return req.do(ctx, http.MethodTrace, url, opts...)
}

// 用于单独执行一个请求的goroutine，便于自定义请求过程
func (req *Request) Do(request *http.Request) (*Response, error) {
	return req.exec(request)
}

func (req Request) do(ctx context.Context, method string, url string, opts ...DialOption) (*Response, error) {
	for _, opt := range opts {
		opt(&req.opts)
	}

	// set params error
	if req.opts.err != nil {
		return nil, req.opts.err
	}

	if req.opts.query != "" {
		url += "?" + req.opts.query
	}
	request, err := http.NewRequestWithContext(ctx, method, url, req.opts.body)
	if err != nil {
		return nil, err
	}

	//set request headers
	for k, v := range req.opts.headers {
		request.Header.Set(k, v)
	}
	//set cookies
	if !req.opts.isSession {
		req.opts.client.Jar = nil
	}
	if len(req.opts.cookies) > 0 {
		if req.opts.client.Jar == nil {
			jar, err := cookiejar.New(nil)
			if err != nil {
				return nil, err
			}
			req.opts.client.Jar = jar
		}
		req.opts.client.Jar.SetCookies(request.URL, req.opts.cookies)
	}

	return req.exec(request)
}

func (req *Request) exec(request *http.Request) (*Response, error) {
	if request.Context() == nil {
		request = request.Clone(context.Background())
	}
	req.Request = request
	for _, before := range req.opts.before {
		if err := before(req); err != nil {
			return nil, err
		}
	}
	c := make(chan error)
	var resp *Response
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("panic: %+v\n", r)
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
		var (
			r   *http.Response
			err error
		)
		client := req.opts.client
		if retry := req.opts.retry; retry != nil {
			r, err = retry.backoff(func() (*http.Response, error) {
				return client.Do(req.Request)
			})
		} else {
			r, err = client.Do(req.Request)
		}

		if err != nil {
			c <- err
			return
		}
		resp = &Response{
			resp: r,
		}
		// request after middleware
		for _, after := range req.opts.after {
			if err := after(resp); err != nil {
				resp = nil
				c <- err
				return
			}
		}

		c <- nil
	}()

	ctx := req.Request.Context()
	select {
	case <-ctx.Done():
		<-c
		return nil, ctx.Err()
	case err := <-c:
		return resp, err
	}
}
