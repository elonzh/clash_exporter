package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"io"
	"net/url"
	"strconv"
	"sync"
	"time"
)
import "net/http"

type IClient interface {
	GetVersion() (*Version, error)
	GetProxies() (map[string]*Proxy, error)
	GetProxyDelay(proxyName string, testUrl string, timeout time.Duration) (uint16, error)
	GetProvidersProxies() (map[string]*Provider, error)
	ProviderProxiesHealthCheck(providerName string) error
	GetConnections() (*Snapshot, error)
}

const (
	DefaultTestUrl        = "http://www.gstatic.com/generate_204"
	DefaultTestUrlTimeout = 3 * time.Second
	DefaultClientTimeout  = 5 * time.Second
)

var (
	proxiesUrl, _          = url.Parse("/proxies")
	providersProxiesUrl, _ = url.Parse("/providers/proxies")
	connectionsUrl, _      = url.Parse("/connections")
	versionUrl, _          = url.Parse("/version")
)

type Client struct {
	BaseUrl *url.URL
	Secret  string
	client  *http.Client
}

func NewClient(baseUrl string, secret string) (*Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Timeout: DefaultClientTimeout,
	}
	return &Client{
		BaseUrl: u,
		Secret:  secret,
		client:  c,
	}, nil
}

func (c *Client) request(u *url.URL, v interface{}) error {
	u = c.BaseUrl.ResolveReference(u)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Secret))
	resp, err := c.client.Do(req)
	if err != nil || v == nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	//debug request
	//buf := bytes.NewBuffer(nil)
	//buf.WriteString(fmt.Sprintln(req.Method, req.URL.String()))
	//err = json.Indent(buf, data, "", "  ")
	//if err != nil {
	//	return err
	//}
	//fmt.Println(buf.String())

	_ = resp.Body.Close()
	err = json.Unmarshal(data, v)
	return err
}

func (c *Client) GetVersion() (*Version, error) {
	container := new(Version)
	if err := c.request(versionUrl, &container); err != nil {
		return nil, err
	}
	return container, nil
}

func (c *Client) GetProxies() (map[string]*Proxy, error) {
	container := make(map[string]map[string]*Proxy)
	if err := c.request(proxiesUrl, &container); err != nil {
		return nil, err
	}
	return container["proxies"], nil
}

func (c *Client) GetProxyDelay(proxyName string, testUrl string, timeout time.Duration) (uint16, error) {
	proxyDelayUrl, err := url.Parse(fmt.Sprintf("/proxies/%s/delay", proxyName))
	if err != nil {
		return 0, err
	}
	q := proxyDelayUrl.Query()
	q.Add("url", testUrl)
	q.Add("timeout", strconv.Itoa(int(timeout.Milliseconds())))
	proxyDelayUrl.RawQuery = q.Encode()
	container := make(map[string]uint16)
	if err := c.request(proxyDelayUrl, &container); err != nil {
		return 0, err
	}
	return container["delay"], nil
}

func (c *Client) GetProvidersProxies() (map[string]*Provider, error) {
	container := make(map[string]map[string]*Provider)
	if err := c.request(providersProxiesUrl, &container); err != nil {
		return nil, err
	}
	return container["providers"], nil
}

func (c *Client) ProviderProxiesHealthCheck(providerName string) error {
	u, err := url.Parse(fmt.Sprintf("/providers/proxies/%s/healthcheck", providerName))
	if err != nil {
		return err
	}
	return c.request(u, nil)
}

func (c *Client) GetConnections() (*Snapshot, error) {
	container := new(Snapshot)
	if err := c.request(connectionsUrl, &container); err != nil {
		return nil, err
	}
	return container, nil
}

func GetAllProxyDelay(proxies map[string]*Proxy, filter func(*Proxy) bool, client IClient, testUrl string, testUrlTimeout time.Duration) map[string]uint16 {
	if filter == nil {
		filter = func(*Proxy) bool {
			return true
		}
	}

	wg := sync.WaitGroup{}
	ch := make(chan struct {
		proxyName string
		delay     uint16
	})
	for _, proxy := range proxies {
		if filter(proxy) {
			wg.Add(1)
			go func(proxy *Proxy) {
				defer wg.Done()
				delay, err := client.GetProxyDelay(proxy.Name, testUrl, testUrlTimeout)
				if err != nil {
					level.Warn(logger).Log("msg", "error when get proxy delay", "err", err, "proxyType", proxy.Type, "proxyName", proxy.Name)
					delay = MaxDelay
				}
				ch <- struct {
					proxyName string
					delay     uint16
				}{proxyName: proxy.Name, delay: delay}
			}(proxy)
		}
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	rv := make(map[string]uint16)
	for v := range ch {
		rv[v.proxyName] = v.delay
	}
	return rv
}
