package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"os"
	"path"
	"testing"
	"time"
)

func expectMetrics(t *testing.T, c prometheus.Collector, fixture string) {
	exp, err := os.Open(path.Join("test", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}

type testClient struct {
}

func (c *testClient) GetVersion() (*Version, error) {
	return &Version{
		Premium: true,
		Version: "2021.04.08",
	}, nil
}

func (c *testClient) makeProxies(proxyNameTemplate string) []*Proxy {
	proxies := make([]*Proxy, 0, len(AllProxyTypes))
	now := time.Now()
	for i, t := range AllProxyTypes {
		n := fmt.Sprintf(proxyNameTemplate, t)
		proxies = append(proxies, &Proxy{
			Type: t,
			Name: n,
			History: []*ProxyDelay{
				{
					Time:  now,
					Delay: uint16(i),
				},
			},
		})
	}
	return proxies
}

func (c *testClient) GetProxies() (map[string]*Proxy, error) {
	proxies := make(map[string]*Proxy, len(AllProxyTypes))
	for _, p := range c.makeProxies("proxy_%s") {
		proxies[p.Name] = p
	}
	return proxies, nil
}

func (c *testClient) GetProxyDelay(proxyName string, testUrl string, timeout time.Duration) (uint16, error) {
	return 666, nil
}

func (c *testClient) GetProvidersProxies() (map[string]*Provider, error) {
	return map[string]*Provider{
		"provider_1": {
			Type:        "Proxy",
			Name:        "provider_1",
			VehicleType: VehicleTypeHTTP,
			UpdatedAt:   time.Time{},
			Proxies:     c.makeProxies("provider_1_proxy_%s"),
		},
		"provider_2": {
			Type:        "Proxy",
			Name:        "provider_2",
			VehicleType: VehicleTypeFile,
			Proxies:     c.makeProxies("provider_2_proxy_%s"),
		},
		"provider_3": {
			Type:        "Proxy",
			Name:        "provider_2",
			VehicleType: VehicleTypeCompatible,
			Proxies:     c.makeProxies("provider_3_proxy_%s"),
		},
	}, nil
}

func (c *testClient) ProviderProxiesHealthCheck(providerName string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (c *testClient) GetConnections() (*Snapshot, error) {
	return &Snapshot{
		DownloadTotal: 111,
		UploadTotal:   222,
		Connections:   nil,
	}, nil
}

func TestExporter(t *testing.T) {
	e, err := NewExporter(&testClient{}, DefaultTestUrl, DefaultTestUrlTimeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("expect test metrics", func(t *testing.T) {
		expectMetrics(t, e, "normal.metrics")
	})

	t.Run("CollectToText", func(t *testing.T) {
		text, err := CollectToText(e)
		if err != nil || text == "" {
			t.Errorf(text, err)
		}
		//f, err := os.Create("test/test.metrics")
		//f.WriteString(text)
		//f.Close()
	})
}
