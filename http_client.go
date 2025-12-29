package youtube_transcript_api

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// HTTPClient HTTP 客户端包装
type HTTPClient struct {
	client     *http.Client
	Headers    map[string]string
	HTTPProxy  *url.URL
	HTTPSProxy *url.URL
	Jar        *cookiejar.Jar
}

// NewHTTPClient 创建新的 HTTP 客户端
func NewHTTPClient() (*HTTPClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return &HTTPClient{
		client:  client,
		Headers: make(map[string]string),
		Jar:     jar,
	}, nil
}

// Get 发送 GET 请求
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// 设置代理
	transport := &http.Transport{}
	if c.HTTPProxy != nil {
		transport.Proxy = http.ProxyURL(c.HTTPProxy)
	}
	if c.HTTPSProxy != nil {
		transport.Proxy = http.ProxyURL(c.HTTPSProxy)
	}
	c.client.Transport = transport

	return c.client.Do(req)
}

// Post 发送 POST 请求
func (c *HTTPClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", contentType)

	// 设置代理
	transport := &http.Transport{}
	if c.HTTPProxy != nil {
		transport.Proxy = http.ProxyURL(c.HTTPProxy)
	}
	if c.HTTPSProxy != nil {
		transport.Proxy = http.ProxyURL(c.HTTPSProxy)
	}
	c.client.Transport = transport

	return c.client.Do(req)
}
