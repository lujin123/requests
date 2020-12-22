package requests

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type (
	Value map[string]string
	File  struct {
		Path   string
		Name   string
		Extras Value
	}
)

type dialOptions struct {
	err       error
	debug     bool
	client    *http.Client
	headers   Value
	cookies   []*http.Cookie
	query     string
	body      io.Reader
	isSession bool
	after     []AfterRequest
	before    []BeforeRequest
	retry     *retry
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

func WithDebug(debug bool) DialOption {
	return func(opts *dialOptions) {
		opts.debug = debug
		if debug {
			opts.before = append([]BeforeRequest{beforeMiddleLog}, opts.before...)
			opts.after = append([]AfterRequest{afterMiddleLog}, opts.after...)
		}
	}
}

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

func WithJSON(data interface{}) DialOption {
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

//直接设置一个请求body
func WithBody(body io.Reader) DialOption {
	return func(opts *dialOptions) {
		opts.body = body
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

func WithCookies(cookies ...*http.Cookie) DialOption {
	return func(opts *dialOptions) {
		opts.cookies = cookies
	}
}

//是否清空cookies
//如果设置成true，后续的请求都会带上前面请求返回的cookie，所以不要随便设置，只有在清楚需要的时候再设置
func WithSession(session bool) DialOption {
	return func(opts *dialOptions) {
		opts.isSession = session
	}
}

func WithBefore(middles ...BeforeRequest) DialOption {
	return func(opts *dialOptions) {
		opts.before = append(opts.before, middles...)
	}
}

func WithAfter(middles ...AfterRequest) DialOption {
	return func(opts *dialOptions) {
		opts.after = append(opts.after, middles...)
	}
}

func WithRetry(retries ...Retry) DialOption {
	return func(opts *dialOptions) {
		var retry Retry
		if len(retries) > 0 {
			retry = retries[0]
		} else {
			retry = &defaultRetry{}
		}
		opts.retry = newRetry(retry)
	}
}

//上传文件
func WithFile(file *File) DialOption {
	return func(opts *dialOptions) {
		f, err := os.Open(file.Path)
		if err != nil {
			opts.err = err
			return
		}
		defer f.Close()

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile(file.Name, filepath.Base(file.Path))
		if err != nil {
			opts.err = err
			return
		}
		if _, err := io.Copy(part, f); err != nil {
			opts.err = err
			return
		}

		for k, v := range file.Extras {
			_ = writer.WriteField(k, v)
		}

		if err := writer.Close(); err != nil {
			opts.err = err
			return
		}

		opts.body = &body
		opts.setContentType(writer.FormDataContentType())
	}
}
