# requests

## http method
- [x] GET
- [x] POST
- [x] PUT
- [x] PATCH
- [x] DELETE
- [x] HEAD
- [x] CONNECT
- [x] OPTIONS
- [x] TRACE

## TODO
- [ ] request middleware
- [ ] request retry?

## Example

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/lujin123/requests"
)

type User struct {
	Id   int    `form:"id" xml:"id"`
	Name string `form:"name" xml:"name"`
}

func main() {
	endpoint := "https://www.baidu.com/"
	//直接请求
	_, _ = requests.Get(context.Background(), endpoint)
	//使用字符串map作为请求参数
	_, _ = requests.Get(context.Background(), endpoint, requests.WithParam(map[string]string{
		"id":  "1",
		"key": "abc",
	}))
	//使用结构体作为请求参数
	//参数key使用form的tag,如果不指定form直接使用field name
	query := User{
		Id:   1,
		Name: "hello",
	}
	_, _ = requests.Get(context.Background(), endpoint, requests.WithQuery(&query))
	//空body的POST请求
	_, _ = requests.Post(context.Background(), endpoint)

	//使用json格式数据作为请求body
	jsonData := User{
		Id:   2,
		Name: "json",
	}
	_, _ = requests.Post(context.Background(), endpoint, requests.WithJson(jsonData))

	//使用xml格式数据作为请求body
	xmlData := User{
		Id:   3,
		Name: "xml",
	}
	_, _ = requests.Post(context.Background(), endpoint, requests.WithXML(xmlData))

	//设置请求头
	headers := map[string]string{
		"Content-Type":  "application/xml",
		"custom-header": "custom",
	}
	_, _ = requests.Post(context.Background(), endpoint, requests.WithXML(xmlData), requests.WithHeaders(headers))

	//自定义本地请求的client
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	_, _ = requests.Post(context.Background(), endpoint, requests.WithClient(&client))

	//自定义一个请求client
	newRequests := requests.New(requests.WithClient(&client))
	_, _ = newRequests.Post(context.Background(), endpoint)

	//处理请求结果
	resp, err := requests.Get(context.Background(), endpoint)
	if err != nil {
		return
	}
	//返回值只能读取一次
	//返回内容字符串
	fmt.Println(resp.Text())
	//json
	var jsonResp interface{}
	_ = resp.Json(&jsonResp)
	fmt.Println(jsonResp)
	//xml
	var xmlResp interface{}
	_ = resp.XML(&xmlResp)
	fmt.Println(xmlResp)
	//读取字节流，resp.Raw()返回的是一个channel
	var buffs bytes.Buffer
	for buf := range resp.Raw() {
		buffs.Write(buf)
	}
	fmt.Println(buffs.Bytes())
}

```
