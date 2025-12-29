package youtube_transcript_api

import (
	"fmt"
	"net/url"
	"strings"
)

// InvalidProxyConfig 代理配置无效错误
type InvalidProxyConfig struct {
	Message string
}

func (e *InvalidProxyConfig) Error() string {
	return e.Message
}

// ProxyConfig 代理配置接口
type ProxyConfig interface {
	// ToProxyURLs 返回代理 URL 映射（http 和 https）
	ToProxyURLs() (httpURL, httpsURL string)
	// PreventKeepingConnectionsAlive 是否阻止保持连接（用于轮换代理）
	PreventKeepingConnectionsAlive() bool
	// RetriesWhenBlocked 被阻止时的重试次数
	RetriesWhenBlocked() int
}

// GenericProxyConfig 通用 HTTP/HTTPS/SOCKS 代理配置
type GenericProxyConfig struct {
	HTTPURL  string
	HTTPSURL string
}

// NewGenericProxyConfig 创建通用代理配置
func NewGenericProxyConfig(httpURL, httpsURL string) (*GenericProxyConfig, error) {
	if httpURL == "" && httpsURL == "" {
		return nil, &InvalidProxyConfig{
			Message: "GenericProxyConfig requires you to define at least one of the two: http or https",
		}
	}
	return &GenericProxyConfig{
		HTTPURL:  httpURL,
		HTTPSURL: httpsURL,
	}, nil
}

func (g *GenericProxyConfig) ToProxyURLs() (httpURL, httpsURL string) {
	if g.HTTPURL != "" {
		httpURL = g.HTTPURL
	} else {
		httpURL = g.HTTPSURL
	}
	if g.HTTPSURL != "" {
		httpsURL = g.HTTPSURL
	} else {
		httpsURL = g.HTTPURL
	}
	return
}

func (g *GenericProxyConfig) PreventKeepingConnectionsAlive() bool {
	return false
}

func (g *GenericProxyConfig) RetriesWhenBlocked() int {
	return 0
}

// WebshareProxyConfig Webshare 轮换住宅代理配置
type WebshareProxyConfig struct {
	*GenericProxyConfig
	ProxyUsername           string
	ProxyPassword           string
	FilterIPLocations       []string
	RetriesWhenBlockedCount int
	DomainName              string
	ProxyPort               int
}

const (
	WebshareDefaultDomainName = "p.webshare.io"
	WebshareDefaultPort       = 80
)

// NewWebshareProxyConfig 创建 Webshare 代理配置
func NewWebshareProxyConfig(
	proxyUsername string,
	proxyPassword string,
	filterIPLocations []string,
	retriesWhenBlocked int,
	domainName string,
	proxyPort int,
) (*WebshareProxyConfig, error) {
	if domainName == "" {
		domainName = WebshareDefaultDomainName
	}
	if proxyPort == 0 {
		proxyPort = WebshareDefaultPort
	}

	// 创建基础配置（URL 会在 URL() 方法中生成）
	baseConfig, err := NewGenericProxyConfig("", "")
	if err != nil {
		return nil, err
	}

	return &WebshareProxyConfig{
		GenericProxyConfig:      baseConfig,
		ProxyUsername:           proxyUsername,
		ProxyPassword:           proxyPassword,
		FilterIPLocations:       filterIPLocations,
		RetriesWhenBlockedCount: retriesWhenBlocked,
		DomainName:              domainName,
		ProxyPort:               proxyPort,
	}, nil
}

// URL 生成 Webshare 代理 URL
func (w *WebshareProxyConfig) URL() string {
	var locationCodes strings.Builder
	for _, locationCode := range w.FilterIPLocations {
		locationCodes.WriteString(fmt.Sprintf("-%s", strings.ToUpper(locationCode)))
	}

	proxyURL := fmt.Sprintf("http://%s%s-rotate:%s@%s:%d/",
		w.ProxyUsername,
		locationCodes.String(),
		w.ProxyPassword,
		w.DomainName,
		w.ProxyPort,
	)
	return proxyURL
}

func (w *WebshareProxyConfig) ToProxyURLs() (httpURL, httpsURL string) {
	url := w.URL()
	return url, url
}

func (w *WebshareProxyConfig) PreventKeepingConnectionsAlive() bool {
	return true
}

func (w *WebshareProxyConfig) RetriesWhenBlocked() int {
	return w.RetriesWhenBlockedCount
}

// SetupHTTPClientProxy 为 HTTP 客户端设置代理
func SetupHTTPClientProxy(client *HTTPClient, proxyConfig ProxyConfig) error {
	if proxyConfig == nil {
		return nil
	}

	httpURL, httpsURL := proxyConfig.ToProxyURLs()

	// 解析代理 URL
	if httpURL != "" {
		httpProxyURL, err := url.Parse(httpURL)
		if err != nil {
			return fmt.Errorf("invalid HTTP proxy URL: %w", err)
		}
		client.HTTPProxy = httpProxyURL
	}

	if httpsURL != "" {
		httpsProxyURL, err := url.Parse(httpsURL)
		if err != nil {
			return fmt.Errorf("invalid HTTPS proxy URL: %w", err)
		}
		client.HTTPSProxy = httpsProxyURL
	}

	// 如果配置要求阻止保持连接，设置 Connection: close 头
	if proxyConfig.PreventKeepingConnectionsAlive() {
		client.Headers["Connection"] = "close"
	}

	return nil
}
