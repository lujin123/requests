package requests

import (
	"context"
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
	_, _ = req.Post(context.Background(), host, WithJson(map[string]interface{}{
		"abc":   1,
		"hello": "world",
	}))
	//xml
	_, _ = req.Post(context.Background(), host, WithXML(map[string]interface{}{
		"abc":   1,
		"hello": "world",
	}))
}
