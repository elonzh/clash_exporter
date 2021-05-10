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

func (t *testClient) GetVersion() (*Version, error) {
	return &Version{
		Premium: true,
		Version: "2021.04.08",
	}, nil
}

func (t *testClient) GetProxies() (map[string]*Proxy, error) {
	proxies := make(map[string]*Proxy, len(AllProxyTypes))
	for _, t := range AllProxyTypes {
		n := fmt.Sprintf("proxy_%s", t)
		proxies[n] = &Proxy{
			Type: t,
			Name: n,
		}
	}
	return proxies, nil
}

func (t *testClient) GetProxyDelay(proxyName string, testUrl string, timeout time.Duration) (uint16, error) {
	return 666, nil
}

func (t *testClient) GetConnections() (*Snapshot, error) {
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
	expectMetrics(t, e, "normal.metrics")
}
