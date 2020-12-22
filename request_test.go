package requests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
)

var (
	host = "https://www.baidu.com/"
	req  = New(WithClient(http.DefaultClient))
)

func TestRequests_Get(t *testing.T) {
	// with param
	_, _ = Get(context.Background(), host, WithParam(map[string]string{
		"abc": "1",
	}))
	// struct参数匿名嵌套参数
	type Page struct {
		Offset int `form:"offset"`
	}
	type UserVip struct {
		IsVip bool `form:"is_vip"`
		Page
	}
	query := struct {
		Id   int    `form:"id"`
		Name string `form:"name"`
		UserVip
	}{
		Id:      1,
		Name:    "golang",
		UserVip: UserVip{IsVip: true},
	}
	_, _ = Get(context.Background(), host, WithQuery(query))
	//map参数
	query1 := map[string]interface{}{
		"a": 1,
	}
	_, _ = Get(context.Background(), host, WithQuery(query1))
	//struct指针参数
	query2 := UserVip{IsVip: true}
	_, _ = Get(context.Background(), host, WithQuery(&query2))

	// set headers
	_, _ = Get(context.Background(), host, WithHeaders(map[string]string{
		"x-session": "session",
	}))
}

func TestRequests_Post(t *testing.T) {
	//form
	_, _ = req.Post(context.Background(), host, WithForm(map[string]string{
		"id":    "1",
		"hello": "world",
	}))
	//json
	_, _ = req.Post(context.Background(), host, WithJSON(map[string]interface{}{
		"abc":   1,
		"hello": "world",
	}))
	//xml
	_, _ = req.Post(context.Background(), host, WithXML(map[string]interface{}{
		"abc":   1,
		"hello": "world",
	}))
}

func TestRequests_SetCookie(t *testing.T) {
	cookies := []*http.Cookie{
		{
			Name:  "hello",
			Value: "123",
		},
	}
	_, _ = Get(context.Background(), host, WithCookies(cookies...))
	cookies2 := []*http.Cookie{
		{
			Name:  "hello2",
			Value: "456",
		},
	}
	_, _ = Get(context.Background(), host, WithCookies(cookies2...), WithSession(false))
}

func TestRequests_Middleware(t *testing.T) {
	before := []BeforeRequest{
		func(request *Request) error {
			ctx := context.WithValue(request.Request.Context(), "token", "1234")
			request.Request = request.Request.Clone(ctx)
			fmt.Printf("method=%s, url=%s, body=%s\n", request.Request.Method, request.Request.URL, request.Request.Body)
			return nil
		},
	}
	after := []AfterRequest{
		func(resp *Response) error {
			ctx := resp.Response().Request.Context()
			fmt.Printf("token=%s\n", ctx.Value("token"))
			fmt.Printf("resp:%+v\n", resp)
			return nil
		},
	}
	ctx := context.Background()
	fmt.Println(ctx)
	_, _ = Get(ctx, host, WithParam(map[string]string{"id": "1"}), WithBefore(before...), WithAfter(after...))
}

func TestRequests_Retry(t *testing.T) {
	_, _ = Get(context.Background(), host, WithRetry())
}

func TestRequest_Debug(t *testing.T) {
	endpoint := "http://localhost:8080/v1/task/status"
	r, err := Get(context.Background(),
		endpoint,
		WithDebug(true),
		WithParam(map[string]string{
			"event": "event",
		}),
		WithHeaders(map[string]string{
			"Authorization": "Bearer xxx",
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	var result interface{}
	if err := r.JSON(&result); err != nil {
		log.Fatal(err)
	}
}

func TestRequest_File(t *testing.T) {
	endpoint := "http://localhost:8080/v1/task/status"
	_, err := Post(context.Background(),
		endpoint,
		WithDebug(true),
		WithFile(&File{
			Path: "LICENSE",
			Name: "license",
			Extras: map[string]string{
				"author":      "test",
				"title":       "My Document",
				"description": "A document with all the Go programming language secrets",
			},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
