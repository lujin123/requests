package requests

import (
	"log"
	"net/http/httputil"
)

type BeforeRequest func(req *Request) error
type AfterRequest func(resp *Response) error

func beforeMiddleLog(req *Request) error {
	buf, err := httputil.DumpRequest(req.Request, true)
	if err != nil {
		return err
	}
	log.Printf("[request]\n%s\n", buf)
	return nil
}

func afterMiddleLog(resp *Response) error {
	response := resp.Response()
	buf, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	log.Printf("[response]\n%s\n", buf)
	return nil
}
