package requests

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

type (
	Value map[string]string
)

type dialOptions struct {
	err     error
	client  *http.Client
	headers Value
	query   string
	body    io.Reader
}

func (opts *dialOptions) setContentType(contentType string) {
	opts.setHeader("Content-Type", contentType)
}

func (opts *dialOptions) setHeader(k, v string) {
	if opts.headers == nil {
		opts.headers = make(Value)
	}
	opts.headers[k] = v
}

type DialOption func(opts *dialOptions)

func WithParam(query Value) DialOption {
	return func(opts *dialOptions) {
		opts.query = mapToValues(query).Encode()
	}
}

//直接传递一个结构体指针作为query参数
//注意：
//1. 支持map、struct，其他的类型会直接panic，struct使用`form`指定字段名称，未指定的使用默认值
//2. 支持匿名嵌套，但不支持命名嵌套，内容不会解析，直接变成一个字符串
func WithQuery(i interface{}) DialOption {
	return func(opts *dialOptions) {
		opts.query = structToValues(i).Encode()
	}
}

func WithForm(form Value) DialOption {
	return func(opts *dialOptions) {
		s := mapToValues(form).Encode()
		opts.body = bytes.NewBufferString(s)
		opts.setContentType("application/x-www-form-urlencoded")
	}
}

func WithJson(data interface{}) DialOption {
	return func(opts *dialOptions) {
		buf, err := json.Marshal(data)
		if err != nil {
			opts.err = err
			return
		}
		opts.body = bytes.NewBuffer(buf)
		opts.setContentType("application/json")
	}
}

func WithXML(data interface{}) DialOption {
	return func(opts *dialOptions) {
		buf, err := xml.Marshal(data)
		if err != nil {
			opts.err = err
			return
		}
		opts.body = bytes.NewBuffer(buf)
		opts.setContentType("application/xml")
	}
}

func WithHeaders(headers Value) DialOption {
	return func(opts *dialOptions) {
		for k, v := range headers {
			opts.setHeader(k, v)
		}
	}
}

func WithClient(client *http.Client) DialOption {
	return func(opts *dialOptions) {
		opts.client = client
	}
}
