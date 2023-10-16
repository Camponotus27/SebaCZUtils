package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type Response struct {
	HTTPResponse *http.Response
	Code         int
	Bytes        []byte
	Body         string
	Error        error
}

func Deserialize[T any](r *Response) (result T, err error) {
	if r == nil {
		err = errors.New("response is nil")
		return
	}

	if r.Error != nil {
		err = fmt.Errorf("previus error: %v", r.Error)
		return
	}

	err = json.Unmarshal(r.Bytes, &result)
	return

	// TODO: validate is read all fields
	valueOf := reflect.ValueOf(result)
	a := valueOf.MapKeys()
	fmt.Println(a)
	// TODO: validate, not worke
	if valueOf.IsNil() || valueOf.IsZero() || valueOf.IsValid() {
		err = fmt.Errorf("JSON is not %T: %s", result, string(r.Bytes))
	}

	return
}

type Client struct {
	host   string
	method string
	path   string
	Header map[string]string
	client http.Client
	body   interface{}
}

func NewClientCall(host string) *Client {
	if host == "" {
		panic("not host provided")
	}
	return &Client{host: host}
}

func (c *Client) SetBody(body interface{}) *Client {
	c.body = body
	return c
}

func (c *Client) Get(path string) *Response {
	if path == "" {
		return &Response{Error: errors.New("not path provided")}
	}
	c.method = http.MethodGet
	c.path = path
	return c.Do()
}

func (c *Client) Post(path string) *Response {
	if path == "" {
		return &Response{Error: errors.New("not path provided")}
	}
	c.method = http.MethodPost
	c.path = path
	return c.Do()
}

func (c *Client) SetAuthorizationHeader(token string) *Client {
	if c.Header == nil {
		c.Header = make(map[string]string)
	}
	c.Header["Authorization"] = token
	return c
}

func (c *Client) SetMethodPost() *Client {
	c.method = http.MethodPost
	return c
}

func (c *Client) SetMethodGet() *Client {
	c.method = http.MethodGet
	return c
}

func (c *Client) SetPath(path string) *Client {
	c.path = path
	return c
}

func (c *Client) Do() *Response {

	bodyBytes, marshalErr := json.Marshal(c.body)
	if marshalErr != nil {
		return &Response{Error: marshalErr}
	}

	reader := strings.NewReader(string(bodyBytes))
	req, reqErr := http.NewRequest(c.method,
		fmt.Sprintf("%s/%s", c.host, c.path),
		reader)
	if reqErr != nil {
		return &Response{
			Error: reqErr,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	for header, value := range c.Header {
		req.Header.Set(header, value)
	}

	resp, callErr := c.client.Do(req)
	if callErr != nil {
		return &Response{
			HTTPResponse: resp,
			Code:         resp.StatusCode,
			Error:        callErr,
		}
	}
	defer resp.Body.Close()

	bytes, readByteErr := io.ReadAll(resp.Body)

	return &Response{
		HTTPResponse: resp,
		Code:         resp.StatusCode,
		Bytes:        bytes,
		Body:         string(bytes),
		Error:        readByteErr,
	}
}
